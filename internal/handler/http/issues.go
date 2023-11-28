package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/emzola/issuetracker/internal/model"
	"github.com/emzola/issuetracker/internal/service"
	"github.com/emzola/issuetracker/pkg/validator"
)

type issueService interface {
	CreateIssue(ctx context.Context, title, description, reportedDate string, reporterID, projectID int64, assignedTo *int64, priority, targetResolutionDate, createdBy, modifiedBy string) (*model.Issue, error)
	GetIssue(ctx context.Context, id int64) (*model.Issue, error)
	GetAllIssues(ctx context.Context, title, reportedDate string, projectID, assignedTo int64, status, priority string, filters model.Filters, v *validator.Validator) ([]*model.Issue, model.Metadata, error)
	UpdateIssue(ctx context.Context, id int64, title, description *string, assignedTo *int64, priority, targetResolutionDate, progress, actualResolutionDate, resolutionSummary *string, modifiedBy string) (*model.Issue, error)
	DeleteIssue(ctx context.Context, id int64) error
}

func (h *Handler) createIssue(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Title                string `json:"title"`
		Description          string `json:"description"`
		ReportedDate         string `json:"reported_date"`
		ProjectID            int64  `json:"project_id"`
		AssignedTo           *int64 `json:"assigned_to"`
		Priority             string `json:"priority"`
		TargetResolutionDate string `json:"target_resolution_date"`
	}
	err := h.decodeJSON(w, r, &requestBody)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	userFromContext := h.contextGetUser(r)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	issue, err := h.service.CreateIssue(ctx, requestBody.Title, requestBody.Description, requestBody.ReportedDate, userFromContext.ID, requestBody.ProjectID, requestBody.AssignedTo, requestBody.Priority, requestBody.TargetResolutionDate, userFromContext.Name, userFromContext.Name)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		case errors.Is(err, service.ErrNotFound):
			h.notFoundResponse(w, r)
		case errors.Is(err, service.ErrFailedValidation):
			h.failedValidationResponse(w, r, err)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}
	err = h.encodeJSON(w, http.StatusCreated, envelop{"issue": issue}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

func (h *Handler) getIssue(w http.ResponseWriter, r *http.Request) {
	issueID, err := h.readIDParam(r, "issue_id")
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	issue, err := h.service.GetIssue(ctx, issueID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		case errors.Is(err, service.ErrNotFound):
			h.notFoundResponse(w, r)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}
	err = h.encodeJSON(w, http.StatusOK, envelop{"issue": issue}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

func (h *Handler) getAllIssues(w http.ResponseWriter, r *http.Request) {
	var requestQuery struct {
		Title        string
		ReportedDate string
		projectID    int64
		AssignedTo   int64
		Status       string
		Priority     string
		Filters      model.Filters
	}
	v := validator.New()
	qs := r.URL.Query()
	requestQuery.Title = h.readString(qs, "title", "")
	requestQuery.ReportedDate = h.readString(qs, "reported_date", "")
	requestQuery.projectID = int64(h.readInt(qs, "project_id", 0, v))
	requestQuery.AssignedTo = int64(h.readInt(qs, "assigned_to", 0, v))
	requestQuery.Status = h.readString(qs, "status", "")
	requestQuery.Priority = h.readString(qs, "priority", "")
	requestQuery.Filters.Page = h.readInt(qs, "page", 1, v)
	requestQuery.Filters.PageSize = h.readInt(qs, "page_size", 20, v)
	requestQuery.Filters.Sort = h.readString(qs, "sort", "id")
	requestQuery.Filters.SortSafelist = []string{"id", "title", "reported_date", "project_id", "assigned_to", "status", "priority", "-id", "-title", "-reported_date", "-project_id", "-assigned_to", "-status", "-priority"}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	issues, metadata, err := h.service.GetAllIssues(ctx, requestQuery.Title, requestQuery.ReportedDate, requestQuery.projectID, requestQuery.AssignedTo, requestQuery.Status, requestQuery.Priority, requestQuery.Filters, v)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFailedValidation):
			h.failedValidationResponse(w, r, err)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}
	err = h.encodeJSON(w, http.StatusOK, envelop{"issues": issues, "metadata": metadata}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

func (h *Handler) updateIssue(w http.ResponseWriter, r *http.Request) {
	issueID, err := h.readIDParam(r, "issue_id")
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}
	var requestBody struct {
		Title                *string `json:"title"`
		Description          *string `json:"description"`
		AssignedTo           *int64  `json:"assigned_to"`
		Priority             *string `json:"priority"`
		TargetResolutionDate *string `json:"target_resolution_date"`
		Progress             *string `json:"progress"`
		ActualResolutionDate *string `json:"actual_resolution_date"`
		ResolutionSummary    *string `json:"resolution_summary"`
	}
	err = h.decodeJSON(w, r, &requestBody)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	userFromContext := h.contextGetUser(r)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	issue, err := h.service.UpdateIssue(ctx, issueID, requestBody.Title, requestBody.Description, requestBody.AssignedTo, requestBody.Priority, requestBody.TargetResolutionDate, requestBody.Progress, requestBody.ActualResolutionDate, requestBody.ResolutionSummary, userFromContext.Name)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		case errors.Is(err, service.ErrNotFound):
			h.notFoundResponse(w, r)
		case errors.Is(err, service.ErrFailedValidation):
			h.failedValidationResponse(w, r, err)
		case errors.Is(err, service.ErrEditConflict):
			h.editConflictResponse(w, r)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}
	err = h.encodeJSON(w, http.StatusOK, envelop{"issue": issue}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

func (h *Handler) deleteIssue(w http.ResponseWriter, r *http.Request) {
	issueID, err := h.readIDParam(r, "issue_id")
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	err = h.service.DeleteIssue(ctx, issueID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		case errors.Is(err, service.ErrNotFound):
			h.notFoundResponse(w, r)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}
	err = h.encodeJSON(w, http.StatusOK, envelop{"message": "issue successfully deleted"}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}
