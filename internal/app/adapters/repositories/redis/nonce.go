package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/errors"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/models"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/ports"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/infrastructure/config"
)

type NonceCache struct {
	client    *redis.Client
	nonceTTL  time.Duration
	keyPrefix string
}

var _ ports.NonceCache = &NonceCache{}

func NewNonceCache(client *redis.Client, cfg *config.AppConfig) *NonceCache {
	return &NonceCache{
		client:    client,
		nonceTTL:  time.Duration(cfg.NonceTTL) * time.Minute,
		keyPrefix: "nonce:",
	}
}

func (c *NonceCache) GetNonce(ctx context.Context, nonceID string) (*models.Nonce, error) {
	key := c.keyPrefix + nonceID
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.ErrNonceNotFound
		}
		return nil, err
	}

	var nonce models.Nonce
	if err := json.Unmarshal([]byte(data), &nonce); err != nil {
		return nil, err
	}

	return &nonce, nil
}

func (c *NonceCache) CreateNonce(ctx context.Context, nonce *models.Nonce) error {
	key := c.keyPrefix + nonce.ID
	data, err := json.Marshal(nonce)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, c.nonceTTL).Err()
}

func (c *NonceCache) DeleteNonce(ctx context.Context, nonceID string) error {
	key := c.keyPrefix + nonceID
	return c.client.Del(ctx, key).Err()
}
