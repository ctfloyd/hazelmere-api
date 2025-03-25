package snapshot

import (
	"context"
	"errors"
	"github.com/ctfloyd/hazelmere-api/src/internal/common/database"
	"github.com/ctfloyd/hazelmere-api/src/internal/common/logger"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/sync/errgroup"
	"time"
)

type SnapshotRepository interface {
	GetLatestSnapshotForUser(ctx context.Context, userId string) (HiscoreSnapshotData, error)
	GetAllSnapshotsForUser(ctx context.Context, userId string) ([]HiscoreSnapshotData, error)
	InsertSnapshot(ctx context.Context, snapshot HiscoreSnapshotData) (HiscoreSnapshotData, error)
	GetSnapshotForUserNearestTimestamp(ctx context.Context, userId string, timestamp time.Time) (HiscoreSnapshotData, error)
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

func (sr *mongoSnapshotRepository) GetSnapshotForUserNearestTimestamp(ctx context.Context, userId string, timestamp time.Time) (HiscoreSnapshotData, error) {
	group, ctx := errgroup.WithContext(ctx)

	var lessThan HiscoreSnapshotData
	var greaterThan HiscoreSnapshotData
	group.Go(func() error {
		result, err := sr.getSnapshotForUserNearestTimestampLessThan(ctx, userId, timestamp)
		if err == nil {
			lessThan = result
		}
		return err
	})
	group.Go(func() error {
		result, err := sr.getSnapshotForUserNearestTimestampGreaterThan(ctx, userId, timestamp)
		if err == nil {
			greaterThan = result
		}
		return err
	})
	if err := group.Wait(); err != nil {
		return HiscoreSnapshotData{}, errors.Join(database.ErrGeneric, err)
	}

	ltDiff := timestamp.Sub(lessThan.Timestamp)
	gtDiff := greaterThan.Timestamp.Sub(lessThan.Timestamp)

	if ltDiff < gtDiff {
		return lessThan, nil
	}

	return greaterThan, nil
}

func (sr *mongoSnapshotRepository) getSnapshotForUserNearestTimestampLessThan(ctx context.Context, userId string, timestamp time.Time) (HiscoreSnapshotData, error) {
	sort := options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: -1}})
	filter := bson.M{"userId": userId, "timestamp": bson.M{"$lte": timestamp}}

	result := sr.collection.FindOne(ctx, filter, sort)
	if result.Err() != nil {
		return HiscoreSnapshotData{}, errors.Join(database.ErrGeneric, result.Err())
	}

	var hs HiscoreSnapshotData
	err := result.Decode(&hs)
	if err != nil {
		return HiscoreSnapshotData{}, errors.Join(database.ErrGeneric, err)
	}

	return hs, nil
}

func (sr *mongoSnapshotRepository) getSnapshotForUserNearestTimestampGreaterThan(ctx context.Context, userId string, timestamp time.Time) (HiscoreSnapshotData, error) {
	sort := options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: 1}})
	filter := bson.M{"userId": userId, "timestamp": bson.M{"$gte": timestamp}}

	result := sr.collection.FindOne(ctx, filter, sort)
	if result.Err() != nil {
		return HiscoreSnapshotData{}, errors.Join(database.ErrGeneric, result.Err())
	}

	var hs HiscoreSnapshotData
	err := result.Decode(&hs)
	if err != nil {
		return HiscoreSnapshotData{}, errors.Join(database.ErrGeneric, err)
	}

	return hs, nil
}
