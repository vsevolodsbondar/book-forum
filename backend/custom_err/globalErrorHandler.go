package custom_err

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type AppHandler func(w http.ResponseWriter, r *http.Request) error

// handlers will be wrapped with it like this:
// mux.HandleFunc("POST /api/comment", error.GlobalErrorHandler(commentHandler.PostComment))
func GlobalErrorHandler(h AppHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			handleError(w, err)
		}
	}
}

func handleError(w http.ResponseWriter, err error) {
	errResp := ErrorResponse{}

	switch {
	case errors.Is(err, ErrExpiredSession),
		errors.Is(err, ErrSessionValidationFail):
		errResp.Status = http.StatusUnauthorized
		errResp.Message = err.Error()

	case errors.Is(err, ErrPostNotFound),
		errors.Is(err, ErrCommentNotFound),
		errors.Is(err, ErrCategoryNotFound),
		errors.Is(err, ErrUserNotFound):
		errResp.Status = http.StatusNotFound
		errResp.Message = err.Error()

	case errors.Is(err, ErrInvalidInput),
		errors.Is(err, ErrUserRegisterError),
		errors.Is(err, ErrNoChange),
		errors.Is(err, ErrInvalidID),
		errors.Is(err, ErrEmptyUsername),
		errors.Is(err, ErrUsernameTooLong),
		errors.Is(err, ErrDescTooLong),
		errors.Is(err, ErrBadRequest):
		errResp.Status = http.StatusBadRequest
		errResp.Message = err.Error()

	case errors.Is(err, ErrWrongCredentials):
		errResp.Status = http.StatusUnauthorized
		errResp.Message = err.Error()
	case errors.Is(err, ErrForbidden):
		errResp.Status = http.StatusForbidden
		errResp.Message = err.Error()

	case errors.Is(err, ErrGetRowsAffected),
		errors.Is(err, ErrGetUser),
		errors.Is(err, ErrCreateUser),
		errors.Is(err, ErrUpdateUser),
		errors.Is(err, ErrDeleteUser):
		errResp.Status = http.StatusInternalServerError
		errResp.Message = err.Error()

	default:
		log.Printf("internal error: %v", err)
		errResp.Status = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errResp.Status)

	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		log.Printf("failed to encode error response: %v", err)
	}
}
