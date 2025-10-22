package ports

import (
	"context"

	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/models"
)

type NonceRepository interface {
	GetNonce(ctx context.Context, nonceID string) (*models.Nonce, error)
	CreateNonce(ctx context.Context, peerID string) (*models.Nonce, error)
	ConsumeNonce(ctx context.Context, nonceID string, peerID string) error
	DeleteExpiredNonces(ctx context.Context) error
}

type NonceCache interface {
	GetNonce(ctx context.Context, nonceID string) (*models.Nonce, error)
	CreateNonce(ctx context.Context, nonce *models.Nonce) error
	DeleteNonce(ctx context.Context, nonceID string) error
}

type NonceService interface {
	CreateNonce(ctx context.Context, peerID string) (*models.Nonce, error)
	VerifyNonce(ctx context.Context, request *models.NonceRequest) error
}

type NonceCleaner interface {
	Run(ctx context.Context) error
}
