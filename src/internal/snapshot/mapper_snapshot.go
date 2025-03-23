package snapshot

import "api/src/pkg/api"

func MapApiToDomain(snapshot api.HiscoreSnapshot) HiscoreSnapshot {
	return HiscoreSnapshot{
		Id:         snapshot.Id,
		UserId:     snapshot.UserId,
		Timestamp:  snapshot.Timestamp,
		Skills:     mapApiSkillsToDomainSkills(snapshot.Skills),
		Bosses:     mapApiBossesToDomainBosses(snapshot.Bosses),
		Activities: mapApiActivitiesToDomainActivities(snapshot.Activities),
	}
}

func MapManyDomainToApi(snapshot []HiscoreSnapshot) []api.HiscoreSnapshot {
	apiSnapshots := make([]api.HiscoreSnapshot, len(snapshot))
	for i := 0; i < len(snapshot); i++ {
		apiSnapshots[i] = MapDomainToApi(snapshot[i])
	}
	return apiSnapshots
}

func MapDomainToApi(snapshot HiscoreSnapshot) api.HiscoreSnapshot {
	return api.HiscoreSnapshot{
		Id:         snapshot.Id,
		UserId:     snapshot.UserId,
		Timestamp:  snapshot.Timestamp,
		Skills:     mapDomainSkillsToApiSkills(snapshot.Skills),
		Bosses:     mapDomainBossesToApiBosses(snapshot.Bosses),
		Activities: mapDomainActivitiesToApiActivities(snapshot.Activities),
	}

}

func MapDomainToData(snapshot HiscoreSnapshot) HiscoreSnapshotData {
	return HiscoreSnapshotData{
		Id:         snapshot.Id,
		UserId:     snapshot.UserId,
		Timestamp:  snapshot.Timestamp,
		Skills:     mapDomainSkillsToDataSkills(snapshot.Skills),
		Bosses:     mapDomainBossesToDataBosses(snapshot.Bosses),
		Activities: mapDomainActivitiesToDataActivities(snapshot.Activities),
	}
}

func mapDomainSkillsToDataSkills(skills []SkillSnapshot) []SkillSnapshotData {
	skillsData := make([]SkillSnapshotData, len(skills))
	for i := 0; i < len(skills); i++ {
		skillsData[i] = mapDomainSkillToDataSkill(skills[i])
	}
	return skillsData
}

func mapDomainSkillToDataSkill(skill SkillSnapshot) SkillSnapshotData {
	return SkillSnapshotData{
		ActivityType: string(skill.ActivityType),
		Name:         skill.Name,
		Level:        skill.Level,
		Experience:   skill.Experience,
		Rank:         skill.Rank,
	}
}

func mapDomainBossesToDataBosses(bosses []BossSnapshot) []BossSnapshotData {
	bossData := make([]BossSnapshotData, len(bosses))
	for i := 0; i < len(bosses); i++ {
		bossData[i] = mapDomainBossToDataBoss(bosses[i])
	}
	return bossData
}

func mapDomainBossToDataBoss(boss BossSnapshot) BossSnapshotData {
	return BossSnapshotData{
		ActivityType: string(boss.ActivityType),
		Name:         boss.Name,
		KillCount:    boss.KillCount,
		Rank:         boss.Rank,
	}
}

func mapDomainActivitiesToDataActivities(activities []ActivitySnapshot) []ActivitySnapshotData {
	activityData := make([]ActivitySnapshotData, len(activities))
	for i := 0; i < len(activities); i++ {
		activityData[i] = mapDomainActivityToDataActivity(activities[i])
	}
	return activityData
}

func mapDomainActivityToDataActivity(activity ActivitySnapshot) ActivitySnapshotData {
	return ActivitySnapshotData{
		ActivityType: string(activity.ActivityType),
		Name:         activity.Name,
		Score:        activity.Score,
		Rank:         activity.Rank,
	}
}

func MapManyDataToDomain(snapshots []HiscoreSnapshotData) []HiscoreSnapshot {
	domainSnapshots := make([]HiscoreSnapshot, len(snapshots))
	for i := 0; i < len(snapshots); i++ {
		domainSnapshots[i] = MapDataToDomain(snapshots[i])
	}
	return domainSnapshots
}

func MapDataToDomain(snapshot HiscoreSnapshotData) HiscoreSnapshot {
	return HiscoreSnapshot{
		Id:         snapshot.Id,
		UserId:     snapshot.UserId,
		Timestamp:  snapshot.Timestamp,
		Skills:     mapDataSkillsToDomainSkills(snapshot.Skills),
		Bosses:     mapDataBossesToDomainBosses(snapshot.Bosses),
		Activities: mapDataActivitiesToDomainActivities(snapshot.Activities),
	}
}

func mapApiSkillsToDomainSkills(skills []api.SkillSnapshot) []SkillSnapshot {
	domain := make([]SkillSnapshot, len(skills))
	for i := 0; i < len(skills); i++ {
		domain[i] = mapApiSkillToDomainSkill(skills[i])
	}
	return domain
}

func mapApiSkillToDomainSkill(skill api.SkillSnapshot) SkillSnapshot {
	return SkillSnapshot{
		ActivityType: mapApiActivityTypeToDomainActivityType(skill.ActivityType),
		Name:         skill.Name,
		Level:        skill.Level,
		Experience:   skill.Experience,
		Rank:         skill.Rank,
	}
}

func mapApiBossesToDomainBosses(bosses []api.BossSnapshot) []BossSnapshot {
	domain := make([]BossSnapshot, len(bosses))
	for i := 0; i < len(bosses); i++ {
		domain[i] = mapApiBossToDomainBoss(bosses[i])
	}
	return domain
}

