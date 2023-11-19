package service

import (
	"context"
	"errors"
	"time"

	"github.com/emzola/bugtracker/internal/model"
	"github.com/emzola/bugtracker/internal/repository"
	"github.com/emzola/bugtracker/pkg/validator"
)

type projectRepository interface {
	CreateProject(ctx context.Context, project *model.Project) error
	GetProject(ctx context.Context, id int64) (*model.Project, error)
	GetAllProjects(ctx context.Context, name string, startDate, targetEndDate, actualEndDate time.Time, createdby string, filters model.Filters) ([]*model.Project, model.Metadata, error)
	UpdateProject(ctx context.Context, project *model.Project) error
	DeleteProject(ctx context.Context, id int64) error
}

// CreateProject adds a new project.
func (s *Service) CreateProject(ctx context.Context, name, description, startDate, targetEndDate, createdBy, modifiedBy string) (*model.Project, error) {
	project := &model.Project{
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
		ModifiedBy:  modifiedBy,
	}
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, err
	}
	project.StartDate = start
	targetEnd, err := time.Parse("2006-01-02", targetEndDate)
	if err != nil {
		return nil, err
	}
	project.TargetEndDate = targetEnd
	v := validator.New()
	if project.Validate(v); !v.Valid() {
		return nil, failedValidationErr(v.Errors)
	}
	err = s.repo.CreateProject(ctx, project)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateKey):
			v.AddError("name", "a project with this name already exists")
			return nil, failedValidationErr(v.Errors)
		default:
			return nil, err
		}
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
func (s *Service) GetAllProjects(ctx context.Context, name, startDate, targetEndDate, actualEndDate, createdby string, filters model.Filters, v *validator.Validator) ([]*model.Project, model.Metadata, error) {
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
	projects, metadata, err := s.repo.GetAllProjects(ctx, name, start, targetEnd, actualEnd, createdby, filters)
	if err != nil {
		return nil, model.Metadata{}, err
	}
	return projects, metadata, nil
}

// UpdateProject updates a project's details.
func (s *Service) UpdateProject(ctx context.Context, id int64, name, description, startDate, targetEndDate, actualEndDate *string, modifiedBy string) (*model.Project, error) {
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
