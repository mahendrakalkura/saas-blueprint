# API Documentation

Base URL: `http://localhost:8080/api/v1`

## Authentication Endpoints

### Register
Create a new user account.

**Endpoint:** `POST /auth/register`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Response:** `201 Created`
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "abc123...",
  "expires_in": 900,
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "email_verified": false,
    "is_active": true,
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
}
```

### Login
Authenticate a user.

**Endpoint:** `POST /auth/login`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "abc123...",
  "expires_in": 900,
  "user": { ... }
}
```

### Refresh Token
Get a new access token using a refresh token.

**Endpoint:** `POST /auth/refresh`

**Request Body:**
```json
{
  "refresh_token": "abc123..."
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "def456...",
  "expires_in": 900,
  "user": { ... }
}
```

### Logout
Revoke the current session.

**Endpoint:** `POST /auth/logout`

**Request Body:**
```json
{
  "refresh_token": "abc123..."
}
```

**Response:** `200 OK`
```json
{
  "message": "Logged out successfully"
}
```

### Get Current User
Get the authenticated user's information.

**Endpoint:** `GET /auth/me`

**Headers:**
```
Authorization: Bearer {access_token}
```

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "avatar_url": null,
  "email_verified": true,
  "is_active": true,
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

### Verify Email
Verify email address with token.

**Endpoint:** `POST /auth/verify-email`

**Request Body:**
```json
{
  "token": "verification_token_here"
}
```

**Response:** `200 OK`
```json
{
  "message": "Email verified successfully"
}
```

### Request Password Reset
Send a password reset email.

**Endpoint:** `POST /auth/request-password-reset`

**Request Body:**
```json
{
  "email": "user@example.com"
}
```

**Response:** `200 OK`
```json
{
  "message": "If the email exists, a password reset link has been sent"
}
```

### Reset Password
Reset password with token.

**Endpoint:** `POST /auth/reset-password`

**Request Body:**
```json
{
  "token": "reset_token_here",
  "password": "new_password123"
}
```

**Response:** `200 OK`
```json
{
  "message": "Password reset successfully"
}
```

## Health Check Endpoints

### Main Health Check
**Endpoint:** `GET /api/v1/health`

**Response:** `200 OK`
```json
{
  "status": "ok",
  "timestamp": "2025-01-01T00:00:00Z",
  "services": {
    "database": "healthy"
  }
}
```

### Liveness Probe
**Endpoint:** `GET /api/v1/health/live`

### Readiness Probe
**Endpoint:** `GET /api/v1/health/ready`

## Monitoring Endpoints

All monitoring endpoints require authentication.

### Queue Statistics
Get statistics about background job queues.

**Endpoint:** `GET /api/v1/monitoring/queues`

**Headers:**
```
Authorization: Bearer {access_token}
```

**Response:** `200 OK`
```json
{
  "queues": {
    "critical": {
      "active": 2,
      "pending": 5,
      "scheduled": 10,
      "retry": 1,
      "archived": 0,
      "completed": 100,
      "aggregating": 0,
      "size": 18,
      "latency": "2.5s",
      "memory_usage": 1024
    },
    "default": {
      "active": 1,
      "pending": 3,
      "scheduled": 5,
      "retry": 0,
      "archived": 0,
      "completed": 50,
      "aggregating": 0,
      "size": 9,
      "latency": "1.2s",
      "memory_usage": 512
    },
    "low": {
      "active": 0,
      "pending": 1,
      "scheduled": 2,
      "retry": 0,
      "archived": 0,
      "completed": 25,
      "aggregating": 0,
      "size": 3,
      "latency": "500ms",
      "memory_usage": 256
    }
  }
}
```

### Server Information
Get information about worker servers.

**Endpoint:** `GET /api/v1/monitoring/servers`

**Headers:**
```
Authorization: Bearer {access_token}
```

**Response:** `200 OK`
```json
{
  "servers": [
    {
      "host": "worker-1",
      "pid": 12345,
      "server_id": "server-123",
      "concurrency": 10,
      "queues": {
        "critical": 6,
        "default": 3,
        "low": 1
      },
      "strict_priority": false,
      "status": "active",
      "started_at": "2025-01-01T00:00:00Z",
      "active_workers": 5
    }
  ]
}
```

### Scheduled Tasks
Get information about scheduled periodic tasks.

**Endpoint:** `GET /api/v1/monitoring/scheduled`

**Headers:**
```
Authorization: Bearer {access_token}
```

**Response:** `200 OK`
```json
{
  "scheduled_tasks": [
    {
      "id": "cleanup-sessions",
      "spec": "@every 1h",
      "task": "cleanup:sessions",
      "opts": [],
      "next_run": "2025-01-01T01:00:00Z",
      "prev_run": "2025-01-01T00:00:00Z"
    }
  ]
}
```

## Error Responses

All errors follow this format:

```json
{
  "error": "Error message here"
}
```

For validation errors:

```json
{
  "error": "Validation failed",
  "fields": {
    "email": "Invalid email format",
    "password": "Password must be at least 8 characters"
  }
}
```

## Status Codes

- `200 OK` - Success
- `201 Created` - Resource created
- `400 Bad Request` - Invalid request
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource already exists
- `422 Unprocessable Entity` - Validation failed
- `500 Internal Server Error` - Server error
