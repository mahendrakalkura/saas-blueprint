FROM golang:1.23-alpine

WORKDIR /app

# Install air for hot reloading and sqlc
RUN go install github.com/air-verse/air@latest && \
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Expose port
EXPOSE 8080

# Run with air for hot reloading
CMD ["air", "-c", ".air.toml"]
