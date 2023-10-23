package service

import (
	"context"

	"github.com/emzola/bugtracker/internal/model"
	"github.com/emzola/bugtracker/pkg/validator"
)

type ProjectRepository interface {
	CreateProject(ctx context.Context, project *model.Project) error
	// GetProject(ctx context.Context, id int64) (*model.Project, error)
	// GetAllProjects(ctx context.Context) ([]*model.Project, model.Metadata, error)
	// UpdateProject(ctx context.Context, project *model.Project) error
	// DeleteProject(ctx context.Context, id int64) error
}

func (s *Service) CreateProject(ctx context.Context, name, description, owner, startDate, endDate string, publicAccess bool, createdBy, modifiedBy string) (*model.Project, error) {
	project := &model.Project{
		Name:         name,
		Description:  description,
		Owner:        owner,
		Status:       "active",
		StartDate:    startDate,
		EndDate:      endDate,
		PublicAccess: publicAccess,
		CreatedBy:    createdBy,
		ModifiedBy:   modifiedBy,
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
