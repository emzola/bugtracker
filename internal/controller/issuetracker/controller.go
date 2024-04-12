package issuetracker

import (
	"sync"

	"github.com/emzola/issuetracker/config"
	"go.uber.org/zap"
)

type issueTrackerRepository interface {
	projectRepository
	userRepository
	tokenRepository
	issueRepository
	issuesReportRepository
}

type Controller struct {
	repo   issueTrackerRepository
	Config config.App
	wg     *sync.WaitGroup
	Logger *zap.Logger
}

func New(repo issueTrackerRepository, cfg config.App, wg *sync.WaitGroup, logger *zap.Logger) *Controller {
	return &Controller{repo, cfg, wg, logger}
}
