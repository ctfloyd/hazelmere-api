package snapshot

import (
	"api/src/internal/common/logger"
	"context"
)

type SnapshotService interface {
	GetAllSnapshotsForUser(ctx context.Context, userId string) ([]HiscoreSnapshot, error)
	//GetLatestSnapshotForUser(ctx context.Context, userId string) (HiscoreSnapshotData, error)
	//InsertSnapshot(ctx context.Context, snapshot HiscoreSnapshotData) (HiscoreSnapshotData, error)
}

type snapshotService struct {
	logger     logger.Logger
	repository SnapshotRepository
}

func NewSnapshotService(logger logger.Logger, repository SnapshotRepository) SnapshotService {
	return &snapshotService{
		logger:     logger,
		repository: repository,
	}
}

func (ss *snapshotService) GetAllSnapshotsForUser(ctx context.Context, userId string) ([]HiscoreSnapshot, error) {
	data, err := ss.repository.GetAllSnapshotsForUser(ctx, userId)
	if err != nil {
		return nil, err
	}
	return MapManyDataToDomain(data), nil
}
