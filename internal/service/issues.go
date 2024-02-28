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

func (s *Service) CreateIssue(ctx context.Context, title, description string, reporterID, projectID int64, assignedTo *int64, priority, targetResolutionDate, createdBy, modifiedBy string) (*model.Issue, error) {
	if priority == "" {
		priority = "low"
	}
	issue := &model.Issue{
		Title:       title,
		Description: description,
		ReporterID:  reporterID,
		ProjectID:   projectID,
		Priority:    priority,
		Status:      "open",
		CreatedBy:   createdBy,
		ModifiedBy:  modifiedBy,
	}
	if targetResolutionDate != "" {
		targetResolution, err := time.Parse("2006-01-02", targetResolutionDate)
		if err != nil {
			return nil, err
		}
		issue.TargetResolutionDate = targetResolution
	}
	// Issues can only be assigned to users associated with a project with role 'member'.
	// Before issue is assigned, attempt to fetch the assignee. If the assignee's role is
	// not 'member', return an error.
	var assignee *model.User
	var err error
	if assignedTo != nil {
		assignee, err = s.repo.GetProjectUser(ctx, issue.ProjectID, *assignedTo)
		if err != nil {
			switch {
			case errors.Is(err, repository.ErrNotFound):
				return nil, ErrNotFound
			default:
				return nil, err
			}
		}
		if assignee.Role != "member" {
			return nil, ErrInvalidRole
		}
		// Assign issue to member
		issue.AssignedTo = &assignee.ID
	}
	v := validator.New()
	if issue.Validate(v); !v.Valid() {
		return nil, failedValidationErr(v.Errors)
	}
	err = s.repo.CreateIssue(ctx, issue)
	if err != nil {
		return nil, err
	}
	// Send email notification to assigned user if issue is assigned.
	if assignedTo != nil {
		data := map[string]string{
			"name":          assignee.Name,
			"issueID":       strconv.Itoa(int(issue.ID)),
			"issueTitle":    issue.Title,
			"issuePriority": issue.Priority,
		}
		s.SendEmail(data, assignee.Email, "issue_assign.tmpl")
	}
	return issue, nil
}

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

func (s *Service) UpdateIssue(ctx context.Context, id int64, title, description *string, assignedTo *int64, status, priority, targetResolutionDate, progress, actualResolutionDate, resolutionSummary *string, user *model.User) (*model.Issue, error) {
	issue, err := s.repo.GetIssue(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	// Check whether user has permission to update issue. Besides managers and leads,
	// members can update issue details only if it's assigned to or reported by them.
	if user.Role == "member" && *issue.AssignedTo != user.ID && issue.ReporterID != user.ID {
		return nil, ErrNotPermitted
	}
	// At this point, update issue as usual.
	if title != nil {
		issue.Title = *title
	}
	if description != nil {
		issue.Description = *description
	}
	// Issues can only be assigned to users with role 'member'.
	// Before issue is assigned, attempt to fetch the assignee.
	// If the assignee's role is not 'member', return an error.
	var assignee *model.User
	if assignedTo != nil {
		assignee, err = s.repo.GetProjectUser(ctx, issue.ProjectID, *assignedTo)
		if err != nil {
			switch {
			case errors.Is(err, repository.ErrNotFound):
				return nil, ErrNotFound
			default:
				return nil, err
			}
		}
		if assignee.Role != "member" {
			return nil, ErrInvalidRole
		}
		// Assign issue to member
		issue.AssignedTo = &assignee.ID
	}
	if status != nil {
		issue.Status = *status
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
	issue.ModifiedBy = user.ModifiedBy
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
	// Send email notification to assignee if issue is assigned.
	if assignedTo != nil {
		data := map[string]string{
			"name":          assignee.Name,
			"issueID":       strconv.Itoa(int(issue.ID)),
			"issueTitle":    issue.Title,
			"issuePriority": issue.Priority,
		}
		s.SendEmail(data, assignee.Email, "issue_assign.tmpl")
	}
	return issue, nil
}

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
