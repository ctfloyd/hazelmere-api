package user

import (
	"context"
	"errors"

	"github.com/ctfloyd/hazelmere-api/src/internal/database"
	"github.com/ctfloyd/hazelmere-api/src/internal/foundation/monitor"
	"github.com/google/uuid"
)

var ErrUserGeneric = errors.New("an service_error occurred while performing the user operation")
var ErrUserNotFound = errors.New("user not found")
var ErrUserValidation = errors.New("user is invalid")
var ErrRunescapeNameTracked = errors.New("runescape name tracked")

type UserService interface {
	GetUserById(ctx context.Context, id string) (User, error)
	GetAllUsers(ctx context.Context) ([]User, error)
	CreateUser(ctx context.Context, user User) (User, error)
	UpdateUser(ctx context.Context, user User) (User, error)
}

type userService struct {
	monitor    *monitor.Monitor
	validator  UserValidator
	repository UserRepository
}

func NewUserService(mon *monitor.Monitor, repository UserRepository, validator UserValidator) UserService {
	return &userService{
		monitor:    mon,
		validator:  validator,
		repository: repository,
	}
}

func (us *userService) GetUserById(ctx context.Context, id string) (User, error) {
	ctx, span := us.monitor.StartSpan(ctx, "userService.GetUserById")
	defer span.End()

	data, err := us.repository.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return User{}, ErrUserNotFound
		}

		return User{}, errors.Join(ErrUserGeneric, err)
	}
	return User{}.FromData(data), nil
}

func (us *userService) GetAllUsers(ctx context.Context) ([]User, error) {
	ctx, span := us.monitor.StartSpan(ctx, "userService.GetAllUsers")
	defer span.End()

	data, err := us.repository.GetAllUsers(ctx)
	if err != nil {
		return []User{}, errors.Join(ErrUserGeneric, err)
	}
	return User{}.ManyFromData(data), nil
}

func (us *userService) CreateUser(ctx context.Context, user User) (User, error) {
	ctx, span := us.monitor.StartSpan(ctx, "userService.CreateUser")
	defer span.End()

	user.Id = uuid.New().String()

	err := us.validator.ValidateUser(user)
	if err != nil {
		return User{}, errors.Join(ErrUserValidation, err)
	}

	_, err = us.repository.GetUserByRunescapeName(ctx, user.RunescapeName)
	if err == nil {
		return User{}, ErrRunescapeNameTracked
	} else {
		if !errors.Is(err, database.ErrNotFound) {
			return User{}, errors.Join(ErrUserGeneric, err)
		}
	}

	data, err := us.repository.CreateUser(ctx, user.ToData())
	if err != nil {
		return User{}, errors.Join(ErrUserGeneric, err)
	}

	return User{}.FromData(data), nil
}

func (us *userService) UpdateUser(ctx context.Context, user User) (User, error) {
	ctx, span := us.monitor.StartSpan(ctx, "userService.UpdateUser")
	defer span.End()

	_, err := us.GetUserById(ctx, user.Id)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return User{}, ErrUserNotFound
		}
		return User{}, errors.Join(ErrUserGeneric, err)
	}

	err = us.validator.ValidateUser(user)
	if err != nil {
		return User{}, errors.Join(ErrUserValidation, err)
	}

	existingUser, err := us.repository.GetUserByRunescapeName(ctx, user.RunescapeName)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return User{}, errors.Join(ErrUserGeneric, err)
		}
	} else if existingUser.Id != user.Id {
		return User{}, ErrRunescapeNameTracked
	}

	data, err := us.repository.UpdateUser(ctx, user.ToData())
	if err != nil {
		return User{}, errors.Join(ErrUserGeneric, err)
	}

	return User{}.FromData(data), nil
}
