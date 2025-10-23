package validation

import (
	"context"
	"encoding/base64"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/handlers/http/keys"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/handlers/http/validation"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/errors"
)

func TestValidateHeader(t *testing.T) {
	tests := []struct {
		name          string
		headerName    string
		headerValue   string
		config        validation.ValidationConfig
		expectedValue string
		expectedError error
	}{
		{
			name:          "valid header",
			headerName:    "X-Test",
			headerValue:   "test-value",
			config:        validation.DefaultValidationConfig(),
			expectedValue: "test-value",
			expectedError: nil,
		},
		{
			name:          "missing header",
			headerName:    "X-Missing",
			headerValue:   "",
			config:        validation.DefaultValidationConfig(),
			expectedValue: "",
			expectedError: errors.ErrMissingHeaders,
		},
		{
			name:        "empty header with allow empty",
			headerName:  "X-Empty",
			headerValue: "",
			config: func() validation.ValidationConfig {
				config := validation.DefaultValidationConfig()
				config.AllowEmpty = true
				return config
			}(),
			expectedValue: "",
			expectedError: nil,
		},
		{
			name:        "header too short",
			headerName:  "X-Short",
			headerValue: "a",
			config: func() validation.ValidationConfig {
				config := validation.DefaultValidationConfig()
				config.MinLength = 5
				return config
			}(),
			expectedValue: "",
			expectedError: errors.ErrInvalidPubkey,
		},
		{
			name:        "header too long",
			headerName:  "X-Long",
			headerValue: string(make([]byte, 3000)),
			config: func() validation.ValidationConfig {
				config := validation.DefaultValidationConfig()
				config.MaxLength = 1000
				return config
			}(),
			expectedValue: "",
			expectedError: errors.ErrInvalidPubkey,
		},
		{
			name:          "header with whitespace",
			headerName:    "X-Whitespace",
			headerValue:   "  test-value  ",
			config:        validation.DefaultValidationConfig(),
			expectedValue: "test-value",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.headerValue != "" {
				req.Header.Set(tt.headerName, tt.headerValue)
			}

			result := validation.ValidateHeader(req, tt.headerName, tt.config)

			assert.Equal(t, tt.expectedValue, result.Value)
			if tt.expectedError != nil {
				assert.Error(t, result.Error)
				assert.Contains(t, result.Error.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, result.Error)
			}
		})
	}
}

func TestValidateURLParam(t *testing.T) {
	tests := []struct {
		name          string
		paramName     string
		paramValue    string
		config        validation.ValidationConfig
		expectedValue string
		expectedError error
	}{
		{
			name:          "valid URL param",
			paramName:     "peerID",
			paramValue:    "peer123",
			config:        validation.DefaultValidationConfig(),
			expectedValue: "peer123",
			expectedError: nil,
		},
		{
			name:          "missing URL param",
			paramName:     "missing",
			paramValue:    "",
			config:        validation.DefaultValidationConfig(),
			expectedValue: "",
			expectedError: errors.ErrMissingHeaders,
		},
		{
			name:       "empty URL param with allow empty",
			paramName:  "empty",
			paramValue: "",
			config: func() validation.ValidationConfig {
				config := validation.DefaultValidationConfig()
				config.AllowEmpty = true
				return config
			}(),
			expectedValue: "",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)

			// Create chi context with URL parameters
			rctx := chi.NewRouteContext()
			if tt.paramValue != "" {
				rctx.URLParams.Add(tt.paramName, tt.paramValue)
			}
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			result := validation.ValidateURLParam(req, tt.paramName, tt.config)

			assert.Equal(t, tt.expectedValue, result.Value)
			if tt.expectedError != nil {
				assert.Error(t, result.Error)
				assert.Contains(t, result.Error.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, result.Error)
			}
		})
	}
}

