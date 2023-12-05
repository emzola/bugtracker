package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/emzola/issuetracker/config"
	httpHandler "github.com/emzola/issuetracker/internal/handler/http"
	"github.com/emzola/issuetracker/internal/repository/postgres"
	"github.com/emzola/issuetracker/internal/service"
	"github.com/emzola/issuetracker/pkg/rbac"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	// Load roles.
	roles, err := rbac.LoadRoles("./roles.json")
	if err != nil {
		logger.Fatal("Failed to load roles", zap.Error(err))
	}
	var cfg config.AppConfiguration
	// Read server settings from command-line flags into the config struct.
	flag.IntVar(&cfg.Port, "port", 8080, "API server port")
	flag.StringVar(&cfg.Env, "env", "development", "Environment(development|staging|production)")
	// Read database connection pool settings from command-line flags into the config struct.
	flag.StringVar(&cfg.Database.Dsn, "db-dsn", os.Getenv("DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.Database.MaxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.Database.MaxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.Database.MaxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection")
	// Read SMTP settings from command-line flags into the config struct.
	flag.StringVar(&cfg.Smtp.Host, "smtp-host", os.Getenv("SMTP_HOST"), "SMTP host")
	flag.IntVar(&cfg.Smtp.Port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&cfg.Smtp.Username, "smtp-username", os.Getenv("SMTP_USERNAME"), "SMTP username")
	flag.StringVar(&cfg.Smtp.Password, "smtp-password", os.Getenv("SMTP_PASSWORD"), "SMTP password")
	flag.StringVar(&cfg.Smtp.Sender, "smtp-sender", "Issue Tracker <no-reply@github.com/emzola/issuetracker>", "SMTP sender")
	// Read JWT signing secret from command-line flags into the config struct.
	flag.StringVar(&cfg.Jwt.Secret, "jwt-secret", "", "JWT secret")
	// Read Rate Limiter settings from command-line flags into the config struct.
	flag.Float64Var(&cfg.Limiter.Rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.Limiter.Burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.Limiter.Enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.Parse()
	// Establish database connection pool.
	logger.Info("Starting the application", zap.Int("port", cfg.Port))
	db, err := dbConn(cfg)
	if err != nil {
		logger.Fatal("Failed to establish database connection pool", zap.Error(err))
	}
	// Instantiate app layers.
	repo := postgres.New(db)
	service := service.New(repo, cfg, logger)
	handler := httpHandler.New(service, cfg, roles)
	// Start server.
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), handler.Routes()); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
