# Performance Optimization Guide

This document describes the performance optimizations implemented in the SaaS Blueprint and best practices for maintaining optimal performance.

## Table of Contents

- [Overview](#overview)
- [Database Optimizations](#database-optimizations)
- [Caching Strategy](#caching-strategy)
- [HTTP Caching](#http-caching)
- [Frontend Optimizations](#frontend-optimizations)
- [Monitoring and Profiling](#monitoring-and-profiling)
- [Performance Best Practices](#performance-best-practices)

## Overview

The application implements multiple layers of optimization:

1. **Database Layer**: Indexes, query optimization, connection pooling
2. **Application Layer**: Redis caching, efficient algorithms
3. **HTTP Layer**: Response caching, compression, CDN-ready
4. **Frontend Layer**: Code splitting, lazy loading, PWA caching

## Database Optimizations

### Indexes

The application includes comprehensive indexes for optimal query performance:

#### Users Table
```sql
-- Email lookup (case-insensitive)
idx_users_email_lower ON users (LOWER(email))

-- Organization membership
idx_users_organization_id ON users (organization_id)

-- Active users
idx_users_email_verified ON users (email_verified) WHERE email_verified = true

-- Temporal queries
idx_users_created_at_desc ON users (created_at DESC)
```

#### Sessions Table
```sql
-- User session lookup
idx_sessions_user_id ON sessions (user_id)

-- Token validation
idx_sessions_refresh_token ON sessions (refresh_token)

-- Active sessions (partial index)
idx_sessions_expires_at ON sessions (expires_at) WHERE expires_at > NOW()

-- Composite for session management
idx_sessions_user_expires ON sessions (user_id, expires_at DESC)
```

#### Organizations & Members
```sql
-- Organization owner
idx_organizations_owner_id ON organizations (owner_id)

-- Membership queries (composite)
idx_org_members_composite ON organization_members (organization_id, user_id)

-- Role-based queries
idx_org_members_role ON organization_members (organization_id, role)
```

#### Full-Text Search
```sql
-- GIN indexes for fast full-text search
idx_users_search_vector ON users USING GIN (search_vector)
idx_organizations_search_vector ON organizations USING GIN (search_vector)
idx_notifications_search_vector ON notifications USING GIN (search_vector)
```

### Query Optimization Tips

1. **Use Prepared Statements**: All queries use prepared statements via sqlc
2. **Avoid N+1 Queries**: Use joins or batch loading
3. **Limit Result Sets**: Always use LIMIT for list queries
4. **Use Partial Indexes**: For queries with common WHERE clauses
5. **Index Foreign Keys**: All foreign keys are indexed

### Connection Pooling

PostgreSQL connection pool configuration:

```go
// Recommended settings for production
db.SetMaxOpenConns(25)           // Maximum open connections
db.SetMaxIdleConns(5)            // Maximum idle connections
db.SetConnMaxLifetime(5 * time.Minute)  // Connection max lifetime
db.SetConnMaxIdleTime(2 * time.Minute)  // Idle connection timeout
```

### Monitoring Database Performance

View database statistics:

```sql
-- Index usage statistics
SELECT * FROM performance_statistics;

-- Table statistics
SELECT * FROM table_statistics;

-- Unused indexes
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan
FROM pg_stat_user_indexes
WHERE idx_scan = 0
ORDER BY pg_relation_size(indexrelid) DESC;

-- Slow queries (requires pg_stat_statements extension)
SELECT
    query,
    mean_exec_time,
    calls,
    total_exec_time
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;
```

## Caching Strategy

### Redis Caching Layers

#### 1. Response Caching

Cache entire API responses for GET requests:

```go
import "github.com/mahendrakalkura/saas-blueprint/internal/middleware"

// Cache configuration
cacheConfig := middleware.CacheConfig{
    TTL:                5 * time.Minute,
    KeyPrefix:          "api:cache:",
    IncludeUserID:      true,  // User-specific cache
    IncludeQueryString: true,  // Include query params in key
}

// Apply to routes
r.With(middleware.Cache(redisClient, cacheConfig)).Get("/users", handler)
```

**When to use:**
- Data that changes infrequently
- Expensive database queries
- Aggregated data or statistics
- Search results

**When NOT to use:**
- Real-time data
- User-specific sensitive data
- POST/PUT/DELETE requests

#### 2. Rate Limiting

Uses Redis for distributed rate limiting:

```go
// Rate limit configuration
rateLimitConfig := ratelimit.Config{
    Requests: 100,               // Max requests
    Window:   1 * time.Minute,   // Time window
}

r.With(middleware.RateLimit(limiter, rateLimitConfig)).Get("/api", handler)
```

#### 3. Session Storage

Sessions stored in Redis for fast access:

```go
// Session key format: session:{session_id}
// TTL matches session expiration
```

### Cache Invalidation

Invalidate cache when data changes:

```go
invalidator := middleware.NewCacheInvalidator(redisClient)

// Invalidate specific key
invalidator.InvalidateKey("api:cache:/users/123")

// Invalidate by path
invalidator.InvalidatePath("/users", "api:cache:")
```

### Cache Key Strategy

Cache keys follow a hierarchical pattern:

```
api:cache:{path}:user:{user_id}:query:{query_hash}
```

Examples:
- `api:cache:/users:query:abc123` - Public user list
- `api:cache:/notifications:user:456:query:def789` - User notifications

## HTTP Caching

### Cache-Control Headers

The application uses intelligent Cache-Control headers:

#### Public Resources (Static Assets)

```go
// Long cache for immutable assets
middleware.PublicCache(365 * 24 * time.Hour)
```

Applied to:
- JavaScript bundles with hash (e.g., `main.abc123.js`)
- CSS files with hash
- Images with hash
- Fonts

#### Private Resources (User-Specific)

```go
// Short cache for user data
middleware.PrivateCache(5 * time.Minute)
```

Applied to:
- User profile data
- User-specific lists
- Dashboard data

#### No Cache (Real-time Data)

```go
middleware.NoCache
```

Applied to:
- Authentication endpoints
- Real-time notifications
- WebSocket connections

### ETag Support

ETags for conditional requests:

```go
r.With(middleware.ETag).Get("/api/resource", handler)
```

Client behavior:
1. First request: Full response with `ETag: "abc123"`
2. Subsequent requests: `If-None-Match: "abc123"`
3. If unchanged: `304 Not Modified` (no body)
4. If changed: `200 OK` with new content and ETag

### Compression

Nginx handles gzip compression:

```nginx
gzip on;
gzip_vary on;
gzip_min_length 1024;
gzip_types text/plain text/css text/xml text/javascript
           application/x-javascript application/xml+rss
           application/json application/javascript;
```

## Frontend Optimizations

### Code Splitting

Vite automatically splits code into chunks:

```javascript
// vite.config.js
export default {
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          'ui-vendor': ['lucide-react'],
          'query-vendor': ['@tanstack/react-query'],
        },
      },
    },
  },
}
```

### Lazy Loading

Load routes on demand:

```javascript
const Dashboard = lazy(() => import('./pages/Dashboard'))
const Settings = lazy(() => import('./pages/Settings'))
```

### Image Optimization

1. **Use appropriate formats**: WebP for photos, SVG for icons
2. **Lazy load images**: Below the fold
3. **Responsive images**: Multiple sizes for different devices
4. **CDN delivery**: Use CDN for static assets

### Bundle Size Optimization

Production build optimization:

```javascript
// Remove console.log in production
terserOptions: {
  compress: {
    drop_console: true,
    drop_debugger: true,
  },
}
```

### PWA Caching

Service worker caches assets for offline use:

- **Precache**: Critical assets (HTML, manifest)
- **Runtime cache**: API responses, images
- **Network first**: API requests
- **Cache first**: Static assets

## Monitoring and Profiling

### Prometheus Metrics

The application exports 40+ metrics for monitoring:

#### HTTP Metrics
```
http_requests_total{method, endpoint, status}
http_request_duration_seconds{method, endpoint}
http_request_size_bytes{method, endpoint}
http_response_size_bytes{method, endpoint}
```

#### Database Metrics
```
database_queries_total{operation}
database_query_duration_seconds{operation}
database_connections_active
database_connections_idle
```

#### Cache Metrics
```
cache_operations_total{operation, status}
rate_limit_hits_total{tier}
rate_limit_exceeded_total{tier}
```

### Profiling Go Applications

Enable pprof for profiling:

```go
import _ "net/http/pprof"

// In development only
if os.Getenv("ENABLE_PPROF") == "true" {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
}
```

Access profiling endpoints:
- CPU: `http://localhost:6060/debug/pprof/profile?seconds=30`
- Memory: `http://localhost:6060/debug/pprof/heap`
- Goroutines: `http://localhost:6060/debug/pprof/goroutine`

### Database Query Analysis

Enable query logging in development:

```sql
-- Show slow queries (> 100ms)
ALTER DATABASE saas_development SET log_min_duration_statement = 100;

-- View query plans
EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'test@example.com';
```

### Load Testing

Use k6 for load testing:

```javascript
import http from 'k6/http';
import { check } from 'k6';

export const options = {
  stages: [
    { duration: '1m', target: 50 },  // Ramp up
    { duration: '3m', target: 50 },  // Stay at 50 users
    { duration: '1m', target: 0 },   // Ramp down
  ],
};

export default function () {
  const res = http.get('https://your-app.com/api/health');
  check(res, { 'status is 200': (r) => r.status === 200 });
}
```

## Performance Best Practices

### Backend

1. **Use Indexes**: Index all foreign keys and frequently queried columns
2. **Batch Operations**: Group multiple operations when possible
3. **Async Processing**: Use background jobs for expensive operations
4. **Connection Pooling**: Configure appropriate pool sizes
5. **Cache Frequently**: Cache data that doesn't change often
6. **Rate Limiting**: Protect against abuse and DoS
7. **Pagination**: Always paginate large result sets
8. **N+1 Queries**: Use joins or eager loading

### Frontend

1. **Code Splitting**: Load only what's needed
2. **Lazy Loading**: Defer non-critical resources
3. **Image Optimization**: Compress and resize images
4. **Bundle Size**: Monitor and minimize bundle size
5. **PWA Caching**: Cache assets for offline use
6. **Debouncing**: Debounce search and input handlers
7. **Memoization**: Cache expensive computations
8. **Virtual Lists**: For long lists

### Database

1. **Normalize Data**: Avoid data duplication
2. **Use Appropriate Types**: Choose optimal column types
3. **Vacuum Regularly**: Clean up dead tuples
4. **Update Statistics**: Keep query planner statistics current
5. **Monitor Indexes**: Remove unused indexes
6. **Partition Tables**: For very large tables
7. **Use Constraints**: Let database enforce rules
8. **Prepared Statements**: Always use prepared statements

### Infrastructure

1. **CDN**: Use CDN for static assets
2. **HTTP/2**: Enable HTTP/2 for multiplexing
3. **Compression**: Enable gzip/brotli compression
4. **Keep-Alive**: Use persistent connections
5. **SSL Session Resumption**: Enable for faster HTTPS
6. **Resource Limits**: Set appropriate container limits
7. **Horizontal Scaling**: Scale out when needed
8. **Load Balancing**: Distribute traffic across instances

## Performance Targets

Aim for these performance targets:

| Metric | Target | Good | Needs Improvement |
|--------|--------|------|-------------------|
| Page Load (First Contentful Paint) | < 1s | < 2s | > 2s |
| API Response (p95) | < 100ms | < 200ms | > 200ms |
| Database Query (p95) | < 50ms | < 100ms | > 100ms |
| Time to Interactive | < 3s | < 5s | > 5s |
| Cache Hit Rate | > 80% | > 60% | < 60% |
| Error Rate | < 0.1% | < 1% | > 1% |

## Troubleshooting Performance Issues

### Slow API Responses

1. Check Prometheus metrics for slow endpoints
2. Review database query execution plans
3. Check cache hit rates
4. Look for N+1 query patterns
5. Check connection pool saturation

### High Database Load

1. Review slow query log
2. Check for missing indexes
3. Analyze query execution plans
4. Check for lock contention
5. Consider read replicas

### Memory Leaks

1. Check Prometheus memory metrics
2. Profile with pprof
3. Look for goroutine leaks
4. Check connection leaks
5. Review caching behavior

### Frontend Performance

1. Run Lighthouse audit
2. Check bundle sizes
3. Analyze network waterfall
4. Check for render blocking resources
5. Profile with Chrome DevTools

## Resources

- [PostgreSQL Performance Tips](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [Go Performance](https://github.com/golang/go/wiki/Performance)
- [Web Performance](https://web.dev/performance/)
- [Redis Best Practices](https://redis.io/topics/best-practices)
