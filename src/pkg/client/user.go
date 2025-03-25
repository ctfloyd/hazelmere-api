package client

import (
	"errors"
	"fmt"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
)

var ErrUserNotFound = errors.Join(ErrHazelmereClient, errors.New("user not found"))
var ErrInvalidUser = errors.Join(ErrHazelmereClient, errors.New("invalid user"))

type User struct {
	prefix string
	client *HazelmereClient
}

func newUser(client *HazelmereClient) *Snapshot {
	mappings := map[string]error{
		api.ErrorCodeUserNotFound:    ErrUserNotFound,
		api.ErrorCodeInvalidSnapshot: ErrInvalidUser,
	}
	client.AddErrorMappings(mappings)

	return &Snapshot{
		prefix: "user",
		client: client,
	}
}

func (user *User) GetAllUsers() (api.GetAllUsersResponse, error) {
	var response api.GetAllUsersResponse
	err := user.client.Get(user.getBaseUrl(), &response)
	if err != nil {
		return api.GetAllUsersResponse{}, err
	}
	return response, nil
}

func (user *User) GetUserById(id string) (api.GetUserByIdResponse, error) {
	url := fmt.Sprintf("%s/%s", user.getBaseUrl(), id)
	var response api.GetUserByIdResponse
	err := user.client.Get(url, &response)
	if err != nil {
		return api.GetUserByIdResponse{}, err
	}
	return response, nil
}

func (user *User) CreateUser(request api.CreateUserRequest) (api.CreateUserResponse, error) {
	var response api.CreateUserResponse
	err := user.client.Post(user.getBaseUrl(), request, &response)
	if err != nil {
		return api.CreateUserResponse{}, err
	}
	return response, nil
}

func (user *User) UpdateUser(request api.UpdateUserRequest) (api.UpdateUserResponse, error) {
	var response api.UpdateUserResponse
	err := user.client.Patch(user.getBaseUrl(), request, &response)
	if err != nil {
		return api.UpdateUserResponse{}, err
	}
	return response, nil
}

func (user *User) getBaseUrl() string {
	return fmt.Sprintf("%s/%s", user.client.GetV1Url(), user.prefix)
}
