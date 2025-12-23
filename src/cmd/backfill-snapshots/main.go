package main

import (
	"context"
	"github.com/ctfloyd/hazelmere-api/src/cmd/backfill-snapshots/wom"
	"github.com/ctfloyd/hazelmere-api/src/internal/core/snapshot"
	"github.com/ctfloyd/hazelmere-api/src/internal/core/user"
	"github.com/ctfloyd/hazelmere-api/src/internal/initialize"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_config"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log/slog"
	"maps"
	"os"
	"os/signal"
	"slices"
	"strings"
	"time"
)

var needBackfill = []string{"msk"}

var userRepo user.UserRepository
var snapshotRepo snapshot.SnapshotRepository
var wiseOldMan *wom.Client
var am map[string]string

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	config := hz_config.NewConfigFromPath("config/dev.json")
	if err := config.Read(); err != nil {
		panic(err)
	}

	client, err := initialize.MongoClient(
		config.ValueOrPanic("mongo.connection.host"),
		config.ValueOrPanic("mongo.connection.username"),
		config.ValueOrPanic("mongo.connection.password"),
	)
	if err != nil {
		panic(err)
	}
	defer initialize.MongoCleanup(ctx, client)

	dbName := config.ValueOrPanic("mongo.database.name")
	sCName := config.ValueOrPanic("mongo.database.collections.snapshot")
	uCName := config.ValueOrPanic("mongo.database.collections.user")

	sCollection := client.Database(dbName).Collection(sCName)
	uCollection := client.Database(dbName).Collection(uCName)

	logger := hz_logger.NewZeroLogAdapater(hz_logger.LogLevelInfo)

	userRepo = user.NewUserRepository(uCollection, logger)
	snapshotRepo = snapshot.NewSnapshotRepository(sCollection, logger)
	wiseOldMan = wom.NewClient(logger)

	am = populateActivityMap(ctx, sCollection)

	users, err := userRepo.GetAllUsers(ctx)
	if err != nil {
		slog.Error("could not get all users", slog.Any("error", err))
		os.Exit(1)
	}

	for _, user := range users {
		if slices.Contains(needBackfill, user.RunescapeName) && user.TrackingStatus == "ENABLED" {
			backfillMissingDataForUser(ctx, user.RunescapeName, user.Id)
		}
	}

}

type MissingRange struct {
	Start, End time.Time
}

func backfillMissingDataForUser(ctx context.Context, username string, userId string) {
	slog.Info("backfilling missing data for user", slog.String("username", username))

	ranges := gatherRanges(ctx, username, userId)
	if len(ranges) == 0 {
		slog.Warn("no ranges for user", slog.String("user", username))
		return
	}

	womSnapshots := getWomSnapshots(ranges, username)
	slog.Info("no dedupe wom snapshot cnt", slog.Int("cnt", len(womSnapshots)))
	womSnapshots = dedupeWomSnapshots(womSnapshots)
	slog.Info("dedupe wom snapshot cnt", slog.Int("cnt", len(womSnapshots)))

	for i, ws := range womSnapshots {
		slog.Info("insert snapshot", slog.String("username", username), slog.Int("idx", i), slog.Int("cnt", len(womSnapshots)))
		ds := convertToSnapshotData(userId, ws)
		_, err := snapshotRepo.InsertSnapshot(ctx, ds)
		if err != nil {
			slog.Error("failed to insert snapshot data for user", slog.String("user", username), slog.Any("error", err))
		}
	}
}

func dedupeWomSnapshots(snapshots []wom.Snapshot) []wom.Snapshot {
	snapshotByDay := make(map[time.Time]wom.Snapshot)
	for _, snapshot := range snapshots {
		day := startOfDay(snapshot.CreatedAt)
		if exist, ok := snapshotByDay[day]; ok {
			if snapshot.CreatedAt.After(exist.CreatedAt) {
				snapshotByDay[day] = snapshot
			}
		} else {
			snapshotByDay[day] = snapshot
		}
	}
	return slices.Collect(maps.Values(snapshotByDay))
}

func getWomSnapshots(ranges []MissingRange, username string) []wom.Snapshot {
	womSnapshots := make([]wom.Snapshot, 0, 1000)

	for _, rng := range ranges {
		if rng.Start.Before(rng.End) {
			slog.Info("Getting snapshot data from WOM for range", slog.Time("start", rng.Start), slog.Time("end", rng.End), slog.String("username", username))
			snapshots, err := wiseOldMan.GetPlayerSnapshots(username, rng.Start, rng.End)
			if err != nil {
				slog.Error("Failed to get snapshot data from WOM for user.",
					slog.String("user", username),
					slog.Any("error", err),
					slog.Time("start", rng.Start),
					slog.Time("end", rng.End),
				)
			}

			if len(snapshots) > 0 {
				womSnapshots = append(womSnapshots, snapshots...)
			}
		}
	}

	return womSnapshots
}

