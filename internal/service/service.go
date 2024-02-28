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

type Service struct {
	repo   RepositoryLayer
	Config config.AppConfiguration
	wg     *sync.WaitGroup
	Logger *zap.Logger
}

func New(repo RepositoryLayer, cfg config.AppConfiguration, wg *sync.WaitGroup, logger *zap.Logger) *Service {
	return &Service{repo, cfg, wg, logger}
}
