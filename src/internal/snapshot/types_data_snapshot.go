package snapshot

import "time"

type HiscoreSnapshotData struct {
	Id         string                 `bson:"_id"`
	UserId     string                 `bson:"userId"`
	Timestamp  time.Time              `bson:"timestamp"`
	Skills     []SkillSnapshotData    `bson:"skills"`
	Bosses     []BossSnapshotData     `bson:"bosses"`
	Activities []ActivitySnapshotData `bson:"activities"`
}

type SkillSnapshotData struct {
	ActivityType string `bson:"activityType"`
	Name         string `bson:"name"`
	Level        int    `bson:"level"`
	Experience   int    `bson:"experience"`
	Rank         int    `bson:"rank"`
}

type BossSnapshotData struct {
	ActivityType string `bson:"activityType"`
	Name         string `bson:"name"`
	KillCount    int    `bson:"killCount"`
	Rank         int    `bson:"rank"`
}

type ActivitySnapshotData struct {
	ActivityType string `json:"activityType"`
	Name         string `json:"name"`
	Score        int    `json:"score"`
	Rank         int    `json:"rank"`
}
