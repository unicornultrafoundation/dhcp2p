package validation

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"

	"github.com/duchuongnguyen/dhcp2p/internal/app/adapters/handlers/http/keys"
	"github.com/duchuongnguyen/dhcp2p/internal/app/adapters/handlers/http/utils"
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
	decoded, err := base64.StdEncoding.DecodeString(pubkey)
	if err != nil {
		return ValidationResult{Error: errors.ErrInvalidPubkey}
	}

	// Validate decoded length (reasonable bounds for public keys)
	if len(decoded) < 32 || len(decoded) > 1024 {
		return ValidationResult{Error: errors.ErrInvalidPubkey}
	}

	return ValidationResult{Value: pubkey}
}

// ValidateBase64Signature validates and decodes a base64-encoded signature
func ValidateBase64Signature(signature string) ValidationResult {
	config := DefaultValidationConfig()
	config.MaxLength = 2048

	result := validateString(signature, "signature", config)
	if result.Error != nil {
		return result
	}

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return ValidationResult{Error: errors.ErrInvalidSignature}
	}

	// Validate decoded length (reasonable bounds for signatures)
	if len(decoded) < 32 || len(decoded) > 1024 {
		return ValidationResult{Error: errors.ErrInvalidSignature}
	}

	return ValidationResult{Value: signature}
}

// validateString performs common string validation
func validateString(value, fieldName string, config ValidationConfig) ValidationResult {
	// Trim whitespace if configured
	if config.TrimWhitespace {
		value = strings.TrimSpace(value)
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
		default:
			return ValidationResult{Error: errors.ErrMissingHeaders}
		}
	}

	// Check if empty is allowed
	if !config.AllowEmpty && value == "" {
		return ValidationResult{Error: errors.ErrMissingHeaders}
	}

	// Check minimum length
	if config.MinLength > 0 && len(value) < config.MinLength {
		return ValidationResult{Error: errors.ErrInvalidPubkey}
	}

	// Check maximum length
	if config.MaxLength > 0 && len(value) > config.MaxLength {
		return ValidationResult{Error: errors.ErrInvalidPubkey}
	}

	return ValidationResult{Value: value}
}

// WriteValidationError writes a validation error response
func WriteValidationError(w http.ResponseWriter, result ValidationResult) {
	utils.WriteDomainError(w, result.Error)
}
