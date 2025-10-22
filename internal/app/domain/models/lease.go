package models

import (
	"time"
)

type Lease struct {
	TokenID   int64     `json:"token_id"`
	PeerID    string    `json:"peer_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Ttl       int32     `json:"ttl"`
}
