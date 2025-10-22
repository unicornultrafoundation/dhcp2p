package http

import (
	"encoding/json"
	"errors"
	"net/http"

	domainErrors "github.com/duchuongnguyen/dhcp2p/internal/app/domain/errors"
)

func writeErrorResponse(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ParseRequestBody parses the request body into the given struct
func ParseRequestBody(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// mapErrorToStatus returns appropriate HTTP status code for domain errors.
func mapErrorToStatus(err error) int {
	switch {
	case errors.Is(err, domainErrors.ErrMissingHeaders),
		errors.Is(err, domainErrors.ErrMissingPubkey),
		errors.Is(err, domainErrors.ErrMissingPeerID),
		errors.Is(err, domainErrors.ErrMissingTokenID),
		errors.Is(err, domainErrors.ErrInvalidTokenID),
		errors.Is(err, domainErrors.ErrInvalidPubkey):
		return http.StatusBadRequest

	case errors.Is(err, domainErrors.ErrInvalidSignature):
		return http.StatusUnauthorized

	case errors.Is(err, domainErrors.ErrPubkeyMismatch):
		return http.StatusForbidden

	case errors.Is(err, domainErrors.ErrLeaseNotFound):
		return http.StatusNotFound

	default:
		return http.StatusInternalServerError
	}
}

// writeDomainError maps domain error to HTTP status and writes JSON error.
func writeDomainError(w http.ResponseWriter, err error) {
	writeErrorResponse(w, mapErrorToStatus(err), err)
}
