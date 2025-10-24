//go:build integration && !unit

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/handlers/http"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/errors"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/models"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/infrastructure/config"
	"github.com/unicornultrafoundation/dhcp2p/tests/mocks"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
)

func TestHTTPHandlers_Integration_ErrorScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockLeaseRepo := mocks.NewMockLeaseRepository(ctrl)
	mockNonceRepo := mocks.NewMockNonceRepository(ctrl)
	mockAuthRepo := mocks.NewMockAuthRepository(ctrl)

	// Create handlers
	cfg := &config.AppConfig{
		Server: config.ServerConfig{
			Port: 8080,
			Host: "localhost",
		},
	}
	logger := zap.NewNop()

	leaseHandler := http.NewLeaseHandler(mockLeaseRepo, logger)
	authHandler := http.NewAuthHandler(mockAuthRepo, logger)
	healthHandler := http.NewHealthHandler(logger)

	// Setup router
	router := chi.NewRouter()
	leaseHandler.RegisterRoutes(router)
	authHandler.RegisterRoutes(router)
	healthHandler.RegisterRoutes(router)

	// Test server
	server := httptest.NewServer(router)
	defer server.Close()

	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		setupMocks     func()
		description    string
	}{
		{
			name:           "allocate_lease_database_error",
			method:         "POST",
			path:           "/api/v1/leases",
			body:           map[string]string{"peer_id": "test-peer"},
			expectedStatus: http.StatusInternalServerError,
			setupMocks: func() {
				mockLeaseRepo.EXPECT().
					GetLeaseByPeerID(gomock.Any(), "test-peer").
					Return(nil, assert.AnError)
			},
			description: "Database errors should return 500",
		},
		{
			name:           "get_lease_not_found",
			method:         "GET",
			path:           "/api/v1/leases/nonexistent-peer",
			body:           nil,
			expectedStatus: http.StatusNotFound,
			setupMocks: func() {
				mockLeaseRepo.EXPECT().
					GetLeaseByPeerID(gomock.Any(), "nonexistent-peer").
					Return(nil, errors.ErrLeaseNotFound)
			},
			description: "Non-existent lease should return 404",
		},
		{
			name:           "renew_lease_validation_error",
			method:         "PUT",
			path:           "/api/v1/leases/12345/renew",
			body:           nil,
			expectedStatus: http.StatusBadRequest,
			setupMocks: func() {
				// Mock validation error
				mockLeaseRepo.EXPECT().
					RenewLease(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.ErrInvalidTokenID)
			},
			description: "Invalid renewal request should return 400",
		},
		{
			name:           "release_lease_permission_denied",
			method:         "DELETE",
			path:           "/api/v1/leases/12345/release",
			body:           nil,
			expectedStatus: http.StatusForbidden,
			setupMocks: func() {
				mockLeaseRepo.EXPECT().
					ReleaseLease(gomock.Any(), int64(12345), gomock.Any()).
					Return(errors.ErrMissingDependencies)
			},
			description: "Unauthorized release should return 403",
		},
		{
			name:           "health_check_database_timeout",
			method:         "GET",
			path:           "/health",
			body:           nil,
			expectedStatus: http.StatusServiceUnavailable,
			setupMocks: func() {
				// Simulate database timeout
				mockLeaseRepo.EXPECT().
					HealthCheck(gomock.Any()).
					Return(context.DeadlineExceeded)
			},
			description: "Database timeout should return 503",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			var bodyBytes []byte
			var err error

			if tc.body != nil {
				bodyBytes, err = json.Marshal(tc.body)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(tc.method, tc.path, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code, tc.description)

			// Validate error response format
			if tc.expectedStatus >= 400 {
				var errorResp map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &errorResp)
				require.NoError(t, err)

				assert.Contains(t, errorResp, "error", "Error response should contain error field")
			}
		})
	}
}

