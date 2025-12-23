package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ctfloyd/hazelmere-api/src/internal/core/user"
	"github.com/ctfloyd/hazelmere-api/src/internal/rest/middleware"
	"github.com/ctfloyd/hazelmere-api/src/internal/rest/service_error"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_handler"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"github.com/go-chi/chi/v5"
	chiWare "github.com/go-chi/chi/v5/middleware"
)

type UserHandler struct {
	logger  hz_logger.Logger
	service user.UserService
}

func NewUserHandler(logger hz_logger.Logger, service user.UserService) *UserHandler {
	return &UserHandler{logger, service}
}

func (uh *UserHandler) RegisterRoutes(mux *chi.Mux, version ApiVersion, authorizer *middleware.Authorizer) {
	if version == ApiVersionV1 {
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

	u, err := uh.service.GetUserById(r.Context(), id)
	if err != nil {
		if errors.Is(user.ErrUserNotFound, err) {
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
		User: u.ToAPI(),
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
		Users: user.User{}.ManyToAPI(users),
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

	domainUser := user.User{}.FromCreateRequest(createUserRequest)

	u, err := uh.service.CreateUser(r.Context(), domainUser)
	if err != nil {
		if errors.Is(err, user.ErrRunescapeNameTracked) {
			uh.logger.WarnArgs(r.Context(), "Runescape name already tracked: %s", createUserRequest.RunescapeName)
			hz_handler.Error(w, service_error.RunescapeNameAlreadyTracked, "The runescape name is already associated with a user.")
		}

		uh.logger.ErrorArgs(r.Context(), "An unexpected error occurred while creating user: %+v", err)
		hz_handler.Error(w, service_error.Internal, "An unexpected service_error occurred while performing the user operation.")
		return
	}

	response := api.CreateUserResponse{
		User: u.ToAPI(),
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

	domainUser := user.User{}.FromUpdateRequest(updateUserRequest)

	u, err := uh.service.UpdateUser(r.Context(), domainUser)
	if err != nil {
		if errors.Is(err, user.ErrRunescapeNameTracked) {
			uh.logger.WarnArgs(r.Context(), "Runescape name already tracked for user: %s", updateUserRequest.Id)
			hz_handler.Error(w, service_error.RunescapeNameAlreadyTracked, "The runescape name is already associated with a user.")
		}

		uh.logger.ErrorArgs(r.Context(), "An unexpected error occurred while updating user: %+v", err)
		hz_handler.Error(w, service_error.Internal, "An unexpected service_error occurred while performing the user operation.")
		return
	}

	response := api.CreateUserResponse{
		User: u.ToAPI(),
	}

	hz_handler.Ok(w, response)
}
