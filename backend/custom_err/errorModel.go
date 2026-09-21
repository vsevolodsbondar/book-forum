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
	//input validation related
	ErrInvalidID       = errors.New("id must be a positive integer")
	ErrEmptyUsername   = errors.New("username cannot be empty")
	ErrUsernameTooLong = errors.New("username must be at most 32 characters")
	ErrDescTooLong     = errors.New("description must be at most 500 characters")
	//db related
	ErrGetRowsAffected  = errors.New("failed to get rows affected")
	ErrGetUser          = errors.New("failed to get user")
	ErrCreateUser       = errors.New("failed to create user")
	ErrUpdateUser       = errors.New("failed to update user")
	ErrDeleteUser       = errors.New("failed to delete user")
	ErrPostNotFound     = errors.New("post not found")
	ErrCommentNotFound  = errors.New("comment not found")
	ErrCategoryNotFound = errors.New("category not found")
	ErrUserNotFound     = errors.New("user not found")
	//probably should be more specific, but atm do not know it yet
	ErrInvalidInput      = errors.New("invalid input")
	ErrExpiredSession    = errors.New("session expired")
	ErrAuthService       = errors.New("error occured on auth side")
	ErrJSONDecodeFailed  = errors.New("can't decode JSON")
	ErrUserRegisterError = errors.New("can't register user")
	ErrWrongCredentials  = errors.New("invalid credentials")
	ErrInvalidToken      = errors.New("invalid token")
	ErrInvalidSession    = errors.New("invalid session")
	ErrBadRequest        = errors.New("failed to process request")
	// more errors
	ErrNoChange  = errors.New("the text wasn't changed")
	ErrForbidden = errors.New("forbidden")
)
