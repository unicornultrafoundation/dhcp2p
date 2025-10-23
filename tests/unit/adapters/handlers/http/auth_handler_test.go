package http

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	handlers "github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/handlers/http"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/errors"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/models"
	"github.com/unicornultrafoundation/dhcp2p/tests/mocks"
	"go.uber.org/mock/gomock"
)

// generateValidPubkey creates a valid-sized pubkey for testing (64 bytes)
func generateValidPubkey(t *testing.T) []byte {
	return make([]byte, 64) // 64 bytes is within the valid range (32-1024)
}

func TestAuthHandler_RequestAuth(t *testing.T) {
	tests := []struct {
		name             string
		headers          map[string]string
		mockSetup        func(*gomock.Controller, *mocks.MockAuthService)
		expectedStatus   int
		expectedError    bool
		expectedResponse *handlers.AuthResponse
	}{
		{
			name: "successful auth request",
			headers: map[string]string{
				"X-Pubkey": base64.StdEncoding.EncodeToString([]byte("valid-pubkey-data")),
			},
			mockSetup: func(ctrl *gomock.Controller, mockService *mocks.MockAuthService) {
				mockService.EXPECT().RequestAuth(gomock.Any(), &models.AuthRequest{
					Pubkey: []byte("valid-pubkey-data"),
				}).Return(&models.AuthResponse{
					NonceID: "test-nonce-id",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
			expectedResponse: &handlers.AuthResponse{
				Pubkey: base64.StdEncoding.EncodeToString([]byte("valid-pubkey-data")),
				Nonce:  "test-nonce-id",
			},
		},
		{
			name:    "missing pubkey header",
			headers: map[string]string{},
			mockSetup: func(ctrl *gomock.Controller, mockService *mocks.MockAuthService) {
				// No expectations for missing header
			},
			expectedStatus:   http.StatusBadRequest,
			expectedError:    true,
			expectedResponse: nil,
		},
		{
			name: "invalid base64 pubkey",
			headers: map[string]string{
				"X-Pubkey": "invalid-base64!@#",
			},
			mockSetup: func(ctrl *gomock.Controller, mockService *mocks.MockAuthService) {
				// No expectations for invalid base64
			},
			expectedStatus:   http.StatusBadRequest,
			expectedError:    true,
			expectedResponse: nil,
		},
		{
			name: "empty pubkey",
			headers: map[string]string{
				"X-Pubkey": "",
			},
			mockSetup: func(ctrl *gomock.Controller, mockService *mocks.MockAuthService) {
				// No expectations for empty pubkey
			},
			expectedStatus:   http.StatusBadRequest,
			expectedError:    true,
			expectedResponse: nil,
		},
		{
			name: "pubkey too short",
			headers: map[string]string{
				"X-Pubkey": base64.StdEncoding.EncodeToString([]byte("short")),
			},
			mockSetup: func(ctrl *gomock.Controller, mockService *mocks.MockAuthService) {
				// No expectations for too short pubkey
			},
			expectedStatus:   http.StatusBadRequest,
			expectedError:    true,
			expectedResponse: nil,
		},
		{
			name: "pubkey too long",
			headers: map[string]string{
				"X-Pubkey": base64.StdEncoding.EncodeToString(make([]byte, 2000)), // Too long
			},
			mockSetup: func(ctrl *gomock.Controller, mockService *mocks.MockAuthService) {
				// No expectations for too long pubkey
			},
			expectedStatus:   http.StatusBadRequest,
			expectedError:    true,
			expectedResponse: nil,
		},
		{
			name: "auth service error",
			headers: map[string]string{
				"X-Pubkey": base64.StdEncoding.EncodeToString([]byte("valid-pubkey-data")),
			},
			mockSetup: func(ctrl *gomock.Controller, mockService *mocks.MockAuthService) {
				mockService.EXPECT().RequestAuth(gomock.Any(), &models.AuthRequest{
					Pubkey: []byte("valid-pubkey-data"),
				}).Return(nil, errors.ErrMissingPeerID)
			},
			expectedStatus:   http.StatusBadRequest,
			expectedError:    true,
			expectedResponse: nil,
		},
		{
			name: "whitespace in pubkey header",
			headers: map[string]string{
				"X-Pubkey": " " + base64.StdEncoding.EncodeToString([]byte("valid-pubkey-data")) + " ",
			},
			mockSetup: func(ctrl *gomock.Controller, mockService *mocks.MockAuthService) {
				mockService.EXPECT().RequestAuth(gomock.Any(), &models.AuthRequest{
					Pubkey: []byte("valid-pubkey-data"),
				}).Return(&models.AuthResponse{
					NonceID: "test-nonce-id",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
			expectedResponse: &handlers.AuthResponse{
				Pubkey: base64.StdEncoding.EncodeToString([]byte("valid-pubkey-data")),
				Nonce:  "test-nonce-id",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockAuthService(ctrl)
			tt.mockSetup(ctrl, mockService)

			handler := handlers.NewAuthHandler(mockService)

			req := httptest.NewRequest("POST", "/request-auth", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			w := httptest.NewRecorder()

			handler.RequestAuth(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response struct {
					Data handlers.AuthResponse `json:"data"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResponse.Pubkey, response.Data.Pubkey)
				assert.Equal(t, tt.expectedResponse.Nonce, response.Data.Nonce)
			} else {
				// Check that response contains error information
				assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}

func TestAuthHandler_RequestAuth_EdgeCases(t *testing.T) {
	t.Run("context cancellation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mocks.NewMockAuthService(ctrl)
		handler := handlers.NewAuthHandler(mockService)

		req := httptest.NewRequest("POST", "/request-auth", nil)
		req.Header.Set("X-Pubkey", base64.StdEncoding.EncodeToString([]byte("valid-pubkey-data")))

		// Create a cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		req = req.WithContext(ctx)

		mockService.EXPECT().RequestAuth(gomock.Any(), gomock.Any()).Return(nil, context.Canceled)

		w := httptest.NewRecorder()
		handler.RequestAuth(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("very large pubkey", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mocks.NewMockAuthService(ctrl)
		handler := handlers.NewAuthHandler(mockService)

		// Create a very large pubkey (but still within limits)
		largePubkey := make([]byte, 1000)
		req := httptest.NewRequest("POST", "/request-auth", nil)
		req.Header.Set("X-Pubkey", base64.StdEncoding.EncodeToString(largePubkey))

		mockService.EXPECT().RequestAuth(gomock.Any(), gomock.Any()).Return(&models.AuthResponse{
			NonceID: "test-nonce-id",
		}, nil)

		w := httptest.NewRecorder()
		handler.RequestAuth(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Data handlers.AuthResponse `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, base64.StdEncoding.EncodeToString(largePubkey), response.Data.Pubkey)
		assert.Equal(t, "test-nonce-id", response.Data.Nonce)
	})

	t.Run("special characters in pubkey", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mocks.NewMockAuthService(ctrl)
		handler := handlers.NewAuthHandler(mockService)

		// Create pubkey with special characters
		specialPubkey := []byte("valid-pubkey-with-special-chars!@#$%^&*()")
		req := httptest.NewRequest("POST", "/request-auth", nil)
		req.Header.Set("X-Pubkey", base64.StdEncoding.EncodeToString(specialPubkey))

		mockService.EXPECT().RequestAuth(gomock.Any(), gomock.Any()).Return(&models.AuthResponse{
			NonceID: "test-nonce-id",
		}, nil)

		w := httptest.NewRecorder()
		handler.RequestAuth(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Data handlers.AuthResponse `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, base64.StdEncoding.EncodeToString(specialPubkey), response.Data.Pubkey)
		assert.Equal(t, "test-nonce-id", response.Data.Nonce)
	})

	t.Run("multiple pubkey headers", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mocks.NewMockAuthService(ctrl)
		handler := handlers.NewAuthHandler(mockService)

		req := httptest.NewRequest("POST", "/request-auth", nil)
		req.Header.Set("X-Pubkey", base64.StdEncoding.EncodeToString([]byte("valid-pubkey-data")))
		req.Header.Add("X-Pubkey", base64.StdEncoding.EncodeToString([]byte("second-pubkey"))) // This should be ignored

		mockService.EXPECT().RequestAuth(gomock.Any(), &models.AuthRequest{
			Pubkey: []byte("valid-pubkey-data"), // Should use the first value
		}).Return(&models.AuthResponse{
			NonceID: "test-nonce-id",
		}, nil)

		w := httptest.NewRecorder()
		handler.RequestAuth(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Data handlers.AuthResponse `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("valid-pubkey-data")), response.Data.Pubkey)
	})

	t.Run("case insensitive header", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mocks.NewMockAuthService(ctrl)
		handler := handlers.NewAuthHandler(mockService)

		req := httptest.NewRequest("POST", "/request-auth", nil)
		req.Header.Set("x-pubkey", base64.StdEncoding.EncodeToString([]byte("valid-pubkey-data"))) // lowercase

		mockService.EXPECT().RequestAuth(gomock.Any(), &models.AuthRequest{
			Pubkey: []byte("valid-pubkey-data"),
		}).Return(&models.AuthResponse{
			NonceID: "test-nonce-id",
		}, nil)

		w := httptest.NewRecorder()
		handler.RequestAuth(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Data handlers.AuthResponse `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("valid-pubkey-data")), response.Data.Pubkey)
	})
}

func TestAuthHandler_RequestAuth_Concurrent(t *testing.T) {
	t.Run("concurrent requests", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mocks.NewMockAuthService(ctrl)
		handler := handlers.NewAuthHandler(mockService)

		const numRequests = 10
		results := make(chan struct {
			status int
			err    error
		}, numRequests)

		// Setup expectations for concurrent calls
		for i := 0; i < numRequests; i++ {
			pubkey := []byte("valid-pubkey-data")
			mockService.EXPECT().RequestAuth(gomock.Any(), &models.AuthRequest{
				Pubkey: pubkey,
			}).Return(&models.AuthResponse{
				NonceID: "test-nonce-id",
			}, nil)
		}

		// Start multiple goroutines
		for i := 0; i < numRequests; i++ {
			go func() {
				req := httptest.NewRequest("POST", "/request-auth", nil)
				req.Header.Set("X-Pubkey", base64.StdEncoding.EncodeToString([]byte("valid-pubkey-data")))
				w := httptest.NewRecorder()

				handler.RequestAuth(w, req)

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
}
