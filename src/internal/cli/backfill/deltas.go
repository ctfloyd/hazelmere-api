package backfill

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ctfloyd/hazelmere-api/src/internal/initialize"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_config"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	numUserWorkers = 8
	writeBatchSize = 500
)

type snapshotData struct {
	Id         string         `bson:"_id"`
	UserId     string         `bson:"userId"`
	Timestamp  time.Time      `bson:"timestamp"`
	Skills     []skillData    `bson:"skills"`
	Bosses     []bossData     `bson:"bosses"`
	Activities []activityData `bson:"activities"`
}

type skillData struct {
	ActivityType string `bson:"activityType"`
	Name         string `bson:"name"`
	Level        int    `bson:"level"`
	Experience   int    `bson:"experience"`
	Rank         int    `bson:"rank"`
}

type bossData struct {
	ActivityType string `bson:"activityType"`
	Name         string `bson:"name"`
	KillCount    int    `bson:"killCount"`
	Rank         int    `bson:"rank"`
}

type activityData struct {
	ActivityType string `bson:"activityType"`
	Name         string `bson:"name"`
	Score        int    `bson:"score"`
	Rank         int    `bson:"rank"`
}

type deltaData struct {
	Id                  string              `bson:"_id"`
	UserId              string              `bson:"userId"`
	SnapshotId          string              `bson:"snapshotId"`
	PreviousSnapshotId  string              `bson:"previousSnapshotId"`
	Timestamp           time.Time           `bson:"timestamp"`
	Skills              []skillDeltaData    `bson:"skills,omitempty"`
	Bosses              []bossDeltaData     `bson:"bosses,omitempty"`
	Activities          []activityDeltaData `bson:"activities,omitempty"`
	TotalExperienceGain int                 `bson:"totalExperienceGain"`
}

type skillDeltaData struct {
	ActivityType   string `bson:"activityType"`
	Name           string `bson:"name"`
	ExperienceGain int    `bson:"experienceGain"`
	LevelGain      int    `bson:"levelGain"`
}

type bossDeltaData struct {
	ActivityType  string `bson:"activityType"`
	Name          string `bson:"name"`
	KillCountGain int    `bson:"killCountGain"`
}

type activityDeltaData struct {
	ActivityType string `bson:"activityType"`
	Name         string `bson:"name"`
	ScoreGain    int    `bson:"scoreGain"`
}

type userResult struct {
	userId        string
	snapshotCount int
	deltaCount    int
	skillChanges  int
	bossChanges   int
	actChanges    int
	totalXpGain   int
	err           error
}

func RunDeltas(configPath string, args []string) error {
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
	snapshotCollName := config.ValueOrPanic("mongo.database.collections.snapshot")
	deltaCollName := config.ValueOrPanic("mongo.database.collections.delta")

	snapshotCollection := client.Database(dbName).Collection(snapshotCollName)
	deltaCollection := client.Database(dbName).Collection(deltaCollName)

	fmt.Println("=== Delta Backfill Script ===")
	fmt.Printf("Database: %s\n", dbName)
	fmt.Printf("Snapshot Collection: %s\n", snapshotCollName)
	fmt.Printf("Delta Collection: %s\n", deltaCollName)
	fmt.Printf("Workers: %d\n", numUserWorkers)
	fmt.Printf("Batch Size: %d\n\n", writeBatchSize)

	return backfillDeltas(ctx, snapshotCollection, deltaCollection)
}

