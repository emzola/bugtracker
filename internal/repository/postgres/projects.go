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

// CreateProject adds a new project record.
func (r *Repository) CreateProject(ctx context.Context, project *model.Project) error {
	query := `
		INSERT INTO projects (name, description, start_date, target_end_date, created_by, modified_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_on, modified_on, version`
	args := []interface{}{project.Name, project.Description, project.StartDate, project.TargetEndDate, project.CreatedBy, project.ModifiedBy}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&project.ID, &project.CreatedOn, &project.ModifiedOn, &project.Version)
	if err != nil {
		switch {
		case err.Error() == "ERROR: canceling statement due to user request":
			return fmt.Errorf("%v: %w", err, ctx.Err())
		case err.Error() == `ERROR: duplicate key value violates unique constraint "projects_name_key" (SQLSTATE 23505)`:
			return repository.ErrDuplicateKey
		default:
			return err
		}
	}
	return nil
}

// AssignProject assigns a project record to a user record.
func (r *Repository) AssignProject(ctx context.Context, projectID, userID int64) error {
	query := `
		INSERT INTO projects_users 
		SELECT $1, users.id FROM users WHERE users.id = $2`
	args := []interface{}{projectID, userID}
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		switch {
		case err.Error() == "ERROR: canceling statement due to user request":
			return fmt.Errorf("%v: %w", err, ctx.Err())
		case err.Error() == `ERROR: duplicate key value violates unique constraint "projects_users_pkey" (SQLSTATE 23505)`:
			return repository.ErrDuplicateKey
		default:
			return err
		}

	}
	return nil
}

// GetProject retrieves a project record by its id.
func (r *Repository) GetProject(ctx context.Context, id int64) (*model.Project, error) {
	if id < 1 {
		return nil, repository.ErrNotFound
	}
	query := `
		SELECT id, name, description, start_date, target_end_date, actual_end_date, created_on, modified_on, created_by, modified_by, version
		FROM projects
		WHERE id = $1`
	var project model.Project
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.StartDate,
		&project.TargetEndDate,
		&project.ActualEndDate,
		&project.CreatedOn,
		&project.ModifiedOn,
		&project.CreatedBy,
		&project.ModifiedBy,
		&project.Version,
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
	return &project, nil
}

// GetAllProjects returns a paginated list of all projects. List can be filtered and sorted.
func (r *Repository) GetAllProjects(ctx context.Context, name string, startDate, targetEndDate, actualEndDate time.Time, createdBy string, filters model.Filters) ([]*model.Project, model.Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, name, description, start_date, target_end_date, actual_end_date, created_on, modified_on, created_by, modified_by, version
		FROM projects
		WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (start_date = $2 OR $2 = '0001-01-01')
		AND (target_end_date = $3 OR $3 = '0001-01-01')
		AND (actual_end_date = $4 OR $4 = '0001-01-01')
		AND (LOWER(created_by) = LOWER($5) OR $5 = '')
		ORDER BY %s %s, id ASC 
		LIMIT $6 OFFSET $7`, filters.SortColumn(), filters.SortDirection())
	args := []interface{}{name, startDate, targetEndDate, actualEndDate, createdBy, filters.Limit(), filters.Offset()}
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
	projects := []*model.Project{}
	for rows.Next() {
		var project model.Project
		err := rows.Scan(
			&totalRecords,
			&project.ID,
			&project.Name,
			&project.Description,
			&project.StartDate,
			&project.TargetEndDate,
			&project.ActualEndDate,
			&project.CreatedOn,
			&project.ModifiedOn,
			&project.CreatedBy,
			&project.ModifiedBy,
			&project.Version,
		)
		if err != nil {
			return nil, model.Metadata{}, err
		}
		projects = append(projects, &project)
	}
	if err = rows.Err(); err != nil {
		return nil, model.Metadata{}, err
	}
	metadata := model.CalculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return projects, metadata, nil
}

// GetAllProjectsForUser returns a paginated list of all projects for a specific user.
func (r *Repository) GetAllProjectsForUser(ctx context.Context, userID int64, filters model.Filters) ([]*model.Project, model.Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), projects.id, projects.name, projects.description, projects.start_date, projects.target_end_date, projects.actual_end_date, projects.created_on, projects.modified_on, projects.created_by, projects.modified_by, projects.version
		FROM projects
		INNER JOIN projects_users ON projects_users.project_id = projects.id
		INNER JOIN users ON projects_users.user_id = users.id
		WHERE users.id = $1
		ORDER BY %s %s, id ASC 
		LIMIT $2 OFFSET $3`, filters.SortColumn(), filters.SortDirection())
	args := []interface{}{userID, filters.Limit(), filters.Offset()}
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
	projects := []*model.Project{}
	for rows.Next() {
		var project model.Project
		err := rows.Scan(
			&totalRecords,
			&project.ID,
			&project.Name,
			&project.Description,
			&project.StartDate,
			&project.TargetEndDate,
			&project.ActualEndDate,
			&project.CreatedOn,
			&project.ModifiedOn,
			&project.CreatedBy,
			&project.ModifiedBy,
			&project.Version,
		)
		if err != nil {
			return nil, model.Metadata{}, err
		}
		projects = append(projects, &project)
	}
	if err = rows.Err(); err != nil {
		return nil, model.Metadata{}, err
	}
	metadata := model.CalculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return projects, metadata, nil
}

// UpdateProject updates a project's record.
func (r *Repository) UpdateProject(ctx context.Context, project *model.Project) error {
	query := `
		UPDATE projects
		SET name = $1, description = $2, start_date = $3, target_end_date = $4, actual_end_date = $5, modified_by = $6, modified_on = CURRENT_TIMESTAMP(0), version = version + 1
		WHERE id = $7 AND version = $8
		RETURNING modified_on, version`
	args := []interface{}{project.Name, project.Description, project.StartDate, project.TargetEndDate, project.ActualEndDate, project.ModifiedBy, project.ID, project.Version}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&project.ModifiedOn, &project.Version)
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

// DeleteProject removes a project record by its id.
func (r *Repository) DeleteProject(ctx context.Context, id int64) error {
	if id < 1 {
		return repository.ErrNotFound
	}
	query := `
		DELETE FROM projects
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
