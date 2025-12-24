package delta

import (
	"context"
	"errors"
	"time"

	"github.com/ctfloyd/hazelmere-api/src/internal/database"
	"github.com/ctfloyd/hazelmere-api/src/internal/foundation/monitor"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type DeltaRepository interface {
	InsertDelta(ctx context.Context, delta HiscoreDeltaData) (HiscoreDeltaData, error)
	GetDeltaById(ctx context.Context, id string) (HiscoreDeltaData, error)
	GetLatestDeltaForUser(ctx context.Context, userId string) (HiscoreDeltaData, error)
	GetDeltasInRange(ctx context.Context, userId string, startTime, endTime time.Time) ([]HiscoreDeltaData, error)
	GetAllDeltasForUser(ctx context.Context, userId string) ([]HiscoreDeltaData, error)
	CountDeltasForUser(ctx context.Context, userId string) (int64, error)
}

type mongoDeltaRepository struct {
	monitor    *monitor.Monitor
	collection *mongo.Collection
}

func NewDeltaRepository(deltaCollection *mongo.Collection, mon *monitor.Monitor) DeltaRepository {
	return &mongoDeltaRepository{
		collection: deltaCollection,
		monitor:    mon,
	}
}

func (dr *mongoDeltaRepository) InsertDelta(ctx context.Context, delta HiscoreDeltaData) (HiscoreDeltaData, error) {
	ctx, span := dr.monitor.StartSpan(ctx, "mongoDeltaRepository.InsertDelta")
	defer span.End()

	_, err := dr.collection.InsertOne(ctx, delta)
	if err != nil {
		return HiscoreDeltaData{}, errors.Join(database.ErrGeneric, err)
	}
	return delta, nil
}

func (dr *mongoDeltaRepository) GetDeltaById(ctx context.Context, id string) (HiscoreDeltaData, error) {
	ctx, span := dr.monitor.StartSpan(ctx, "mongoDeltaRepository.GetDeltaById")
	defer span.End()

	filter := bson.M{"_id": id}
	result := dr.collection.FindOne(ctx, filter)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return HiscoreDeltaData{}, errors.Join(database.ErrNotFound, result.Err())
		}
		return HiscoreDeltaData{}, errors.Join(database.ErrGeneric, result.Err())
	}

	var delta HiscoreDeltaData
	err := result.Decode(&delta)
	if err != nil {
		return HiscoreDeltaData{}, errors.Join(database.ErrGeneric, err)
	}
	return delta, nil
}

func (dr *mongoDeltaRepository) GetLatestDeltaForUser(ctx context.Context, userId string) (HiscoreDeltaData, error) {
	ctx, span := dr.monitor.StartSpan(ctx, "mongoDeltaRepository.GetLatestDeltaForUser")
	defer span.End()

	sort := options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: -1}})
	filter := bson.M{"userId": userId}

	result := dr.collection.FindOne(ctx, filter, sort)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return HiscoreDeltaData{}, errors.Join(database.ErrNotFound, result.Err())
		}
		return HiscoreDeltaData{}, errors.Join(database.ErrGeneric, result.Err())
	}

	var delta HiscoreDeltaData
	err := result.Decode(&delta)
	if err != nil {
		return HiscoreDeltaData{}, errors.Join(database.ErrGeneric, err)
	}
	return delta, nil
}

func (dr *mongoDeltaRepository) GetDeltasInRange(ctx context.Context, userId string, startTime, endTime time.Time) ([]HiscoreDeltaData, error) {
	ctx, span := dr.monitor.StartSpan(ctx, "mongoDeltaRepository.GetDeltasInRange")
	defer span.End()

	filter := bson.M{
		"userId": userId,
		"timestamp": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})
	cursor, err := dr.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, errors.Join(database.ErrGeneric, err)
	}

	var results []HiscoreDeltaData
	if err = cursor.All(ctx, &results); err != nil {
		return nil, errors.Join(database.ErrGeneric, err)
	}

	return results, nil
}

func (dr *mongoDeltaRepository) GetAllDeltasForUser(ctx context.Context, userId string) ([]HiscoreDeltaData, error) {
	ctx, span := dr.monitor.StartSpan(ctx, "mongoDeltaRepository.GetAllDeltasForUser")
	defer span.End()

	filter := bson.M{"userId": userId}
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})

	cursor, err := dr.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, errors.Join(database.ErrGeneric, err)
	}

	var results []HiscoreDeltaData
	if err = cursor.All(ctx, &results); err != nil {
		return nil, errors.Join(database.ErrGeneric, err)
	}

	return results, nil
}

func (dr *mongoDeltaRepository) CountDeltasForUser(ctx context.Context, userId string) (int64, error) {
	ctx, span := dr.monitor.StartSpan(ctx, "mongoDeltaRepository.CountDeltasForUser")
	defer span.End()

	filter := bson.M{"userId": userId}
	count, err := dr.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, errors.Join(database.ErrGeneric, err)
	}
	return count, nil
}
