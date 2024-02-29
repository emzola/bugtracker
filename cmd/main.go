package main

import (
	"flag"
	"os"
	"strings"
	"sync"

	"github.com/emzola/issuetracker/config"
	_ "github.com/emzola/issuetracker/docs"
	httpHandler "github.com/emzola/issuetracker/internal/handler/http"
	"github.com/emzola/issuetracker/internal/repository/postgres"
	"github.com/emzola/issuetracker/internal/service"
	"github.com/emzola/issuetracker/pkg/rbac"

	"go.uber.org/zap"
)

// @title  Issue Tracker API
// @version 1.0.0
// @description This is an API service for an issue tracker.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email emma.idika@yahoo.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host https://issuetracker-api-dev.fl0.io
// @BasePath /
func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	// Load roles.
	roles, err := rbac.LoadRoles("./roles.json")
	if err != nil {
		logger.Fatal("failed to load roles", zap.Error(err))
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
	flag.Float64Var(&cfg.Limiter.Rps, "limiter-rps", 4, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.Limiter.Burst, "limiter-burst", 8, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.Limiter.Enabled, "limiter-enabled", true, "Enable rate limiter")
	// Read CORS configuration from command-line flags into the config struct.
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(s string) error {
		cfg.Cors.TrustedOrigins = strings.Fields(s)
		return nil
	})
	flag.Parse()
	// Establish database connection pool.
	db, err := dbConn(cfg)
	if err != nil {
		logger.Fatal("failed to establish database connection pool", zap.Error(err))
	}
	logger.Info("database connection pool established")
	var wg sync.WaitGroup
	// Instantiate app layers.
	repo := postgres.New(db)
	service := service.New(repo, cfg, &wg, logger)
	handler := httpHandler.New(service, cfg, roles)
	// Start server.
	err = serve(handler.Routes(), cfg, &wg, logger)
	if err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
}
