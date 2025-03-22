package initialize

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func MongoClient(uri string, username string, password string) (*mongo.Client, error) {
	credentials := options.Credential{Username: username, Password: password}
	return mongo.Connect(options.Client().ApplyURI(uri).SetAuth(credentials))
}

func MongoCleanup(ctx context.Context, client *mongo.Client) {
	if client == nil {
		return
	}

	if err := client.Disconnect(ctx); err != nil {
		panic(err)
	}
}
