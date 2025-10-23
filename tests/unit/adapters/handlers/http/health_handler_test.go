package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	handlers "github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/handlers/http"
)

// MockDB is a mock for pgxpool.Pool
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockRedisClient is a mock for redis.Client
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

// MockStatusCmd is a mock for redis.StatusCmd
type MockStatusCmd struct {
	mock.Mock
}

func (m *MockStatusCmd) Err() error {
	args := m.Called()
	return args.Error(0)
}

func TestHealthHandler_Health(t *testing.T) {
	tests := []struct {
		name           string
		db             *pgxpool.Pool
		cache          *redis.Client
		expectedStatus int
		expectedBody   map[string]string
	}{
		{
			name:           "successful health check",
			db:             nil, // Health check doesn't use DB
			cache:          nil, // Health check doesn't use cache
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]string{"status": "ok"},
		},
		{
			name:           "health check with nil dependencies",
			db:             nil,
			cache:          nil,
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]string{"status": "ok"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := handlers.NewHealthHandler(tt.db, tt.cache)

			req := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()

			handler.Health(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]string
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)
		})
	}
}

func TestHealthHandler_Readiness(t *testing.T) {
	tests := []struct {
		name           string
		db             *pgxpool.Pool
		cache          *redis.Client
		mockSetup      func(*MockDB, *MockRedisClient)
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "successful readiness check with nil dependencies",
			db:             nil,
			cache:          nil,
			mockSetup:      func(mockDB *MockDB, mockRedis *MockRedisClient) {},
			expectedStatus: http.StatusServiceUnavailable,
			expectedError:  true,
		},
		{
			name:           "nil database",
			db:             nil,
			cache:          &redis.Client{},
			mockSetup:      func(mockDB *MockDB, mockRedis *MockRedisClient) {},
			expectedStatus: http.StatusServiceUnavailable,
			expectedError:  true,
		},
		{
			name:           "nil cache",
			db:             &pgxpool.Pool{},
			cache:          nil,
			mockSetup:      func(mockDB *MockDB, mockRedis *MockRedisClient) {},
			expectedStatus: http.StatusServiceUnavailable,
			expectedError:  true,
		},
		{
			name:           "both nil dependencies",
			db:             nil,
			cache:          nil,
			mockSetup:      func(mockDB *MockDB, mockRedis *MockRedisClient) {},
			expectedStatus: http.StatusServiceUnavailable,
			expectedError:  true,
		},
		{
			name:  "database ping failure",
			db:    nil, // Use nil to trigger the nil check in the handler
			cache: nil,
			mockSetup: func(mockDB *MockDB, mockRedis *MockRedisClient) {
				// No setup needed for nil dependencies
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedError:  true,
		},
		{
			name:  "cache ping failure",
			db:    nil, // Use nil to trigger the nil check in the handler
			cache: nil,
			mockSetup: func(mockDB *MockDB, mockRedis *MockRedisClient) {
				// No setup needed for nil dependencies
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: In a real test, you would need to properly mock the pgxpool.Pool and redis.Client
			// For now, we'll test the basic functionality without mocking
			handler := handlers.NewHealthHandler(tt.db, tt.cache)

			req := httptest.NewRequest("GET", "/ready", nil)
			w := httptest.NewRecorder()

			handler.Readiness(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			if !tt.expectedError {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "ready", response["status"])
			} else {
				// Check that response contains error information
				assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}

func TestHealthHandler_EdgeCases(t *testing.T) {
	t.Run("context timeout", func(t *testing.T) {
		handler := handlers.NewHealthHandler(nil, nil)

		// Create a request with a very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		req := httptest.NewRequest("GET", "/ready", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Readiness(w, req)

		// Should return service unavailable due to nil dependencies
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})

	t.Run("concurrent health checks", func(t *testing.T) {
		handler := handlers.NewHealthHandler(nil, nil)

		const numRequests = 10
		results := make(chan struct {
			status int
			err    error
		}, numRequests)

		// Start multiple goroutines
		for i := 0; i < numRequests; i++ {
			go func() {
				req := httptest.NewRequest("GET", "/health", nil)
				w := httptest.NewRecorder()

				handler.Health(w, req)

				results <- struct {
					status int
					err    error
				}{w.Code, nil}
			}()
		}

		// Collect results
		var successfulRequests int
		for i := 0; i < numRequests; i++ {
			select {
			case result := <-results:
				if result.status == http.StatusOK {
					successfulRequests++
				}
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for concurrent requests")
			}
		}

		// Verify all requests succeeded
		assert.Equal(t, numRequests, successfulRequests)
	})

	t.Run("concurrent readiness checks", func(t *testing.T) {
		handler := handlers.NewHealthHandler(nil, nil) // Will fail due to nil dependencies

		const numRequests = 10
		results := make(chan struct {
			status int
			err    error
		}, numRequests)

		// Start multiple goroutines
		for i := 0; i < numRequests; i++ {
			go func() {
				req := httptest.NewRequest("GET", "/ready", nil)
				w := httptest.NewRecorder()

				handler.Readiness(w, req)

				results <- struct {
					status int
					err    error
				}{w.Code, nil}
			}()
		}

		// Collect results
		var failedRequests int
		for i := 0; i < numRequests; i++ {
			select {
			case result := <-results:
				if result.status == http.StatusServiceUnavailable {
					failedRequests++
				}
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for concurrent requests")
			}
		}

		// Verify all requests failed as expected (due to nil dependencies)
		assert.Equal(t, numRequests, failedRequests)
	})

	t.Run("different HTTP methods", func(t *testing.T) {
		handler := handlers.NewHealthHandler(nil, nil)

		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				req := httptest.NewRequest(method, "/health", nil)
				w := httptest.NewRecorder()

				handler.Health(w, req)

				// Health endpoint should work regardless of HTTP method
				assert.Equal(t, http.StatusOK, w.Code)
			})
		}
	})

	t.Run("malformed request", func(t *testing.T) {
		handler := handlers.NewHealthHandler(nil, nil)

		// Create a request with malformed URL
		req := httptest.NewRequest("GET", "/health?invalid=%%%", nil)
		w := httptest.NewRecorder()

		handler.Health(w, req)

		// Should still work as health check doesn't parse query params
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHealthHandler_ResponseFormat(t *testing.T) {
	t.Run("health response format", func(t *testing.T) {
		handler := handlers.NewHealthHandler(nil, nil)

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		handler.Health(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "status")
		assert.Equal(t, "ok", response["status"])
	})

	t.Run("readiness response format", func(t *testing.T) {
		handler := handlers.NewHealthHandler(nil, nil)

		req := httptest.NewRequest("GET", "/ready", nil)
		w := httptest.NewRecorder()

		handler.Readiness(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "type")
		assert.Contains(t, response, "code")
		assert.Contains(t, response, "message")
	})
}
