# MVP-AUTH-003: Authentication Endpoints

## Status: ✅ Complete

## Overview
Implement REST API endpoints for user registration, login, token refresh, logout, and getting current user info.

## Implementation

### File: `internal/auth/handler.go`

**Handler Structure:**
```go
type Handler struct {
    service *Service
    logger  *logrus.Logger
}
```

## Endpoints

### 1. POST /api/v1/auth/register

**Purpose**: Create new user account

**Request Body:**
```json
{
  "email": "user@example.com",
  "name": "John Doe",
  "password": "securePassword123"
}
```

**Validation:**
- Email must be valid format (Gin binding validation)
- Name required, 2-100 characters
- Password required, minimum 8 characters

**Success Response (201 Created):**
```json
{
  "message": "User registered successfully",
  "user": {
    "id": "users/user_abc123",
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2026-01-28T10:00:00Z",
    "updated_at": "2026-01-28T10:00:00Z",
    "is_active": true
  }
}
```

**Error Responses:**
- **400 Bad Request**: Invalid JSON or validation failed
- **409 Conflict**: Email already exists

### 2. POST /api/v1/auth/login

**Purpose**: Authenticate user and return tokens

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123"
}
```

**Success Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "xK9mP2nQ4rS6tU8vW0yA1bC3dE5fG7hI9jK",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "users/user_abc123",
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2026-01-28T10:00:00Z",
    "updated_at": "2026-01-28T10:00:00Z",
    "is_active": true
  }
}
```

**Error Responses:**
- **400 Bad Request**: Invalid JSON
- **401 Unauthorized**: Invalid credentials or account disabled

### 3. POST /api/v1/auth/refresh

**Purpose**: Obtain new access token using refresh token

**Request Body:**
```json
{
  "refresh_token": "xK9mP2nQ4rS6tU8vW0yA1bC3dE5fG7hI9jK"
}
```

**Success Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "xK9mP2nQ4rS6tU8vW0yA1bC3dE5fG7hI9jK",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "users/user_abc123",
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2026-01-28T10:00:00Z",
    "updated_at": "2026-01-28T10:00:00Z",
    "is_active": true
  }
}
```

**Error Responses:**
- **400 Bad Request**: Missing refresh_token
- **401 Unauthorized**: Invalid or expired refresh token

### 4. POST /api/v1/auth/logout

**Purpose**: Revoke refresh token

**Request Body:**
```json
{
  "refresh_token": "xK9mP2nQ4rS6tU8vW0yA1bC3dE5fG7hI9jK"
}
```

**Success Response (200 OK):**
```json
{
  "message": "Logged out successfully"
}
```

**Note**: If no refresh_token provided but user is authenticated (from context), logout succeeds but token revocation is skipped.

### 5. GET /api/v1/auth/me

**Purpose**: Get current authenticated user info

**Authentication**: Requires valid access token in Authorization header

**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Success Response (200 OK):**
```json
{
  "id": "users/user_abc123",
  "email": "user@example.com",
  "name": "John Doe",
  "created_at": "2026-01-28T10:00:00Z",
  "updated_at": "2026-01-28T10:00:00Z",
  "is_active": true
}
```

**Error Responses:**
- **401 Unauthorized**: Missing or invalid access token
- **404 Not Found**: User not found in database

## Route Registration

**File**: `internal/app/routes_api.go`

```go
// Authentication endpoints (MVP-AUTH-003) - Public routes
if a.authService != nil {
    authHandler := auth.NewHandler(a.authService, a.logger)
    authHandler.RegisterRoutes(apiV1)
    a.logger.Info("Authentication endpoints registered")
}
```

**Routes Registered:**
```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
POST   /api/v1/auth/logout
GET    /api/v1/auth/me
```

## Error Handling

**Generic Error Format:**
```json
{
  "error": "Error message",
  "details": "Optional detailed error information"
}
```

**Logging:**
- Registration attempts logged with email
- Login failures logged with email (not password)
- Token refresh failures logged
- All errors logged with logrus

**Security Considerations:**
- Passwords never logged
- Generic error messages prevent email enumeration
- Failed login attempts should be monitored for brute force

## Testing with cURL

**Register User:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "name": "Test User",
    "password": "password123"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

**Get Current User:**
```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Refresh Token:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

**Logout:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

## Integration with Frontend

**CodeValdFortex AuthService** (Dart):
```dart
final response = await _dio.post(
  '/api/v1/auth/login',
  data: credentials.toJson(),
);

return AuthResponse.fromJson(response.data);
```

**Expected Frontend Behavior:**
1. Store access_token in memory (state management)
2. Store refresh_token in flutter_secure_storage
3. Add access_token to all API requests via interceptor
4. On 401 response, attempt token refresh
5. On logout, clear both tokens and navigate to login

## Dependencies
- `github.com/gin-gonic/gin` - HTTP router and handlers
- `github.com/sirupsen/logrus` - Logging
- `internal/auth.Service` - Business logic

## Future Enhancements
- [ ] Rate limiting on auth endpoints (5 attempts per minute)
- [ ] CAPTCHA on repeated failed logins
- [ ] Email verification required before login
- [ ] Password reset endpoint
- [ ] Change password endpoint (authenticated)
- [ ] List active sessions endpoint
- [ ] Revoke all sessions endpoint
