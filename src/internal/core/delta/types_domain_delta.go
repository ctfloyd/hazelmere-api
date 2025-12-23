package delta

import (
	"time"

	"github.com/ctfloyd/hazelmere-api/src/internal/core/snapshot"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
)

type HiscoreDelta struct {
	Id                 string
	UserId             string
	SnapshotId         string
	PreviousSnapshotId string
	Timestamp          time.Time
	Skills             []SkillDelta
	Bosses             []BossDelta
	Activities         []ActivityDelta
}

type SkillDelta struct {
	ActivityType   snapshot.ActivityType
	Name           string
	ExperienceGain int
	LevelGain      int
}

type BossDelta struct {
	ActivityType  snapshot.ActivityType
	Name          string
	KillCountGain int
}

type ActivityDelta struct {
	ActivityType snapshot.ActivityType
	Name         string
	ScoreGain    int
}

// ToAPI converts the domain HiscoreDelta to an API HiscoreDelta
func (hd HiscoreDelta) ToAPI() api.HiscoreDelta {
	skills := make([]api.SkillDelta, len(hd.Skills))
	for i := range hd.Skills {
		skills[i] = api.SkillDelta{
			ActivityType:   hd.Skills[i].ActivityType.ToAPI(),
			Name:           hd.Skills[i].Name,
			ExperienceGain: hd.Skills[i].ExperienceGain,
			LevelGain:      hd.Skills[i].LevelGain,
		}
	}
	bosses := make([]api.BossDelta, len(hd.Bosses))
	for i := range hd.Bosses {
		bosses[i] = api.BossDelta{
			ActivityType:  hd.Bosses[i].ActivityType.ToAPI(),
			Name:          hd.Bosses[i].Name,
			KillCountGain: hd.Bosses[i].KillCountGain,
		}
	}
	activities := make([]api.ActivityDelta, len(hd.Activities))
	for i := range hd.Activities {
		activities[i] = api.ActivityDelta{
			ActivityType: hd.Activities[i].ActivityType.ToAPI(),
			Name:         hd.Activities[i].Name,
			ScoreGain:    hd.Activities[i].ScoreGain,
		}
	}
	return api.HiscoreDelta{
		Id:                 hd.Id,
		UserId:             hd.UserId,
		SnapshotId:         hd.SnapshotId,
		PreviousSnapshotId: hd.PreviousSnapshotId,
		Timestamp:          hd.Timestamp,
		Skills:             skills,
		Bosses:             bosses,
		Activities:         activities,
	}
}

// ToData converts the domain HiscoreDelta to a data layer HiscoreDeltaData
func (hd HiscoreDelta) ToData() HiscoreDeltaData {
	skills := make([]SkillDeltaData, len(hd.Skills))
	for i := range hd.Skills {
		skills[i] = hd.Skills[i].ToData()
	}
	bosses := make([]BossDeltaData, len(hd.Bosses))
	for i := range hd.Bosses {
		bosses[i] = hd.Bosses[i].ToData()
	}
	activities := make([]ActivityDeltaData, len(hd.Activities))
	for i := range hd.Activities {
		activities[i] = hd.Activities[i].ToData()
	}
	return HiscoreDeltaData{
		Id:                 hd.Id,
		UserId:             hd.UserId,
		SnapshotId:         hd.SnapshotId,
		PreviousSnapshotId: hd.PreviousSnapshotId,
		Timestamp:          hd.Timestamp,
		Skills:             skills,
		Bosses:             bosses,
		Activities:         activities,
	}
}

// FromData creates a domain HiscoreDelta from data layer HiscoreDeltaData
func (HiscoreDelta) FromData(data HiscoreDeltaData) HiscoreDelta {
	skills := make([]SkillDelta, len(data.Skills))
	for i := range data.Skills {
		skills[i] = SkillDelta{}.FromData(data.Skills[i])
	}
	bosses := make([]BossDelta, len(data.Bosses))
	for i := range data.Bosses {
		bosses[i] = BossDelta{}.FromData(data.Bosses[i])
	}
	activities := make([]ActivityDelta, len(data.Activities))
	for i := range data.Activities {
		activities[i] = ActivityDelta{}.FromData(data.Activities[i])
	}
	return HiscoreDelta{
		Id:                 data.Id,
		UserId:             data.UserId,
		SnapshotId:         data.SnapshotId,
		PreviousSnapshotId: data.PreviousSnapshotId,
		Timestamp:          data.Timestamp,
		Skills:             skills,
		Bosses:             bosses,
		Activities:         activities,
	}
}

// ManyToAPI converts a slice of domain HiscoreDeltas to API HiscoreDeltas
func (HiscoreDelta) ManyToAPI(deltas []HiscoreDelta) []api.HiscoreDelta {
	apiDeltas := make([]api.HiscoreDelta, len(deltas))
	for i := range deltas {
		apiDeltas[i] = deltas[i].ToAPI()
	}
	return apiDeltas
}

// ManyFromData converts a slice of HiscoreDeltaData to domain HiscoreDeltas
func (HiscoreDelta) ManyFromData(data []HiscoreDeltaData) []HiscoreDelta {
	deltas := make([]HiscoreDelta, len(data))
	for i := range data {
		deltas[i] = HiscoreDelta{}.FromData(data[i])
	}
	return deltas
}

// ToData converts the domain SkillDelta to a data layer SkillDeltaData
func (sd SkillDelta) ToData() SkillDeltaData {
	return SkillDeltaData{
		ActivityType:   string(sd.ActivityType),
		Name:           sd.Name,
		ExperienceGain: sd.ExperienceGain,
		LevelGain:      sd.LevelGain,
	}
}

// FromData creates a domain SkillDelta from data layer SkillDeltaData
func (SkillDelta) FromData(data SkillDeltaData) SkillDelta {
	return SkillDelta{
		ActivityType:   snapshot.ActivityTypeFromValue(data.ActivityType),
		Name:           data.Name,
		ExperienceGain: data.ExperienceGain,
		LevelGain:      data.LevelGain,
	}
}

// ToData converts the domain BossDelta to a data layer BossDeltaData
func (bd BossDelta) ToData() BossDeltaData {
	return BossDeltaData{
		ActivityType:  string(bd.ActivityType),
		Name:          bd.Name,
		KillCountGain: bd.KillCountGain,
	}
}

// FromData creates a domain BossDelta from data layer BossDeltaData
func (BossDelta) FromData(data BossDeltaData) BossDelta {
	return BossDelta{
		ActivityType:  snapshot.ActivityTypeFromValue(data.ActivityType),
		Name:          data.Name,
		KillCountGain: data.KillCountGain,
	}
}

// ToData converts the domain ActivityDelta to a data layer ActivityDeltaData
func (ad ActivityDelta) ToData() ActivityDeltaData {
	return ActivityDeltaData{
		ActivityType: string(ad.ActivityType),
		Name:         ad.Name,
		ScoreGain:    ad.ScoreGain,
	}
}

// FromData creates a domain ActivityDelta from data layer ActivityDeltaData
func (ActivityDelta) FromData(data ActivityDeltaData) ActivityDelta {
	return ActivityDelta{
		ActivityType: snapshot.ActivityTypeFromValue(data.ActivityType),
		Name:         data.Name,
		ScoreGain:    data.ScoreGain,
	}
}
