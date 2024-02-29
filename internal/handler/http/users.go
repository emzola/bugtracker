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

type userService interface {
	CreateUser(ctx context.Context, name, email, password, role, createdBy, modifiedBy string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
	GetAllUsers(ctx context.Context, name, email, role string, filters model.Filters, v *validator.Validator) ([]*model.User, model.Metadata, error)
	GetUserForToken(ctx context.Context, tokenScope, tokenPlaintext string) (*model.User, error)
	ActivateUser(ctx context.Context, user *model.User, modifiedBy string) error
	UpdateUser(ctx context.Context, id int64, name, email, role *string, modifiedby string) (*model.User, error)
	DeleteUser(ctx context.Context, id int64) error
	AssignUserToProject(ctx context.Context, userID, projectID int64) error
	GetAllProjectsForUser(ctx context.Context, userID int64, filters model.Filters, v *validator.Validator) ([]*model.Project, model.Metadata, error)
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with the request payload
// @Tags users
// @Accept  json
// @Produce json
// @Param token header string true "Bearer token"
// @Param payload body createUserPayload true "Request payload"
// @Success 202 {object} model.User
// @Failure 400
// @Failure 422
// @Failure 500
// @Router /v1/users [post]
func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	var requestPayload createUserPayload
	err := h.decodeJSON(w, r, &requestPayload)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	userFromContext := h.contextGetUser(r)
	user, err := h.service.CreateUser(ctx, requestPayload.Name, requestPayload.Email, requestPayload.Password, requestPayload.Role, userFromContext.Name, userFromContext.Name)
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
	err = h.encodeJSON(w, http.StatusAccepted, envelop{"user": user}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

// ActivateUser godoc
// @Summary Activate a new user
// @Description Activate a new user with the request payload
// @Tags users
// @Accept  json
// @Produce json
// @Param token header string true "Bearer token"
// @Param payload body activateUserPayload true "Request payload"
// @Success 200 {object} model.User
// @Failure 400
// @Failure 409
// @Failure 422
// @Failure 500
// @Router /v1/users/activated [put]
func (h *Handler) activateUser(w http.ResponseWriter, r *http.Request) {
	var requestPayload activateUserPayload
	err := h.decodeJSON(w, r, &requestPayload)
	if err != nil {
		h.badRequestResponse(w, r, err)
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	user, err := h.service.GetUserForToken(ctx, model.ScopeActivation, requestPayload.Token)
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
	userFromContext := h.contextGetUser(r)
	err = h.service.ActivateUser(ctx, user, userFromContext.Name)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		case errors.Is(err, service.ErrEditConflict):
			h.editConflictResponse(w, r)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}
	err = h.encodeJSON(w, http.StatusOK, envelop{"user": user}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

// GetUser godoc
// @Summary Get user by ID
// @Description This endpoint gets a user by ID
// @Tags users
// @Produce json
// @Param token header string true "Bearer token"
// @Param user_id path string true "ID of user to get"
// @Success 200 {object} model.User
// @Failure 404
// @Failure 500
// @Router /v1/users/{user_id} [get]
func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	userID, err := h.readIDParam(r, "user_id")
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	user, err := h.service.GetUserByID(ctx, userID)
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
	err = h.encodeJSON(w, http.StatusOK, envelop{"user": user}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

// GetAllUsers godoc
// @Summary Get all users
// @Description This endpoint gets all users
// @Tags users
// @Produce json
// @Param token header string true "Bearer token"
// @Param name query string false "Query string param for name"
// @Param email query string false "Query string param for email"
// @Param role query string false "Query string param for role"
// @Param page query string false "Query string param for pagination (min 1)"
// @Param page_size query string false "Query string param for pagination (max 100)"
// @Param sort query string false "Sort by asc or desc order. Asc: id, name, email, created_on, modified_on | Desc: -id, -name, -email, -created_on, -modified_on"
// @Success 200 {array} model.User
// @Failure 422
// @Failure 500
// @Router /v1/users [get]
func (h *Handler) getAllUsers(w http.ResponseWriter, r *http.Request) {
	var requestQuery struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Role    string `json:"role"`
		Filters model.Filters
	}
	v := validator.New()
	qs := r.URL.Query()
	requestQuery.Name = h.readString(qs, "name", "")
	requestQuery.Email = h.readString(qs, "email", "")
	requestQuery.Role = h.readString(qs, "role", "")
	requestQuery.Filters.Page = h.readInt(qs, "page", 1, v)
	requestQuery.Filters.PageSize = h.readInt(qs, "page_size", 20, v)
	requestQuery.Filters.Sort = h.readString(qs, "sort", "id")
	requestQuery.Filters.SortSafelist = []string{"id", "name", "email", "created_on", "modified_on", "-id", "-name", "-email", "-created_on", "-modified_on"}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	users, metadata, err := h.service.GetAllUsers(ctx, requestQuery.Name, requestQuery.Email, requestQuery.Role, requestQuery.Filters, v)
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
	err = h.encodeJSON(w, http.StatusOK, envelop{"users": users, "metadata": metadata}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

// UpdateUser godoc
// @Summary Update a user
// @Description This endpoint updates a user
// @Tags users
// @Accept  json
// @Produce json
// @Param token header string true "Bearer token"
// @Param payload body updateUserPayload true "Request payload"
// @Param user_id path string true "ID of user to update"
// @Success 200 {object} model.User
// @Failure 400
// @Failure 404
// @Failure 409
// @Failure 422
// @Failure 500
// @Router /v1/users/{user_id} [patch]
func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	var requestPayload updateUserPayload
	userID, err := h.readIDParam(r, "user_id")
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
	user, err := h.service.UpdateUser(ctx, userID, requestPayload.Name, requestPayload.Email, requestPayload.Role, userFromContext.Name)
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
	err = h.encodeJSON(w, http.StatusOK, envelop{"user": user}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

// DeleteUser godoc
// @Summary Delete a user
// @Description This endpoint deletes a user
// @Tags users
// @Produce json
// @Param token header string true "Bearer token"
// @Param user_id path string true "ID of user to delete"
// @Success 200
// @Failure 404
// @Failure 500
// @Router /v1/users/{user_id} [delete]
func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := h.readIDParam(r, "user_id")
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	err = h.service.DeleteUser(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			h.notFoundResponse(w, r)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}
	err = h.encodeJSON(w, http.StatusOK, envelop{"message": "user successfully deleted"}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

// AssignUserToProject godoc
// @Summary Assign a user to a project
// @Description Assign a user to a project with the request payload
// @Tags users
// @Accept  json
// @Produce json
// @Param token header string true "Bearer token"
// @Param payload body assignUserToProjectPayload true "Request payload"
// @Success 200
// @Failure 400
// @Failure 403
// @Failure 404
// @Failure 422
// @Failure 500
// @Router /v1/users/{user_id}/projects [post]
func (h *Handler) assignUserToProject(w http.ResponseWriter, r *http.Request) {
	var requestPayload assignUserToProjectPayload
	userID, err := h.readIDParam(r, "user_id")
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
	err = h.service.AssignUserToProject(ctx, userID, requestPayload.ProjectID)
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
	err = h.encodeJSON(w, http.StatusOK, envelop{"message": "user successfully assigned to project"}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

// GetAllProjectsForUser godoc
// @Summary Get all projects for user
// @Description This endpoint gets all projects for a user
// @Tags users
// @Produce json
// @Param token header string true "Bearer token"
// @Param page query string false "Query string param for pagination (min 1)"
// @Param page_size query string false "Query string param for pagination (max 100)"
// @Param sort query string false "Sort by asc or desc order. Asc: id | Desc: -id"
// @Success 200 {array} model.User
// @Failure 422
// @Failure 500
// @Router /v1/users/{user_id}/projects [get]
func (h *Handler) getAllProjectsForUser(w http.ResponseWriter, r *http.Request) {
	var queryParams struct {
		Filters model.Filters
	}
	userID, err := h.readIDParam(r, "user_id")
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}
	v := validator.New()
	qs := r.URL.Query()
	queryParams.Filters.Page = h.readInt(qs, "page", 1, v)
	queryParams.Filters.PageSize = h.readInt(qs, "page_size", 20, v)
	queryParams.Filters.Sort = h.readString(qs, "sort", "id")
	queryParams.Filters.SortSafelist = []string{"id", "-id"}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	projects, metadata, err := h.service.GetAllProjectsForUser(ctx, userID, queryParams.Filters, v)
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
