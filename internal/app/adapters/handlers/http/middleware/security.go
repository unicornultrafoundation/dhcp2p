package middleware

import (
	"net/http"

	"github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/handlers/http/utils"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/handlers/http/validation"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/errors"
)

// SecurityMiddleware provides comprehensive request validation and sanitization
func SecurityMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Validate and sanitize the request
			if err := validation.ValidateAndSanitizeRequest(r); err != nil {
				utils.WriteDomainError(w, err)
				return
			}

			// Continue to the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// RequestSizeMiddleware limits the size of incoming requests
func RequestSizeMiddleware(maxSize int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit request body size
			if r.ContentLength > maxSize {
				utils.WriteDomainError(w, errors.ErrInvalidPubkey) // Use appropriate error
				return
			}

			// Limit request size (including headers)
			if r.ContentLength > 0 {
				r.Body = http.MaxBytesReader(w, r.Body, maxSize)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersMiddleware adds comprehensive security headers to responses
func SecurityHeadersMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Prevent MIME type sniffing
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// Prevent clickjacking
			w.Header().Set("X-Frame-Options", "DENY")

			// XSS protection
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Referrer policy
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Content Security Policy - more permissive for API usage
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none';")

			// Strict Transport Security (only for HTTPS)
			if r.TLS != nil {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			}

			// Permissions Policy (formerly Feature Policy)
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), speaker=(), vibrate=(), fullscreen=(self), sync-xhr=()")

			// Cross-Origin policies
			w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
			w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
			w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")

			// Cache control for sensitive endpoints
			if r.URL.Path == "/request-auth" || r.URL.Path == "/allocate-ip" ||
				r.URL.Path == "/renew-lease" || r.URL.Path == "/release-lease" {
				w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware adds CORS headers for cross-origin requests
func CORSMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Pubkey, X-Nonce, X-Signature")
			w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CombinedSecurityMiddleware combines all security middlewares
func CombinedSecurityMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return CORSMiddleware()(
			SecurityHeadersMiddleware()(
				RequestSizeMiddleware(1024 * 1024)( // 1MB limit
					SecurityMiddleware()(next),
				),
			),
		)
	}
}
