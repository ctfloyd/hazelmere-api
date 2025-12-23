package main

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
	Timestamp               time.Time   `bson:"timestamp"`
	Source                  string      `bson:"source"`
	OverallExperienceChange int         `bson:"overallExperienceChange"`
	Skills                  []skillData `bson:"skills"`
}

type skillData struct {
	ActivityType string `bson:"activityType"`
	Experience   int    `bson:"experience"`
}

type writeOp struct {
	id            string
	correctChange int
	currentChange int
	shouldDelete  bool
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

	fmt.Println("=== Fix Snapshot XP Change Script ===")
	fmt.Printf("Database: %s\n", dbName)
	fmt.Printf("Collection: %s\n", collName)
	fmt.Printf("Workers: %d\n", numWriteWorkers)
	fmt.Printf("Batch Size: %d\n\n", writeBatchSize)

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
	updateChan := make(chan []writeOp, numWriteWorkers*2)
	errChan := make(chan error, numWriteWorkers)
	var writeWg sync.WaitGroup
	var writtenCount atomic.Int64
	var deletedCount atomic.Int64

	for i := 0; i < numWriteWorkers; i++ {
		writeWg.Add(1)
		go func() {
			defer writeWg.Done()
			for batch := range updateChan {
				updated, deleted, err := writeBatch(ctx, collection, batch)
				if err != nil {
					errChan <- err
					return
				}
				writtenCount.Add(int64(updated))
				deletedCount.Add(int64(deleted))
			}
		}()
	}

	// Fetch sorted by userId, timestamp - this ensures we process each user's snapshots in order
	// Don't use $elemMatch - fetch all skills and filter in code
	opts := options.Find().
		SetSort(bson.D{{Key: "userId", Value: 1}, {Key: "timestamp", Value: 1}}).
		SetBatchSize(int32(fetchBatchSize)).
		SetProjection(bson.M{
			"_id":                     1,
			"userId":                  1,
			"timestamp":               1,
			"source":                  1,
			"overallExperienceChange": 1,
			"skills.activityType":     1,
			"skills.experience":       1,
		})

	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return fmt.Errorf("failed to find snapshots: %w", err)
	}
	defer cursor.Close(ctx)

	// State tracking per user
	type userState struct {
		lastExp       int
		lastTimestamp time.Time
		lastId        string
	}
	prevUserState := make(map[string]*userState)

	var pendingOps []writeOp
	var processedCount int64
	var totalUpdates int64
	var totalDeletes int64
	var totalCorrect int64
	var totalNoOverall int64
	var sampleUpdates []writeOp
	var noOverallIds []string

	fmt.Println("Processing snapshots...")

	for cursor.Next(ctx) {
		var s snapshotData
		if err := cursor.Decode(&s); err != nil {
			return fmt.Errorf("failed to decode snapshot: %w", err)
		}

		currentExp := getOverallExp(s.Skills)
		if currentExp == -1 {
			// No OVERALL skill found - skip but count and record ID
			totalNoOverall++
			noOverallIds = append(noOverallIds, s.Id)
			processedCount++
			continue
		}

		correctChange := 0
		state := prevUserState[s.UserId]

		if state != nil {
			diff := currentExp - state.lastExp
			if diff > 0 {
				correctChange = diff
			} else if diff < 0 && s.Source == "WOM_BACKFILL_122025" {
				// XP went down and source is WOM_BACKFILL - delete this invalid snapshot
				pendingOps = append(pendingOps, writeOp{
					id:           s.Id,
					shouldDelete: true,
				})
				totalDeletes++
				// Don't update state - skip this snapshot entirely
				goto checkBatch
			}
			// If diff <= 0 and not WOM_BACKFILL, correctChange stays 0
		}

		if s.OverallExperienceChange != correctChange {
			op := writeOp{
				id:            s.Id,
				correctChange: correctChange,
				currentChange: s.OverallExperienceChange,
			}
			pendingOps = append(pendingOps, op)
			totalUpdates++

			// Capture first 10 sample updates for output
			if len(sampleUpdates) < 10 {
				sampleUpdates = append(sampleUpdates, op)
			}
		} else {
			totalCorrect++
		}

		// Update state for this user
		if state == nil {
			prevUserState[s.UserId] = &userState{
				lastExp:       currentExp,
				lastTimestamp: s.Timestamp,
				lastId:        s.Id,
			}
		} else {
			state.lastExp = currentExp
			state.lastTimestamp = s.Timestamp
			state.lastId = s.Id
		}

	checkBatch:
		// Send batch to writers when full
		if len(pendingOps) >= writeBatchSize {
			updateChan <- pendingOps
			pendingOps = nil
		}
		processedCount++

		if processedCount%5000 == 0 {
			pct := float64(processedCount) / float64(totalCount) * 100
			fmt.Printf("Progress: %d/%d (%.1f%%) | Correct: %d | Updates: %d (written: %d) | Deletes: %d (deleted: %d) | NoOverall: %d\n",
				processedCount, totalCount, pct, totalCorrect, totalUpdates, writtenCount.Load(), totalDeletes, deletedCount.Load(), totalNoOverall)
		}
	}

	if err := cursor.Err(); err != nil {
		return fmt.Errorf("cursor error: %w", err)
	}

	// Send remaining operations
	if len(pendingOps) > 0 {
		updateChan <- pendingOps
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

	fmt.Printf("\n")
	fmt.Printf("=====================================\n")
	fmt.Printf("              COMPLETE               \n")
	fmt.Printf("=====================================\n")
	fmt.Printf("Total processed:     %d\n", processedCount)
	fmt.Printf("Already correct:     %d\n", totalCorrect)
	fmt.Printf("Updated:             %d\n", writtenCount.Load())
	fmt.Printf("Deleted:             %d\n", deletedCount.Load())
	fmt.Printf("No OVERALL skill:    %d\n", totalNoOverall)
	fmt.Printf("Users tracked:       %d\n", len(prevUserState))
	fmt.Printf("=====================================\n")

	// Show sample updates
	if len(sampleUpdates) > 0 {
		fmt.Printf("\n--- Sample Updates (first %d) ---\n", len(sampleUpdates))
		for i, op := range sampleUpdates {
			fmt.Printf("%d. ID: %s\n", i+1, op.id)
			fmt.Printf("   Current value: %d -> Correct value: %d\n", op.currentChange, op.correctChange)
		}
		fmt.Printf("---------------------------------\n")
	}

	// Show IDs of snapshots missing OVERALL skill
	if len(noOverallIds) > 0 {
		fmt.Printf("\n--- Snapshots Missing OVERALL Skill (%d) ---\n", len(noOverallIds))
		for _, id := range noOverallIds {
			fmt.Println(id)
		}
		fmt.Printf("--------------------------------------------\n")
	}

	return nil
}

func getOverallExp(skills []skillData) int {
	for _, skill := range skills {
		if skill.ActivityType == "OVERALL" {
			return skill.Experience
		}
	}
	return -1 // Return -1 to indicate not found (0 could be valid)
}

func writeBatch(ctx context.Context, collection *mongo.Collection, ops []writeOp) (updated int, deleted int, err error) {
	var writeOps []mongo.WriteModel
	for _, op := range ops {
		if op.shouldDelete {
			writeOps = append(writeOps, mongo.NewDeleteOneModel().SetFilter(bson.M{"_id": op.id}))
			deleted++
		} else {
			writeOps = append(writeOps, mongo.NewUpdateOneModel().
				SetFilter(bson.M{"_id": op.id}).
				SetUpdate(bson.M{"$set": bson.M{"overallExperienceChange": op.correctChange}}))
			updated++
		}
	}

	if len(writeOps) == 0 {
		return 0, 0, nil
	}

	if _, err := collection.BulkWrite(ctx, writeOps); err != nil {
		return 0, 0, fmt.Errorf("bulk write failed: %w", err)
	}
	return updated, deleted, nil
}
