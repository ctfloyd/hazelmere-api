package api

import "time"

type ActivityType string

const (
	ActivityTypeUnknown  ActivityType = "UNKNOWN"
	ActivityTypeOverall  ActivityType = "OVERALL"
	ActivityTypeJad      ActivityType = "JAD"
	ActivityTypeSoulWars ActivityType = "SOULWARS"
)

var AllActivityTypes = []ActivityType{
	ActivityTypeUnknown,
	ActivityTypeOverall,
	ActivityTypeJad,
	ActivityTypeSoulWars,
}

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

type GetAllHiscoreSnapshotsForUserResponse struct {
	Snapshots []HiscoreSnapshot `json:"snapshots"`
}
