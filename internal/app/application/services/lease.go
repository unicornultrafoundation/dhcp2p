package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/models"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/ports"
	"github.com/duchuongnguyen/dhcp2p/internal/app/infrastructure/config"
	"go.uber.org/zap"
)

type LeaseService struct {
	repo       ports.LeaseRepository
	logger     *zap.Logger
	maxRetries int
	retryDelay time.Duration
}

var _ ports.LeaseService = &LeaseService{}

func NewLeaseService(appConfig *config.AppConfig, repo ports.LeaseRepository, logger *zap.Logger) *LeaseService {
	return &LeaseService{repo, logger, appConfig.MaxLeaseRetries, time.Duration(appConfig.LeaseRetryDelay) * time.Millisecond}
}

func (s *LeaseService) AllocateIP(ctx context.Context, peerID string) (*models.Lease, error) {
	var lease *models.Lease
	var err error

	// Check if the lease is already allocated
	lease, err = s.repo.GetLeaseByPeerID(ctx, peerID)
	if lease != nil && err == nil {
		return lease, nil
	}

	maxRetries := s.maxRetries
	retries := 0

	// Try to find an expired lease for reuse
	for {
		retries++
		if retries > maxRetries {
			// If we've tried too many times, try to allocate a new lease
			break
		}

		lease, err = s.repo.FindAndReuseExpiredLease(ctx, peerID)
		if err != nil {
			// If we encounter an error, try again
			s.logger.With(zap.String("retries", strconv.Itoa(retries)), zap.String("peerID", peerID)).Error("error finding and reusing expired lease", zap.Error(err))

			// Sleep for 500 milliseconds
			time.Sleep(s.retryDelay)
			continue
		}

		if lease != nil {
			// Already reused, return the lease
			return lease, nil
		}
	}

	// If no expired lease found, allocate a new one
	retries = 0
	for {
		retries++
		if retries > maxRetries {
			return nil, fmt.Errorf("failed to allocate new lease: %v", err)
		}

		lease, err = s.repo.AllocateNewLease(ctx, peerID)
		if err != nil {
			s.logger.
				With(zap.String("retries", strconv.Itoa(retries)), zap.String("peerID", peerID)).
				Error("error allocating new lease", zap.Error(err))

			// Sleep for retry delay
			time.Sleep(s.retryDelay)
			continue
		}

		if lease != nil {
			return lease, nil
		}
	}
}

func (s *LeaseService) GetLeaseByPeerID(ctx context.Context, peerID string) (*models.Lease, error) {
	return s.repo.GetLeaseByPeerID(ctx, peerID)
}

func (s *LeaseService) GetLeaseByTokenID(ctx context.Context, tokenID int64) (*models.Lease, error) {
	return s.repo.GetLeaseByTokenID(ctx, tokenID)
}

func (s *LeaseService) RenewLease(ctx context.Context, tokenID int64, peerID string) (*models.Lease, error) {
	return s.repo.RenewLease(ctx, tokenID, peerID)
}

func (s *LeaseService) ReleaseLease(ctx context.Context, tokenID int64, peerID string) error {
	return s.repo.ReleaseLease(ctx, tokenID, peerID)
}
