package delta

import "time"

type HiscoreDeltaData struct {
	Id                 string              `bson:"_id"`
	UserId             string              `bson:"userId"`
	SnapshotId         string              `bson:"snapshotId"`
	PreviousSnapshotId string              `bson:"previousSnapshotId"`
	Timestamp          time.Time           `bson:"timestamp"`
	Skills             []SkillDeltaData    `bson:"skills,omitempty"`
	Bosses             []BossDeltaData     `bson:"bosses,omitempty"`
	Activities         []ActivityDeltaData `bson:"activities,omitempty"`
}

type SkillDeltaData struct {
	ActivityType   string `bson:"activityType"`
	Name           string `bson:"name"`
	ExperienceGain int    `bson:"experienceGain"`
	LevelGain      int    `bson:"levelGain"`
}

type BossDeltaData struct {
	ActivityType  string `bson:"activityType"`
	Name          string `bson:"name"`
	KillCountGain int    `bson:"killCountGain"`
}

type ActivityDeltaData struct {
	ActivityType string `bson:"activityType"`
	Name         string `bson:"name"`
	ScoreGain    int    `bson:"scoreGain"`
}
