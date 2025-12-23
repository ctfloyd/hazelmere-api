package database

import "go.mongodb.org/mongo-driver/v2/mongo"

type MongoFactoryConfig struct {
	DatabaseName           string
	SnapshotCollectionName string
	UserCollectionName     string
	DeltaCollectionName    string
}

type MongoFactory struct {
	client *mongo.Client
	config MongoFactoryConfig
}

func NewMongoFactory(client *mongo.Client, config MongoFactoryConfig) *MongoFactory {
	return &MongoFactory{
		client: client,
		config: config,
	}
}

func (mf *MongoFactory) NewSnapshotCollection() *mongo.Collection {
	return mf.client.Database(mf.config.DatabaseName).Collection(mf.config.SnapshotCollectionName)
}

func (mf *MongoFactory) NewUserCollection() *mongo.Collection {
	return mf.client.Database(mf.config.DatabaseName).Collection(mf.config.UserCollectionName)
}

func (mf *MongoFactory) NewDeltaCollection() *mongo.Collection {
	return mf.client.Database(mf.config.DatabaseName).Collection(mf.config.DeltaCollectionName)
}
