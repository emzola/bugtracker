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
	err = h.encodeJSON(w, http.StatusCreated, envelop{"user": user}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}
