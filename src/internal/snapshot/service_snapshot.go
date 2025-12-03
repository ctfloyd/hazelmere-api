package snapshot

import (
	"context"
	"errors"
	"github.com/ctfloyd/hazelmere-api/src/internal/database"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"github.com/google/uuid"
	"time"
)

const MaxIntervalDuration = 365 * 2 * 24 * time.Hour

var ErrSnapshotGeneric = errors.New("an unexpected error occurred while performing snapshot operation")
var ErrSnapshotValidation = errors.New("snapshot is invalid")
var ErrSnapshotNotFound = errors.New("snapshot not found")
var ErrInvalidIntervalRequest = errors.New("invalid interval request")

type SnapshotService interface {
	CreateSnapshot(ctx context.Context, snapshot HiscoreSnapshot) (HiscoreSnapshot, error)
	GetSnapshotById(ctx context.Context, id string) (HiscoreSnapshot, error)
	GetSnapshotInterval(ctx context.Context, userId string, startTime time.Time, endTime time.Time) ([]HiscoreSnapshot, error)
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

func (ss *snapshotService) GetSnapshotInterval(ctx context.Context, userId string, startTime time.Time, endTime time.Time) ([]HiscoreSnapshot, error) {
	startTime, endTime, err := validateSnapshotInterval(startTime, endTime)
	if err != nil {
		return nil, err
	}

	data, err := ss.repository.GetSnapshotInterval(ctx, userId, startTime, endTime)
	if err != nil {
		return nil, errors.Join(ErrSnapshotGeneric, err)
	}

	snapshots := MapManyDataToDomain(data)

	return snapshots, nil
}

func (ss *snapshotService) GetAllSnapshotsForUser(ctx context.Context, userId string) ([]HiscoreSnapshot, error) {
	data, err := ss.repository.GetAllSnapshotsForUser(ctx, userId)
	if err != nil {
		return nil, errors.Join(ErrSnapshotGeneric, err)
	}
	return MapManyDataToDomain(data), nil
}

func (ss *snapshotService) GetSnapshotById(ctx context.Context, id string) (HiscoreSnapshot, error) {
	data, err := ss.repository.GetSnapshotById(ctx, id)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return HiscoreSnapshot{}, ErrSnapshotNotFound
		}
		return HiscoreSnapshot{}, errors.Join(ErrSnapshotGeneric, err)
	}
	return MapDataToDomain(data), nil
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

	return MapDataToDomain(data), nil
}

func (ss *snapshotService) CreateSnapshot(ctx context.Context, snapshot HiscoreSnapshot) (HiscoreSnapshot, error) {
	snapshot.Id = uuid.New().String()

	err := ss.validator.ValidateSnapshot(snapshot)
	if err != nil {
		return HiscoreSnapshot{}, errors.Join(ErrSnapshotValidation, err)
	}

	xpChange := 0
	previousSnapshot, err := ss.GetSnapshotForUserNearestTimestamp(ctx, snapshot.UserId, time.Now().Unix())
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return HiscoreSnapshot{}, errors.Join(ErrSnapshotGeneric, err)
	} else {
		xpChange = snapshot.GetSkill(ActivityTypeOverall).Experience - previousSnapshot.GetSkill(ActivityTypeOverall).Experience
	}

	dataSnapshot := MapDomainToData(snapshot)
	dataSnapshot.OverallExperienceChange = xpChange

	data, err := ss.repository.InsertSnapshot(ctx, MapDomainToData(snapshot))
	if err != nil {
		return HiscoreSnapshot{}, errors.Join(ErrSnapshotGeneric, err)
	}

	return MapDataToDomain(data), nil
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

func filterUnchangedSnapshots(snapshots []HiscoreSnapshot) []HiscoreSnapshot {
	if len(snapshots) == 0 {
		return []HiscoreSnapshot{}
	}

	deltaSnapshots := make([]HiscoreSnapshot, 0)
	deltaSnapshots = append(deltaSnapshots, snapshots[0])

	for i := 1; i < len(snapshots); i++ {
		previous := snapshots[i-1]
		current := snapshots[i]

		if !current.Equals(previous) {
			deltaSnapshots = append(deltaSnapshots, current)
		}
	}

	return deltaSnapshots
}
