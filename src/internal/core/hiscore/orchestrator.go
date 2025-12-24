package hiscore

import (
	"context"
	"errors"
	"time"

	"github.com/ctfloyd/hazelmere-api/src/internal/core/delta"
	"github.com/ctfloyd/hazelmere-api/src/internal/core/snapshot"
	"github.com/ctfloyd/hazelmere-api/src/internal/database"
	"github.com/ctfloyd/hazelmere-api/src/internal/foundation/monitor"
	"github.com/google/uuid"
)

var ErrSnapshotNotFound = errors.New("snapshot not found")
var ErrSnapshotValidation = errors.New("snapshot validation failed")

type HiscoreOrchestrator interface {
	CreateSnapshotWithDelta(ctx context.Context, snap snapshot.HiscoreSnapshot) (CreateSnapshotResponse, error)
	GetDeltaSummary(ctx context.Context, userId string, startTime, endTime time.Time) (DeltaSummaryResponse, error)
}

type hiscoreOrchestrator struct {
	monitor         *monitor.Monitor
	snapshotService snapshot.SnapshotService
	deltaService    delta.DeltaService
	txManager       *database.TransactionManager
}

func NewHiscoreOrchestrator(
	mon *monitor.Monitor,
	snapshotService snapshot.SnapshotService,
	deltaService delta.DeltaService,
	txManager *database.TransactionManager,
) HiscoreOrchestrator {
	return &hiscoreOrchestrator{
		monitor:         mon,
		snapshotService: snapshotService,
		deltaService:    deltaService,
		txManager:       txManager,
	}
}

func (o *hiscoreOrchestrator) CreateSnapshotWithDelta(ctx context.Context, snap snapshot.HiscoreSnapshot) (CreateSnapshotResponse, error) {
	ctx, span := o.monitor.StartSpan(ctx, "hiscoreOrchestrator.CreateSnapshotWithDelta")
	defer span.End()

	var createdSnapshot snapshot.HiscoreSnapshot
	var createdDelta *delta.HiscoreDelta
	var previousSnapshot snapshot.HiscoreSnapshot
	hasPreviousSnapshot := false

	// Get previous snapshot first (outside transaction - read-only)
	prev, err := o.snapshotService.GetLatestSnapshotForUser(ctx, snap.UserId)
	if err == nil {
		previousSnapshot = prev
		hasPreviousSnapshot = true
	} else if !errors.Is(err, snapshot.ErrSnapshotNotFound) {
		return CreateSnapshotResponse{}, err
	}

	err = o.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		created, err := o.snapshotService.CreateSnapshot(txCtx, snap)
		if err != nil {
			return err
		}
		createdSnapshot = created

		if hasPreviousSnapshot {
			d := o.computeDelta(ctx, previousSnapshot, createdSnapshot)
			insertedDelta, err := o.deltaService.CreateDelta(txCtx, d)
			if err != nil {
				return err
			}
			// Only set if delta was actually created (has changes)
			if insertedDelta.Id != "" {
				createdDelta = &insertedDelta
			}
			o.monitor.Logger().DebugArgs(txCtx, "Created delta for snapshot %s", createdSnapshot.Id)
		}
		return nil
	})

	if err != nil {
		return CreateSnapshotResponse{}, err
	}

	return CreateSnapshotResponse{
		Snapshot: createdSnapshot,
		Delta:    createdDelta,
	}, nil
}

