package http

import (
	"net/http"
	"strconv"

	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/errors"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/ports"
	"github.com/go-chi/chi/v5"
)

type LeaseHandler struct {
	leaseService ports.LeaseService
}

func NewLeaseHandler(leaseService ports.LeaseService) *LeaseHandler {
	return &LeaseHandler{leaseService}
}

func (h *LeaseHandler) AllocateIP(w http.ResponseWriter, r *http.Request) {
	peerIDValue := r.Context().Value(PeerIDContextKey)
	if peerIDValue == nil {
		writeDomainError(w, errors.ErrMissingPeerID)
		return
	}
	peerID, ok := peerIDValue.(string)
	if !ok || peerID == "" {
		writeDomainError(w, errors.ErrMissingPeerID)
		return
	}
	lease, err := h.leaseService.AllocateIP(r.Context(), peerID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeResponse(w, http.StatusOK, lease)
}

func (h *LeaseHandler) GetLeaseByPeerID(w http.ResponseWriter, r *http.Request) {
	peerID := chi.URLParam(r, "peerID")
	if peerID == "" {
		writeDomainError(w, errors.ErrMissingPeerID)
		return
	}
	lease, err := h.leaseService.GetLeaseByPeerID(r.Context(), peerID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeResponse(w, http.StatusOK, lease)
}

func (h *LeaseHandler) GetLeaseByTokenID(w http.ResponseWriter, r *http.Request) {
	tokenIDStr := chi.URLParam(r, "tokenID")
	if tokenIDStr == "" {
		writeDomainError(w, errors.ErrMissingTokenID)
		return
	}
	tokenID, err := strconv.ParseInt(tokenIDStr, 10, 64)
	if err != nil {
		writeDomainError(w, errors.ErrInvalidTokenID)
		return
	}
	lease, err := h.leaseService.GetLeaseByTokenID(r.Context(), tokenID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeResponse(w, http.StatusOK, lease)
}

func (h *LeaseHandler) RenewLease(w http.ResponseWriter, r *http.Request) {
	peerIDValue := r.Context().Value(PeerIDContextKey)
	if peerIDValue == nil {
		writeDomainError(w, errors.ErrMissingPeerID)
		return
	}
	peerID, ok := peerIDValue.(string)
	if !ok || peerID == "" {
		writeDomainError(w, errors.ErrMissingPeerID)
		return
	}
	tokenIDStr := r.URL.Query().Get("tokenID")
	if tokenIDStr == "" {
		writeDomainError(w, errors.ErrMissingTokenID)
		return
	}
	tokenID, err := strconv.ParseInt(tokenIDStr, 10, 64)
	if err != nil {
		writeDomainError(w, errors.ErrInvalidTokenID)
		return
	}
	lease, err := h.leaseService.RenewLease(r.Context(), tokenID, peerID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeResponse(w, http.StatusOK, lease)
}

func (h *LeaseHandler) ReleaseLease(w http.ResponseWriter, r *http.Request) {
	peerIDValue := r.Context().Value(PeerIDContextKey)
	if peerIDValue == nil {
		writeDomainError(w, errors.ErrMissingPeerID)
		return
	}
	peerID, ok := peerIDValue.(string)
	if !ok || peerID == "" {
		writeDomainError(w, errors.ErrMissingPeerID)
		return
	}
	tokenIDStr := r.URL.Query().Get("tokenID")
	if tokenIDStr == "" {
		writeDomainError(w, errors.ErrMissingTokenID)
		return
	}
	tokenID, err := strconv.ParseInt(tokenIDStr, 10, 64)
	if err != nil {
		writeDomainError(w, errors.ErrInvalidTokenID)
		return
	}

	err = h.leaseService.ReleaseLease(r.Context(), tokenID, peerID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeResponse(w, http.StatusOK, nil)
}
