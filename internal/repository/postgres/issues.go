package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/emzola/issuetracker/internal/model"
	"github.com/emzola/issuetracker/internal/repository"
)

func (r *Repository) CreateIssue(ctx context.Context, issue *model.Issue) error {
	query := `
		INSERT INTO issues (title, description, reporter_id, project_id, assigned_to, status, priority, target_resolution_date, created_by, modified_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, reported_date, created_on, modified_on, version`
	args := []interface{}{issue.Title, issue.Description, issue.ReporterID, issue.ProjectID, issue.AssignedTo, issue.Status, issue.Priority, issue.TargetResolutionDate, issue.CreatedBy, issue.ModifiedBy}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&issue.ID, &issue.ReportedDate, &issue.CreatedOn, &issue.ModifiedOn, &issue.Version)
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

func (r *Repository) GetIssue(ctx context.Context, id int64) (*model.Issue, error) {
	if id < 1 {
		return nil, repository.ErrNotFound
	}
	query := `
		SELECT id, title, description, reporter_id, reported_date, project_id, assigned_to, status, priority, target_resolution_date, progress, actual_resolution_date, resolution_summary, created_on, created_by, modified_on, modified_by, version
		FROM issues
		WHERE id = $1`
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

func (r *Repository) GetAllIssues(ctx context.Context, title string, reportedDate time.Time, projectID, assignedTo int64, status, priority string, filters model.Filters) ([]*model.Issue, model.Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, title, description, reporter_id, reported_date, project_id, assigned_to, status, priority, target_resolution_date, progress, actual_resolution_date, resolution_summary, created_on, created_by, modified_on, modified_by, version
		FROM issues
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (reported_date = $2 OR $2 = '0001-01-01')
		AND (project_id = $3 OR $3 = 0)
		AND (assigned_to = $4 OR $4 = 0)
		AND (LOWER(status) = LOWER($5) OR $5 = '')
		AND (LOWER(priority) = LOWER($6) OR $6 = '')
		ORDER BY %s %s, id ASC 
		LIMIT $7 OFFSET $8`, filters.SortColumn(), filters.SortDirection())
	args := []interface{}{title, reportedDate, projectID, assignedTo, status, priority, filters.Limit(), filters.Offset()}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		switch {
		case err.Error() == "ERROR: canceling statement due to user request":
			return nil, model.Metadata{}, fmt.Errorf("%v: %w", err, ctx.Err())
		default:
			return nil, model.Metadata{}, err
		}
	}
	defer rows.Close()
	totalRecords := 0
	issues := []*model.Issue{}
	for rows.Next() {
		var issue model.Issue
		err := rows.Scan(
			&totalRecords,
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
			return nil, model.Metadata{}, err
		}
		issues = append(issues, &issue)
	}
	if err = rows.Err(); err != nil {
		return nil, model.Metadata{}, err
	}
	metadata := model.CalculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return issues, metadata, nil
}

func (r *Repository) UpdateIssue(ctx context.Context, issue *model.Issue) error {
	query := `
		UPDATE issues
		SET title = $1, description = $2, assigned_to = $3, status = $4, priority = $5, target_resolution_date = $6, progress = $7, actual_resolution_date = $8, resolution_summary = $9, modified_on = CURRENT_TIMESTAMP(0), modified_by = $10, version = version + 1
		WHERE id = $11 AND version = $12
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

func (r *Repository) DeleteIssue(ctx context.Context, id int64) error {
	if id < 1 {
		return repository.ErrNotFound
	}
	query := `
		DELETE FROM issues
		WHERE id = $1`
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
