package postgres

import (
	"context"
	"fmt"

	"github.com/duchuongnguyen/dhcp2p/internal/app/infrastructure/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

type DBPool struct {
	pool *pgxpool.Pool
}

func NewDBPool(lc fx.Lifecycle, cfg *config.AppConfig) (*pgxpool.Pool, error) {
	dbURL := cfg.DatabaseURL
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable not set")
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err = pool.Ping(ctx); err != nil {
				pool.Close()
				return fmt.Errorf("failed to ping database: %w", err)
			}
			return err
		},
		OnStop: func(ctx context.Context) error {
			pool.Close()
			return nil
		},
	})

	return pool, nil
}
