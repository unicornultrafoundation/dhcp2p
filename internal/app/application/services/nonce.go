package services

import (
	"context"

	"github.com/unicornultrafoundation/dhcp2p/internal/app/application/utils"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/models"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/ports"
)

type NonceService struct {
	repo              ports.NonceRepository
	signatureVerifier ports.SignatureVerifier
}

var _ ports.NonceService = &NonceService{}

func NewNonceService(repo ports.NonceRepository, signatureVerifier ports.SignatureVerifier) *NonceService {
	return &NonceService{repo, signatureVerifier}
}

func (s *NonceService) CreateNonce(ctx context.Context, peerID string) (*models.Nonce, error) {
	nonce, err := s.repo.CreateNonce(ctx, peerID)
	if err != nil {
		return nil, err
	}

	return nonce, nil
}

func (s *NonceService) VerifyNonce(ctx context.Context, request *models.NonceRequest) error {
	// Verify signature
	err := s.signatureVerifier.VerifySignature(ctx, request.Pubkey, request.Payload, request.Signature)
	if err != nil {
		return err
	}

	// Get nonce from database
	nonce, err := s.repo.GetNonce(ctx, request.NonceID)
	if err != nil {
		return err
	}

	peerID, err := utils.GetPeerIDFromPubkey(request.Pubkey)
	if err != nil {
		return err
	}

	// Try to consume nonce
	err = s.repo.ConsumeNonce(ctx, nonce.ID, peerID)
	if err != nil {
		return err
	}

	return nil
}
