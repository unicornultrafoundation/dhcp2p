package unit

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/errors"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/models"
	"github.com/unicornultrafoundation/dhcp2p/tests/config"
	"github.com/unicornultrafoundation/dhcp2p/tests/fixtures"
)

func TestSecurity_LeaseValidation(t *testing.T) {
	builder := fixtures.NewTestBuilder()

	testCases := []struct {
		name        string
		buildLease  func() *models.Lease
		expectError error
		description string
	}{
		{
			name: "valid_lease",
			buildLease: func() *models.Lease {
				return builder.NewLease().Build()
			},
			expectError: nil,
			description: "A properly formed lease should be valid",
		},
		{
			name: "lease_with_network_address",
			buildLease: func() *models.Lease {
				return builder.NewLease().WithTokenID(config.NetworkAddress).Build()
			},
			expectError: errors.ErrInvalidTokenID,
			description: "Network address should not be assignable",
		},
		{
			name: "lease_with_broadcast_address",
			buildLease: func() *models.Lease {
				return builder.NewLease().WithTokenID(config.BroadcastAddress).Build()
			},
			expectError: errors.ErrInvalidTokenID,
			description: "Broadcast address should not be assignable",
		},
		{
			name: "lease_with_token_id_below_range",
			buildLease: func() *models.Lease {
				return builder.NewLease().WithTokenID(config.DefaultStartTokenID - 1).Build()
			},
			expectError: errors.ErrInvalidTokenID,
			description: "Token ID below DHCP range should be invalid",
		},
		{
			name: "lease_with_token_id_above_range",
			buildLease: func() *models.Lease {
				return builder.NewLease().WithTokenID(config.DefaultEndTokenID + 1).Build()
			},
			expectError: errors.ErrInvalidTokenID,
			description: "Token ID above DHCP range should be invalid",
		},
		{
			name: "lease_with_negative_token_id",
			buildLease: func() *models.Lease {
				return builder.NewLease().WithTokenID(-1).Build()
			},
			expectError: errors.ErrInvalidTokenID,
			description: "Negative token ID should be invalid",
		},
		{
			name: "lease_with_empty_peer_id",
			buildLease: func() *models.Lease {
				return builder.NewLease().WithPeerID("").Build()
			},
			expectError: errors.ErrMissingPeerID,
			description: "Empty peer ID should be invalid",
		},
		{
			name: "lease_with_too_long_peer_id",
			buildLease: func() *models.Lease {
				longPeerID := ""
				for i := 0; i < 129; i++ {
					longPeerID += "a"
				}
				return builder.NewLease().WithPeerID(longPeerID).Build()
			},
			expectError: errors.ErrMissingPeerID,
			description: "Peer ID longer than 128 characters should be invalid",
		},
		{
			name: "lease_expires_before_creation",
			buildLease: func() *models.Lease {
				now := time.Now()
				return builder.NewLease().
					WithCreatedAt(now).
					WithExpiresAt(now.Add(-time.Hour)).Build()
			},
			expectError: errors.ErrInvalidTokenID,
			description: "Lease that expires before creation should be invalid",
		},
		{
			name: "lease_with_negative_ttl",
			buildLease: func() *models.Lease {
				return builder.NewLease().WithTTL(-1).Build()
			},
			expectError: errors.ErrInvalidTokenID,
			description: "Negative TTL should be invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lease := tc.buildLease()
			err := validateLeaseSecurity(lease)

			if tc.expectError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectError, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

func TestSecurity_NonceValidation(t *testing.T) {
	builder := fixtures.NewTestBuilder()

	testCases := []struct {
		name        string
		buildNonce  func() *models.Nonce
		expectError error
		description string
	}{
		{
			name: "valid_nonce",
			buildNonce: func() *models.Nonce {
				return builder.NewNonce().Build()
			},
			expectError: nil,
			description: "A properly formed nonce should be valid",
		},
		{
			name: "nonce_with_empty_id",
			buildNonce: func() *models.Nonce {
				return builder.NewNonce().WithID("").Build()
			},
			expectError: errors.ErrMissingPeerID,
			description: "Empty nonce ID should be invalid",
		},
		{
			name: "nonce_with_empty_peer_id",
			buildNonce: func() *models.Nonce {
				return builder.NewNonce().WithPeerID("").Build()
			},
			expectError: errors.ErrMissingPeerID,
			description: "Empty peer ID should be invalid",
		},
		{
			name: "nonce_expires_before_issuance",
			buildNonce: func() *models.Nonce {
				now := time.Now()
				return builder.NewNonce().
					WithIssuedAt(now).
					WithExpiresAt(now.Add(-time.Hour)).Build()
			},
			expectError: errors.ErrNonceNotFound,
			description: "Nonce that expires before issuance should be invalid",
		},
		{
			name: "nonce_marked_used_without_timestamp",
			buildNonce: func() *models.Nonce {
				nonce := builder.NewNonce().Build()
				nonce.Used = true
				nonce.UsedAt = time.Time{} // This should be invalid (zero time)
				return nonce
			},
			expectError: errors.ErrNonceNotFound,
			description: "Used nonce without timestamp should be invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nonce := tc.buildNonce()
			err := validateNonceSecurity(nonce)

			if tc.expectError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectError, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

func TestSecurity_InputSanitization(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal_input",
			input:    "peer-123",
			expected: "peer-123",
		},
		{
			name:     "input_with_sql_injection",
			input:    "peer'; DROP TABLE leases; --",
			expected: "peer DROP TABLE leases ",
		},
		{
			name:     "input_with_xss",
			input:    "<script>alert('xss')</script>",
			expected: "scriptalert(xss)/script",
		},
		{
			name:     "input_with_control_chars",
			input:    "peer\x00\x01\x02",
			expected: "peer",
		},
		{
			name:     "empty_input",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sanitized := sanitizeInput(tc.input)
			assert.Equal(t, tc.expected, sanitized)
		})
	}
}

func TestSecurity_RateLimiting(t *testing.T) {
	// Simulate rate limiting logic
	rateLimiter := NewMockRateLimiter(5, time.Minute) // 5 requests per minute

	testCases := []struct {
		name        string
		requests    int
		expectError bool
		description string
	}{
		{
			name:        "within_limit",
			requests:    3,
			expectError: false,
			description: "Requests within rate limit should succeed",
		},
		{
			name:        "at_limit",
			requests:    5,
			expectError: false,
			description: "Requests at rate limit should succeed",
		},
		{
			name:        "exceeds_limit",
			requests:    6,
			expectError: true,
			description: "Requests exceeding rate limit should fail",
		},
		{
			name:        "far_exceeds_limit",
			requests:    10,
			expectError: true,
			description: "Requests far exceeding rate limit should fail",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset rate limiter
			rateLimiter.Reset()

			var lastErr error
			for i := 0; i < tc.requests; i++ {
				lastErr = rateLimiter.AllowRequest("test-peer")
			}

			if tc.expectError {
				assert.Error(t, lastErr, tc.description)
			} else {
				assert.NoError(t, lastErr, tc.description)
			}
		})
	}
}

// Helper functions for security testing

func validateLeaseSecurity(lease *models.Lease) error {
	if lease == nil {
		return errors.ErrInvalidTokenID
	}

	// Validate token ID range
	if lease.TokenID <= config.NetworkAddress || lease.TokenID >= config.BroadcastAddress {
		return errors.ErrInvalidTokenID
	}

	// Validate peer ID
	if lease.PeerID == "" {
		return errors.ErrMissingPeerID
	}
	if len(lease.PeerID) > 128 {
		return errors.ErrMissingPeerID
	}

	// Validate timestamps
	if lease.ExpiresAt.Before(lease.CreatedAt) {
		return errors.ErrInvalidTokenID
	}

	// Validate TTL
	if lease.Ttl < 0 {
		return errors.ErrInvalidTokenID
	}

	return nil
}

func validateNonceSecurity(nonce *models.Nonce) error {
	if nonce == nil {
		return errors.ErrNonceNotFound
	}

	// Validate ID and peer ID
	if nonce.ID == "" || nonce.PeerID == "" {
		return errors.ErrMissingPeerID
	}

	// Validate timestamps
	if nonce.ExpiresAt.Before(nonce.IssuedAt) {
		return errors.ErrNonceNotFound
	}

	// Validate used state
	if nonce.Used && nonce.UsedAt.IsZero() {
		return errors.ErrNonceNotFound
	}

	return nil
}

func sanitizeInput(input string) string {
	// Remove SQL injection characters
	result := input
	result = strings.ReplaceAll(result, "'", "")
	result = strings.ReplaceAll(result, ";", "")
	result = strings.ReplaceAll(result, "--", "")

	// Remove XSS characters
	result = strings.ReplaceAll(result, "<", "")
	result = strings.ReplaceAll(result, ">", "")
	result = strings.ReplaceAll(result, "\"", "")
	result = strings.ReplaceAll(result, "'", "")

	// Remove control characters
	result = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, result)

	return result
}

// MockRateLimiter for testing rate limiting
type MockRateLimiter struct {
	limit    int
	window   time.Duration
	requests map[string][]time.Time
	mu       sync.RWMutex
}

func NewMockRateLimiter(limit int, window time.Duration) *MockRateLimiter {
	return &MockRateLimiter{
		limit:    limit,
		window:   window,
		requests: make(map[string][]time.Time),
	}
}

func (r *MockRateLimiter) AllowRequest(peerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// Clean old requests
	if times, exists := r.requests[peerID]; exists {
		var validTimes []time.Time
		for _, t := range times {
			if now.Sub(t) < r.window {
				validTimes = append(validTimes, t)
			}
		}
		r.requests[peerID] = validTimes
	}

	// Check limit
	if len(r.requests[peerID]) >= r.limit {
		return errors.ErrMissingDependencies // Using existing error
	}

	// Add new request
	r.requests[peerID] = append(r.requests[peerID], now)
	return nil
}

func (r *MockRateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests = make(map[string][]time.Time)
}
