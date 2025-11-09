.PHONY: test test-unit test-integration test-coverage lint vet

# Run all tests
test:
	go test -v -race ./...

# Run unit tests only
test-unit:
	go test -v -race -short ./...

# Run integration tests only
test-integration:
	go test -v -race -tags=integration ./...

# Run tests with coverage
test-coverage:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	golangci-lint run ./...

# Run go vet
vet:
	go vet ./...

# Run staticcheck
staticcheck:
	staticcheck ./...

# Run all quality checks
check: vet lint test

# Clean test cache and coverage files
clean:
	go clean -testcache
	rm -f coverage.out coverage.html

# Install development tools
tools:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run tests in watch mode (requires entr)
watch:
	find . -name "*.go" | entr -c go test -v ./...
