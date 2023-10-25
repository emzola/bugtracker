package http

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (h *Handler) Routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(h.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(h.methodNotAllowedResponse)
	// router.HandlerFunc(http.MethodGet, "/v1/projects", h.getAllProjects)
	router.HandlerFunc(http.MethodPost, "/v1/projects", h.createProject)
	router.HandlerFunc(http.MethodGet, "/v1/projects/:projectId", h.getProject)
	router.HandlerFunc(http.MethodPatch, "/v1/projects/:projectId", h.updateProject)
	// router.HandlerFunc(http.MethodDelete, "/v1/projects/:id", h.deleteProject)
	return router
}
