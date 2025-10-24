package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	handlers "github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/handlers/http"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/handlers/http/keys"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/models"
	"github.com/unicornultrafoundation/dhcp2p/tests/mocks"
	"github.com/golang/mock/gomock"
)

// Helper function to create a request with chi URL parameters
func createRequestWithURLParams(method, url string, params map[string]string) *http.Request {
	req := httptest.NewRequest(method, url, nil)

	// Create chi context with URL parameters
	rctx := chi.NewRouteContext()
	for key, value := range params {
		rctx.URLParams.Add(key, value)
	}

	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	return req.WithContext(ctx)
}

func TestLeaseHandler_AllocateIP(t *testing.T) {
	tests := []struct {
		name           string
		peerID         string
		mockSetup      func(*gomock.Controller, *mocks.MockLeaseService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "successful allocation",
			peerID: "peer123",
			mockSetup: func(ctrl *gomock.Controller, mockService *mocks.MockLeaseService) {
				mockService.EXPECT().AllocateIP(gomock.Any(), "peer123").Return(&models.Lease{
					TokenID:   167772161,
					PeerID:    "peer123",
					CreatedAt: time.Now(),
					ExpiresAt: time.Now().Add(time.Hour),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "missing peer ID in context",
			peerID:         "",
			mockSetup:      func(ctrl *gomock.Controller, mockService *mocks.MockLeaseService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockLeaseService(ctrl)
			tt.mockSetup(ctrl, mockService)

			handler := handlers.NewLeaseHandler(mockService)

			// Create request with proper context (set by auth middleware)
			req := httptest.NewRequest("POST", "/allocate-ip", nil)
			if tt.peerID != "" {
				req = req.WithContext(context.WithValue(req.Context(), keys.PeerIDContextKey, tt.peerID))
			}
			w := httptest.NewRecorder()

			handler.AllocateIP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response struct {
					Data models.Lease `json:"data"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.peerID, response.Data.PeerID)
			}
		})
	}
}

func TestLeaseHandler_GetLeaseByPeerID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockLeaseService(ctrl)
	handler := handlers.NewLeaseHandler(mockService)

	expectedLease := &models.Lease{
		TokenID:   167772161,
		PeerID:    "peer123",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	mockService.EXPECT().GetLeaseByPeerID(gomock.Any(), "peer123").Return(expectedLease, nil)

	// Create request with chi URL parameters
	req := createRequestWithURLParams("GET", "/lease/peer-id/peer123", map[string]string{"peerID": "peer123"})
	w := httptest.NewRecorder()

	handler.GetLeaseByPeerID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data models.Lease `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedLease.TokenID, response.Data.TokenID)
	assert.Equal(t, expectedLease.PeerID, response.Data.PeerID)
}

func TestLeaseHandler_GetLeaseByTokenID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockLeaseService(ctrl)
	handler := handlers.NewLeaseHandler(mockService)

	expectedLease := &models.Lease{
		TokenID:   167772161,
		PeerID:    "peer123",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	mockService.EXPECT().GetLeaseByTokenID(gomock.Any(), int64(167772161)).Return(expectedLease, nil)

	// Create request with chi URL parameters
	req := createRequestWithURLParams("GET", "/lease/token-id/167772161", map[string]string{"tokenID": "167772161"})
	w := httptest.NewRecorder()

	handler.GetLeaseByTokenID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data models.Lease `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedLease.TokenID, response.Data.TokenID)
	assert.Equal(t, expectedLease.PeerID, response.Data.PeerID)
}

func TestLeaseHandler_RenewLease(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockLeaseService(ctrl)
	handler := handlers.NewLeaseHandler(mockService)

	expectedLease := &models.Lease{
		TokenID:   167772161,
		PeerID:    "peer123",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	mockService.EXPECT().RenewLease(gomock.Any(), int64(167772161), "peer123").Return(expectedLease, nil)

	// Create request with proper context and query parameter
	req := httptest.NewRequest("POST", "/renew-lease?tokenID=167772161", nil)
	req = req.WithContext(context.WithValue(req.Context(), keys.PeerIDContextKey, "peer123"))
	w := httptest.NewRecorder()

	handler.RenewLease(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data models.Lease `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedLease.TokenID, response.Data.TokenID)
	assert.Equal(t, expectedLease.PeerID, response.Data.PeerID)
}

func TestLeaseHandler_ReleaseLease(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockLeaseService(ctrl)
	handler := handlers.NewLeaseHandler(mockService)

	mockService.EXPECT().ReleaseLease(gomock.Any(), int64(167772161), "peer123").Return(nil)

	// Create request with proper context and query parameter
	req := httptest.NewRequest("POST", "/release-lease?tokenID=167772161", nil)
	req = req.WithContext(context.WithValue(req.Context(), keys.PeerIDContextKey, "peer123"))
	w := httptest.NewRecorder()

	handler.ReleaseLease(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data map[string]string `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Data["status"])
}

// Test validation error cases
func TestLeaseHandler_ValidationErrors(t *testing.T) {
	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		expectedStatus int
	}{
		{
			name: "missing peer ID in context for AllocateIP",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("POST", "/allocate-ip", nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing peer ID parameter for GetLeaseByPeerID",
			setupRequest: func() *http.Request {
				return createRequestWithURLParams("GET", "/lease/peer-id/", map[string]string{"peerID": ""})
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing token ID parameter for GetLeaseByTokenID",
			setupRequest: func() *http.Request {
				return createRequestWithURLParams("GET", "/lease/token-id/", map[string]string{"tokenID": ""})
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing token ID query parameter for RenewLease",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("POST", "/renew-lease", nil)
				return req.WithContext(context.WithValue(req.Context(), keys.PeerIDContextKey, "peer123"))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing token ID query parameter for ReleaseLease",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("POST", "/release-lease", nil)
				return req.WithContext(context.WithValue(req.Context(), keys.PeerIDContextKey, "peer123"))
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockLeaseService(ctrl)
			handler := handlers.NewLeaseHandler(mockService)

			req := tt.setupRequest()
			w := httptest.NewRecorder()

			// Determine which handler to test based on the URL
			switch req.URL.Path {
			case "/allocate-ip":
				handler.AllocateIP(w, req)
			case "/lease/peer-id/":
				handler.GetLeaseByPeerID(w, req)
			case "/lease/token-id/":
				handler.GetLeaseByTokenID(w, req)
			case "/renew-lease":
				handler.RenewLease(w, req)
			case "/release-lease":
				handler.ReleaseLease(w, req)
			}

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
