package snapshot

import (
	"context"
	"errors"
	"github.com/ctfloyd/hazelmere-api/src/internal/common/database"
	"github.com/ctfloyd/hazelmere-api/src/internal/common/logger"
	"github.com/google/uuid"
	"time"
)

var ErrSnapshotGeneric = errors.New("an unexpected error occurred while performing snapshot operation")
var ErrSnapshotValidation = errors.New("snapshot is invalid")
var ErrSnapshotNotFound = errors.New("snapshot not found")

type SnapshotService interface {
	CreateSnapshot(ctx context.Context, snapshot HiscoreSnapshot) (HiscoreSnapshot, error)
	GetAllSnapshotsForUser(ctx context.Context, userId string) ([]HiscoreSnapshot, error)
	GetSnapshotForUserNearestTimestamp(ctx context.Context, userId string, timestamp int64) (HiscoreSnapshot, error)
}

type snapshotService struct {
	logger     logger.Logger
	validator  SnapshotValidator
	repository SnapshotRepository
}

func NewSnapshotService(logger logger.Logger, repository SnapshotRepository, validator SnapshotValidator) SnapshotService {
	return &snapshotService{
		logger:     logger,
		validator:  validator,
		repository: repository,
	}
}

func (ss *snapshotService) GetAllSnapshotsForUser(ctx context.Context, userId string) ([]HiscoreSnapshot, error) {
	data, err := ss.repository.GetAllSnapshotsForUser(ctx, userId)
	if err != nil {
		return nil, errors.Join(ErrSnapshotGeneric, err)
	}
	return MapManyDataToDomain(data), nil
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

	data, err := ss.repository.InsertSnapshot(ctx, MapDomainToData(snapshot))
	if err != nil {
		return HiscoreSnapshot{}, errors.Join(ErrSnapshotGeneric, err)
	}

	return MapDataToDomain(data), nil
}
