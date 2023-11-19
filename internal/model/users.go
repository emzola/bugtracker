package model

import (
	"errors"
	"time"

	"github.com/emzola/bugtracker/pkg/validator"
	"golang.org/x/crypto/bcrypt"
)

// User defines user data.
type User struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Password   password  `json:"-"`
	Activated  bool      `json:"activated"`
	Role       string    `json:"role"`
	CreatedOn  time.Time `json:"created_on"`
	CreatedBy  string    `json:"created_by"`
	ModifiedOn time.Time `json:"modified_on"`
	ModifiedBy string    `json:"modified_by"`
	Version    int       `json:"-"`
}

// password contains the plaintext and hashed versions of the password for a user.
type password struct {
	Plaintext *string
	Hash      []byte
}

// Set calculates the bcrypt hash of a plaintext password, and stores both
// the hash and the plaintext versions in the struct.
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.Plaintext = &plaintextPassword
	p.Hash = hash
	return nil
}

// Matches checks whether the provided plaintext password matches the hashed password
// stored in the struct, returning true if it matches and false otherwise.
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.Hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

// Validate user.
func (u User) Validate(v *validator.Validator) {
	v.Check(u.Name != "", "name", "must be provided")
	v.Check(len(u.Name) >= 3, "name", "must not be less than 3 bytes long")
	v.Check(len(u.Name) <= 500, "name", "must not be more than 500 bytes long")
	ValidateEmail(v, u.Email)
	if u.Password.Plaintext != nil {
		ValidatePasswordPlaintext(v, *u.Password.Plaintext)
	}
	if u.Password.Hash == nil {
		panic("missing password hash for user")
	}
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}
