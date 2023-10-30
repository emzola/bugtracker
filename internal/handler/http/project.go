package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/emzola/bugtracker/internal/model"
	"github.com/emzola/bugtracker/internal/service"
	"github.com/emzola/bugtracker/pkg/validator"
)

type projectService interface {
	CreateProject(ctx context.Context, name, description, owner, startDate, endDate, access, createdBy, modifiedBy string) (*model.Project, error)
	GetProject(ctx context.Context, id int64) (*model.Project, error)
	GetAllProjects(ctx context.Context, name, owner, status, createdby, modifiedBy, access string, filters model.Filters, v *validator.Validator) ([]*model.Project, model.Metadata, error)
	UpdateProject(ctx context.Context, id int64, name, description, owner, status, startDate, endDate, completedOn, access *string) (*model.Project, error)
	DeleteProject(ctx context.Context, id int64) error
}

func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Owner       string `json:"owner"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date,omitempty"`
		Access      string `json:"access"`
	}
	err := h.decodeJSON(w, r, &requestBody)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	project, err := h.service.CreateProject(ctx, requestBody.Name, requestBody.Description, requestBody.Owner, requestBody.StartDate, requestBody.EndDate, requestBody.Access, "Emzo", "Emzo")
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
	id, err := h.readIDParam(r, "project_id")
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

func (h *Handler) getAllProjects(w http.ResponseWriter, r *http.Request) {
	var rQuery struct {
		Name         string
		Owner        string
		Status       string
		StartDate    string
		EndDate      string
		CompletedOn  string
		CreatedOn    string
		LastModified string
		CreatedBy    string
		ModifiedBy   string
		Access       string
		Filters      model.Filters
	}
	v := validator.New()
	qs := r.URL.Query()
	rQuery.Name = h.readString(qs, "name", "")
	rQuery.Owner = h.readString(qs, "owner", "")
	rQuery.Status = h.readString(qs, "status", "")
	rQuery.CreatedBy = h.readString(qs, "created_by", "")
	rQuery.ModifiedBy = h.readString(qs, "modified_by", "")
	rQuery.Access = h.readString(qs, "access", "")
	rQuery.Filters.Page = h.readInt(qs, "page", 1, v)
	rQuery.Filters.PageSize = h.readInt(qs, "page_size", 20, v)
	rQuery.Filters.Sort = h.readString(qs, "sort", "id")
	rQuery.Filters.SortSafelist = []string{"id", "name", "owner", "start_date", "end_date", "completed_on", "created_on", "last_modified", "created_by", "modified_by", "-id", "-name", "-owner", "-start_date", "-end_date", "-completed_on", "-created_on", "-last_modified", "-created_by", "-modified_by"}
	rQuery.Filters.StatusSafelist = []string{"active", "in progress", "on track", "delayed", "in testing", "on hold", "approved", "cancelled", "completed"}
	rQuery.Filters.AccessSafelist = []string{"private", "public"}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	projects, metadata, err := h.service.GetAllProjects(ctx, rQuery.Name, rQuery.Owner, rQuery.Status, rQuery.CreatedBy, rQuery.ModifiedBy, rQuery.Access, rQuery.Filters, v)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFailedValidation):
			h.failedValidationResponse(w, r, err)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}
	err = h.encodeJSON(w, http.StatusOK, envelop{"projects": projects, "metadata": metadata}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

func (h *Handler) updateProject(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Owner       *string `json:"owner"`
		Status      *string `json:"status"`
		StartDate   *string `json:"start_date"`
		EndDate     *string `json:"end_date"`
		CompletedOn *string `json:"completed_on"`
		Access      *string `json:"access"`
	}
	err := h.decodeJSON(w, r, &requestBody)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	id, err := h.readIDParam(r, "project_id")
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	project, err := h.service.UpdateProject(ctx, id, requestBody.Name, requestBody.Description, requestBody.Owner, requestBody.Status, requestBody.StartDate, requestBody.EndDate, requestBody.CompletedOn, requestBody.Access)
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
	err = h.encodeJSON(w, http.StatusOK, envelop{"project": project}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

func (h *Handler) deleteProject(w http.ResponseWriter, r *http.Request) {
	id, err := h.readIDParam(r, "project_id")
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	err = h.service.DeleteProject(ctx, id)
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
	err = h.encodeJSON(w, http.StatusOK, envelop{"message": "project successfully deleted"}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}
