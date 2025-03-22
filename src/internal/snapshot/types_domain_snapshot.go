package snapshot

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

func ActivityTypeFromValue(value string) ActivityType {
	for _, at := range AllActivityTypes {
		if value == string(at) {
			return at
		}
	}
	return ActivityTypeUnknown
}

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
