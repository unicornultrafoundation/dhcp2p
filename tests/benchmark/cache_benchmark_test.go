//go:build benchmark

package benchmark

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/duchuongnguyen/dhcp2p/internal/app/adapters/repositories/redis"
	"github.com/duchuongnguyen/dhcp2p/internal/app/infrastructure/config"
	"github.com/duchuongnguyen/dhcp2p/tests/helpers"
	"github.com/duchuongnguyen/dhcp2p/tests/fixtures"
	redisclient "github.com/redis/go-redis/v9"
)

func BenchmarkRedisCache_SetLease(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark test")
	}

	ctx := context.Background()
	redisContainer, err := helpers.StartRedisContainer(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer redisContainer.Terminate(ctx)

	redisHost, err := redisContainer.Host(ctx)
	if err != nil {
		b.Fatal(err)
	}

	redisPort, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		b.Fatal(err)
	}

	redisClient := redisclient.NewClient(&redisclient.Options{
		Addr:     redisHost + ":" + redisPort.Port(),
		Password: "",
		DB:       0,
	})

	cfg := &config.AppConfig{
		LeaseTTL: 3600,
	}

	leaseCache := redis.NewLeaseCache(redisClient, cfg)
	builder := fixtures.NewTestBuilder()

	// Pre-warm Redis
	for i := 0; i < 100; i++ {
		lease := builder.NewLease().WithTokenID(int64(167772161 + i)).Build()
		leaseCache.SetLease(ctx, lease)
		leaseCache.DeleteLease(ctx, lease.PeerID, lease.TokenID)
	}

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
	redisContainer, err := helpers.StartRedisContainer(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer redisContainer.Terminate(ctx)

	redisHost, err := redisContainer.Host(ctx)
	if err != nil {
		b.Fatal(err)
	}

	redisPort, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		b.Fatal(err)
	}

	redisClient := redisclient.NewClient(&redisclient.Options{
		Addr:     redisHost + ":" + redisPort.Port(),
		Password: "",
		DB:       0,
	})

	cfg := &config.AppConfig{
		LeaseTTL: 3600,
	}

	leaseCache := redis.NewLeaseCache(redisClient, cfg)
	builder := fixtures.NewTestBuilder()

	// Pre-populate cache
	for i := 0; i < b.N; i++ {
		lease := builder.NewLease().
			WithTokenID(int64(167772161 + i)).
			WithPeerID("benchmark-peer").
			Build()
		leaseCache.SetLease(ctx, lease)
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
	redisContainer, err := helpers.StartRedisContainer(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer redisContainer.Terminate(ctx)

	redisHost, err := redisContainer.Host(ctx)
	if err != nil {
		b.Fatal(err)
	}

	redisPort, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		b.Fatal(err)
	}

	redisClient := redisclient.NewClient(&redisclient.Options{
		Addr:     redisHost + ":" + redisPort.Port(),
		Password: "",
		DB:       0,
	})

	cfg := &config.AppConfig{
		LeaseTTL: 3600,
	}

	leaseCache := redis.NewLeaseCache(redisClient, cfg)
	builder := fixtures.NewTestBuilder()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			tokenID := int64(167772161 + i)
			peerID := "benchmark-peer"
			
			lease := builder.NewLease().WithTokenID(tokenID).WithPeerID(peerID).Build()
			
			// Set lease
			err := leaseCache.SetLease(ctx, lease)
			if err != nil {
				b.Fatal(err)
			}
			
			// Get lease
			_, err = leaseCache.GetLeaseByPeerID(ctx, peerID)
			if err != nil {
				b.Fatal(err)
			}
			
			// Delete lease
			err = leaseCache.DeleteLease(ctx, peerID, tokenID)
			if err != nil {
				b.Fatal(err)
			}
			
			i++
		}
	})
}
