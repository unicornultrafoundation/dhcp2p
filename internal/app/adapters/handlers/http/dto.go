package http

import "github.com/unicornultrafoundation/dhcp2p/internal/app/domain/models"

type AuthResponse struct {
	Pubkey string `json:"pubkey"`
	Nonce  string `json:"nonce"`
}

type AllocateRequestedIPRequest struct {
	IP string `json:"ip"`
}

type AllocateRequestedIPResponse struct {
	Lease *models.Lease `json:"lease"`
}

type AllocateDynamicIPResponse struct {
	Lease *models.Lease `json:"lease,omitempty"`
}

// Request data structures for type safety
type AuthRequestData struct {
	Pubkey []byte
}

type LeaseRequestData struct {
	PeerID string
}

type TokenIDRequestData struct {
	PeerID  string
	TokenID int64
}

type PeerIDRequestData struct {
	PeerID string
}
