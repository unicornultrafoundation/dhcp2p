package fixtures

import (
	"time"

	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/models"
)

// TestBuilder provides factory methods for creating test data
type TestBuilder struct{}

// NewTestBuilder creates a new test builder instance
func NewTestBuilder() *TestBuilder {
	return &TestBuilder{}
}

// LeaseBuilder provides fluent interface for building test leases
type LeaseBuilder struct {
	lease *models.Lease
}

// NewLease creates a new lease builder with default values
func (tb *TestBuilder) NewLease() *LeaseBuilder {
	now := time.Now()
	return &LeaseBuilder{
		lease: &models.Lease{
			TokenID:   167772161, // Default DHCP range start
			PeerID:    "test-peer-123",
			CreatedAt: now,
			UpdatedAt: now,
			ExpiresAt: now.Add(time.Hour),
			Ttl:       3600, // 1 hour
		},
	}
}

// WithTokenID sets the token ID
func (lb *LeaseBuilder) WithTokenID(tokenID int64) *LeaseBuilder {
	lb.lease.TokenID = tokenID
	return lb
}

// WithPeerID sets the peer ID
func (lb *LeaseBuilder) WithPeerID(peerID string) *LeaseBuilder {
	lb.lease.PeerID = peerID
	return lb
}

// WithCreatedAt sets the creation time
func (lb *LeaseBuilder) WithCreatedAt(t time.Time) *LeaseBuilder {
	lb.lease.CreatedAt = t
	lb.lease.UpdatedAt = t
	return lb
}

// WithExpiresAt sets the expiration time
func (lb *LeaseBuilder) WithExpiresAt(t time.Time) *LeaseBuilder {
	lb.lease.ExpiresAt = t
	return lb
}

// WithTTL sets the TTL in seconds
func (lb *LeaseBuilder) WithTTL(ttl int64) *LeaseBuilder {
	lb.lease.Ttl = int32(ttl)
	return lb
}

// WithExpired creates an expired lease
func (lb *LeaseBuilder) WithExpired() *LeaseBuilder {
	now := time.Now()
	lb.lease.ExpiresAt = now.Add(-time.Hour)
	lb.lease.Ttl = 0
	return lb
}

// WithExpiringSoon creates a lease that expires soon
func (lb *LeaseBuilder) WithExpiringSoon() *LeaseBuilder {
	lb.lease.ExpiresAt = time.Now().Add(5 * time.Minute)
	lb.lease.Ttl = 300 // 5 minutes
	return lb
}

// Build returns the constructed lease
func (lb *LeaseBuilder) Build() *models.Lease {
	return lb.lease
}

// NonceBuilder provides fluent interface for building test nonces
type NonceBuilder struct {
	nonce *models.Nonce
}

// NewNonce creates a new nonce builder with default values
func (tb *TestBuilder) NewNonce() *NonceBuilder {
	now := time.Now()
	return &NonceBuilder{
		nonce: &models.Nonce{
			ID:        "test-nonce-123",
			PeerID:    "test-peer-123",
			IssuedAt:  now,
			ExpiresAt: now.Add(5 * time.Minute),
			Used:      false,
		},
	}
}

// WithID sets the nonce ID
func (nb *NonceBuilder) WithID(id string) *NonceBuilder {
	nb.nonce.ID = id
	return nb
}

// WithPeerID sets the peer ID
func (nb *NonceBuilder) WithPeerID(peerID string) *NonceBuilder {
	nb.nonce.PeerID = peerID
	return nb
}

// WithIssuedAt sets the issued time
func (nb *NonceBuilder) WithIssuedAt(t time.Time) *NonceBuilder {
	nb.nonce.IssuedAt = t
	return nb
}

// WithExpiresAt sets the expiration time
func (nb *NonceBuilder) WithExpiresAt(t time.Time) *NonceBuilder {
	nb.nonce.ExpiresAt = t
	return nb
}

// WithUsed marks the nonce as used
func (nb *NonceBuilder) WithUsed() *NonceBuilder {
	nb.nonce.Used = true
	nb.nonce.UsedAt = time.Now()
	return nb
}

// WithExpired creates an expired nonce
func (nb *NonceBuilder) WithExpired() *NonceBuilder {
	now := time.Now()
	nb.nonce.ExpiresAt = now.Add(-time.Hour)
	return nb
}

// Build returns the constructed nonce
func (nb *NonceBuilder) Build() *models.Nonce {
	return nb.nonce
}

// TestConfigs provides predefined test configurations
var TestConfigs = struct {
	ValidLeaseRanges   map[string][]int64
	InvalidLeaseRanges []int64
	TestPeers          []string
	TestNonces         []string
	TimeDurations      map[string]time.Duration
}{
	ValidLeaseRanges: map[string][]int64{
		"small":  {167772161, 167772162, 167772163},   // 10.0.0.1-10.0.0.3
		"medium": {167772161, 184418304},              // Full DHCP range
		"large":  generateRange(167772161, 167772180), // 10.0.0.1-10.0.0.20
	},
	InvalidLeaseRanges: []int64{
		0,         // Invalid
		167772160, // Network address 10.0.0.0
		184418305, // Broadcast address 10.255.255.255
		-1,        // Negative
	},
	TestPeers: []string{
		"peer-001",
		"peer-002",
		"peer-003",
		"test-peer-123",
		"integration-test-peer",
	},
	TestNonces: []string{
		"nonce-001",
		"nonce-002",
		"nonce-003",
		"test-nonce-123",
		"integration-nonce",
	},
	TimeDurations: map[string]time.Duration{
		"no_expiration": 0,
		"short":         time.Minute,
		"medium":        5 * time.Minute,
		"long":          time.Hour,
		"very_long":     24 * time.Hour,
	},
}

// generateRange creates a slice of consecutive integers
func generateRange(start, end int64) []int64 {
	result := make([]int64, 0, end-start+1)
	for i := start; i <= end; i++ {
		result = append(result, i)
	}
	return result
}
