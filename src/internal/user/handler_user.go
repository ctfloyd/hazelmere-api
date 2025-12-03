package user

import (
	"errors"
	"fmt"
	"github.com/ctfloyd/hazelmere-api/src/internal/common/handler"
	"github.com/ctfloyd/hazelmere-api/src/internal/middleware"
	"github.com/ctfloyd/hazelmere-api/src/internal/service_error"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_handler"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"github.com/go-chi/chi/v5"
	chiWare "github.com/go-chi/chi/v5/middleware"
	"net/http"
	"time"
)

type UserHandler struct {
	logger  hz_logger.Logger
	service UserService
}

func NewUserHandler(logger hz_logger.Logger, service UserService) *UserHandler {
	return &UserHandler{logger, service}
}

func (uh *UserHandler) RegisterRoutes(mux *chi.Mux, version handler.ApiVersion, authorizer *middleware.Authorizer) {
	if version == handler.ApiVersionV1 {
		mux.Group(func(r chi.Router) {
			r.Use(chiWare.Timeout(5000 * time.Millisecond))
			r.Get(fmt.Sprintf("/v1/user/{id:%s}", hz_handler.RegexUuid), uh.GetUserById)
			r.Get("/v1/user", uh.GetAllUsers)
			r.Group(func(secure chi.Router) {
				secure.Use(authorizer.Authorize)
				secure.Post("/v1/user", uh.CreateUser)
				secure.Put("/v1/user", uh.UpdateUser)
			})
		})
	}
}

func (uh *UserHandler) GetUserById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	uh.logger.InfoArgs(r.Context(), "Getting user by id: %s", id)

	user, err := uh.service.GetUserById(r.Context(), id)
	if err != nil {
		if errors.Is(ErrUserNotFound, err) {
			uh.logger.WarnArgs(r.Context(), "User not found: %s", id)
			hz_handler.Error(w, service_error.UserNotFound, "User not found.")
			return
		} else {
			uh.logger.ErrorArgs(r.Context(), "An unexpected error occurred while getting user by id: %+v", err)
			hz_handler.Error(w, service_error.Internal, "An unexpected service_error occurred while performing the user operation.")
			return
		}
	}

	response := api.GetUserByIdResponse{
		User: MapDomainToApi(user),
	}

	hz_handler.Ok(w, response)
}

func (uh *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	uh.logger.Info(r.Context(), "Getting all users")

	users, err := uh.service.GetAllUsers(r.Context())
	if err != nil {
		uh.logger.ErrorArgs(r.Context(), "An unexpected error occurred while getting all users: %+v", err)
		hz_handler.Error(w, service_error.Internal, "An unexpected service_error occurred while performing the user operation.")
		return
	}

	response := api.GetAllUsersResponse{
		Users: MapManyDomainToApi(users),
	}

	hz_handler.Ok(w, response)
}

func (uh *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var createUserRequest api.CreateUserRequest
	if ok := hz_handler.ReadBody(w, r, &createUserRequest); !ok {
		uh.logger.Warn(r.Context(), "Failed to read request body for create user")
		hz_handler.Error(w, service_error.BadRequest, "Request body could not be read.")
		return
	}

	uh.logger.InfoArgs(r.Context(), "Creating user: %s", createUserRequest.RunescapeName)

	domainUser := MapCreateUserRequestToDomainUser(createUserRequest)

	user, err := uh.service.CreateUser(r.Context(), domainUser)
	if err != nil {
		if errors.Is(err, ErrRunescapeNameTracked) {
			uh.logger.WarnArgs(r.Context(), "Runescape name already tracked: %s", createUserRequest.RunescapeName)
			hz_handler.Error(w, service_error.RunescapeNameAlreadyTracked, "The runescape name is already associated with a user.")
		}

		uh.logger.ErrorArgs(r.Context(), "An unexpected error occurred while creating user: %+v", err)
		hz_handler.Error(w, service_error.Internal, "An unexpected service_error occurred while performing the user operation.")
		return
	}

	response := api.CreateUserResponse{
		User: MapDomainToApi(user),
	}

	hz_handler.Ok(w, response)
}

func (uh *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var updateUserRequest api.UpdateUserRequest
	if ok := hz_handler.ReadBody(w, r, &updateUserRequest); !ok {
		uh.logger.Warn(r.Context(), "Failed to read request body for update user")
		hz_handler.Error(w, service_error.BadRequest, "Request body could not be read.")
		return
	}

	uh.logger.InfoArgs(r.Context(), "Updating user: %s", updateUserRequest.Id)

	domainUser := MapUpdateUserRequestToDomainUser(updateUserRequest)

	user, err := uh.service.UpdateUser(r.Context(), domainUser)
	if err != nil {
		if errors.Is(err, ErrRunescapeNameTracked) {
			uh.logger.WarnArgs(r.Context(), "Runescape name already tracked for user: %s", updateUserRequest.Id)
			hz_handler.Error(w, service_error.RunescapeNameAlreadyTracked, "The runescape name is already associated with a user.")
		}

		uh.logger.ErrorArgs(r.Context(), "An unexpected error occurred while updating user: %+v", err)
		hz_handler.Error(w, service_error.Internal, "An unexpected service_error occurred while performing the user operation.")
		return
	}

	response := api.CreateUserResponse{
		User: MapDomainToApi(user),
	}

	hz_handler.Ok(w, response)
}
