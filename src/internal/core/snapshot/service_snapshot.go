package snapshot

import (
	"context"
	"errors"
	"time"

	"github.com/ctfloyd/hazelmere-api/src/internal/core/user"
	"github.com/ctfloyd/hazelmere-api/src/internal/database"
	"github.com/ctfloyd/hazelmere-api/src/internal/foundation/monitor"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/google/uuid"
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
	GetLatestSnapshotForUser(ctx context.Context, userId string) (HiscoreSnapshot, error)
}

type snapshotService struct {
	monitor        *monitor.Monitor
	validator      SnapshotValidator
	repository     SnapshotRepository
	userRepository user.UserRepository
}

func NewSnapshotService(mon *monitor.Monitor, repository SnapshotRepository, validator SnapshotValidator, userRepository user.UserRepository) SnapshotService {
	return &snapshotService{
		monitor:        mon,
		validator:      validator,
		repository:     repository,
		userRepository: userRepository,
	}
}

const (
	DailyMaxDuration  = 366 * 24 * time.Hour
	WeeklyMaxDuration = 2 * 366 * 24 * time.Hour
)

var ErrInvalidAggregationWindow = errors.New("invalid aggregation window for requested time range")

func (ss *snapshotService) GetSnapshotInterval(ctx context.Context, userId string, startTime time.Time, endTime time.Time, aggregationWindow api.AggregationWindow) (SnapshotIntervalResponse, error) {
	ctx, span := ss.monitor.StartSpan(ctx, "snapshotService.GetSnapshotInterval")
	defer span.End()

	startTime, endTime, err := validateSnapshotInterval(startTime, endTime)
	if err != nil {
		return SnapshotIntervalResponse{}, err
	}

	aggregationWindow = normalizeAggregationWindow(aggregationWindow)

	if err := validateAggregationWindow(startTime, endTime, aggregationWindow); err != nil {
		return SnapshotIntervalResponse{}, err
	}

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

func validateAggregationWindow(startTime, endTime time.Time, window api.AggregationWindow) error {
	duration := endTime.Sub(startTime)

	switch window {
	case api.AggregationWindowDaily:
		if duration > DailyMaxDuration {
			return errors.Join(ErrInvalidAggregationWindow, errors.New("daily aggregation requires time range <= 366 days"))
		}
	case api.AggregationWindowWeekly:
		if duration > WeeklyMaxDuration {
			return errors.Join(ErrInvalidAggregationWindow, errors.New("weekly aggregation requires time range <= 2 years"))
		}
	case api.AggregationWindowMonthly:
		// Monthly is always valid
	}

	return nil
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
	ctx, span := ss.monitor.StartSpan(ctx, "snapshotService.GetAllSnapshotsForUser")
	defer span.End()

	data, err := ss.repository.GetAllSnapshotsForUser(ctx, userId)
	if err != nil {
		return nil, errors.Join(ErrSnapshotGeneric, err)
	}
	return HiscoreSnapshot{}.ManyFromData(data), nil
}

func (ss *snapshotService) GetSnapshotById(ctx context.Context, id string) (HiscoreSnapshot, error) {
	ctx, span := ss.monitor.StartSpan(ctx, "snapshotService.GetSnapshotById")
	defer span.End()

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
	ctx, span := ss.monitor.StartSpan(ctx, "snapshotService.GetSnapshotForUserNearestTimestamp")
	defer span.End()

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
	ctx, span := ss.monitor.StartSpan(ctx, "snapshotService.CreateSnapshot")
	defer span.End()

	snapshot.Id = uuid.New().String()

	err := ss.validator.ValidateSnapshot(snapshot)
	if err != nil {
		return HiscoreSnapshot{}, errors.Join(ErrSnapshotValidation, err)
	}

	xpChange := 0

	previousSnapshotData, err := ss.repository.GetLatestSnapshotForUser(ctx, snapshot.UserId)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return HiscoreSnapshot{}, errors.Join(ErrSnapshotGeneric, err)
	} else if err == nil {
		previousSnapshot := HiscoreSnapshot{}.FromData(previousSnapshotData)
		xpChange = snapshot.GetSkill(ActivityTypeOverall).Experience - previousSnapshot.GetSkill(ActivityTypeOverall).Experience
	}

	dataSnapshot := snapshot.ToData()
	dataSnapshot.OverallExperienceChange = xpChange

	data, err := ss.repository.InsertSnapshot(ctx, dataSnapshot)
	if err != nil {
		return HiscoreSnapshot{}, errors.Join(ErrSnapshotGeneric, err)
	}

	createdSnapshot := HiscoreSnapshot{}.FromData(data)
	ss.monitor.Logger().DebugArgs(ctx, "Created snapshot for user: %s", snapshot.UserId)

	return createdSnapshot, nil
}

func (ss *snapshotService) GetLatestSnapshotForUser(ctx context.Context, userId string) (HiscoreSnapshot, error) {
	ctx, span := ss.monitor.StartSpan(ctx, "snapshotService.GetLatestSnapshotForUser")
	defer span.End()

	data, err := ss.repository.GetLatestSnapshotForUser(ctx, userId)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return HiscoreSnapshot{}, ErrSnapshotNotFound
		}
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
