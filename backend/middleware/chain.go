package middleware

import (
	"net/http"
	"time"
)

func Chain(handler http.Handler) http.Handler {
	handler = CORS(handler)
	handler = Timeout(5 * time.Second)(handler)
	handler = Recovery(handler)

	return handler
}
