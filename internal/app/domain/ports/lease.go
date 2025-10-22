package ports

import (
	"context"

	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/models"
)

type LeaseService interface {
	GetLeaseByPeerID(ctx context.Context, peerID string) (*models.Lease, error)
	GetLeaseByTokenID(ctx context.Context, tokenID int64) (*models.Lease, error)
	RenewLease(ctx context.Context, tokenID int64, peerID string) (*models.Lease, error)
	ReleaseLease(ctx context.Context, tokenID int64, peerID string) error
	AllocateIP(ctx context.Context, peerID string) (*models.Lease, error)
}

type LeaseRepository interface {
	FindAndReuseExpiredLease(ctx context.Context, peerID string) (*models.Lease, error)
	AllocateNewLease(ctx context.Context, peerID string) (*models.Lease, error)
	GetLeaseByTokenID(ctx context.Context, tokenID int64) (*models.Lease, error)
	GetLeaseByPeerID(ctx context.Context, peerID string) (*models.Lease, error)
	RenewLease(ctx context.Context, tokenID int64, peerID string) (*models.Lease, error)
	ReleaseLease(ctx context.Context, tokenID int64, peerID string) error
}

type LeaseCache interface {
	GetLeaseByPeerID(ctx context.Context, peerID string) (*models.Lease, error)
	GetLeaseByTokenID(ctx context.Context, tokenID int64) (*models.Lease, error)
	SetLease(ctx context.Context, lease *models.Lease) error
	DeleteLease(ctx context.Context, peerID string, tokenID int64) error
}
