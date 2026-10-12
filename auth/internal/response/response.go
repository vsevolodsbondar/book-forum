package response

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse is the shared JSON error body.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains the public error code and message.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewError creates the shared JSON error body.
func NewError(code, message string) ErrorResponse {
	return ErrorResponse{
		Error: ErrorDetail{Code: code, Message: message},
	}
}

// WriteJSON writes the status and JSON body, returning encoding or write errors.
// Headers are committed before encoding, so callers should not send a fallback response.
func WriteJSON(w http.ResponseWriter, data any, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}
