package snapshot

import (
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"time"
)

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

// ToAPI converts the domain HiscoreSnapshot to an API HiscoreSnapshot
func (hs HiscoreSnapshot) ToAPI() api.HiscoreSnapshot {
	skills := make([]api.SkillSnapshot, len(hs.Skills))
	for i := range hs.Skills {
		skills[i] = hs.Skills[i].ToAPI()
	}
	bosses := make([]api.BossSnapshot, len(hs.Bosses))
	for i := range hs.Bosses {
		bosses[i] = hs.Bosses[i].ToAPI()
	}
	activities := make([]api.ActivitySnapshot, len(hs.Activities))
	for i := range hs.Activities {
		activities[i] = hs.Activities[i].ToAPI()
	}
	return api.HiscoreSnapshot{
		Id:         hs.Id,
		UserId:     hs.UserId,
		Timestamp:  hs.Timestamp,
		Skills:     skills,
		Bosses:     bosses,
		Activities: activities,
	}
}

// FromAPI creates a domain HiscoreSnapshot from an API HiscoreSnapshot (call as HiscoreSnapshot{}.FromAPI(...))
func (HiscoreSnapshot) FromAPI(snapshot api.HiscoreSnapshot) HiscoreSnapshot {
	skills := make([]SkillSnapshot, len(snapshot.Skills))
	for i := range snapshot.Skills {
		skills[i] = SkillSnapshot{}.FromAPI(snapshot.Skills[i])
	}
	bosses := make([]BossSnapshot, len(snapshot.Bosses))
	for i := range snapshot.Bosses {
		bosses[i] = BossSnapshot{}.FromAPI(snapshot.Bosses[i])
	}
	activities := make([]ActivitySnapshot, len(snapshot.Activities))
	for i := range snapshot.Activities {
		activities[i] = ActivitySnapshot{}.FromAPI(snapshot.Activities[i])
	}
	return HiscoreSnapshot{
		Id:         snapshot.Id,
		UserId:     snapshot.UserId,
		Timestamp:  snapshot.Timestamp,
		Skills:     skills,
		Bosses:     bosses,
		Activities: activities,
	}
}

// ToData converts the domain HiscoreSnapshot to a data layer HiscoreSnapshotData
func (hs HiscoreSnapshot) ToData() HiscoreSnapshotData {
	skills := make([]SkillSnapshotData, len(hs.Skills))
	for i := range hs.Skills {
		skills[i] = hs.Skills[i].ToData()
	}
	bosses := make([]BossSnapshotData, len(hs.Bosses))
	for i := range hs.Bosses {
		bosses[i] = hs.Bosses[i].ToData()
	}
	activities := make([]ActivitySnapshotData, len(hs.Activities))
	for i := range hs.Activities {
		activities[i] = hs.Activities[i].ToData()
	}
	return HiscoreSnapshotData{
		Id:         hs.Id,
		UserId:     hs.UserId,
		Timestamp:  hs.Timestamp,
		Skills:     skills,
		Bosses:     bosses,
		Activities: activities,
	}
}

// FromData creates a domain HiscoreSnapshot from data layer HiscoreSnapshotData (call as HiscoreSnapshot{}.FromData(...))
func (HiscoreSnapshot) FromData(snapshot HiscoreSnapshotData) HiscoreSnapshot {
	skills := make([]SkillSnapshot, len(snapshot.Skills))
	for i := range snapshot.Skills {
		skills[i] = SkillSnapshot{}.FromData(snapshot.Skills[i])
	}
	bosses := make([]BossSnapshot, len(snapshot.Bosses))
	for i := range snapshot.Bosses {
		bosses[i] = BossSnapshot{}.FromData(snapshot.Bosses[i])
	}
	activities := make([]ActivitySnapshot, len(snapshot.Activities))
	for i := range snapshot.Activities {
		activities[i] = ActivitySnapshot{}.FromData(snapshot.Activities[i])
	}
	return HiscoreSnapshot{
		Id:         snapshot.Id,
		UserId:     snapshot.UserId,
		Timestamp:  snapshot.Timestamp,
		Skills:     skills,
		Bosses:     bosses,
		Activities: activities,
	}
}

// ManyToAPI converts a slice of domain HiscoreSnapshots to API HiscoreSnapshots (call as HiscoreSnapshot{}.ManyToAPI(...))
func (HiscoreSnapshot) ManyToAPI(snapshots []HiscoreSnapshot) []api.HiscoreSnapshot {
	apiSnapshots := make([]api.HiscoreSnapshot, len(snapshots))
	for i := range snapshots {
		apiSnapshots[i] = snapshots[i].ToAPI()
	}
	return apiSnapshots
}

// ManyFromData converts a slice of HiscoreSnapshotData to domain HiscoreSnapshots (call as HiscoreSnapshot{}.ManyFromData(...))
func (HiscoreSnapshot) ManyFromData(data []HiscoreSnapshotData) []HiscoreSnapshot {
	snapshots := make([]HiscoreSnapshot, len(data))
	for i := range data {
		snapshots[i] = HiscoreSnapshot{}.FromData(data[i])
	}
	return snapshots
}

