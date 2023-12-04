package postgres

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"time"

	"github.com/emzola/issuetracker/internal/model"
)

// CreateToken is a shortcut method which creates a new token struct
// and inserts the data in the tokens table.
func (r *Repository) CreateToken(ctx context.Context, userID int64, ttl time.Duration, scope string) (*model.Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}
	err = r.InsertToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// generateToken generates a new token instance containing the userID, expiry and scope information.
func generateToken(userID int64, ttl time.Duration, scope string) (*model.Token, error) {
	token := &model.Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]
	return token, nil
}

// InsertToken inserts a token record into the tokens table.
func (r *Repository) InsertToken(ctx context.Context, token *model.Token) error {
	query := `
		INSERT INTO tokens(hash, user_id, expiry, scope)
		VALUES ($1, $2, $3, $4)`
	args := []interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		switch {
		case err.Error() == "ERROR: canceling statement due to user request":
			return fmt.Errorf("%v: %w", err, ctx.Err())
		default:
			return err
		}
	}
	return nil
}

// DeleteAllTokensForUser deletes all tokens for a specific user and scope.
func (r *Repository) DeleteAllTokensForUser(ctx context.Context, scope string, userID int64) error {
	query := `
		DELETE FROM tokens
		WHERE scope = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, scope, userID)
	if err != nil {
		switch {
		case err.Error() == "ERROR: canceling statement due to user request":
			return fmt.Errorf("%v: %w", err, ctx.Err())
		default:
			return err
		}
	}
	return nil
}
