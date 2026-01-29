package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Handler handles authentication HTTP requests
type Handler struct {
	service    *Service
	middleware *Middleware
	logger     *logrus.Logger
}

// NewHandler creates a new auth handler
func NewHandler(service *Service, logger *logrus.Logger) *Handler {
	return &Handler{
		service:    service,
		middleware: NewMiddleware(service, logger),
		logger:     logger,
	}
}

// Register handles POST /api/v1/auth/register
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid registration request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate password strength (basic validation)
	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Password must be at least 8 characters long",
		})
		return
	}

	user, err := h.service.Register(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to register user")
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "User with this email already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to register user",
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("User registered successfully")

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    user.ToResponse(),
	})
}

// Login handles POST /api/v1/auth/login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid login request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	authResponse, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).WithField("email", req.Email).Warn("Login failed")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": authResponse.User.ID,
		"email":   authResponse.User.Email,
	}).Info("User logged in successfully")

	c.JSON(http.StatusOK, authResponse)
}

// Refresh handles POST /api/v1/auth/refresh
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid refresh token request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	authResponse, err := h.service.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.logger.WithError(err).Warn("Token refresh failed")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid or expired refresh token",
		})
		return
	}

	h.logger.WithField("user_id", authResponse.User.ID).Info("Access token refreshed")

	c.JSON(http.StatusOK, authResponse)
}

// Logout handles POST /api/v1/auth/logout
func (h *Handler) Logout(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no refresh token provided in body, try to get from auth context
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Refresh token required",
			})
			return
		}

		h.logger.WithField("user_id", userID).Info("User logged out (token revocation skipped)")
		c.JSON(http.StatusOK, gin.H{
			"message": "Logged out successfully",
		})
		return
	}

	if err := h.service.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		h.logger.WithError(err).Warn("Failed to revoke refresh token")
		// Don't fail logout if token revocation fails
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// Me handles GET /api/v1/auth/me
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

// RegisterRoutes registers authentication routes
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		// Public routes
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/logout", h.Logout)

		// Protected routes (require authentication)
		auth.GET("/me", h.middleware.RequireAuth(), h.Me)
	}
}
