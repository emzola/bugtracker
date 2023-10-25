package service

import (
	"context"
	"errors"
	"time"

	"github.com/emzola/bugtracker/internal/model"
	"github.com/emzola/bugtracker/internal/repository"
	"github.com/emzola/bugtracker/pkg/validator"
)

type ProjectRepository interface {
	CreateProject(ctx context.Context, project *model.Project) error
	GetProject(ctx context.Context, id int64) (*model.Project, error)
	// GetAllProjects(ctx context.Context) ([]*model.Project, model.Metadata, error)
	UpdateProject(ctx context.Context, project *model.Project) error
	// DeleteProject(ctx context.Context, id int64) error
}

// CreateProject adds a new project.
func (s *Service) CreateProject(ctx context.Context, name, description, owner, startDate, endDate string, publicAccess bool, createdBy, modifiedBy string) (*model.Project, error) {
	project := &model.Project{
		Name:         name,
		Description:  description,
		Owner:        owner,
		Status:       "active",
		PublicAccess: publicAccess,
		CreatedBy:    createdBy,
		ModifiedBy:   modifiedBy,
	}
	if startDate != "" {
		// Convert startDate from string to time.Time and assign it to project.
		start, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return nil, err
		}
		project.StartDate = &start
	}
	if endDate != "" {
		// Convert endDate from string to time.Time and assign it to project.
		end, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			return nil, err
		}
		project.EndDate = &end
	}
	v := validator.New()
	if project.Validate(v); !v.Valid() {
		return nil, failedValidationErr(v.Errors)
	}
	err := s.repo.CreateProject(ctx, project)
	if err != nil {
		return nil, err
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

// UpdateProject updates a project's details.
func (s *Service) UpdateProject(ctx context.Context, id int64, name, description, owner, status, startDate, endDate, completedOn *string, publicAccess *bool) (*model.Project, error) {
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
	if owner != nil {
		project.Owner = *owner
	}
	if status != nil {
		project.Status = *status
	}
	if startDate != nil {
		start, err := time.Parse("2006-01-02", *startDate)
		if err != nil {
			return nil, err
		}
		project.StartDate = &start
	}
	if endDate != nil {
		end, err := time.Parse("2006-01-02", *endDate)
		if err != nil {
			return nil, err
		}
		project.EndDate = &end
	}
	if completedOn != nil {
		completed, err := time.Parse("2006-01-02", *completedOn)
		if err != nil {
			return nil, err
		}
		project.CompletedOn = &completed
	}
	if publicAccess != nil {
		project.PublicAccess = *publicAccess
	}
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
