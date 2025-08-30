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

func (hs HiscoreSnapshot) GetSkill(activityType ActivityType) SkillSnapshot {
	for _, skill := range hs.Skills {
		if skill.ActivityType == activityType {
			return skill
		}
	}
	return SkillSnapshot{}
}

func (hs HiscoreSnapshot) GetBoss(activityType ActivityType) BossSnapshot {
	for _, boss := range hs.Bosses {
		if boss.ActivityType == activityType {
			return boss
		}
	}
	return BossSnapshot{}
}

func (hs HiscoreSnapshot) GetActivity(activityType ActivityType) ActivitySnapshot {
	for _, activity := range hs.Activities {
		if activity.ActivityType == activityType {
			return activity
		}
	}
	return ActivitySnapshot{}
}

func (hs HiscoreSnapshot) Equals(other HiscoreSnapshot) bool {
	if hs.UserId != other.UserId {
		return false
	}

	for _, skill := range hs.Skills {
		if skill.ActivityType != ActivityTypeUnknown {
			otherSkill := other.GetSkill(skill.ActivityType)
			if !skill.Equals(otherSkill) {
				return false
			}
		}
	}

	for _, boss := range hs.Bosses {
		if boss.ActivityType != ActivityTypeUnknown {
			otherBoss := other.GetBoss(boss.ActivityType)
			if !boss.Equals(otherBoss) {
				return false
			}
		}
	}

	for _, activity := range hs.Activities {
		if activity.ActivityType != ActivityTypeUnknown {
			otherActivity := other.GetActivity(activity.ActivityType)
			if !activity.Equals(otherActivity) {
				return false
			}
		}
	}

	return true
}

type SkillSnapshot struct {
	ActivityType ActivityType
	Name         string
	Level        int
	Experience   int
	Rank         int
}

func (ss SkillSnapshot) Equals(other SkillSnapshot) bool {
	return ss.ActivityType == other.ActivityType &&
		ss.Level == other.Level &&
		ss.Experience == other.Experience
}

type BossSnapshot struct {
	ActivityType ActivityType
	Name         string
	KillCount    int
	Rank         int
}

func (bs BossSnapshot) Equals(other BossSnapshot) bool {
	return bs.ActivityType == other.ActivityType &&
		bs.Name == other.Name &&
		bs.KillCount == other.KillCount
}

type ActivitySnapshot struct {
	ActivityType ActivityType
	Name         string
	Score        int
	Rank         int
}

func (as ActivitySnapshot) Equals(other ActivitySnapshot) bool {
	return as.ActivityType == other.ActivityType &&
		as.Name == other.Name &&
		as.Score == other.Score
}
