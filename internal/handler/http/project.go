package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/emzola/bugtracker/internal/model"
	"github.com/emzola/bugtracker/internal/service"
)

type projectService interface {
	CreateProject(ctx context.Context, name, description, owner, startDate, endDate string, publicAccess bool, createdBy, modifiedBy string) (*model.Project, error)
	GetProject(ctx context.Context, id int64) (*model.Project, error)
}

func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Name         string `json:"name"`
		Description  string `json:"description,omitempty"`
		Owner        string `json:"owner"`
		StartDate    string `json:"start_date"`
		EndDate      string `json:"end_date,omitempty"`
		PublicAccess bool   `json:"public_access"`
	}
	err := h.decodeJSON(w, r, &requestBody)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	project, err := h.service.CreateProject(ctx, requestBody.Name, requestBody.Description, requestBody.Owner, requestBody.StartDate, requestBody.EndDate, requestBody.PublicAccess, "Emzo", "Emzo")
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFailedValidation):
			h.failedValidationResponse(w, r, err)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}
	header := make(http.Header)
	header.Set("Location", fmt.Sprintf("/v1/projects/%d", project.ID))
	err = h.encodeJSON(w, http.StatusCreated, envelop{"project": project}, header)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

func (h *Handler) getProject(w http.ResponseWriter, r *http.Request) {
	id, err := h.readIDParam(r, "projectId")
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	project, err := h.service.GetProject(ctx, id)
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
	err = h.encodeJSON(w, http.StatusOK, envelop{"project": project}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}
