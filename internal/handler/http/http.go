package http

import (
	"github.com/emzola/issuetracker/config"
	"github.com/emzola/issuetracker/pkg/rbac"
)

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
	roles   rbac.Roles
}

// New creates a new HTTP handler.
func New(service ServiceLayer, cfg config.AppConfiguration, roles rbac.Roles) *Handler {
	return &Handler{service, cfg, roles}
}
