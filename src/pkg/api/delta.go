package api

import "time"

type HiscoreDelta struct {
	Id                 string          `json:"id"`
	UserId             string          `json:"userId"`
	SnapshotId         string          `json:"snapshotId"`
	PreviousSnapshotId string          `json:"previousSnapshotId"`
	Timestamp          time.Time       `json:"timestamp"`
	Skills             []SkillDelta    `json:"skills,omitempty"`
	Bosses             []BossDelta     `json:"bosses,omitempty"`
	Activities         []ActivityDelta `json:"activities,omitempty"`
}

type SkillDelta struct {
	ActivityType   ActivityType `json:"activityType"`
	Name           string       `json:"name"`
	ExperienceGain int          `json:"experienceGain"`
	LevelGain      int          `json:"levelGain"`
}

type BossDelta struct {
	ActivityType  ActivityType `json:"activityType"`
	Name          string       `json:"name"`
	KillCountGain int          `json:"killCountGain"`
}

type ActivityDelta struct {
	ActivityType ActivityType `json:"activityType"`
	Name         string       `json:"name"`
	ScoreGain    int          `json:"scoreGain"`
}

type GetLatestDeltaResponse struct {
	Delta HiscoreDelta `json:"delta"`
}

type GetDeltaIntervalRequest struct {
	UserId    string    `json:"userId"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type GetDeltaIntervalResponse struct {
	Deltas      []HiscoreDelta `json:"deltas"`
	TotalDeltas int            `json:"totalDeltas"`
}

type GetDeltaSummaryRequest struct {
	UserId    string    `json:"userId"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type GetDeltaSummaryResponse struct {
	UserId              string                 `json:"userId"`
	StartTime           time.Time              `json:"startTime"`
	EndTime             time.Time              `json:"endTime"`
	TotalExperienceGain int                    `json:"totalExperienceGain"`
	Skills              []SkillDeltaSummary    `json:"skills"`
	Bosses              []BossDeltaSummary     `json:"bosses"`
	Activities          []ActivityDeltaSummary `json:"activities"`
	DeltaCount          int                    `json:"deltaCount"`
}

type SkillDeltaSummary struct {
	ActivityType        ActivityType `json:"activityType"`
	Name                string       `json:"name"`
	TotalExperienceGain int          `json:"totalExperienceGain"`
	TotalLevelGain      int          `json:"totalLevelGain"`
}

type BossDeltaSummary struct {
	ActivityType       ActivityType `json:"activityType"`
	Name               string       `json:"name"`
	TotalKillCountGain int          `json:"totalKillCountGain"`
}

type ActivityDeltaSummary struct {
	ActivityType   ActivityType `json:"activityType"`
	Name           string       `json:"name"`
	TotalScoreGain int          `json:"totalScoreGain"`
}

// GetSnapshotWithDeltasRequest is used to request a snapshot and all deltas in a time range
type GetSnapshotWithDeltasRequest struct {
	UserId    string    `json:"userId"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

// GetSnapshotWithDeltasResponse contains a snapshot and all deltas in the requested range
type GetSnapshotWithDeltasResponse struct {
	Snapshot HiscoreSnapshot `json:"snapshot"`
	Deltas   []HiscoreDelta  `json:"deltas"`
}
