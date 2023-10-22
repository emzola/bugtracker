package postgresql

import (
	"context"

	"github.com/emzola/bugtracker/pkg/model"
)

// Createproject adds a new project record.
func (r *Repository) CreateProject(ctx context.Context, project *model.Project) error {
	query := `
			INSERT INTO project (name, description, start_date, end_date, public_access, owner, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		  	RETURNING id, created_on, version`
	args := []interface{}{project.Name, project.Description, project.StartDate, project.EndDate, project.PublicAccess, project.Owner, project.CreatedBy}
	return r.db.QueryRowContext(ctx, query, args...).Scan(&project.ID, &project.CreatedOn, &project.Version)
}