func TestValidateQueryParam(t *testing.T) {
	tests := []struct {
		name          string
		paramName     string
		paramValue    string
		config        validation.ValidationConfig
		expectedValue string
		expectedError error
	}{
		{
			name:          "valid query param",
			paramName:     "tokenID",
			paramValue:    "12345",
			config:        validation.DefaultValidationConfig(),
			expectedValue: "12345",
			expectedError: nil,
		},
		{
			name:          "missing query param",
			paramName:     "missing",
			paramValue:    "",
			config:        validation.DefaultValidationConfig(),
			expectedValue: "",
			expectedError: errors.ErrMissingHeaders,
		},
		{
			name:       "empty query param with allow empty",
			paramName:  "empty",
			paramValue: "",
			config: func() validation.ValidationConfig {
				config := validation.DefaultValidationConfig()
				config.AllowEmpty = true
				return config
			}(),
			expectedValue: "",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/test"
			if tt.paramValue != "" {
				url += "?" + tt.paramName + "=" + tt.paramValue
			}
			req := httptest.NewRequest("GET", url, nil)

			result := validation.ValidateQueryParam(req, tt.paramName, tt.config)

			assert.Equal(t, tt.expectedValue, result.Value)
			if tt.expectedError != nil {
				assert.Error(t, result.Error)
				assert.Contains(t, result.Error.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, result.Error)
			}
		})
	}
}

func TestValidatePeerIDFromContext(t *testing.T) {
	tests := []struct {
		name          string
		contextValue  interface{}
		expectedValue string
		expectedError error
	}{
		{
			name:          "valid peer ID in context",
			contextValue:  "peer123",
			expectedValue: "peer123",
			expectedError: nil,
		},
		{
			name:          "nil context value",
			contextValue:  nil,
			expectedValue: "",
			expectedError: errors.ErrMissingPeerID,
		},
		{
			name:          "empty peer ID",
			contextValue:  "",
			expectedValue: "",
			expectedError: errors.ErrMissingPeerID,
		},
		{
			name:          "non-string context value",
			contextValue:  123,
			expectedValue: "",
			expectedError: errors.ErrMissingPeerID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.contextValue != nil {
				req = req.WithContext(context.WithValue(req.Context(), keys.PeerIDContextKey, tt.contextValue))
			}

			result := validation.ValidatePeerIDFromContext(req)

			assert.Equal(t, tt.expectedValue, result.Value)
			if tt.expectedError != nil {
				assert.Error(t, result.Error)
				assert.Contains(t, result.Error.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, result.Error)
			}
		})
	}
}

func TestValidateTokenID(t *testing.T) {
	tests := []struct {
		name          string
		tokenIDStr    string
		expectedValue string
		expectedError error
	}{
		{
			name:          "valid token ID",
			tokenIDStr:    "12345",
			expectedValue: "12345",
			expectedError: nil,
		},
		{
			name:          "valid large token ID",
			tokenIDStr:    "9223372036854775807", // Max int64
			expectedValue: "9223372036854775807",
			expectedError: nil,
		},
		{
			name:          "empty token ID",
			tokenIDStr:    "",
			expectedValue: "",
			expectedError: errors.ErrMissingTokenID,
		},
		{
			name:          "invalid token ID",
			tokenIDStr:    "abc",
			expectedValue: "",
			expectedError: errors.ErrInvalidTokenID,
		},
		{
			name:          "negative token ID",
			tokenIDStr:    "-123",
			expectedValue: "",
			expectedError: errors.ErrInvalidTokenID,
		},
		{
			name:          "token ID with decimal",
			tokenIDStr:    "123.45",
			expectedValue: "",
			expectedError: errors.ErrInvalidTokenID,
		},
		{
			name:          "token ID too large",
			tokenIDStr:    "9223372036854775808", // Max int64 + 1
			expectedValue: "",
			expectedError: errors.ErrInvalidTokenID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validation.ValidateTokenID(tt.tokenIDStr)

			assert.Equal(t, tt.expectedValue, result.Value)
			if tt.expectedError != nil {
				assert.Error(t, result.Error)
				assert.Contains(t, result.Error.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, result.Error)
			}
		})
	}
}

