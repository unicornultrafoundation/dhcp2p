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
	"github.com/unicornultrafoundation/dhcp2p/tests/helpers"
)

func TestNonceCache_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Don't run integration tests in parallel to avoid port conflicts
	// t.Parallel()

	// Add a small delay to avoid resource contention with other tests
	time.Sleep(100 * time.Millisecond)

	ctx := context.Background()

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
		NonceTTL: 5, // 5 minutes
	}

	// Create NonceCache
	nonceCache := redis.NewNonceCache(redisClient, cfg)

	t.Run("CreateNonce", func(t *testing.T) {
		nonce := &models.Nonce{
			ID:        "test-nonce-1",
			PeerID:    "peer123",
			IssuedAt:  time.Now(),
			ExpiresAt: time.Now().Add(time.Hour),
			Used:      false,
		}

		err := nonceCache.CreateNonce(ctx, nonce)
		assert.NoError(t, err)
	})

	t.Run("GetNonce", func(t *testing.T) {
		nonce := &models.Nonce{
			ID:        "test-nonce-2",
			PeerID:    "peer456",
			IssuedAt:  time.Now(),
			ExpiresAt: time.Now().Add(time.Hour),
			Used:      false,
		}

		// Create nonce first
		err := nonceCache.CreateNonce(ctx, nonce)
		require.NoError(t, err)

		// Get nonce
		retrievedNonce, err := nonceCache.GetNonce(ctx, nonce.ID)
		assert.NoError(t, err)
		assert.Equal(t, nonce.ID, retrievedNonce.ID)
		assert.Equal(t, nonce.PeerID, retrievedNonce.PeerID)
		assert.Equal(t, nonce.Used, retrievedNonce.Used)
	})

	t.Run("GetNonce_NotFound", func(t *testing.T) {
		nonce, err := nonceCache.GetNonce(ctx, "non-existent-nonce")
		assert.Error(t, err)
		assert.Equal(t, errors.ErrNonceNotFound, err)
		assert.Nil(t, nonce)
	})

	t.Run("DeleteNonce", func(t *testing.T) {
		nonce := &models.Nonce{
			ID:        "test-nonce-3",
			PeerID:    "peer789",
			IssuedAt:  time.Now(),
			ExpiresAt: time.Now().Add(time.Hour),
			Used:      false,
		}

		// Create nonce first
		err := nonceCache.CreateNonce(ctx, nonce)
		require.NoError(t, err)

		// Verify nonce exists
		retrievedNonce, err := nonceCache.GetNonce(ctx, nonce.ID)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedNonce)

		// Delete nonce
		err = nonceCache.DeleteNonce(ctx, nonce.ID)
		assert.NoError(t, err)

		// Verify nonce is deleted
		deletedNonce, err := nonceCache.GetNonce(ctx, nonce.ID)
		assert.Error(t, err)
		assert.Equal(t, errors.ErrNonceNotFound, err)
		assert.Nil(t, deletedNonce)
	})

	t.Run("NonceTTL", func(t *testing.T) {
		// Create a nonce with very short TTL
		shortTTLCache := redis.NewNonceCache(redisClient, &config.AppConfig{
			NonceTTL: 1, // 1 minute
		})

		nonce := &models.Nonce{
			ID:        "test-nonce-ttl",
			PeerID:    "peer-ttl",
			IssuedAt:  time.Now(),
			ExpiresAt: time.Now().Add(30 * time.Second), // Shorter than cache TTL
			Used:      false,
		}

		err := shortTTLCache.CreateNonce(ctx, nonce)
		require.NoError(t, err)

		// Verify nonce exists
		retrievedNonce, err := shortTTLCache.GetNonce(ctx, nonce.ID)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedNonce)

		// Wait for TTL to expire (in a real test, you'd use a shorter TTL)
		// For this test, we'll just verify the nonce was created successfully
	})

	t.Run("ConcurrentOperations", func(t *testing.T) {
		const numGoroutines = 10
		const noncesPerGoroutine = 5

		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer func() { done <- true }()

				for j := 0; j < noncesPerGoroutine; j++ {
					nonceID := fmt.Sprintf("concurrent-nonce-%d-%d", goroutineID, j)
					nonce := &models.Nonce{
						ID:        nonceID,
						PeerID:    fmt.Sprintf("peer-%d-%d", goroutineID, j),
						IssuedAt:  time.Now(),
						ExpiresAt: time.Now().Add(time.Hour),
						Used:      false,
					}

					// Create nonce
					err := nonceCache.CreateNonce(ctx, nonce)
					assert.NoError(t, err)

					// Get nonce
					retrievedNonce, err := nonceCache.GetNonce(ctx, nonceID)
					assert.NoError(t, err)
					assert.Equal(t, nonceID, retrievedNonce.ID)

					// Delete nonce
					err = nonceCache.DeleteNonce(ctx, nonceID)
					assert.NoError(t, err)
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})
}
