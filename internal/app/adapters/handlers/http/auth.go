package http

import (
	"context"
	"encoding/base64"
	"net/http"

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
	sc := &ServiceCall{Handler: w, Request: r}
	sc.ExecuteWithValidation(
		h.handleAuthRequest,
		ValidateAuthRequest,
	)
}

// handleAuthRequest is the business logic handler for auth requests
func (h *AuthHandler) handleAuthRequest(ctx context.Context, req interface{}) (interface{}, error) {
	authReq := req.(*AuthRequestData)

	nonce, err := h.authService.RequestAuth(ctx, &models.AuthRequest{
		Pubkey: authReq.Pubkey,
	})
	if err != nil {
		return nil, err
	}

	// Encode pubkey back to base64 for response
	pubkeyStr := base64.StdEncoding.EncodeToString(authReq.Pubkey)

	return &AuthResponse{
		Pubkey: pubkeyStr,
		Nonce:  nonce.NonceID,
	}, nil
}
