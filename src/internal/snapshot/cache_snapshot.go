package snapshot

import (
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/patrickmn/go-cache"
	"sort"
	"time"
)

const (
	CacheCleanupInterval = 1 * time.Hour
	DailyThreshold       = 365 * 24 * time.Hour
	WeeklyThreshold      = 2 * 365 * 24 * time.Hour
)

type CachedUserData struct {
	Daily            []HiscoreSnapshotData
	Weekly           []HiscoreSnapshotData
	Monthly          []HiscoreSnapshotData
	TotalSnapshots   int
	SnapshotsWithGains int
	StartTime        time.Time
	EndTime          time.Time
	CachedAt         time.Time
}

type SnapshotCache struct {
	cache *cache.Cache
}

func NewSnapshotCache() *SnapshotCache {
	return &SnapshotCache{
		cache: cache.New(cache.NoExpiration, CacheCleanupInterval),
	}
}

func (sc *SnapshotCache) GetUserData(userId string) (CachedUserData, bool) {
	if cached, found := sc.cache.Get(userId); found {
		return cached.(CachedUserData), true
	}
	return CachedUserData{}, false
}

func (sc *SnapshotCache) SetUserData(userId string, data CachedUserData) {
	sc.cache.Set(userId, data, cache.NoExpiration)
}

func (sc *SnapshotCache) InvalidateUser(userId string) {
	sc.cache.Delete(userId)
}

func (sc *SnapshotCache) IsCached(userId string) bool {
	_, found := sc.cache.Get(userId)
	return found
}

func (sc *SnapshotCache) GetAggregatedData(userId string, startTime, endTime time.Time, window api.AggregationWindow) ([]HiscoreSnapshotData, int, int, bool) {
	cached, found := sc.GetUserData(userId)
	if !found {
		return nil, 0, 0, false
	}

	var source []HiscoreSnapshotData
	switch window {
	case api.AggregationWindowDaily:
		source = cached.Daily
	case api.AggregationWindowWeekly:
		source = cached.Weekly
	case api.AggregationWindowMonthly:
		source = cached.Monthly
	default:
		source = cached.Daily
	}

	var filtered []HiscoreSnapshotData
	for _, s := range source {
		if (s.Timestamp.Equal(startTime) || s.Timestamp.After(startTime)) &&
			(s.Timestamp.Equal(endTime) || s.Timestamp.Before(endTime)) {
			filtered = append(filtered, s)
		}
	}

	return filtered, cached.TotalSnapshots, cached.SnapshotsWithGains, true
}

func (sc *SnapshotCache) BuildAndSetUserData(userId string, rawSnapshots []HiscoreSnapshotData, startTime, endTime time.Time) {
	totalCount := len(rawSnapshots)
	gainsCount := 0

	var withChanges []HiscoreSnapshotData
	for _, s := range rawSnapshots {
		if s.OverallExperienceChange > 0 {
			gainsCount++
		}
		if s.OverallExperienceChange != 0 {
			withChanges = append(withChanges, s)
		}
	}

	daily := aggregateByWindow(withChanges, api.AggregationWindowDaily)
	weekly := aggregateByWindow(withChanges, api.AggregationWindowWeekly)
	monthly := aggregateByWindow(withChanges, api.AggregationWindowMonthly)

	sc.SetUserData(userId, CachedUserData{
		Daily:              daily,
		Weekly:             weekly,
		Monthly:            monthly,
		TotalSnapshots:     totalCount,
		SnapshotsWithGains: gainsCount,
		StartTime:          startTime,
		EndTime:            endTime,
		CachedAt:           time.Now(),
	})
}

func (sc *SnapshotCache) AppendSnapshot(userId string, snapshot HiscoreSnapshotData) {
	cached, found := sc.GetUserData(userId)
	if !found {
		return
	}

	cached.TotalSnapshots++
	if snapshot.OverallExperienceChange > 0 {
		cached.SnapshotsWithGains++
	}

	if snapshot.OverallExperienceChange != 0 {
		cached.Daily = appendAndReaggregate(cached.Daily, snapshot, api.AggregationWindowDaily)
		cached.Weekly = appendAndReaggregate(cached.Weekly, snapshot, api.AggregationWindowWeekly)
		cached.Monthly = appendAndReaggregate(cached.Monthly, snapshot, api.AggregationWindowMonthly)
	}

	if snapshot.Timestamp.After(cached.EndTime) {
		cached.EndTime = snapshot.Timestamp
	}
	cached.CachedAt = time.Now()
	sc.SetUserData(userId, cached)
}

func appendAndReaggregate(existing []HiscoreSnapshotData, newSnapshot HiscoreSnapshotData, window api.AggregationWindow) []HiscoreSnapshotData {
	dateFormat := getDateFormatString(window)
	newKey := newSnapshot.Timestamp.Format(dateFormat)

	for i, s := range existing {
		if s.Timestamp.Format(dateFormat) == newKey {
			best := selectBestSnapshot([]HiscoreSnapshotData{s, newSnapshot})
			existing[i] = best
			return existing
		}
	}

	existing = append(existing, newSnapshot)
	sort.Slice(existing, func(i, j int) bool {
		return existing[i].Timestamp.Before(existing[j].Timestamp)
	})
	return existing
}

func aggregateByWindow(snapshots []HiscoreSnapshotData, window api.AggregationWindow) []HiscoreSnapshotData {
	if len(snapshots) == 0 {
		return snapshots
	}

	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Timestamp.Before(snapshots[j].Timestamp)
	})

	dateFormat := getDateFormatString(window)
	groups := make(map[string][]HiscoreSnapshotData)

	for _, s := range snapshots {
		key := s.Timestamp.Format(dateFormat)
		groups[key] = append(groups[key], s)
	}

	var result []HiscoreSnapshotData
	for _, group := range groups {
		best := selectBestSnapshot(group)
		result = append(result, best)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})

	return result
}

func getDateFormatString(window api.AggregationWindow) string {
	switch window {
	case api.AggregationWindowWeekly:
		return "2006-W01"
	case api.AggregationWindowMonthly:
		return "2006-01"
	default:
		return "2006-01-02"
	}
}

func selectBestSnapshot(snapshots []HiscoreSnapshotData) HiscoreSnapshotData {
	if len(snapshots) == 0 {
		return HiscoreSnapshotData{}
	}

	best := snapshots[0]
	bestExp := getOverallExperience(best)

	for _, s := range snapshots[1:] {
		exp := getOverallExperience(s)
		if exp > bestExp || (exp == bestExp && s.Timestamp.After(best.Timestamp)) {
			best = s
			bestExp = exp
		}
	}

	return best
}

func getOverallExperience(s HiscoreSnapshotData) int {
	for _, skill := range s.Skills {
		if skill.ActivityType == "OVERALL" {
			return skill.Experience
		}
	}
	return 0
}
