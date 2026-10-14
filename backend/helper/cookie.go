package helper

import (
	"errors"
	"fmt"
	"forum_backend/custom_err"
	"net/http"
	"time"
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

const sessionCookieName = "session_id"

func SetCookie(w http.ResponseWriter, token string, expiresAt int64) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  time.Unix(expiresAt, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

}
