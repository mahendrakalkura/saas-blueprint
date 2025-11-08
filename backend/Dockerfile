# Development stage
FROM golang:1.23-alpine AS development

WORKDIR /app

# Install development tools
RUN apk add --no-cache git ca-certificates && \
    go install github.com/air-verse/air@latest && \
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest && \
    go install github.com/pressly/goose/v3/cmd/goose@latest

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Expose port
EXPOSE 8080

# Run with air for hot reloading
CMD ["air", "-c", ".air.toml"]

# Builder stage for production
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o main .

# Production stage
FROM alpine:3.19 AS production

WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser

# Copy binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/db/migrations ./db/migrations

# Use non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Run the application
CMD ["./main"]
