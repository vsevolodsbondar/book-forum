package handler

import (
	"log/slog"
	"net/http"

	"auth/internal/response"
)

// HeartbeatResponse is the JSON body returned by the heartbeat endpoint.
type HeartbeatResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

// Heartbeat reports that the HTTP service is responding.
// It does not check the database or other dependencies.
func Heartbeat(w http.ResponseWriter, r *http.Request) {
	data := HeartbeatResponse{
		Service: "auth",
		Status:  "ok",
	}

	if err := response.WriteJSON(w, data, http.StatusOK); err != nil {
		slog.ErrorContext(r.Context(), "write heartbeat response", "err", err)
	}
}
