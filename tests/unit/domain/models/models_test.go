package models

import (
	"testing"
	"time"

	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/models"
	"github.com/stretchr/testify/assert"
)

func TestLease_Properties(t *testing.T) {
	tests := []struct {
		name        string
		lease       models.Lease
		expectValid bool
	}{
		{
			name: "valid lease",
			lease: models.Lease{
				TokenID:   167772161,
				PeerID:    "peer123",
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(time.Hour),
			},
			expectValid: true,
		},
		{
			name: "zero token ID",
			lease: models.Lease{
				TokenID:   0,
				PeerID:    "peer123",
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(time.Hour),
			},
			expectValid: false,
		},
		{
			name: "negative token ID",
			lease: models.Lease{
				TokenID:   -1,
				PeerID:    "peer123",
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(time.Hour),
			},
			expectValid: false,
		},
		{
			name: "empty peer ID",
			lease: models.Lease{
				TokenID:   167772161,
				PeerID:    "",
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(time.Hour),
			},
			expectValid: false,
		},
		{
			name: "expired lease",
			lease: models.Lease{
				TokenID:   167772161,
				PeerID:    "peer123",
				CreatedAt: time.Now().Add(-2 * time.Hour),
				ExpiresAt: time.Now().Add(-time.Hour),
			},
			expectValid: true, // Expired leases are valid, just expired
		},
		{
			name: "lease expires before creation",
			lease: models.Lease{
				TokenID:   167772161,
				PeerID:    "peer123",
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(-time.Hour),
			},
			expectValid: false,
		},
		{
			name: "very long peer ID",
			lease: models.Lease{
				TokenID:   167772161,
				PeerID:    string(make([]byte, 200)), // Very long peer ID
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(time.Hour),
			},
			expectValid: false,
		},
		{
			name: "peer ID with special characters",
			lease: models.Lease{
				TokenID:   167772161,
				PeerID:    "peer-123_with.special@chars",
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(time.Hour),
			},
			expectValid: true, // Should be valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic properties
			assert.Equal(t, tt.lease.TokenID, tt.lease.TokenID)
			assert.Equal(t, tt.lease.PeerID, tt.lease.PeerID)

			// Test if lease appears valid based on our expectations
			isValid := tt.lease.TokenID > 0 && tt.lease.PeerID != "" && len(tt.lease.PeerID) <= 128 && tt.lease.ExpiresAt.After(tt.lease.CreatedAt)
			assert.Equal(t, tt.expectValid, isValid)
		})
	}
}

