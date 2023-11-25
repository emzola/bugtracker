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
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
	CreateToken(ctx context.Context, userID int64, ttl time.Duration, scope string) (*model.Token, error)
	GetUserForToken(ctx context.Context, tokenScope, tokenPlaintext string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User, modifiedby string) error
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

// GetUserByID retrieves a user by ID.
func (s *Service) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return user, nil
}

// GetUserForToken retrieves a user whose records matches a token.
func (s *Service) GetUserForToken(ctx context.Context, tokenScope, tokenPlaintext string) (*model.User, error) {
	v := validator.New()
	if model.ValidateTokenPlaintext(v, tokenPlaintext); !v.Valid() {
		return nil, failedValidationErr(v.Errors)
	}
	user, err := s.repo.GetUserForToken(ctx, model.ScopeActivation, tokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			v.AddError("token", "invalid or expired activation token")
			return nil, failedValidationErr(v.Errors)
		default:
			return nil, err
		}
	}
	return user, nil
}

// ActivateUser activates a user.
func (s *Service) ActivateUser(ctx context.Context, user *model.User, modifiedBy string) error {
	// Update user.
	user.Activated = true
	err := s.repo.UpdateUser(ctx, user, modifiedBy)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrEditConflict):
			return ErrEditConflict
		default:
			return err
		}
	}
	// Delete all activation tokens for user.
	err = s.repo.DeleteAllTokensForUser(ctx, model.ScopeActivation, user.ID)
	if err != nil {
		return err
	}
	return nil
}
