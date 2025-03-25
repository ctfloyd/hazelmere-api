package user

import "api/src/pkg/api"

func MapCreateUserRequestToDomainUser(request api.CreateUserRequest) User {
	return User{
		RunescapeName:  request.RunescapeName,
		TrackingStatus: TrackingStatusFromValue(string(request.TrackingStatus)),
	}
}

func MapUpdateUserRequestToDomainUser(request api.UpdateUserRequest) User {
	return User{
		Id:             request.Id,
		RunescapeName:  request.RunescapeName,
		TrackingStatus: TrackingStatusFromValue(string(request.TrackingStatus)),
	}

}

func MapManyDomainToApi(users []User) []api.User {
	apiUsers := make([]api.User, len(users))
	for i := 0; i < len(users); i++ {
		apiUsers[i] = MapDomainToApi(users[i])
	}
	return apiUsers
}

func MapDomainToApi(user User) api.User {
	return api.User{
		Id:             user.Id,
		RunescapeName:  user.RunescapeName,
		TrackingStatus: api.TrackingStatusFromValue(string(user.TrackingStatus)),
	}
}

func MapDomainToData(user User) UserData {
	return UserData{
		Id:             user.Id,
		RunescapeName:  user.RunescapeName,
		TrackingStatus: string(user.TrackingStatus),
	}
}

func MapManyDataToDomain(userData []UserData) []User {
	users := make([]User, len(userData))
	for i := 0; i < len(userData); i++ {
		users[i] = MapDataToDomain(userData[i])
	}
	return users
}

func MapDataToDomain(userData UserData) User {
	return User{
		Id:             userData.Id,
		RunescapeName:  userData.RunescapeName,
		TrackingStatus: TrackingStatusFromValue(userData.TrackingStatus),
	}
}
