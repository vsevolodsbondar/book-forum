package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// 1st layer: MW factory (normal function that returns a middleware)
func Timeout(duration time.Duration) func(http.Handler) http.Handler {
	// 2nd layer: Middleware that returns a handler function
	return func(next http.Handler) http.Handler {
		//3rd layer: Handler function itself that adds logic to the chain of handlers
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			runWithTimeOut(w, r, next, duration)
		})
	}
}

// the middlewarelogic
func runWithTimeOut(w http.ResponseWriter, r *http.Request, next http.Handler, duration time.Duration) {
	//add timer to r's context and create CancelFunc
	ctx, cancel := context.WithTimeout(r.Context(), duration)
	//cancels the timer in case of success so it does not stay in memory
	defer cancel()

	//prep. channell for a signall of successfull operation
	done := make(chan struct{})

	//execute the request in parallell and fill the "done" channel when done
	go func() {
		defer close(done)
		next.ServeHTTP(w, r.WithContext(ctx))
	}()

	//see which channell gets filled first: 1.Operation success(done) OR 2. Time is out (ctx.Done)
	select {
	case <-done: //success
		return
	case <-ctx.Done(): //time is out
		writeTimeOutResponse(w, r)
	}
}

func writeTimeOutResponse(w http.ResponseWriter, r *http.Request) {
	slog.WarnContext(r.Context(), "request timed out", "path", r.URL.Path)
	http.Error(w, "request timed out", http.StatusGatewayTimeout)
}
