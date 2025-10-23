package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	qDb "github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/repositories/postgres/db"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/models"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/ports"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/infrastructure/config"
)

type LeaseRepository struct {
	pool     *pgxpool.Pool
	queries  *qDb.Queries
	leaseTTL time.Duration
}

var _ ports.LeaseRepository = &LeaseRepository{}

func NewLeaseRepository(cfg *config.AppConfig, db *pgxpool.Pool) *LeaseRepository {
	return &LeaseRepository{db, qDb.New(db), time.Duration(cfg.LeaseTTL) * time.Minute}
}

func (r *LeaseRepository) FindAndReuseExpiredLease(ctx context.Context, peerID string) (*models.Lease, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := r.queries.WithTx(tx)

	expired, err := q.FindExpiredLeaseForReuse(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	lease, err := q.ReuseLease(ctx, qDb.ReuseLeaseParams{
		PeerID:  peerID,
		TokenID: expired.TokenID,
		Ttl:     int32(r.leaseTTL.Minutes()),
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &models.Lease{
		TokenID:   lease.TokenID,
		PeerID:    lease.PeerID,
		ExpiresAt: lease.ExpiresAt.Time,
		CreatedAt: lease.CreatedAt.Time,
		UpdatedAt: lease.UpdatedAt.Time,
		Ttl:       lease.Ttl,
	}, nil
}

func (r *LeaseRepository) AllocateNewLease(ctx context.Context, peerID string) (*models.Lease, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := r.queries.WithTx(tx)

	tokenID, err := q.AllocateNextTokenID(ctx)
	if err != nil {
		return nil, err
	}

	lease, err := q.InsertLease(ctx, qDb.InsertLeaseParams{
		TokenID: tokenID,
		PeerID:  peerID,
		Ttl:     int32(r.leaseTTL.Minutes()),
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &models.Lease{
		TokenID:   lease.TokenID,
		PeerID:    lease.PeerID,
		ExpiresAt: lease.ExpiresAt.Time,
		CreatedAt: lease.CreatedAt.Time,
		UpdatedAt: lease.UpdatedAt.Time,
		Ttl:       lease.Ttl,
	}, nil
}

func (r *LeaseRepository) GetLeaseByTokenID(ctx context.Context, leaseID int64) (*models.Lease, error) {
	lease, err := r.queries.GetLeaseByTokenID(ctx, leaseID)
	if err != nil {
		return nil, err
	}
	return &models.Lease{
		TokenID:   lease.TokenID,
		PeerID:    lease.PeerID,
		ExpiresAt: lease.ExpiresAt.Time,
		CreatedAt: lease.CreatedAt.Time,
		UpdatedAt: lease.UpdatedAt.Time,
		Ttl:       lease.Ttl,
	}, nil
}

func (r *LeaseRepository) GetLeaseByPeerID(ctx context.Context, peerID string) (*models.Lease, error) {
	lease, err := r.queries.GetLeaseByPeerID(ctx, peerID)
	if err != nil {
		return nil, err
	}
	return &models.Lease{
		TokenID:   lease.TokenID,
		PeerID:    lease.PeerID,
		ExpiresAt: lease.ExpiresAt.Time,
		CreatedAt: lease.CreatedAt.Time,
		UpdatedAt: lease.UpdatedAt.Time,
		Ttl:       lease.Ttl,
	}, nil
}

func (r *LeaseRepository) RenewLease(ctx context.Context, tokenID int64, peerID string) (*models.Lease, error) {
	lease, err := r.queries.RenewLease(ctx, qDb.RenewLeaseParams{
		TokenID: tokenID,
		PeerID:  peerID,
		Ttl:     int32(r.leaseTTL.Minutes()),
	})
	if err != nil {
		return nil, err
	}
	return &models.Lease{
		TokenID:   lease.TokenID,
		PeerID:    lease.PeerID,
		ExpiresAt: lease.ExpiresAt.Time,
		CreatedAt: lease.CreatedAt.Time,
		UpdatedAt: lease.UpdatedAt.Time,
		Ttl:       lease.Ttl,
	}, nil
}

func (r *LeaseRepository) ReleaseLease(ctx context.Context, tokenID int64, peerID string) error {
	err := r.queries.ReleaseLease(ctx, qDb.ReleaseLeaseParams{
		TokenID: tokenID,
		PeerID:  peerID,
	})
	if err != nil {
		return err
	}
	return nil
}
