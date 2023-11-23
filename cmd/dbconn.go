package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/emzola/bugtracker/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// dbConn configures a database connection pool.
func dbConn(cfg config.AppConfiguration) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.Database.Dsn)
	if err != nil {
		return nil, err
	}
	duration, err := time.ParseDuration(cfg.Database.MaxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxIdleTime(duration)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
