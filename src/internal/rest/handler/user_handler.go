package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ctfloyd/hazelmere-api/src/internal/core/user"
	"github.com/ctfloyd/hazelmere-api/src/internal/foundation/middleware"
	"github.com/ctfloyd/hazelmere-api/src/internal/foundation/monitor"
	"github.com/ctfloyd/hazelmere-api/src/internal/rest/service_error"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_handler"
	"github.com/go-chi/chi/v5"
	chiWare "github.com/go-chi/chi/v5/middleware"
)

type UserHandler struct {
	monitor *monitor.Monitor
	service user.UserService
}

func NewUserHandler(mon *monitor.Monitor, service user.UserService) *UserHandler {
	return &UserHandler{mon, service}
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
	ctx, span := uh.monitor.StartSpan(r.Context(), "UserHandler.GetUserById")
	defer span.End()

	id := chi.URLParam(r, "id")
	uh.monitor.Logger().InfoArgs(ctx, "Getting user by id: %s", id)

	u, err := uh.service.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(user.ErrUserNotFound, err) {
			uh.monitor.Logger().WarnArgs(ctx, "User not found: %s", id)
			hz_handler.Error(w, service_error.UserNotFound, "User not found.")
			return
		} else {
			uh.monitor.Logger().ErrorArgs(ctx, "An unexpected error occurred while getting user by id: %+v", err)
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
	ctx, span := uh.monitor.StartSpan(r.Context(), "UserHandler.GetAllUsers")
	defer span.End()

	uh.monitor.Logger().Info(ctx, "Getting all users")

	users, err := uh.service.GetAllUsers(ctx)
	if err != nil {
		uh.monitor.Logger().ErrorArgs(ctx, "An unexpected error occurred while getting all users: %+v", err)
		hz_handler.Error(w, service_error.Internal, "An unexpected service_error occurred while performing the user operation.")
		return
	}

	response := api.GetAllUsersResponse{
		Users: user.User{}.ManyToAPI(users),
	}

	hz_handler.Ok(w, response)
}

func (uh *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := uh.monitor.StartSpan(r.Context(), "UserHandler.CreateUser")
	defer span.End()

	var createUserRequest api.CreateUserRequest
	if ok := hz_handler.ReadBody(w, r, &createUserRequest); !ok {
		uh.monitor.Logger().Warn(ctx, "Failed to read request body for create user")
		hz_handler.Error(w, service_error.BadRequest, "Request body could not be read.")
		return
	}

	uh.monitor.Logger().InfoArgs(ctx, "Creating user: %s", createUserRequest.RunescapeName)

	domainUser := user.User{}.FromCreateRequest(createUserRequest)

	u, err := uh.service.CreateUser(ctx, domainUser)
	if err != nil {
		if errors.Is(err, user.ErrRunescapeNameTracked) {
			uh.monitor.Logger().WarnArgs(ctx, "Runescape name already tracked: %s", createUserRequest.RunescapeName)
			hz_handler.Error(w, service_error.RunescapeNameAlreadyTracked, "The runescape name is already associated with a user.")
		}

		uh.monitor.Logger().ErrorArgs(ctx, "An unexpected error occurred while creating user: %+v", err)
		hz_handler.Error(w, service_error.Internal, "An unexpected service_error occurred while performing the user operation.")
		return
	}

	response := api.CreateUserResponse{
		User: u.ToAPI(),
	}

	hz_handler.Ok(w, response)
}

func (uh *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := uh.monitor.StartSpan(r.Context(), "UserHandler.UpdateUser")
	defer span.End()

	var updateUserRequest api.UpdateUserRequest
	if ok := hz_handler.ReadBody(w, r, &updateUserRequest); !ok {
		uh.monitor.Logger().Warn(ctx, "Failed to read request body for update user")
		hz_handler.Error(w, service_error.BadRequest, "Request body could not be read.")
		return
	}

	uh.monitor.Logger().InfoArgs(ctx, "Updating user: %s", updateUserRequest.Id)

	domainUser := user.User{}.FromUpdateRequest(updateUserRequest)

	u, err := uh.service.UpdateUser(ctx, domainUser)
	if err != nil {
		if errors.Is(err, user.ErrRunescapeNameTracked) {
			uh.monitor.Logger().WarnArgs(ctx, "Runescape name already tracked for user: %s", updateUserRequest.Id)
			hz_handler.Error(w, service_error.RunescapeNameAlreadyTracked, "The runescape name is already associated with a user.")
		}

		uh.monitor.Logger().ErrorArgs(ctx, "An unexpected error occurred while updating user: %+v", err)
		hz_handler.Error(w, service_error.Internal, "An unexpected service_error occurred while performing the user operation.")
		return
	}

	response := api.CreateUserResponse{
		User: u.ToAPI(),
	}

	hz_handler.Ok(w, response)
}
