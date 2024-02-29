package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/emzola/issuetracker/internal/model"
	"github.com/emzola/issuetracker/pkg/validator"
)

type issuesReportService interface {
	GetIssuesStatusReport(ctx context.Context, projectID int64) ([]*model.IssuesStatus, error)
	GetIssuesAssigneeReport(ctx context.Context, projectID int64) ([]*model.IssuesAssignee, error)
	GetIssuesReporterReport(ctx context.Context, projectID int64) ([]*model.IssuesReporter, error)
	GetIssuesPriorityLevelReport(ctx context.Context, projectID int64) ([]*model.IssuesPriority, error)
	GetIssuesTargetDateReport(ctx context.Context, projectID int64) ([]*model.IssuesTargetDate, error)
}

// GetIssuesStatusReport godoc
// @Summary Get report of issue status for a project
// @Description This endpoint gets report of issue status for a project
// @Tags issuesreport
// @Produce json
// @Param token header string true "Bearer token"
// @Param project_id query string true "Query string param for project_id"
// @Success 200 {array} model.IssuesStatus
// @Failure 500
// @Router /v1/issuesreport/status [get]
func (h *Handler) getIssuesStatusReport(w http.ResponseWriter, r *http.Request) {
	var queryParams struct {
		ProjectID int64
	}
	v := validator.New()
	qs := r.URL.Query()
	queryParams.ProjectID = int64(h.readInt(qs, "project_id", 0, v))
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	statuses, err := h.service.GetIssuesStatusReport(ctx, queryParams.ProjectID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}
	err = h.encodeJSON(w, http.StatusOK, envelop{"report": statuses}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

// GetIssuesAssigneeReport godoc
// @Summary Get report of issue assignees for a project
// @Description This endpoint gets report of issue assignees for a project
// @Tags issuesreport
// @Produce json
// @Param token header string true "Bearer token"
// @Param project_id query string true "Query string param for project_id"
// @Success 200 {array} model.IssuesAssignee
// @Failure 500
// @Router /v1/issuesreport/assignee [get]
func (h *Handler) getIssuesAssigneeReport(w http.ResponseWriter, r *http.Request) {
	var queryParams struct {
		ProjectID int64
	}
	v := validator.New()
	qs := r.URL.Query()
	queryParams.ProjectID = int64(h.readInt(qs, "project_id", 0, v))
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	assignees, err := h.service.GetIssuesAssigneeReport(ctx, queryParams.ProjectID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}
	err = h.encodeJSON(w, http.StatusOK, envelop{"report": assignees}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

// GetIssuesReporterReport godoc
// @Summary Get report of issues reporter for a project
// @Description This endpoint gets report of issues reporter for a project
// @Tags issuesreport
// @Produce json
// @Param token header string true "Bearer token"
// @Param project_id query string true "Query string param for project_id"
// @Success 200 {array} model.IssuesReporter
// @Failure 500
// @Router /v1/issuesreport/reporter [get]
func (h *Handler) getIssuesReporterReport(w http.ResponseWriter, r *http.Request) {
	var queryParams struct {
		ProjectID int64
	}
	v := validator.New()
	qs := r.URL.Query()
	queryParams.ProjectID = int64(h.readInt(qs, "project_id", 0, v))
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	reporters, err := h.service.GetIssuesReporterReport(ctx, queryParams.ProjectID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}
	err = h.encodeJSON(w, http.StatusOK, envelop{"report": reporters}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

// GetIssuesPriorityLevelReport godoc
// @Summary Get report of issues priority level for a project
// @Description This endpoint gets report of issues priority level for a project
// @Tags issuesreport
// @Produce json
// @Param token header string true "Bearer token"
// @Param project_id query string true "Query string param for project_id"
// @Success 200 {array} model.IssuesPriority
// @Failure 500
// @Router /v1/issuesreport/priority [get]
func (h *Handler) getIssuesPriorityLevelReport(w http.ResponseWriter, r *http.Request) {
	var queryParams struct {
		ProjectID int64
	}
	v := validator.New()
	qs := r.URL.Query()
	queryParams.ProjectID = int64(h.readInt(qs, "project_id", 0, v))
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	priorityLevels, err := h.service.GetIssuesPriorityLevelReport(ctx, queryParams.ProjectID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}
	err = h.encodeJSON(w, http.StatusOK, envelop{"report": priorityLevels}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

// GetIssuesTargetDateReport godoc
// @Summary Get report of issues target date for a project
// @Description This endpoint gets report of issue target date for a project
// @Tags issuesreport
// @Produce json
// @Param token header string true "Bearer token"
// @Param project_id query string true "Query string param for project_id"
// @Success 200 {array} model.IssuesTargetDate
// @Failure 500
// @Router /v1/issuesreport/date [get]
func (h *Handler) getIssuesTargetDateReport(w http.ResponseWriter, r *http.Request) {
	var queryParams struct {
		ProjectID int64
	}
	v := validator.New()
	qs := r.URL.Query()
	queryParams.ProjectID = int64(h.readInt(qs, "project_id", 0, v))
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	targetDates, err := h.service.GetIssuesTargetDateReport(ctx, queryParams.ProjectID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}
	err = h.encodeJSON(w, http.StatusOK, envelop{"report": targetDates}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}
