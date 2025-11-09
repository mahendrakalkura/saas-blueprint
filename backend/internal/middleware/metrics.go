package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mahendrakalkura/saas-blueprint/internal/metrics"
)

// responseWriter wraps http.ResponseWriter to capture status code
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int
}

func newMetricsResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (mrw *metricsResponseWriter) WriteHeader(code int) {
	mrw.statusCode = code
	mrw.ResponseWriter.WriteHeader(code)
}

func (mrw *metricsResponseWriter) Write(b []byte) (int, error) {
	n, err := mrw.ResponseWriter.Write(b)
	mrw.written += n
	return n, err
}

// Metrics middleware records HTTP metrics
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code and size
		mrw := newMetricsResponseWriter(w)

		// Get route pattern if available
		routePattern := r.URL.Path
		if rctx := chi.RouteContext(r.Context()); rctx != nil && rctx.RoutePattern() != "" {
			routePattern = rctx.RoutePattern()
		}

		// Record request size
		if r.ContentLength > 0 {
			metrics.HTTPRequestSize.WithLabelValues(
				r.Method,
				routePattern,
			).Observe(float64(r.ContentLength))
		}

		// Process request
		next.ServeHTTP(mrw, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(mrw.statusCode)

		metrics.HTTPRequestsTotal.WithLabelValues(
			r.Method,
			routePattern,
			statusCode,
		).Inc()

		metrics.HTTPRequestDuration.WithLabelValues(
			r.Method,
			routePattern,
		).Observe(duration)

		metrics.HTTPResponseSize.WithLabelValues(
			r.Method,
			routePattern,
		).Observe(float64(mrw.written))
	})
}
