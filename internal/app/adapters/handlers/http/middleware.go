package http

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"

	"github.com/duchuongnguyen/dhcp2p/internal/app/application/utils"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/errors"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/models"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/ports"
)

const (
	PeerIDContextKey = "peerID"
)

func WithAuth(authService ports.AuthService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pubkey := r.Header.Get("X-Pubkey")
			nonce := r.Header.Get("X-Nonce")
			signature := r.Header.Get("X-Signature")

			if pubkey == "" || nonce == "" || signature == "" {
				writeErrorResponse(w, http.StatusBadRequest, errors.ErrMissingHeaders)
				return
			}

			pub, err := base64.StdEncoding.DecodeString(pubkey)
			if err != nil {
				writeErrorResponse(w, http.StatusBadRequest, errors.ErrInvalidPubkey)
				return
			}

			sig, err := base64.StdEncoding.DecodeString(signature)
			if err != nil {
				writeErrorResponse(w, http.StatusBadRequest, errors.ErrInvalidSignature)
				return
			}

			res, err := authService.VerifyAuth(r.Context(), &models.AuthVerifyRequest{
				Pubkey:    pub,
				NonceID:   nonce,
				Signature: sig,
			})
			if err != nil {
				writeErrorResponse(w, http.StatusInternalServerError, err)
				return
			}

			if !bytes.Equal(res.Pubkey, pub) {
				writeErrorResponse(w, http.StatusBadRequest, errors.ErrPubkeyMismatch)
				return
			}

			// Set peerID to context
			peerID, err := utils.GetPeerIDFromPubkey(res.Pubkey)
			if err != nil {
				writeErrorResponse(w, http.StatusInternalServerError, err)
				return
			}
			ctx := context.WithValue(r.Context(), PeerIDContextKey, peerID)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)

		})
	}
}
