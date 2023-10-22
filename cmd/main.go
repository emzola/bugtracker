package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	httpHandler "github.com/emzola/bugtracker/internal/handler/http"
	"github.com/emzola/bugtracker/internal/repository/postgresql"
	"github.com/emzola/bugtracker/internal/service"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	f, err := os.Open(filepath.FromSlash("../configs/base.yaml"))
	if err != nil {
		logger.Fatal("Failed to open configuration", zap.Error(err))
	}
	var cfg config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		logger.Fatal("Failed to parse configuration", zap.Error(err))
	}
	port := cfg.API.Port
	logger.Info("Starting the application", zap.Int("port", port))
	repo, err := postgresql.New()
	if err != nil {
		logger.Fatal("Failed to establish database connection pool", zap.Error(err))
	}
	service := service.New(repo)
	handler := httpHandler.New(service)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), handler.Routes()); err != nil {
		panic(err)
	}
}
