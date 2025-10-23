package unit

import (
	"testing"

	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/errors"
	"github.com/stretchr/testify/assert"
)

func TestDomainErrors_Error(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrLeaseNotFound",
			err:      errors.ErrLeaseNotFound,
			expected: "LEASE_NOT_FOUND: Lease not found",
		},
		{
			name:     "ErrMissingPeerID",
			err:      errors.ErrMissingPeerID,
			expected: "MISSING_PEER_ID: Peer ID is required",
		},
		{
			name:     "ErrMissingTokenID",
			err:      errors.ErrMissingTokenID,
			expected: "MISSING_TOKEN_ID: Token ID is required",
		},
		{
			name:     "ErrInvalidTokenID",
			err:      errors.ErrInvalidTokenID,
			expected: "INVALID_TOKEN_ID: Invalid token ID format",
		},
		{
			name:     "ErrNonceNotFound",
			err:      errors.ErrNonceNotFound,
			expected: "NONCE_NOT_FOUND: Nonce not found",
		},
		{
			name:     "ErrMissingHeaders",
			err:      errors.ErrMissingHeaders,
			expected: "MISSING_HEADERS: Required headers are missing",
		},
		{
			name:     "ErrMissingPubkey",
			err:      errors.ErrMissingPubkey,
			expected: "MISSING_PUBKEY: Public key is required",
		},
		{
			name:     "ErrMissingDependencies",
			err:      errors.ErrMissingDependencies,
			expected: "MISSING_DEPENDENCIES: Missing required dependencies",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.err.Error())
			assert.Error(t, tc.err)
		})
	}
}

func TestErrorComparison(t *testing.T) {
	// Test error comparison
	leaseErr := errors.ErrLeaseNotFound
	nonceErr := errors.ErrNonceNotFound

	assert.NotEqual(t, leaseErr, nonceErr)
	assert.Equal(t, errors.ErrLeaseNotFound, leaseErr)
}

// Test error edge cases
func TestErrorUsage(t *testing.T) {
	t.Run("Error is Error", func(t *testing.T) {
		err := errors.ErrLeaseNotFound
		assert.Error(t, err)
		assert.NotNil(t, err)
	})

	t.Run("Error Message", func(t *testing.T) {
		err := errors.ErrLeaseNotFound
		assert.NotEmpty(t, err.Error())
		assert.Contains(t, err.Error(), "LEASE_NOT_FOUND")
	})
}
