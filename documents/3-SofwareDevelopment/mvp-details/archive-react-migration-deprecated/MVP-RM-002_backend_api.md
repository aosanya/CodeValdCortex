# MVP-RM-002: Backend API Infrastructure

**Domain**: React Migration  
**Priority**: P1  
**Effort**: Medium  
**Skills**: Go, Gin Framework, REST API Design  
**Dependencies**: None

---

## Overview

Create the foundational REST API infrastructure in CodeValdCortex backend to support the React frontend. This includes establishing the `/api/v1` package structure, common middleware, error handling, and response formatting.

---

## Requirements

### 1. Package Structure

Create new package structure in CodeValdCortex:
```
internal/
├── api/
│   ├── router.go           # Main API router setup
│   ├── middleware/         # API middleware
│   │   ├── cors.go         # CORS configuration
│   │   ├── auth.go         # Authentication middleware
│   │   └── logger.go       # Request logging
│   ├── response/           # Response formatting
│   │   ├── success.go      # Success responses
│   │   └── error.go        # Error responses
│   └── v1/                 # API version 1
│       ├── router.go       # v1 route registration
│       └── workitems/      # Work items endpoints (created in MVP-RM-003)
```

### 2. API Router Setup

**Main API Router (`internal/api/router.go`):**
```go
package api

import (
    "github.com/gin-gonic/gin"
    "github.com/yourusername/codevaldcortex/internal/api/middleware"
    v1 "github.com/yourusername/codevaldcortex/internal/api/v1"
)

// SetupRouter creates and configures the API router
func SetupRouter(r *gin.Engine) {
    // Apply middleware
    r.Use(middleware.CORS())
    r.Use(middleware.RequestLogger())
    
    // API v1 routes
    apiV1 := r.Group("/api/v1")
    {
        v1.RegisterRoutes(apiV1)
    }
}
```

**V1 Router (`internal/api/v1/router.go`):**
```go
package v1

import (
    "github.com/gin-gonic/gin"
)

// RegisterRoutes registers all v1 API routes
func RegisterRoutes(rg *gin.RouterGroup) {
    // Health check
    rg.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "ok",
            "version": "v1",
        })
    })
    
    // Work items routes will be registered here in MVP-RM-003
    // RegisterWorkItemsRoutes(rg)
}
```

### 3. CORS Middleware

**CORS Configuration (`internal/api/middleware/cors.go`):**
```go
package middleware

import (
    "github.com/gin-gonic/gin"
)

// CORS middleware for React frontend
func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Development: Allow localhost:5173
        // Production: Use environment variable for allowed origins
        origin := c.GetHeader("Origin")
        
        // In development, allow localhost:5173
        if origin == "http://localhost:5173" || origin == "http://localhost:3000" {
            c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
        }
        
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}
```

### 4. Request Logger Middleware

**Logger Middleware (`internal/api/middleware/logger.go`):**
```go
package middleware

import (
    "time"

    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
)

// RequestLogger logs API requests
func RequestLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        startTime := time.Now()
        
        // Process request
        c.Next()
        
        // Log after request
        duration := time.Since(startTime)
        
        logrus.WithFields(logrus.Fields{
            "method":     c.Request.Method,
            "path":       c.Request.URL.Path,
            "status":     c.Writer.Status(),
            "duration":   duration.Milliseconds(),
            "client_ip":  c.ClientIP(),
            "user_agent": c.Request.UserAgent(),
        }).Info("API request")
    }
}
```

### 5. Authentication Middleware (Placeholder)

**Auth Middleware (`internal/api/middleware/auth.go`):**
```go
package middleware

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/yourusername/codevaldcortex/internal/api/response"
)

// RequireAuth ensures request is authenticated
// Currently uses session-based auth (existing system)
// Will be enhanced with JWT in future phases
func RequireAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // TODO: Implement proper session validation
        // For now, allow all requests in development
        
        // Example session check:
        // session, err := getSession(c)
        // if err != nil || session == nil {
        //     response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
        //     c.Abort()
        //     return
        // }
        
        c.Next()
    }
}
```

### 6. Response Formatting

**Success Response (`internal/api/response/success.go`):**
```go
package response

import (
    "github.com/gin-gonic/gin"
)

// SuccessResponse represents a successful API response
type SuccessResponse struct {
    Data       interface{}            `json:"data"`
    Pagination *PaginationMeta        `json:"pagination,omitempty"`
    Meta       map[string]interface{} `json:"meta,omitempty"`
}

// PaginationMeta contains pagination information
type PaginationMeta struct {
    Total   int  `json:"total"`
    Limit   int  `json:"limit"`
    Offset  int  `json:"offset"`
    HasMore bool `json:"has_more"`
}

// Success sends a successful response
func Success(c *gin.Context, statusCode int, data interface{}) {
    c.JSON(statusCode, SuccessResponse{
        Data: data,
    })
}

// SuccessWithPagination sends a successful response with pagination
func SuccessWithPagination(c *gin.Context, statusCode int, data interface{}, pagination *PaginationMeta) {
    c.JSON(statusCode, SuccessResponse{
        Data:       data,
        Pagination: pagination,
    })
}
```

