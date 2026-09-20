package custom_err

import "errors"

type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// in places where error occurs we can add some context to base error:
// return fmt.Errorf("%w: %w", ErrInvalidInput, err) OR
// return just base error without context "return error.ErrInvalidInput"
var (
	ErrPostNotFound     = errors.New("post not found")
	ErrCommentNotFound  = errors.New("comment not found")
	ErrCategoryNotFound = errors.New("category not found")
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidInput     = errors.New("invalid input")
	ErrExpiredSession   = errors.New("session expired")
	//probably should be more specific, but atm do not know it yet
	ErrAuthService       = errors.New("error occured on auth side")
	ErrJSONDecodeFailed  = errors.New("can't decode JSON")
	ErrUserRegisterError = errors.New("can't register user")
	ErrWrongCredentials  = errors.New("invalid credentials")
	ErrInvalidToken      = errors.New("invalid token")
	ErrInvalidSession    = errors.New("invalid session")
	ErrBadRequest        = errors.New("failed to process request")
)
