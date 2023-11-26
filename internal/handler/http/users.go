package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/emzola/issuetracker/internal/model"
	"github.com/emzola/issuetracker/internal/service"
)

type userService interface {
	CreateUser(ctx context.Context, name, email, password, role, createdBy, modifiedBy string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
	GetUserForToken(ctx context.Context, tokenScope, tokenPlaintext string) (*model.User, error)
	ActivateUser(ctx context.Context, user *model.User, modifiedBy string) error
	UpdateUser(ctx context.Context, id int64, name, email, role *string, modifiedby string) (*model.User, error)
	DeleteUser(ctx context.Context, id int64) error
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	err := h.decodeJSON(w, r, &requestBody)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	userFromContext := h.contextGetUser(r)
	user, err := h.service.CreateUser(ctx, requestBody.Name, requestBody.Email, requestBody.Password, requestBody.Role, userFromContext.Name, userFromContext.Name)
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

func (h *Handler) activateUser(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Token string `json:"token"`
	}
	err := h.decodeJSON(w, r, &requestBody)
	if err != nil {
		h.badRequestResponse(w, r, err)
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	user, err := h.service.GetUserForToken(ctx, model.ScopeActivation, requestBody.Token)
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

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := h.readIDParam(r, "user_id")
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}
	var requestBody struct {
		Name  *string `json:"name"`
		Email *string `json:"email"`
		Role  *string `json:"role"`
	}
	err = h.decodeJSON(w, r, &requestBody)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	userFromContext := h.contextGetUser(r)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	user, err := h.service.UpdateUser(ctx, userID, requestBody.Name, requestBody.Email, requestBody.Role, userFromContext.Name)
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
		}
		return
	}
	err = h.encodeJSON(w, http.StatusOK, envelop{"user": user}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

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
