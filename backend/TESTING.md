# Testing Guide

This document describes the testing infrastructure and best practices for the SaaS Blueprint backend.

## Table of Contents

- [Overview](#overview)
- [Running Tests](#running-tests)
- [Test Structure](#test-structure)
- [Writing Tests](#writing-tests)
- [Test Fixtures](#test-fixtures)
- [Continuous Integration](#continuous-integration)
- [Best Practices](#best-practices)

## Overview

The backend uses Go's built-in testing framework with additional utilities for:

- **Unit Tests**: Test individual functions and methods in isolation
- **Integration Tests**: Test API endpoints with a real database
- **Test Fixtures**: Reusable test data and helpers

## Running Tests

### Quick Start

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -cover ./...
```

### Using Makefile

```bash
# Run all tests
make test

# Run only unit tests
make test-unit

# Run integration tests
make test-integration

# Generate coverage report
make test-coverage

# Run linter
make lint

# Run all quality checks (vet, lint, test)
make check
```

### Running Specific Tests

```bash
# Run tests in a specific package
go test ./internal/auth/...

# Run a specific test
go test -run TestHashPassword ./internal/auth/...

# Run tests matching a pattern
go test -run "TestAuth.*" ./...
```

## Test Structure

```
backend/
├── internal/
│   ├── auth/
│   │   ├── auth.go
│   │   └── auth_test.go          # Unit tests for auth package
│   ├── mfa/
│   │   ├── service.go
│   │   └── service_test.go       # Unit tests for MFA service
│   ├── api/
│   │   └── v1/
│   │       ├── auth.go
│   │       └── auth_test.go      # Integration tests for auth endpoints
│   └── testutil/
│       ├── database.go           # Test database utilities
│       └── fixtures.go           # Test data fixtures
└── TESTING.md                    # This file
```

## Writing Tests

### Unit Tests

Unit tests should test individual functions in isolation without external dependencies.

```go
package auth

import "testing"

func TestHashPassword(t *testing.T) {
    password := "SecurePassword123!"

    hash, err := HashPassword(password)
    if err != nil {
        t.Fatalf("HashPassword() error = %v", err)
    }

    if hash == "" {
        t.Error("HashPassword() returned empty hash")
    }

    // Verify the password
    err = CheckPassword(password, hash)
    if err != nil {
        t.Errorf("CheckPassword() failed: %v", err)
    }
}
```

### Integration Tests

Integration tests test API endpoints with a real database connection.

```go
package v1

import (
    "testing"
    "net/http/httptest"

    "github.com/mahendrakalkura/saas-blueprint/internal/testutil"
)

func TestAuthHandler_Register(t *testing.T) {
    // Setup test database
    testDB := testutil.SetupTestDB(t)
    defer testDB.TeardownTestDB(t)

    // Create handler with test dependencies
    handler := NewAuthHandler(/* ... */)

    // Create test request
    req := httptest.NewRequest("POST", "/auth/register", payload)
    w := httptest.NewRecorder()

    // Call handler
    handler.Register(w, req)

    // Assert response
    if w.Code != http.StatusCreated {
        t.Errorf("Expected status 201, got %d", w.Code)
    }
}
```

### Table-Driven Tests

Use table-driven tests for testing multiple scenarios:

```go
func TestValidateAccessToken(t *testing.T) {
    tests := []struct {
        name    string
        token   string
        secret  string
        wantErr bool
    }{
        {
            name:    "valid token",
            token:   validToken,
            secret:  secret,
            wantErr: false,
        },
        {
            name:    "invalid secret",
            token:   validToken,
            secret:  "wrong-secret",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := ValidateAccessToken(tt.token, tt.secret)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Test Fixtures

The `testutil` package provides utilities for creating test data.

### Database Setup

```go
// Setup test database
testDB := testutil.SetupTestDB(t)
defer testDB.TeardownTestDB(t)

// Clean specific tables between tests
testDB.CleanTables(t, "users", "sessions")
```

### Creating Test Users

```go
userRepo := repository.NewUserRepository(testDB.DB.DB)

// Create a basic test user
user := testutil.CreateTestUser(t, userRepo, "test@example.com")

// Create a user with MFA enabled
userWithMFA := testutil.CreateTestUserWithMFA(t, userRepo, "mfa@example.com", "secret")
```

### Generating Test Tokens

```go
// Generate a test JWT
token := testutil.GenerateTestJWT(t, user.ID, user.Email, "test-secret")

// Create a test session
session := testutil.CreateTestSession(t, sessionRepo, user.ID)
```

## Continuous Integration

Tests run automatically on every push and pull request via GitHub Actions.

### CI Workflow

The CI pipeline includes:

1. **Backend Tests**
   - Run all unit and integration tests
   - Check for race conditions
   - Generate coverage reports
   - Run `go vet`
   - Run `staticcheck`

2. **Frontend Tests**
   - Run linter
   - Type checking
   - Build verification

3. **Integration Tests**
   - Full stack testing with PostgreSQL, Redis, and MinIO
   - Database migrations
   - End-to-end API tests

### Local CI Simulation

```bash
# Run the same checks as CI
make check

# Run tests with race detection (like CI)
go test -race ./...

# Run staticcheck (like CI)
staticcheck ./...
```

## Best Practices

### General Guidelines

1. **Use `t.Helper()`** in utility functions to get better error locations
2. **Use `t.Cleanup()`** for cleanup instead of `defer` when possible
3. **Test error cases** as well as success cases
4. **Use meaningful test names** that describe what is being tested
5. **Keep tests independent** - each test should be able to run alone
6. **Use table-driven tests** for testing multiple scenarios
7. **Don't use `t.Fatal()` in goroutines** - use `t.Error()` instead

### Database Tests

1. **Create a fresh database** for each test or test suite
2. **Clean up** after tests using `defer testDB.TeardownTestDB(t)`
3. **Use transactions** for tests that need to be isolated
4. **Avoid hardcoded IDs** - use generated UUIDs
5. **Test database constraints** and foreign key relationships

### API Tests

1. **Test all HTTP methods** (GET, POST, PUT, DELETE)
2. **Test different status codes** (200, 400, 401, 404, etc.)
3. **Test request validation** (missing fields, invalid formats)
4. **Test authentication** (missing token, invalid token, expired token)
5. **Test authorization** (insufficient permissions)
6. **Verify response structure** (not just status codes)

### Coverage Goals

- **Overall coverage**: Aim for 70%+ coverage
- **Critical paths**: Aim for 90%+ coverage (auth, payments, security)
- **Focus on**:
  - Business logic
  - Error handling
  - Edge cases
- **Don't obsess over**:
  - 100% coverage
  - Simple getters/setters
  - Generated code

### Performance

1. **Use `-short` flag** for quick feedback during development
2. **Run integration tests** separately when needed
3. **Use parallel tests** with `t.Parallel()` when safe
4. **Profile slow tests** and optimize if needed

### CI Optimization

1. **Cache Go modules** in CI
2. **Run linter before tests** to fail fast
3. **Use matrix builds** for different Go versions if needed
4. **Parallelize independent test suites**

## Troubleshooting

### Tests Fail Locally But Pass in CI

- Check environment variables
- Verify database state (did previous test leave data?)
- Check for timing issues (use fixed times in tests)
- Verify dependencies are up to date

### Flaky Tests

- Avoid time-dependent tests (use fixed times)
- Ensure proper cleanup between tests
- Check for race conditions with `-race` flag
- Don't rely on external services

### Slow Tests

- Use `go test -v ./... | grep -E '(PASS|FAIL).*\([0-9]+\.[0-9]+s\)'` to find slow tests
- Consider mocking external dependencies
- Use `t.Parallel()` for independent tests
- Optimize database queries in tests

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Go Test Examples](https://go.dev/blog/examples)
- [Table Driven Tests](https://go.dev/wiki/TableDrivenTests)
- [Testify Package](https://github.com/stretchr/testify) (if we add it)

## Examples

See the following files for complete examples:

- `internal/auth/auth_test.go` - Unit tests
- `internal/mfa/service_test.go` - Service tests
- `internal/api/v1/auth_test.go` - Integration tests
- `internal/testutil/` - Test utilities and fixtures
