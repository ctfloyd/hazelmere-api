package snapshot

import (
	"context"
	"errors"
	"github.com/ctfloyd/hazelmere-api/src/internal/database"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/sync/errgroup"
	"time"
)

type SnapshotIntervalResult struct {
	Snapshots          []HiscoreSnapshotData
	TotalSnapshots     int
	SnapshotsWithGains int
}

type SnapshotRepository interface {
	GetSnapshotById(ctx context.Context, id string) (HiscoreSnapshotData, error)
	GetLatestSnapshotForUser(ctx context.Context, userId string) (HiscoreSnapshotData, error)
	GetSnapshotInterval(ctx context.Context, userId string, startTime time.Time, endTime time.Time, aggregationWindow api.AggregationWindow) (SnapshotIntervalResult, error)
	GetAllSnapshotsForUser(ctx context.Context, userId string) ([]HiscoreSnapshotData, error)
	InsertSnapshot(ctx context.Context, snapshot HiscoreSnapshotData) (HiscoreSnapshotData, error)
	GetSnapshotForUserNearestTimestamp(ctx context.Context, userId string, timestamp time.Time) (HiscoreSnapshotData, error)
}

type mongoSnapshotRepository struct {
	logger     hz_logger.Logger
	collection *mongo.Collection
}

func NewSnapshotRepository(snapshotCollection *mongo.Collection, logger hz_logger.Logger) SnapshotRepository {
	return &mongoSnapshotRepository{
		collection: snapshotCollection,
		logger:     logger,
	}
}

func (sr *mongoSnapshotRepository) GetSnapshotInterval(ctx context.Context, userId string, startTime time.Time, endTime time.Time, aggregationWindow api.AggregationWindow) (SnapshotIntervalResult, error) {
	dateFormat := getDateFormatForAggregationWindow(aggregationWindow)

	baseFilter := bson.M{
		"userId": userId,
		"timestamp": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}

	group, ctx := errgroup.WithContext(ctx)

	var snapshots []HiscoreSnapshotData
	var totalCount int64
	var gainsCount int64

	// Run aggregation pipeline for snapshots
	group.Go(func() error {
		pipeline := []bson.M{
			{
				"$match": bson.M{
					"userId": userId,
					"timestamp": bson.M{
						"$gte": startTime,
						"$lte": endTime,
					},
					"overallExperienceChange": bson.M{
						"$ne": 0,
					},
				},
			},
			{
				"$addFields": bson.M{
					"dateOnly": bson.M{
						"$dateToString": bson.M{
							"format": dateFormat,
							"date":   "$timestamp",
						},
					},
					"overallExperience": bson.M{
						"$let": bson.M{
							"vars": bson.M{
								"overallSkill": bson.M{
									"$arrayElemAt": bson.A{
										bson.M{
											"$filter": bson.M{
												"input": "$skills",
												"cond":  bson.M{"$eq": bson.A{"$$this.activityType", "OVERALL"}},
											},
										},
										0,
									},
								},
							},
							"in": "$$overallSkill.experience",
						},
					},
				},
			},
			{
				"$sort": bson.D{
					{Key: "dateOnly", Value: 1},
					{Key: "overallExperience", Value: -1},
					{Key: "timestamp", Value: -1},
				},
			},
			{
				"$group": bson.M{
					"_id": "$dateOnly",
					"snapshot": bson.M{
						"$first": "$$ROOT",
					},
				},
			},
			{
				"$replaceRoot": bson.M{
					"newRoot": "$snapshot",
				},
			},
			{
				"$project": bson.M{
					"dateOnly":          0,
					"overallExperience": 0,
				},
			},
			{
				"$sort": bson.M{"timestamp": 1},
			},
		}

		cursor, err := sr.collection.Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}

		return cursor.All(ctx, &snapshots)
	})

	// Count total snapshots in interval
	group.Go(func() error {
		count, err := sr.collection.CountDocuments(ctx, baseFilter)
		if err != nil {
			return err
		}
		totalCount = count
		return nil
	})

	// Count snapshots with gains (positive experience change)
	group.Go(func() error {
		gainsFilter := bson.M{
			"userId": userId,
			"timestamp": bson.M{
				"$gte": startTime,
				"$lte": endTime,
			},
			"overallExperienceChange": bson.M{
				"$gt": 0,
			},
		}
		count, err := sr.collection.CountDocuments(ctx, gainsFilter)
		if err != nil {
			return err
		}
		gainsCount = count
		return nil
	})

	if err := group.Wait(); err != nil {
		return SnapshotIntervalResult{}, errors.Join(database.ErrGeneric, err)
	}

	return SnapshotIntervalResult{
		Snapshots:          snapshots,
		TotalSnapshots:     int(totalCount),
		SnapshotsWithGains: int(gainsCount),
	}, nil
}

func getDateFormatForAggregationWindow(window api.AggregationWindow) string {
	switch window {
	case api.AggregationWindowWeekly:
		return "%G-W%V"
	case api.AggregationWindowMonthly:
		return "%Y-%m"
	default:
		return "%Y-%m-%d"
	}
}

func (sr *mongoSnapshotRepository) InsertSnapshot(ctx context.Context, snapshot HiscoreSnapshotData) (HiscoreSnapshotData, error) {
	_, err := sr.collection.InsertOne(ctx, snapshot)
	if err != nil {
		return HiscoreSnapshotData{}, errors.Join(database.ErrGeneric, err)
	}
	return snapshot, nil
}

func (sr *mongoSnapshotRepository) GetSnapshotById(ctx context.Context, id string) (HiscoreSnapshotData, error) {
	filter := bson.M{"_id": id}
	result := sr.collection.FindOne(ctx, filter)
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
		if err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				return err
			}
		}
		lessThan = result
		return nil
	})
	group.Go(func() error {
		result, err := sr.getSnapshotForUserNearestTimestampGreaterThan(ctx, userId, timestamp)
		if err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				return err
			}
		}
		greaterThan = result
		return nil
	})
	if err := group.Wait(); err != nil {
		return HiscoreSnapshotData{}, errors.Join(database.ErrGeneric, err)
	}

	if lessThan.Timestamp.IsZero() && greaterThan.Timestamp.IsZero() {
		return HiscoreSnapshotData{}, database.ErrNotFound
	}

	if lessThan.Timestamp.IsZero() {
		return greaterThan, nil
	}

	if greaterThan.Timestamp.IsZero() {
		return lessThan, nil
	}

	ltDiff := timestamp.Sub(lessThan.Timestamp)
	gtDiff := greaterThan.Timestamp.Sub(timestamp)

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
