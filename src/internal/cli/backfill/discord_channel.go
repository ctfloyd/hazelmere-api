package backfill

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/ctfloyd/hazelmere-api/src/internal/initialize"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_config"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// RunDiscordChannel backfills the discordChannelId field on existing user
// documents. Any user that predates the field (missing it, or holding an empty
// value) is assigned the default channel so their updates continue to be posted
// somewhere.
func RunDiscordChannel(configPath string, args []string) error {
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
	userCollName := config.ValueOrPanic("mongo.database.collections.user")
	userCollection := client.Database(dbName).Collection(userCollName)

	fmt.Println("=== Discord Channel Backfill Script ===")
	fmt.Printf("Database: %s\n", dbName)
	fmt.Printf("User Collection: %s\n", userCollName)
	fmt.Printf("Default Channel: %s\n\n", api.DefaultDiscordChannelId)

	// Match users that either have no discordChannelId field at all or have it
	// set to an empty string.
	filter := bson.M{
		"$or": bson.A{
			bson.M{"discordChannelId": bson.M{"$exists": false}},
			bson.M{"discordChannelId": ""},
		},
	}
	update := bson.M{"$set": bson.M{"discordChannelId": api.DefaultDiscordChannelId}}

	toUpdate, err := userCollection.CountDocuments(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to count users needing backfill: %w", err)
	}
	fmt.Printf("Users needing backfill: %d\n", toUpdate)

	result, err := userCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to backfill discord channel: %w", err)
	}

	fmt.Printf("\n")
	fmt.Printf("=====================================\n")
	fmt.Printf("           BACKFILL COMPLETE         \n")
	fmt.Printf("=====================================\n")
	fmt.Printf("Users matched:   %d\n", result.MatchedCount)
	fmt.Printf("Users modified:  %d\n", result.ModifiedCount)
	fmt.Printf("Assigned channel: %s\n", api.DefaultDiscordChannelId)
	fmt.Printf("=====================================\n")

	return nil
}
