package backfill

import (
	"context"
	"fmt"
	wom2 "github.com/ctfloyd/hazelmere-api/src/internal/dependency/wom"
	"log/slog"
	"maps"
	"os"
	"os/signal"
	"slices"
	"strings"
	"time"

	"github.com/ctfloyd/hazelmere-api/src/internal/core/snapshot"
	"github.com/ctfloyd/hazelmere-api/src/internal/core/user"
	"github.com/ctfloyd/hazelmere-api/src/internal/initialize"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_config"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Users to backfill - can be customized via args in the future
var needBackfill = []string{"msk"}

type snapshotBackfiller struct {
	userRepo     user.UserRepository
	snapshotRepo snapshot.SnapshotRepository
	wiseOldMan   *wom2.Client
	activityMap  map[string]string
}

func RunSnapshots(configPath string, args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	config := hz_config.NewConfigFromPath(configPath)
	if err := config.Read(); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	client, err := initialize.MongoClient(
		config.ValueOrPanic("mongo.connection.host"),
		config.ValueOrPanic("mongo.connection.username"),
		config.ValueOrPanic("mongo.connection.password"),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer initialize.MongoCleanup(ctx, client)

	dbName := config.ValueOrPanic("mongo.database.name")
	sCName := config.ValueOrPanic("mongo.database.collections.snapshot")
	uCName := config.ValueOrPanic("mongo.database.collections.user")

	sCollection := client.Database(dbName).Collection(sCName)
	uCollection := client.Database(dbName).Collection(uCName)

	logger := hz_logger.NewZeroLogAdapater(hz_logger.LogLevelInfo)

	backfiller := &snapshotBackfiller{
		userRepo:     user.NewUserRepository(uCollection, logger),
		snapshotRepo: snapshot.NewSnapshotRepository(sCollection, logger),
		wiseOldMan:   wom2.NewClient(logger),
		activityMap:  populateActivityMap(ctx, sCollection),
	}

	fmt.Println("=== Snapshot Backfill Script ===")
	fmt.Printf("Database: %s\n", dbName)
	fmt.Printf("Users to backfill: %v\n\n", needBackfill)

	users, err := backfiller.userRepo.GetAllUsers(ctx)
	if err != nil {
		return fmt.Errorf("could not get all users: %w", err)
	}

	for _, u := range users {
		if slices.Contains(needBackfill, u.RunescapeName) && u.TrackingStatus == "ENABLED" {
			backfiller.backfillMissingDataForUser(ctx, u.RunescapeName, u.Id)
		}
	}

	fmt.Println("\n=== Backfill Complete ===")
	return nil
}

type missingRange struct {
	Start, End time.Time
}

func (b *snapshotBackfiller) backfillMissingDataForUser(ctx context.Context, username string, userId string) {
	slog.Info("backfilling missing data for user", slog.String("username", username))

	ranges := b.gatherRanges(ctx, username, userId)
	if len(ranges) == 0 {
		slog.Warn("no ranges for user", slog.String("user", username))
		return
	}

	womSnapshots := b.getWomSnapshots(ranges, username)
	slog.Info("no dedupe wom snapshot cnt", slog.Int("cnt", len(womSnapshots)))
	womSnapshots = dedupeWomSnapshots(womSnapshots)
	slog.Info("dedupe wom snapshot cnt", slog.Int("cnt", len(womSnapshots)))

	for i, ws := range womSnapshots {
		slog.Info("insert snapshot", slog.String("username", username), slog.Int("idx", i), slog.Int("cnt", len(womSnapshots)))
		ds := b.convertToSnapshotData(userId, ws)
		_, err := b.snapshotRepo.InsertSnapshot(ctx, ds)
		if err != nil {
			slog.Error("failed to insert snapshot data for user", slog.String("user", username), slog.Any("error", err))
		}
	}
}

func dedupeWomSnapshots(snapshots []wom2.Snapshot) []wom2.Snapshot {
	snapshotByDay := make(map[time.Time]wom2.Snapshot)
	for _, s := range snapshots {
		day := startOfDay(s.CreatedAt)
		if exist, ok := snapshotByDay[day]; ok {
			if s.CreatedAt.After(exist.CreatedAt) {
				snapshotByDay[day] = s
			}
		} else {
			snapshotByDay[day] = s
		}
	}
	return slices.Collect(maps.Values(snapshotByDay))
}

func (b *snapshotBackfiller) getWomSnapshots(ranges []missingRange, username string) []wom2.Snapshot {
	womSnapshots := make([]wom2.Snapshot, 0, 1000)

	for _, rng := range ranges {
		if rng.Start.Before(rng.End) {
			slog.Info("Getting snapshot data from WOM for range", slog.Time("start", rng.Start), slog.Time("end", rng.End), slog.String("username", username))
			snapshots, err := b.wiseOldMan.GetPlayerSnapshots(username, rng.Start, rng.End)
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

func (b *snapshotBackfiller) gatherRanges(ctx context.Context, username string, userId string) []missingRange {
	details, err := b.wiseOldMan.GetPlayerDetails(username)
	if err != nil {
		slog.Error("failed to get wom player details", slog.String("username", username), slog.Any("error", err))
		return nil
	}

	times, err := b.snapshotRepo.GetAllTimestampsForUser(ctx, userId)
	if err != nil {
		slog.Error("failed to get all timestamps for user", slog.String("username", username), slog.Any("error", err))
		return nil
	}

	snapshotsByDay := make(map[time.Time]struct{})
	for _, t := range times {
		snapshotsByDay[startOfDay(t.Timestamp)] = struct{}{}
	}

	var ranges []missingRange

	rangeStart := time.Time{}
	for day := startOfDay(details.RegisteredAt); day.Before(endOfDay(time.Now())); day = day.AddDate(0, 0, 1) {
		if _, ok := snapshotsByDay[day]; !ok {
			if rangeStart.IsZero() {
				rangeStart = day
			}
		} else {
			if !rangeStart.IsZero() {
				ranges = append(ranges, missingRange{Start: rangeStart, End: day})
				rangeStart = time.Time{}
			}
		}
	}

	if !rangeStart.IsZero() {
		ranges = append(ranges, missingRange{Start: rangeStart, End: startOfDay(time.Now().AddDate(0, 0, 1))})
	}

	return ranges
}

func (b *snapshotBackfiller) convertToSnapshotData(userId string, ss wom2.Snapshot) snapshot.HiscoreSnapshotData {
	return snapshot.HiscoreSnapshotData{
		Id:         uuid.NewString(),
		UserId:     userId,
		Timestamp:  ss.CreatedAt,
		Skills:     b.convertSkills(ss.Data.Skills),
		Bosses:     b.convertBosses(ss.Data.Bosses),
		Activities: b.convertActivities(ss.Data.Activities),
		Source:     "WOM_BACKFILL_122025",
	}
}

func (b *snapshotBackfiller) convertActivities(a map[string]wom2.SnapshotActivity) []snapshot.ActivitySnapshotData {
	convert := make([]snapshot.ActivitySnapshotData, 0, len(snapshot.AllActivityActivityTypes))
	for _, activity := range snapshot.AllActivityActivityTypes {
		if ac, ok := a[strings.ToLower(string(activity))]; ok {
			convert = append(convert, snapshot.ActivitySnapshotData{
				ActivityType: string(activity),
				Name:         b.activityMap[string(activity)],
				Rank:         ac.Rank,
				Score:        ac.Score,
			})
		} else {
			convert = append(convert, snapshot.ActivitySnapshotData{
				ActivityType: string(activity),
				Name:         b.activityMap[string(activity)],
				Rank:         -1,
				Score:        -1,
			})
		}
	}
	return convert
}

func (b *snapshotBackfiller) convertBosses(bosses map[string]wom2.SnapshotBoss) []snapshot.BossSnapshotData {
	convert := make([]snapshot.BossSnapshotData, 0, len(snapshot.AllBossActivityTypes))
	for _, boss := range snapshot.AllBossActivityTypes {
		if bo, ok := bosses[strings.ToLower(string(boss))]; ok {
			convert = append(convert, snapshot.BossSnapshotData{
				ActivityType: string(boss),
				Name:         b.activityMap[string(boss)],
				Rank:         bo.Rank,
				KillCount:    bo.Kills,
			})
		} else {
			convert = append(convert, snapshot.BossSnapshotData{
				ActivityType: string(boss),
				Name:         b.activityMap[string(boss)],
				Rank:         -1,
				KillCount:    -1,
			})
		}
	}
	return convert
}

func (b *snapshotBackfiller) convertSkills(s map[string]wom2.SnapshotSkill) []snapshot.SkillSnapshotData {
	convert := make([]snapshot.SkillSnapshotData, 0, len(snapshot.AllSkillActivityTypes))
	for _, skill := range snapshot.AllSkillActivityTypes {
		key := strings.ToLower(string(skill))
		if skill == snapshot.ActivityTypeRunecraft {
			key = "runecrafting"
		}

		if sk, ok := s[key]; ok {
			convert = append(convert, snapshot.SkillSnapshotData{
				ActivityType: string(skill),
				Name:         b.activityMap[string(skill)],
				Level:        sk.Level,
				Experience:   sk.Experience,
				Rank:         sk.Rank,
			})
		} else {
			convert = append(convert, snapshot.SkillSnapshotData{
				ActivityType: string(skill),
				Name:         b.activityMap[string(skill)],
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
