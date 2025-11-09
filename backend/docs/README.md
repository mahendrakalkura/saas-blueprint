# API Documentation

This directory contains the OpenAPI/Swagger documentation for the SaaS Blueprint API.

## Accessing the Documentation

### Swagger UI

Visit the interactive API documentation at:
```
http://localhost:8080/api/v1/docs
```

The Swagger UI provides:
- Interactive API testing
- Request/response examples
- Schema definitions
- Authentication testing

### OpenAPI Specification

The raw OpenAPI 3.0 specification is available at:
```
http://localhost:8080/api/v1/docs/swagger.json
```

## Authentication

Most endpoints require JWT authentication. To use authenticated endpoints in Swagger UI:

1. Click the "Authorize" button at the top of the page
2. Enter your JWT token in the format: `Bearer your_token_here`
3. Click "Authorize"
4. All subsequent requests will include the Authorization header

## Rate Limiting

The API implements rate limiting to prevent abuse:

- **Authentication endpoints**: 10 requests per 5 minutes (per IP)
- **Authenticated requests**: 1000 requests per minute (per user)
- **General endpoints**: 100 requests per minute (per IP)

Rate limit headers are included in responses:
- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Requests remaining in current window
- `X-RateLimit-Reset`: Unix timestamp when limit resets
- `Retry-After`: Seconds to wait before retrying (when rate limited)

## API Features

### Authentication
- User registration and login
- Email verification
- Password reset
- JWT-based authentication
- Refresh tokens
- Multi-Factor Authentication (TOTP)

### Organizations
- Multi-tenant organization management
- Role-based access control (Owner, Admin, Member)
- Member invitations
- Organization settings

### Files
- File upload with validation
- Avatar management
- File listing and deletion

### Billing
- Stripe integration
- Subscription management
- Customer portal
- Webhook handling

### Notifications
- Real-time notifications via WebSocket
- Notification history
- Mark as read/unread

### WebSocket
- Real-time bidirectional communication
- Connection statistics
- User-specific messaging

## Development

### Generating Documentation

The OpenAPI specification is maintained manually in `swagger.json`. To make changes:

1. Edit `swagger.json` directly
2. Restart the server
3. Refresh the Swagger UI to see changes

### Adding New Endpoints

To document a new endpoint:

1. Add Swagger annotations to the handler function in Go code
2. Update `swagger.json` with the endpoint definition
3. Include request/response schemas in the components section

Example annotation:
```go
// @Summary Get current user
// @Description Get the currently authenticated user's information
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.User
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
    // handler code
}
```

## Response Codes

Common HTTP status codes used in the API:

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request payload
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource already exists
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

## Support

For API support, please contact:
- Email: support@example.com
- Documentation: https://example.com/docs
