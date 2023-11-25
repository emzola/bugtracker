package http

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/emzola/bugtracker/internal/model"
	"github.com/emzola/bugtracker/internal/service"
	"github.com/pascaldekloe/jwt"
)

// authenticate handles user authentication.
func (h *Handler) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			r = h.contextSetUser(r, model.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			h.invalidAuthenticationTokenResponse(w, r)
			return
		}
		token := headerParts[1]
		// Parse JWT and extract claims.
		claims, err := jwt.HMACCheck([]byte(token), []byte(h.Config.Jwt.Secret))
		if err != nil {
			h.invalidAuthenticationTokenResponse(w, r)
			return
		}
		// Check if JWT is still valid at this moment in time.
		if !claims.Valid(time.Now()) {
			h.invalidAuthenticationTokenResponse(w, r)
			return
		}
		// Check that the issuer is our application.
		if claims.Issuer != "github.com/emzola/bug-tracker" {
			h.invalidAuthenticationTokenResponse(w, r)
			return
		}
		// Check that our application is in the expected audiences for the JWT.
		if !claims.AcceptAudience("github.com/emzola/bug-tracker") {
			h.invalidAuthenticationTokenResponse(w, r)
			return
		}
		// Extract userID from claims subject and convert it from string to int64.
		userID, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			h.serverErrorResponse(w, r, err)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		// Lookup the user record from the database.
		user, err := h.service.GetUserByID(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				return
			case errors.Is(err, service.ErrNotFound):
				h.invalidAuthenticationTokenResponse(w, r)
			default:
				h.serverErrorResponse(w, r, err)
			}
			return
		}
		// Add the user record to the request context and continue as normal.
		r = h.contextSetUser(r, user)
		next.ServeHTTP(w, r)
	})
}

// requireAuthenticatedUser checks that a user is not anonymous.
func (h *Handler) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := h.contextGetUser(r)
		if user.IsAnonymous() {
			h.authenticationRequiredResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requireActivatedUser checks that a user is both authenticated and activated.
func (h *Handler) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := h.contextGetUser(r)
		if !user.Activated {
			h.inactiveAccountResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
	return h.requireAuthenticatedUser(fn)
}
