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
	CreateProject(ctx context.Context, name, description, startDate, targetEndDate, createdBy, modifiedBy string) (*model.Project, error)
	GetProject(ctx context.Context, id int64) (*model.Project, error)
	GetAllProjects(ctx context.Context, name, startDate, targetEndDate, actualEndDate, createdby string, filters model.Filters, v *validator.Validator) ([]*model.Project, model.Metadata, error)
	UpdateProject(ctx context.Context, id int64, name, description, startDate, targetEndDate, actualEnddate *string, modifiedBy string) (*model.Project, error)
	DeleteProject(ctx context.Context, id int64) error
}

func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		StartDate     string `json:"start_date"`
		TargetEndDate string `json:"target_end_date"`
	}
	err := h.decodeJSON(w, r, &requestBody)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	user := "ems"
	project, err := h.service.CreateProject(ctx, requestBody.Name, requestBody.Description, requestBody.StartDate, requestBody.TargetEndDate, user, user)
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
	var requestQuery struct {
		Name          string
		StartDate     string
		TargetEndDate string
		ActualEndDate string
		CreatedBy     string
		Filters       model.Filters
	}
	v := validator.New()
	qs := r.URL.Query()
	requestQuery.Name = h.readString(qs, "project_name", "")
	requestQuery.StartDate = h.readString(qs, "start_date", "")
	requestQuery.TargetEndDate = h.readString(qs, "target_end_date", "")
	requestQuery.ActualEndDate = h.readString(qs, "actual_end_date", "")
	requestQuery.CreatedBy = h.readString(qs, "created_by", "")
	requestQuery.Filters.Page = h.readInt(qs, "page", 1, v)
	requestQuery.Filters.PageSize = h.readInt(qs, "page_size", 20, v)
	requestQuery.Filters.Sort = h.readString(qs, "sort", "project_id")
	requestQuery.Filters.SortSafelist = []string{"project_id", "project_name", "start_date", "target_end_date", "actual_end_date", "created_by", "-project_id", "-project_name", "-start_date", "-target_end_date", "-actual_end_date", "-created_by"}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	projects, metadata, err := h.service.GetAllProjects(ctx, requestQuery.Name, requestQuery.StartDate, requestQuery.TargetEndDate, requestQuery.ActualEndDate, requestQuery.CreatedBy, requestQuery.Filters, v)
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
	id, err := h.readIDParam(r, "project_id")
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	var requestBody struct {
		Name          *string `json:"name"`
		Description   *string `json:"description"`
		StartDate     *string `json:"start_date"`
		TargetEndDate *string `json:"target_end_date"`
		ActualEndDate *string `json:"actual_end_date"`
	}
	err = h.decodeJSON(w, r, &requestBody)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	user := "Ems"
	project, err := h.service.UpdateProject(ctx, id, requestBody.Name, requestBody.Description, requestBody.StartDate, requestBody.TargetEndDate, requestBody.ActualEndDate, user)
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
