package user

import (
	"errors"
	"fmt"
	"github.com/ctfloyd/hazelmere-api/src/internal/service_error"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_handler"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

func (uh *UserHandler) RegisterRoutes(mux *chi.Mux, version hz_handler.ApiVersion) {
	if version == hz_handler.ApiVersionV1 {
		mux.Group(func(r chi.Router) {
			r.Use(middleware.Timeout(5000 * time.Millisecond))
			r.Get(fmt.Sprintf("/v1/user/{id:%s}", hz_handler.RegexUuid), uh.GetUserById)
			r.Get("/v1/user", uh.GetAllUsers)
			r.Post("/v1/user", uh.CreateUser)
			r.Put("/v1/user", uh.UpdateUser)
		})
	}
}

func (uh *UserHandler) GetUserById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := uh.service.GetUserById(r.Context(), id)
	if err != nil {
		if errors.Is(ErrUserNotFound, err) {
			hz_handler.Error(w, service_error.UserNotFound, "User not found.")
			return
		} else {
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
	users, err := uh.service.GetAllUsers(r.Context())
	if err != nil {
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
		hz_handler.Error(w, service_error.BadRequest, "Request body could not be read.")
		return
	}

	domainUser := MapCreateUserRequestToDomainUser(createUserRequest)

	user, err := uh.service.CreateUser(r.Context(), domainUser)
	if err != nil {
		if errors.Is(err, ErrRunescapeNameTracked) {
			hz_handler.Error(w, service_error.RunescapeNameAlreadyTracked, "The runescape name is already associated with a user.")
		}

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
		hz_handler.Error(w, service_error.BadRequest, "Request body could not be read.")
		return
	}

	domainUser := MapUpdateUserRequestToDomainUser(updateUserRequest)

	user, err := uh.service.UpdateUser(r.Context(), domainUser)
	if err != nil {
		if errors.Is(err, ErrRunescapeNameTracked) {
			hz_handler.Error(w, service_error.RunescapeNameAlreadyTracked, "The runescape name is already associated with a user.")
		}

		hz_handler.Error(w, service_error.Internal, "An unexpected service_error occurred while performing the user operation.")
		return
	}

	response := api.CreateUserResponse{
		User: MapDomainToApi(user),
	}

	hz_handler.Ok(w, response)
}
