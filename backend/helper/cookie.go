package helper

import (
	"errors"
	"fmt"
	"forum_backend/custom_err"
	"net/http"
)

func ExtractSessionCookie(r *http.Request) (string, error) {
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", fmt.Errorf("%w: user is unauthorized", custom_err.ErrForbidden)

		}

		return "", custom_err.ErrBadRequest
	}

	return sessionCookie.Value, nil
}