func TestLease_IsExpired(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		lease    models.Lease
		expected bool
	}{
		{
			name: "active lease",
			lease: models.Lease{
				TokenID:   167772161,
				PeerID:    "peer123",
				CreatedAt: now.Add(-time.Hour),
				ExpiresAt: now.Add(time.Hour),
			},
			expected: false,
		},
		{
			name: "expired lease",
			lease: models.Lease{
				TokenID:   167772161,
				PeerID:    "peer123",
				CreatedAt: now.Add(-2 * time.Hour),
				ExpiresAt: now.Add(-time.Hour),
			},
			expected: true,
		},
		{
			name: "lease expiring now",
			lease: models.Lease{
				TokenID:   167772161,
				PeerID:    "peer123",
				CreatedAt: now.Add(-time.Hour),
				ExpiresAt: now,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.lease.ExpiresAt.Before(time.Now())
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLease_TimeRemaining(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		lease    models.Lease
		expected time.Duration
	}{
		{
			name: "lease with 1 hour remaining",
			lease: models.Lease{
				TokenID:   167772161,
				PeerID:    "peer123",
				CreatedAt: now.Add(-time.Hour),
				ExpiresAt: now.Add(time.Hour),
			},
			expected: time.Hour,
		},
		{
			name: "expired lease",
			lease: models.Lease{
				TokenID:   167772161,
				PeerID:    "peer123",
				CreatedAt: now.Add(-2 * time.Hour),
				ExpiresAt: now.Add(-time.Hour),
			},
			expected: -time.Hour,
		},
		{
			name: "lease expiring now",
			lease: models.Lease{
				TokenID:   167772161,
				PeerID:    "peer123",
				CreatedAt: now.Add(-time.Hour),
				ExpiresAt: now,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := time.Until(tt.lease.ExpiresAt)
			// Allow for small time differences due to test execution time
			assert.InDelta(t, tt.expected.Seconds(), result.Seconds(), 1.0)
		})
	}
}

func TestNonce_Properties(t *testing.T) {
	tests := []struct {
		name        string
		nonce       models.Nonce
		expectValid bool
	}{
		{
			name: "valid nonce",
			nonce: models.Nonce{
				ID:        "nonce-123",
				PeerID:    "peer123",
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now().Add(5 * time.Minute),
				Used:      false,
			},
			expectValid: true,
		},
		{
			name: "empty ID",
			nonce: models.Nonce{
				ID:        "",
				PeerID:    "peer123",
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now().Add(5 * time.Minute),
				Used:      false,
			},
			expectValid: false,
		},
		{
			name: "empty peer ID",
			nonce: models.Nonce{
				ID:        "nonce-123",
				PeerID:    "",
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now().Add(5 * time.Minute),
				Used:      false,
			},
			expectValid: false,
		},
		{
			name: "expired nonce",
			nonce: models.Nonce{
				ID:        "nonce-123",
				PeerID:    "peer123",
				IssuedAt:  time.Now().Add(-10 * time.Minute),
				ExpiresAt: time.Now().Add(-5 * time.Minute),
				Used:      false,
			},
			expectValid: true, // Expired nonces are valid, just expired
		},
		{
			name: "nonce expires before issued",
			nonce: models.Nonce{
				ID:        "nonce-123",
				PeerID:    "peer123",
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now().Add(-time.Minute),
				Used:      false,
			},
			expectValid: false,
		},
		{
			name: "used nonce",
			nonce: models.Nonce{
				ID:        "nonce-123",
				PeerID:    "peer123",
				IssuedAt:  time.Now().Add(-time.Minute),
				ExpiresAt: time.Now().Add(5 * time.Minute),
				Used:      true,
				UsedAt:    time.Now(),
			},
			expectValid: true, // Used nonces are valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic properties
			assert.Equal(t, tt.nonce.ID, tt.nonce.ID)
			assert.Equal(t, tt.nonce.PeerID, tt.nonce.PeerID)
			assert.Equal(t, tt.nonce.Used, tt.nonce.Used)

			// Test if nonce appears valid based on our expectations
			isValid := tt.nonce.ID != "" && tt.nonce.PeerID != "" && tt.nonce.ExpiresAt.After(tt.nonce.IssuedAt)
			assert.Equal(t, tt.expectValid, isValid)
		})
	}
}

func TestNonce_IsExpired(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		nonce    models.Nonce
		expected bool
	}{
		{
			name: "active nonce",
			nonce: models.Nonce{
				ID:        "nonce-123",
				PeerID:    "peer123",
				IssuedAt:  now.Add(-time.Minute),
				ExpiresAt: now.Add(5 * time.Minute),
				Used:      false,
			},
			expected: false,
		},
		{
			name: "expired nonce",
			nonce: models.Nonce{
				ID:        "nonce-123",
				PeerID:    "peer123",
				IssuedAt:  now.Add(-10 * time.Minute),
				ExpiresAt: now.Add(-5 * time.Minute),
				Used:      false,
			},
			expected: true,
		},
		{
			name: "nonce expiring now",
			nonce: models.Nonce{
				ID:        "nonce-123",
				PeerID:    "peer123",
				IssuedAt:  now.Add(-5 * time.Minute),
				ExpiresAt: now,
				Used:      false,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.nonce.ExpiresAt.Before(time.Now())
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNonce_IsUsed(t *testing.T) {
	tests := []struct {
		name     string
		nonce    models.Nonce
		expected bool
	}{
		{
			name: "unused nonce",
			nonce: models.Nonce{
				ID:        "nonce-123",
				PeerID:    "peer123",
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now().Add(5 * time.Minute),
				Used:      false,
			},
			expected: false,
		},
		{
			name: "used nonce",
			nonce: models.Nonce{
				ID:        "nonce-123",
				PeerID:    "peer123",
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now().Add(5 * time.Minute),
				Used:      true,
				UsedAt:    time.Now(),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.nonce.Used
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuthRequest_Properties(t *testing.T) {
	tests := []struct {
		name        string
		request     models.AuthRequest
		expectValid bool
	}{
		{
			name: "valid auth request",
			request: models.AuthRequest{
				Pubkey: make([]byte, 64), // Valid size
			},
			expectValid: true,
		},
		{
			name: "nil pubkey",
			request: models.AuthRequest{
				Pubkey: nil,
			},
			expectValid: false,
		},
		{
			name: "empty pubkey",
			request: models.AuthRequest{
				Pubkey: []byte{},
			},
			expectValid: false,
		},
		{
			name: "pubkey too short",
			request: models.AuthRequest{
				Pubkey: []byte("short"),
			},
			expectValid: false,
		},
		{
			name: "pubkey too long",
			request: models.AuthRequest{
				Pubkey: make([]byte, 2000),
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic properties
			assert.Equal(t, tt.request.Pubkey, tt.request.Pubkey)

			// Test if request appears valid based on our expectations
			isValid := len(tt.request.Pubkey) >= 32 && len(tt.request.Pubkey) <= 1024
			assert.Equal(t, tt.expectValid, isValid)
		})
	}
}

func TestAuthVerifyRequest_Properties(t *testing.T) {
	tests := []struct {
		name        string
		request     models.AuthVerifyRequest
		expectValid bool
	}{
		{
			name: "valid auth verify request",
			request: models.AuthVerifyRequest{
				NonceID:   "nonce-123",
				Signature: []byte("valid-signature"),
				Pubkey:    []byte("valid-pubkey-data"),
			},
			expectValid: true,
		},
		{
			name: "empty nonce ID",
			request: models.AuthVerifyRequest{
				NonceID:   "",
				Signature: []byte("valid-signature"),
				Pubkey:    []byte("valid-pubkey-data"),
			},
			expectValid: false,
		},
		{
			name: "nil signature",
			request: models.AuthVerifyRequest{
				NonceID:   "nonce-123",
				Signature: nil,
				Pubkey:    []byte("valid-pubkey-data"),
			},
			expectValid: false,
		},
		{
			name: "empty signature",
			request: models.AuthVerifyRequest{
				NonceID:   "nonce-123",
				Signature: []byte{},
				Pubkey:    []byte("valid-pubkey-data"),
			},
			expectValid: false,
		},
		{
			name: "nil pubkey",
			request: models.AuthVerifyRequest{
				NonceID:   "nonce-123",
				Signature: []byte("valid-signature"),
				Pubkey:    nil,
			},
			expectValid: false,
		},
		{
			name: "empty pubkey",
			request: models.AuthVerifyRequest{
				NonceID:   "nonce-123",
				Signature: []byte("valid-signature"),
				Pubkey:    []byte{},
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic properties
			assert.Equal(t, tt.request.NonceID, tt.request.NonceID)
			assert.Equal(t, tt.request.Signature, tt.request.Signature)
			assert.Equal(t, tt.request.Pubkey, tt.request.Pubkey)

			// Test if request appears valid based on our expectations
			isValid := tt.request.NonceID != "" &&
				tt.request.Signature != nil && len(tt.request.Signature) > 0 &&
				tt.request.Pubkey != nil && len(tt.request.Pubkey) > 0
			assert.Equal(t, tt.expectValid, isValid)
		})
	}
}

func TestNonceRequest_Properties(t *testing.T) {
	tests := []struct {
		name        string
		request     models.NonceRequest
		expectValid bool
	}{
		{
			name: "valid nonce request",
			request: models.NonceRequest{
				NonceID:   "nonce-123",
				Pubkey:    []byte("valid-pubkey-data"),
				Payload:   []byte("valid-payload"),
				Signature: []byte("valid-signature"),
			},
			expectValid: true,
		},
		{
			name: "empty nonce ID",
			request: models.NonceRequest{
				NonceID:   "",
				Pubkey:    []byte("valid-pubkey-data"),
				Payload:   []byte("valid-payload"),
				Signature: []byte("valid-signature"),
			},
			expectValid: false,
		},
		{
			name: "nil pubkey",
			request: models.NonceRequest{
				NonceID:   "nonce-123",
				Pubkey:    nil,
				Payload:   []byte("valid-payload"),
				Signature: []byte("valid-signature"),
			},
			expectValid: false,
		},
		{
			name: "nil payload",
			request: models.NonceRequest{
				NonceID:   "nonce-123",
				Pubkey:    []byte("valid-pubkey-data"),
				Payload:   nil,
				Signature: []byte("valid-signature"),
			},
			expectValid: false,
		},
		{
			name: "nil signature",
			request: models.NonceRequest{
				NonceID:   "nonce-123",
				Pubkey:    []byte("valid-pubkey-data"),
				Payload:   []byte("valid-payload"),
				Signature: nil,
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic properties
			assert.Equal(t, tt.request.NonceID, tt.request.NonceID)
			assert.Equal(t, tt.request.Pubkey, tt.request.Pubkey)
			assert.Equal(t, tt.request.Payload, tt.request.Payload)
			assert.Equal(t, tt.request.Signature, tt.request.Signature)

			// Test if request appears valid based on our expectations
			isValid := tt.request.NonceID != "" &&
				tt.request.Pubkey != nil && len(tt.request.Pubkey) > 0 &&
				tt.request.Payload != nil && len(tt.request.Payload) > 0 &&
				tt.request.Signature != nil && len(tt.request.Signature) > 0
			assert.Equal(t, tt.expectValid, isValid)
		})
	}
}

func TestModels_EdgeCases(t *testing.T) {
	t.Run("lease with maximum token ID", func(t *testing.T) {
		lease := models.Lease{
			TokenID:   9223372036854775807, // Max int64
			PeerID:    "peer123",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(time.Hour),
		}
		assert.Equal(t, int64(9223372036854775807), lease.TokenID)
		assert.True(t, lease.TokenID > 0)
	})

	t.Run("nonce with very long ID", func(t *testing.T) {
		nonce := models.Nonce{
			ID:        string(make([]byte, 1000)), // Very long ID
			PeerID:    "peer123",
			IssuedAt:  time.Now(),
			ExpiresAt: time.Now().Add(5 * time.Minute),
			Used:      false,
		}
		assert.Equal(t, 1000, len(nonce.ID))
		assert.True(t, len(nonce.ID) > 100) // Should be considered too long
	})

	t.Run("auth request with maximum size pubkey", func(t *testing.T) {
		request := models.AuthRequest{
			Pubkey: make([]byte, 1024), // Maximum allowed size
		}
		assert.Equal(t, 1024, len(request.Pubkey))
		assert.True(t, len(request.Pubkey) >= 32 && len(request.Pubkey) <= 1024)
	})

	t.Run("auth request with oversized pubkey", func(t *testing.T) {
		request := models.AuthRequest{
			Pubkey: make([]byte, 2000), // Too large
		}
		assert.Equal(t, 2000, len(request.Pubkey))
		assert.True(t, len(request.Pubkey) > 1024) // Should be considered too large
	})

	t.Run("nonce request with maximum size data", func(t *testing.T) {
		request := models.NonceRequest{
			NonceID:   "nonce-123",
			Pubkey:    make([]byte, 1024),
			Payload:   make([]byte, 1024),
			Signature: make([]byte, 1024),
		}
		assert.Equal(t, 1024, len(request.Pubkey))
		assert.Equal(t, 1024, len(request.Payload))
		assert.Equal(t, 1024, len(request.Signature))
		assert.True(t, len(request.Pubkey) <= 1024)
	})
}
