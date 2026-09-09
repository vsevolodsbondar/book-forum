package server

import (
	"log/slog"
	"net/http"

	"auth/internal/response"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(
			r.Context(),
			"http request",
			"method", r.Method,
			"path", r.URL.Path,
		)

		next.ServeHTTP(w, r)
	})
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.ErrorContext(r.Context(), "panic recovered", "panic", err)
				if err := response.WriteJSON(w, response.InternalServerError("internal server error"), http.StatusInternalServerError); err != nil {
					slog.ErrorContext(r.Context(), "failed to write response", "err", err)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
