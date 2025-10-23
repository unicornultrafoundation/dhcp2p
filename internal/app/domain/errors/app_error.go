package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorType represents the type of error
type ErrorType string

const (
	ErrorTypeValidation ErrorType = "validation_error"
	ErrorTypeAuth       ErrorType = "auth_error"
	ErrorTypeNotFound   ErrorType = "not_found"
	ErrorTypeConflict   ErrorType = "conflict"
	ErrorTypeInternal   ErrorType = "internal_error"
	ErrorTypeRateLimit  ErrorType = "rate_limit_error"
	ErrorTypeBadRequest ErrorType = "bad_request"
)

// AppError represents a structured application error
type AppError struct {
	Type    ErrorType `json:"type"`
	Code    string    `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
	Cause   error     `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *AppError) Unwrap() error {
	return e.Cause
}

// HTTPStatus returns the appropriate HTTP status code
func (e *AppError) HTTPStatus() int {
	switch e.Type {
	case ErrorTypeValidation, ErrorTypeBadRequest:
		return http.StatusBadRequest
	case ErrorTypeAuth:
		return http.StatusUnauthorized
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeRateLimit:
		return http.StatusTooManyRequests
	case ErrorTypeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// NewAppError creates a new application error
func NewAppError(errorType ErrorType, code, message string, cause error) *AppError {
	return &AppError{
		Type:    errorType,
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// NewValidationError creates a validation error
func NewValidationError(code, message string, cause error) *AppError {
	return NewAppError(ErrorTypeValidation, code, message, cause)
}

// NewAuthError creates an authentication error
func NewAuthError(code, message string, cause error) *AppError {
	return NewAppError(ErrorTypeAuth, code, message, cause)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(code, message string, cause error) *AppError {
	return NewAppError(ErrorTypeNotFound, code, message, cause)
}

// NewConflictError creates a conflict error
func NewConflictError(code, message string, cause error) *AppError {
	return NewAppError(ErrorTypeConflict, code, message, cause)
}

// NewInternalError creates an internal error
func NewInternalError(code, message string, cause error) *AppError {
	return NewAppError(ErrorTypeInternal, code, message, cause)
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(code, message string, cause error) *AppError {
	return NewAppError(ErrorTypeRateLimit, code, message, cause)
}

// WrapError wraps an existing error with additional context
func WrapError(err error, errorType ErrorType, code, message string) *AppError {
	return &AppError{
		Type:    errorType,
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError extracts AppError from an error
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}

// Predefined application errors
var (
	// Validation errors
	ErrMissingPeerID      = NewValidationError("MISSING_PEER_ID", "Peer ID is required", nil)
	ErrMissingTokenID     = NewValidationError("MISSING_TOKEN_ID", "Token ID is required", nil)
	ErrMissingPubkey      = NewValidationError("MISSING_PUBKEY", "Public key is required", nil)
	ErrMissingNonce       = NewValidationError("MISSING_NONCE", "Nonce is required", nil)
	ErrMissingSignature   = NewValidationError("MISSING_SIGNATURE", "Signature is required", nil)
	ErrMissingHeaders     = NewValidationError("MISSING_HEADERS", "Required headers are missing", nil)
	ErrInvalidPeerID      = NewValidationError("INVALID_PEER_ID", "Invalid peer ID format", nil)
	ErrInvalidTokenID     = NewValidationError("INVALID_TOKEN_ID", "Invalid token ID format", nil)
	ErrInvalidPubkey      = NewValidationError("INVALID_PUBKEY", "Invalid public key format", nil)
	ErrInvalidNonce       = NewValidationError("INVALID_NONCE", "Invalid nonce format", nil)
	ErrInvalidSignature   = NewValidationError("INVALID_SIGNATURE", "Invalid signature format", nil)
	ErrInvalidRequest     = NewValidationError("INVALID_REQUEST", "Invalid request format", nil)
	ErrInvalidContentType = NewValidationError("INVALID_CONTENT_TYPE", "Invalid content type", nil)
	ErrRequestTooLarge    = NewValidationError("REQUEST_TOO_LARGE", "Request size exceeds limit", nil)
	ErrInvalidURL         = NewValidationError("INVALID_URL", "Invalid URL format", nil)
	ErrInvalidHeader      = NewValidationError("INVALID_HEADER", "Invalid header format", nil)

	// Authentication errors
	ErrNonceExpired          = NewAuthError("NONCE_EXPIRED", "Nonce has expired", nil)
	ErrNonceNotFound         = NewAuthError("NONCE_NOT_FOUND", "Nonce not found", nil)
	ErrNonceUsed             = NewAuthError("NONCE_USED", "Nonce has already been used", nil)
	ErrPubkeyMismatch        = NewAuthError("PUBKEY_MISMATCH", "Public key mismatch", nil)
	ErrSignatureVerification = NewAuthError("SIGNATURE_VERIFICATION_FAILED", "Signature verification failed", nil)

	// Not found errors
	ErrLeaseNotFound    = NewNotFoundError("LEASE_NOT_FOUND", "Lease not found", nil)
	ErrNonceNotFoundErr = NewNotFoundError("NONCE_NOT_FOUND", "Nonce not found", nil)

	// Conflict errors
	ErrLeaseAlreadyExists = NewConflictError("LEASE_ALREADY_EXISTS", "Lease already exists", nil)
	ErrLeaseExpired       = NewConflictError("LEASE_EXPIRED", "Lease has expired", nil)

	// Internal errors
	ErrDatabaseConnection  = NewInternalError("DATABASE_CONNECTION_FAILED", "Database connection failed", nil)
	ErrRedisConnection     = NewInternalError("REDIS_CONNECTION_FAILED", "Redis connection failed", nil)
	ErrMissingDependencies = NewInternalError("MISSING_DEPENDENCIES", "Missing required dependencies", nil)
	ErrAllocationFailed    = NewInternalError("ALLOCATION_FAILED", "Failed to allocate lease", nil)

	// Rate limit errors
	ErrRateLimitExceeded = NewRateLimitError("RATE_LIMIT_EXCEEDED", "Rate limit exceeded", nil)
)
