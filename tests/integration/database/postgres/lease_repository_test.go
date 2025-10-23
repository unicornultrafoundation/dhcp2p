//go:build integration

package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/duchuongnguyen/dhcp2p/internal/app/adapters/repositories/postgres"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/models"
	"github.com/duchuongnguyen/dhcp2p/internal/app/infrastructure/config"
	"github.com/duchuongnguyen/dhcp2p/tests/helpers"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	postgresModule "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestLeaseRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()

	// Start PostgreSQL container
	postgresContainer, err := postgresModule.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgresModule.WithDatabase("dhcp2p_test"),
		postgresModule.WithUsername("test"),
		postgresModule.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)
	defer postgresContainer.Terminate(ctx)

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Run migrations
	err = helpers.RunMigrations(connStr)
	require.NoError(t, err)

	// Create repository
	cfg := &config.AppConfig{
		LeaseTTL: 60, // 60 minutes
	}

	// Create pgxpool from connection string
	dbPool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	defer dbPool.Close()

	repo := postgres.NewLeaseRepository(cfg, dbPool)

	t.Run("AllocateNewLease", func(t *testing.T) {
		lease, err := repo.AllocateNewLease(ctx, "peer123")
		assert.NoError(t, err)
		assert.NotNil(t, lease)
		assert.Equal(t, "peer123", lease.PeerID)
		assert.True(t, lease.TokenID > 0)
		assert.True(t, lease.ExpiresAt.After(time.Now()))
	})

	t.Run("GetLeaseByPeerID", func(t *testing.T) {
		// First allocate a lease
		lease, err := repo.AllocateNewLease(ctx, "peer456")
		require.NoError(t, err)

		// Then retrieve it
		retrieved, err := repo.GetLeaseByPeerID(ctx, "peer456")
		assert.NoError(t, err)
		assert.Equal(t, lease.TokenID, retrieved.TokenID)
		assert.Equal(t, lease.PeerID, retrieved.PeerID)
	})

	t.Run("GetLeaseByTokenID", func(t *testing.T) {
		// First allocate a lease
		lease, err := repo.AllocateNewLease(ctx, "peer789")
		require.NoError(t, err)

		// Then retrieve it by token ID
		retrieved, err := repo.GetLeaseByTokenID(ctx, lease.TokenID)
		assert.NoError(t, err)
		assert.Equal(t, lease.TokenID, retrieved.TokenID)
		assert.Equal(t, lease.PeerID, retrieved.PeerID)
	})

	t.Run("RenewLease", func(t *testing.T) {
		// First allocate a lease
		lease, err := repo.AllocateNewLease(ctx, "peer-renew")
		require.NoError(t, err)

		// Renew the lease
		renewed, err := repo.RenewLease(ctx, lease.TokenID, "peer-renew")
		assert.NoError(t, err)
		assert.Equal(t, lease.TokenID, renewed.TokenID)
		assert.Equal(t, lease.PeerID, renewed.PeerID)
		assert.True(t, renewed.ExpiresAt.After(lease.ExpiresAt))
	})

	t.Run("ReleaseLease", func(t *testing.T) {
		// First allocate a lease
		lease, err := repo.AllocateNewLease(ctx, "peer-release")
		require.NoError(t, err)

		// Release the lease
		err = repo.ReleaseLease(ctx, lease.TokenID, "peer-release")
		assert.NoError(t, err)

		// Verify the lease is no longer retrievable
		retrieved, err := repo.GetLeaseByPeerID(ctx, "peer-release")
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("FindAndReuseExpiredLease", func(t *testing.T) {
		// Create an expired lease manually
		dbHelper, err := helpers.NewDatabaseHelper(connStr)
		require.NoError(t, err)
		defer dbHelper.Close()

		expiredTime := time.Now().Add(-time.Hour) // 1 hour ago
		err = dbHelper.InsertTestLease(ctx, 167772200, "expired-peer", expiredTime)
		require.NoError(t, err)

		// Try to find and reuse the expired lease
		reusedLease, err := repo.FindAndReuseExpiredLease(ctx, "expired-peer")
		assert.NoError(t, err)
		assert.NotNil(t, reusedLease)
		assert.Equal(t, int64(167772200), reusedLease.TokenID)
		assert.Equal(t, "expired-peer", reusedLease.PeerID)
		assert.True(t, reusedLease.ExpiresAt.After(time.Now()))
	})

	t.Run("ConcurrentAllocations", func(t *testing.T) {
		const numGoroutines = 10
		const peerPrefix = "concurrent-peer"

		results := make(chan *models.Lease, numGoroutines)
		errors := make(chan error, numGoroutines)

		// Start multiple goroutines to allocate leases concurrently
		for i := 0; i < numGoroutines; i++ {
			go func(peerID string) {
				lease, err := repo.AllocateNewLease(ctx, peerID)
				if err != nil {
					errors <- err
					return
				}
				results <- lease
			}(fmt.Sprintf("%s-%d", peerPrefix, i))
		}

		// Collect results
		var leases []*models.Lease
		for i := 0; i < numGoroutines; i++ {
			select {
			case lease := <-results:
				leases = append(leases, lease)
			case err := <-errors:
				t.Errorf("Error in goroutine: %v", err)
			case <-time.After(10 * time.Second):
				t.Fatal("Timeout waiting for concurrent allocations")
			}
		}

		// Verify all leases were allocated successfully
		assert.Len(t, leases, numGoroutines)

		// Verify all token IDs are unique
		tokenIDs := make(map[int64]bool)
		for _, lease := range leases {
			assert.False(t, tokenIDs[lease.TokenID], "Duplicate token ID: %d", lease.TokenID)
			tokenIDs[lease.TokenID] = true
		}
	})
}
