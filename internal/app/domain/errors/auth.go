package errors

import "errors"

var (
	ErrMissingHeaders = errors.New("missing auth headers")
	ErrMissingPubkey  = errors.New("missing pubkey")
)
