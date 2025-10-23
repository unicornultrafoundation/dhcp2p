package validation

import (
	"encoding/base64"
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-chi/chi/v5"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/errors"
)

// SanitizationConfig holds configuration for input sanitization
type SanitizationConfig struct {
	MaxLength        int
	MinLength        int
	AllowedChars     *regexp.Regexp
	RemoveWhitespace bool
	ToLowercase      bool
	ToUppercase      bool
}

// DefaultSanitizationConfig returns sensible defaults for sanitization
func DefaultSanitizationConfig() SanitizationConfig {
	return SanitizationConfig{
		MaxLength:        2048,
		MinLength:        1,
		RemoveWhitespace: true,
		ToLowercase:      false,
		ToUppercase:      false,
	}
}

// PeerIDSanitizationConfig returns configuration for peer ID sanitization
func PeerIDSanitizationConfig() SanitizationConfig {
	config := DefaultSanitizationConfig()
	config.MaxLength = 128
	config.MinLength = 1
	// Peer IDs typically contain alphanumeric characters and some special chars
	config.AllowedChars = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return config
}

// PubkeySanitizationConfig returns configuration for public key sanitization
func PubkeySanitizationConfig() SanitizationConfig {
	config := DefaultSanitizationConfig()
	config.MaxLength = 2048
	config.MinLength = 32
	// Base64 characters only
	config.AllowedChars = regexp.MustCompile(`^[A-Za-z0-9+/=]+$`)
	return config
}

// SignatureSanitizationConfig returns configuration for signature sanitization
func SignatureSanitizationConfig() SanitizationConfig {
	config := DefaultSanitizationConfig()
	config.MaxLength = 2048
	config.MinLength = 32
	// Base64 characters only
	config.AllowedChars = regexp.MustCompile(`^[A-Za-z0-9+/=]+$`)
	return config
}

// SanitizeString sanitizes a string according to the provided configuration
func SanitizeString(input string, config SanitizationConfig) (string, error) {
	if input == "" {
		if config.MinLength > 0 {
			return "", errors.ErrMissingHeaders
		}
		return "", nil
	}

	// Remove whitespace if configured
	if config.RemoveWhitespace {
		input = strings.TrimSpace(input)
	}

	// Convert case if configured
	if config.ToLowercase {
		input = strings.ToLower(input)
	} else if config.ToUppercase {
		input = strings.ToUpper(input)
	}

	// Check length constraints
	if len(input) < config.MinLength {
		return "", errors.ErrInvalidPubkey
	}
	if len(input) > config.MaxLength {
		return "", errors.ErrInvalidPubkey
	}

	// Check allowed characters if regex is provided
	if config.AllowedChars != nil && !config.AllowedChars.MatchString(input) {
		return "", errors.ErrInvalidPubkey
	}

	return input, nil
}

// SanitizeHeader sanitizes a header value
func SanitizeHeader(r *http.Request, headerName string, config SanitizationConfig) (string, error) {
	value := r.Header.Get(headerName)
	return SanitizeString(value, config)
}

// SanitizeURLParam sanitizes a URL parameter
func SanitizeURLParam(r *http.Request, paramName string, config SanitizationConfig) (string, error) {
	value := chi.URLParam(r, paramName)
	return SanitizeString(value, config)
}

// SanitizeQueryParam sanitizes a query parameter
func SanitizeQueryParam(r *http.Request, paramName string, config SanitizationConfig) (string, error) {
	value := r.URL.Query().Get(paramName)
	return SanitizeString(value, config)
}

// SanitizeBase64Data sanitizes and validates base64-encoded data
func SanitizeBase64Data(input string, config SanitizationConfig) (string, error) {
	// First sanitize the string
	sanitized, err := SanitizeString(input, config)
	if err != nil {
		return "", err
	}

	// Validate base64 encoding
	_, err = base64.StdEncoding.DecodeString(sanitized)
	if err != nil {
		return "", errors.ErrInvalidPubkey
	}

	return sanitized, nil
}

// SanitizePeerID sanitizes a peer ID
func SanitizePeerID(peerID string) (string, error) {
	return SanitizeString(peerID, PeerIDSanitizationConfig())
}

// SanitizePubkey sanitizes a public key
func SanitizePubkey(pubkey string) (string, error) {
	return SanitizeBase64Data(pubkey, PubkeySanitizationConfig())
}

// SanitizeSignature sanitizes a signature
func SanitizeSignature(signature string) (string, error) {
	return SanitizeBase64Data(signature, SignatureSanitizationConfig())
}

// SanitizeTokenID sanitizes and validates a token ID string
func SanitizeTokenID(tokenIDStr string) (string, error) {
	config := DefaultSanitizationConfig()
	config.MaxLength = 20 // Reasonable limit for int64 string representation
	config.MinLength = 1
	config.AllowedChars = regexp.MustCompile(`^[0-9]+$`) // Only digits

	sanitized, err := SanitizeString(tokenIDStr, config)
	if err != nil {
		return "", err
	}

	// Additional validation: ensure it's a valid int64
	if len(sanitized) > 19 { // Max digits for int64
		return "", errors.ErrInvalidTokenID
	}

	return sanitized, nil
}

// RemoveControlCharacters removes control characters from input
func RemoveControlCharacters(input string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\t' && r != '\n' && r != '\r' {
			return -1 // Remove character
		}
		return r
	}, input)
}

// ValidateAndSanitizeRequest performs comprehensive validation and sanitization
func ValidateAndSanitizeRequest(r *http.Request) error {
	// Check for suspicious patterns in headers
	for name, values := range r.Header {
		for _, value := range values {
			// Check for potential injection attempts
			if strings.Contains(strings.ToLower(value), "<script") ||
				strings.Contains(strings.ToLower(value), "javascript:") ||
				strings.Contains(strings.ToLower(value), "onload=") ||
				strings.Contains(strings.ToLower(value), "onerror=") {
				return errors.ErrInvalidPubkey
			}

			// Check header length
			if len(value) > 8192 { // 8KB limit per header
				return errors.ErrInvalidPubkey
			}
		}

		// Check header name length
		if len(name) > 256 {
			return errors.ErrInvalidPubkey
		}
	}

	// Check URL length
	if len(r.URL.String()) > 8192 {
		return errors.ErrInvalidPubkey
	}

	return nil
}
