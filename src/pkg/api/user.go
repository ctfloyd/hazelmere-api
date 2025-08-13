package api

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

func (u *User) IsTrackingEnabled() bool {
	return u.TrackingStatus == TrackingStatusEnabled
}

func (u *User) IsTrackingDisabled() bool {
	return u.TrackingStatus == TrackingStatusDisabled
}

func (u *User) IsNormal() bool {
	return u.AccountType == AccountTypeNormal
}

func (u *User) IsIronman() bool {
	return u.AccountType == AccountTypeIronman
}

func (u *User) IsHardcoreIronman() bool {
	return u.AccountType == AccountTypeHardcoreIronman
}

func (u *User) IsUltimateIronman() bool {
	return u.AccountType == AccountTypeUltimateIronman
}

func (u *User) IsGroupIronman() bool {
	return u.AccountType == AccountTypeGroupIronman
}

type GetAllUsersResponse struct {
	Users []User `json:"users"`
}

type GetUserByIdResponse struct {
	User User `json:"user"`
}

type CreateUserRequest struct {
	RunescapeName  string         `json:"runescapeName"`
	TrackingStatus TrackingStatus `json:"trackingStatus"`
	AccountType    AccountType    `json:"accountType"`
}
type CreateUserResponse struct {
	User User `json:"user"`
}

type UpdateUserRequest struct {
	Id             string         `json:"id"`
	RunescapeName  string         `json:"runescapeName"`
	TrackingStatus TrackingStatus `json:"trackingStatus"`
	AccountType    AccountType    `json:"accountType"`
}
type UpdateUserResponse struct {
	User User `json:"user"`
}
