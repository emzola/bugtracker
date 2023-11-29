package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/emzola/issuetracker/internal/model"
	"github.com/emzola/issuetracker/internal/repository"
	"github.com/emzola/issuetracker/pkg/validator"
)

type projectRepository interface {
	CreateProject(ctx context.Context, project *model.Project) error
	GetProject(ctx context.Context, id int64) (*model.Project, error)
	GetAllProjects(ctx context.Context, name string, startDate, targetEndDate, actualEndDate time.Time, assignedTo int64, createdby string, filters model.Filters) ([]*model.Project, model.Metadata, error)
	UpdateProject(ctx context.Context, project *model.Project) error
	DeleteProject(ctx context.Context, id int64) error
}

// CreateProject adds a new project.
func (s *Service) CreateProject(ctx context.Context, name, description, startDate, targetEndDate string, assignedTo *int64, createdBy, modifiedBy string) (*model.Project, error) {
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
	if assignedTo != nil {
		project.AssignedTo = assignedTo
	}
	v := validator.New()
	if project.Validate(v); !v.Valid() {
		return nil, failedValidationErr(v.Errors)
	}
	err := s.repo.CreateProject(ctx, project)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateKey):
			v.AddError("name", "a project with this name already exists")
			return nil, failedValidationErr(v.Errors)
		default:
			return nil, err
		}
	}
	// Send email notification to assigned project lead if project is assigned.
	if assignedTo != nil {
		projectLead, err := s.repo.GetUserByID(ctx, *assignedTo)
		if err != nil {
			switch {
			case errors.Is(err, repository.ErrNotFound):
				return nil, ErrNotFound
			default:
				return nil, err
			}
		}
		data := map[string]string{
			"name":        projectLead.Name,
			"projectID":   strconv.Itoa(int(project.ID)),
			"projectName": project.Name,
		}
		s.SendEmail(data, projectLead.Email, "project_assign.tmpl")
	}
	return project, nil
}

// GetProject retrieves a project by id.
func (s *Service) GetProject(ctx context.Context, id int64) (*model.Project, error) {
	project, err := s.repo.GetProject(ctx, id)
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

// GetAllProjects returns a paginated list of all projects.
// List can be filtered and sorted.
func (s *Service) GetAllProjects(ctx context.Context, name, startDate, targetEndDate, actualEndDate string, AssignedTo int64, createdBy string, filters model.Filters, v *validator.Validator) ([]*model.Project, model.Metadata, error) {
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
	projects, metadata, err := s.repo.GetAllProjects(ctx, name, start, targetEnd, actualEnd, AssignedTo, createdBy, filters)
	if err != nil {
		return nil, model.Metadata{}, err
	}
	return projects, metadata, nil
}

// UpdateProject updates a project's details.
func (s *Service) UpdateProject(ctx context.Context, id int64, name, description, startDate, targetEndDate, actualEndDate *string, assignedTo *int64, modifiedBy string) (*model.Project, error) {
	project, err := s.repo.GetProject(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
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
	if assignedTo != nil {
		project.AssignedTo = assignedTo
	}
	project.ModifiedBy = modifiedBy
	v := validator.New()
	if project.Validate(v); !v.Valid() {
		return nil, failedValidationErr(v.Errors)
	}
	err = s.repo.UpdateProject(ctx, project)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrEditConflict):
			return nil, ErrEditConflict
		default:
			return nil, err
		}
	}
	// Send email notification to assigned project lead if project is assigned.
	if assignedTo != nil {
		projectLead, err := s.repo.GetUserByID(ctx, *assignedTo)
		if err != nil {
			switch {
			case errors.Is(err, repository.ErrNotFound):
				return nil, ErrNotFound
			default:
				return nil, err
			}
		}
		data := map[string]string{
			"name":        projectLead.Name,
			"projectID":   strconv.Itoa(int(project.ID)),
			"projectName": project.Name,
		}
		s.SendEmail(data, projectLead.Email, "project_assign.tmpl")
	}
	return project, nil
}

// DeleteProject deletes a project by id.
func (s *Service) DeleteProject(ctx context.Context, id int64) error {
	err := s.repo.DeleteProject(ctx, id)
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
