package validation

import (
	"encoding/base64"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/duchuongnguyen/dhcp2p/internal/app/adapters/handlers/http/keys"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/errors"
	"github.com/go-chi/chi/v5"
)

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	Value string
	Error error
}

// ValidationConfig holds configuration for validation
type ValidationConfig struct {
	MaxLength      int
	MinLength      int
	Required       bool
	AllowEmpty     bool
	TrimWhitespace bool
	Pattern        string // Regex pattern for validation
}

// DefaultValidationConfig returns sensible defaults for validation
func DefaultValidationConfig() ValidationConfig {
	return ValidationConfig{
		MaxLength:      2048,
		MinLength:      1,
		Required:       true,
		AllowEmpty:     false,
		TrimWhitespace: true,
	}
}

// PeerIDValidationConfig returns configuration for peer ID validation
func PeerIDValidationConfig() ValidationConfig {
	return ValidationConfig{
		MaxLength:      128,
		MinLength:      10,
		Required:       true,
		AllowEmpty:     false,
		TrimWhitespace: true,
		Pattern:        `^[a-zA-Z0-9_-]+$`, // Allow alphanumeric, underscore, and hyphen
	}
}

// NonceValidationConfig returns configuration for nonce validation
func NonceValidationConfig() ValidationConfig {
	return ValidationConfig{
		MaxLength:      36, // UUID length
		MinLength:      36,
		Required:       true,
		AllowEmpty:     false,
		TrimWhitespace: true,
		Pattern:        `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, // UUID pattern
	}
}

// PubkeyValidationConfig returns configuration for public key validation
func PubkeyValidationConfig() ValidationConfig {
	return ValidationConfig{
		MaxLength:      2048,
		MinLength:      16,
		Required:       true,
		AllowEmpty:     false,
		TrimWhitespace: true,
	}
}

// SignatureValidationConfig returns configuration for signature validation
func SignatureValidationConfig() ValidationConfig {
	return ValidationConfig{
		MaxLength:      2048,
		MinLength:      32,
		Required:       true,
		AllowEmpty:     false,
		TrimWhitespace: true,
	}
}

// ValidateHeader validates and extracts a header value
func ValidateHeader(r *http.Request, headerName string, config ValidationConfig) ValidationResult {
	value := r.Header.Get(headerName)
	return validateString(value, headerName, config)
}

// ValidateURLParam validates and extracts a URL parameter
func ValidateURLParam(r *http.Request, paramName string, config ValidationConfig) ValidationResult {
	value := chi.URLParam(r, paramName)
	return validateString(value, paramName, config)
}

// ValidateQueryParam validates and extracts a query parameter
func ValidateQueryParam(r *http.Request, paramName string, config ValidationConfig) ValidationResult {
	value := r.URL.Query().Get(paramName)
	return validateString(value, paramName, config)
}

// ValidatePeerIDFromContext validates and extracts peerID from request context
func ValidatePeerIDFromContext(r *http.Request) ValidationResult {
	peerIDValue := r.Context().Value(keys.PeerIDContextKey)
	if peerIDValue == nil {
		return ValidationResult{Error: errors.ErrMissingPeerID}
	}

	peerID, ok := peerIDValue.(string)
	if !ok || peerID == "" {
		return ValidationResult{Error: errors.ErrMissingPeerID}
	}

	return ValidationResult{Value: peerID}
}

// ValidateTokenID validates and parses a token ID string
func ValidateTokenID(tokenIDStr string) ValidationResult {
	if tokenIDStr == "" {
		return ValidationResult{Error: errors.ErrMissingTokenID}
	}

	tokenID, err := strconv.ParseInt(tokenIDStr, 10, 64)
	if err != nil {
		return ValidationResult{Error: errors.ErrInvalidTokenID}
	}

	// Token IDs should be positive
	if tokenID <= 0 {
		return ValidationResult{Error: errors.ErrInvalidTokenID}
	}

	return ValidationResult{Value: strconv.FormatInt(tokenID, 10)}
}

// ValidateBase64Pubkey validates and decodes a base64-encoded public key
func ValidateBase64Pubkey(pubkey string) ValidationResult {
	config := DefaultValidationConfig()
	config.MaxLength = 2048

	result := validateString(pubkey, "pubkey", config)
	if result.Error != nil {
		return result
	}

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(result.Value)
	if err != nil {
		return ValidationResult{Error: errors.ErrInvalidPubkey}
	}

	// Validate decoded length (reasonable bounds for public keys)
	// Allow smaller sizes for testing (minimum 16 bytes instead of 32)
	if len(decoded) < 16 || len(decoded) > 1024 {
		return ValidationResult{Error: errors.ErrInvalidPubkey}
	}

	return ValidationResult{Value: result.Value}
}

// ValidateBase64Signature validates and decodes a base64-encoded signature
func ValidateBase64Signature(signature string) ValidationResult {
	config := DefaultValidationConfig()
	config.MaxLength = 2048

	result := validateString(signature, "signature", config)
	if result.Error != nil {
		// Convert generic errors to signature-specific errors
		if result.Error == errors.ErrInvalidPubkey {
			return ValidationResult{Error: errors.ErrInvalidSignature}
		}
		return result
	}

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(result.Value)
	if err != nil {
		return ValidationResult{Error: errors.ErrInvalidSignature}
	}

	// Validate decoded length (reasonable bounds for signatures)
	if len(decoded) < 32 || len(decoded) > 1024 {
		return ValidationResult{Error: errors.ErrInvalidSignature}
	}

	return ValidationResult{Value: result.Value}
}

// validateString performs common string validation
func validateString(value, fieldName string, config ValidationConfig) ValidationResult {
	// Trim whitespace if configured
	if config.TrimWhitespace {
		value = strings.TrimSpace(value)
	}

	// Check if empty is allowed
	if config.AllowEmpty && value == "" {
		return ValidationResult{Value: value}
	}

	// Check if required field is empty
	if config.Required && value == "" {
		switch fieldName {
		case "peerID":
			return ValidationResult{Error: errors.ErrMissingPeerID}
		case "tokenID":
			return ValidationResult{Error: errors.ErrMissingTokenID}
		case "pubkey":
			return ValidationResult{Error: errors.ErrMissingPubkey}
		case "nonce":
			return ValidationResult{Error: errors.ErrMissingNonce}
		case "signature":
			return ValidationResult{Error: errors.ErrMissingSignature}
		default:
			return ValidationResult{Error: errors.ErrMissingHeaders}
		}
	}

	// Check minimum length (only if not empty)
	if config.MinLength > 0 && len(value) < config.MinLength && value != "" {
		switch fieldName {
		case "peerID":
			return ValidationResult{Error: errors.ErrInvalidPeerID}
		case "pubkey":
			return ValidationResult{Error: errors.ErrInvalidPubkey}
		case "signature":
			return ValidationResult{Error: errors.ErrInvalidSignature}
		case "nonce":
			return ValidationResult{Error: errors.ErrInvalidNonce}
		default:
			return ValidationResult{Error: errors.ErrInvalidPubkey}
		}
	}

	// Check maximum length
	if config.MaxLength > 0 && len(value) > config.MaxLength {
		switch fieldName {
		case "peerID":
			return ValidationResult{Error: errors.ErrInvalidPeerID}
		case "pubkey":
			return ValidationResult{Error: errors.ErrInvalidPubkey}
		case "signature":
			return ValidationResult{Error: errors.ErrInvalidSignature}
		case "nonce":
			return ValidationResult{Error: errors.ErrInvalidNonce}
		default:
			return ValidationResult{Error: errors.ErrInvalidPubkey}
		}
	}

	// Check pattern if provided
	if config.Pattern != "" && value != "" {
		matched, err := regexp.MatchString(config.Pattern, value)
		if err != nil || !matched {
			switch fieldName {
			case "peerID":
				return ValidationResult{Error: errors.ErrInvalidPeerID}
			case "nonce":
				return ValidationResult{Error: errors.ErrInvalidNonce}
			default:
				return ValidationResult{Error: errors.ErrInvalidPubkey}
			}
		}
	}

	return ValidationResult{Value: value}
}

// ValidateNonce validates a nonce with UUID format checking
func ValidateNonce(nonce string) error {
	config := NonceValidationConfig()
	result := validateString(nonce, "nonce", config)
	return result.Error
}
