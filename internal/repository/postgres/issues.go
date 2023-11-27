package postgres

import (
	"context"
	"fmt"

	"github.com/emzola/issuetracker/internal/model"
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
