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
