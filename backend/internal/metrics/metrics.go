package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	HTTPRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "endpoint"},
	)

	HTTPResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "endpoint"},
	)

	// Authentication Metrics
	AuthRegistrationsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_registrations_total",
			Help: "Total number of user registrations",
		},
	)

	AuthLoginsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_logins_total",
			Help: "Total number of login attempts",
		},
		[]string{"status"}, // success, failed, mfa_required
	)

	AuthMFAVerificationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_mfa_verifications_total",
			Help: "Total number of MFA verification attempts",
		},
		[]string{"status", "method"}, // status: success/failed, method: totp/backup
	)

	AuthPasswordResetsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_password_resets_total",
			Help: "Total number of password reset requests",
		},
	)

	// Database Metrics
	DatabaseQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation"}, // select, insert, update, delete
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	DatabaseConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "database_connections_active",
			Help: "Number of active database connections",
		},
	)

	DatabaseConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "database_connections_idle",
			Help: "Number of idle database connections",
		},
	)

	// WebSocket Metrics
	WebSocketConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_connections_active",
			Help: "Number of active WebSocket connections",
		},
	)

	WebSocketMessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "websocket_messages_total",
			Help: "Total number of WebSocket messages",
		},
		[]string{"direction"}, // sent, received
	)

	// Background Job Metrics
	BackgroundJobsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "background_jobs_total",
			Help: "Total number of background jobs",
		},
		[]string{"job_type", "status"}, // status: success, failed
	)

	BackgroundJobDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "background_job_duration_seconds",
			Help:    "Background job duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"job_type"},
	)

	BackgroundJobQueueSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "background_job_queue_size",
			Help: "Number of jobs in the queue",
		},
		[]string{"queue"},
	)

	// File Storage Metrics
	FileUploadsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "file_uploads_total",
			Help: "Total number of file uploads",
		},
		[]string{"status"}, // success, failed
	)

	FileUploadSize = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "file_upload_size_bytes",
			Help:    "File upload size in bytes",
			Buckets: prometheus.ExponentialBuckets(1024, 10, 8), // 1KB to ~10GB
		},
	)

	// Payment Metrics
	StripeWebhooksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "stripe_webhooks_total",
			Help: "Total number of Stripe webhook events",
		},
		[]string{"event_type", "status"},
	)

	SubscriptionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "subscriptions_active",
			Help: "Number of active subscriptions",
		},
	)

	// Rate Limiting Metrics
	RateLimitHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"tier"}, // global, auth, authenticated
	)

	RateLimitExceededTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_exceeded_total",
			Help: "Total number of rate limit exceeded events",
		},
		[]string{"tier"},
	)

	// Cache Metrics (Redis)
	CacheOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_operations_total",
			Help: "Total number of cache operations",
		},
		[]string{"operation", "status"}, // operation: get/set/del, status: hit/miss/error
	)

	// Application Metrics
	ApplicationStartTime = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "application_start_time_seconds",
			Help: "Application start time in unix timestamp",
		},
	)

	ApplicationInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "application_info",
			Help: "Application version info",
		},
		[]string{"version", "environment"},
	)
)

// Initialize sets up initial metric values
func Initialize(version, environment string) {
	ApplicationInfo.WithLabelValues(version, environment).Set(1)
}
