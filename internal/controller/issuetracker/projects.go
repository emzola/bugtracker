package issuetracker

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/emzola/issuetracker/internal/repository"
	"github.com/emzola/issuetracker/pkg/model"
	"github.com/emzola/issuetracker/pkg/validator"
)

type projectRepository interface {
	CreateProject(ctx context.Context, project *model.Project) error
	GetProject(ctx context.Context, id int64) (*model.Project, error)
	GetAllProjects(ctx context.Context, name string, assignedTo int64, startDate, targetEndDate, actualEndDate time.Time, createdBy string, filters model.Filters) ([]*model.Project, model.Metadata, error)
	UpdateProject(ctx context.Context, project *model.Project) error
	DeleteProject(ctx context.Context, id int64) error
	GetProjectUsers(ctx context.Context, projectID int64, role string, filters model.Filters) ([]*model.User, model.Metadata, error)
	GetProjectUser(ctx context.Context, projectID, userID int64) (*model.User, error)
}

func (c *Controller) CreateProject(ctx context.Context, name, description string, assignedTo *int64, startDate, targetEndDate, createdBy, modifiedBy string) (*model.Project, error) {
	project := &model.Project{
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
		ModifiedBy:  modifiedBy,
	}
	if startDate != "" {
		start, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return nil, err
		}
		project.StartDate = start
	}
	if targetEndDate != "" {
		targetEnd, err := time.Parse("2006-01-02", targetEndDate)
		if err != nil {
			return nil, err
		}
		project.TargetEndDate = targetEnd
	}
	// Projects can only be assigned to users with role 'lead'.
	// Before project is assigned, attempt to fetch the assignee.
	// If the assignee's role is not 'lead', return an error.
	var assignee *model.User
	var err error
	if assignedTo != nil {
		assignee, err = c.repo.GetUserByID(ctx, *assignedTo)
		if err != nil {
			switch {
			case errors.Is(err, repository.ErrNotFound):
				return nil, ErrNotFound
			default:
				return nil, err
			}
		}
		if assignee.Role != "lead" {
			return nil, ErrInvalidRole
		}
		// Assign lead to project.
		project.AssignedTo = &assignee.ID
	}
	v := validator.New()
	if project.Validate(v); !v.Valid() {
		return nil, failedValidationErr(v.Errors)
	}
	err = c.repo.CreateProject(ctx, project)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateKey):
			v.AddError("name", "a project with this name already exists")
			return nil, failedValidationErr(v.Errors)
		default:
			return nil, err
		}
	}
	// Send email notification to assigned user if project is assigned.
	if assignedTo != nil {
		data := map[string]string{
			"name":        assignee.Name,
			"projectID":   strconv.Itoa(int(project.ID)),
			"projectName": project.Name,
		}
		c.SendEmail(data, assignee.Email, "project_assign.tmpl")
	}
	return project, nil
}

func (c *Controller) GetProject(ctx context.Context, id int64) (*model.Project, error) {
	project, err := c.repo.GetProject(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return project, nil
}

func (c *Controller) GetAllProjects(ctx context.Context, name string, assignedTo int64, startDate, targetEndDate, actualEndDate, createdBy string, filters model.Filters, v *validator.Validator) ([]*model.Project, model.Metadata, error) {
	if filters.Validate(v); !v.Valid() {
		return nil, model.Metadata{}, failedValidationErr(v.Errors)
	}
	var start, targetEnd, actualEnd time.Time
	var err error
	if startDate != "" {
		start, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			return nil, model.Metadata{}, err
		}
	}
	if targetEndDate != "" {
		targetEnd, err = time.Parse("2006-01-02", targetEndDate)
		if err != nil {
			return nil, model.Metadata{}, err
		}
	}
	if actualEndDate != "" {
		actualEnd, err = time.Parse("2006-01-02", actualEndDate)
		if err != nil {
			return nil, model.Metadata{}, err
		}
	}
	projects, metadata, err := c.repo.GetAllProjects(ctx, name, assignedTo, start, targetEnd, actualEnd, createdBy, filters)
	if err != nil {
		return nil, model.Metadata{}, err
	}
	return projects, metadata, nil
}

func (c *Controller) UpdateProject(ctx context.Context, id int64, name, description *string, assignedTo *int64, startDate, targetEndDate, actualEndDate *string, user *model.User) (*model.Project, error) {
	project, err := c.repo.GetProject(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	// Check whether user has permission to update project.
	// Leads can update project details only if it's assigned to them.
	if user.Role == "lead" && *project.AssignedTo != user.ID {
		return nil, ErrNotPermitted
	}
	// At this point, update project as usual.
	if name != nil {
		project.Name = *name
	}
	if description != nil {
		project.Description = *description
	}
	if startDate != nil {
		start, err := time.Parse("2006-01-02", *startDate)
		if err != nil {
			return nil, err
		}
		project.StartDate = start
	}
	if targetEndDate != nil {
		targetEnd, err := time.Parse("2006-01-02", *targetEndDate)
		if err != nil {
			return nil, err
		}
		project.TargetEndDate = targetEnd
	}
	if actualEndDate != nil {
		actualEnd, err := time.Parse("2006-01-02", *actualEndDate)
		if err != nil {
			return nil, err
		}
		project.ActualEndDate = &actualEnd
	}
	project.ModifiedBy = user.ModifiedBy
	// Only managers can assign projects to leads. Before project is assigned,
	// attempt to fetch the assignee. If the assignee's role is not 'lead', return an error.
	var assignee *model.User
	if assignedTo != nil && user.Role == "manager" {
		assignee, err = c.repo.GetUserByID(ctx, *assignedTo)
		if err != nil {
			switch {
			case errors.Is(err, repository.ErrNotFound):
				return nil, ErrNotFound
			default:
				return nil, err
			}
		}
		if assignee.Role != "lead" {
			return nil, ErrInvalidRole
		}
		// Assign lead to project.
		project.AssignedTo = &assignee.ID
	}
	v := validator.New()
	if project.Validate(v); !v.Valid() {
		return nil, failedValidationErr(v.Errors)
	}
	err = c.repo.UpdateProject(ctx, project)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrEditConflict):
			return nil, ErrEditConflict
		default:
			return nil, err
		}
	}
	// Send email notification to assigned lead if project is assigned.
	if assignedTo != nil && user.Role == "manager" {
		data := map[string]string{
			"name":        assignee.Name,
			"projectID":   strconv.Itoa(int(project.ID)),
			"projectName": project.Name,
		}
		c.SendEmail(data, assignee.Email, "project_assign.tmpl")
	}
	return project, nil
}

func (c *Controller) DeleteProject(ctx context.Context, id int64) error {
	err := c.repo.DeleteProject(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return ErrNotFound
		default:
			return err
		}
	}
	return nil
}

func (c *Controller) GetProjectUsers(ctx context.Context, projectID int64, role string, filters model.Filters, v *validator.Validator) ([]*model.User, model.Metadata, error) {
	if filters.Validate(v); !v.Valid() {
		return nil, model.Metadata{}, failedValidationErr(v.Errors)
	}
	users, metadata, err := c.repo.GetProjectUsers(ctx, projectID, role, filters)
	if err != nil {
		return nil, model.Metadata{}, err
	}
	return users, metadata, nil
}

func (c *Controller) GetProjectUser(ctx context.Context, projectID, userID int64) (*model.User, error) {
	user, err := c.repo.GetProjectUser(ctx, projectID, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return user, nil
}
