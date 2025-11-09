package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mahendrakalkura/saas-blueprint/internal/auth"
	"github.com/mahendrakalkura/saas-blueprint/internal/ratelimit"
	"github.com/rs/zerolog/log"
)

// RateLimitConfig defines different rate limit tiers
type RateLimitConfig struct {
	// For unauthenticated requests (by IP)
	Global ratelimit.Config

	// For authentication endpoints (by IP)
	Auth ratelimit.Config

	// For authenticated users (by user ID)
	Authenticated ratelimit.Config
}

// DefaultRateLimitConfig returns sensible defaults
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		// 100 requests per minute for general unauthenticated traffic
		Global: ratelimit.Config{
			Requests: 100,
			Window:   time.Minute,
		},
		// 10 login attempts per 5 minutes per IP (prevents brute force)
		Auth: ratelimit.Config{
			Requests: 10,
			Window:   5 * time.Minute,
		},
		// 1000 requests per minute for authenticated users
		Authenticated: ratelimit.Config{
			Requests: 1000,
			Window:   time.Minute,
		},
	}
}

// RateLimit creates a rate limiting middleware
func RateLimit(limiter *ratelimit.Limiter, config ratelimit.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to get user ID from context (for authenticated requests)
			key := getClientIP(r)
			userCtx, ok := auth.GetUserFromContext(r)
			if ok {
				key = fmt.Sprintf("user:%s", userCtx.UserID)
			}

			result, err := limiter.Allow(r.Context(), key, config)
			if err != nil {
				log.Error().Err(err).Str("key", key).Msg("Rate limit check failed")
				// Don't block requests if rate limiting fails
				next.ServeHTTP(w, r)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", result.Limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", result.RetryAfter.Unix()))

			if !result.Allowed {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(time.Until(result.RetryAfter).Seconds())))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"Rate limit exceeded. Please try again later."}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (if behind a proxy)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP if there are multiple
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
