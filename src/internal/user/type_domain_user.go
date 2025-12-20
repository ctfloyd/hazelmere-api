package user

import "github.com/ctfloyd/hazelmere-api/src/pkg/api"

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

// ToAPI converts the domain User to an API User
func (u User) ToAPI() api.User {
	return api.User{
		Id:             u.Id,
		RunescapeName:  u.RunescapeName,
		TrackingStatus: api.TrackingStatusFromValue(string(u.TrackingStatus)),
		AccountType:    api.AccountTypeFromValue(string(u.AccountType)),
	}
}

// FromAPI creates a domain User from an API User (call as User{}.FromAPI(...))
func (User) FromAPI(user api.User) User {
	return User{
		Id:             user.Id,
		RunescapeName:  user.RunescapeName,
		TrackingStatus: TrackingStatusFromValue(string(user.TrackingStatus)),
		AccountType:    AccountTypeFromValue(string(user.AccountType)),
	}
}

// FromCreateRequest creates a domain User from a CreateUserRequest (call as User{}.FromCreateRequest(...))
func (User) FromCreateRequest(request api.CreateUserRequest) User {
	return User{
		RunescapeName:  request.RunescapeName,
		TrackingStatus: TrackingStatusFromValue(string(request.TrackingStatus)),
		AccountType:    AccountTypeFromValue(string(request.AccountType)),
	}
}

// FromUpdateRequest creates a domain User from an UpdateUserRequest (call as User{}.FromUpdateRequest(...))
func (User) FromUpdateRequest(request api.UpdateUserRequest) User {
	return User{
		Id:             request.Id,
		RunescapeName:  request.RunescapeName,
		TrackingStatus: TrackingStatusFromValue(string(request.TrackingStatus)),
		AccountType:    AccountTypeFromValue(string(request.AccountType)),
	}
}

// ToData converts the domain User to a data layer UserData
func (u User) ToData() UserData {
	return UserData{
		Id:             u.Id,
		RunescapeName:  u.RunescapeName,
		TrackingStatus: string(u.TrackingStatus),
		AccountType:    string(u.AccountType),
	}
}

// FromData creates a domain User from data layer UserData (call as User{}.FromData(...))
func (User) FromData(userData UserData) User {
	return User{
		Id:             userData.Id,
		RunescapeName:  userData.RunescapeName,
		TrackingStatus: TrackingStatusFromValue(userData.TrackingStatus),
		AccountType:    AccountTypeFromValue(userData.AccountType),
	}
}

// ManyToAPI converts a slice of domain Users to API Users (call as User{}.ManyToAPI(...))
func (User) ManyToAPI(users []User) []api.User {
	apiUsers := make([]api.User, len(users))
	for i := range users {
		apiUsers[i] = users[i].ToAPI()
	}
	return apiUsers
}

// ManyFromData converts a slice of UserData to domain Users (call as User{}.ManyFromData(...))
func (User) ManyFromData(userData []UserData) []User {
	users := make([]User, len(userData))
	for i := range userData {
		users[i] = User{}.FromData(userData[i])
	}
	return users
}
