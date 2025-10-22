package hybrid

import (
	"context"

	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/models"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/ports"
	"go.uber.org/zap"
)

type LeaseRepository struct {
	dbRepo ports.LeaseRepository
	cache  ports.LeaseCache
	logger *zap.Logger
}

var _ ports.LeaseRepository = &LeaseRepository{}

func NewLeaseRepository(dbRepo ports.LeaseRepository, cache ports.LeaseCache, logger *zap.Logger) *LeaseRepository {
	return &LeaseRepository{dbRepo, cache, logger}
}

func (r *LeaseRepository) GetLeaseByPeerID(ctx context.Context, peerID string) (*models.Lease, error) {
	// Try cache first
	lease, err := r.cache.GetLeaseByPeerID(ctx, peerID)
	if err == nil {
		return lease, nil
	}
	// Log cache errors and fall back to DB
	r.logger.Debug("cache GetLeaseByPeerID failed, falling back to DB", zap.Error(err), zap.String("peerID", peerID))

	// Fallback to database
	lease, err = r.dbRepo.GetLeaseByPeerID(ctx, peerID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if cacheErr := r.cache.SetLease(ctx, lease); cacheErr != nil {
		r.logger.Warn("Failed to cache lease", zap.Error(cacheErr))
	}

	return lease, nil
}

func (r *LeaseRepository) GetLeaseByTokenID(ctx context.Context, tokenID int64) (*models.Lease, error) {
	// Try cache first
	lease, err := r.cache.GetLeaseByTokenID(ctx, tokenID)
	if err == nil {
		return lease, nil
	}
	r.logger.Debug("cache GetLeaseByTokenID failed, falling back to DB", zap.Error(err), zap.Int64("tokenID", tokenID))

	// Fallback to database
	lease, err = r.dbRepo.GetLeaseByTokenID(ctx, tokenID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if cacheErr := r.cache.SetLease(ctx, lease); cacheErr != nil {
		r.logger.Warn("Failed to cache lease", zap.Error(cacheErr))
	}

	return lease, nil
}

func (r *LeaseRepository) FindAndReuseExpiredLease(ctx context.Context, peerID string) (*models.Lease, error) {
	// This operation always goes to database (complex query)
	lease, err := r.dbRepo.FindAndReuseExpiredLease(ctx, peerID)
	if err != nil || lease == nil {
		return lease, err
	}

	// Cache the reused lease
	if cacheErr := r.cache.SetLease(ctx, lease); cacheErr != nil {
		r.logger.Warn("Failed to cache reused lease", zap.Error(cacheErr))
	}

	return lease, nil
}

func (r *LeaseRepository) AllocateNewLease(ctx context.Context, peerID string) (*models.Lease, error) {
	// Create in database
	lease, err := r.dbRepo.AllocateNewLease(ctx, peerID)
	if err != nil {
		return nil, err
	}

	// Cache the new lease
	if cacheErr := r.cache.SetLease(ctx, lease); cacheErr != nil {
		r.logger.Warn("Failed to cache new lease", zap.Error(cacheErr))
	}

	return lease, nil
}

func (r *LeaseRepository) RenewLease(ctx context.Context, tokenID int64, peerID string) (*models.Lease, error) {
	// Update database
	lease, err := r.dbRepo.RenewLease(ctx, tokenID, peerID)
	if err != nil {
		return nil, err
	}

	// Cache the renewed lease
	if cacheErr := r.cache.SetLease(ctx, lease); cacheErr != nil {
		r.logger.Warn("Failed to cache renewed lease", zap.Error(cacheErr))
	}

	return lease, nil
}

func (r *LeaseRepository) ReleaseLease(ctx context.Context, tokenID int64, peerID string) error {
	// Update database
	err := r.dbRepo.ReleaseLease(ctx, tokenID, peerID)
	if err != nil {
		return err
	}

	// Remove from cache
	if cacheErr := r.cache.DeleteLease(ctx, peerID, tokenID); cacheErr != nil {
		r.logger.Warn("Failed to remove lease from cache", zap.Error(cacheErr))
	}

	return nil
}
