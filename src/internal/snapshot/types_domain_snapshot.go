package snapshot

import "time"

type HiscoreSnapshot struct {
	Id         string
	UserId     string
	Timestamp  time.Time
	Skills     []SkillSnapshot
	Bosses     []BossSnapshot
	Activities []ActivitySnapshot
}

type SkillSnapshot struct {
	ActivityType ActivityType
	Name         string
	Level        int
	Experience   int
	Rank         int
}

type BossSnapshot struct {
	ActivityType ActivityType
	Name         string
	KillCount    int
	Rank         int
}

type ActivitySnapshot struct {
	ActivityType ActivityType
	Name         string
	Score        int
	Rank         int
}
