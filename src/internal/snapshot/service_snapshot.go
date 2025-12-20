package snapshot

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ctfloyd/hazelmere-api/src/internal/database"
	"github.com/ctfloyd/hazelmere-api/src/internal/user"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
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
	PrimeCache(ctx context.Context) error
}

type snapshotService struct {
	logger         hz_logger.Logger
	validator      SnapshotValidator
	repository     SnapshotRepository
	userRepository user.UserRepository
	cache          *SnapshotCache
}

func NewSnapshotService(logger hz_logger.Logger, repository SnapshotRepository, validator SnapshotValidator, cache *SnapshotCache, userRepository user.UserRepository) SnapshotService {
	return &snapshotService{
		logger:         logger,
		validator:      validator,
		repository:     repository,
		userRepository: userRepository,
		cache:          cache,
	}
}

const (
	DailyMaxDuration  = 366 * 24 * time.Hour
	WeeklyMaxDuration = 2 * 366 * 24 * time.Hour
)

var ErrInvalidAggregationWindow = errors.New("invalid aggregation window for requested time range")

func (ss *snapshotService) GetSnapshotInterval(ctx context.Context, userId string, startTime time.Time, endTime time.Time, aggregationWindow api.AggregationWindow) (SnapshotIntervalResponse, error) {
	startTime, endTime, err := validateSnapshotInterval(startTime, endTime)
	if err != nil {
		return SnapshotIntervalResponse{}, err
	}

	aggregationWindow = normalizeAggregationWindow(aggregationWindow)

	if err := validateAggregationWindow(startTime, endTime, aggregationWindow); err != nil {
		return SnapshotIntervalResponse{}, err
	}

	if !ss.cache.IsCached(userId) {
		ss.logger.DebugArgs(ctx, "Cache miss for user %s, falling back to repository", userId)
		return ss.getIntervalFromRepository(ctx, userId, startTime, endTime, aggregationWindow)
	}

	snapshots, totalCount, gainsCount, ok := ss.cache.GetAggregatedData(userId, startTime, endTime, aggregationWindow)
	if !ok {
		ss.logger.WarnArgs(ctx, "Cache miss after check for user %s", userId)
		return ss.getIntervalFromRepository(ctx, userId, startTime, endTime, aggregationWindow)
	}

	ss.logger.DebugArgs(ctx, "Cache hit for interval query: userId=%s, window=%s", userId, aggregationWindow)
	return SnapshotIntervalResponse{
		Snapshots:          HiscoreSnapshot{}.ManyFromData(snapshots),
		TotalSnapshots:     totalCount,
		SnapshotsWithGains: gainsCount,
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

func (ss *snapshotService) getIntervalFromRepository(ctx context.Context, userId string, startTime, endTime time.Time, aggregationWindow api.AggregationWindow) (SnapshotIntervalResponse, error) {
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

	ss.cache.AppendSnapshot(snapshot.UserId, data)
	ss.logger.DebugArgs(ctx, "Appended snapshot to cache for user: %s", snapshot.UserId)

	return HiscoreSnapshot{}.FromData(data), nil
}

func (ss *snapshotService) PrimeCache(ctx context.Context) error {
	users, err := ss.userRepository.GetUsersWithTrackingEnabled(ctx)
	if err != nil {
		return err
	}

	ss.logger.InfoArgs(ctx, "Priming cache for %d users with tracking enabled", len(users))

	var wg sync.WaitGroup
	for _, u := range users {
		wg.Add(1)
		go func(userId string) {
			defer wg.Done()
			ss.primeUserCache(ctx, userId)
		}(u.Id)
	}
	wg.Wait()

	ss.logger.Info(ctx, "Cache priming complete")
	return nil
}

const batchSize = 30 * 24 * time.Hour // 30 days per batch

func (ss *snapshotService) primeUserCache(ctx context.Context, userId string) {
	oldest, err := ss.repository.GetOldestSnapshotForUser(ctx, userId)
	if err != nil {
		ss.logger.DebugArgs(ctx, "No snapshots found for user %s, skipping cache prime", userId)
		return
	}

	latest, err := ss.repository.GetLatestSnapshotForUser(ctx, userId)
	if err != nil {
		ss.logger.WarnArgs(ctx, "Failed to get latest snapshot for user %s: %v", userId, err)
		return
	}

	ss.logger.InfoArgs(ctx, "Priming user %s: oldest=%v, latest=%v", userId, oldest.Timestamp, latest.Timestamp)

	batches := ss.calculateBatches(oldest.Timestamp, latest.Timestamp)
	if len(batches) == 0 {
		ss.logger.WarnArgs(ctx, "No batches calculated for user %s", userId)
		return
	}

	results := make([][]HiscoreSnapshotData, len(batches))
	var wg sync.WaitGroup

	for i, batch := range batches {
		wg.Add(1)
		go func(idx int, start, end time.Time) {
			defer wg.Done()
			snapshots, err := ss.repository.GetSnapshotsInRange(ctx, userId, start, end)
			if err != nil {
				ss.logger.WarnArgs(ctx, "Failed to fetch batch %d for user %s: %v", idx, userId, err)
				return
			}
			results[idx] = snapshots
		}(i, batch.start, batch.end)
	}
	wg.Wait()

	var allSnapshots []HiscoreSnapshotData
	for _, batch := range results {
		allSnapshots = append(allSnapshots, batch...)
	}

	if len(allSnapshots) == 0 {
		ss.logger.DebugArgs(ctx, "No snapshots loaded for user %s", userId)
		return
	}

	ss.cache.BuildAndSetUserData(userId, allSnapshots, oldest.Timestamp, latest.Timestamp)

	ss.logger.DebugArgs(ctx, "Primed cache for user %s with %d snapshots in %d batches", userId, len(allSnapshots), len(batches))
}

type timeBatch struct {
	start time.Time
	end   time.Time
}

func (ss *snapshotService) calculateBatches(oldest, latest time.Time) []timeBatch {
	var batches []timeBatch
	current := oldest

	// Handle case where oldest equals latest (single snapshot or same timestamp)
	if !current.Before(latest) {
		return []timeBatch{{start: oldest, end: latest}}
	}

	for current.Before(latest) {
		batchEnd := current.Add(batchSize)
		if batchEnd.After(latest) {
			batchEnd = latest
		}
		batches = append(batches, timeBatch{start: current, end: batchEnd})
		current = batchEnd
	}

	// Ensure we always have at least one batch
	if len(batches) == 0 {
		batches = append(batches, timeBatch{start: oldest, end: latest})
	}

	return batches
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
