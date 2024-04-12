package http

import (
	"github.com/emzola/issuetracker/config"
	"github.com/emzola/issuetracker/internal/controller/issuetracker"
	"github.com/emzola/issuetracker/pkg/rbac"
)

type Handler struct {
	ctrl   *issuetracker.Controller
	Config config.App
	roles  rbac.Roles
}

func New(ctrl *issuetracker.Controller, cfg config.App, roles rbac.Roles) *Handler {
	return &Handler{ctrl, cfg, roles}
}
