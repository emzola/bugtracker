package http

import "github.com/emzola/issuetracker/config"

type ServiceLayer interface {
	projectService
	userService
	tokenService
	issueService
	issuesReportService
}

// Handler defines the app's HTTP handler.
type Handler struct {
	service ServiceLayer
	Config  config.AppConfiguration
}

// New creates a new HTTP handler.
func New(service ServiceLayer, cfg config.AppConfiguration) *Handler {
	return &Handler{service, cfg}
}
