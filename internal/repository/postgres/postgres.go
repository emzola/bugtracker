package postgres

import (
	"database/sql"
)

// Repository defines a PostgreSQL-based project repository.
type Repository struct {
	db *sql.DB
}

// New creates a new PostgreSQL-based repository.
func New(db *sql.DB) *Repository {
	return &Repository{db}
}
