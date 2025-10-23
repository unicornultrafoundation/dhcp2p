package http

import (
	"context"
	"net/http"

	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/ports"
)

type LeaseHandler struct {
	leaseService ports.LeaseService
}

func NewLeaseHandler(leaseService ports.LeaseService) *LeaseHandler {
	return &LeaseHandler{leaseService}
}

func (h *LeaseHandler) AllocateIP(w http.ResponseWriter, r *http.Request) {
	sc := &ServiceCall{Handler: w, Request: r}
	sc.ExecuteWithValidation(
		h.handleAllocateIP,
		ValidateLeaseRequest,
	)
}

func (h *LeaseHandler) GetLeaseByPeerID(w http.ResponseWriter, r *http.Request) {
	sc := &ServiceCall{Handler: w, Request: r}
	sc.ExecuteWithValidation(
		h.handleGetLeaseByPeerID,
		ValidatePeerIDParamRequest,
	)
}

func (h *LeaseHandler) GetLeaseByTokenID(w http.ResponseWriter, r *http.Request) {
	sc := &ServiceCall{Handler: w, Request: r}
	sc.ExecuteWithValidation(
		h.handleGetLeaseByTokenID,
		ValidateTokenIDParamRequest,
	)
}

func (h *LeaseHandler) RenewLease(w http.ResponseWriter, r *http.Request) {
	sc := &ServiceCall{Handler: w, Request: r}
	sc.ExecuteWithValidation(
		h.handleRenewLease,
		ValidateTokenIDRequest,
	)
}

func (h *LeaseHandler) ReleaseLease(w http.ResponseWriter, r *http.Request) {
	sc := &ServiceCall{Handler: w, Request: r}
	sc.ExecuteWithValidation(
		h.handleReleaseLease,
		ValidateTokenIDRequest,
	)
}

// Business logic handlers

func (h *LeaseHandler) handleAllocateIP(ctx context.Context, req interface{}) (interface{}, error) {
	leaseReq := req.(*LeaseRequestData)
	return h.leaseService.AllocateIP(ctx, leaseReq.PeerID)
}

func (h *LeaseHandler) handleGetLeaseByPeerID(ctx context.Context, req interface{}) (interface{}, error) {
	peerReq := req.(*PeerIDRequestData)
	return h.leaseService.GetLeaseByPeerID(ctx, peerReq.PeerID)
}

func (h *LeaseHandler) handleGetLeaseByTokenID(ctx context.Context, req interface{}) (interface{}, error) {
	tokenReq := req.(*TokenIDRequestData)
	return h.leaseService.GetLeaseByTokenID(ctx, tokenReq.TokenID)
}

func (h *LeaseHandler) handleRenewLease(ctx context.Context, req interface{}) (interface{}, error) {
	tokenReq := req.(*TokenIDRequestData)
	return h.leaseService.RenewLease(ctx, tokenReq.TokenID, tokenReq.PeerID)
}

func (h *LeaseHandler) handleReleaseLease(ctx context.Context, req interface{}) (interface{}, error) {
	tokenReq := req.(*TokenIDRequestData)
	err := h.leaseService.ReleaseLease(ctx, tokenReq.TokenID, tokenReq.PeerID)
	if err != nil {
		return nil, err
	}
	return map[string]string{"status": "success"}, nil
}
