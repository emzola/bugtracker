package service

import (
	"fmt"
	"sort"
	"strings"
)

// ErrNotFound is returned when a requested record is not found.
var ErrNotFound error

// ErrFailedValidation is returned when there is a validation error.
var ErrFailedValidation error

// ErrEditConflict is returned when there is an edit conflict error.
var ErrEditConflict error

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
