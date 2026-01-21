// Package auth provides authentication middleware.
package auth

import (
	"context"
	"net/http"
	"strings"

	"connectrpc.com/connect"

	"github.com/voidarchive/ntx/internal/portfolio"
)

// AuthInterceptor validates auth tokens for protected routes.
type AuthInterceptor struct {
	authService *AuthService
}

// NewAuthInterceptor creates a new auth interceptor.
func NewAuthInterceptor(authService *AuthService) *AuthInterceptor {
	return &AuthInterceptor{authService: authService}
}

// WrapUnary wraps unary handlers with auth validation.
func (i *AuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		// Skip auth for login and register endpoints
		if strings.Contains(req.Spec().Procedure, "AuthService/Login") ||
			strings.Contains(req.Spec().Procedure, "AuthService/Register") {
			return next(ctx, req)
		}

		// Skip auth for public endpoints (CompanyService, PriceService)
		if strings.Contains(req.Spec().Procedure, "CompanyService") ||
			strings.Contains(req.Spec().Procedure, "PriceService") {
			return next(ctx, req)
		}

		// Extract token from Authorization header
		authHeader := req.Header().Get("Authorization")
		if authHeader == "" {
			return nil, connect.NewError(connect.CodeUnauthenticated, nil)
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			return nil, connect.NewError(connect.CodeUnauthenticated, nil)
		}

		userID, ok := i.authService.ValidateToken(token)
		if !ok {
			return nil, connect.NewError(connect.CodeUnauthenticated, nil)
		}

		// Add user ID to context
		ctx = context.WithValue(ctx, portfolio.UserIDKey, userID)
		return next(ctx, req)
	}
}

// WrapStreamingClient is required by the interface but not used.
func (i *AuthInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// WrapStreamingHandler wraps streaming handlers with auth validation.
func (i *AuthInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		// For streaming, we would need to implement similar logic
		// For now, just pass through (no streaming handlers in this service)
		return next(ctx, conn)
	}
}

// MiddlewareFunc creates an http.Handler middleware for non-Connect routes.
func MiddlewareFunc(authService *AuthService, requireAuth func(r *http.Request) bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if requireAuth != nil && !requireAuth(r) {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userID, ok := authService.ValidateToken(token)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), portfolio.UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
