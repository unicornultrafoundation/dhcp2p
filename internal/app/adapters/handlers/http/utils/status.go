package utils

import (
	"errors"
	"net/http"

	domainErrors "github.com/duchuongnguyen/dhcp2p/internal/app/domain/errors"
)

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
