package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"

	"github.com/ctfloyd/hazelmere-api/src/internal/initialize"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_config"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	numWriteWorkers = 16
	fetchBatchSize  = 1000
	writeBatchSize  = 500
)

type snapshotData struct {
	Id                      string      `bson:"_id"`
	UserId                  string      `bson:"userId"`
	OverallExperienceChange int         `bson:"overallExperienceChange"`
	Skills                  []skillData `bson:"skills"`
}

type skillData struct {
	ActivityType string `bson:"activityType"`
	Experience   int    `bson:"experience"`
}

type updateOp struct {
	id            string
	correctChange int
}

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
	collName := config.ValueOrPanic("mongo.database.collections.snapshot")
	collection := client.Database(dbName).Collection(collName)

	if err := fixSnapshotExperienceChanges(ctx, collection); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func fixSnapshotExperienceChanges(ctx context.Context, collection *mongo.Collection) error {
	totalCount, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to count documents: %w", err)
	}
	fmt.Printf("Total snapshots: %d\n\n", totalCount)

	// Start write workers
	updateChan := make(chan []updateOp, numWriteWorkers*2)
	errChan := make(chan error, numWriteWorkers)
	var writeWg sync.WaitGroup
	var writtenCount atomic.Int64

	for i := 0; i < numWriteWorkers; i++ {
		writeWg.Add(1)
		go func() {
			defer writeWg.Done()
			for batch := range updateChan {
				if err := writeBatch(ctx, collection, batch); err != nil {
					errChan <- err
					return
				}
				writtenCount.Add(int64(len(batch)))
			}
		}()
	}

	// Fetch sorted by userId, timestamp - this ensures we process each user's snapshots in order
	opts := options.Find().
		SetSort(bson.D{{Key: "userId", Value: 1}, {Key: "timestamp", Value: 1}}).
		SetBatchSize(int32(fetchBatchSize)).
		SetProjection(bson.M{
			"_id":                     1,
			"userId":                  1,
			"overallExperienceChange": 1,
			"skills": bson.M{
				"$elemMatch": bson.M{"activityType": "OVERALL"},
			},
		})

	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return fmt.Errorf("failed to find snapshots: %w", err)
	}
	defer cursor.Close(ctx)

	// State tracking
	prevUserExp := make(map[string]int) // userId -> last known experience
	var pendingUpdates []updateOp
	var processedCount int64
	var totalUpdates int64

	for cursor.Next(ctx) {
		var s snapshotData
		if err := cursor.Decode(&s); err != nil {
			return fmt.Errorf("failed to decode snapshot: %w", err)
		}

		currentExp := getOverallExp(s.Skills)
		correctChange := 0

		if prevExp, exists := prevUserExp[s.UserId]; exists {
			diff := currentExp - prevExp
			if diff > 0 {
				correctChange = diff
			}
		}

		if s.OverallExperienceChange != correctChange {
			pendingUpdates = append(pendingUpdates, updateOp{
				id:            s.Id,
				correctChange: correctChange,
			})
			totalUpdates++
		}

		// Send batch to writers when full
		if len(pendingUpdates) >= writeBatchSize {
			updateChan <- pendingUpdates
			pendingUpdates = nil
		}

		prevUserExp[s.UserId] = currentExp
		processedCount++

		if processedCount%5000 == 0 {
			pct := float64(processedCount) / float64(totalCount) * 100
			fmt.Printf("Progress: %d/%d (%.1f%%) - Updates queued: %d - Written: %d\n",
				processedCount, totalCount, pct, totalUpdates, writtenCount.Load())
		}
	}

	if err := cursor.Err(); err != nil {
		return fmt.Errorf("cursor error: %w", err)
	}

	// Send remaining updates
	if len(pendingUpdates) > 0 {
		updateChan <- pendingUpdates
	}
	close(updateChan)

	// Wait for writers to finish
	writeWg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	fmt.Printf("\nComplete! Processed: %d, Updated: %d\n", processedCount, writtenCount.Load())
	return nil
}

func getOverallExp(skills []skillData) int {
	for _, skill := range skills {
		if skill.ActivityType == "OVERALL" {
			return skill.Experience
		}
	}
	return 0
}

func writeBatch(ctx context.Context, collection *mongo.Collection, updates []updateOp) error {
	var ops []mongo.WriteModel
	for _, u := range updates {
		op := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": u.id}).
			SetUpdate(bson.M{"$set": bson.M{"overallExperienceChange": u.correctChange}})
		ops = append(ops, op)
	}

	if _, err := collection.BulkWrite(ctx, ops); err != nil {
		return fmt.Errorf("bulk write failed: %w", err)
	}
	return nil
}
