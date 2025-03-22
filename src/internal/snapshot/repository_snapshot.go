package snapshot

import (
	"api/src/internal/common/database"
	"api/src/internal/common/logger"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type SnapshotRepository interface {
	GetLatestSnapshotForUser(ctx context.Context, userId string) (HiscoreSnapshotData, error)
	GetAllSnapshotsForUser(ctx context.Context, userId string) ([]HiscoreSnapshotData, error)
	InsertSnapshot(ctx context.Context, snapshot HiscoreSnapshotData) (HiscoreSnapshotData, error)
}

type mongoSnapshotRepository struct {
	logger     logger.Logger
	collection *mongo.Collection
}

func NewSnapshotRepository(snapshotCollection *mongo.Collection, logger logger.Logger) SnapshotRepository {
	return &mongoSnapshotRepository{
		collection: snapshotCollection,
		logger:     logger,
	}
}

func (sr *mongoSnapshotRepository) InsertSnapshot(ctx context.Context, snapshot HiscoreSnapshotData) (HiscoreSnapshotData, error) {
	_, err := sr.collection.InsertOne(ctx, snapshot)
	if err != nil {
		return HiscoreSnapshotData{}, errors.Join(database.ErrGeneric, err)
	}
	return snapshot, nil
}

func (sr *mongoSnapshotRepository) GetLatestSnapshotForUser(ctx context.Context, userId string) (HiscoreSnapshotData, error) {
	sort := options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: -1}})
	filter := bson.M{"userId": userId}

	result := sr.collection.FindOne(ctx, filter, sort)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return HiscoreSnapshotData{}, errors.Join(database.ErrNotFound, result.Err())
		}

		return HiscoreSnapshotData{}, errors.Join(database.ErrGeneric, result.Err())
	}

	var snapshot HiscoreSnapshotData
	err := result.Decode(&snapshot)
	if err != nil {
		return HiscoreSnapshotData{}, errors.Join(database.ErrGeneric, result.Err())
	}
	return snapshot, nil
}

func (sr *mongoSnapshotRepository) GetAllSnapshotsForUser(ctx context.Context, userId string) ([]HiscoreSnapshotData, error) {
	filter := bson.M{"userId": userId}

	cursor, err := sr.collection.Find(ctx, filter)
	if err != nil {
		return []HiscoreSnapshotData{}, errors.Join(database.ErrGeneric, err)
	}

	var results []HiscoreSnapshotData
	if err = cursor.All(ctx, &results); err != nil {
		return []HiscoreSnapshotData{}, errors.Join(database.ErrGeneric, err)
	}

	return results, nil
}
