package config

import (
	"time"
)

// TestConstants provides centralized test constants
type TestConstants struct{}

const (
	// DHCP Range Constants
	DefaultStartTokenID = 167772161 // 10.0.0.1
	DefaultEndTokenID   = 184418304 // 10.255.255.255
	NetworkAddress      = 167772160 // 10.0.0.0
	BroadcastAddress    = 184418305 // 10.255.255.255
	
	// Test Time Constants
	DefaultTTL           = 3600 // 1 hour in seconds
	DefaultNonceTTL      = 300  // 5 minutes in seconds
	ShortTTL            = 60   // 1 minute
	TestTimeout         = 30 * time.Second
	ContainerTimeout    = 60 * time.Second
	DatabaseTimeout     = 10 * time.Second
	
	// Container Constants  
	PostgresPort = "5432"
	RedisPort    = "6379"
	TestDatabase = "dhcp2p_test"
	TestUser     = "test"
	TestPassword = "test"
	
	// Test Data Constants
	DefaultPeerID    = "test-peer-123"
	DefaultNonceID   = "test-nonce-123"
	TestRequestID    = "test-request-123"
	TestIPAddress    = "10.0.0.1"
	
	// Performance Test Constants
 BenchmarkIterations = 1000
 LoadTestDuration   = 30 * time.Second
 ConcurrentUsers    = 50
 RequestRate        = 100 // requests per second
 
 // Error Constants
 TestErrorMessage = "test error"
	
	// Integration Test Constants
 IntegrationRetryAttempts = 3
 IntegrationRetryDelay    = 100 * time.Millisecond
)

// TestTimeouts provides configurable timeout values
var TestTimeouts = struct {
	ContainerStartup time.Duration
	DatabaseConnect  time.Duration
	ServiceReady     time.Duration
	TestExecution    time.Duration
	RequestTimeout   time.Duration
}{
	ContainerStartup: 60 * time.Second,
	DatabaseConnect:  10 * time.Second,
	ServiceReady:     30 * time.Second,
	TestExecution:    2 * time.Minute,
	RequestTimeout:   5 * time.Second,
}

// TestPorts provides configurable port numbers
var TestPorts = struct {
	Postgres int
	Redis    int
	HTTP     int
	GRPC     int
}{
	Postgres: 5432,
	Redis:    6379,
	HTTP:     8080,
	GRPC:     9090,
}

// TestImages provides container image names
var TestImages = struct {
	Postgres string
	Redis    string
	App      string
}{
	Postgres: "postgres:15-alpine",
	Redis:    "redis:7-alpine",
	App:      "dhcp2p:test",
}

// TestLimits provides test limits and constraints
var TestLimits = struct {
	MaxRetries       int
	MaxConcurrency   int
	MaxTestDuration  time.Duration
	ContainerCleanup time.Duration
}{
	MaxRetries:       5,
	MaxConcurrency:   100,
	MaxTestDuration:  10 * time.Minute,
	ContainerCleanup: 30 * time.Second,
}
