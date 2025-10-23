package utils

import (
	"encoding/json"
	"net/http"
)

func WriteErrorResponse(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func WriteResponse(w http.ResponseWriter, statusCode int, data interface{}) {
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

// writeDomainError maps domain error to HTTP status and writes JSON error.
func WriteDomainError(w http.ResponseWriter, err error) {
	WriteErrorResponse(w, mapErrorToStatus(err), err)
}

// WriteSuccessResponse writes a successful response
func WriteSuccessResponse(w http.ResponseWriter, data interface{}) {
	WriteResponse(w, http.StatusOK, data)
}
