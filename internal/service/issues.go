package service

import (
	"context"
	"time"

	"github.com/emzola/issuetracker/internal/model"
	"github.com/emzola/issuetracker/pkg/validator"
)

type issueRepository interface {
	CreateIssue(ctx context.Context, issue *model.Issue) error
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
