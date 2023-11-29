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

type issueRepository interface {
	CreateIssue(ctx context.Context, issue *model.Issue) error
	GetIssue(ctx context.Context, id int64) (*model.Issue, error)
	GetAllIssues(ctx context.Context, title string, reportedDate time.Time, projectID, assignedTo int64, status, priority string, filters model.Filters) ([]*model.Issue, model.Metadata, error)
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
	// Send email notification to assigned user if issue is assigned.
	if assignedTo != nil {
		member, err := s.repo.GetUserByID(ctx, *assignedTo)
		if err != nil {
			switch {
			case errors.Is(err, repository.ErrNotFound):
				return nil, ErrNotFound
			default:
				return nil, err
			}
		}
		data := map[string]string{
			"name":          member.Name,
			"issueID":       strconv.Itoa(int(issue.ID)),
			"issueTitle":    issue.Title,
			"issuePriority": issue.Priority,
		}
		s.SendEmail(data, member.Email, "issue_assign.tmpl")
	}
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

// GetAllIssues returns a paginated list of all issues. List can be filtered and sorted.
func (s *Service) GetAllIssues(ctx context.Context, title, reportedDate string, projectID, assignedTo int64, status, priority string, filters model.Filters, v *validator.Validator) ([]*model.Issue, model.Metadata, error) {
	if filters.Validate(v); !v.Valid() {
		return nil, model.Metadata{}, failedValidationErr(v.Errors)
	}
	var reported time.Time
	var err error
	if reportedDate != "" {
		reported, err = time.Parse("2006-01-02", reportedDate)
		if err != nil {
			return nil, model.Metadata{}, err
		}
	}
	issues, metadata, err := s.repo.GetAllIssues(ctx, title, reported, projectID, assignedTo, status, priority, filters)
	if err != nil {
		return nil, model.Metadata{}, err
	}
	return issues, metadata, nil
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
	// Send email notification to assigned user if issue is assigned.
	if assignedTo != nil {
		member, err := s.repo.GetUserByID(ctx, *assignedTo)
		if err != nil {
			switch {
			case errors.Is(err, repository.ErrNotFound):
				return nil, ErrNotFound
			default:
				return nil, err
			}
		}
		data := map[string]string{
			"name":          member.Name,
			"issueID":       strconv.Itoa(int(issue.ID)),
			"issueTitle":    issue.Title,
			"issuePriority": issue.Priority,
		}
		s.SendEmail(data, member.Email, "issue_assign.tmpl")
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
