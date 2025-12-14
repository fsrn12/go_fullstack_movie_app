package store

import (
	"context"
	"fmt"
	"time"

	"multipass/pkg/logging"
	"multipass/pkg/provider"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	DBMaxOpenConns = 10
	DBMaxIdleConns = 5
)

func Open(connStr string, logger logging.Logger) (*pgxpool.Pool, error) {
	if logger == nil {
		logger, _ = provider.GetLogger()
	}

	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		logger.Errorf("failed to parse database config: %w", err)
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	cfg.MaxConns = 10
	// db.SetMaxIdleConns(DBMaxIdleConns)
	// db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbPool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		logger.Errorf("failed to connect to database: %w", err)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := dbPool.Ping(ctx); err != nil {
		logger.Errorf("failed to ping database %w", err)
		return nil, fmt.Errorf("failed to ping database %w", err)
	}

	return dbPool, nil
}
