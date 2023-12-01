package postgres

import (
	"context"
	"fmt"

	"github.com/emzola/issuetracker/internal/model"
)

// GetIssuesReportStatus retrieves the count of issue statuses for a specific project record.
func (r *Repository) GetIssuesStatusReport(ctx context.Context, projectID int64) ([]*model.IssuesStatus, error) {
	query := `
		SELECT status, COUNT(status)
		FROM issues
		WHERE project_id = $1
		GROUP BY status`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		switch {
		case err.Error() == "ERROR: canceling statement due to user request":
			return nil, fmt.Errorf("%v: %w", err, ctx.Err())
		default:
			return nil, err
		}
	}
	defer rows.Close()
	statuses := []*model.IssuesStatus{}
	for rows.Next() {
		var status model.IssuesStatus
		err := rows.Scan(
			&status.Status,
			&status.IssuesCount,
		)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, &status)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return statuses, nil
}

// GetIssuesAssigneeReport retrieves the count of issue assignees for a specific project record.
func (r *Repository) GetIssuesAssigneeReport(ctx context.Context, projectID int64) ([]*model.IssuesAssignee, error) {
	query := `
		SELECT users.id, users.name, COUNT(users.id)
		FROM users
		LEFT JOIN issues
		ON users.id = issues.assigned_to
		WHERE project_id = $1
		GROUP BY users.id`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		switch {
		case err.Error() == "ERROR: canceling statement due to user request":
			return nil, fmt.Errorf("%v: %w", err, ctx.Err())
		default:
			return nil, err
		}
	}
	defer rows.Close()
	assignees := []*model.IssuesAssignee{}
	for rows.Next() {
		var assignee model.IssuesAssignee
		err := rows.Scan(
			&assignee.AssigneeID,
			&assignee.AssigneeName,
			&assignee.IssuesAssigned,
		)
		if err != nil {
			return nil, err
		}
		assignees = append(assignees, &assignee)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return assignees, nil
}

// GetIssuesAssigneeReport retrieves the count of issue assignees for a specific project record.
func (r *Repository) GetIssuesReporterReport(ctx context.Context, projectID int64) ([]*model.IssuesReporter, error) {
	query := `
		SELECT users.id, users.name, COUNT(users.id)
		FROM users
		LEFT JOIN issues
		ON users.id = issues.reporter_id
		WHERE project_id = $1
		GROUP BY users.id`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		switch {
		case err.Error() == "ERROR: canceling statement due to user request":
			return nil, fmt.Errorf("%v: %w", err, ctx.Err())
		default:
			return nil, err
		}
	}
	defer rows.Close()
	reporters := []*model.IssuesReporter{}
	for rows.Next() {
		var reporter model.IssuesReporter
		err := rows.Scan(
			&reporter.ReporterID,
			&reporter.ReporterName,
			&reporter.IssuesReported,
		)
		if err != nil {
			return nil, err
		}
		reporters = append(reporters, &reporter)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return reporters, nil
}
