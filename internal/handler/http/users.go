package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/emzola/bugtracker/internal/model"
	"github.com/emzola/bugtracker/internal/service"
)

type userService interface {
	CreateUser(ctx context.Context, name, email, password, role, createdBy, modifiedBy string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserForToken(ctx context.Context, tokenScope, tokenPlaintext string) (*model.User, error)
	ActivateUser(ctx context.Context, user *model.User, modifiedBy string) error
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
	creator := "ems"
	user, err := h.service.CreateUser(ctx, requestBody.Name, requestBody.Email, requestBody.Password, requestBody.Role, creator, creator)
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
	modifiedby := "ems"
	err = h.service.ActivateUser(ctx, user, modifiedby)
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
