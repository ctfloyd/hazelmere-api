package delta

import (
	"sort"
	"sync"
	"time"

	"github.com/ctfloyd/hazelmere-api/src/internal/core/snapshot"
)

// CachedUserDeltas holds pre-aggregated daily deltas for a user
type CachedUserDeltas struct {
	// DailyDeltas maps date string (YYYY-MM-DD) to aggregated delta for that day
	DailyDeltas map[string]HiscoreDelta
	CachedAt    time.Time
}

// DeltaCache stores pre-aggregated daily deltas in memory per user
type DeltaCache struct {
	mu    sync.RWMutex
	cache map[string]*CachedUserDeltas
}

func NewDeltaCache() *DeltaCache {
	return &DeltaCache{
		cache: make(map[string]*CachedUserDeltas),
	}
}

// SetUserDeltas aggregates raw deltas by day and stores them (used during priming)
func (dc *DeltaCache) SetUserDeltas(userId string, rawDeltas []HiscoreDeltaData) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dailyDeltas := aggregateDeltasByDay(rawDeltas)

	dc.cache[userId] = &CachedUserDeltas{
		DailyDeltas: dailyDeltas,
		CachedAt:    time.Now(),
	}
}

// AppendDelta adds a new delta to the user's cache, merging with existing daily aggregate
func (dc *DeltaCache) AppendDelta(userId string, deltaData HiscoreDeltaData) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	cached, exists := dc.cache[userId]
	if !exists {
		// If user not in cache yet, create new entry with this delta
		domainDelta := HiscoreDelta{}.FromData(deltaData)
		dateKey := deltaData.Timestamp.Format("2006-01-02")
		dc.cache[userId] = &CachedUserDeltas{
			DailyDeltas: map[string]HiscoreDelta{dateKey: domainDelta},
			CachedAt:    time.Now(),
		}
		return
	}

	// Merge with existing daily aggregate
	dateKey := deltaData.Timestamp.Format("2006-01-02")
	domainDelta := HiscoreDelta{}.FromData(deltaData)

	if existing, ok := cached.DailyDeltas[dateKey]; ok {
		// Merge the deltas
		merged := mergeTwoDeltas(existing, domainDelta)
		cached.DailyDeltas[dateKey] = merged
	} else {
		cached.DailyDeltas[dateKey] = domainDelta
	}
	cached.CachedAt = time.Now()
}

// GetLatestDelta returns the most recent daily aggregated delta for a user
func (dc *DeltaCache) GetLatestDelta(userId string) (HiscoreDelta, bool) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	cached, exists := dc.cache[userId]
	if !exists || len(cached.DailyDeltas) == 0 {
		return HiscoreDelta{}, false
	}

	// Find the latest date
	var latestDate string
	for dateKey := range cached.DailyDeltas {
		if dateKey > latestDate {
			latestDate = dateKey
		}
	}

	return cached.DailyDeltas[latestDate], true
}

// GetDeltasInRange returns daily aggregated deltas within the specified time range
func (dc *DeltaCache) GetDeltasInRange(userId string, startTime, endTime time.Time) ([]HiscoreDelta, bool) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	cached, exists := dc.cache[userId]
	if !exists {
		return nil, false
	}

	startDate := startTime.Format("2006-01-02")
	endDate := endTime.Format("2006-01-02")

	var result []HiscoreDelta
	for dateKey, d := range cached.DailyDeltas {
		if dateKey >= startDate && dateKey <= endDate {
			result = append(result, d)
		}
	}

	// Sort by timestamp
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})

	return result, true
}

// IsCached checks if a user has cached deltas
func (dc *DeltaCache) IsCached(userId string) bool {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	_, exists := dc.cache[userId]
	return exists
}

// GetDeltaCount returns the number of daily aggregated deltas cached for a user
func (dc *DeltaCache) GetDeltaCount(userId string) int {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	cached, exists := dc.cache[userId]
	if !exists {
		return 0
	}
	return len(cached.DailyDeltas)
}

