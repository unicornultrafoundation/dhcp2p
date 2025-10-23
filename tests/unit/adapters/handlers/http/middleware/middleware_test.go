package middleware

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/duchuongnguyen/dhcp2p/internal/app/adapters/handlers/http/keys"
	"github.com/duchuongnguyen/dhcp2p/internal/app/adapters/handlers/http/middleware"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/errors"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/models"
	"github.com/duchuongnguyen/dhcp2p/tests/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWithAuth(t *testing.T) {
	tests := []struct {
		name           string
		headers        map[string]string
		mockSetup      func(*gomock.Controller, *mocks.MockAuthService)
		expectedStatus int
		expectedError  bool
		expectedPeerID string
	}{
		{
			name: "successful authentication",
			headers: map[string]string{
				"X-Pubkey":    base64.StdEncoding.EncodeToString(make([]byte, 32)), // 32 bytes for valid pubkey
				"X-Nonce":     "12345678-1234-1234-1234-123456789012",              // Valid UUID
				"X-Signature": base64.StdEncoding.EncodeToString(make([]byte, 64)), // 64 bytes for valid signature
			},
			mockSetup: func(ctrl *gomock.Controller, mockService *mocks.MockAuthService) {
				mockService.EXPECT().VerifyAuth(gomock.Any(), &models.AuthVerifyRequest{
					Pubkey:    make([]byte, 32),
					NonceID:   "12345678-1234-1234-1234-123456789012",
					Signature: make([]byte, 64),
				}).Return(&models.AuthVerifyResponse{
					Pubkey: make([]byte, 32),
				}, nil)
			},
			expectedStatus: http.StatusBadRequest, // Will fail at GetPeerIDFromPubkey
			expectedError:  true,
			expectedPeerID: "", // Will be set by GetPeerIDFromPubkey
		},
		{
			name: "missing pubkey header",
			headers: map[string]string{
				"X-Nonce":     "12345678-1234-1234-1234-123456789012",
				"X-Signature": base64.StdEncoding.EncodeToString([]byte("valid-signature")),
			},
			mockSetup: func(ctrl *gomock.Controller, mockService *mocks.MockAuthService) {
				// No expectations for missing header
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			expectedPeerID: "",
		},
		{
			name: "missing nonce header",
			headers: map[string]string{
				"X-Pubkey":    base64.StdEncoding.EncodeToString(make([]byte, 32)),
				"X-Signature": base64.StdEncoding.EncodeToString(make([]byte, 64)),
			},
			mockSetup: func(ctrl *gomock.Controller, mockService *mocks.MockAuthService) {
				// No expectations for missing header
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			expectedPeerID: "",
		},
		{
			name: "missing signature header",
			headers: map[string]string{
				"X-Pubkey": base64.StdEncoding.EncodeToString(make([]byte, 32)),
				"X-Nonce":  "12345678-1234-1234-1234-123456789012",
			},
			mockSetup: func(ctrl *gomock.Controller, mockService *mocks.MockAuthService) {
				// No expectations for missing header
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			expectedPeerID: "",
		},
		{
			name: "invalid base64 pubkey",
			headers: map[string]string{
				"X-Pubkey":    "invalid-base64!@#",
				"X-Nonce":     "12345678-1234-1234-1234-123456789012",
				"X-Signature": base64.StdEncoding.EncodeToString(make([]byte, 64)),
			},
			mockSetup: func(ctrl *gomock.Controller, mockService *mocks.MockAuthService) {
				// No expectations for invalid base64
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			expectedPeerID: "",
		},
		{
			name: "invalid base64 signature",
			headers: map[string]string{
				"X-Pubkey":    base64.StdEncoding.EncodeToString(make([]byte, 32)),
				"X-Nonce":     "12345678-1234-1234-1234-123456789012",
				"X-Signature": "invalid-base64!@#",
			},
			mockSetup: func(ctrl *gomock.Controller, mockService *mocks.MockAuthService) {
				// No expectations for invalid base64
			},
			expectedStatus: http.StatusBadRequest, // Validation errors return BadRequest
			expectedError:  true,
			expectedPeerID: "",
		},
		{
			name: "pubkey too short",
			headers: map[string]string{
				"X-Pubkey":    base64.StdEncoding.EncodeToString(make([]byte, 8)), // Too short
				"X-Nonce":     "12345678-1234-1234-1234-123456789012",
				"X-Signature": base64.StdEncoding.EncodeToString(make([]byte, 64)),
			},
			mockSetup: func(ctrl *gomock.Controller, mockService *mocks.MockAuthService) {
				// No expectations for too short pubkey
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			expectedPeerID: "",
		},
		{
			name: "signature too short",
			headers: map[string]string{
				"X-Pubkey":    base64.StdEncoding.EncodeToString(make([]byte, 32)),
				"X-Nonce":     "12345678-1234-1234-1234-123456789012",
				"X-Signature": base64.StdEncoding.EncodeToString(make([]byte, 16)), // Too short
			},
			mockSetup: func(ctrl *gomock.Controller, mockService *mocks.MockAuthService) {
				// No expectations for too short signature
			},
			expectedStatus: http.StatusBadRequest, // Validation errors return BadRequest
			expectedError:  true,
			expectedPeerID: "",
		},
		{
			name: "auth service verification failure",
			headers: map[string]string{
				"X-Pubkey":    base64.StdEncoding.EncodeToString(make([]byte, 32)),
				"X-Nonce":     "12345678-1234-1234-1234-123456789012",
				"X-Signature": base64.StdEncoding.EncodeToString(make([]byte, 64)),
			},
			mockSetup: func(ctrl *gomock.Controller, mockService *mocks.MockAuthService) {
				mockService.EXPECT().VerifyAuth(gomock.Any(), gomock.Any()).Return(nil, errors.ErrSignatureVerification)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
			expectedPeerID: "",
		},
		{
			name: "pubkey mismatch",
			headers: map[string]string{
				"X-Pubkey":    base64.StdEncoding.EncodeToString(make([]byte, 32)),
				"X-Nonce":     "12345678-1234-1234-1234-123456789012",
				"X-Signature": base64.StdEncoding.EncodeToString(make([]byte, 64)),
			},
			mockSetup: func(ctrl *gomock.Controller, mockService *mocks.MockAuthService) {
				mockService.EXPECT().VerifyAuth(gomock.Any(), gomock.Any()).Return(&models.AuthVerifyResponse{
					Pubkey: make([]byte, 32), // Different pubkey (all zeros)
				}, nil)
			},
			expectedStatus: http.StatusBadRequest, // Will fail at GetPeerIDFromPubkey
			expectedError:  true,
			expectedPeerID: "",
		},
		{
			name: "whitespace in headers",
			headers: map[string]string{
				"X-Pubkey":    " " + base64.StdEncoding.EncodeToString(make([]byte, 32)) + " ",
				"X-Nonce":     " 12345678-1234-1234-1234-123456789012 ",
				"X-Signature": " " + base64.StdEncoding.EncodeToString(make([]byte, 64)) + " ",
			},
			mockSetup: func(ctrl *gomock.Controller, mockService *mocks.MockAuthService) {
				mockService.EXPECT().VerifyAuth(gomock.Any(), &models.AuthVerifyRequest{
					Pubkey:    make([]byte, 32),
					NonceID:   "12345678-1234-1234-1234-123456789012",
					Signature: make([]byte, 64),
				}).Return(&models.AuthVerifyResponse{
					Pubkey: make([]byte, 32),
				}, nil)
			},
			expectedStatus: http.StatusBadRequest, // Will fail at GetPeerIDFromPubkey
			expectedError:  true,
			expectedPeerID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockAuthService(ctrl)
			tt.mockSetup(ctrl, mockService)

			// Create a test handler that checks if peerID is set in context
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				peerID := r.Context().Value(keys.PeerIDContextKey)
				if peerID != nil {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("authenticated"))
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("no peer ID"))
				}
			})

			// Apply auth middleware
			authMiddleware := middleware.WithAuth(mockService)
			handler := authMiddleware(testHandler)

			req := httptest.NewRequest("POST", "/test", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				assert.Equal(t, "authenticated", w.Body.String())
			} else {
				assert.NotEqual(t, "authenticated", w.Body.String())
			}
		})
	}
}

func TestSecurityMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		request        func() *http.Request
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "valid request",
			request: func() *http.Request {
				return httptest.NewRequest("GET", "/test", nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "request with suspicious header",
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Custom", "<script>alert('xss')</script>")
				return req
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "request with javascript header",
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Custom", "javascript:alert('xss')")
				return req
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "request with onload header",
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Custom", "onload=alert('xss')")
				return req
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "request with onerror header",
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Custom", "onerror=alert('xss')")
				return req
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "request with very long header",
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Custom", string(make([]byte, 9000))) // Too long
				return req
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "request with very long header name",
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set(string(make([]byte, 300)), "value") // Too long name
				return req
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "request with very long URL",
			request: func() *http.Request {
				// Create a URL with query parameters that exceed the limit
				longQuery := "param=" + strings.Repeat("a", 10000) // Much longer than 8192
				req := httptest.NewRequest("GET", "/test?"+longQuery, nil)
				return req
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
			})

			// Apply security middleware
			securityMiddleware := middleware.SecurityMiddleware()
			handler := securityMiddleware(testHandler)

			req := tt.request()
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				assert.Equal(t, "ok", w.Body.String())
			}
		})
	}
}

func TestRequestSizeMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		contentLength  int64
		maxSize        int64
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "request within size limit",
			contentLength:  1024,
			maxSize:        2048,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "request at size limit",
			contentLength:  1024,
			maxSize:        1024,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "request exceeds size limit",
			contentLength:  2048,
			maxSize:        1024,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "zero content length",
			contentLength:  0,
			maxSize:        1024,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
			})

			// Apply request size middleware
			sizeMiddleware := middleware.RequestSizeMiddleware(tt.maxSize)
			handler := sizeMiddleware(testHandler)

			req := httptest.NewRequest("POST", "/test", bytes.NewReader(make([]byte, tt.contentLength)))
			req.ContentLength = tt.contentLength
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				assert.Equal(t, "ok", w.Body.String())
			}
		})
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	t.Run("security headers are set", func(t *testing.T) {
		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})

		// Apply security headers middleware
		headersMiddleware := middleware.SecurityHeadersMiddleware()
		handler := headersMiddleware(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
		assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
		assert.Equal(t, "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none';", w.Header().Get("Content-Security-Policy"))
	})
}

func TestCombinedSecurityMiddleware(t *testing.T) {
	t.Run("all security middlewares work together", func(t *testing.T) {
		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})

		// Apply combined security middleware
		combinedMiddleware := middleware.CombinedSecurityMiddleware()
		handler := combinedMiddleware(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "ok", w.Body.String())

		// Check that security headers are set
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	})

	t.Run("suspicious request is blocked", func(t *testing.T) {
		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})

		// Apply combined security middleware
		combinedMiddleware := middleware.CombinedSecurityMiddleware()
		handler := combinedMiddleware(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Custom", "<script>alert('xss')</script>")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.NotEqual(t, "ok", w.Body.String())
	})
}

func TestMiddleware_EdgeCases(t *testing.T) {
	t.Run("concurrent middleware execution", func(t *testing.T) {
		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})

		// Apply combined middleware
		combinedMiddleware := middleware.CombinedSecurityMiddleware()
		handler := combinedMiddleware(testHandler)

		const numRequests = 10
		results := make(chan struct {
			status int
			err    error
		}, numRequests)

		// Start multiple goroutines
		for i := 0; i < numRequests; i++ {
			go func() {
				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()

				handler.ServeHTTP(w, req)

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
