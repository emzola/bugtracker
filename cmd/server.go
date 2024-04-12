package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/emzola/issuetracker/config"
	"go.uber.org/zap"
)

func serve(handler http.Handler, cfg config.App, wg *sync.WaitGroup, logger *zap.Logger) error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	// Graceful shutdown.
	shutdownErr := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		logger.Info("shutting down server", zap.Any("properties", map[string]string{
			"signal": s.String(),
		}))
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownErr <- err
		}
		logger.Info("completing background tasks", zap.Any("properties", map[string]string{
			"addr": srv.Addr,
		}))
		wg.Wait()
		shutdownErr <- nil
	}()
	// Start server.
	logger.Info("starting server", zap.Any("properties", map[string]string{
		"addr": srv.Addr,
		"env":  cfg.Env,
	}))
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	err = <-shutdownErr
	if err != nil {
		return err
	}
	logger.Info("server stopped", zap.Any("properties", map[string]string{
		"addr": srv.Addr,
	}))
	return nil
}
