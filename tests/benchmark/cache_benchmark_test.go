//go:build benchmark

package benchmark

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	redisclient "github.com/redis/go-redis/v9"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/repositories/redis"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/infrastructure/config"
	"github.com/unicornultrafoundation/dhcp2p/tests/fixtures"
)

func setupRedisForBenchmark(b *testing.B) *redisclient.Client {
	// Try to connect to external Redis
	client := redisclient.NewClient(&redisclient.Options{
		Addr:     "127.0.0.1:6380",  // Use port 6380 to avoid conflicts
		Password: "",
		DB:       0,
	})

	// Try to ping Redis with retries - give it more time since container just started
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := client.Ping(ctx).Err()
		cancel()
		
		if err == nil {
			// Only log on the first successful connection
			if i == 0 {
				b.Logf("Successfully connected to Redis")
			}
			return client
		}
		
		if i < 9 {
			// Only log failures on the first attempt to reduce noise
			if i == 0 {
				b.Logf("Redis connection attempt failed, retrying...")
			}
			time.Sleep(1 * time.Second)
		}
	}

	b.Skip("Skipping Redis benchmark: Could not connect to Redis at 127.0.0.1:6380 after multiple attempts")
	return nil
}

func BenchmarkRedisCache_SetLease(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark test")
	}

	ctx := context.Background()
	redisClient := setupRedisForBenchmark(b)

	cfg := &config.AppConfig{
		LeaseTTL: 3600,
	}

	leaseCache := redis.NewLeaseCache(redisClient, cfg)
	builder := fixtures.NewTestBuilder()

	var tokenCounter int64
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			currentToken := atomic.AddInt64(&tokenCounter, 1)
			lease := builder.NewLease().
				WithTokenID(int64(167772161 + currentToken - 1)).
				WithPeerID("benchmark-peer").
				Build()

			err := leaseCache.SetLease(ctx, lease)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkRedisCache_GetLeaseByPeerID(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark test")
	}

	ctx := context.Background()
	redisClient := setupRedisForBenchmark(b)

	cfg := &config.AppConfig{
		LeaseTTL: 3600,
	}

	leaseCache := redis.NewLeaseCache(redisClient, cfg)
	builder := fixtures.NewTestBuilder()

	// Clear existing data before running benchmark
	redisClient.FlushDB(ctx)

	// Pre-populate cache with a fixed set before timer starts
	for i := 0; i < 100; i++ {
		lease := builder.NewLease().
			WithTokenID(int64(167772161 + i)).
			WithPeerID("benchmark-peer").
			Build()
		if err := leaseCache.SetLease(ctx, lease); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := leaseCache.GetLeaseByPeerID(ctx, "benchmark-peer")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkRedisCache_ConcurrentOperations(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark test")
	}

	ctx := context.Background()
	redisClient := setupRedisForBenchmark(b)

	cfg := &config.AppConfig{
		LeaseTTL: 3600,
	}

	leaseCache := redis.NewLeaseCache(redisClient, cfg)
	builder := fixtures.NewTestBuilder()

	var tokenCounter int64
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			currentToken := atomic.AddInt64(&tokenCounter, 1)
			// Use unique peerID per goroutine to avoid race conditions
			workerID := currentToken % 100 // 100 workers max
			tokenID := int64(167772161 + (currentToken-1)%1000) // Smaller range
			peerID := fmt.Sprintf("worker-%d", workerID)

			lease := builder.NewLease().WithTokenID(tokenID).WithPeerID(peerID).Build()

			// Set lease
			err := leaseCache.SetLease(ctx, lease)
			if err != nil {
				b.Fatal(err)
			}

			// Get lease by peerID (should find the lease we just set)
			retrievedLease, err := leaseCache.GetLeaseByPeerID(ctx, peerID)
			if err != nil {
				b.Fatal(err)
			}

			// Get lease by tokenID (should also work)
			retrievedLease2, err := leaseCache.GetLeaseByTokenID(ctx, tokenID)
			if err != nil {
				b.Fatal(err)
			}

			// Delete lease (use the retrieved lease's actual values)
			if retrievedLease != nil {
				err = leaseCache.DeleteLease(ctx, retrievedLease.PeerID, retrievedLease.TokenID)
				if err != nil {
					b.Fatal(err)
				}
			} else if retrievedLease2 != nil {
				err = leaseCache.DeleteLease(ctx, retrievedLease2.PeerID, retrievedLease2.TokenID)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}
