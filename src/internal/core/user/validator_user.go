package user

type UserValidator interface {
	ValidateUser(user User) error
}

type userValidator struct {
}

func NewUserValidator() UserValidator {
	return &userValidator{}
}

func (uv *userValidator) ValidateUser(user User) error {
	return nil
}
