package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/aosanya/CodeValdCortex/internal/pool"
)

// PoolHandler handles HTTP requests for pool management
type PoolHandler struct {
	poolManager *pool.Manager
}

// NewPoolHandler creates a new pool handler
func NewPoolHandler(poolManager *pool.Manager) *PoolHandler {
	return &PoolHandler{
		poolManager: poolManager,
	}
}

// CreatePoolRequest represents a request to create a new pool
type CreatePoolRequest struct {
	Name                  string                        `json:"name"`
	Description           string                        `json:"description"`
	LoadBalancingStrategy pool.LoadBalancingStrategy    `json:"load_balancing_strategy"`
	MinAgents             int                           `json:"min_agents"`
	MaxAgents             int                           `json:"max_agents"`
	HealthCheckInterval   int64                         `json:"health_check_interval_ms"`
	ResourceLimits        pool.ResourceLimits           `json:"resource_limits"`
	AutoScaling           pool.AutoScalingConfig        `json:"auto_scaling"`
}

// PoolResponse represents a pool in API responses
type PoolResponse struct {
	ID                    string                        `json:"id"`
	Name                  string                        `json:"name"`
	Description           string                        `json:"description"`
	LoadBalancingStrategy pool.LoadBalancingStrategy    `json:"load_balancing_strategy"`
	MinAgents             int                           `json:"min_agents"`
	MaxAgents             int                           `json:"max_agents"`
	HealthCheckInterval   int64                         `json:"health_check_interval_ms"`
	ResourceLimits        pool.ResourceLimits           `json:"resource_limits"`
	AutoScaling           pool.AutoScalingConfig        `json:"auto_scaling"`
	Status                pool.PoolStatus               `json:"status"`
	Metrics               *pool.PoolMetrics             `json:"metrics,omitempty"`
	CreatedAt             time.Time                     `json:"created_at"`
	UpdatedAt             time.Time                     `json:"updated_at"`
}

// CreatePool handles POST /api/v1/pools
func (ph *PoolHandler) CreatePool(c *gin.Context) {
	var req CreatePoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Validate request
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pool name is required"})
		return
	}

	if req.MaxAgents <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Max agents must be greater than 0"})
		return
	}

	if req.MinAgents < 0 || req.MinAgents > req.MaxAgents {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid min agents value"})
		return
	}

	// Set defaults
	if req.HealthCheckInterval <= 0 {
		req.HealthCheckInterval = 30000 // 30 seconds
	}

	if req.LoadBalancingStrategy == "" {
		req.LoadBalancingStrategy = pool.LoadBalancingRoundRobin
	}

	// Create pool config
	config := pool.PoolConfig{
		Name:                  req.Name,
		Description:           req.Description,
		LoadBalancingStrategy: req.LoadBalancingStrategy,
		MinAgents:             req.MinAgents,
		MaxAgents:             req.MaxAgents,
		HealthCheckInterval:   time.Duration(req.HealthCheckInterval) * time.Millisecond,
		ResourceLimits:        req.ResourceLimits,
		AutoScaling:           req.AutoScaling,
	}

	// Create pool
	createdPool, err := ph.poolManager.CreatePool(c.Request.Context(), config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pool: " + err.Error()})
		return
	}

	// Convert to response
	response := ph.poolToResponse(createdPool, true)

	c.JSON(http.StatusCreated, response)
}

// ListPools handles GET /api/v1/pools
func (ph *PoolHandler) ListPools(c *gin.Context) {
	// Parse query parameters
	statusFilter := pool.PoolStatus(c.Query("status"))

	// Get pools
	pools, err := ph.poolManager.ListPools(c.Request.Context(), statusFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list pools: " + err.Error()})
		return
	}

	// Convert to responses
	var responses []PoolResponse
	for _, p := range pools {
		response := ph.poolToResponse(p, false) // Don't include metrics in list
		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, responses)
}

// GetPool handles GET /api/v1/pools/{id}
func (ph *PoolHandler) GetPool(c *gin.Context) {
	poolID := c.Param("id")

	if poolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pool ID is required"})
		return
	}

	// Get pool
	p, err := ph.poolManager.GetPool(c.Request.Context(), poolID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pool not found: " + err.Error()})
		return
	}

	// Convert to response with metrics
	response := ph.poolToResponse(p, true)

	c.JSON(http.StatusOK, response)
}

// DeletePool handles DELETE /api/v1/pools/{id}
func (ph *PoolHandler) DeletePool(c *gin.Context) {
	poolID := c.Param("id")

	if poolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pool ID is required"})
		return
	}

	// Delete pool
	err := ph.poolManager.DeletePool(c.Request.Context(), poolID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete pool: " + err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetPoolMetrics handles GET /api/v1/pools/{id}/metrics
func (ph *PoolHandler) GetPoolMetrics(c *gin.Context) {
	poolID := c.Param("id")

	if poolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pool ID is required"})
		return
	}

	// Get pool metrics
	metrics, err := ph.poolManager.GetPoolMetrics(c.Request.Context(), poolID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pool metrics: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// Helper methods

// poolToResponse converts a pool to API response format
func (ph *PoolHandler) poolToResponse(p *pool.AgentPool, includeMetrics bool) PoolResponse {
	response := PoolResponse{
		ID:                    p.ID,
		Name:                  p.Config.Name,
		Description:           p.Config.Description,
		LoadBalancingStrategy: p.Config.LoadBalancingStrategy,
		MinAgents:             p.Config.MinAgents,
		MaxAgents:             p.Config.MaxAgents,
		HealthCheckInterval:   p.Config.HealthCheckInterval.Milliseconds(),
		ResourceLimits:        p.Config.ResourceLimits,
		AutoScaling:           p.Config.AutoScaling,
		Status:                p.Status,
		CreatedAt:             p.CreatedAt,
		UpdatedAt:             p.UpdatedAt,
	}

	if includeMetrics {
		response.Metrics = p.GetMetrics(nil)
	}

	return response
}

// RegisterRoutes registers pool management routes with the router
func (ph *PoolHandler) RegisterRoutes(router *gin.Engine) {
	pools := router.Group("/api/v1/pools")
	{
		pools.POST("", ph.CreatePool)
		pools.GET("", ph.ListPools)
		pools.GET("/:id", ph.GetPool)
		pools.DELETE("/:id", ph.DeletePool)
		pools.GET("/:id/metrics", ph.GetPoolMetrics)
	}
}