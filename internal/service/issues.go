package service

import (
	"context"
	"errors"
	"time"

	"github.com/emzola/issuetracker/internal/model"
	"github.com/emzola/issuetracker/internal/repository"
	"github.com/emzola/issuetracker/pkg/validator"
)

type issueRepository interface {
	CreateIssue(ctx context.Context, issue *model.Issue) error
	GetIssue(ctx context.Context, id int64) (*model.Issue, error)
	UpdateIssue(ctx context.Context, issue *model.Issue) error
	DeleteIssue(ctx context.Context, id int64) error
}

// CreateIssue adds a new issue.
func (s *Service) CreateIssue(ctx context.Context, title, description, reportedDate string, reporterID, projectID int64, assignedTo *int64, priority, targetResolutionDate, createdBy, modifiedBy string) (*model.Issue, error) {
	if priority == "" {
		priority = "low"
	}
	issue := &model.Issue{
		Title:       title,
		Description: description,
		ReporterID:  reporterID,
		ProjectID:   projectID,
		AssignedTo:  assignedTo,
		Priority:    priority,
		Status:      "open",
		CreatedBy:   createdBy,
		ModifiedBy:  modifiedBy,
	}
	if reportedDate != "" {
		reported, err := time.Parse("2006-01-02", reportedDate)
		if err != nil {
			return nil, err
		}
		issue.ReportedDate = reported
	}
	if targetResolutionDate != "" {
		targetResolution, err := time.Parse("2006-01-02", targetResolutionDate)
		if err != nil {
			return nil, err
		}
		issue.TargetResolutionDate = targetResolution
	}
	v := validator.New()
	if issue.Validate(v); !v.Valid() {
		return nil, failedValidationErr(v.Errors)
	}
	err := s.repo.CreateIssue(ctx, issue)
	if err != nil {
		return nil, err
	}

	// WORK ON FEATURE TO SEND EMAIL TO ASSIGNEE!!!!!!

	return issue, nil
}

// GetIssue retrieves an issue by id.
func (s *Service) GetIssue(ctx context.Context, id int64) (*model.Issue, error) {
	issue, err := s.repo.GetIssue(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return issue, nil
}

// UpdateIssue updates an issue's details.
func (s *Service) UpdateIssue(ctx context.Context, id int64, title, description *string, assignedTo *int64, priority, targetResolutionDate, progress, actualResolutionDate, resolutionSummary *string, modifiedBy string) (*model.Issue, error) {
	issue, err := s.repo.GetIssue(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	if title != nil {
		issue.Title = *title
	}
	if description != nil {
		issue.Description = *description
	}
	if assignedTo != nil {
		issue.AssignedTo = assignedTo
	}
	if priority != nil {
		issue.Priority = *priority
	}
	if targetResolutionDate != nil {
		targetResolution, err := time.Parse("2006-01-02", *targetResolutionDate)
		if err != nil {
			return nil, err
		}
		issue.TargetResolutionDate = targetResolution
	}
	if progress != nil {
		issue.Progress = *progress
	}
	if actualResolutionDate != nil {
		actualResolution, err := time.Parse("2006-01-02", *actualResolutionDate)
		if err != nil {
			return nil, err
		}
		issue.ActualResolutionDate = &actualResolution
		issue.Status = "closed"
	}
	if resolutionSummary != nil {
		issue.ResolutionSummary = *resolutionSummary
	}
	issue.ModifiedBy = modifiedBy
	v := validator.New()
	if issue.Validate(v); !v.Valid() {
		return nil, failedValidationErr(v.Errors)
	}
	err = s.repo.UpdateIssue(ctx, issue)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrEditConflict):
			return nil, ErrEditConflict
		default:
			return nil, err
		}
	}
	return issue, nil
}

// DeleteIssue deletes an issue by id.
func (s *Service) DeleteIssue(ctx context.Context, id int64) error {
	err := s.repo.DeleteIssue(ctx, id)
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