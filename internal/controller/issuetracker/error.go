package issuetracker

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrFailedValidation   = errors.New("failed validation")
	ErrEditConflict       = errors.New("edit conflict")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidRole        = errors.New("invalid role")
	ErrActivated          = errors.New("invalid role")
	ErrNotPermitted       = errors.New("not permitted")
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
