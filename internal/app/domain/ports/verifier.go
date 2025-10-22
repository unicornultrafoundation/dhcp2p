package ports

import (
	"context"
)

type SignatureVerifier interface {
	VerifySignature(ctx context.Context, pubKey []byte, payload []byte, signature []byte) error
}
