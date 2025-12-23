package user

import (
	"context"
	"errors"
	"github.com/ctfloyd/hazelmere-api/src/internal/database"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
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
	logger     hz_logger.Logger
	validator  UserValidator
	repository UserRepository
}

func NewUserService(logger hz_logger.Logger, repository UserRepository, validator UserValidator) UserService {
	return &userService{
		logger:     logger,
		validator:  validator,
		repository: repository,
	}
}

func (us *userService) GetUserById(ctx context.Context, id string) (User, error) {
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
	data, err := us.repository.GetAllUsers(ctx)
	if err != nil {
		return []User{}, errors.Join(ErrUserGeneric, err)
	}
	return User{}.ManyFromData(data), nil
}

func (us *userService) CreateUser(ctx context.Context, user User) (User, error) {
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
		// Successful operation, the RSN that's being updated to is already tracked.
		return User{}, ErrRunescapeNameTracked
	}

	data, err := us.repository.UpdateUser(ctx, user.ToData())
	if err != nil {
		return User{}, errors.Join(ErrUserGeneric, err)
	}

	return User{}.FromData(data), nil
}
