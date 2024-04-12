package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/emzola/issuetracker/internal/controller/issuetracker"

	"github.com/emzola/issuetracker/pkg/model"
	"github.com/emzola/issuetracker/pkg/validator"
)

// CreateProject godoc
// @Summary Create a new project
// @Description Create a new project with the request payload
// @Tags projects
// @Accept  json
// @Produce json
// @Param token header string true "Bearer token"
// @Param payload body createProjectPayload true "Request payload"
// @Success 201 {object} model.Project
// @Failure 403
// @Failure 404
// @Failure 422
// @Failure 500
// @Router /v1/projects [post]
func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		AssignedTo    *int64 `json:"assigned_to"`
		StartDate     string `json:"start_date"`
		TargetEndDate string `json:"target_end_date"`
	}
	err := h.decodeJSON(w, r, &requestPayload)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	userFromContext := h.contextGetUser(r)
	project, err := h.ctrl.CreateProject(ctx, requestPayload.Name, requestPayload.Description, requestPayload.AssignedTo, requestPayload.StartDate, requestPayload.TargetEndDate, userFromContext.Name, userFromContext.Name)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		case errors.Is(err, issuetracker.ErrNotFound):
			h.notFoundResponse(w, r)
		case errors.Is(err, issuetracker.ErrInvalidRole):
			h.invalidRoleResponse(w, r)
		case errors.Is(err, issuetracker.ErrFailedValidation):
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

