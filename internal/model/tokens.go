package model

import (
	"time"

	"github.com/emzola/issuetracker/pkg/validator"
)

const (
	ScopeActivation = "activation"
)

// Token holds data for an individual token.
type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"-"`
	Scope     string    `json:"-"`
}

// Validate token plaintext.
func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}
