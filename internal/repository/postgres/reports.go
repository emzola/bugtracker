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
			&status.Count,
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
