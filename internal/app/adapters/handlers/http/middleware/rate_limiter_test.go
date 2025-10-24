package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/unicornultrafoundation/dhcp2p/internal/app/infrastructure/config"
)

func TestRateLimiter_ExtractClientIP(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.AppConfig{
		RateLimitEnabled:           true,
		RateLimitRequestsPerMinute: 100,
		RateLimitBurst:             20,
		RateLimitTrustedProxies:    []string{"127.0.0.1", "10.0.0.0/8"},
	}

	rl := NewRateLimiter(cfg, logger)
	defer rl.Stop()

	tests := []struct {
		name        string
		remoteAddr  string
		headers     map[string]string
		expectedIP  string
		description string
	}{
		{
			name:        "Direct connection",
			remoteAddr:  "192.168.1.100:12345",
			headers:     map[string]string{},
			expectedIP:  "192.168.1.100",
			description: "Should use RemoteAddr when no proxy headers",
		},
		{
			name:       "X-Real-IP from trusted proxy",
			remoteAddr: "127.0.0.1:12345",
			headers: map[string]string{
				"X-Real-IP": "203.0.113.1",
			},
			expectedIP:  "203.0.113.1",
			description: "Should use X-Real-IP when from trusted proxy",
		},
		{
			name:       "X-Forwarded-For from trusted proxy",
			remoteAddr: "127.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.1, 198.51.100.1",
			},
			expectedIP:  "203.0.113.1",
			description: "Should use first IP from X-Forwarded-For when from trusted proxy",
		},
		{
			name:       "Untrusted proxy ignores headers",
			remoteAddr: "192.168.1.1:12345",
			headers: map[string]string{
				"X-Real-IP":       "203.0.113.1",
				"X-Forwarded-For": "203.0.113.1",
			},
			expectedIP:  "192.168.1.1",
			description: "Should ignore proxy headers when not from trusted proxy",
		},
		{
			name:       "Invalid IP in header",
			remoteAddr: "127.0.0.1:12345",
			headers: map[string]string{
				"X-Real-IP": "invalid-ip",
			},
			expectedIP:  "127.0.0.1",
			description: "Should fallback to RemoteAddr when header contains invalid IP",
		},
		{
			name:       "Empty X-Forwarded-For",
			remoteAddr: "127.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "",
			},
			expectedIP:  "127.0.0.1",
			description: "Should fallback to RemoteAddr when X-Forwarded-For is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			actualIP := rl.extractClientIP(req)
			assert.Equal(t, tt.expectedIP, actualIP, tt.description)
		})
	}
}

func TestRateLimiter_TokenBucketBehavior(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.AppConfig{
		RateLimitEnabled:           true,
		RateLimitRequestsPerMinute: 60, // 1 request per second
		RateLimitBurst:             2,  // Allow burst of 2
		RateLimitTrustedProxies:    []string{},
	}

	rl := NewRateLimiter(cfg, logger)
	defer rl.Stop()

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	// First two requests should be allowed (burst capacity)
	allowed1, retryAfter1, remaining1 := rl.Allow(req)
	assert.True(t, allowed1, "First request should be allowed")
	assert.Equal(t, time.Duration(0), retryAfter1)
	assert.Equal(t, 1, remaining1)

	allowed2, retryAfter2, remaining2 := rl.Allow(req)
	assert.True(t, allowed2, "Second request should be allowed (burst)")
	assert.Equal(t, time.Duration(0), retryAfter2)
	assert.Equal(t, 0, remaining2)

	// Third request should be rate limited
	allowed3, retryAfter3, remaining3 := rl.Allow(req)
	assert.False(t, allowed3, "Third request should be rate limited")
	assert.True(t, retryAfter3 > 0, "Should have retry after delay")
	assert.Equal(t, 0, remaining3)
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.AppConfig{
		RateLimitEnabled:           true,
		RateLimitRequestsPerMinute: 2, // Very low limit for testing
		RateLimitBurst:             1,
		RateLimitTrustedProxies:    []string{},
	}

	rl := NewRateLimiter(cfg, logger)
	defer rl.Stop()

	// Create requests from different IPs
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.100:12345"

	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.101:12345"

	// Both IPs should be able to make requests independently
	allowed1, _, _ := rl.Allow(req1)
	assert.True(t, allowed1, "First IP should be allowed")

	allowed2, _, _ := rl.Allow(req2)
	assert.True(t, allowed2, "Second IP should be allowed")

	// Second request from first IP should be rate limited
	allowed3, _, _ := rl.Allow(req1)
	assert.False(t, allowed3, "Second request from first IP should be rate limited")

	// But second IP should still be allowed
	allowed4, _, _ := rl.Allow(req2)
	assert.False(t, allowed4, "Second request from second IP should also be rate limited")
}

