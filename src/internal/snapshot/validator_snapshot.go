package snapshot

import (
	"errors"
	"slices"
)

type SnapshotValidator interface {
	ValidateSnapshot(snapshot HiscoreSnapshot) error
}

type snapshotValidator struct {
}

func NewSnapshotValidator() SnapshotValidator {
	return &snapshotValidator{}
}

func (sv *snapshotValidator) ValidateSnapshot(snapshot HiscoreSnapshot) error {
	if snapshot.Timestamp.IsZero() {
		return errors.New("snapshot timestamp is zero")
	}

	if snapshot.UserId == "" {
		return errors.New("snapshot user id is empty")
	}

	if len(snapshot.Skills) < len(AllSkillActivityTypes) {
		return errors.New("snapshot must contain all skills")
	}

	for _, sk := range snapshot.Skills {
		if !slices.Contains(AllSkillActivityTypes, sk.ActivityType) {
			return errors.New(string(sk.ActivityType) + " is not a skill activity type")
		}
	}

	if len(snapshot.Bosses) < len(AllBossActivityTypes) {
		return errors.New("snapshot must contain all bosses")
	}

	for _, sk := range snapshot.Bosses {
		if !slices.Contains(AllBossActivityTypes, sk.ActivityType) {
			return errors.New(string(sk.ActivityType) + " is not a boss activity type")
		}
	}

	if len(snapshot.Activities) < len(AllActivityActivityTypes) {
		return errors.New("snapshot must contain all activities")
	}

	for _, sk := range snapshot.Activities {
		if !slices.Contains(AllActivityActivityTypes, sk.ActivityType) {
			return errors.New(string(sk.ActivityType) + " is not an activity activity type")
		}
	}

	return nil
}