func (o *hiscoreOrchestrator) GetDeltaSummary(ctx context.Context, userId string, startTime, endTime time.Time) (DeltaSummaryResponse, error) {
	ctx, span := o.monitor.StartSpan(ctx, "hiscoreOrchestrator.GetDeltaSummary")
	defer span.End()

	// Get snapshot nearest to start time
	startMs := startTime.UnixNano() / int64(time.Millisecond)
	snap, err := o.snapshotService.GetSnapshotForUserNearestTimestamp(ctx, userId, startMs)
	if err != nil {
		if errors.Is(err, snapshot.ErrSnapshotNotFound) {
			return DeltaSummaryResponse{}, ErrSnapshotNotFound
		}
		return DeltaSummaryResponse{}, err
	}

	// Get all deltas in the range
	deltaResp, err := o.deltaService.GetDeltasInRange(ctx, userId, startTime, endTime)
	if err != nil {
		return DeltaSummaryResponse{}, err
	}

	return DeltaSummaryResponse{
		Snapshot: snap,
		Deltas:   deltaResp.Deltas,
	}, nil
}

func (o *hiscoreOrchestrator) computeDelta(ctx context.Context, previousSnapshot, currentSnapshot snapshot.HiscoreSnapshot) delta.HiscoreDelta {
	_, span := o.monitor.StartSpan(ctx, "hiscoreOrchestrator.computeDelta")
	defer span.End()

	d := delta.HiscoreDelta{
		Id:                 uuid.New().String(),
		UserId:             currentSnapshot.UserId,
		SnapshotId:         currentSnapshot.Id,
		PreviousSnapshotId: previousSnapshot.Id,
		Timestamp:          currentSnapshot.Timestamp,
	}

	d.Skills = o.computeSkillDeltas(previousSnapshot.Skills, currentSnapshot.Skills)
	d.Bosses = o.computeBossDeltas(previousSnapshot.Bosses, currentSnapshot.Bosses)
	d.Activities = o.computeActivityDeltas(previousSnapshot.Activities, currentSnapshot.Activities)

	return d
}

func (o *hiscoreOrchestrator) computeSkillDeltas(previous, current []snapshot.SkillSnapshot) []delta.SkillDelta {
	var deltas []delta.SkillDelta

	prevMap := make(map[snapshot.ActivityType]snapshot.SkillSnapshot)
	for _, skill := range previous {
		prevMap[skill.ActivityType] = skill
	}

	for _, curr := range current {
		prev, exists := prevMap[curr.ActivityType]
		if !exists {
			continue
		}

		xpGain := curr.Experience - prev.Experience
		levelGain := curr.Level - prev.Level

		if xpGain != 0 || levelGain != 0 {
			deltas = append(deltas, delta.SkillDelta{
				ActivityType:   curr.ActivityType,
				Name:           curr.Name,
				ExperienceGain: xpGain,
				LevelGain:      levelGain,
			})
		}
	}

	return deltas
}

func (o *hiscoreOrchestrator) computeBossDeltas(previous, current []snapshot.BossSnapshot) []delta.BossDelta {
	var deltas []delta.BossDelta

	prevMap := make(map[snapshot.ActivityType]snapshot.BossSnapshot)
	for _, boss := range previous {
		prevMap[boss.ActivityType] = boss
	}

	for _, curr := range current {
		prev, exists := prevMap[curr.ActivityType]
		if !exists {
			continue
		}

		kcGain := curr.KillCount - prev.KillCount

		if kcGain != 0 {
			deltas = append(deltas, delta.BossDelta{
				ActivityType:  curr.ActivityType,
				Name:          curr.Name,
				KillCountGain: kcGain,
			})
		}
	}

	return deltas
}

func (o *hiscoreOrchestrator) computeActivityDeltas(previous, current []snapshot.ActivitySnapshot) []delta.ActivityDelta {
	var deltas []delta.ActivityDelta

	prevMap := make(map[snapshot.ActivityType]snapshot.ActivitySnapshot)
	for _, activity := range previous {
		prevMap[activity.ActivityType] = activity
	}

	for _, curr := range current {
		prev, exists := prevMap[curr.ActivityType]
		if !exists {
			continue
		}

		scoreGain := curr.Score - prev.Score

		if scoreGain != 0 {
			deltas = append(deltas, delta.ActivityDelta{
				ActivityType: curr.ActivityType,
				Name:         curr.Name,
				ScoreGain:    scoreGain,
			})
		}
	}

	return deltas
}
