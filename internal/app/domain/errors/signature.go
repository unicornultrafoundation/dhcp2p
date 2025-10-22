package errors

import "errors"

var (
	ErrInvalidSignature = errors.New("invalid signature")
	ErrInvalidPubkey    = errors.New("invalid pubkey")
	ErrPubkeyMismatch   = errors.New("pubkey mismatch")
)
