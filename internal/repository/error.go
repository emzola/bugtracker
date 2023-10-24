package repository

import "errors"

var (
	// ErrNotFound is returned when a requested record is not found.
	ErrNotFound = errors.New("not found")

	// ErrFailedValidation is returned when there is a validation error.
	ErrFailedValidation = errors.New("failed validation")

	// ErrEditConflict is returned when there is an edit conflict error.
	ErrEditConflict = errors.New("edit conflict")
)
