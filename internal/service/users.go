package service

import (
	"context"
	"errors"
	"time"

	"github.com/emzola/bugtracker/internal/model"
	"github.com/emzola/bugtracker/internal/repository"
	"github.com/emzola/bugtracker/pkg/validator"
)

type userRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	CreateToken(ctx context.Context, userID int64, ttl time.Duration, scope string) (*model.Token, error)
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
	// Generate an activation token.
	token, err := s.repo.CreateToken(ctx, user.ID, 3*24*time.Hour, model.ScopeActivation)
	if err != nil {
		return nil, err
	}
	// Send welcome email with activation token in a background goroutine.
	data := map[string]string{
		"activationToken": token.Plaintext,
		"name":            user.Name,
	}
	s.SendEmail(data, user.Email, "user_welcome.tmpl")
	return user, nil
}

// GetUserByEmail retrieves a user by email.
func (s *Service) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	v := validator.New()
	if model.ValidateEmail(v, email); !v.Valid() {
		return nil, failedValidationErr(v.Errors)
	}
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			v.AddError("email", "no matching email address found")
			return nil, failedValidationErr(v.Errors)
		default:
			return nil, err
		}
	}
	return user, nil
}
