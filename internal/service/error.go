package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	// ErrNotFound is returned when a requested record is not found.
	ErrNotFound = errors.New("not found")

	// ErrFailedValidation is returned when there is a validation error.
	ErrFailedValidation = errors.New("failed validation")

	// ErrEditConflict is returned when there is an edit conflict error.
	ErrEditConflict = errors.New("edit conflict")

	// ErrInvalidCredentials is returned when there is no match between password and hash.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrInvalidRole is returned when user role is not allowed.
	ErrInvalidRole = errors.New("invalid role")
)

// failedValidationErr loops through an errors map and returns ErrFailedValidation
// which contains the keys and values of the errors map.
func failedValidationErr(errors map[string]string) error {
	if len(errors) == 0 {
		return nil
	}
	keys := make([]string, len(errors))
	i := 0
	for key := range errors {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	var s strings.Builder
	for i, key := range keys {
		if i > 0 {
			s.WriteString("; ")
		}
		fmt.Fprintf(&s, "%v: %v", key, errors[key])
	}
	s.WriteString(".")
	ErrFailedValidation = fmt.Errorf("%s", s.String())
	return ErrFailedValidation
}
