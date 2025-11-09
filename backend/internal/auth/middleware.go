package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/mahendrakalkura/saas-blueprint/internal/config"
)

type contextKey string

const UserContextKey = contextKey("user")

type UserContext struct {
	UserID string
	Email  string
}

// AuthMiddleware validates JWT tokens and adds user info to context
func AuthMiddleware(cfg *config.JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Check for Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Validate token
			claims, err := ValidateAccessToken(tokenString, cfg.Secret)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Add user info to context
			userCtx := UserContext{
				UserID: claims.UserID,
				Email:  claims.Email,
			}

			ctx := context.WithValue(r.Context(), UserContextKey, userCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext retrieves user info from request context
func GetUserFromContext(r *http.Request) (*UserContext, bool) {
	user, ok := r.Context().Value(UserContextKey).(UserContext)
	return &user, ok
}

// OptionalAuthMiddleware adds user info to context if token is present, but doesn't require it
func OptionalAuthMiddleware(cfg *config.JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					claims, err := ValidateAccessToken(parts[1], cfg.Secret)
					if err == nil {
						userCtx := UserContext{
							UserID: claims.UserID,
							Email:  claims.Email,
						}
						ctx := context.WithValue(r.Context(), UserContextKey, userCtx)
						r = r.WithContext(ctx)
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
