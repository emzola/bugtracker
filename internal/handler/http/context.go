package http

import (
	"context"
	"net/http"

	"github.com/emzola/bugtracker/internal/model"
)

type contextKey string

const userContextKey = contextKey("user")

func (h *Handler) contextSetUser(r *http.Request, user *model.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (h *Handler) contextGetUser(r *http.Request) *model.User {
	user, ok := r.Context().Value(userContextKey).(*model.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