func mapApiBossToDomainBoss(boss api.BossSnapshot) BossSnapshot {
	return BossSnapshot{
		ActivityType: mapApiActivityTypeToDomainActivityType(boss.ActivityType),
		Name:         boss.Name,
		KillCount:    boss.KillCount,
		Rank:         boss.Rank,
	}
}

func mapApiActivitiesToDomainActivities(activities []api.ActivitySnapshot) []ActivitySnapshot {
	domain := make([]ActivitySnapshot, len(activities))
	for i := 0; i < len(activities); i++ {
		domain[i] = mapApiActivityToDomainActivity(activities[i])
	}
	return domain
}

func mapApiActivityToDomainActivity(activity api.ActivitySnapshot) ActivitySnapshot {
	return ActivitySnapshot{
		ActivityType: mapApiActivityTypeToDomainActivityType(activity.ActivityType),
		Name:         activity.Name,
		Score:        activity.Score,
		Rank:         activity.Rank,
	}
}

func mapApiActivityTypeToDomainActivityType(activityType api.ActivityType) ActivityType {
	for _, activity := range AllActivityTypes {
		if string(activityType) == string(activity) {
			return activity
		}
	}
	return ActivityTypeUnknown
}

func mapDomainSkillsToApiSkills(skills []SkillSnapshot) []api.SkillSnapshot {
	apiSkills := make([]api.SkillSnapshot, len(skills))
	for i := 0; i < len(skills); i++ {
		apiSkills[i] = mapDomainSkillToApiSkill(skills[i])
	}
	return apiSkills
}

func mapDomainSkillToApiSkill(skill SkillSnapshot) api.SkillSnapshot {
	return api.SkillSnapshot{
		ActivityType: mapDomainActivityTypeToApiActivityType(skill.ActivityType),
		Name:         skill.Name,
		Level:        skill.Level,
		Experience:   skill.Experience,
		Rank:         skill.Rank,
	}
}

func mapDomainBossesToApiBosses(bosses []BossSnapshot) []api.BossSnapshot {
	apiBosses := make([]api.BossSnapshot, len(bosses))
	for i := 0; i < len(bosses); i++ {
		apiBosses[i] = mapDomainBossToApiBoss(bosses[i])
	}
	return apiBosses
}

func mapDomainBossToApiBoss(boss BossSnapshot) api.BossSnapshot {
	return api.BossSnapshot{
		ActivityType: mapDomainActivityTypeToApiActivityType(boss.ActivityType),
		Name:         boss.Name,
		KillCount:    boss.KillCount,
		Rank:         boss.Rank,
	}
}

func mapDomainActivitiesToApiActivities(activities []ActivitySnapshot) []api.ActivitySnapshot {
	apiActivities := make([]api.ActivitySnapshot, len(activities))
	for i := 0; i < len(activities); i++ {
		apiActivities[i] = mapDomainActivityToApiActivity(activities[i])
	}
	return apiActivities
}

func mapDomainActivityToApiActivity(activity ActivitySnapshot) api.ActivitySnapshot {
	return api.ActivitySnapshot{
		ActivityType: mapDomainActivityTypeToApiActivityType(activity.ActivityType),
		Name:         activity.Name,
		Score:        activity.Score,
		Rank:         activity.Rank,
	}
}

func mapDomainActivityTypeToApiActivityType(domainActivityType ActivityType) api.ActivityType {
	for _, at := range api.AllActivityTypes {
		if string(at) == string(domainActivityType) {
			return at
		}
	}
	return api.ActivityTypeUnknown
}

func mapDataSkillsToDomainSkills(skills []SkillSnapshotData) []SkillSnapshot {
	domainSkills := make([]SkillSnapshot, len(skills))
	for i := 0; i < len(skills); i++ {
		domainSkills[i] = mapDataSkillToDomainSkill(skills[i])
	}
	return domainSkills
}

func mapDataSkillToDomainSkill(skill SkillSnapshotData) SkillSnapshot {
	return SkillSnapshot{
		ActivityType: ActivityTypeFromValue(skill.ActivityType),
		Name:         skill.Name,
		Level:        skill.Level,
		Experience:   skill.Experience,
		Rank:         skill.Rank,
	}
}

func mapDataBossesToDomainBosses(bosses []BossSnapshotData) []BossSnapshot {
	domainBosses := make([]BossSnapshot, len(bosses))
	for i := 0; i < len(bosses); i++ {
		domainBosses[i] = mapDataBossToDomainBoss(bosses[i])
	}
	return domainBosses
}

func mapDataBossToDomainBoss(boss BossSnapshotData) BossSnapshot {
	return BossSnapshot{
		ActivityType: ActivityTypeFromValue(boss.ActivityType),
		Name:         boss.Name,
		KillCount:    boss.KillCount,
		Rank:         boss.Rank,
	}
}

func mapDataActivitiesToDomainActivities(activities []ActivitySnapshotData) []ActivitySnapshot {
	domainActivities := make([]ActivitySnapshot, len(activities))
	for i := 0; i < len(activities); i++ {
		domainActivities[i] = mapDataActivityToDomainActivity(activities[i])
	}
	return domainActivities
}

func mapDataActivityToDomainActivity(activity ActivitySnapshotData) ActivitySnapshot {
	return ActivitySnapshot{
		ActivityType: ActivityTypeFromValue(activity.ActivityType),
		Name:         activity.Name,
		Score:        activity.Score,
		Rank:         activity.Rank,
	}
}
