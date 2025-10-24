package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/handlers/http/utils"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/errors"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/infrastructure/config"
)

// RateLimiter manages rate limiting for HTTP requests
type RateLimiter struct {
	config        *config.AppConfig
	logger        *zap.Logger
	limiters      sync.Map // map[string]*rate.Limiter
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(cfg *config.AppConfig, logger *zap.Logger) *RateLimiter {
	rl := &RateLimiter{
		config:      cfg,
		logger:      logger,
		stopCleanup: make(chan struct{}),
	}

	// Start cleanup goroutine to remove unused limiters
	rl.startCleanup()

	return rl
}

// startCleanup starts a background goroutine to clean up unused limiters
func (rl *RateLimiter) startCleanup() {
	rl.cleanupTicker = time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-rl.cleanupTicker.C:
				rl.cleanupUnusedLimiters()
			case <-rl.stopCleanup:
				rl.cleanupTicker.Stop()
				return
			}
		}
	}()
}

// Stop stops the rate limiter and cleans up resources
func (rl *RateLimiter) Stop() {
	select {
	case <-rl.stopCleanup:
		// Already closed
	default:
		close(rl.stopCleanup)
	}
}

// cleanupUnusedLimiters removes limiters that haven't been used recently
func (rl *RateLimiter) cleanupUnusedLimiters() {
	// Simple cleanup: remove all limiters periodically
	// In a production system, you'd want to track last access time per limiter
	// For now, we'll just clear all limiters every 5 minutes to prevent memory leaks
	rl.limiters.Range(func(key, value interface{}) bool {
		rl.limiters.Delete(key)
		return true
	})
}

// extractClientIP extracts the client IP from the request, considering proxy headers
func (rl *RateLimiter) extractClientIP(r *http.Request) string {
	// Check X-Real-IP header first (if from trusted proxy)
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		if rl.isTrustedProxy(r.RemoteAddr) {
			if ip := rl.parseIP(realIP); ip != "" {
				return ip
			}
		}
	}

	// Check X-Forwarded-For header (if from trusted proxy)
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		if rl.isTrustedProxy(r.RemoteAddr) {
			// X-Forwarded-For can contain multiple IPs, take the first one
			ips := strings.Split(forwardedFor, ",")
			if len(ips) > 0 {
				if ip := rl.parseIP(strings.TrimSpace(ips[0])); ip != "" {
					return ip
				}
			}
		}
	}

	// Fallback to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If SplitHostPort fails, try parsing as IP directly
		if net.ParseIP(r.RemoteAddr) != nil {
			return r.RemoteAddr
		}
		return "unknown"
	}
	return ip
}

// isTrustedProxy checks if the given IP is in the list of trusted proxies
func (rl *RateLimiter) isTrustedProxy(proxyIP string) bool {
	if len(rl.config.RateLimitTrustedProxies) == 0 {
		return false
	}

	ip, _, err := net.SplitHostPort(proxyIP)
	if err != nil {
		ip = proxyIP
	}

	for _, trustedProxy := range rl.config.RateLimitTrustedProxies {
		if trustedProxy == ip {
			return true
		}
		// Check if it's a CIDR block
		if _, network, err := net.ParseCIDR(trustedProxy); err == nil {
			if network.Contains(net.ParseIP(ip)) {
				return true
			}
		}
	}
	return false
}

// parseIP validates and returns a valid IP address
func (rl *RateLimiter) parseIP(ipStr string) string {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return ""
	}
	return ip.String()
}

// getOrCreateLimiter gets an existing limiter for the IP or creates a new one
func (rl *RateLimiter) getOrCreateLimiter(clientIP string) *rate.Limiter {
	// Try to load existing limiter first
	if value, exists := rl.limiters.Load(clientIP); exists {
		return value.(*rate.Limiter)
	}

	// Create new limiter with token bucket algorithm
	// Rate is requests per minute, burst is the maximum burst capacity
	ratePerSecond := float64(rl.config.RateLimitRequestsPerMinute) / 60.0
	newLimiter := rate.NewLimiter(rate.Limit(ratePerSecond), rl.config.RateLimitBurst)

	// Try to store the new limiter, but if another goroutine beat us to it,
	// use the existing one
	if actual, loaded := rl.limiters.LoadOrStore(clientIP, newLimiter); loaded {
		// Another goroutine created a limiter, use that one
		return actual.(*rate.Limiter)
	}

	// Our limiter was stored successfully
	return newLimiter
}

// Allow checks if the request should be allowed based on rate limiting
func (rl *RateLimiter) Allow(r *http.Request) (allowed bool, retryAfter time.Duration, remaining int) {
	if !rl.config.RateLimitEnabled {
		return true, 0, rl.config.RateLimitRequestsPerMinute
	}

	clientIP := rl.extractClientIP(r)
	limiter := rl.getOrCreateLimiter(clientIP)

	// Check if request is allowed
	now := time.Now()
	if !limiter.AllowN(now, 1) {
		// Rate limit exceeded - calculate when next token will be available
		reservation := limiter.ReserveN(now, 1)
		retryAfter = reservation.DelayFrom(now)
		return false, retryAfter, 0
	}

	// Calculate remaining tokens
	tokens := limiter.TokensAt(now)
	remaining = int(tokens)
	if remaining < 0 {
		remaining = 0
	}

	return true, 0, remaining
}

// RateLimitMiddleware creates a middleware that enforces rate limiting
func RateLimitMiddleware(cfg *config.AppConfig, logger *zap.Logger) func(next http.Handler) http.Handler {
	rateLimiter := NewRateLimiter(cfg, logger)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, retryAfter, remaining := rateLimiter.Allow(r)

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(cfg.RateLimitRequestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

			// Calculate reset time (next minute)
			resetTime := time.Now().Add(time.Minute).Unix()
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

			if !allowed {
				// Rate limit exceeded
				w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
				utils.WriteDomainError(w, errors.ErrRateLimitExceeded)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
