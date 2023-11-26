package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/emzola/issuetracker/internal/model"
	"github.com/emzola/issuetracker/internal/service"
)

type tokenService interface {
	CreateActivationToken(ctx context.Context, user *model.User) error
	CreateAuthenticationToken(ctx context.Context, email, password string) ([]byte, error)
}

func (h *Handler) createActivationToken(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Email string `json:"email"`
	}
	err := h.decodeJSON(w, r, &requestBody)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	user, err := h.service.GetUserByEmail(ctx, requestBody.Email)
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
	err = h.service.CreateActivationToken(ctx, user)
	if err != nil {
		h.serverErrorResponse(w, r, err)
		return
	}
	err = h.encodeJSON(w, http.StatusOK, envelop{"message": "an email will be sent to you containing activation instructions"}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

func (h *Handler) createAuthenticationToken(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := h.decodeJSON(w, r, &requestBody)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	jwtBytes, err := h.service.CreateAuthenticationToken(ctx, requestBody.Email, requestBody.Password)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		case errors.Is(err, service.ErrFailedValidation):
			h.failedValidationResponse(w, r, err)
		case errors.Is(err, service.ErrInvalidCredentials):
			h.invalidCredentialsResponse(w, r)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}
	err = h.encodeJSON(w, http.StatusCreated, envelop{"authentication_token": string(jwtBytes)}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}
