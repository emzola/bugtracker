package model

import (
	"time"

	"github.com/emzola/bugtracker/pkg/validator"
)

const (
	ScopeActivation = "activation"
)

// Token holds data for an individual token.
type Token struct {
	Plaintext string
	Hash      []byte
	UserID    int64
	Expiry    time.Time
	Scope     string
}

// Validate token plaintext.
func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}