func TestRateLimiter_Disabled(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.AppConfig{
		RateLimitEnabled:           false,
		RateLimitRequestsPerMinute: 1,
		RateLimitBurst:             1,
		RateLimitTrustedProxies:    []string{},
	}

	rl := NewRateLimiter(cfg, logger)
	defer rl.Stop()

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	// All requests should be allowed when rate limiting is disabled
	for i := 0; i < 10; i++ {
		allowed, retryAfter, remaining := rl.Allow(req)
		assert.True(t, allowed, "Request %d should be allowed when rate limiting is disabled", i+1)
		assert.Equal(t, time.Duration(0), retryAfter)
		assert.Equal(t, cfg.RateLimitRequestsPerMinute, remaining)
	}
}

func TestRateLimitMiddleware_Headers(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.AppConfig{
		RateLimitEnabled:           true,
		RateLimitRequestsPerMinute: 100,
		RateLimitBurst:             20,
		RateLimitTrustedProxies:    []string{},
	}

	middleware := RateLimitMiddleware(cfg, logger)

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with middleware
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	// Test successful request
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "100", w.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "19", w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
}

func TestRateLimitMiddleware_RateLimitExceeded(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.AppConfig{
		RateLimitEnabled:           true,
		RateLimitRequestsPerMinute: 1, // Very low limit
		RateLimitBurst:             1,
		RateLimitTrustedProxies:    []string{},
	}

	middleware := RateLimitMiddleware(cfg, logger)

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with middleware
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	// First request should succeed
	w1 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w1, req)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request should be rate limited
	w2 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w2, req)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	assert.NotEmpty(t, w2.Header().Get("Retry-After"))

	// Check response body contains rate limit error
	assert.Contains(t, w2.Body.String(), "rate_limit_error")
}

func TestRateLimiter_ConcurrentRequests(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.AppConfig{
		RateLimitEnabled:           true,
		RateLimitRequestsPerMinute: 10,
		RateLimitBurst:             5,
		RateLimitTrustedProxies:    []string{},
	}

	rl := NewRateLimiter(cfg, logger)
	defer rl.Stop()

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	// Test concurrent requests from same IP
	results := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			allowed, _, _ := rl.Allow(req)
			results <- allowed
		}()
	}

	// Collect results
	allowedCount := 0
	for i := 0; i < 10; i++ {
		if <-results {
			allowedCount++
		}
	}

	// Should allow exactly burst capacity (5) requests
	assert.Equal(t, 5, allowedCount, "Should allow exactly burst capacity requests")
}

func TestRateLimiter_TrustedProxyCIDR(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.AppConfig{
		RateLimitEnabled:           true,
		RateLimitRequestsPerMinute: 100,
		RateLimitBurst:             20,
		RateLimitTrustedProxies:    []string{"10.0.0.0/8", "172.16.0.0/12"},
	}

	rl := NewRateLimiter(cfg, logger)
	defer rl.Stop()

	tests := []struct {
		name       string
		remoteAddr string
		xRealIP    string
		expectedIP string
	}{
		{
			name:       "Proxy in 10.0.0.0/8 range",
			remoteAddr: "10.1.1.1:12345",
			xRealIP:    "203.0.113.1",
			expectedIP: "203.0.113.1",
		},
		{
			name:       "Proxy in 172.16.0.0/12 range",
			remoteAddr: "172.16.1.1:12345",
			xRealIP:    "203.0.113.2",
			expectedIP: "203.0.113.2",
		},
		{
			name:       "Proxy not in trusted range",
			remoteAddr: "192.168.1.1:12345",
			xRealIP:    "203.0.113.3",
			expectedIP: "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			req.Header.Set("X-Real-IP", tt.xRealIP)

			actualIP := rl.extractClientIP(req)
			assert.Equal(t, tt.expectedIP, actualIP)
		})
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.AppConfig{
		RateLimitEnabled:           true,
		RateLimitRequestsPerMinute: 100,
		RateLimitBurst:             20,
		RateLimitTrustedProxies:    []string{},
	}

	rl := NewRateLimiter(cfg, logger)
	defer rl.Stop()

	// Create some limiters
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.100:12345"
	rl.Allow(req1)

	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.101:12345"
	rl.Allow(req2)

	// Verify limiters exist
	_, exists1 := rl.limiters.Load("192.168.1.100")
	assert.True(t, exists1, "Limiter for first IP should exist")

	_, exists2 := rl.limiters.Load("192.168.1.101")
	assert.True(t, exists2, "Limiter for second IP should exist")

	// Trigger cleanup
	rl.cleanupUnusedLimiters()

	// After cleanup, limiters should be removed
	_, exists1After := rl.limiters.Load("192.168.1.100")
	assert.False(t, exists1After, "Limiter for first IP should be cleaned up")

	_, exists2After := rl.limiters.Load("192.168.1.101")
	assert.False(t, exists2After, "Limiter for second IP should be cleaned up")
}

func TestRateLimiter_Stop(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.AppConfig{
		RateLimitEnabled:           true,
		RateLimitRequestsPerMinute: 100,
		RateLimitBurst:             20,
		RateLimitTrustedProxies:    []string{},
	}

	rl := NewRateLimiter(cfg, logger)

	// Stop should not panic
	assert.NotPanics(t, func() {
		rl.Stop()
	})

	// Calling Stop multiple times should not panic
	assert.NotPanics(t, func() {
		rl.Stop()
	})
}
