package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/emzola/issuetracker/internal/repository"
	"github.com/emzola/issuetracker/pkg/model"
)

func (r *Repository) CreateProject(ctx context.Context, project *model.Project) error {
	query := `
		INSERT INTO projects (name, description, assigned_to, start_date, target_end_date, created_by, modified_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_on, modified_on, version`
	args := []interface{}{project.Name, project.Description, project.AssignedTo, project.StartDate, project.TargetEndDate, project.CreatedBy, project.ModifiedBy}
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

func (r *Repository) GetProject(ctx context.Context, id int64) (*model.Project, error) {
	if id < 1 {
		return nil, repository.ErrNotFound
	}
	query := `
		SELECT id, name, description, assigned_to, start_date, target_end_date, actual_end_date, created_on, modified_on, created_by, modified_by, version
		FROM projects
		WHERE id = $1`
	var project model.Project
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.AssignedTo,
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

func (r *Repository) GetAllProjects(ctx context.Context, name string, assignedTo int64, startDate, targetEndDate, actualEndDate time.Time, createdBy string, filters model.Filters) ([]*model.Project, model.Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, name, description, assigned_to, start_date, target_end_date, actual_end_date, created_on, modified_on, created_by, modified_by, version
		FROM projects
		WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (assigned_to = $2 OR $2 = 0)
		AND (start_date = $3 OR $3 = '0001-01-01')
		AND (target_end_date = $4 OR $4 = '0001-01-01')
		AND (actual_end_date = $5 OR $5 = '0001-01-01')
		AND (LOWER(created_by) = LOWER($6) OR $6 = '')
		ORDER BY %s %s, id ASC 
		LIMIT $7 OFFSET $8`, filters.SortColumn(), filters.SortDirection())
	args := []interface{}{name, assignedTo, startDate, targetEndDate, actualEndDate, createdBy, filters.Limit(), filters.Offset()}
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
			&project.AssignedTo,
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

func (r *Repository) UpdateProject(ctx context.Context, project *model.Project) error {
	query := `
		UPDATE projects
		SET name = $1, description = $2, assigned_to = $3, start_date = $4, target_end_date = $5, actual_end_date = $6, modified_by = $7, modified_on = CURRENT_TIMESTAMP(0), version = version + 1
		WHERE id = $8 AND version = $9
		RETURNING modified_on, version`
	args := []interface{}{project.Name, project.Description, project.AssignedTo, project.StartDate, project.TargetEndDate, project.ActualEndDate, project.ModifiedBy, project.ID, project.Version}
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

func (r *Repository) GetProjectUsers(ctx context.Context, projectID int64, role string, filters model.Filters) ([]*model.User, model.Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), users.id, users.name, users.email, users.password_hash, users.activated, users.role, users.created_on, users.created_by, users.modified_on, users.modified_by, users.version
		FROM users
		INNER JOIN projects_users ON projects_users.user_id = users.id
		INNER JOIN projects ON projects_users.project_id = projects.id
		WHERE projects.id = $1
		AND (LOWER(users.role) = LOWER($2) OR $2 = '')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.SortColumn(), filters.SortDirection())
	args := []interface{}{projectID, role, filters.Limit(), filters.Offset()}
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
	users := []*model.User{}
	for rows.Next() {
		var user model.User
		err := rows.Scan(
			&totalRecords,
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
			return nil, model.Metadata{}, err
		}
		users = append(users, &user)
	}
	if err = rows.Err(); err != nil {
		return nil, model.Metadata{}, err
	}
	metadata := model.CalculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return users, metadata, nil
}

func (r *Repository) GetProjectUser(ctx context.Context, projectID, userID int64) (*model.User, error) {
	query := `
		SELECT users.id, users.name, users.email, users.password_hash, users.activated, users.role, users.created_on, users.created_by, users.modified_on, users.modified_by, users.version
		FROM users
		INNER JOIN projects_users ON projects_users.user_id = users.id
		INNER JOIN projects ON projects_users.project_id = projects.id
		WHERE projects.id = $1 AND users.id = $2`
	args := []interface{}{projectID, userID}
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
