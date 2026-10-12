package server

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"auth/internal/database"
	"auth/internal/handler"
	"auth/internal/repository"
	"auth/internal/service"
)

func TestRegisterRoutesHealth(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	healthRepo := repository.NewHealthRepository(db)
	healthService := service.NewHealthService(healthRepo)
	healthHandler := handler.NewHealthHandler(healthService)

	mux := http.NewServeMux()
	registerRoutes(mux, healthHandler, handler.NewUserHandler(nil, nil), handler.NewSessionHandler(nil, nil))

	tests := []struct {
		name   string
		path   string
		status int
	}{
		{
			name:   "liveness",
			path:   "/health/live",
			status: http.StatusOK,
		},
		{
			name:   "readiness",
			path:   "/health/ready",
			status: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			recorder := httptest.NewRecorder()

			mux.ServeHTTP(recorder, request)

			if recorder.Code != tt.status {
				t.Fatalf("got status %d, want %d", recorder.Code, tt.status)
			}
		})
	}
}