func backfillDeltas(ctx context.Context, snapshotCollection, deltaCollection *mongo.Collection) error {
	fmt.Println("Fetching distinct user IDs...")
	result := snapshotCollection.Distinct(ctx, "userId", bson.M{})
	var userIds []interface{}
	if err := result.Decode(&userIds); err != nil {
		return fmt.Errorf("failed to get distinct user IDs: %w", err)
	}

	totalUsers := len(userIds)
	fmt.Printf("Found %d users to process\n\n", totalUsers)

	userChan := make(chan string, numUserWorkers*2)
	resultChan := make(chan userResult, numUserWorkers*2)
	var wg sync.WaitGroup
	var processedUsers atomic.Int64
	var totalDeltas atomic.Int64
	var totalSkillChanges atomic.Int64
	var totalBossChanges atomic.Int64
	var totalActivityChanges atomic.Int64
	var totalXpGain atomic.Int64
	var errorCount atomic.Int64
	var skippedUsers atomic.Int64

	var allDeltasMu sync.Mutex
	var allDeltas []deltaData

	done := make(chan struct{})
	go func() {
		for res := range resultChan {
			if res.err != nil {
				fmt.Printf("  ERROR [%s]: %v\n", res.userId, res.err)
				errorCount.Add(1)
				continue
			}

			if res.deltaCount == 0 {
				skippedUsers.Add(1)
				if res.snapshotCount < 2 {
					fmt.Printf("  SKIP  [%s]: Only %d snapshot(s), need at least 2\n", res.userId, res.snapshotCount)
				} else {
					fmt.Printf("  SKIP  [%s]: %d snapshots but no changes detected\n", res.userId, res.snapshotCount)
				}
				continue
			}

			fmt.Printf("  OK    [%s]: %d snapshots -> %d deltas (skills: %d, bosses: %d, activities: %d, xp: +%d)\n",
				res.userId, res.snapshotCount, res.deltaCount,
				res.skillChanges, res.bossChanges, res.actChanges, res.totalXpGain)

			totalDeltas.Add(int64(res.deltaCount))
			totalSkillChanges.Add(int64(res.skillChanges))
			totalBossChanges.Add(int64(res.bossChanges))
			totalActivityChanges.Add(int64(res.actChanges))
			totalXpGain.Add(int64(res.totalXpGain))
		}
		close(done)
	}()

	for i := 0; i < numUserWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for userId := range userChan {
				deltas, snapshotCount, err := processUser(ctx, snapshotCollection, deltaCollection, userId)

				res := userResult{
					userId:        userId,
					snapshotCount: snapshotCount,
					err:           err,
				}

				if err == nil && len(deltas) > 0 {
					res.deltaCount = len(deltas)
					for _, d := range deltas {
						res.skillChanges += len(d.Skills)
						res.bossChanges += len(d.Bosses)
						res.actChanges += len(d.Activities)
						res.totalXpGain += d.TotalExperienceGain
					}

					allDeltasMu.Lock()
					allDeltas = append(allDeltas, deltas...)
					allDeltasMu.Unlock()
				}

				resultChan <- res

				processed := processedUsers.Add(1)
				if processed%10 == 0 {
					pct := float64(processed) / float64(totalUsers) * 100
					fmt.Printf("\n--- Progress: %d/%d users (%.1f%%) ---\n\n",
						processed, totalUsers, pct)
				}
			}
		}(i)
	}

	for _, uid := range userIds {
		userId, ok := uid.(string)
		if !ok {
			continue
		}
		userChan <- userId
	}
	close(userChan)

	wg.Wait()
	close(resultChan)
	<-done

	fmt.Printf("\n")
	fmt.Printf("=====================================\n")
	fmt.Printf("           BACKFILL COMPLETE         \n")
	fmt.Printf("=====================================\n")
	fmt.Printf("Users processed:      %d\n", processedUsers.Load())
	fmt.Printf("Users with deltas:    %d\n", processedUsers.Load()-skippedUsers.Load()-errorCount.Load())
	fmt.Printf("Users skipped:        %d\n", skippedUsers.Load())
	fmt.Printf("Errors:               %d\n", errorCount.Load())
	fmt.Printf("-------------------------------------\n")
	fmt.Printf("Total deltas created: %d\n", totalDeltas.Load())
	fmt.Printf("Total skill changes:  %d\n", totalSkillChanges.Load())
	fmt.Printf("Total boss changes:   %d\n", totalBossChanges.Load())
	fmt.Printf("Total activity changes: %d\n", totalActivityChanges.Load())
	fmt.Printf("Total XP gained:      %d\n", totalXpGain.Load())
	fmt.Printf("=====================================\n")

	return nil
}

