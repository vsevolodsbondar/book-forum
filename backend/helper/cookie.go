package helper

import (
	"errors"
	"fmt"
	"forum_backend/custom_err"
	"net/http"
	"time"
)

const sessionCookieName = "token"

func ExtractSessionCookie(r *http.Request) (string, error) {
	sessionCookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", fmt.Errorf("%w: user is unauthorized", custom_err.ErrForbidden)

		}

		return "", custom_err.ErrBadRequest
	}

	return sessionCookie.Value, nil
}

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