// GetProject godoc
// @Summary Get project by ID
// @Description This endpoint gets a project by ID
// @Tags projects
// @Produce json
// @Param token header string true "Bearer token"
// @Param project_id path string true "ID of project to get"
// @Success 200 {object} model.Project
// @Failure 404
// @Failure 500
// @Router /v1/projects/{project_id} [get]
func (h *Handler) getProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.readIDParam(r, "project_id")
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	project, err := h.ctrl.GetProject(ctx, projectID)
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
	err = h.encodeJSON(w, http.StatusOK, envelop{"project": project}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

// GetAllProjects godoc
// @Summary Get all projects
// @Description This endpoint gets all projects
// @Tags projects
// @Produce json
// @Param token header string true "Bearer token"
// @Param name query string false "Query string param for name"
// @Param assigned_to query string false "Query string param for assigned_to"
// @Param start_date query string false "Query string param for start_Date"
// @Param target_end_date query string false "Query string param for target_end_date"
// @Param actual_end_date query string false "Query string param for actual_end_date"
// @Param created_by query string false "Query string param for created_by"
// @Param page query string false "Query string param for pagination (min 1)"
// @Param page_size query string false "Query string param for pagination (max 100)"
// @Param sort query string false "Sort by asc or desc order. Asc: id, name, assigned_to, start_date, target_end_date, actual_end_date, created_by | Desc: -id, -name, -assigned_to, -start_date, -target_end_date, -actual_end_date, -created_by"
// @Success 200 {array} model.Project
// @Failure 422
// @Failure 500
// @Router /v1/projects [get]
func (h *Handler) getAllProjects(w http.ResponseWriter, r *http.Request) {
	var queryParams struct {
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
	queryParams.Name = h.readString(qs, "name", "")
	queryParams.AssignedTo = int64(h.readInt(qs, "assigned_to", 0, v))
	queryParams.StartDate = h.readString(qs, "start_date", "")
	queryParams.TargetEndDate = h.readString(qs, "target_end_date", "")
	queryParams.ActualEndDate = h.readString(qs, "actual_end_date", "")
	queryParams.CreatedBy = h.readString(qs, "created_by", "")
	queryParams.Filters.Page = h.readInt(qs, "page", 1, v)
	queryParams.Filters.PageSize = h.readInt(qs, "page_size", 20, v)
	queryParams.Filters.Sort = h.readString(qs, "sort", "id")
	queryParams.Filters.SortSafelist = []string{"id", "name", "assigned_to", "start_date", "target_end_date", "actual_end_date", "created_by", "-id", "-name", "-assigned_to", "-start_date", "-target_end_date", "-actual_end_date", "-created_by"}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	projects, metadata, err := h.ctrl.GetAllProjects(ctx, queryParams.Name, queryParams.AssignedTo, queryParams.StartDate, queryParams.TargetEndDate, queryParams.ActualEndDate, queryParams.CreatedBy, queryParams.Filters, v)
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
	err = h.encodeJSON(w, http.StatusOK, envelop{"projects": projects, "metadata": metadata}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

// UpdateProject godoc
// @Summary Update a project
// @Description This endpoint updates a project
// @Tags projects
// @Accept  json
// @Produce json
// @Param token header string true "Bearer token"
// @Param payload body updateProjectPayload true "Request payload"
// @Param project_id path string true "ID of project to update"
// @Success 200 {object} model.Project
// @Failure 400
// @Failure 403
// @Failure 404
// @Failure 409
// @Failure 422
// @Failure 500
// @Router /v1/projects/{project_id} [patch]
func (h *Handler) updateProject(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Name          *string `json:"name"`
		Description   *string `json:"description"`
		AssignedTo    *int64  `json:"assigned_to"`
		StartDate     *string `json:"start_date"`
		TargetEndDate *string `json:"target_end_date"`
		ActualEndDate *string `json:"actual_end_date"`
	}
	projectID, err := h.readIDParam(r, "project_id")
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}
	err = h.decodeJSON(w, r, &requestPayload)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	userFromContext := h.contextGetUser(r)
	project, err := h.ctrl.UpdateProject(ctx, projectID, requestPayload.Name, requestPayload.Description, requestPayload.AssignedTo, requestPayload.StartDate, requestPayload.TargetEndDate, requestPayload.ActualEndDate, userFromContext)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		case errors.Is(err, issuetracker.ErrNotPermitted):
			h.notPermittedResponse(w, r)
		case errors.Is(err, issuetracker.ErrNotFound):
			h.notFoundResponse(w, r)
		case errors.Is(err, issuetracker.ErrFailedValidation):
			h.failedValidationResponse(w, r, err)
		case errors.Is(err, issuetracker.ErrEditConflict):
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

// DeleteProject godoc
// @Summary Delete a project
// @Description This endpoint deletes a project
// @Tags projects
// @Produce json
// @Param token header string true "Bearer token"
// @Param project_id path string true "ID of project to delete"
// @Success 200
// @Failure 404
// @Failure 500
// @Router /v1/projects/{project_id} [delete]
func (h *Handler) deleteProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.readIDParam(r, "project_id")
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	err = h.ctrl.DeleteProject(ctx, projectID)
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
	err = h.encodeJSON(w, http.StatusOK, envelop{"message": "project successfully deleted"}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

// GetProjectUsers godoc
// @Summary Get project users
// @Description This endpoint gets all project users
// @Tags projects
// @Produce json
// @Param token header string true "Bearer token"
// @Param project_id path string true "ID of project to get users"
// @Param role query string false "Query string param for role"
// @Param page query string false "Query string param for pagination (min 1)"
// @Param page_size query string false "Query string param for pagination (max 100)"
// @Param sort query string false "Sort by asc or desc order. Asc: id | Desc: -id"
// @Success 200 {array} model.User
// @Failure 422
// @Failure 500
// @Router /v1/projects/{project_id}/users [get]
func (h *Handler) getProjectUsers(w http.ResponseWriter, r *http.Request) {
	var queryParams struct {
		Role    string
		Filters model.Filters
	}
	projectID, err := h.readIDParam(r, "project_id")
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}
	v := validator.New()
	qs := r.URL.Query()
	queryParams.Role = h.readString(qs, "role", "")
	queryParams.Filters.Page = h.readInt(qs, "page", 1, v)
	queryParams.Filters.PageSize = h.readInt(qs, "page_size", 20, v)
	queryParams.Filters.Sort = h.readString(qs, "sort", "id")
	queryParams.Filters.SortSafelist = []string{"id", "-id"}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	users, metadata, err := h.ctrl.GetProjectUsers(ctx, projectID, queryParams.Role, queryParams.Filters, v)
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
	err = h.encodeJSON(w, http.StatusOK, envelop{"users": users, "metadata": metadata}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}
