package config

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func DbConn(app App) (*sql.DB, error) {
	db, err := sql.Open("pgx", app.Database.Dsn)
	if err != nil {
		return nil, err
	}
	duration, err := time.ParseDuration(app.Database.MaxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(app.Database.MaxOpenConns)
	db.SetMaxIdleConns(app.Database.MaxIdleConns)
	db.SetConnMaxIdleTime(duration)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
