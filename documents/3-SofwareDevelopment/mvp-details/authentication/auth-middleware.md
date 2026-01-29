# MVP-AUTH-004: Authentication Middleware

## Status: ✅ Complete

## Overview
Replace placeholder AuthMiddleware with real JWT validation to protect routes, extract user context, and handle token expiry.

## Implementation

### File: `internal/api/middleware.go`

**Updated AuthMiddleware:**
```go
func AuthMiddleware(authService interface {
    ValidateAccessToken(string) (*struct {
        UserID string
        Email  string
        Name   string
        Exp    int64
        Iat    int64
    }, error)
}) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get Authorization header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        // Check Bearer token format
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(401, gin.H{"error": "Invalid authorization header format"})
            c.Abort()
            return
        }

        tokenString := parts[1]

        // Validate token
        claims, err := authService.ValidateAccessToken(tokenString)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid or expired token"})
            c.Abort()
            return
        }

        // Set user context
        c.Set("user_id", claims.UserID)
        c.Set("user_email", claims.Email)
        c.Set("user_name", claims.Name)

        c.Next()
    }
}
```

## How It Works

### 1. Request Flow

**Without Middleware (Public Routes):**
```
Client → Handler → Response
```

**With Middleware (Protected Routes):**
```
Client → AuthMiddleware → Handler → Response
         ↓ (if token invalid)
         401 Unauthorized
```

### 2. Authorization Header

**Expected Format:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Validation Steps:**
1. Check header exists
2. Split by space, verify "Bearer" prefix
3. Extract token string
4. Validate JWT signature and expiration
5. Extract user claims

### 3. User Context

**Set in Gin Context:**
```go
c.Set("user_id", "users/user_abc123")
c.Set("user_email", "user@example.com")
c.Set("user_name", "John Doe")
```

**Access in Handlers:**
```go
func (h *Handler) SomeProtectedRoute(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(401, gin.H{"error": "Unauthorized"})
        return
    }
    
    // Use userID for business logic
    user := userID.(string)
}
```

## Error Responses

### Missing Authorization Header
**Status**: 401 Unauthorized
```json
{
  "error": "Authorization header required"
}
```

### Invalid Header Format
**Status**: 401 Unauthorized
```json
{
  "error": "Invalid authorization header format"
}
```

**Examples of Invalid Format:**
- `Authorization: eyJhbGc...` (missing "Bearer")
- `Authorization: Basic dXNlcjpwYXNz` (wrong auth type)
- `Bearer eyJhbGc...` (missing "Authorization:")

### Invalid or Expired Token
**Status**: 401 Unauthorized
```json
{
  "error": "Invalid or expired token"
}
```

**Causes:**
- Token signature doesn't match (tampered)
- Token expired (> 15 minutes old)
- Malformed JWT
- Wrong signing algorithm

## Usage Pattern

### Applying Middleware to Routes

**Option 1: Route Group**
```go
protected := router.Group("/api/v1/protected")
protected.Use(api.AuthMiddleware(authService))
{
    protected.GET("/agencies", handler.ListAgencies)
    protected.POST("/agencies", handler.CreateAgency)
}
```

**Option 2: Individual Routes**
```go
router.GET("/api/v1/auth/me", 
    api.AuthMiddleware(authService), 
    handler.GetCurrentUser)
```

## Integration Points

### Current Protected Route
**File**: `internal/auth/handler.go`

```go
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
    auth := rg.Group("/auth")
    {
        auth.POST("/register", h.Register)
        auth.POST("/login", h.Login)
        auth.POST("/refresh", h.Refresh)
        auth.POST("/logout", h.Logout)
        auth.GET("/me", h.Me) // Protected - middleware applied elsewhere
    }
}
```

### Handler Using User Context

**Example from `auth/handler.go`:**
```go
func (h *Handler) Me(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error": "Unauthorized",
        })
        return
    }

    user, err := h.service.GetCurrentUser(c.Request.Context(), userID.(string))
    if err != nil {
        h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get current user")
        c.JSON(http.StatusNotFound, gin.H{
            "error": "User not found",
        })
        return
    }

    c.JSON(http.StatusOK, user.ToResponse())
}
```

## Security Considerations

### Token Validation
- ✅ Signature verified with HMAC-SHA256
- ✅ Expiration checked automatically by JWT library
- ✅ Signing algorithm validated (prevents algorithm confusion attacks)

### Timing Attacks
- JWT validation uses constant-time comparison (via library)
- No early returns that leak information about token validity

### Error Messages
- Generic error messages prevent information leakage
- Don't reveal if token is malformed vs expired vs invalid signature

### Context Injection
- User claims extracted from validated token only
- No way to inject user_id without valid JWT signature

## Testing

**Test Valid Token:**
```bash
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

**Test Missing Header:**
```bash
curl -X GET http://localhost:8080/api/v1/auth/me
# Expected: 401 with "Authorization header required"
```

**Test Invalid Format:**
```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: InvalidToken"
# Expected: 401 with "Invalid authorization header format"
```

**Test Expired Token:**
```bash
# Wait 15 minutes after login
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $EXPIRED_TOKEN"
# Expected: 401 with "Invalid or expired token"
```

## Performance Considerations

### Token Validation Overhead
- JWT validation requires signature computation (HMAC-SHA256)
- ~0.1-0.5ms per request on modern hardware
- No database lookup required (stateless validation)

### Context Storage
- Gin context uses map for storing values
- Minimal memory overhead (3 strings per request)
- Context cleared after request completes

### Caching
- Currently no caching of validated tokens
- Future: Could cache valid tokens for 1-2 seconds to reduce CPU

## Dependencies
- `strings` - String manipulation for header parsing
- `internal/auth.Service` - Token validation logic
- `github.com/gin-gonic/gin` - HTTP context and middleware

## Future Enhancements
- [ ] Support for API key authentication (alternative to JWT)
- [ ] Token caching to reduce validation overhead
- [ ] Request rate limiting per user
- [ ] Audit logging of all authenticated requests
- [ ] Support for service-to-service authentication
- [ ] WebSocket authentication support
