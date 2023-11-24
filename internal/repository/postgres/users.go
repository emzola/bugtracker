package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/emzola/bugtracker/internal/model"
	"github.com/emzola/bugtracker/internal/repository"
)

// CreateUser adds a new user record.
func (r *Repository) CreateUser(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (name, email, password_hash, activated, role, created_by, modified_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_on, modified_on, version`
	args := []interface{}{user.Name, user.Email, user.Password.Hash, user.Activated, user.Role, user.CreatedBy, user.ModifiedBy}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedOn, &user.ModifiedOn, &user.Version)
	if err != nil {
		switch {
		case err.Error() == "ERROR: canceling statement due to user request":
			return fmt.Errorf("%v: %w", err, ctx.Err())
		case err.Error() == `ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)`:
			return repository.ErrDuplicateKey
		default:
			return err
		}
	}
	return nil
}

// GetUserByEmail retrieves a user record by email.
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, name, email, password_hash, activated, role, created_on, created_by, modified_on, modified_by, version
		FROM users
		WHERE email = $1`
	var user model.User
	if err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password.Hash,
		&user.Activated,
		&user.Role,
		&user.CreatedOn,
		&user.CreatedBy,
		&user.ModifiedOn,
		&user.ModifiedBy,
		&user.Version,
	); err != nil {
		switch {
		case err.Error() == "ERROR: canceling statement due to user request":
			return nil, fmt.Errorf("%v: %w", err, ctx.Err())
		case errors.Is(err, sql.ErrNoRows):
			return nil, repository.ErrNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

// UpdateUser updates a user's record.
func (r *Repository) UpdateUser(ctx context.Context, user *model.User, modifiedBy string) error {
	query := `
		UPDATE users
		SET name = $1, email = $2, password_hash = $3, activated = $4, role = $5, modified_by = $6, version = version + 1
		WHERE id = $7 AND version = $8
		RETURNING version`
	args := []interface{}{user.Name, user.Email, user.Password.Hash, user.Activated, user.Role, modifiedBy, user.ID, user.Version}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == "ERROR: canceling statement due to user request":
			return fmt.Errorf("%v: %w", err, ctx.Err())
		case err.Error() == `ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)`:
			return repository.ErrDuplicateKey
		case errors.Is(err, sql.ErrNoRows):
			return repository.ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// GetUserForToken retrieves a user record from the tokens table.
func (r *Repository) GetUserForToken(ctx context.Context, tokenScope, tokenPlaintext string) (*model.User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))
	query := `
		SELECT users.id, users.name, users.email, users.password_hash, users.activated, users.role, users.created_on, users.created_by, users.modified_on, users.modified_by, users.version
		FROM users
		INNER JOIN tokens
		ON users.id = tokens.user_id
		WHERE tokens.hash = $1
		AND tokens.scope = $2 
		AND tokens.expiry > $3`
	args := []interface{}{tokenHash[:], tokenScope, time.Now()}
	var user model.User
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password.Hash,
		&user.Activated,
		&user.Role,
		&user.CreatedOn,
		&user.CreatedBy,
		&user.ModifiedOn,
		&user.ModifiedBy,
		&user.Version,
	)
	if err != nil {
		switch {
		case err.Error() == "ERROR: canceling statement due to user request":
			return nil, fmt.Errorf("%v: %w", err, ctx.Err())
		case errors.Is(err, sql.ErrNoRows):
			return nil, repository.ErrNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}
