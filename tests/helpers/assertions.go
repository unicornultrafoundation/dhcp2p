package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/models"
)

// TestAssertions provides custom assertion helpers
type TestAssertions struct {
	t assert.TestingT
}

// NewTestAssertions creates a new test assertions helper
func NewTestAssertions(t assert.TestingT) *TestAssertions {
	return &TestAssertions{t: t}
}

// AssertLeaseEqual validates that two leases are equal
func (ta *TestAssertions) AssertLeaseEqual(expected, actual *models.Lease) {
	if expected == nil && actual == nil {
		return
	}

	assert.NotNil(ta.t, expected, "expected lease should not be nil")
	assert.NotNil(ta.t, actual, "actual lease should not be nil")

	if expected != nil && actual != nil {
		assert.Equal(ta.t, expected.TokenID, actual.TokenID, "TokenID should match")
		assert.Equal(ta.t, expected.PeerID, actual.PeerID, "PeerID should match")
		assert.Equal(ta.t, expected.Ttl, actual.Ttl, "TTL should match")

		// Allow small time differences for timestamps
		assert.WithinDuration(ta.t, expected.CreatedAt, actual.CreatedAt, time.Second, "CreatedAt should be close")
		assert.WithinDuration(ta.t, expected.ExpiresAt, actual.ExpiresAt, time.Second, "ExpiresAt should be close")
	}
}

// AssertLeaseValid validates that a lease is properly formed
func (ta *TestAssertions) AssertLeaseValid(lease *models.Lease) {
	assert.NotNil(ta.t, lease, "lease should not be nil")
	assert.Greater(ta.t, lease.TokenID, int64(167772160), "TokenID should be in valid range")
	assert.Less(ta.t, lease.TokenID, int64(184418305), "TokenID should be in valid range")
	assert.NotEmpty(ta.t, lease.PeerID, "PeerID should not be empty")
	assert.Greater(ta.t, lease.CreatedAt, time.Time{}, "CreatedAt should be set")
	assert.Greater(ta.t, lease.ExpiresAt, lease.CreatedAt, "ExpiresAt should be after CreatedAt")
	assert.GreaterOrEqual(ta.t, lease.Ttl, int64(0), "TTL should be non-negative")
}

// AssertLeaseExpired validates that a lease is expired
func (ta *TestAssertions) AssertLeaseExpired(lease *models.Lease) {
	assert.NotNil(ta.t, lease, "lease should not be nil")
	assert.True(ta.t, time.Now().After(lease.ExpiresAt), "lease should be expired")
}

// AssertLeaseNotExpired validates that a lease is not expired
func (ta *TestAssertions) AssertLeaseNotExpired(lease *models.Lease) {
	assert.NotNil(ta.t, lease, "lease should not be nil")
	assert.False(ta.t, time.Now().After(lease.ExpiresAt), "lease should not be expired")
}

// AssertNonceEqual validates that two nonces are equal
func (ta *TestAssertions) AssertNonceEqual(expected, actual *models.Nonce) {
	if expected == nil && actual == nil {
		return
	}

	assert.NotNil(ta.t, expected, "expected nonce should not be nil")
	assert.NotNil(ta.t, actual, "actual nonce should not be nil")

	if expected != nil && actual != nil {
		assert.Equal(ta.t, expected.ID, actual.ID, "ID should match")
		assert.Equal(ta.t, expected.PeerID, actual.PeerID, "PeerID should match")
		assert.Equal(ta.t, expected.Used, actual.Used, "Used flag should match")

		// Allow small time differences for timestamps
		assert.WithinDuration(ta.t, expected.IssuedAt, actual.IssuedAt, time.Second, "IssuedAt should be close")
		assert.WithinDuration(ta.t, expected.ExpiresAt, actual.ExpiresAt, time.Second, "ExpiresAt should be close")

		if !expected.UsedAt.IsZero() && !actual.UsedAt.IsZero() {
			assert.WithinDuration(ta.t, expected.UsedAt, actual.UsedAt, time.Second, "UsedAt should be close")
		} else {
			assert.Equal(ta.t, expected.UsedAt.IsZero(), actual.UsedAt.IsZero(), "UsedAt should both be zero or both be set")
		}
	}
}

// AssertNonceValid validates that a nonce is properly formed
func (ta *TestAssertions) AssertNonceValid(nonce *models.Nonce) {
	assert.NotNil(ta.t, nonce, "nonce should not be nil")
	assert.NotEmpty(ta.t, nonce.ID, "ID should not be empty")
	assert.NotEmpty(ta.t, nonce.PeerID, "PeerID should not be empty")
	assert.Greater(ta.t, nonce.IssuedAt, time.Time{}, "IssuedAt should be set")
	assert.Greater(ta.t, nonce.ExpiresAt, nonce.IssuedAt, "ExpiresAt should be after IssuedAt")

	// If used, UsedAt should be set and after IssuedAt
	if nonce.Used {
		assert.NotNil(ta.t, nonce.UsedAt, "UsedAt should be set when nonce is used")
		assert.True(ta.t, nonce.UsedAt.After(nonce.IssuedAt), "UsedAt should be after IssuedAt")
	}
}

// AssertNonceExpired validates that a nonce is expired
func (ta *TestAssertions) AssertNonceExpired(nonce *models.Nonce) {
	assert.NotNil(ta.t, nonce, "nonce should not be nil")
	assert.True(ta.t, time.Now().After(nonce.ExpiresAt), "nonce should be expired")
}

// AssertNonceUsed validates that a nonce is used
func (ta *TestAssertions) AssertNonceUsed(nonce *models.Nonce) {
	assert.NotNil(ta.t, nonce, "nonce should not be nil")
	assert.True(ta.t, nonce.Used, "nonce should be marked as used")
	assert.NotNil(ta.t, nonce.UsedAt, "UsedAt should be set")
}

// AssertWithRetry retries an assertion with exponential backoff
func (ta *TestAssertions) AssertWithRetry(assertion func() bool, maxRetries int, baseDelay time.Duration, message string) {
	for i := 0; i < maxRetries; i++ {
		if assertion() {
			return
		}

		if i < maxRetries-1 {
			delay := time.Duration(i+1) * baseDelay
			time.Sleep(delay)
		}
	}

	assert.Fail(ta.t, fmt.Sprintf("%s after %d retries", message, maxRetries))
}

// AssertEventually asserts that a condition becomes true within timeout
func (ta *TestAssertions) AssertEventually(ctx context.Context, condition func() bool, timeout time.Duration, message string) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timeoutChan := time.After(timeout)

	for {
		select {
		case <-ctx.Done():
			assert.Fail(ta.t, fmt.Sprintf("%s: context cancelled", message))
			return
		case <-timeoutChan:
			assert.Fail(ta.t, fmt.Sprintf("%s: timeout after %v", message, timeout))
			return
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}
