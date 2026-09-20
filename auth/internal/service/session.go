package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
	"uuid"

	"auth/internal/password"
	"auth/internal/repository"
)

// ErrInvalidCredentials is returned when login credentials do not match a user.
var ErrInvalidCredentials = errors.New("invalid credentials")

// SessionService handles login and session lifecycle operations.
type SessionService struct {
	userRepo        *repository.UserRepository
	sessionRepo     *repository.SessionRepository
	hasher          *password.Hasher
	sessionLifetime time.Duration
	dummyHash       string
}

// LoginSession contains the session data returned after login.
type LoginSession struct {
	ID        string
	Token     string
	ExpiresAt int64
}

// LoginResult contains the authenticated user and newly created session.
type LoginResult struct {
	User    repository.User
	Session LoginSession
}

// NewSessionService creates a session service and prepares its reusable dummy hash.
func NewSessionService(userRepo *repository.UserRepository, sessionRepo *repository.SessionRepository, hasher *password.Hasher, sessionLifetime time.Duration) (*SessionService, error) {
	dummyHash, err := hasher.Hash("dummy password")
	if err != nil {
		return nil, fmt.Errorf("create dummy password hash: %w", err)
	}

	return &SessionService{
		userRepo:        userRepo,
		sessionRepo:     sessionRepo,
		hasher:          hasher,
		sessionLifetime: sessionLifetime,
		dummyHash:       dummyHash,
	}, nil
}

// Login verifies credentials and creates a new session.
func (s *SessionService) Login(ctx context.Context, email, passwordValue string) (LoginResult, error) {
	credentials, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return LoginResult{}, err
		}

		if _, err := s.hasher.Verify(passwordValue, s.dummyHash); err != nil {
			return LoginResult{}, fmt.Errorf("verify dummy password: %w", err)
		}

		return LoginResult{}, ErrInvalidCredentials
	}

	matches, err := s.hasher.Verify(passwordValue, credentials.PasswordHash)
	if err != nil {
		return LoginResult{}, fmt.Errorf("verify password: %w", err)
	}
	if !matches {
		return LoginResult{}, ErrInvalidCredentials
	}

	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return LoginResult{}, fmt.Errorf("generate session secret: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(s.sessionLifetime).Unix()
	tokenHash := sha256.Sum256(secret)

	session := repository.Session{
		ID:         uuid.New().String(),
		UserID:     credentials.ID,
		TokenHash:  tokenHash[:],
		CreatedAt:  now.Unix(),
		LastSeenAt: now.Unix(),
		ExpiresAt:  expiresAt,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		User: credentials.User,
		Session: LoginSession{
			ID:        session.ID,
			Token:     base64.RawURLEncoding.EncodeToString(secret),
			ExpiresAt: expiresAt,
		},
	}, nil
}
