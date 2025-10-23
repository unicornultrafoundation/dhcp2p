package utils

import (
	"encoding/json"
	"net/http"

	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/errors"
)

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse represents a successful response
type SuccessResponse struct {
	Data interface{} `json:"data"`
}

// WriteErrorResponse writes a structured error response
func WriteErrorResponse(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	var appErr *errors.AppError
	if errors.IsAppError(err) {
		appErr = errors.GetAppError(err)
	} else {
		// Wrap unknown errors as internal errors
		appErr = errors.WrapError(err, errors.ErrorTypeInternal, "UNKNOWN_ERROR", "An unexpected error occurred")
	}

	w.WriteHeader(appErr.HTTPStatus())

	errorResp := ErrorResponse{
		Type:    string(appErr.Type),
		Code:    appErr.Code,
		Message: appErr.Message,
		Details: appErr.Details,
	}

	if encodeErr := json.NewEncoder(w).Encode(errorResp); encodeErr != nil {
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

// WriteSuccessResponse writes a successful response
func WriteSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := SuccessResponse{Data: data}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// WriteResponse writes a response with custom status code
func WriteResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// ParseRequestBody parses the request body into the given struct
func ParseRequestBody(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// WriteDomainError maps domain error to HTTP status and writes JSON error
func WriteDomainError(w http.ResponseWriter, err error) {
	WriteErrorResponse(w, err)
}
