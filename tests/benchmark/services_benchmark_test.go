//go:build benchmark

package benchmark

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/unicornultrafoundation/dhcp2p/internal/app/application/services"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/infrastructure/config"
	"github.com/unicornultrafoundation/dhcp2p/tests/fixtures"
	"github.com/unicornultrafoundation/dhcp2p/tests/mocks"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func BenchmarkLeaseService_AllocateIP(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockLeaseRepository(ctrl)
	builder := fixtures.NewTestBuilder()
	service := services.NewLeaseService(&config.AppConfig{
		MaxLeaseRetries: 3,
		LeaseRetryDelay: 100,
	}, mockRepo, zap.NewNop())

	lease := builder.NewLease().Build()

	mockRepo.EXPECT().GetLeaseByPeerID(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockRepo.EXPECT().FindAndReuseExpiredLease(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockRepo.EXPECT().AllocateNewLease(gomock.Any(), gomock.Any()).Return(lease, nil).AnyTimes()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := service.AllocateIP(context.Background(), "benchmark-peer")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkLeaseService_GetLeaseByPeerID(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockLeaseRepository(ctrl)
	builder := fixtures.NewTestBuilder()
	service := services.NewLeaseService(&config.AppConfig{}, mockRepo, zap.NewNop())

	lease := builder.NewLease().Build()

	mockRepo.EXPECT().GetLeaseByPeerID(gomock.Any(), "benchmark-peer").Return(lease, nil).AnyTimes()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := service.GetLeaseByPeerID(context.Background(), "benchmark-peer")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkLeaseService_RenewLease(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockLeaseRepository(ctrl)
	builder := fixtures.NewTestBuilder()
	service := services.NewLeaseService(&config.AppConfig{}, mockRepo, zap.NewNop())

	lease := builder.NewLease().Build()

	mockRepo.EXPECT().RenewLease(gomock.Any(), gomock.Any(), gomock.Any()).Return(lease, nil).AnyTimes()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := service.RenewLease(context.Background(), 167772161, "benchmark-peer")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkLeaseBuilder_CreateLeases(b *testing.B) {
	builder := fixtures.NewTestBuilder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lease := builder.NewLease().
			WithTokenID(int64(167772161 + i)).
			WithPeerID("benchmark-peer").
			Build()
		_ = lease
	}
}

func BenchmarkNonceBuilder_CreateNonces(b *testing.B) {
	builder := fixtures.NewTestBuilder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nonce := builder.NewNonce().
			WithID("benchmark-nonce").
			WithPeerID("benchmark-peer").
			Build()
		_ = nonce
	}
}

func BenchmarkConcurrentLeaseAllocation(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockLeaseRepository(ctrl)
	builder := fixtures.NewTestBuilder()
	service := services.NewLeaseService(&config.AppConfig{
		MaxLeaseRetries: 3,
		LeaseRetryDelay: 10, // Lower delay for benchmarking
	}, mockRepo, zap.NewNop())

	lease := builder.NewLease().Build()

	mockRepo.EXPECT().GetLeaseByPeerID(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockRepo.EXPECT().FindAndReuseExpiredLease(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockRepo.EXPECT().AllocateNewLease(gomock.Any(), gomock.Any()).Return(lease, nil).AnyTimes()

	var allocCounter int64
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			current := atomic.AddInt64(&allocCounter, 1)
			peerID := fmt.Sprintf("benchmark-peer-%d", current-1)
			_, err := service.AllocateIP(context.Background(), peerID)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
