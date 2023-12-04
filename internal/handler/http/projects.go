package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/emzola/issuetracker/internal/model"
	"github.com/emzola/issuetracker/internal/service"
	"github.com/emzola/issuetracker/pkg/validator"
)

type projectService interface {
	CreateProject(ctx context.Context, name, description string, assignedTo *int64, startDate, targetEndDate, createdBy, modifiedBy string) (*model.Project, error)
	GetProject(ctx context.Context, id int64) (*model.Project, error)
	GetAllProjects(ctx context.Context, name string, assignedTo int64, startDate, targetEndDate, actualEndDate, createdBy string, filters model.Filters, v *validator.Validator) ([]*model.Project, model.Metadata, error)
	UpdateProject(ctx context.Context, id int64, name, description *string, assignedTo *int64, startDate, targetEndDate, actualEnddate *string, user *model.User) (*model.Project, error)
	DeleteProject(ctx context.Context, id int64) error
	GetProjectUsers(ctx context.Context, projectID int64, role string, filters model.Filters, v *validator.Validator) ([]*model.User, model.Metadata, error)
	GetProjectUser(ctx context.Context, projectID, userID int64) (*model.User, error)
}

func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		AssignedTo    *int64 `json:"assigned_to"`
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
	userFromContext := h.contextGetUser(r)
	project, err := h.service.CreateProject(ctx, requestBody.Name, requestBody.Description, requestBody.AssignedTo, requestBody.StartDate, requestBody.TargetEndDate, userFromContext.Name, userFromContext.Name)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		case errors.Is(err, service.ErrNotFound):
			h.notFoundResponse(w, r)
		case errors.Is(err, service.ErrInvalidRole):
			h.invalidRoleResponse(w, r)
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
	projectID, err := h.readIDParam(r, "project_id")
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	project, err := h.service.GetProject(ctx, projectID)
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
		AssignedTo    int64
		StartDate     string
		TargetEndDate string
		ActualEndDate string
		CreatedBy     string
		Filters       model.Filters
	}
	v := validator.New()
	qs := r.URL.Query()
	requestQuery.Name = h.readString(qs, "name", "")
	requestQuery.AssignedTo = int64(h.readInt(qs, "assigned_to", 0, v))
	requestQuery.StartDate = h.readString(qs, "start_date", "")
	requestQuery.TargetEndDate = h.readString(qs, "target_end_date", "")
	requestQuery.ActualEndDate = h.readString(qs, "actual_end_date", "")
	requestQuery.CreatedBy = h.readString(qs, "created_by", "")
	requestQuery.Filters.Page = h.readInt(qs, "page", 1, v)
	requestQuery.Filters.PageSize = h.readInt(qs, "page_size", 20, v)
	requestQuery.Filters.Sort = h.readString(qs, "sort", "id")
	requestQuery.Filters.SortSafelist = []string{"id", "name", "assigned_to", "start_date", "target_end_date", "actual_end_date", "created_by", "-id", "-name", "-assigned_to", "-start_date", "-target_end_date", "-actual_end_date", "-created_by"}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	projects, metadata, err := h.service.GetAllProjects(ctx, requestQuery.Name, requestQuery.AssignedTo, requestQuery.StartDate, requestQuery.TargetEndDate, requestQuery.ActualEndDate, requestQuery.CreatedBy, requestQuery.Filters, v)
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
	err = h.encodeJSON(w, http.StatusOK, envelop{"projects": projects, "metadata": metadata}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

func (h *Handler) updateProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.readIDParam(r, "project_id")
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}
	var requestBody struct {
		Name          *string `json:"name"`
		Description   *string `json:"description"`
		AssignedTo    *int64  `json:"assigned_to"`
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
	userFromContext := h.contextGetUser(r)
	project, err := h.service.UpdateProject(ctx, projectID, requestBody.Name, requestBody.Description, requestBody.AssignedTo, requestBody.StartDate, requestBody.TargetEndDate, requestBody.ActualEndDate, userFromContext)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		case errors.Is(err, service.ErrNotPermitted):
			h.notPermittedResponse(w, r)
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
	projectID, err := h.readIDParam(r, "project_id")
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	err = h.service.DeleteProject(ctx, projectID)
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

func (h *Handler) getProjectUsers(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.readIDParam(r, "project_id")
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}
	var requestQuery struct {
		Role    string
		Filters model.Filters
	}
	v := validator.New()
	qs := r.URL.Query()
	requestQuery.Role = h.readString(qs, "role", "")
	requestQuery.Filters.Page = h.readInt(qs, "page", 1, v)
	requestQuery.Filters.PageSize = h.readInt(qs, "page_size", 20, v)
	requestQuery.Filters.Sort = h.readString(qs, "sort", "id")
	requestQuery.Filters.SortSafelist = []string{"id", "-id"}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	users, metadata, err := h.service.GetProjectUsers(ctx, projectID, requestQuery.Role, requestQuery.Filters, v)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}
	err = h.encodeJSON(w, http.StatusOK, envelop{"users": users, "metadata": metadata}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}
