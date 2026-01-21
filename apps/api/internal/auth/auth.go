// Package auth provides authentication service implementation.
package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"sync"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/crypto/bcrypt"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

// Session stores user session info.
type Session struct {
	UserID    int64
	ExpiresAt time.Time
}

// AuthService implements the AuthService gRPC service.
type AuthService struct {
	queries  *sqlc.Queries
	sessions map[string]Session
	mu       sync.RWMutex
}

// NewAuthService creates a new AuthService.
func NewAuthService(queries *sqlc.Queries) *AuthService {
	return &AuthService{
		queries:  queries,
		sessions: make(map[string]Session),
	}
}

// Login validates credentials and returns a session token.
func (s *AuthService) Login(
	ctx context.Context,
	req *connect.Request[ntxv1.LoginRequest],
) (*connect.Response[ntxv1.LoginResponse], error) {
	email := req.Msg.Email
	password := req.Msg.Password

	if email == "" || password == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("email and password required"))
	}

	// Check if we should use env-based single user auth
	envEmail := os.Getenv("AUTH_EMAIL")
	envPasswordHash := os.Getenv("AUTH_PASSWORD_HASH")

	var userID int64

	if envEmail != "" && envPasswordHash != "" {
		// Env-based auth (single user mode)
		if email != envEmail {
			return nil, connect.NewError(connect.CodeUnauthenticated, ErrInvalidCredentials)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(envPasswordHash), []byte(password)); err != nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, ErrInvalidCredentials)
		}
		userID = 1 // Fake user ID for env-based auth
	} else {
		// Database-based auth
		user, err := s.queries.GetUserByEmail(ctx, email)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, ErrInvalidCredentials)
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, ErrInvalidCredentials)
		}
		userID = user.ID
	}

	// Generate session token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to generate token"))
	}
	token := hex.EncodeToString(tokenBytes)

	// Store session
	s.mu.Lock()
	s.sessions[token] = Session{
		UserID:    userID,
		ExpiresAt: time.Now().Add(24 * time.Hour * 7), // 7 days
	}
	s.mu.Unlock()

	return connect.NewResponse(&ntxv1.LoginResponse{
		Token:  token,
		UserId: userID,
	}), nil
}

// ValidateToken checks if a token is valid and returns the user ID.
func (s *AuthService) ValidateToken(token string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[token]
	if !ok {
		return 0, false
	}

	if time.Now().After(session.ExpiresAt) {
		return 0, false
	}

	return session.UserID, true
}

// Register creates a new user account.
func (s *AuthService) Register(
	ctx context.Context,
	req *connect.Request[ntxv1.RegisterRequest],
) (*connect.Response[ntxv1.RegisterResponse], error) {
	email := req.Msg.Email
	password := req.Msg.Password

	if email == "" || password == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("email and password required"))
	}

	if len(password) < 6 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("password must be at least 6 characters"))
	}

	// Check if user already exists
	_, err := s.queries.GetUserByEmail(ctx, email)
	if err == nil {
		return nil, connect.NewError(connect.CodeAlreadyExists, errors.New("email already registered"))
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to hash password"))
	}

	// Create user
	user, err := s.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        email,
		PasswordHash: string(hash),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to create user"))
	}

	return connect.NewResponse(&ntxv1.RegisterResponse{
		UserId: user.ID,
	}), nil
}
