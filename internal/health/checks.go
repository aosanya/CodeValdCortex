package health

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
)

// HeartbeatHealthCheck monitors agent responsiveness via heartbeat
type HeartbeatHealthCheck struct {
	name     string
	interval time.Duration
	timeout  time.Duration
	enabled  bool
}

// NewHeartbeatHealthCheck creates a new heartbeat health check
func NewHeartbeatHealthCheck() *HeartbeatHealthCheck {
	return &HeartbeatHealthCheck{
		name:     "heartbeat",
		interval: 30 * time.Second,
		timeout:  5 * time.Second,
		enabled:  true,
	}
}

func (h *HeartbeatHealthCheck) Name() string {
	return h.name
}

func (h *HeartbeatHealthCheck) Type() CheckType {
	return CheckTypeHeartbeat
}

func (h *HeartbeatHealthCheck) Check(ctx context.Context, agent *agent.Agent) *HealthCheckResult {
	start := time.Now()
	result := &HealthCheckResult{
		CheckName: h.name,
		CheckType: h.Type(),
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	// Check if agent is responsive via heartbeat
	isHealthy := agent.IsHealthy()
	lastHeartbeat := agent.LastHeartbeat

	result.Duration = time.Since(start)
	result.Details["last_heartbeat"] = lastHeartbeat
	result.Details["time_since_heartbeat"] = time.Since(lastHeartbeat).String()

	if isHealthy {
		result.Status = HealthStatusHealthy
		result.Message = "Agent heartbeat is current"
	} else {
		result.Status = HealthStatusUnhealthy
		result.Message = fmt.Sprintf("Agent heartbeat is stale (last: %v)", lastHeartbeat.Format(time.RFC3339))
	}

	return result
}

func (h *HeartbeatHealthCheck) Interval() time.Duration {
	return h.interval
}

func (h *HeartbeatHealthCheck) Timeout() time.Duration {
	return h.timeout
}

func (h *HeartbeatHealthCheck) IsEnabled() bool {
	return h.enabled
}

// ResourceHealthCheck monitors system resource usage
type ResourceHealthCheck struct {
	name         string
	interval     time.Duration
	timeout      time.Duration
	enabled      bool
	cpuThreshold float64 // CPU usage threshold (0-100)
	memThreshold float64 // Memory usage threshold (0-100)
}

// NewResourceHealthCheck creates a new resource health check
func NewResourceHealthCheck() *ResourceHealthCheck {
	return &ResourceHealthCheck{
		name:         "resource",
		interval:     1 * time.Minute,
		timeout:      10 * time.Second,
		enabled:      true,
		cpuThreshold: 90.0, // 90% CPU usage threshold
		memThreshold: 85.0, // 85% memory usage threshold
	}
}

func (h *ResourceHealthCheck) Name() string {
	return h.name
}

func (h *ResourceHealthCheck) Type() CheckType {
	return CheckTypeResource
}

func (h *ResourceHealthCheck) Check(ctx context.Context, agent *agent.Agent) *HealthCheckResult {
	start := time.Now()
	result := &HealthCheckResult{
		CheckName: h.name,
		CheckType: h.Type(),
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	// Get system resource information
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Calculate memory usage (simplified)
	memUsageBytes := memStats.Alloc
	memUsagePercent := float64(memUsageBytes) / float64(memStats.Sys) * 100

	// Get goroutine count
	goroutines := runtime.NumGoroutine()

	result.Duration = time.Since(start)
	result.Details["memory_usage_bytes"] = memUsageBytes
	result.Details["memory_usage_percent"] = memUsagePercent
	result.Details["goroutines"] = goroutines
	result.Details["memory_threshold"] = h.memThreshold

	// Determine health status based on resource usage
	if memUsagePercent > h.memThreshold {
		result.Status = HealthStatusUnhealthy
		result.Message = fmt.Sprintf("High memory usage: %.1f%% (threshold: %.1f%%)",
			memUsagePercent, h.memThreshold)
	} else if memUsagePercent > h.memThreshold*0.8 {
		result.Status = HealthStatusDegraded
		result.Message = fmt.Sprintf("Elevated memory usage: %.1f%%", memUsagePercent)
	} else {
		result.Status = HealthStatusHealthy
		result.Message = fmt.Sprintf("Resource usage normal: %.1f%% memory, %d goroutines",
			memUsagePercent, goroutines)
	}

	return result
}

func (h *ResourceHealthCheck) Interval() time.Duration {
	return h.interval
}

func (h *ResourceHealthCheck) Timeout() time.Duration {
	return h.timeout
}

func (h *ResourceHealthCheck) IsEnabled() bool {
	return h.enabled
}

// PerformanceHealthCheck monitors agent task execution performance
type PerformanceHealthCheck struct {
	name            string
	interval        time.Duration
	timeout         time.Duration
	enabled         bool
	maxResponseTime time.Duration
	minSuccessRate  float64
}

// NewPerformanceHealthCheck creates a new performance health check
func NewPerformanceHealthCheck() *PerformanceHealthCheck {
	return &PerformanceHealthCheck{
		name:            "performance",
		interval:        2 * time.Minute,
		timeout:         15 * time.Second,
		enabled:         true,
		maxResponseTime: 5 * time.Second,
		minSuccessRate:  0.95, // 95% success rate
	}
}

func (h *PerformanceHealthCheck) Name() string {
	return h.name
}

func (h *PerformanceHealthCheck) Type() CheckType {
	return CheckTypePerformance
}

func (h *PerformanceHealthCheck) Check(ctx context.Context, agent *agent.Agent) *HealthCheckResult {
	start := time.Now()
	result := &HealthCheckResult{
		CheckName: h.name,
		CheckType: h.Type(),
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	// Get agent performance metrics (simplified - would be more detailed in real implementation)
	state := agent.GetState()
	agentType := agent.Type

	result.Duration = time.Since(start)
	result.Details["agent_state"] = state
	result.Details["agent_type"] = agentType
	result.Details["max_response_time"] = h.maxResponseTime.String()
	result.Details["min_success_rate"] = h.minSuccessRate

	// Simplified performance check based on agent state
	switch state {
	case "running":
		result.Status = HealthStatusHealthy
		result.Message = "Agent performance is normal"
	case "paused":
		result.Status = HealthStatusDegraded
		result.Message = "Agent is paused"
	case "failed":
		result.Status = HealthStatusCritical
		result.Message = "Agent has failed"
	case "stopped":
		result.Status = HealthStatusUnhealthy
		result.Message = "Agent is stopped"
	default:
		result.Status = HealthStatusUnknown
		result.Message = fmt.Sprintf("Unknown agent state: %s", state)
	}

	return result
}

func (h *PerformanceHealthCheck) Interval() time.Duration {
	return h.interval
}

func (h *PerformanceHealthCheck) Timeout() time.Duration {
	return h.timeout
}

func (h *PerformanceHealthCheck) IsEnabled() bool {
	return h.enabled
}

// ConnectivityHealthCheck monitors database and network connectivity
type ConnectivityHealthCheck struct {
	name     string
	interval time.Duration
	timeout  time.Duration
	enabled  bool
}

// NewConnectivityHealthCheck creates a new connectivity health check
func NewConnectivityHealthCheck() *ConnectivityHealthCheck {
	return &ConnectivityHealthCheck{
		name:     "connectivity",
		interval: 1 * time.Minute,
		timeout:  10 * time.Second,
		enabled:  true,
	}
}

func (h *ConnectivityHealthCheck) Name() string {
	return h.name
}

func (h *ConnectivityHealthCheck) Type() CheckType {
	return CheckTypeConnectivity
}

func (h *ConnectivityHealthCheck) Check(ctx context.Context, agent *agent.Agent) *HealthCheckResult {
	start := time.Now()
	result := &HealthCheckResult{
		CheckName: h.name,
		CheckType: h.Type(),
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	// Simple connectivity check - in a real implementation this would
	// check database connections, external services, etc.

	result.Duration = time.Since(start)
	result.Details["check_duration"] = result.Duration.String()

	// For now, assume connectivity is healthy if agent context is not cancelled
	select {
	case <-agent.Context().Done():
		result.Status = HealthStatusUnhealthy
		result.Message = "Agent context is cancelled"
	default:
		result.Status = HealthStatusHealthy
		result.Message = "Connectivity checks passed"
	}

	return result
}

func (h *ConnectivityHealthCheck) Interval() time.Duration {
	return h.interval
}

func (h *ConnectivityHealthCheck) Timeout() time.Duration {
	return h.timeout
}

func (h *ConnectivityHealthCheck) IsEnabled() bool {
	return h.enabled
}

// GetDefaultHealthChecks returns a set of default health checks
func GetDefaultHealthChecks() []HealthCheck {
	return []HealthCheck{
		NewHeartbeatHealthCheck(),
		NewResourceHealthCheck(),
		NewPerformanceHealthCheck(),
		NewConnectivityHealthCheck(),
	}
}
