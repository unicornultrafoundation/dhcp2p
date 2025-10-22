package libp2p

import (
	"context"

	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/errors"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/ports"
	"github.com/libp2p/go-libp2p/core/crypto"
)

type SignatureVerifier struct {
}

var _ ports.SignatureVerifier = &SignatureVerifier{}

func NewSignatureVerifier() *SignatureVerifier {
	return &SignatureVerifier{}
}

func (s *SignatureVerifier) VerifySignature(ctx context.Context, publicKey []byte, payload []byte, signature []byte) error {
	pubKey, err := crypto.UnmarshalPublicKey(publicKey)
	if err != nil {
		return err
	}

	ok, err := pubKey.Verify(payload, signature)
	if err != nil {
		return err
	}
	if !ok {
		return errors.ErrInvalidSignature
	}
	return nil
}