func TestHTTPHandlers_Integration_TimeoutScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLeaseRepo := mocks.NewMockLeaseRepository(ctrl)
	logger := zap.NewNop()

	leaseHandler := http.NewLeaseHandler(mockLeaseRepo, logger)
	router := chi.NewRouter()
	leaseHandler.RegisterRoutes(router)

	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("slow_database_response", func(t *testing.T) {
		// Setup mock with delay
		mockLeaseRepo.EXPECT().
			GetLeaseByPeerID(gomock.Any(), "slow-peer").
			DoAndReturn(func(ctx context.Context, peerID string) (*models.Lease, error) {
				time.Sleep(2 * time.Second) // Simulate slow response
				return nil, context.DeadlineExceeded
			})

		req := httptest.NewRequest("GET", "/api/v1/leases/slow-peer", nil)
		ctx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
		req = req.WithContext(ctx)
		defer cancel()

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusRequestTimeout, w.Code)
	})
}

func TestHTTPHandlers_Integration_ConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLeaseRepo := mocks.NewMockLeaseRepository(ctrl)
	logger := zap.NewNop()

	leaseHandler := http.NewLeaseHandler(mockLeaseRepo, logger)
	router := chi.NewRouter()
	leaseHandler.RegisterRoutes(router)

	server := httptest.NewServer(router)
	defer server.Close()

	lease := &models.Lease{
		TokenID:   12345,
		PeerID:    "concurrent-peer",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		Ttl:       3600,
	}

	// Setup mock to handle concurrent requests
	mockLeaseRepo.EXPECT().
		GetLeaseByPeerID(gomock.Any(), "concurrent-peer").
		Return(lease, nil).
		AnyTimes()

	const numRequests = 50
	const numGoroutines = 10

	var wg sync.WaitGroup
	var mu sync.Mutex
	responses := make([]int, numRequests)

	startTime := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numRequests/numGoroutines; j++ {
				reqNum := goroutineID*(numRequests/numGoroutines) + j

				req := httptest.NewRequest("GET", "/api/v1/leases/concurrent-peer", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				mu.Lock()
				responses[reqNum] = w.Code
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// Validate all requests succeeded
	for i, statusCode := range responses {
		assert.Equal(t, http.StatusOK, statusCode, "Request %d should succeed", i)
	}

	// Performance assertion - should handle concurrent requests efficiently
	assert.Less(t, duration, 5*time.Second, "Concurrent requests should complete within 5 seconds")

	t.Logf("Processed %d concurrent requests in %v (%.2f req/sec)",
		numRequests, duration, float64(numRequests)/duration.Seconds())
}

func TestHTTPHandlers_Integration_MalformedRequests(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLeaseRepo := mocks.NewMockLeaseRepository(ctrl)
	logger := zap.NewNop()

	leaseHandler := http.NewLeaseHandler(mockLeaseRepo, logger)
	router := chi.NewRouter()
	leaseHandler.RegisterRoutes(router)

	server := httptest.NewServer(router)
	defer server.Close()

	testCases := []struct {
		name           string
		method         string
		path           string
		body           string
		contentType    string
		expectedStatus int
		description    string
	}{
		{
			name:           "invalid_json",
			method:         "POST",
			path:           "/api/v1/leases",
			body:           `{"peer_id":}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			description:    "Invalid JSON should return 400",
		},
		{
			name:           "missing_fields",
			method:         "POST",
			path:           "/api/v1/leases",
			body:           `{}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			description:    "Missing required fields should return 400",
		},
		{
			name:           "wrong_content_type",
			method:         "POST",
			path:           "/api/v1/leases",
			body:           `{"peer_id": "test"}`,
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
			description:    "Wrong content type should return 400",
		},
		{
			name:           "empty_body",
			method:         "POST",
			path:           "/api/v1/leases",
			body:           "",
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			description:    "Empty body should return 400",
		},
		{
			name:           "too_large_body",
			method:         "POST",
			path:           "/api/v1/leases",
			body:           string(make([]byte, 1024*1024)), // 1MB
			contentType:    "application/json",
			expectedStatus: http.StatusRequestEntityTooLarge,
			description:    "Too large body should return 413",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code, tc.description)
		})
	}
}
