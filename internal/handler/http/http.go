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

type Handler struct {
	service ServiceLayer
	Config  config.AppConfiguration
	roles   rbac.Roles
}

func New(service ServiceLayer, cfg config.AppConfiguration, roles rbac.Roles) *Handler {
	return &Handler{service, cfg, roles}
}
