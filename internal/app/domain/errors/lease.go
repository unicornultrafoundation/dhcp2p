package errors

import "errors"

var (
	ErrMissingPeerID  = errors.New("missing peer ID")
	ErrMissingTokenID = errors.New("missing token ID")
	ErrInvalidTokenID = errors.New("invalid token ID")
	ErrLeaseNotFound  = errors.New("lease not found")
)
