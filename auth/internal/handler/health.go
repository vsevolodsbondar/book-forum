package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"auth/internal/response"
	"auth/internal/service"
)

// HealthHandler handles service health requests.
type HealthHandler struct {
	healthService *service.HealthService
}

// NewHealthHandler creates a health handler.
func NewHealthHandler(healthService *service.HealthService) *HealthHandler {
	return &HealthHandler{healthService: healthService}
}

// HealthCheck contains the result of one named health check.
type HealthCheck struct {
	Status string `json:"status"`
}

// HealthResponse is the JSON body returned by health endpoints.
type HealthResponse struct {
	Service string                 `json:"service"`
	Status  string                 `json:"status"`
	Checks  map[string]HealthCheck `json:"checks,omitempty"`
}

// Live reports that the HTTP server is responding without checking dependencies.
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	data := HealthResponse{
		Service: "auth",
		Status:  "ok",
	}

	if err := response.WriteJSON(w, data, http.StatusOK); err != nil {
		slog.ErrorContext(r.Context(), "write liveness response", "err", err)
	}
}

// Ready reports whether required dependencies are available.
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	result := h.healthService.Check(ctx)
	data := toHealthResponse(result)

	httpStatus := http.StatusOK
	if !result.Ready {
		httpStatus = http.StatusServiceUnavailable
	}

	if err := response.WriteJSON(w, data, httpStatus); err != nil {
		slog.ErrorContext(r.Context(), "write readiness response", "err", err)
	}
}

func toHealthResponse(result service.HealthResult) HealthResponse {
	status := "ok"
	if !result.Ready {
		status = "unavailable"
	}

	checks := make(map[string]HealthCheck, len(result.Checks))
	for name, check := range result.Checks {
		checks[name] = HealthCheck{Status: check.Status}
	}

	return HealthResponse{
		Service: "auth",
		Status:  status,
		Checks:  checks,
	}
}
