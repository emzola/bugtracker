package postgresql

import (
	"context"
	"time"

	"github.com/emzola/bugtracker/internal/model"
)

// Createproject adds a new project record.
func (r *Repository) CreateProject(ctx context.Context, project *model.Project) error {
	query := `
			INSERT INTO project (name, description, owner, status, start_date, end_date, public_access, created_by, modified_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		  	RETURNING id, created_on, last_modified, version`
	var startDate, endDate time.Time
	var err error
	// Convert start date from string to time.Time so it can be compatible with postgresql date type.
	if project.StartDate != "" {
		startDate, err = time.Parse("2006-01-02", project.StartDate)
		if err != nil {
			return err
		}
	}
	// Convert end date from string to time.Time so it can be compatible with postgresql date type.
	if project.EndDate != "" {
		endDate, err = time.Parse("2006-01-02", project.EndDate)
		if err != nil {
			return err
		}
	}
	args := []interface{}{project.Name, project.Description, project.Owner, project.Status, startDate, endDate, project.PublicAccess, project.CreatedBy, project.ModifiedBy}
	return r.db.QueryRowContext(ctx, query, args...).Scan(&project.ID, &project.CreatedOn, &project.LastModified, &project.Version)
}
