package postgres

import (
	"context"
	"time"

	qDb "github.com/duchuongnguyen/dhcp2p/internal/app/adapters/repositories/postgres/db"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/models"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/ports"
	"github.com/duchuongnguyen/dhcp2p/internal/app/infrastructure/config"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NonceRepository struct {
	query    *qDb.Queries
	nonceTTL time.Duration
}

var _ ports.NonceRepository = &NonceRepository{}

func NewNonceRepository(cfg *config.AppConfig, db *pgxpool.Pool) *NonceRepository {
	return &NonceRepository{qDb.New(db), time.Duration(cfg.NonceTTL) * time.Minute}
}

func (r *NonceRepository) GetNonce(ctx context.Context, nonceID string) (*models.Nonce, error) {
	var id pgtype.UUID
	err := id.Scan(nonceID)
	if err != nil {
		return nil, err
	}

	nonce, err := r.query.GetNonce(ctx, id)
	if err != nil {
		return nil, err
	}
	return &models.Nonce{
		ID:        nonce.ID.String(),
		PeerID:    nonce.PeerID,
		IssuedAt:  nonce.IssuedAt.Time,
		ExpiresAt: nonce.ExpiresAt.Time,
		Used:      nonce.Used,
		UsedAt:    nonce.UsedAt.Time,
	}, nil
}

func (r *NonceRepository) CreateNonce(ctx context.Context, peerID string) (*models.Nonce, error) {
	params := qDb.CreateNonceParams{
		PeerID: peerID,
		Ttl:    int32(r.nonceTTL.Minutes()),
	}

	nonce, err := r.query.CreateNonce(ctx, params)
	if err != nil {
		return nil, err
	}

	return &models.Nonce{
		ID:        nonce.ID.String(),
		PeerID:    peerID,
		IssuedAt:  nonce.IssuedAt.Time,
		ExpiresAt: nonce.ExpiresAt.Time,
		Used:      nonce.Used,
		UsedAt:    nonce.UsedAt.Time,
	}, nil
}

func (r *NonceRepository) ConsumeNonce(ctx context.Context, nonceID string, peerID string) error {
	var id pgtype.UUID
	err := id.Scan(nonceID)
	if err != nil {
		return err
	}
	_, err = r.query.ConsumeNonce(ctx, qDb.ConsumeNonceParams{
		ID:     id,
		PeerID: peerID,
	})
	return err
}

func (r *NonceRepository) DeleteExpiredNonces(ctx context.Context) error {
	return r.query.DeleteExpiredNonces(ctx)
}
