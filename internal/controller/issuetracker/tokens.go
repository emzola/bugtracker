package issuetracker

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/emzola/issuetracker/internal/repository"
	"github.com/emzola/issuetracker/pkg/model"
	"github.com/emzola/issuetracker/pkg/validator"
	"github.com/pascaldekloe/jwt"
)

type tokenRepository interface {
	CreateToken(ctx context.Context, userID int64, ttl time.Duration, scope string) (*model.Token, error)
	DeleteAllTokensForUser(ctx context.Context, scope string, userID int64) error
}

func (c *Controller) CreateActivationToken(ctx context.Context, user *model.User) error {
	if user.Activated {
		return ErrActivated
	}
	token, err := c.repo.CreateToken(ctx, user.ID, 3*24*time.Hour, model.ScopeActivation)
	if err != nil {
		return err
	}
	// Send email with activation token in a background goroutine.
	data := map[string]string{
		"activationToken": token.Plaintext,
		"name":            user.Name,
	}
	c.SendEmail(data, user.Email, "token_activation.tmpl")
	return nil
}

func (c *Controller) CreateAuthenticationToken(ctx context.Context, email, password string) ([]byte, error) {
	v := validator.New()
	model.ValidateEmail(v, email)
	model.ValidatePasswordPlaintext(v, password)
	if !v.Valid() {
		return nil, failedValidationErr(v.Errors)
	}
	user, err := c.repo.GetUserByEmail(ctx, email)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return nil, ErrInvalidCredentials
		default:
			return nil, err
		}
	}
	match, err := user.Password.Matches(password)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, ErrInvalidCredentials
	}
	var claims jwt.Claims
	claims.Subject = strconv.FormatInt(user.ID, 10)
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(time.Now().Add(24 * time.Hour))
	claims.Issuer = "github.com/emzola/issuetracker"
	claims.Audiences = []string{"github.com/emzola/issuetracker"}
	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(c.Config.Jwt.Secret))
	if err != nil {
		return nil, err
	}
	return jwtBytes, nil
}
