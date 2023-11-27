package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/emzola/issuetracker/internal/model"
	"github.com/emzola/issuetracker/internal/repository"
)

// CreateIssue adds a new issue record.
func (r *Repository) CreateIssue(ctx context.Context, issue *model.Issue) error {
	query := `
		INSERT INTO issues (title, description, reported_date, reporter_id, project_id, assigned_to, status, priority, target_resolution_date, created_by, modified_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING issue_id, created_on, modified_on, version`
	args := []interface{}{issue.Title, issue.Description, issue.ReportedDate, issue.ReporterID, issue.ProjectID, issue.AssignedTo, issue.Status, issue.Priority, issue.TargetResolutionDate, issue.CreatedBy, issue.ModifiedBy}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&issue.ID, &issue.CreatedOn, &issue.ModifiedOn, &issue.Version)
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

// GetIssue retrieves an issue record by its id.
func (r *Repository) GetIssue(ctx context.Context, id int64) (*model.Issue, error) {
	if id < 1 {
		return nil, repository.ErrNotFound
	}
	query := `
		SELECT issue_id, title, description, reporter_id, reported_date, project_id, assigned_to, status, priority, target_resolution_date, progress, actual_resolution_date, resolution_summary, created_on, created_by, modified_on, modified_by, version
		FROM issues
		WHERE issue_id = $1`
	var issue model.Issue
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&issue.ID,
		&issue.Title,
		&issue.Description,
		&issue.ReporterID,
		&issue.ReportedDate,
		&issue.ProjectID,
		&issue.AssignedTo,
		&issue.Status,
		&issue.Priority,
		&issue.TargetResolutionDate,
		&issue.Progress,
		&issue.ActualResolutionDate,
		&issue.ResolutionSummary,
		&issue.CreatedOn,
		&issue.CreatedBy,
		&issue.ModifiedOn,
		&issue.ModifiedBy,
		&issue.Version,
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
	return &issue, nil
}

// UpdateIssueupdates an issue record.
func (r *Repository) UpdateIssue(ctx context.Context, issue *model.Issue) error {
	query := `
		UPDATE issues
		SET title = $1, description = $2, assigned_to = $3, status = $4, priority = $5, target_resolution_date = $6, progress = $7, actual_resolution_date = $8, resolution_summary = $9, modified_on = CURRENT_TIMESTAMP(0), modified_by = $10, version = version + 1
		WHERE issue_id = $11 AND version = $12
		RETURNING modified_on, version`
	args := []interface{}{issue.Title, issue.Description, issue.AssignedTo, issue.Status, issue.Priority, issue.TargetResolutionDate, issue.Progress, issue.ActualResolutionDate, issue.ResolutionSummary, issue.ModifiedBy, issue.ID, issue.Version}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&issue.ModifiedOn, &issue.Version)
	if err != nil {
		switch {
		case err.Error() == "ERROR: canceling statement due to user request":
			return fmt.Errorf("%v: %w", err, ctx.Err())
		case errors.Is(err, sql.ErrNoRows):
			return repository.ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// DeleteIssue removes an issue record by its id.
func (r *Repository) DeleteIssue(ctx context.Context, id int64) error {
	if id < 1 {
		return repository.ErrNotFound
	}
	query := `
		DELETE FROM issues
		WHERE issue_id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		switch {
		case err.Error() == "ERROR: canceling statement due to user request":
			return fmt.Errorf("%v: %w", err, ctx.Err())
		default:
			return err
		}
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}
