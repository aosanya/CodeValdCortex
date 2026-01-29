# Authentication & User Management Domain

**Priority**: P0 (CRITICAL - FRONTEND PREREQUISITE)  
**Status**: 4/5 tasks complete  
**Dependencies**: None

## Overview

This domain provides secure user authentication and session management for CodeValdCortex. The authentication system uses JWT tokens with a dual-token approach (access + refresh tokens) for security and user experience. All authentication endpoints are exposed via REST API at `/api/v1/auth/*` to support both the legacy Templ UI and the new Flutter frontend (CodeValdFortex).

## Architecture

### Token Strategy

**Dual Token Approach**:
- **Access Token**: Short-lived (15 minutes), used for API authentication
- **Refresh Token**: Long-lived (7 days), used to obtain new access tokens

**Security Features**:
- Passwords hashed with bcrypt (cost factor 12)
- Refresh tokens hashed with SHA-256 before storage
- Token revocation support via database
- JWT signature validation on every request

### Data Model

**User Model**:
```go
type User struct {
    Key          string    `json:"_key,omitempty"`
    ID           string    `json:"_id,omitempty"`
    Rev          string    `json:"_rev,omitempty"`
    Email        string    `json:"email" binding:"required,email"`
    Name         string    `json:"name" binding:"required"`
    PasswordHash string    `json:"password_hash,omitempty"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    IsActive     bool      `json:"is_active"`
}
```

**RefreshToken Model**:
```go
type RefreshToken struct {
    Key       string     `json:"_key,omitempty"`
    ID        string     `json:"_id,omitempty"`
    Rev       string     `json:"_rev,omitempty"`
    UserID    string     `json:"user_id" binding:"required"`
    Token     string     `json:"token" binding:"required"` // SHA-256 hash
    ExpiresAt time.Time  `json:"expires_at"`
    CreatedAt time.Time  `json:"created_at"`
    IsRevoked bool       `json:"is_revoked"`
    RevokedAt *time.Time `json:"revoked_at,omitempty"`
}
```

### API Endpoints

| Endpoint | Method | Auth Required | Description |
|----------|--------|---------------|-------------|
| `/api/v1/auth/register` | POST | No | Register new user account |
| `/api/v1/auth/login` | POST | No | Login with email/password |
| `/api/v1/auth/refresh` | POST | No | Refresh access token |
| `/api/v1/auth/logout` | POST | No | Revoke refresh token |
| `/api/v1/auth/me` | GET | Yes | Get current user profile |

## Task Index

<!-- MVP-AUTH-001 -->
### ✅ MVP-AUTH-001: User Model & Repository

**Status**: Complete  
**Priority**: P0  
**Effort**: Low  
**Dependencies**: None

**Implementation**:
- Created `User` model in `internal/auth/models.go`
- Implemented ArangoDB repository in `internal/auth/repository.go`
- CRUD operations: CreateUser, GetUserByID, GetUserByEmail, UpdateUser
- Password hashing with bcrypt (cost factor 12)
- Email validation via struct tags
- Automatic timestamps (created_at, updated_at)
- User activation status tracking

**Files Modified**:
- `internal/auth/models.go`
- `internal/auth/repository.go`

**Acceptance Criteria**: ✅ All Met
- ✅ User model with all required fields
- ✅ ArangoDB repository with CRUD operations
- ✅ Password hashing with bcrypt
- ✅ Email validation
- ✅ Collection creation on initialization

---

<!-- MVP-AUTH-002 -->
### ✅ MVP-AUTH-002: JWT Token Service

**Status**: Complete  
**Priority**: P0  
**Effort**: Medium  
**Dependencies**: MVP-AUTH-001 ✅

**Implementation**:
- JWT token generation/validation in `internal/auth/service.go`
- Access tokens: 15 minute expiry, signed with HS256
- Refresh tokens: 7 day expiry, stored as SHA-256 hashes
- Token claims include: user_id, email, name, iat, exp
- Refresh token storage in ArangoDB
- Token revocation via database flags

**Key Functions**:
- `generateAccessToken()` - Creates JWT access token
- `generateAndSaveRefreshToken()` - Creates and stores refresh token
- `ValidateAccessToken()` - Validates JWT and extracts claims
- `hashToken()` - SHA-256 hashing for refresh tokens

**Files Modified**:
- `internal/auth/service.go`
- `internal/auth/repository.go` (refresh token operations)

**Acceptance Criteria**: ✅ All Met
- ✅ Access tokens with 15min expiry
- ✅ Refresh tokens with 7 day expiry
- ✅ Secure token storage (hashed)
- ✅ Token validation logic
- ✅ Token revocation support

---

<!-- MVP-AUTH-003 -->
### ✅ MVP-AUTH-003: Authentication Endpoints

**Status**: Complete  
**Priority**: P0  
**Effort**: Medium  
**Dependencies**: MVP-AUTH-002 ✅

**Implementation**:
- REST API endpoints in `internal/auth/handler.go`
- Route registration in `internal/app/routes_api.go`
- Request/response models in `internal/auth/models.go`
- Comprehensive error handling
- Input validation using Gin binding

**Endpoints Implemented**:

1. **POST /api/v1/auth/register**
   - Request: `{"email": "...", "name": "...", "password": "..."}`
   - Response: `{"message": "...", "user": {...}}`
   - Validation: Email format, password length (min 8 chars)

2. **POST /api/v1/auth/login**
   - Request: `{"email": "...", "password": "..."}`
   - Response: `{"access_token": "...", "refresh_token": "...", "token_type": "Bearer", "expires_in": 900, "user": {...}}`
   - Returns JWT tokens on successful authentication

3. **POST /api/v1/auth/refresh**
   - Request: `{"refresh_token": "..."}`
   - Response: Same as login
   - Validates refresh token and issues new access token

4. **POST /api/v1/auth/logout**
   - Request: `{"refresh_token": "..."}`
   - Response: `{"message": "Logged out successfully"}`
   - Revokes refresh token

5. **GET /api/v1/auth/me**
   - Headers: `Authorization: Bearer <access_token>`
   - Response: `{"id": "...", "email": "...", "name": "...", ...}`
   - Requires authentication middleware

**Files Modified**:
- `internal/auth/handler.go`
- `internal/auth/models.go`
- `internal/app/routes_api.go`

**Bug Fixes Applied**:
- Fixed `GetUserByID` to handle full document IDs (e.g., "users/123" → "123")
- Added proper error messages for each endpoint
- Implemented proper HTTP status codes

**Acceptance Criteria**: ✅ All Met
- ✅ All 5 endpoints implemented
- ✅ Proper input validation
- ✅ Error handling with appropriate status codes
- ✅ Route registration in API router
- ✅ Tested and working

---

<!-- MVP-AUTH-004 -->
### ✅ MVP-AUTH-004: Authentication Middleware

**Status**: Complete  
**Priority**: P0  
**Effort**: Medium  
**Dependencies**: MVP-AUTH-003 ✅

**Implementation**:
- Created `internal/auth/middleware.go` with JWT validation middleware
- Implements `RequireAuth()` for protected routes
- Implements `OptionalAuth()` for conditionally authenticated routes
- Extracts user context from JWT and attaches to Gin context
- Proper 401 responses for invalid/expired tokens

**Middleware Functions**:

1. **RequireAuth()** - Mandatory authentication
   - Validates `Authorization: Bearer <token>` header
   - Extracts and validates JWT claims
   - Sets context variables: `user_id`, `user_email`, `user_name`
   - Returns 401 if token missing/invalid

2. **OptionalAuth()** - Optional authentication
   - Same validation as RequireAuth
   - Continues without error if token missing
   - Useful for endpoints that work with/without auth

**Context Variables Set**:
- `user_id` (string) - User's document ID
- `user_email` (string) - User's email address
- `user_name` (string) - User's display name

**Files Created**:
- `internal/auth/middleware.go`

**Files Modified**:
- `internal/auth/handler.go` - Updated to use middleware

**Acceptance Criteria**: ✅ All Met
- ✅ JWT validation middleware created
- ✅ User context extracted from token
- ✅ Context attached to Gin requests
- ✅ Proper 401 responses for invalid tokens
- ✅ Applied to `/me` endpoint

---

<!-- MVP-AUTH-005 -->
### 📋 MVP-AUTH-005: Protected Routes Integration

**Status**: Not Started  
**Priority**: P0  
**Effort**: Low  
**Dependencies**: MVP-AUTH-004 ✅

**Requirements**:
1. Apply authentication middleware to all protected routes:
   - Agency CRUD endpoints (`/api/v1/agencies/*`)
   - Work item endpoints (`/api/v1/work-items/*`)
   - Instance management endpoints (`/api/v1/agencies/:id/instances/*`)
   - Other domain-specific endpoints

2. Replace hardcoded "system" user with actual user context:
   - Update handlers to get `user_id` from context
   - Use `c.GetString("user_id")` instead of `"system"`
   - Add user tracking to audit logs

3. Add permission checks:
   - Verify user owns/has access to agency
   - Check user permissions for operations
   - Return 403 Forbidden for unauthorized actions

4. Update existing handlers:
   - Agency handlers (`internal/handlers/agency_handler.go`)
   - Work item handlers
   - Instance handlers
   - Any handler using placeholder auth

**Acceptance Criteria**:
- [ ] Authentication middleware applied to all protected routes
- [ ] No more hardcoded "system" or "user" strings in handlers
- [ ] User context properly extracted in all handlers
- [ ] Permission checks implemented for resource access
- [ ] 403 responses for unauthorized operations
- [ ] Audit logs include actual user_id

**Files to Modify**:
- `internal/app/routes_api.go` - Apply middleware to route groups
- `internal/handlers/agency_handler.go` - Replace "system" with user context
- `internal/handlers/*.go` - Update other handlers as needed
- `internal/middleware/*.go` - May need permission checking middleware

**Testing Requirements**:
- [ ] Test protected endpoints return 401 without token
- [ ] Test protected endpoints work with valid token
- [ ] Test permission checks prevent unauthorized access
- [ ] Test user context is properly used in operations

---

## Testing Summary

Comprehensive authentication flow tested successfully:

```bash
# 1. Register user
POST /api/v1/auth/register
✅ Returns 201 with user data

# 2. Login
POST /api/v1/auth/login
✅ Returns 200 with access_token, refresh_token, user data

# 3. Access protected endpoint
GET /api/v1/auth/me (with Authorization header)
✅ Returns 200 with user profile

# 4. Refresh access token
POST /api/v1/auth/refresh
✅ Returns 200 with new access_token

# 5. Use new token
GET /api/v1/auth/me (with new token)
✅ Returns 200 with user profile

# Error cases
GET /api/v1/auth/me (no auth header)
✅ Returns 401 "Authorization header required"

GET /api/v1/auth/me (invalid header format)
✅ Returns 401 "Invalid authorization header format"
```

## Implementation Notes

### Bug Fixes Applied

1. **Refresh Token Bug** (Fixed):
   - **Issue**: Refresh token endpoint returned "invalid refresh token"
   - **Root Cause**: `GetUserByID` expected document key but received full ID
   - **Fix**: Extract key from ID by splitting on "/" (e.g., "users/123" → "123")
   - **File**: `internal/auth/repository.go`

2. **Missing Middleware** (Fixed):
   - **Issue**: `/me` endpoint returned "Unauthorized" even with valid token
   - **Root Cause**: No authentication middleware protecting the endpoint
   - **Fix**: Created `middleware.go` and applied to protected routes
   - **Files**: `internal/auth/middleware.go`, `internal/auth/handler.go`

### Security Considerations

- Passwords never returned in API responses
- Refresh tokens hashed before storage (SHA-256)
- Access tokens short-lived (15 minutes)
- Token revocation supported
- Bcrypt cost factor 12 for password hashing
- Email validation on registration
- User activation status checking

### Frontend Integration

The authentication system is ready for CodeValdFortex (Flutter frontend) integration:
- All required endpoints implemented
- JWT tokens in standard format
- RESTful API design
- Proper error responses with status codes
- CORS headers configured (via API middleware)

## Next Steps

1. **Complete MVP-AUTH-005** - Protected Routes Integration
2. **P2 Tasks** (Future):
   - MVP-026: Password reset & email verification
   - MVP-027: Security hardening (rate limiting, brute force protection)
   - MVP-028: RBAC system for fine-grained permissions

## References

**Files**:
- `internal/auth/models.go` - Data models and request/response types
- `internal/auth/repository.go` - ArangoDB operations
- `internal/auth/service.go` - Business logic and JWT handling
- `internal/auth/handler.go` - HTTP handlers
- `internal/auth/middleware.go` - Authentication middleware
- `internal/app/routes_api.go` - Route registration

**Dependencies**:
- `github.com/golang-jwt/jwt/v5` - JWT token handling
- `golang.org/x/crypto/bcrypt` - Password hashing
- `github.com/gin-gonic/gin` - HTTP framework
- `github.com/arangodb/go-driver` - Database driver
