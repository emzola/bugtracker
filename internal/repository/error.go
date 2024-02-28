package repository

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrFailedValidation = errors.New("failed validation")
	ErrEditConflict     = errors.New("edit conflict")
	ErrDuplicateKey     = errors.New("duplicate key")
)