// GetTotalCachedUsers returns the number of users with cached deltas
func (dc *DeltaCache) GetTotalCachedUsers() int {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	return len(dc.cache)
}

// aggregateDeltasByDay groups raw deltas by day and merges them
func aggregateDeltasByDay(rawDeltas []HiscoreDeltaData) map[string]HiscoreDelta {
	dailyMap := make(map[string]HiscoreDelta)

	for _, raw := range rawDeltas {
		dateKey := raw.Timestamp.Format("2006-01-02")
		domainDelta := HiscoreDelta{}.FromData(raw)

		if existing, ok := dailyMap[dateKey]; ok {
			// Merge with existing
			merged := mergeTwoDeltas(existing, domainDelta)
			dailyMap[dateKey] = merged
		} else {
			dailyMap[dateKey] = domainDelta
		}
	}

	return dailyMap
}

// mergeTwoDeltas merges two deltas into one by summing their gains
func mergeTwoDeltas(a, b HiscoreDelta) HiscoreDelta {
	// Use the later timestamp and IDs from the second delta
	merged := HiscoreDelta{
		Id:                 b.Id,
		UserId:             b.UserId,
		SnapshotId:         b.SnapshotId,
		PreviousSnapshotId: a.PreviousSnapshotId, // Keep the earliest previous
		Timestamp:          b.Timestamp,          // Use the latest timestamp
	}

	// Merge skills
	skillMap := make(map[snapshot.ActivityType]*SkillDelta)
	for _, s := range a.Skills {
		skillMap[s.ActivityType] = &SkillDelta{
			ActivityType:   s.ActivityType,
			Name:           s.Name,
			ExperienceGain: s.ExperienceGain,
			LevelGain:      s.LevelGain,
		}
	}
	for _, s := range b.Skills {
		if existing, ok := skillMap[s.ActivityType]; ok {
			existing.ExperienceGain += s.ExperienceGain
			existing.LevelGain += s.LevelGain
		} else {
			skillMap[s.ActivityType] = &SkillDelta{
				ActivityType:   s.ActivityType,
				Name:           s.Name,
				ExperienceGain: s.ExperienceGain,
				LevelGain:      s.LevelGain,
			}
		}
	}
	for _, s := range skillMap {
		merged.Skills = append(merged.Skills, *s)
	}

	// Merge bosses
	bossMap := make(map[snapshot.ActivityType]*BossDelta)
	for _, b := range a.Bosses {
		bossMap[b.ActivityType] = &BossDelta{
			ActivityType:  b.ActivityType,
			Name:          b.Name,
			KillCountGain: b.KillCountGain,
		}
	}
	for _, b := range b.Bosses {
		if existing, ok := bossMap[b.ActivityType]; ok {
			existing.KillCountGain += b.KillCountGain
		} else {
			bossMap[b.ActivityType] = &BossDelta{
				ActivityType:  b.ActivityType,
				Name:          b.Name,
				KillCountGain: b.KillCountGain,
			}
		}
	}
	for _, b := range bossMap {
		merged.Bosses = append(merged.Bosses, *b)
	}

	// Merge activities
	activityMap := make(map[snapshot.ActivityType]*ActivityDelta)
	for _, act := range a.Activities {
		activityMap[act.ActivityType] = &ActivityDelta{
			ActivityType: act.ActivityType,
			Name:         act.Name,
			ScoreGain:    act.ScoreGain,
		}
	}
	for _, act := range b.Activities {
		if existing, ok := activityMap[act.ActivityType]; ok {
			existing.ScoreGain += act.ScoreGain
		} else {
			activityMap[act.ActivityType] = &ActivityDelta{
				ActivityType: act.ActivityType,
				Name:         act.Name,
				ScoreGain:    act.ScoreGain,
			}
		}
	}
	for _, act := range activityMap {
		merged.Activities = append(merged.Activities, *act)
	}

	return merged
}
