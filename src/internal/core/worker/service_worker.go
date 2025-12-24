package worker

import (
	"context"
	"errors"

	"github.com/ctfloyd/hazelmere-api/src/internal/core/snapshot"
	"github.com/ctfloyd/hazelmere-api/src/internal/foundation/monitor"
	"github.com/ctfloyd/hazelmere-worker/src/pkg/worker_client"
)

var ErrWorkerGeneric = errors.New("an error occurred while performing the worker operation")
var ErrHiscoreTimeout = errors.New("hiscore timeout")

type WorkerService interface {
	GenerateSnapshotOnDemand(ctx context.Context, userId string) (snapshot.HiscoreSnapshot, error)
}

type workerService struct {
	monitor         *monitor.Monitor
	workerClient    *worker_client.HazelmereWorker
	snapshotService snapshot.SnapshotService
}

func NewWorkerService(mon *monitor.Monitor, workerClient *worker_client.HazelmereWorker, snapshotService snapshot.SnapshotService) WorkerService {
	return &workerService{
		monitor:         mon,
		workerClient:    workerClient,
		snapshotService: snapshotService,
	}
}

func (ws *workerService) GenerateSnapshotOnDemand(ctx context.Context, userId string) (snapshot.HiscoreSnapshot, error) {
	ctx, span := ws.monitor.StartSpan(ctx, "workerService.GenerateSnapshotOnDemand")
	defer span.End()

	response, err := ws.workerClient.Snapshot.GenerateSnapshotOnDemand(userId)
	if err != nil {
		if errors.Is(err, worker_client.ErrRunescapeHiscoreTimeout) {
			return snapshot.HiscoreSnapshot{}, ErrHiscoreTimeout
		}
		return snapshot.HiscoreSnapshot{}, errors.Join(ErrWorkerGeneric, err)
	}

	ss, err := ws.snapshotService.GetSnapshotById(ctx, response.SnapshotId)
	if err != nil {
		return snapshot.HiscoreSnapshot{}, errors.Join(ErrWorkerGeneric, err)
	}

	return ss, nil
}
