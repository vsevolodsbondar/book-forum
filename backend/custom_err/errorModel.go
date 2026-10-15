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
	// input validation related status 400
	ErrInvalidID         = errors.New("id must be a positive integer")
	ErrEmptyUsername     = errors.New("username cannot be empty")
	ErrUsernameTooLong   = errors.New("username must be at most 30 characters")
	ErrUsernameTooShort  = errors.New("username must be at least 3 characters")
	ErrInvalidUsername   = errors.New("username must only contain alphanumeric characters and underscores")
	ErrUnchangedUsername = errors.New("username unchanged")
	ErrDescTooLong       = errors.New("description must be at most 500 characters")
	// db malfunction related status 500
	ErrGetRowsAffected       = errors.New("failed to get rows affected")
	ErrGetUser               = errors.New("failed to get user")
	ErrGetUsers              = errors.New("failed to get users")
	ErrCreateUser            = errors.New("failed to create user")
	ErrUpdateUser            = errors.New("failed to update user")
	ErrDeleteUser            = errors.New("failed to delete user")
	ErrSessionValidationFail = errors.New("failed session validation")
	ErrCreateReaction        = errors.New("failed to create reaction")
	ErrUpdateReaction        = errors.New("failed to update reaction")
	ErrDeleteReaction        = errors.New("failed to delete reaction")
	// not found status 404
	ErrPostNotFound     = errors.New("post not found")
	ErrCommentNotFound  = errors.New("comment not found")
	ErrCategoryNotFound = errors.New("category not found")
	ErrUserNotFound     = errors.New("user not found")
	ErrReactionNotFound = errors.New("reaction not found")
	// probably should be more specific, but atm do not know it yet
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
	ErrNoChange      = errors.New("the text wasn't changed")
	ErrForbidden     = errors.New("forbidden")
	ErrConflict      = errors.New("already exists")
	ErrTakenUsername = errors.New("username is taken")
)
