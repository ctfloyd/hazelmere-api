package api

import "time"

type AggregationWindow string

const (
	AggregationWindowDaily   AggregationWindow = "daily"
	AggregationWindowWeekly  AggregationWindow = "weekly"
	AggregationWindowMonthly AggregationWindow = "monthly"
)

type HiscoreSnapshot struct {
	Id         string             `json:"id"`
	UserId     string             `json:"userId"`
	Timestamp  time.Time          `json:"timestamp"`
	Skills     []SkillSnapshot    `json:"skills"`
	Bosses     []BossSnapshot     `json:"bosses"`
	Activities []ActivitySnapshot `json:"activities"`
}

type SkillSnapshot struct {
	ActivityType ActivityType `json:"activityType"`
	Name         string       `json:"name"`
	Level        int          `json:"level"`
	Experience   int          `json:"experience"`
	Rank         int          `json:"rank"`
}

type BossSnapshot struct {
	ActivityType ActivityType `json:"activityType"`
	Name         string       `json:"name"`
	KillCount    int          `json:"killCount"`
	Rank         int          `json:"rank"`
}

type ActivitySnapshot struct {
	ActivityType ActivityType `json:"activityType"`
	Name         string       `json:"name"`
	Score        int          `json:"score"`
	Rank         int          `json:"rank"`
}

type CreateSnapshotRequest struct {
	Snapshot HiscoreSnapshot `json:"snapshot"`
}

type CreateSnapshotResponse struct {
	Snapshot HiscoreSnapshot `json:"snapshot"`
}

type GetSnapshotNearestTimestampResponse struct {
	Snapshot HiscoreSnapshot `json:"snapshot"`
}
type GetAllSnapshotsForUser struct {
	Snapshots []HiscoreSnapshot `json:"snapshots"`
}

type GetSnapshotIntervalRequest struct {
	UserId            string            `json:"userId"`
	StartTime         time.Time         `json:"startTime"`
	EndTime           time.Time         `json:"endTime"`
	AggregationWindow AggregationWindow `json:"aggregationWindow"`
}

type GetSnapshotIntervalResponse struct {
	Snapshots []HiscoreSnapshot `json:"snapshots"`
}
