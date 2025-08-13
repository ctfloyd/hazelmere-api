package user

type TrackingStatus string

const (
	TrackingStatusEnabled  TrackingStatus = "ENABLED"
	TrackingStatusDisabled TrackingStatus = "DISABLED"
)

func TrackingStatusFromValue(value string) TrackingStatus {
	if value == string(TrackingStatusEnabled) {
		return TrackingStatusEnabled
	}

	if value == string(TrackingStatusDisabled) {
		return TrackingStatusDisabled
	}

	return TrackingStatusEnabled
}

type AccountType string

const (
	AccountTypeNormal          AccountType = "NORMAL"
	AccountTypeIronman         AccountType = "IRONMAN"
	AccountTypeHardcoreIronman AccountType = "HARDCORE_IRONMAN"
	AccountTypeUltimateIronman AccountType = "ULTIMATE_IRONMAN"
	AccountTypeGroupIronman    AccountType = "GROUP_IRONMAN"
)

func AccountTypeFromValue(value string) AccountType {
	if value == string(AccountTypeNormal) {
		return AccountTypeNormal
	}

	if value == string(AccountTypeIronman) {
		return AccountTypeIronman
	}

	if value == string(AccountTypeHardcoreIronman) {
		return AccountTypeHardcoreIronman
	}

	if value == string(AccountTypeUltimateIronman) {
		return AccountTypeUltimateIronman
	}

	if value == string(AccountTypeGroupIronman) {
		return AccountTypeGroupIronman
	}

	return AccountTypeNormal
}

type User struct {
	Id             string         `json:"id"`
	RunescapeName  string         `json:"runescapeName"`
	TrackingStatus TrackingStatus `json:"trackingStatus"`
	AccountType    AccountType    `json:"accountType"`
}

func (u User) isTrackingEnabled() bool {
	return u.TrackingStatus == TrackingStatusEnabled
}

func (u User) isTrackingDisabled() bool {
	return u.TrackingStatus == TrackingStatusDisabled
}
