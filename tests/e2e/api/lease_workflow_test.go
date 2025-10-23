//go:build e2e

package api

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unicornultrafoundation/dhcp2p/tests/helpers"
)

func TestLeaseWorkflow_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test")
	}

	ctx := context.Background()

	// Start full stack
	stack, err := helpers.StartTestStack(ctx)
	require.NoError(t, err)
	defer stack.Terminate(ctx)

	// Wait for services to be ready (skip this in unit test environment)
	// In a real e2e environment, you would start the application here
	// and wait for it to be ready before running tests

	// For now, we'll just test that the containers are running
	require.NotEmpty(t, stack.PostgresConnStr)
	require.NotEmpty(t, stack.RedisConnStr)

	// Run migrations to set up the database schema
	err = helpers.RunMigrations(stack.PostgresConnStr)
	require.NoError(t, err)

	// Note: In a real e2e environment, you would start the application here
	// and test against actual HTTP endpoints. For now, we're testing the
	// container setup and database connectivity.

	t.Run("Container Connectivity Test", func(t *testing.T) {
		// Test that we can connect to PostgreSQL
		dbHelper, err := helpers.NewDatabaseHelper(stack.PostgresConnStr)
		require.NoError(t, err)
		defer dbHelper.Close()

		// Test basic database operations
		ctx := context.Background()
		count, err := dbHelper.GetLeaseCount(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count) // Should be empty initially

		// Test that we can insert and query data
		err = dbHelper.InsertTestLease(ctx, 167772161, "test-peer", time.Now().Add(time.Hour))
		require.NoError(t, err)

		count, err = dbHelper.GetLeaseCount(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Clean up
		err = dbHelper.CleanupTables(ctx)
		require.NoError(t, err)
	})

	t.Run("Database Error Handling", func(t *testing.T) {
		// Test database error scenarios
		dbHelper, err := helpers.NewDatabaseHelper(stack.PostgresConnStr)
		require.NoError(t, err)
		defer dbHelper.Close()

		ctx := context.Background()

		// Test inserting duplicate primary key (should fail)
		err = dbHelper.InsertTestLease(ctx, 167772161, "peer1", time.Now().Add(time.Hour))
		require.NoError(t, err)

		// Try to insert another lease with the same token ID (should fail)
		err = dbHelper.InsertTestLease(ctx, 167772161, "peer2", time.Now().Add(time.Hour))
		require.Error(t, err) // Should fail due to primary key constraint

		// Clean up
		err = dbHelper.CleanupTables(ctx)
		require.NoError(t, err)
	})

	t.Run("Concurrent Database Operations", func(t *testing.T) {
		const numConcurrent = 5
		results := make(chan struct {
			peerID string
			err    error
		}, numConcurrent)

		// Start concurrent database operations
		for i := 0; i < numConcurrent; i++ {
			go func(i int) {
				dbHelper, err := helpers.NewDatabaseHelper(stack.PostgresConnStr)
				if err != nil {
					results <- struct {
						peerID string
						err    error
					}{"", err}
					return
				}
				defer dbHelper.Close()

				peerID := fmt.Sprintf("concurrent-peer-%d", i)
				ctx := context.Background()

				err = dbHelper.InsertTestLease(ctx, int64(167772161+i), peerID, time.Now().Add(time.Hour))

				results <- struct {
					peerID string
					err    error
				}{peerID, err}
			}(i)
		}

		// Collect results
		var successfulInserts []string
		for i := 0; i < numConcurrent; i++ {
			select {
			case result := <-results:
				require.NoError(t, result.err, "Error for peer %s", result.peerID)
				successfulInserts = append(successfulInserts, result.peerID)
			case <-time.After(10 * time.Second):
				t.Fatal("Timeout waiting for concurrent database operations")
			}
		}

		// Verify all operations succeeded
		assert.Equal(t, numConcurrent, len(successfulInserts))

		// Clean up
		dbHelper, err := helpers.NewDatabaseHelper(stack.PostgresConnStr)
		require.NoError(t, err)
		defer dbHelper.Close()

		ctx := context.Background()
		err = dbHelper.CleanupTables(ctx)
		require.NoError(t, err)
	})
}