// ToAPI converts the domain SkillSnapshot to an API SkillSnapshot
func (ss SkillSnapshot) ToAPI() api.SkillSnapshot {
	return api.SkillSnapshot{
		ActivityType: ss.ActivityType.ToAPI(),
		Name:         ss.Name,
		Level:        ss.Level,
		Experience:   ss.Experience,
		Rank:         ss.Rank,
	}
}

// FromAPI creates a domain SkillSnapshot from an API SkillSnapshot (call as SkillSnapshot{}.FromAPI(...))
func (SkillSnapshot) FromAPI(skill api.SkillSnapshot) SkillSnapshot {
	return SkillSnapshot{
		ActivityType: ActivityType("").FromAPI(skill.ActivityType),
		Name:         skill.Name,
		Level:        skill.Level,
		Experience:   skill.Experience,
		Rank:         skill.Rank,
	}
}

// ToData converts the domain SkillSnapshot to a data layer SkillSnapshotData
func (ss SkillSnapshot) ToData() SkillSnapshotData {
	return SkillSnapshotData{
		ActivityType: string(ss.ActivityType),
		Name:         ss.Name,
		Level:        ss.Level,
		Experience:   ss.Experience,
		Rank:         ss.Rank,
	}
}

// FromData creates a domain SkillSnapshot from data layer SkillSnapshotData (call as SkillSnapshot{}.FromData(...))
func (SkillSnapshot) FromData(skill SkillSnapshotData) SkillSnapshot {
	return SkillSnapshot{
		ActivityType: ActivityTypeFromValue(skill.ActivityType),
		Name:         skill.Name,
		Level:        skill.Level,
		Experience:   skill.Experience,
		Rank:         skill.Rank,
	}
}

// ToAPI converts the domain BossSnapshot to an API BossSnapshot
func (bs BossSnapshot) ToAPI() api.BossSnapshot {
	return api.BossSnapshot{
		ActivityType: bs.ActivityType.ToAPI(),
		Name:         bs.Name,
		KillCount:    bs.KillCount,
		Rank:         bs.Rank,
	}
}

// FromAPI creates a domain BossSnapshot from an API BossSnapshot (call as BossSnapshot{}.FromAPI(...))
func (BossSnapshot) FromAPI(boss api.BossSnapshot) BossSnapshot {
	return BossSnapshot{
		ActivityType: ActivityType("").FromAPI(boss.ActivityType),
		Name:         boss.Name,
		KillCount:    boss.KillCount,
		Rank:         boss.Rank,
	}
}

// ToData converts the domain BossSnapshot to a data layer BossSnapshotData
func (bs BossSnapshot) ToData() BossSnapshotData {
	return BossSnapshotData{
		ActivityType: string(bs.ActivityType),
		Name:         bs.Name,
		KillCount:    bs.KillCount,
		Rank:         bs.Rank,
	}
}

// FromData creates a domain BossSnapshot from data layer BossSnapshotData (call as BossSnapshot{}.FromData(...))
func (BossSnapshot) FromData(boss BossSnapshotData) BossSnapshot {
	return BossSnapshot{
		ActivityType: ActivityTypeFromValue(boss.ActivityType),
		Name:         boss.Name,
		KillCount:    boss.KillCount,
		Rank:         boss.Rank,
	}
}

// ToAPI converts the domain ActivitySnapshot to an API ActivitySnapshot
func (as ActivitySnapshot) ToAPI() api.ActivitySnapshot {
	return api.ActivitySnapshot{
		ActivityType: as.ActivityType.ToAPI(),
		Name:         as.Name,
		Score:        as.Score,
		Rank:         as.Rank,
	}
}

// FromAPI creates a domain ActivitySnapshot from an API ActivitySnapshot (call as ActivitySnapshot{}.FromAPI(...))
func (ActivitySnapshot) FromAPI(activity api.ActivitySnapshot) ActivitySnapshot {
	return ActivitySnapshot{
		ActivityType: ActivityType("").FromAPI(activity.ActivityType),
		Name:         activity.Name,
		Score:        activity.Score,
		Rank:         activity.Rank,
	}
}

// ToData converts the domain ActivitySnapshot to a data layer ActivitySnapshotData
func (as ActivitySnapshot) ToData() ActivitySnapshotData {
	return ActivitySnapshotData{
		ActivityType: string(as.ActivityType),
		Name:         as.Name,
		Score:        as.Score,
		Rank:         as.Rank,
	}
}

// FromData creates a domain ActivitySnapshot from data layer ActivitySnapshotData (call as ActivitySnapshot{}.FromData(...))
func (ActivitySnapshot) FromData(activity ActivitySnapshotData) ActivitySnapshot {
	return ActivitySnapshot{
		ActivityType: ActivityTypeFromValue(activity.ActivityType),
		Name:         activity.Name,
		Score:        activity.Score,
		Rank:         activity.Rank,
	}
}

// ToAPI converts domain ActivityType to API ActivityType
func (at ActivityType) ToAPI() api.ActivityType {
	for _, apiAt := range api.AllActivityTypes {
		if string(apiAt) == string(at) {
			return apiAt
		}
	}
	return api.ActivityTypeUnknown
}

// FromAPI converts API ActivityType to domain ActivityType (call as ActivityType("").FromAPI(...))
func (ActivityType) FromAPI(activityType api.ActivityType) ActivityType {
	for _, activity := range AllActivityTypes {
		if string(activityType) == string(activity) {
			return activity
		}
	}
	return ActivityTypeUnknown
}
