package client

import (
	"context"
	"errors"
	"time"
)

type AuthStatusError struct {
	Status int
	Err    error
}

// for AuthStatusError to match ERROR interface
func (e *AuthStatusError) Error() string {
	return e.Err.Error()
}

// for error.IS and error.AS to be able to open the struct
func (e *AuthStatusError) Unwrap() error {
	return e.Err
}

func ShouldRetry(err error, safeToRetry bool) bool {
	//user has left no need to retry
	if errors.Is(err, context.Canceled) {
		return false
	}

	var statusErr *AuthStatusError
	if errors.As(err, &statusErr) {
		if statusErr.Status > 500 {
			return safeToRetry
		}
		return false
	}

	//if no response from auth AT ALL
	return false
}

const (
	maxAttempts    = 3
	attemptTimeout = 1 * time.Second
)

func buildBackoff(attempt int) time.Duration {
	delay := 100 * time.Millisecond

	//bigger delay for every attempt
	for i := 0; i < attempt; i++ {
		delay = delay * 2
	}
	return delay
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
