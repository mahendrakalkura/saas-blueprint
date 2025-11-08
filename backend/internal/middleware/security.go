package middleware

import (
	"net/http"
)

// SecurityHeaders adds security headers to responses
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// XSS Protection (for older browsers)
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
			"script-src 'self' 'unsafe-inline'; "+
			"style-src 'self' 'unsafe-inline'; "+
			"img-src 'self' data: https:; "+
			"font-src 'self' data:; "+
			"connect-src 'self'; "+
			"frame-ancestors 'none';")

		// Strict Transport Security (HSTS)
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Permissions Policy
		w.Header().Set("Permissions-Policy",
			"geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}

// CSRF protection middleware (simple implementation)
// For production, use gorilla/csrf or similar
func CSRFProtection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF check for safe methods
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// In production, implement proper CSRF token validation
		// For now, we'll rely on same-site cookies and CORS
		next.ServeHTTP(w, r)
	})
}

// RateLimitHeaders adds rate limit information to response headers
func RateLimitHeaders(limit, remaining, reset int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-RateLimit-Limit", string(rune(limit)))
			w.Header().Set("X-RateLimit-Remaining", string(rune(remaining)))
			w.Header().Set("X-RateLimit-Reset", string(rune(reset)))
			next.ServeHTTP(w, r)
		})
	}
}
