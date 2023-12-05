package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/emzola/issuetracker/internal/model"
	"github.com/emzola/issuetracker/internal/service"
	"github.com/emzola/issuetracker/pkg/rbac"
	"github.com/pascaldekloe/jwt"
	"golang.org/x/time/rate"
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
		if claims.Issuer != "github.com/emzola/issuetracker" {
			h.invalidAuthenticationTokenResponse(w, r)
			return
		}
		// Check that our application is in the expected audiences for the JWT.
		if !claims.AcceptAudience("github.com/emzola/issuetracker") {
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
		// Check RBAC permission for authenticated user.
		rbacAuthorizer := rbac.New(h.roles)
		asset := strings.Split(strings.Trim(r.URL.Path, "/"), "/")[1]
		action := rbacAuthorizer.ActionFromMethod(r.Method)
		if !rbacAuthorizer.HasPermission(user, action, asset) {
			h.notPermittedResponse(w, r)
			return
		}
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

// recoverPanic recovers from app-wide panics.
func (h *Handler) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				h.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// rateLimit implements IP-based rate limiting.
func (h *Handler) rateLimit(next http.Handler) http.Handler {
	// Define a client struct to hold rate limiter and last seen time.
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)
	// Launch a background goroutine which removes old entries from the clients maps
	// once every minute.
	go func() {
		for {
			time.Sleep(time.Minute)
			// Lock the mutex to prevent any rate limiter checks from happening while
			// the cleanup is taking place.
			mu.Lock()
			// Loop through all clients. If they haven't been seen within the last three
			// minutes, delete the corresponding entry from the map.
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			// Unlock the mutex when the cleanup is complete.
			mu.Unlock()
		}
	}()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.Config.Limiter.Enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				h.serverErrorResponse(w, r, err)
				return
			}
			mu.Lock()
			if _, exists := clients[ip]; !exists {
				// Create and add a new client struct to the map if it doesn't already exist.
				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(h.Config.Limiter.Rps), h.Config.Limiter.Burst)}
			}
			// Update the last seen time for the client.
			clients[ip].lastSeen = time.Now()
			// Call the Allow() method on the rate limiter for the current IP address. If
			// the request isn't allowed, unlock the mutex and send a 429 Too Many Requests.
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				h.rateLimitExceededResponse(w, r)
				return
			}
			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}

// enableCORS implements cross origin requests.
func (h *Handler) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")
		origin := r.Header.Get("Origin")
		if origin != "" {
			for i := range h.Config.Cors.TrustedOrigins {
				if origin == h.Config.Cors.TrustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
						w.WriteHeader(http.StatusOK)
						return
					}
					break
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
