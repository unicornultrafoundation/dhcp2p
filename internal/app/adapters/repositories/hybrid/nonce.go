package hybrid

import (
	"context"

	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/models"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/ports"
	"go.uber.org/zap"
)

type NonceRepository struct {
	dbRepo ports.NonceRepository
	cache  ports.NonceCache
	logger *zap.Logger
}

var _ ports.NonceRepository = &NonceRepository{}

func NewNonceRepository(dbRepo ports.NonceRepository, cache ports.NonceCache, logger *zap.Logger) *NonceRepository {
	return &NonceRepository{dbRepo, cache, logger}
}

func (r *NonceRepository) GetNonce(ctx context.Context, nonceID string) (*models.Nonce, error) {
	// Try cache first
	nonce, err := r.cache.GetNonce(ctx, nonceID)
	if err == nil {
		return nonce, nil
	}

	// Fallback to database
	nonce, err = r.dbRepo.GetNonce(ctx, nonceID)
	if err != nil {
		return nil, err
	}

	// Cache the result for future requests
	if cacheErr := r.cache.CreateNonce(ctx, nonce); cacheErr != nil {
		r.logger.Warn("Failed to cache nonce", zap.Error(cacheErr))
	}

	return nonce, nil
}

func (r *NonceRepository) CreateNonce(ctx context.Context, peerID string) (*models.Nonce, error) {
	// Create in database first
	nonce, err := r.dbRepo.CreateNonce(ctx, peerID)
	if err != nil {
		return nil, err
	}

	// Cache the new nonce
	if cacheErr := r.cache.CreateNonce(ctx, nonce); cacheErr != nil {
		r.logger.Warn("Failed to cache new nonce", zap.Error(cacheErr))
	}

	return nonce, nil
}

func (r *NonceRepository) ConsumeNonce(ctx context.Context, nonceID string, peerID string) error {
	// Update database
	err := r.dbRepo.ConsumeNonce(ctx, nonceID, peerID)
	if err != nil {
		return err
	}

	// Remove from cache
	if cacheErr := r.cache.DeleteNonce(ctx, nonceID); cacheErr != nil {
		r.logger.Warn("Failed to remove nonce from cache", zap.Error(cacheErr))
	}

	return nil
}

func (r *NonceRepository) DeleteExpiredNonces(ctx context.Context) error {
	// Only database cleanup needed - Redis TTL handles cache cleanup
	return r.dbRepo.DeleteExpiredNonces(ctx)
}