**Error Response (`internal/api/response/error.go`):**
```go
package response

import (
    "github.com/gin-gonic/gin"
)

// ErrorResponse represents an API error response
type ErrorResponse struct {
    Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}

// Error sends an error response
func Error(c *gin.Context, statusCode int, code string, message string, details map[string]interface{}) {
    c.JSON(statusCode, ErrorResponse{
        Error: ErrorDetail{
            Code:    code,
            Message: message,
            Details: details,
        },
    })
}

// ValidationError sends a validation error response
func ValidationError(c *gin.Context, field string, reason string) {
    Error(c, 400, "INVALID_INPUT", "Validation failed", map[string]interface{}{
        "field":  field,
        "reason": reason,
    })
}

// NotFoundError sends a not found error response
func NotFoundError(c *gin.Context, resource string, id string) {
    Error(c, 404, "NOT_FOUND", resource+" not found", map[string]interface{}{
        "resource": resource,
        "id":       id,
    })
}

// InternalError sends an internal server error response
func InternalError(c *gin.Context, err error) {
    // Log the actual error for debugging
    // Don't expose internal error details to client
    Error(c, 500, "INTERNAL_ERROR", "An unexpected error occurred", nil)
}
```

### 7. Integration with Main App

**Update `cmd/main.go`:**
```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/yourusername/codevaldcortex/internal/api"
    "github.com/yourusername/codevaldcortex/internal/app"
    // ... other imports
)

func main() {
    // ... existing initialization code
    
    // Create Gin router
    r := gin.Default()
    
    // Setup API routes
    api.SetupRouter(r)
    
    // ... existing web routes (Templ pages)
    // These continue to work alongside the new API
    
    // Start server
    r.Run(":8080")
}
```

### 8. Configuration Updates

**Update `config.yaml`:**
```yaml
api:
  version: "v1"
  cors:
    allowed_origins:
      - "http://localhost:5173"  # Development React app
      - "http://localhost:3000"  # Alternative dev port
    allowed_methods:
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
      - "PATCH"
      - "OPTIONS"
    allowed_headers:
      - "Content-Type"
      - "Authorization"
      - "X-Requested-With"
    allow_credentials: true
```

---

## Acceptance Criteria

- [ ] `/api/v1` package structure created
- [ ] CORS middleware configured and tested
- [ ] Request logger middleware working
- [ ] Response formatting (success/error) implemented
- [ ] Health check endpoint accessible: `GET /api/v1/health` returns 200
- [ ] CORS headers present in responses
- [ ] OPTIONS requests handled correctly (preflight)
- [ ] Error responses follow standard format
- [ ] Success responses follow standard format
- [ ] API routes integrated with main app
- [ ] Existing Templ routes continue to work (backward compatible)
- [ ] Configuration updated with API settings
- [ ] All code follows Go conventions (gofmt, golint)

---

## Testing

### Manual Testing

**Test Health Endpoint:**
```bash
# From terminal
curl -X GET http://localhost:8080/api/v1/health

# Expected response:
{
  "status": "ok",
  "version": "v1"
}
```

**Test CORS:**
```bash
# Preflight request
curl -X OPTIONS http://localhost:8080/api/v1/health \
  -H "Origin: http://localhost:5173" \
  -H "Access-Control-Request-Method: GET"

# Should return 204 with CORS headers
```

**Test from React App:**
```typescript
// In React app, create a test component
useEffect(() => {
  fetch('http://localhost:8080/api/v1/health')
    .then(res => res.json())
    .then(data => console.log('Health check:', data))
    .catch(err => console.error('Error:', err));
}, []);
```

### Automated Testing

**Unit Tests (`internal/api/middleware/cors_test.go`):**
```go
package middleware_test

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/yourusername/codevaldcortex/internal/api/middleware"
)

func TestCORS(t *testing.T) {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.Use(middleware.CORS())
    r.GET("/test", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // Test OPTIONS request
    req := httptest.NewRequest("OPTIONS", "/test", nil)
    req.Header.Set("Origin", "http://localhost:5173")
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    assert.Equal(t, 204, w.Code)
    assert.Equal(t, "http://localhost:5173", w.Header().Get("Access-Control-Allow-Origin"))
}
```

**Response Format Tests (`internal/api/response/response_test.go`):**
```go
package response_test

import (
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/yourusername/codevaldcortex/internal/api/response"
)

func TestSuccessResponse(t *testing.T) {
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    response.Success(c, 200, gin.H{"message": "test"})

    assert.Equal(t, 200, w.Code)
    assert.Contains(t, w.Body.String(), "\"data\"")
}
```

---

## Implementation Notes

### File Size Guidelines
- Keep middleware files under 150 lines each
- Response package files under 100 lines each
- Router files under 200 lines

### Code Quality
- Run `go fmt` before committing
- Run `go vet` to catch issues
- Run `golangci-lint` if available
- Ensure all functions have comments

### Error Handling
- Never expose internal error details to API responses
- Log actual errors server-side
- Return generic error messages to client
- Use appropriate HTTP status codes

---

## Dependencies

**None** - This is a foundation task for the backend

---

## Related Tasks

- **MVP-RM-001**: Project Setup (React frontend)
- **MVP-RM-003**: Work Items REST API (first API endpoints)
- **MVP-RM-004**: Work Items Redux Store (consumes this API)

---

## References

- [Gin Web Framework](https://gin-gonic.com/docs/)
- [REST API Best Practices](https://restfulapi.net/)
- [CORS Documentation](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)

---

**Created**: January 26, 2026  
**Last Updated**: January 26, 2026
