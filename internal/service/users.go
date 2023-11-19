package service

import (
	"context"
	"errors"

	"github.com/emzola/bugtracker/internal/model"
	"github.com/emzola/bugtracker/internal/repository"
	"github.com/emzola/bugtracker/pkg/validator"
)

type userRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
}

// CreateUser adds a new user.
func (s *Service) CreateUser(ctx context.Context, name, email, password, role, createdBy, modifiedBy string) (*model.User, error) {
	user := &model.User{
		Name:       name,
		Email:      email,
		Role:       role,
		Activated:  false,
		CreatedBy:  createdBy,
		ModifiedBy: modifiedBy,
	}
	err := user.Password.Set(password)
	if err != nil {
		return nil, err
	}
	v := validator.New()
	if user.Validate(v); !v.Valid() {
		return nil, failedValidationErr(v.Errors)
	}
	err = s.repo.CreateUser(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateKey):
			v.AddError("email", "a user with this email already exists")
			return nil, failedValidationErr(v.Errors)
		default:
			return nil, err
		}
	}
	return user, nil
}
