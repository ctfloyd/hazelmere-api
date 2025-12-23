package hiscore

import (
	"github.com/ctfloyd/hazelmere-api/src/internal/core/delta"
	"github.com/ctfloyd/hazelmere-api/src/internal/core/snapshot"
)

// CreateSnapshotResponse contains all objects created when creating a snapshot with delta
type CreateSnapshotResponse struct {
	Snapshot snapshot.HiscoreSnapshot
	Delta    *delta.HiscoreDelta // nil if no previous snapshot existed
}

// DeltaSummaryResponse contains a snapshot and all deltas for a time range
type DeltaSummaryResponse struct {
	Snapshot snapshot.HiscoreSnapshot
	Deltas   []delta.HiscoreDelta
}