func gatherRanges(ctx context.Context, username string, userId string) []MissingRange {
	details, err := wiseOldMan.GetPlayerDetails(username)
	if err != nil {
		slog.Error("failed to get wom player details", slog.String("username", username), slog.Any("error", err))
		return nil
	}

	times, err := snapshotRepo.GetAllTimestampsForUser(ctx, userId)
	if err != nil {
		slog.Error("failed to get all timestamps for user", slog.String("username", username), slog.Any("error", err))
		return nil
	}

	snapshotsByDay := make(map[time.Time]struct{})
	for _, t := range times {
		snapshotsByDay[startOfDay(t.Timestamp)] = struct{}{}
	}

	var ranges []MissingRange

	rangeStart := time.Time{}
	for day := startOfDay(details.RegisteredAt); day.Before(endOfDay(time.Now())); day = day.AddDate(0, 0, 1) {
		if _, ok := snapshotsByDay[day]; !ok {
			if rangeStart.IsZero() {
				rangeStart = day
			}
		} else {
			if !rangeStart.IsZero() {
				ranges = append(ranges, MissingRange{Start: rangeStart, End: day})
				rangeStart = time.Time{}
			}
		}
	}

	if !rangeStart.IsZero() {
		ranges = append(ranges, MissingRange{Start: rangeStart, End: startOfDay(time.Now().AddDate(0, 0, 1))})
	}

	return ranges
}

func convertToSnapshotData(userId string, ss wom.Snapshot) snapshot.HiscoreSnapshotData {
	return snapshot.HiscoreSnapshotData{
		Id:         uuid.NewString(),
		UserId:     userId,
		Timestamp:  ss.CreatedAt,
		Skills:     convertSkills(ss.Data.Skills),
		Bosses:     convertBosses(ss.Data.Bosses),
		Activities: convertActivities(ss.Data.Activities),
		Source:     "WOM_BACKFILL_122025",
	}
}

func convertActivities(a map[string]wom.SnapshotActivity) []snapshot.ActivitySnapshotData {
	convert := make([]snapshot.ActivitySnapshotData, 0, len(snapshot.AllActivityActivityTypes))
	for _, activity := range snapshot.AllActivityActivityTypes {
		if ac, ok := a[strings.ToLower(string(activity))]; ok {
			convert = append(convert, snapshot.ActivitySnapshotData{
				ActivityType: string(activity),
				Name:         am[string(activity)],
				Rank:         ac.Rank,
				Score:        ac.Score,
			})
		} else {
			convert = append(convert, snapshot.ActivitySnapshotData{
				ActivityType: string(activity),
				Name:         am[string(activity)],
				Rank:         -1,
				Score:        -1,
			})
		}

	}
	return convert

}

func convertBosses(b map[string]wom.SnapshotBoss) []snapshot.BossSnapshotData {
	convert := make([]snapshot.BossSnapshotData, 0, len(snapshot.AllBossActivityTypes))
	for _, boss := range snapshot.AllBossActivityTypes {
		if bo, ok := b[strings.ToLower(string(boss))]; ok {
			convert = append(convert, snapshot.BossSnapshotData{
				ActivityType: string(boss),
				Name:         am[string(boss)],
				Rank:         bo.Rank,
				KillCount:    bo.Kills,
			})
		} else {
			convert = append(convert, snapshot.BossSnapshotData{
				ActivityType: string(boss),
				Name:         am[string(boss)],
				Rank:         -1,
				KillCount:    -1,
			})
		}

	}
	return convert

}

func convertSkills(s map[string]wom.SnapshotSkill) []snapshot.SkillSnapshotData {
	convert := make([]snapshot.SkillSnapshotData, 0, len(snapshot.AllSkillActivityTypes))
	for _, skill := range snapshot.AllSkillActivityTypes {
		key := strings.ToLower(string(skill))
		if skill == snapshot.ActivityTypeRunecraft {
			key = "runecrafting"
		}

		if sk, ok := s[key]; ok {
			convert = append(convert, snapshot.SkillSnapshotData{
				ActivityType: string(skill),
				Name:         am[string(skill)],
				Level:        sk.Level,
				Experience:   sk.Experience,
				Rank:         sk.Rank,
			})
		} else {
			convert = append(convert, snapshot.SkillSnapshotData{
				ActivityType: string(skill),
				Name:         am[string(skill)],
				Level:        -1,
				Experience:   -1,
				Rank:         -1,
			})
		}
	}
	return convert
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func endOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 23, 59, 59, 0, t.Location())
}

func populateActivityMap(ctx context.Context, collection *mongo.Collection) map[string]string {
	pipeline := mongo.Pipeline{
		{
			{Key: "$project", Value: bson.D{
				{Key: "typeToNameMap", Value: bson.D{
					{Key: "$arrayToObject", Value: bson.D{
						{Key: "$concatArrays", Value: bson.A{
							bson.D{{Key: "$map", Value: bson.D{
								{Key: "input", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$skills", bson.A{}}}}},
								{Key: "as", Value: "skill"},
								{Key: "in", Value: bson.D{
									{Key: "k", Value: "$$skill.activityType"},
									{Key: "v", Value: "$$skill.name"},
								}},
							}}},
							bson.D{{Key: "$map", Value: bson.D{
								{Key: "input", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$bosses", bson.A{}}}}},
								{Key: "as", Value: "boss"},
								{Key: "in", Value: bson.D{
									{Key: "k", Value: "$$boss.activityType"},
									{Key: "v", Value: "$$boss.name"},
								}},
							}}},
							bson.D{{Key: "$map", Value: bson.D{
								{Key: "input", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$activities", bson.A{}}}}},
								{Key: "as", Value: "activity"},
								{Key: "in", Value: bson.D{
									{Key: "k", Value: "$$activity.activitytype"},
									{Key: "v", Value: "$$activity.name"},
								}},
							}}},
						}},
					}},
				}},
			}},
		},
		{
			{Key: "$limit", Value: 1},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		panic(err)
	}
	defer cursor.Close(ctx)

	var result struct {
		TypeToNameMap map[string]string `bson:"typeToNameMap"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			panic(err)
		}
		return result.TypeToNameMap
	}

	if err := cursor.Err(); err != nil {
		panic(err)
	}

	panic("no docs found")
}
