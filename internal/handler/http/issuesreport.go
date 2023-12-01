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
}

func (h *Handler) getIssuesStatusReport(w http.ResponseWriter, r *http.Request) {
	var requestQuery struct {
		ProjectID int64
	}
	v := validator.New()
	qs := r.URL.Query()
	requestQuery.ProjectID = int64(h.readInt(qs, "project_id", 0, v))
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	statuses, err := h.service.GetIssuesStatusReport(ctx, requestQuery.ProjectID)
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

func (h *Handler) getIssuesAssigneeReport(w http.ResponseWriter, r *http.Request) {
	var requestQuery struct {
		ProjectID int64
	}
	v := validator.New()
	qs := r.URL.Query()
	requestQuery.ProjectID = int64(h.readInt(qs, "project_id", 0, v))
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	assignees, err := h.service.GetIssuesAssigneeReport(ctx, requestQuery.ProjectID)
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

func (h *Handler) getIssuesReporterReport(w http.ResponseWriter, r *http.Request) {
	var requestQuery struct {
		ProjectID int64
	}
	v := validator.New()
	qs := r.URL.Query()
	requestQuery.ProjectID = int64(h.readInt(qs, "project_id", 0, v))
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	reporters, err := h.service.GetIssuesReporterReport(ctx, requestQuery.ProjectID)
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

func (h *Handler) getIssuesPriorityLevelReport(w http.ResponseWriter, r *http.Request) {
	var requestQuery struct {
		ProjectID int64
	}
	v := validator.New()
	qs := r.URL.Query()
	requestQuery.ProjectID = int64(h.readInt(qs, "project_id", 0, v))
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	priorityLevels, err := h.service.GetIssuesPriorityLevelReport(ctx, requestQuery.ProjectID)
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
