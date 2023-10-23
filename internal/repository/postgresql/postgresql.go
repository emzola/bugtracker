package postgresql

import (
	"database/sql"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Repository defines a PostgreSQL-based project repository.
type Repository struct {
	db *sql.DB
}

// New creates a new PostgreSQL-based repository.
func New() (*Repository, error) {
	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}
	return &Repository{db}, nil
}
