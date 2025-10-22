package models

import (
	"time"
)

type Nonce struct {
	ID        string
	PeerID    string
	IssuedAt  time.Time
	ExpiresAt time.Time
	Used      bool
	UsedAt    time.Time
}

type NonceRequest struct {
	NonceID   string
	Pubkey    []byte
	Payload   []byte
	Signature []byte
}
