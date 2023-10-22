package postgresql

import (
	"database/sql"
	"os"
)

// Repository defines a PostgreSQL-based project repository.
type Repository struct {
	db *sql.DB
}

// New creates a new PostgreSQL-based repository.
func New() (*Repository, error) {
	db, err := sql.Open("postgres", os.Getenv("DB_DSN"))
	if err != nil {
		return nil, err
	}
	return &Repository{db}, nil
}
