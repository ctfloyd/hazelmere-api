package user

import (
	"api/src/internal/common/handler"
	"api/src/internal/common/logger"
	"api/src/internal/common/service_error"
	"api/src/pkg/api"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type UserHandler struct {
	logger  logger.Logger
	service UserService
}

func NewUserHandler(logger logger.Logger, service UserService) *UserHandler {
	return &UserHandler{logger, service}
}

func (uh *UserHandler) RegisterRoutes(mux *chi.Mux, version handler.ApiVersion) {
	if version == handler.ApiVersionV1 {
		mux.Get(fmt.Sprintf("/v1/user/{id:%s}", handler.RegexUuid), uh.GetUserById)
		mux.Get("/v1/user", uh.GetAllUsers)
		mux.Post("/v1/user", uh.CreateUser)
		mux.Put("/v1/user", uh.UpdateUser)
	}
}

func (uh *UserHandler) GetUserById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := uh.service.GetUserById(r.Context(), id)
	if err != nil {
		if errors.Is(ErrUserNotFound, err) {
			handler.Error(w, service_error.UserNotFound, "User not found.")
			return
		} else {
			handler.Error(w, service_error.Internal, "An unexpected error occurred while performing the user operation.")
			return
		}
	}

	response := api.GetUserByIdResponse{
		User: MapDomainToApi(user),
	}

	handler.Ok(w, response)
}

func (uh *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := uh.service.GetAllUsers(r.Context())
	if err != nil {
		handler.Error(w, service_error.Internal, "An unexpected error occurred while performing the user operation.")
		return
	}

	response := api.GetAllUsersResponse{
		Users: MapManyDomainToApi(users),
	}

	handler.Ok(w, response)
}

func (uh *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var createUserRequest api.CreateUserRequest
	if ok := handler.ReadBody(w, r, &createUserRequest); !ok {
		handler.Error(w, service_error.BadRequest, "Request body could not be read.")
		return
	}

	domainUser := MapCreateUserRequestToDomainUser(createUserRequest)

	user, err := uh.service.CreateUser(r.Context(), domainUser)
	if err != nil {
		if errors.Is(err, ErrRunescapeNameTracked) {
			handler.Error(w, service_error.RunescapeNameAlreadyTracked, "The runescape name is already associated with a user.")
		}

		handler.Error(w, service_error.Internal, "An unexpected error occurred while performing the user operation.")
		return
	}

	response := api.CreateUserResponse{
		User: MapDomainToApi(user),
	}

	handler.Ok(w, response)
}

func (uh *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var updateUserRequest api.UpdateUserRequest
	if ok := handler.ReadBody(w, r, &updateUserRequest); !ok {
		handler.Error(w, service_error.BadRequest, "Request body could not be read.")
		return
	}

	domainUser := MapUpdateUserRequestToDomainUser(updateUserRequest)

	user, err := uh.service.UpdateUser(r.Context(), domainUser)
	if err != nil {
		if errors.Is(err, ErrRunescapeNameTracked) {
			handler.Error(w, service_error.RunescapeNameAlreadyTracked, "The runescape name is already associated with a user.")
		}

		handler.Error(w, service_error.Internal, "An unexpected error occurred while performing the user operation.")
		return
	}

	response := api.CreateUserResponse{
		User: MapDomainToApi(user),
	}

	handler.Ok(w, response)
}
