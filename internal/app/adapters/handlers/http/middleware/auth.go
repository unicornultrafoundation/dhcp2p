package middleware

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"

	"github.com/duchuongnguyen/dhcp2p/internal/app/adapters/handlers/http/keys"
	"github.com/duchuongnguyen/dhcp2p/internal/app/adapters/handlers/http/utils"
	"github.com/duchuongnguyen/dhcp2p/internal/app/adapters/handlers/http/validation"
	applicationUtils "github.com/duchuongnguyen/dhcp2p/internal/app/application/utils"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/errors"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/models"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/ports"
)

func WithAuth(authService ports.AuthService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Security middleware handles request validation and sanitization
			// We only need to validate authentication-specific headers here

			// Validate headers using enhanced validation
			pubkeyResult := validation.ValidateHeader(r, "X-Pubkey", validation.PubkeyValidationConfig())
			if pubkeyResult.Error != nil {
				utils.WriteDomainError(w, pubkeyResult.Error)
				return
			}

			nonceResult := validation.ValidateHeader(r, "X-Nonce", validation.NonceValidationConfig())
			if nonceResult.Error != nil {
				utils.WriteDomainError(w, nonceResult.Error)
				return
			}

			signatureResult := validation.ValidateHeader(r, "X-Signature", validation.SignatureValidationConfig())
			if signatureResult.Error != nil {
				utils.WriteDomainError(w, signatureResult.Error)
				return
			}

			// Validate and decode base64 data
			pubkeyValidation := validation.ValidateBase64Pubkey(pubkeyResult.Value)
			if pubkeyValidation.Error != nil {
				utils.WriteDomainError(w, pubkeyValidation.Error)
				return
			}

			signatureValidation := validation.ValidateBase64Signature(signatureResult.Value)
			if signatureValidation.Error != nil {
				utils.WriteDomainError(w, signatureValidation.Error)
				return
			}

			// Decode the validated data
			pub, err := base64.StdEncoding.DecodeString(pubkeyValidation.Value)
			if err != nil {
				utils.WriteDomainError(w, errors.ErrInvalidPubkey)
				return
			}

			sig, err := base64.StdEncoding.DecodeString(signatureValidation.Value)
			if err != nil {
				utils.WriteDomainError(w, errors.ErrInvalidSignature)
				return
			}

			// Verify authentication
			res, err := authService.VerifyAuth(r.Context(), &models.AuthVerifyRequest{
				Pubkey:    pub,
				NonceID:   nonceResult.Value,
				Signature: sig,
			})
			if err != nil {
				utils.WriteDomainError(w, err)
				return
			}

			if !bytes.Equal(res.Pubkey, pub) {
				utils.WriteDomainError(w, errors.ErrPubkeyMismatch)
				return
			}

			// Set peerID to context
			peerID, err := applicationUtils.GetPeerIDFromPubkey(res.Pubkey)
			if err != nil {
				// Convert libp2p errors to validation errors
				utils.WriteDomainError(w, errors.ErrInvalidPubkey)
				return
			}
			ctx := context.WithValue(r.Context(), keys.PeerIDContextKey, peerID)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
