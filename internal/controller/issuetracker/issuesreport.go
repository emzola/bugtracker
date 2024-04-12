package issuetracker

import (
	"context"

	"github.com/emzola/issuetracker/pkg/model"
)

type issuesReportRepository interface {
	GetIssuesStatusReport(ctx context.Context, projectID int64) ([]*model.IssuesStatus, error)
	GetIssuesAssigneeReport(ctx context.Context, projectID int64) ([]*model.IssuesAssignee, error)
	GetIssuesReporterReport(ctx context.Context, projectID int64) ([]*model.IssuesReporter, error)
	GetIssuesPriorityLevelReport(ctx context.Context, projectID int64) ([]*model.IssuesPriority, error)
	GetIssuesTargetDateReport(ctx context.Context, projectID int64) ([]*model.IssuesTargetDate, error)
}

func (c *Controller) GetIssuesStatusReport(ctx context.Context, projectID int64) ([]*model.IssuesStatus, error) {
	statuses, err := c.repo.GetIssuesStatusReport(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return statuses, nil
}

func (c *Controller) GetIssuesAssigneeReport(ctx context.Context, projectID int64) ([]*model.IssuesAssignee, error) {
	assignees, err := c.repo.GetIssuesAssigneeReport(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return assignees, nil
}

func (c *Controller) GetIssuesReporterReport(ctx context.Context, projectID int64) ([]*model.IssuesReporter, error) {
	reporters, err := c.repo.GetIssuesReporterReport(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return reporters, nil
}

func (c *Controller) GetIssuesPriorityLevelReport(ctx context.Context, projectID int64) ([]*model.IssuesPriority, error) {
	priorityLevels, err := c.repo.GetIssuesPriorityLevelReport(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return priorityLevels, nil
}

func (c *Controller) GetIssuesTargetDateReport(ctx context.Context, projectID int64) ([]*model.IssuesTargetDate, error) {
	targetDates, err := c.repo.GetIssuesTargetDateReport(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return targetDates, nil
}
