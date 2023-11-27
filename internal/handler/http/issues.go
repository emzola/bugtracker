package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/emzola/issuetracker/internal/model"
	"github.com/emzola/issuetracker/internal/service"
)

type issueService interface {
	CreateIssue(ctx context.Context, title, description, reportedDate string, reporterID, projectID int64, assignedTo *int64, priority, targetResolutionDate, createdBy, modifiedBy string) (*model.Issue, error)
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
