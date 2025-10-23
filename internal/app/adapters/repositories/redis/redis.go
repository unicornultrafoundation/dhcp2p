package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/infrastructure/config"
	"go.uber.org/fx"
)

func NewRedisClient(lc fx.Lifecycle, cfg *config.AppConfig) (*redis.Client, error) {
	redisURL := cfg.RedisURL
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL environment variable not set")
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:         redisURL,
		Password:     cfg.RedisPassword,
		MaxRetries:   cfg.RedisMaxRetries,
		PoolSize:     cfg.RedisPoolSize,
		MinIdleConns: cfg.RedisMinIdleConns,
		DialTimeout:  time.Duration(cfg.RedisDialTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.RedisReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.RedisWriteTimeout) * time.Second,
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := redisClient.Ping(ctx).Err(); err != nil {
				return fmt.Errorf("failed to ping redis: %w", err)
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return redisClient.Close()
		},
	})

	return redisClient, nil
}
