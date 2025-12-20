package snapshot

import (
	"context"
	"errors"
	"github.com/ctfloyd/hazelmere-api/src/internal/database"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"github.com/google/uuid"
	"time"
)

const MaxIntervalDuration = 5 * 365 * 2 * 24 * time.Hour

var ErrSnapshotGeneric = errors.New("an unexpected error occurred while performing snapshot operation")
var ErrSnapshotValidation = errors.New("snapshot is invalid")
var ErrSnapshotNotFound = errors.New("snapshot not found")
var ErrInvalidIntervalRequest = errors.New("invalid interval request")

type SnapshotIntervalResponse struct {
	Snapshots          []HiscoreSnapshot
	TotalSnapshots     int
	SnapshotsWithGains int
}

type SnapshotService interface {
	CreateSnapshot(ctx context.Context, snapshot HiscoreSnapshot) (HiscoreSnapshot, error)
	GetSnapshotById(ctx context.Context, id string) (HiscoreSnapshot, error)
	GetSnapshotInterval(ctx context.Context, userId string, startTime time.Time, endTime time.Time, aggregationWindow api.AggregationWindow) (SnapshotIntervalResponse, error)
	GetAllSnapshotsForUser(ctx context.Context, userId string) ([]HiscoreSnapshot, error)
	GetSnapshotForUserNearestTimestamp(ctx context.Context, userId string, timestamp int64) (HiscoreSnapshot, error)
}

type snapshotService struct {
	logger     hz_logger.Logger
	validator  SnapshotValidator
	repository SnapshotRepository
}

func NewSnapshotService(logger hz_logger.Logger, repository SnapshotRepository, validator SnapshotValidator) SnapshotService {
	return &snapshotService{
		logger:     logger,
		validator:  validator,
		repository: repository,
	}
}

func (ss *snapshotService) GetSnapshotInterval(ctx context.Context, userId string, startTime time.Time, endTime time.Time, aggregationWindow api.AggregationWindow) (SnapshotIntervalResponse, error) {
	startTime, endTime, err := validateSnapshotInterval(startTime, endTime)
	if err != nil {
		return SnapshotIntervalResponse{}, err
	}

	aggregationWindow = normalizeAggregationWindow(aggregationWindow)

	result, err := ss.repository.GetSnapshotInterval(ctx, userId, startTime, endTime, aggregationWindow)
	if err != nil {
		return SnapshotIntervalResponse{}, errors.Join(ErrSnapshotGeneric, err)
	}

	return SnapshotIntervalResponse{
		Snapshots:          HiscoreSnapshot{}.ManyFromData(result.Snapshots),
		TotalSnapshots:     result.TotalSnapshots,
		SnapshotsWithGains: result.SnapshotsWithGains,
	}, nil
}

func normalizeAggregationWindow(window api.AggregationWindow) api.AggregationWindow {
	switch window {
	case api.AggregationWindowDaily, api.AggregationWindowWeekly, api.AggregationWindowMonthly:
		return window
	default:
		return api.AggregationWindowDaily
	}
}

func (ss *snapshotService) GetAllSnapshotsForUser(ctx context.Context, userId string) ([]HiscoreSnapshot, error) {
	data, err := ss.repository.GetAllSnapshotsForUser(ctx, userId)
	if err != nil {
		return nil, errors.Join(ErrSnapshotGeneric, err)
	}
	return HiscoreSnapshot{}.ManyFromData(data), nil
}

func (ss *snapshotService) GetSnapshotById(ctx context.Context, id string) (HiscoreSnapshot, error) {
	data, err := ss.repository.GetSnapshotById(ctx, id)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return HiscoreSnapshot{}, ErrSnapshotNotFound
		}
		return HiscoreSnapshot{}, errors.Join(ErrSnapshotGeneric, err)
	}
	return HiscoreSnapshot{}.FromData(data), nil
}

func (ss *snapshotService) GetSnapshotForUserNearestTimestamp(ctx context.Context, userId string, timestamp int64) (HiscoreSnapshot, error) {
	date := time.Unix(0, timestamp*int64(time.Millisecond))

	data, err := ss.repository.GetSnapshotForUserNearestTimestamp(ctx, userId, date)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return HiscoreSnapshot{}, ErrSnapshotNotFound
		}

		return HiscoreSnapshot{}, errors.Join(ErrSnapshotGeneric, err)
	}

	return HiscoreSnapshot{}.FromData(data), nil
}

func (ss *snapshotService) CreateSnapshot(ctx context.Context, snapshot HiscoreSnapshot) (HiscoreSnapshot, error) {
	snapshot.Id = uuid.New().String()

	err := ss.validator.ValidateSnapshot(snapshot)
	if err != nil {
		return HiscoreSnapshot{}, errors.Join(ErrSnapshotValidation, err)
	}

	xpChange := 0
	previousSnapshot, err := ss.GetSnapshotForUserNearestTimestamp(ctx, snapshot.UserId, time.Now().Unix())
	if err != nil && !errors.Is(err, ErrSnapshotNotFound) {
		return HiscoreSnapshot{}, errors.Join(ErrSnapshotGeneric, err)
	} else {
		xpChange = snapshot.GetSkill(ActivityTypeOverall).Experience - previousSnapshot.GetSkill(ActivityTypeOverall).Experience
	}

	dataSnapshot := snapshot.ToData()
	dataSnapshot.OverallExperienceChange = xpChange

	data, err := ss.repository.InsertSnapshot(ctx, dataSnapshot)
	if err != nil {
		return HiscoreSnapshot{}, errors.Join(ErrSnapshotGeneric, err)
	}

	return HiscoreSnapshot{}.FromData(data), nil
}

func validateSnapshotInterval(startTime, endTime time.Time) (time.Time, time.Time, error) {
	if startTime.Equal(endTime) {
		return time.Time{}, time.Time{}, errors.Join(ErrInvalidIntervalRequest, errors.New("start time must not equal end time"))
	}

	if endTime.After(time.Now()) {
		endTime = time.Now()
	}

	if endTime.Before(startTime) {
		tmp := endTime
		endTime = startTime
		startTime = tmp
	}

	if endTime.Sub(startTime) > MaxIntervalDuration {
		return time.Time{}, time.Time{}, errors.Join(ErrInvalidIntervalRequest, errors.New("maximum time interval exceeded"))
	}

	return startTime, endTime, nil
}
