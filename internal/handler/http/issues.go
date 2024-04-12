package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/emzola/issuetracker/internal/controller/issuetracker"
	"github.com/emzola/issuetracker/pkg/model"
	"github.com/emzola/issuetracker/pkg/validator"
)

// CreateIssue godoc
// @Summary Create a new issue
// @Description Create a new issue with the request payload
// @Tags issues
// @Accept  json
// @Produce json
// @Param token header string true "Bearer token"
// @Param payload body createIssuePayload true "Request payload"
// @Success 201 {object} model.Issue
// @Failure 403
// @Failure 404
// @Failure 422
// @Failure 500
// @Router /v1/issues [post]
func (h *Handler) createIssue(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Title                string `json:"title"`
		Description          string `json:"description"`
		ProjectID            int64  `json:"project_id"`
		AssignedTo           *int64 `json:"assigned_to"`
		Priority             string `json:"priority"`
		TargetResolutionDate string `json:"target_resolution_date"`
	}
	err := h.decodeJSON(w, r, &requestPayload)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	userFromContext := h.contextGetUser(r)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	issue, err := h.ctrl.CreateIssue(ctx, requestPayload.Title, requestPayload.Description, userFromContext.ID, requestPayload.ProjectID, requestPayload.AssignedTo, requestPayload.Priority, requestPayload.TargetResolutionDate, userFromContext.Name, userFromContext.Name)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		case errors.Is(err, issuetracker.ErrNotFound):
			h.notFoundResponse(w, r)
		case errors.Is(err, issuetracker.ErrFailedValidation):
			h.failedValidationResponse(w, r, err)
		case errors.Is(err, issuetracker.ErrInvalidRole):
			h.invalidRoleResponse(w, r)
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

// GetIssue godoc
// @Summary Get issue by ID
// @Description This endpoint gets an issue by ID
// @Tags issues
// @Produce json
// @Param token header string true "Bearer token"
// @Param issue_id path string true "ID of issue to get"
// @Success 200 {object} model.Issue
// @Failure 404
// @Failure 500
// @Router /v1/issues/{issue_id} [get]
func (h *Handler) getIssue(w http.ResponseWriter, r *http.Request) {
	issueID, err := h.readIDParam(r, "issue_id")
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	issue, err := h.ctrl.GetIssue(ctx, issueID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		case errors.Is(err, issuetracker.ErrNotFound):
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

// GetAllIssues godoc
// @Summary Get all issues
// @Description This endpoint gets all issues
// @Tags issues
// @Produce json
// @Param token header string true "Bearer token"
// @Param title query string false "Query string param for title"
// @Param reported_date query string false "Query string param for reported_date"
// @Param project_id query string false "Query string param for project_id"
// @Param assigned_to query string false "Query string param for assigned_to"
// @Param status query string false "Query string param for status"
// @Param priority query string false "Query string param for priority"
// @Param page query string false "Query string param for pagination (min 1)"
// @Param page_size query string false "Query string param for pagination (max 100)"
// @Param sort query string false "Sort by asc or desc order. Asc: id, title, reported_date, project_id, assigned_to, status, priority | Desc: -id, -title, -reported_date, -project_id, -assigned_to, -status, -priority"
// @Success 200 {array} model.Issue
// @Failure 422
// @Failure 500
// @Router /v1/issues [get]
func (h *Handler) getAllIssues(w http.ResponseWriter, r *http.Request) {
	var queryParams struct {
		Title        string
		ReportedDate string
		ProjectID    int64
		AssignedTo   int64
		Status       string
		Priority     string
		Filters      model.Filters
	}
	v := validator.New()
	qs := r.URL.Query()
	queryParams.Title = h.readString(qs, "title", "")
	queryParams.ReportedDate = h.readString(qs, "reported_date", "")
	queryParams.ProjectID = int64(h.readInt(qs, "project_id", 0, v))
	queryParams.AssignedTo = int64(h.readInt(qs, "assigned_to", 0, v))
	queryParams.Status = h.readString(qs, "status", "")
	queryParams.Priority = h.readString(qs, "priority", "")
	queryParams.Filters.Page = h.readInt(qs, "page", 1, v)
	queryParams.Filters.PageSize = h.readInt(qs, "page_size", 20, v)
	queryParams.Filters.Sort = h.readString(qs, "sort", "id")
	queryParams.Filters.SortSafelist = []string{"id", "title", "reported_date", "project_id", "assigned_to", "status", "priority", "-id", "-title", "-reported_date", "-project_id", "-assigned_to", "-status", "-priority"}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	issues, metadata, err := h.ctrl.GetAllIssues(ctx, queryParams.Title, queryParams.ReportedDate, queryParams.ProjectID, queryParams.AssignedTo, queryParams.Status, queryParams.Priority, queryParams.Filters, v)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		case errors.Is(err, issuetracker.ErrFailedValidation):
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

// UpdateIssue godoc
// @Summary Update an issue
// @Description This endpoint updates an issue
// @Tags issues
// @Accept  json
// @Produce json
// @Param token header string true "Bearer token"
// @Param payload body updateIsssuePayload true "Request payload"
// @Param issue_id path string true "ID of issue to update"
// @Success 200 {object} model.Issue
// @Failure 400
// @Failure 403
// @Failure 404
// @Failure 409
// @Failure 422
// @Failure 500
// @Router /v1/issues/{issue_id} [patch]
func (h *Handler) updateIssue(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Title                *string `json:"title"`
		Description          *string `json:"description"`
		AssignedTo           *int64  `json:"assigned_to"`
		Status               *string `json:"status"`
		Priority             *string `json:"priority"`
		TargetResolutionDate *string `json:"target_resolution_date"`
		Progress             *string `json:"progress"`
		ActualResolutionDate *string `json:"actual_resolution_date"`
		ResolutionSummary    *string `json:"resolution_summary"`
	}
	issueID, err := h.readIDParam(r, "issue_id")
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}
	err = h.decodeJSON(w, r, &requestPayload)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	userFromContext := h.contextGetUser(r)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	issue, err := h.ctrl.UpdateIssue(ctx, issueID, requestPayload.Title, requestPayload.Description, requestPayload.AssignedTo, requestPayload.Status, requestPayload.Priority, requestPayload.TargetResolutionDate, requestPayload.Progress, requestPayload.ActualResolutionDate, requestPayload.ResolutionSummary, userFromContext)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		case errors.Is(err, issuetracker.ErrNotPermitted):
			h.notPermittedResponse(w, r)
		case errors.Is(err, issuetracker.ErrNotFound):
			h.notFoundResponse(w, r)
		case errors.Is(err, issuetracker.ErrInvalidRole):
			h.invalidRoleResponse(w, r)
		case errors.Is(err, issuetracker.ErrFailedValidation):
			h.failedValidationResponse(w, r, err)
		case errors.Is(err, issuetracker.ErrEditConflict):
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

// DeleteIssue godoc
// @Summary Delete an issue
// @Description This endpoint deletes an issue
// @Tags issues
// @Produce json
// @Param token header string true "Bearer token"
// @Param issue_id path string true "ID of issue to delete"
// @Success 200
// @Failure 404
// @Failure 500
// @Router /v1/issues/{issue_id} [delete]
func (h *Handler) deleteIssue(w http.ResponseWriter, r *http.Request) {
	issueID, err := h.readIDParam(r, "issue_id")
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	err = h.ctrl.DeleteIssue(ctx, issueID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		case errors.Is(err, issuetracker.ErrNotFound):
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
