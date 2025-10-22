package http

import (
	"encoding/json"
	"net/http"
)

func writeErrorResponse(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
	err = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeResponse(w http.ResponseWriter, statusCode int, data interface{}) {
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
