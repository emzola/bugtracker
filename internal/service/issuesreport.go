package service

import (
	"context"

	"github.com/emzola/issuetracker/internal/model"
)

type issuesReportRepository interface {
	GetIssuesStatusReport(ctx context.Context, projectID int64) ([]*model.IssuesStatus, error)
	GetIssuesAssigneeReport(ctx context.Context, projectID int64) ([]*model.IssuesAssignee, error)
	GetIssuesReporterReport(ctx context.Context, projectID int64) ([]*model.IssuesReporter, error)
	GetIssuesPriorityLevelReport(ctx context.Context, projectID int64) ([]*model.IssuesPriority, error)
}

// GetIssuesReportStatus retrieves issues status report for a specific project.
func (s *Service) GetIssuesStatusReport(ctx context.Context, projectID int64) ([]*model.IssuesStatus, error) {
	statuses, err := s.repo.GetIssuesStatusReport(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return statuses, nil
}

// GetIssuesAssigneeReport retrieves issues assignee report for a specific project.
func (s *Service) GetIssuesAssigneeReport(ctx context.Context, projectID int64) ([]*model.IssuesAssignee, error) {
	assignees, err := s.repo.GetIssuesAssigneeReport(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return assignees, nil
}

// GetIssuesReporterReport retrieves issues reporter report for a specific project.
func (s *Service) GetIssuesReporterReport(ctx context.Context, projectID int64) ([]*model.IssuesReporter, error) {
	reporters, err := s.repo.GetIssuesReporterReport(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return reporters, nil
}

// GetIssuesPriorityLevelReport retrieves issues priority level report for a specific project.
func (s *Service) GetIssuesPriorityLevelReport(ctx context.Context, projectID int64) ([]*model.IssuesPriority, error) {
	priorityLevels, err := s.repo.GetIssuesPriorityLevelReport(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return priorityLevels, nil
}
