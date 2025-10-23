package services

import (
	"context"
	"crypto/sha256"

	"github.com/unicornultrafoundation/dhcp2p/internal/app/application/utils"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/errors"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/models"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/ports"
)

type AuthService struct {
	nonceService ports.NonceService
}

var _ ports.AuthService = &AuthService{}

func NewAuthService(nonceService ports.NonceService) *AuthService {
	return &AuthService{nonceService}
}

func (s *AuthService) RequestAuth(ctx context.Context, request *models.AuthRequest) (*models.AuthResponse, error) {
	if request == nil {
		return nil, errors.ErrMissingPubkey
	}
	peerID, err := utils.GetPeerIDFromPubkey(request.Pubkey)
	if err != nil {
		return nil, err
	}
	nonce, err := s.nonceService.CreateNonce(ctx, peerID)
	if err != nil {
		return nil, err
	}

	response := &models.AuthResponse{
		NonceID: nonce.ID,
	}

	return response, nil
}

func (s *AuthService) VerifyAuth(ctx context.Context, request *models.AuthVerifyRequest) (*models.AuthVerifyResponse, error) {
	if request == nil {
		return nil, errors.ErrMissingPeerID
	}

	payload := sha256.Sum256([]byte(request.NonceID))

	// Check if signature is not nil
	if request.Signature == nil {
		return nil, errors.ErrInvalidSignature
	}

	// Verify nonce
	err := s.nonceService.VerifyNonce(ctx, &models.NonceRequest{
		NonceID:   request.NonceID,
		Pubkey:    request.Pubkey,
		Payload:   payload[:],
		Signature: request.Signature,
	})
	if err != nil {
		return nil, err
	}

	// Nonce is valid, return the peerID
	response := &models.AuthVerifyResponse{
		Pubkey: request.Pubkey,
	}

	return response, nil
}