func TestValidateBase64Pubkey(t *testing.T) {
	tests := []struct {
		name          string
		pubkey        string
		expectedValue string
		expectedError error
	}{
		{
			name:          "valid base64 pubkey",
			pubkey:        base64.StdEncoding.EncodeToString(make([]byte, 64)),
			expectedValue: base64.StdEncoding.EncodeToString(make([]byte, 64)),
			expectedError: nil,
		},
		{
			name:          "empty pubkey",
			pubkey:        "",
			expectedValue: "",
			expectedError: errors.ErrMissingPubkey,
		},
		{
			name:          "invalid base64",
			pubkey:        "invalid-base64!@#",
			expectedValue: "",
			expectedError: errors.ErrInvalidPubkey,
		},
		{
			name:          "pubkey too short",
			pubkey:        base64.StdEncoding.EncodeToString(make([]byte, 8)), // Less than 16 bytes
			expectedValue: "",
			expectedError: errors.ErrInvalidPubkey,
		},
		{
			name:          "pubkey too long",
			pubkey:        base64.StdEncoding.EncodeToString(make([]byte, 2000)),
			expectedValue: "",
			expectedError: errors.ErrInvalidPubkey,
		},
		{
			name:          "pubkey with whitespace",
			pubkey:        " " + base64.StdEncoding.EncodeToString(make([]byte, 64)) + " ",
			expectedValue: base64.StdEncoding.EncodeToString(make([]byte, 64)),
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validation.ValidateBase64Pubkey(tt.pubkey)

			assert.Equal(t, tt.expectedValue, result.Value)
			if tt.expectedError != nil {
				assert.Error(t, result.Error)
				assert.Contains(t, result.Error.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, result.Error)
			}
		})
	}
}

func TestValidateBase64Signature(t *testing.T) {
	tests := []struct {
		name          string
		signature     string
		expectedValue string
		expectedError error
	}{
		{
			name:          "valid base64 signature",
			signature:     base64.StdEncoding.EncodeToString(make([]byte, 64)),
			expectedValue: base64.StdEncoding.EncodeToString(make([]byte, 64)),
			expectedError: nil,
		},
		{
			name:          "empty signature",
			signature:     "",
			expectedValue: "",
			expectedError: errors.ErrMissingSignature,
		},
		{
			name:          "invalid base64",
			signature:     "invalid-base64!@#",
			expectedValue: "",
			expectedError: errors.ErrInvalidSignature,
		},
		{
			name:          "signature too short",
			signature:     base64.StdEncoding.EncodeToString(make([]byte, 16)),
			expectedValue: "",
			expectedError: errors.ErrInvalidSignature,
		},
		{
			name:          "signature too long",
			signature:     base64.StdEncoding.EncodeToString(make([]byte, 2000)),
			expectedValue: "",
			expectedError: errors.ErrInvalidSignature,
		},
		{
			name:          "signature with whitespace",
			signature:     " " + base64.StdEncoding.EncodeToString(make([]byte, 64)) + " ",
			expectedValue: base64.StdEncoding.EncodeToString(make([]byte, 64)),
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validation.ValidateBase64Signature(tt.signature)

			assert.Equal(t, tt.expectedValue, result.Value)
			if tt.expectedError != nil {
				assert.Error(t, result.Error)
				assert.Contains(t, result.Error.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, result.Error)
			}
		})
	}
}

func TestValidation_EdgeCases(t *testing.T) {
	t.Run("very long header value", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Long", string(make([]byte, 10000)))

		result := validation.ValidateHeader(req, "X-Long", validation.DefaultValidationConfig())

		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), errors.ErrInvalidPubkey.Error())
	})

	t.Run("special characters in header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Special", "test-value-with-special-chars!@#$%^&*()")

		result := validation.ValidateHeader(req, "X-Special", validation.DefaultValidationConfig())

		assert.NoError(t, result.Error)
		assert.Equal(t, "test-value-with-special-chars!@#$%^&*()", result.Value)
	})

	t.Run("unicode characters in header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Unicode", "test-value-with-unicode-🚀-emoji")

		result := validation.ValidateHeader(req, "X-Unicode", validation.DefaultValidationConfig())

		assert.NoError(t, result.Error)
		assert.Equal(t, "test-value-with-unicode-🚀-emoji", result.Value)
	})

	t.Run("case insensitive header names", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("x-test", "test-value") // lowercase

		result := validation.ValidateHeader(req, "X-Test", validation.DefaultValidationConfig())

		assert.NoError(t, result.Error)
		assert.Equal(t, "test-value", result.Value)
	})

	t.Run("multiple header values", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Multiple", "value1")
		req.Header.Add("X-Multiple", "value2")

		result := validation.ValidateHeader(req, "X-Multiple", validation.DefaultValidationConfig())

		assert.NoError(t, result.Error)
		assert.Equal(t, "value1", result.Value) // Should get the first value
	})
}
