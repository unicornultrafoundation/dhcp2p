package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	redisclient "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/repositories/redis"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/errors"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/models"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/infrastructure/config"
	testconfig "github.com/unicornultrafoundation/dhcp2p/tests/config"
	"github.com/unicornultrafoundation/dhcp2p/tests/fixtures"
	"github.com/unicornultrafoundation/dhcp2p/tests/helpers"
)

func TestLeaseCache_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	builder := fixtures.NewTestBuilder()

	// Use container pool for better resource management
	pool := helpers.GetGlobalPool()
	redisContainer, connStr, err := pool.GetRedisContainer(ctx)
	require.NoError(t, err)
	defer pool.ReturnRedisContainer(redisContainer)

	// Create Redis client with retry logic
	redisClient := redisclient.NewClient(&redisclient.Options{
		Addr:         connStr,
		DialTimeout:  testconfig.TestTimeouts.DatabaseConnect,
		ReadTimeout:  testconfig.TestTimeouts.RequestTimeout,
		WriteTimeout: testconfig.TestTimeouts.RequestTimeout,
	})

	// Test connection
	err = redisClient.Ping(ctx).Err()
	require.NoError(t, err, "Failed to connect to Redis")

	// Create test config
	cfg := &config.AppConfig{
		LeaseTTL: testconfig.DefaultTTL,
	}

	// Create LeaseCache
	leaseCache := redis.NewLeaseCache(redisClient, cfg)

	t.Run("SetLease", func(t *testing.T) {
		lease := builder.NewLease().WithPeerID("peer123").Build()

		err := leaseCache.SetLease(ctx, lease)
		assert.NoError(t, err)
	})

	t.Run("GetLeaseByPeerID", func(t *testing.T) {
		lease := builder.NewLease().WithTokenID(67890).WithPeerID("peer456").Build()

		// Set lease first
		err := leaseCache.SetLease(ctx, lease)
		require.NoError(t, err)

		// Get lease by peer ID
		retrievedLease, err := leaseCache.GetLeaseByPeerID(ctx, lease.PeerID)
		assert.NoError(t, err)
		assert.Equal(t, lease.TokenID, retrievedLease.TokenID)
		assert.Equal(t, lease.PeerID, retrievedLease.PeerID)
		assert.Equal(t, lease.Ttl, retrievedLease.Ttl)
	})

	t.Run("GetLeaseByTokenID", func(t *testing.T) {
		lease := builder.NewLease().WithTokenID(11111).WithPeerID("peer789").Build()

		// Set lease first
		err := leaseCache.SetLease(ctx, lease)
		require.NoError(t, err)

		// Get lease by token ID
		retrievedLease, err := leaseCache.GetLeaseByTokenID(ctx, lease.TokenID)
		assert.NoError(t, err)
		assert.Equal(t, lease.TokenID, retrievedLease.TokenID)
		assert.Equal(t, lease.PeerID, retrievedLease.PeerID)
		assert.Equal(t, lease.Ttl, retrievedLease.Ttl)
	})

	t.Run("GetLeaseByPeerID_NotFound", func(t *testing.T) {
		lease, err := leaseCache.GetLeaseByPeerID(ctx, "non-existent-peer")
		assert.Error(t, err)
		assert.Equal(t, errors.ErrLeaseNotFound, err)
		assert.Nil(t, lease)
	})

	t.Run("GetLeaseByTokenID_NotFound", func(t *testing.T) {
		lease, err := leaseCache.GetLeaseByTokenID(ctx, 99999)
		assert.Error(t, err)
		assert.Equal(t, errors.ErrLeaseNotFound, err)
		assert.Nil(t, lease)
	})

	t.Run("DeleteLease", func(t *testing.T) {
		lease := builder.NewLease().WithTokenID(22222).WithPeerID("peer-delete").Build()

		// Set lease first
		err := leaseCache.SetLease(ctx, lease)
		require.NoError(t, err)

		// Verify lease exists by both keys
		retrievedByPeer, err := leaseCache.GetLeaseByPeerID(ctx, lease.PeerID)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedByPeer)

		retrievedByToken, err := leaseCache.GetLeaseByTokenID(ctx, lease.TokenID)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedByToken)

		// Delete lease
		err = leaseCache.DeleteLease(ctx, lease.PeerID, lease.TokenID)
		assert.NoError(t, err)

		// Verify lease is deleted from both keys
		deletedByPeer, err := leaseCache.GetLeaseByPeerID(ctx, lease.PeerID)
		assert.Error(t, err)
		assert.Equal(t, errors.ErrLeaseNotFound, err)
		assert.Nil(t, deletedByPeer)

		deletedByToken, err := leaseCache.GetLeaseByTokenID(ctx, lease.TokenID)
		assert.Error(t, err)
		assert.Equal(t, errors.ErrLeaseNotFound, err)
		assert.Nil(t, deletedByToken)
	})

	t.Run("SetLease_ExpiredTTL", func(t *testing.T) {
		// Create a lease with expired TTL
		lease := builder.NewLease().WithTokenID(33333).WithPeerID("peer-expired").WithExpired().Build()

		// Setting a lease with TTL <= 0 should not cache it
		err := leaseCache.SetLease(ctx, lease)
		assert.NoError(t, err)

		// Verify lease is not cached
		retrievedLease, err := leaseCache.GetLeaseByPeerID(ctx, lease.PeerID)
		assert.Error(t, err)
		assert.Equal(t, errors.ErrLeaseNotFound, err)
		assert.Nil(t, retrievedLease)
	})

	t.Run("ConcurrentOperations", func(t *testing.T) {
		const numGoroutines = 5
		const leasesPerGoroutine = 3

		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer func() { done <- true }()

				for j := 0; j < leasesPerGoroutine; j++ {
					tokenID := int64(goroutineID*1000 + j)
					peerID := fmt.Sprintf("concurrent-peer-%d-%d", goroutineID, j)

					lease := &models.Lease{
						TokenID:   tokenID,
						PeerID:    peerID,
						CreatedAt: time.Now(),
						ExpiresAt: time.Now().Add(time.Hour),
						Ttl:       3600,
					}

					// Set lease
					err := leaseCache.SetLease(ctx, lease)
					assert.NoError(t, err)

					// Get lease by peer ID
					retrievedByPeer, err := leaseCache.GetLeaseByPeerID(ctx, peerID)
					assert.NoError(t, err)
					assert.Equal(t, tokenID, retrievedByPeer.TokenID)

					// Get lease by token ID
					retrievedByToken, err := leaseCache.GetLeaseByTokenID(ctx, tokenID)
					assert.NoError(t, err)
					assert.Equal(t, peerID, retrievedByToken.PeerID)

					// Delete lease
					err = leaseCache.DeleteLease(ctx, peerID, tokenID)
					assert.NoError(t, err)
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})

	t.Run("MultipleKeysForSameLease", func(t *testing.T) {
		lease := builder.NewLease().WithTokenID(44444).WithPeerID("peer-multi-key").Build()

		// Set lease (should create both peer and token keys)
		err := leaseCache.SetLease(ctx, lease)
		require.NoError(t, err)

		// Verify both keys work
		retrievedByPeer, err := leaseCache.GetLeaseByPeerID(ctx, lease.PeerID)
		assert.NoError(t, err)
		assert.Equal(t, lease.TokenID, retrievedByPeer.TokenID)

		retrievedByToken, err := leaseCache.GetLeaseByTokenID(ctx, lease.TokenID)
		assert.NoError(t, err)
		assert.Equal(t, lease.PeerID, retrievedByToken.PeerID)

		// Verify both keys return the same lease data
		assert.Equal(t, retrievedByPeer.TokenID, retrievedByToken.TokenID)
		assert.Equal(t, retrievedByPeer.PeerID, retrievedByToken.PeerID)
		assert.Equal(t, retrievedByPeer.Ttl, retrievedByToken.Ttl)
	})
}
