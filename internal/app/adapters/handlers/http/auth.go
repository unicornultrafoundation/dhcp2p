package http

import (
	"encoding/base64"
	"net/http"

	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/errors"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/models"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/ports"
)

type AuthHandler struct {
	authService ports.AuthService
}

func NewAuthHandler(authService ports.AuthService) *AuthHandler {
	return &AuthHandler{authService}
}

func (h *AuthHandler) RequestAuth(w http.ResponseWriter, r *http.Request) {
	pubkey := r.Header.Get("X-Pubkey")
	pub, err := base64.StdEncoding.DecodeString(pubkey)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, errors.ErrInvalidPubkey)
		return
	}

	nonce, err := h.authService.RequestAuth(r.Context(), &models.AuthRequest{
		Pubkey: pub,
	})
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	res := &AuthResponse{
		Pubkey: pubkey,
		Nonce:  nonce.NonceID,
	}

	writeResponse(w, http.StatusOK, res)
}
