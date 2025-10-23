package http

import (
	"context"
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/duchuongnguyen/dhcp2p/internal/app/adapters/handlers/http/utils"
	"github.com/duchuongnguyen/dhcp2p/internal/app/adapters/handlers/http/validation"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/errors"
)

// HandlerFunc represents a standardized handler function signature
type HandlerFunc func(ctx context.Context, req interface{}) (interface{}, error)

// ServiceCall represents a service method call with standardized error handling
type ServiceCall struct {
	Handler http.ResponseWriter
	Request *http.Request
}

// ExecuteWithValidation executes a handler with standardized validation and error handling
func (sc *ServiceCall) ExecuteWithValidation(
	handler HandlerFunc,
	validator func(*http.Request) (interface{}, error),
) {
	// Validate request
	req, err := validator(sc.Request)
	if err != nil {
		utils.WriteDomainError(sc.Handler, err)
		return
	}

	// Execute handler
	result, err := handler(sc.Request.Context(), req)
	if err != nil {
		utils.WriteDomainError(sc.Handler, err)
		return
	}

	// Write success response
	utils.WriteSuccessResponse(sc.Handler, result)
}

// ExecuteServiceCall executes a service call with standardized error handling
func (sc *ServiceCall) ExecuteServiceCall(
	serviceFunc func(context.Context, interface{}) (interface{}, error),
	req interface{},
) {
	result, err := serviceFunc(sc.Request.Context(), req)
	if err != nil {
		utils.WriteDomainError(sc.Handler, err)
		return
	}

	utils.WriteSuccessResponse(sc.Handler, result)
}

// Common validation functions for different request types

// ValidateAuthRequest validates an authentication request
func ValidateAuthRequest(r *http.Request) (interface{}, error) {
	pubkeyResult := validation.ValidateHeader(r, "X-Pubkey", validation.DefaultValidationConfig())
	if pubkeyResult.Error != nil {
		return nil, pubkeyResult.Error
	}

	pubkeyValidation := validation.ValidateBase64Pubkey(pubkeyResult.Value)
	if pubkeyValidation.Error != nil {
		return nil, pubkeyValidation.Error
	}

	// Decode the validated pubkey
	pubkey, err := base64.StdEncoding.DecodeString(pubkeyValidation.Value)
	if err != nil {
		return nil, errors.ErrInvalidPubkey
	}

	return &AuthRequestData{
		Pubkey: pubkey,
	}, nil
}

// ValidateLeaseRequest validates a lease-related request
func ValidateLeaseRequest(r *http.Request) (interface{}, error) {
	peerIDResult := validation.ValidatePeerIDFromContext(r)
	if peerIDResult.Error != nil {
		return nil, peerIDResult.Error
	}

	return &LeaseRequestData{
		PeerID: peerIDResult.Value,
	}, nil
}

// ValidateTokenIDRequest validates a request that includes a token ID
func ValidateTokenIDRequest(r *http.Request) (interface{}, error) {
	peerIDResult := validation.ValidatePeerIDFromContext(r)
	if peerIDResult.Error != nil {
		return nil, peerIDResult.Error
	}

	tokenIDStr := r.URL.Query().Get("tokenID")
	tokenIDResult := validation.ValidateTokenID(tokenIDStr)
	if tokenIDResult.Error != nil {
		return nil, tokenIDResult.Error
	}

	tokenID, _ := strconv.ParseInt(tokenIDResult.Value, 10, 64)

	return &TokenIDRequestData{
		PeerID:  peerIDResult.Value,
		TokenID: tokenID,
	}, nil
}

// ValidatePeerIDParamRequest validates a request with peerID as URL parameter
func ValidatePeerIDParamRequest(r *http.Request) (interface{}, error) {
	peerIDResult := validation.ValidateURLParam(r, "peerID", validation.DefaultValidationConfig())
	if peerIDResult.Error != nil {
		return nil, peerIDResult.Error
	}

	return &PeerIDRequestData{
		PeerID: peerIDResult.Value,
	}, nil
}

// ValidateTokenIDParamRequest validates a request with tokenID as URL parameter
func ValidateTokenIDParamRequest(r *http.Request) (interface{}, error) {
	tokenIDStrResult := validation.ValidateURLParam(r, "tokenID", validation.DefaultValidationConfig())
	if tokenIDStrResult.Error != nil {
		return nil, tokenIDStrResult.Error
	}

	tokenIDResult := validation.ValidateTokenID(tokenIDStrResult.Value)
	if tokenIDResult.Error != nil {
		return nil, tokenIDResult.Error
	}

	tokenID, _ := strconv.ParseInt(tokenIDResult.Value, 10, 64)

	return &TokenIDRequestData{
		TokenID: tokenID,
	}, nil
}