func processUser(ctx context.Context, snapshotCollection, deltaCollection *mongo.Collection, userId string) ([]deltaData, int, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: 1}})

	cursor, err := snapshotCollection.Find(ctx, bson.M{"userId": userId}, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find snapshots: %w", err)
	}
	defer cursor.Close(ctx)

	var snapshots []snapshotData
	if err := cursor.All(ctx, &snapshots); err != nil {
		return nil, 0, fmt.Errorf("failed to decode snapshots: %w", err)
	}

	snapshotCount := len(snapshots)
	if snapshotCount < 2 {
		return nil, snapshotCount, nil
	}

	var deltas []deltaData
	for i := 1; i < len(snapshots); i++ {
		prev := snapshots[i-1]
		curr := snapshots[i]

		delta := computeDelta(prev, curr)

		if len(delta.Skills) > 0 || len(delta.Bosses) > 0 || len(delta.Activities) > 0 {
			deltas = append(deltas, delta)
		}
	}

	if len(deltas) == 0 {
		return nil, snapshotCount, nil
	}

	for i := 0; i < len(deltas); i += writeBatchSize {
		end := i + writeBatchSize
		if end > len(deltas) {
			end = len(deltas)
		}
		batch := deltas[i:end]

		docs := make([]interface{}, len(batch))
		for j, d := range batch {
			docs[j] = d
		}

		_, err := deltaCollection.InsertMany(ctx, docs, options.InsertMany().SetOrdered(false))
		if err != nil {
			if !mongo.IsDuplicateKeyError(err) {
				return nil, snapshotCount, fmt.Errorf("failed to insert deltas: %w", err)
			}
		}
	}

	return deltas, snapshotCount, nil
}

func computeDelta(prev, curr snapshotData) deltaData {
	d := deltaData{
		Id:                 uuid.New().String(),
		UserId:             curr.UserId,
		SnapshotId:         curr.Id,
		PreviousSnapshotId: prev.Id,
		Timestamp:          curr.Timestamp,
	}

	d.Skills = computeSkillDeltas(prev.Skills, curr.Skills)
	d.Bosses = computeBossDeltas(prev.Bosses, curr.Bosses)
	d.Activities = computeActivityDeltas(prev.Activities, curr.Activities)
	d.TotalExperienceGain = computeTotalExperienceGain(d.Skills)

	return d
}

func computeSkillDeltas(prev, curr []skillData) []skillDeltaData {
	var deltas []skillDeltaData

	prevMap := make(map[string]skillData)
	for _, skill := range prev {
		if skill.ActivityType != "" {
			prevMap[skill.ActivityType] = skill
		}
	}

	for _, c := range curr {
		if c.ActivityType == "" {
			continue
		}

		p, exists := prevMap[c.ActivityType]
		if !exists {
			continue
		}

		if c.Experience < 0 || p.Experience < 0 {
			continue
		}

		xpGain := c.Experience - p.Experience
		levelGain := c.Level - p.Level

		if xpGain > 0 || levelGain > 0 {
			deltas = append(deltas, skillDeltaData{
				ActivityType:   c.ActivityType,
				Name:           c.Name,
				ExperienceGain: xpGain,
				LevelGain:      levelGain,
			})
		}
	}

	return deltas
}

func computeBossDeltas(prev, curr []bossData) []bossDeltaData {
	var deltas []bossDeltaData

	prevMap := make(map[string]bossData)
	for _, boss := range prev {
		if boss.ActivityType != "" {
			prevMap[boss.ActivityType] = boss
		}
	}

	for _, c := range curr {
		if c.ActivityType == "" {
			continue
		}

		p, exists := prevMap[c.ActivityType]
		if !exists {
			continue
		}

		if c.KillCount < 0 || p.KillCount < 0 {
			continue
		}

		kcGain := c.KillCount - p.KillCount

		if kcGain > 0 {
			deltas = append(deltas, bossDeltaData{
				ActivityType:  c.ActivityType,
				Name:          c.Name,
				KillCountGain: kcGain,
			})
		}
	}

	return deltas
}

func computeActivityDeltas(prev, curr []activityData) []activityDeltaData {
	var deltas []activityDeltaData

	prevMap := make(map[string]activityData)
	for _, activity := range prev {
		if activity.ActivityType != "" {
			prevMap[activity.ActivityType] = activity
		}
	}

	for _, c := range curr {
		if c.ActivityType == "" {
			continue
		}

		p, exists := prevMap[c.ActivityType]
		if !exists {
			continue
		}

		if c.Score < 0 || p.Score < 0 {
			continue
		}

		scoreGain := c.Score - p.Score

		if scoreGain > 0 {
			deltas = append(deltas, activityDeltaData{
				ActivityType: c.ActivityType,
				Name:         c.Name,
				ScoreGain:    scoreGain,
			})
		}
	}

	return deltas
}

func computeTotalExperienceGain(skills []skillDeltaData) int {
	total := 0
	for _, skill := range skills {
		if skill.ExperienceGain > 0 {
			total += skill.ExperienceGain
		}
	}
	return total
}
