package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// LoggingMiddleware logs HTTP requests with structured logging
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		requestID := getRequestID(c)

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Build full path with query parameters
		if raw != "" {
			path = path + "?" + raw
		}

		// Log request details
		entry := log.WithFields(log.Fields{
			"request_id":    requestID,
			"method":        c.Request.Method,
			"path":          path,
			"status":        c.Writer.Status(),
			"latency":       latency,
			"client_ip":     c.ClientIP(),
			"user_agent":    c.Request.UserAgent(),
			"response_size": c.Writer.Size(),
		})

		// Log level based on status code
		status := c.Writer.Status()
		switch {
		case status >= 500:
			entry.Error("HTTP request completed")
		case status >= 400:
			entry.Warn("HTTP request completed")
		default:
			entry.Info("HTTP request completed")
		}
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Set CORS headers
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Request-ID, X-Forwarded-For")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RecoveryMiddleware handles panics and returns 500 errors gracefully
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID := getRequestID(c)

		log.WithFields(log.Fields{
			"request_id": requestID,
			"panic":      recovered,
			"path":       c.Request.URL.Path,
			"method":     c.Request.Method,
		}).Error("Panic recovered in HTTP handler")

		InternalError(c, "Internal server error", map[string]interface{}{
			"request_id": requestID,
		})
	})
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")

		// Remove server information
		c.Header("Server", "")

		c.Next()
	}
}

// RateLimitMiddleware implements basic rate limiting (placeholder)
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement actual rate limiting
		// For now, just pass through
		c.Next()
	}
}

// AuthMiddleware handles authentication (placeholder for future implementation)
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement authentication
		// For now, just pass through
		c.Next()
	}
}

// HealthCheckMiddleware bypasses other middleware for health checks
func HealthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/api/v1/health" {
			c.Next()
			return
		}
		c.Next()
	}
}

// ValidateContentTypeMiddleware ensures correct content type for POST/PUT requests
func ValidateContentTypeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if contentType != "" && contentType != "application/json" {
				BadRequestError(c, "Content-Type must be application/json", map[string]string{
					"received": contentType,
					"expected": "application/json",
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// RequestSizeLimitMiddleware limits request body size
func RequestSizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			BadRequestError(c, "Request body too large", map[string]interface{}{
				"max_size": maxSize,
				"received": c.Request.ContentLength,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// TimeoutMiddleware adds request timeout (placeholder)
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement request timeout
		// For now, just pass through
		c.Next()
	}
}
