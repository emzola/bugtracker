package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/emzola/bugtracker/internal/model"
	"github.com/emzola/bugtracker/internal/repository"
	"github.com/emzola/bugtracker/pkg/mailer"
	"github.com/emzola/bugtracker/pkg/validator"
	"go.uber.org/zap"
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
	// Send welcome email in a background goroutine.
	go func() {
		defer func() {
			if err := recover(); err != nil {
				s.Logger.Info(fmt.Sprintf("%s", err))
			}
		}()
		data := map[string]string{
			"name": user.Name,
		}
		mailer := mailer.New(s.Config.Smtp.Host, s.Config.Smtp.Port, s.Config.Smtp.Username, s.Config.Smtp.Password, s.Config.Smtp.Sender)
		err = mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			s.Logger.Info("Failed to send welcome email", zap.Error(err))
		}
	}()
	return user, nil
}
