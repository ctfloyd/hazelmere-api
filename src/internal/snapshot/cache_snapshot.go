package snapshot

import (
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/patrickmn/go-cache"
	"sort"
	"time"
)

const (
	CacheCleanupInterval = 1 * time.Hour
)

type CachedUserData struct {
	Snapshots  []HiscoreSnapshotData
	StartTime  time.Time
	EndTime    time.Time
	CachedAt   time.Time
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

func (sc *SnapshotCache) GetMissingRange(userId string, startTime, endTime time.Time) (needsFetch bool, fetchStart, fetchEnd time.Time) {
	cached, found := sc.GetUserData(userId)
	if !found {
		return true, startTime, endTime
	}

	if cached.StartTime.After(startTime) {
		return true, startTime, endTime
	}

	if cached.EndTime.Before(endTime) {
		return true, cached.EndTime, endTime
	}

	return false, time.Time{}, time.Time{}
}

func (sc *SnapshotCache) AppendSnapshots(userId string, newSnapshots []HiscoreSnapshotData, newEndTime time.Time) {
	cached, found := sc.GetUserData(userId)
	if !found {
		return
	}

	cached.Snapshots = append(cached.Snapshots, newSnapshots...)
	cached.EndTime = newEndTime
	cached.CachedAt = time.Now()
	sc.SetUserData(userId, cached)
}

func (sc *SnapshotCache) AppendSnapshot(userId string, snapshot HiscoreSnapshotData) {
	cached, found := sc.GetUserData(userId)
	if !found {
		return
	}

	cached.Snapshots = append(cached.Snapshots, snapshot)
	if snapshot.Timestamp.After(cached.EndTime) {
		cached.EndTime = snapshot.Timestamp
	}
	cached.CachedAt = time.Now()
	sc.SetUserData(userId, cached)
}

func (sc *SnapshotCache) FilterAndAggregate(userId string, startTime, endTime time.Time, window api.AggregationWindow) ([]HiscoreSnapshotData, int, int, bool) {
	cached, found := sc.GetUserData(userId)
	if !found {
		return nil, 0, 0, false
	}

	if cached.StartTime.After(startTime) {
		return nil, 0, 0, false
	}

	var filtered []HiscoreSnapshotData
	totalCount := 0
	gainsCount := 0

	for _, s := range cached.Snapshots {
		if (s.Timestamp.Equal(startTime) || s.Timestamp.After(startTime)) &&
			(s.Timestamp.Equal(endTime) || s.Timestamp.Before(endTime)) {
			totalCount++
			if s.OverallExperienceChange > 0 {
				gainsCount++
			}
			if s.OverallExperienceChange != 0 {
				filtered = append(filtered, s)
			}
		}
	}

	aggregated := aggregateByWindow(filtered, window)
	return aggregated, totalCount, gainsCount, true
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
