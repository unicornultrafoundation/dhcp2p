package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/errors"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/models"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/ports"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/infrastructure/config"
)

type LeaseCache struct {
	client    *redis.Client
	leaseTTL  time.Duration
	keyPrefix string
}

var _ ports.LeaseCache = &LeaseCache{}

func NewLeaseCache(client *redis.Client, cfg *config.AppConfig) *LeaseCache {
	return &LeaseCache{
		client:    client,
		leaseTTL:  time.Duration(cfg.LeaseTTL) * time.Minute,
		keyPrefix: "lease:",
	}
}

func (c *LeaseCache) GetLeaseByPeerID(ctx context.Context, peerID string) (*models.Lease, error) {
	key := c.keyPrefix + "peer:" + peerID
	return c.getLease(ctx, key)
}

func (c *LeaseCache) GetLeaseByTokenID(ctx context.Context, tokenID int64) (*models.Lease, error) {
	key := c.keyPrefix + "token:" + fmt.Sprintf("%d", tokenID)
	return c.getLease(ctx, key)
}

func (c *LeaseCache) getLease(ctx context.Context, key string) (*models.Lease, error) {
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.ErrLeaseNotFound
		}
		return nil, err
	}

	var lease models.Lease
	if err := json.Unmarshal([]byte(data), &lease); err != nil {
		return nil, err
	}

	return &lease, nil
}

func (c *LeaseCache) SetLease(ctx context.Context, lease *models.Lease) error {
	data, err := json.Marshal(lease)
	if err != nil {
		return err
	}

	// Use TTL from lease object (calculated by database)
	ttl := time.Duration(lease.Ttl) * time.Second
	if ttl <= 0 {
		// Do not cache already expired leases
		return nil
	}

	// Set both peer and token keys
	peerKey := c.keyPrefix + "peer:" + lease.PeerID
	tokenKey := c.keyPrefix + "token:" + fmt.Sprintf("%d", lease.TokenID)

	pipe := c.client.Pipeline()
	pipe.Set(ctx, peerKey, data, ttl)
	pipe.Set(ctx, tokenKey, data, ttl)

	_, err = pipe.Exec(ctx)
	return err
}

func (c *LeaseCache) DeleteLease(ctx context.Context, peerID string, tokenID int64) error {
	peerKey := c.keyPrefix + "peer:" + peerID
	tokenKey := c.keyPrefix + "token:" + fmt.Sprintf("%d", tokenID)

	pipe := c.client.Pipeline()
	pipe.Del(ctx, peerKey)
	pipe.Del(ctx, tokenKey)

	_, err := pipe.Exec(ctx)
	return err
}
