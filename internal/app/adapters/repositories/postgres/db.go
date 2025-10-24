package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/infrastructure/config"
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

	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure pool settings from config
	if cfg.DBMaxConns > 0 {
		poolConfig.MaxConns = int32(cfg.DBMaxConns)
	}
	if cfg.DBMinConns > 0 {
		poolConfig.MinConns = int32(cfg.DBMinConns)
	}
	if cfg.DBMaxConnLifetime > 0 {
		poolConfig.MaxConnLifetime = time.Duration(cfg.DBMaxConnLifetime) * time.Minute
	}
	if cfg.DBMaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime = time.Duration(cfg.DBMaxConnIdleTime) * time.Minute
	}
	if cfg.DBHealthCheckPeriod > 0 {
		poolConfig.HealthCheckPeriod = time.Duration(cfg.DBHealthCheckPeriod) * time.Second
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
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
