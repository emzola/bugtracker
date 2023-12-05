package service

import (
	"sync"

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
	wg     *sync.WaitGroup
	Logger *zap.Logger
}

// New creates a project service controller.
func New(repo RepositoryLayer, cfg config.AppConfiguration, wg *sync.WaitGroup, logger *zap.Logger) *Service {
	return &Service{repo, cfg, wg, logger}
}
