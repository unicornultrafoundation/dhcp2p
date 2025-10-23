package utils

import (
	"net/http/httptest"
	"testing"

	"github.com/duchuongnguyen/dhcp2p/internal/app/adapters/handlers/http/utils"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/errors"
	"github.com/stretchr/testify/assert"
)

func TestWriteDomainError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "missing headers error",
			err:            errors.ErrMissingHeaders,
			expectedStatus: 400,
			expectedBody:   `{"type":"validation_error","code":"MISSING_HEADERS","message":"Required headers are missing"}`,
		},
		{
			name:           "missing pubkey error",
			err:            errors.ErrMissingPubkey,
			expectedStatus: 400,
			expectedBody:   `{"type":"validation_error","code":"MISSING_PUBKEY","message":"Public key is required"}`,
		},
		{
			name:           "missing peer ID error",
			err:            errors.ErrMissingPeerID,
			expectedStatus: 400,
			expectedBody:   `{"type":"validation_error","code":"MISSING_PEER_ID","message":"Peer ID is required"}`,
		},
		{
			name:           "missing token ID error",
			err:            errors.ErrMissingTokenID,
			expectedStatus: 400,
			expectedBody:   `{"type":"validation_error","code":"MISSING_TOKEN_ID","message":"Token ID is required"}`,
		},
		{
			name:           "invalid token ID error",
			err:            errors.ErrInvalidTokenID,
			expectedStatus: 400,
			expectedBody:   `{"type":"validation_error","code":"INVALID_TOKEN_ID","message":"Invalid token ID format"}`,
		},
		{
			name:           "invalid pubkey error",
			err:            errors.ErrInvalidPubkey,
			expectedStatus: 400,
			expectedBody:   `{"type":"validation_error","code":"INVALID_PUBKEY","message":"Invalid public key format"}`,
		},
		{
			name:           "invalid signature error",
			err:            errors.ErrInvalidSignature,
			expectedStatus: 400,
			expectedBody:   `{"type":"validation_error","code":"INVALID_SIGNATURE","message":"Invalid signature format"}`,
		},
		{
			name:           "pubkey mismatch error",
			err:            errors.ErrPubkeyMismatch,
			expectedStatus: 401,
			expectedBody:   `{"type":"auth_error","code":"PUBKEY_MISMATCH","message":"Public key mismatch"}`,
		},
		{
			name:           "lease not found error",
			err:            errors.ErrLeaseNotFound,
			expectedStatus: 404,
			expectedBody:   `{"type":"not_found","code":"LEASE_NOT_FOUND","message":"Lease not found"}`,
		},
		{
			name:           "unknown error",
			err:            assert.AnError,
			expectedStatus: 500,
			expectedBody:   `{"type":"internal_error","code":"UNKNOWN_ERROR","message":"An unexpected error occurred"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			utils.WriteDomainError(w, tt.err)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestWriteErrorResponse(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		err            error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "bad request error",
			statusCode:     400,
			err:            errors.ErrMissingHeaders,
			expectedStatus: 400,
			expectedBody:   `{"type":"validation_error","code":"MISSING_HEADERS","message":"Required headers are missing"}`,
		},
		{
			name:           "internal server error",
			statusCode:     500,
			err:            assert.AnError,
			expectedStatus: 500,
			expectedBody:   `{"type":"internal_error","code":"UNKNOWN_ERROR","message":"An unexpected error occurred"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			utils.WriteErrorResponse(w, tt.err)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestWriteResponse(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		data           interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "success response with data",
			statusCode:     200,
			data:           map[string]string{"message": "success"},
			expectedStatus: 200,
			expectedBody:   `{"message":"success"}`,
		},
		{
			name:           "created response",
			statusCode:     201,
			data:           map[string]int{"id": 123},
			expectedStatus: 201,
			expectedBody:   `{"id":123}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			utils.WriteResponse(w, tt.statusCode, tt.data)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestWriteSuccessResponse(t *testing.T) {
	tests := []struct {
		name         string
		data         interface{}
		expectedBody string
	}{
		{
			name:         "success with data",
			data:         map[string]string{"status": "ok"},
			expectedBody: `{"data":{"status":"ok"}}`,
		},
		{
			name:         "success with array",
			data:         []string{"item1", "item2"},
			expectedBody: `{"data":["item1","item2"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			utils.WriteSuccessResponse(w, tt.data)

			assert.Equal(t, 200, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}
