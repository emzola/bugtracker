package service

import (
	"context"
	"time"

	"github.com/emzola/bugtracker/internal/model"
)

type tokenRepository interface {
	CreateToken(ctx context.Context, userID int64, ttl time.Duration, scope string) (*model.Token, error)
	DeleteAllTokensForUser(ctx context.Context, scope string, userID int64) error
}

// CreateActivationToken creates a new activation token and emails it to user.
func (s *Service) CreateActivationToken(ctx context.Context, user *model.User) error {
	token, err := s.repo.CreateToken(ctx, user.ID, 3*24*time.Hour, model.ScopeActivation)
	if err != nil {
		return err
	}
	// Send email with activation token in a background goroutine.
	data := map[string]string{
		"activationToken": token.Plaintext,
		"name":            user.Name,
	}
	s.SendEmail(data, user.Email, "token_activation.tmpl")
	return nil
}
