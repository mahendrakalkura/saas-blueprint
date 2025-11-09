package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mahendrakalkura/saas-blueprint/internal/auth"
	"github.com/mahendrakalkura/saas-blueprint/internal/redis"
)

// CacheConfig holds configuration for response caching
type CacheConfig struct {
	TTL            time.Duration // Cache TTL
	KeyPrefix      string        // Prefix for cache keys
	IncludeUserID  bool          // Include user ID in cache key
	IncludeQueryString bool       // Include query string in cache key
}

// DefaultCacheConfig returns a sensible default configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		TTL:                5 * time.Minute,
		KeyPrefix:          "api:cache:",
		IncludeUserID:      false,
		IncludeQueryString: true,
	}
}

// cacheResponseWriter wraps http.ResponseWriter to capture response
type cacheResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func newCacheResponseWriter(w http.ResponseWriter) *cacheResponseWriter {
	return &cacheResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
	}
}

func (w *cacheResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *cacheResponseWriter) Write(b []byte) (int, error) {
	// Write to both the buffer and the actual response
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// CacheMiddleware creates a middleware that caches responses in Redis
func Cache(redisClient *redis.Client, config CacheConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only cache GET requests
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			// Generate cache key
			cacheKey := generateCacheKey(r, config)

			// Try to get cached response
			cachedResponse, err := redisClient.Get(r.Context(), cacheKey)
			if err == nil && cachedResponse != "" {
				// Cache hit
				var cached CachedResponse
				if err := json.Unmarshal([]byte(cachedResponse), &cached); err == nil {
					// Set cached headers
					for key, values := range cached.Headers {
						for _, value := range values {
							w.Header().Add(key, value)
						}
					}
					w.Header().Set("X-Cache", "HIT")
					w.Header().Set("X-Cache-Key", cacheKey)

					// Write cached response
					w.WriteHeader(cached.StatusCode)
					w.Write([]byte(cached.Body))
					return
				}
			}

			// Cache miss - capture response
			crw := newCacheResponseWriter(w)
			next.ServeHTTP(crw, r)

			// Only cache successful responses (200-299)
			if crw.statusCode >= 200 && crw.statusCode < 300 {
				// Prepare cached response
				cachedResp := CachedResponse{
					StatusCode: crw.statusCode,
					Headers:    crw.Header(),
					Body:       crw.body.String(),
					CachedAt:   time.Now(),
				}

				// Serialize and cache
				if data, err := json.Marshal(cachedResp); err == nil {
					redisClient.Set(r.Context(), cacheKey, string(data), config.TTL)
				}

				// Add cache miss header
				w.Header().Set("X-Cache", "MISS")
				w.Header().Set("X-Cache-Key", cacheKey)
			}
		})
	}
}

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	StatusCode int         `json:"status_code"`
	Headers    http.Header `json:"headers"`
	Body       string      `json:"body"`
	CachedAt   time.Time   `json:"cached_at"`
}

// generateCacheKey creates a cache key based on request parameters
func generateCacheKey(r *http.Request, config CacheConfig) string {
	var keyParts []string

	// Add prefix
	keyParts = append(keyParts, config.KeyPrefix)

	// Add path
	keyParts = append(keyParts, r.URL.Path)

	// Add user ID if configured
	if config.IncludeUserID {
		if userCtx, ok := auth.GetUserFromContext(r); ok {
			keyParts = append(keyParts, "user", userCtx.UserID)
		}
	}

	// Add query string if configured
	if config.IncludeQueryString && r.URL.RawQuery != "" {
		// Hash the query string to keep keys manageable
		hash := sha256.Sum256([]byte(r.URL.RawQuery))
		queryHash := hex.EncodeToString(hash[:])
		keyParts = append(keyParts, "query", queryHash[:16])
	}

	return strings.Join(keyParts, ":")
}

// CacheHeaders adds HTTP caching headers to responses
func CacheHeaders(maxAge time.Duration, options ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Build Cache-Control header
			var directives []string

			// Add max-age
			if maxAge > 0 {
				directives = append(directives, fmt.Sprintf("max-age=%d", int(maxAge.Seconds())))
			} else {
				directives = append(directives, "no-cache", "no-store", "must-revalidate")
			}

			// Add additional options
			directives = append(directives, options...)

			// Set headers
			w.Header().Set("Cache-Control", strings.Join(directives, ", "))

			if maxAge > 0 {
				// Set Expires header
				expires := time.Now().Add(maxAge).UTC().Format(http.TimeFormat)
				w.Header().Set("Expires", expires)

				// Set Last-Modified header (current time as placeholder)
				lastModified := time.Now().UTC().Format(http.TimeFormat)
				w.Header().Set("Last-Modified", lastModified)
			}

			// Add Vary header for content negotiation
			w.Header().Set("Vary", "Accept-Encoding, Authorization")

			next.ServeHTTP(w, r)
		})
	}
}

// NoCache is a middleware that prevents caching
func NoCache(next http.Handler) http.Handler {
	return CacheHeaders(0, "private")(next)
}

// PublicCache is a middleware for public, cacheable resources
func PublicCache(maxAge time.Duration) func(http.Handler) http.Handler {
	return CacheHeaders(maxAge, "public")
}

// PrivateCache is a middleware for user-specific, cacheable resources
func PrivateCache(maxAge time.Duration) func(http.Handler) http.Handler {
	return CacheHeaders(maxAge, "private")
}

// ETag generates and validates ETags for responses
func ETag(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture response
		crw := newCacheResponseWriter(w)
		next.ServeHTTP(crw, r)

		// Generate ETag from response body
		if crw.statusCode >= 200 && crw.statusCode < 300 {
			hash := sha256.Sum256(crw.body.Bytes())
			etag := `"` + hex.EncodeToString(hash[:])[:16] + `"`

			// Check If-None-Match header
			if match := r.Header.Get("If-None-Match"); match != "" {
				if match == etag {
					// ETag matches - return 304 Not Modified
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}

			// Set ETag header
			w.Header().Set("ETag", etag)
		}
	})
}

// CacheInvalidate provides a way to invalidate cache keys
type CacheInvalidator struct {
	redis *redis.Client
}

// NewCacheInvalidator creates a new cache invalidator
func NewCacheInvalidator(redisClient *redis.Client) *CacheInvalidator {
	return &CacheInvalidator{redis: redisClient}
}

// InvalidateByPattern invalidates all cache keys matching a pattern
func (ci *CacheInvalidator) InvalidateByPattern(pattern string) error {
	// Note: This requires SCAN command which we need to add to redis client
	// For now, we'll invalidate specific keys
	return fmt.Errorf("pattern invalidation not yet implemented")
}

// InvalidateKey invalidates a specific cache key
func (ci *CacheInvalidator) InvalidateKey(key string) error {
	return ci.redis.Del(nil, key)
}

// InvalidatePath invalidates all cache entries for a specific path
func (ci *CacheInvalidator) InvalidatePath(path string, prefix string) error {
	pattern := prefix + path + "*"
	return ci.InvalidateByPattern(pattern)
}

// CacheStats holds cache statistics
type CacheStats struct {
	Hits   int64   `json:"hits"`
	Misses int64   `json:"misses"`
	Size   int64   `json:"size"`
	HitRate float64 `json:"hit_rate"`
}

// GetCacheStats returns cache statistics (placeholder implementation)
func GetCacheStats(redisClient *redis.Client) (*CacheStats, error) {
	// This would need to be implemented with actual Redis commands
	// For now, return empty stats
	return &CacheStats{
		Hits:    0,
		Misses:  0,
		Size:    0,
		HitRate: 0.0,
	}, nil
}
