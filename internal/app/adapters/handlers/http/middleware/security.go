package middleware

import (
	"net/http"

	"github.com/duchuongnguyen/dhcp2p/internal/app/adapters/handlers/http/utils"
	"github.com/duchuongnguyen/dhcp2p/internal/app/adapters/handlers/http/validation"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/errors"
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

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")

			next.ServeHTTP(w, r)
		})
	}
}

// CombinedSecurityMiddleware combines all security middlewares
func CombinedSecurityMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return SecurityHeadersMiddleware()(
			RequestSizeMiddleware(1024 * 1024)( // 1MB limit
				SecurityMiddleware()(next),
			),
		)
	}
}
