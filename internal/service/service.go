package service

import (
	"github.com/emzola/issuetracker/config"
	"go.uber.org/zap"
)

type RepositoryLayer interface {
	projectRepository
	userRepository
	tokenRepository
	issueRepository
	issuesReportRepository
}

// Controller defines a new project service controller.
type Service struct {
	repo   RepositoryLayer
	Config config.AppConfiguration
	Logger *zap.Logger
}

// New creates a project service controller.
func New(repo RepositoryLayer, cfg config.AppConfiguration, logger *zap.Logger) *Service {
	return &Service{repo, cfg, logger}
}
