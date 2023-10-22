package service

import (
	"context"
	"time"

	"github.com/emzola/bugtracker/pkg/model"
)

type ProjectRepository interface {
	CreateProject(ctx context.Context, project *model.Project) error
	// GetProject(ctx context.Context, id int64) (*model.Project, error)
	// GetAllProjects(ctx context.Context) ([]*model.Project, model.Metadata, error)
	// UpdateProject(ctx context.Context, project *model.Project) error
	// DeleteProject(ctx context.Context, id int64) error
}

func (s *Service) CreateProject(ctx context.Context, name, description string, startDate, endDate time.Time, publicAccess bool, owner, createdBy string) (*model.Project, error) {
	project := &model.Project{
		Name:         name,
		Description:  description,
		StartDate:    startDate,
		EndDate:      endDate,
		PublicAccess: publicAccess,
		Owner:        owner,
		CreatedBy:    createdBy,
	}
	// Handle validation first before proceeding
	err := s.repo.CreateProject(ctx, project)
	if err != nil {
		return nil, err
	}
	return project, nil
}
