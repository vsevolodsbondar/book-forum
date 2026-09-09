package response

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse is the shared JSON error body.
type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// WriteJSON writes the status and JSON body, returning encoding or write errors.
// Headers are committed before encoding, so callers should not send a fallback response.
func WriteJSON(w http.ResponseWriter, data any, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

// InternalServerError builds a status-500 error body; it does not send a response.
func InternalServerError(message string) *ErrorResponse {
	return &ErrorResponse{
		Message: message,
		Code:    http.StatusInternalServerError,
	}
}
