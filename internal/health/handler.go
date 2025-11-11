package health

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// HealthHandler provides HTTP endpoints for health monitoring
type HealthHandler struct {
	monitor *Monitor
	logger  *log.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(monitor *Monitor, logger *log.Logger) *HealthHandler {
	if logger == nil {
		logger = log.New()
	}

	return &HealthHandler{
		monitor: monitor,
		logger:  logger,
	}
}

// RegisterRoutes registers health monitoring routes with the HTTP mux
func (h *HealthHandler) RegisterRoutes(mux *http.ServeMux) {
	// Agent health endpoints
	mux.HandleFunc("/api/v1/health/agents", h.GetAllAgentHealth)
	mux.HandleFunc("/api/v1/health/agents/", h.handleAgentRoutes)

	// System health endpoints
	mux.HandleFunc("/api/v1/health/system/metrics", h.GetSystemMetrics)
	mux.HandleFunc("/api/v1/health/system/status", h.GetSystemStatus)

	// Health check management endpoints
	mux.HandleFunc("/api/v1/health/checks", h.GetHealthChecks)
	mux.HandleFunc("/api/v1/health/checks/", h.handleCheckRoutes)

	// Monitoring configuration endpoints
	mux.HandleFunc("/api/v1/health/config", h.handleConfigRoutes)
}

// handleAgentRoutes handles agent-specific routes
func (h *HealthHandler) handleAgentRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Extract agent ID from path
	if len(path) < len("/api/v1/health/agents/") {
		http.Error(w, "Agent ID is required", http.StatusBadRequest)
		return
	}

	agentPath := path[len("/api/v1/health/agents/"):]

	// Parse route
	if agentPath == "" {
		http.Error(w, "Agent ID is required", http.StatusBadRequest)
		return
	}

	// Find agent ID and action
	var agentID, action string
	if idx := len(agentPath); idx > 0 {
		// Simple parsing - in production would use a proper router
		if agentPath[idx-6:] == "/start" {
			agentID = agentPath[:idx-6]
			action = "start"
		} else if agentPath[idx-5:] == "/stop" {
			agentID = agentPath[:idx-5]
			action = "stop"
		} else {
			agentID = agentPath
			action = ""
		}
	}

	switch action {
	case "start":
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.startMonitoringForAgent(w, r, agentID)
	case "stop":
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.stopMonitoringForAgent(w, r, agentID)
	default:
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.getAgentHealthByID(w, r, agentID)
	}
}

// handleCheckRoutes handles health check management routes
func (h *HealthHandler) handleCheckRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if len(path) < len("/api/v1/health/checks/") {
		http.Error(w, "Check name is required", http.StatusBadRequest)
		return
	}

	checkPath := path[len("/api/v1/health/checks/"):]

	// Simple parsing for enable/disable actions
	var checkName, action string
	if idx := len(checkPath); idx > 0 {
		if checkPath[idx-7:] == "/enable" {
			checkName = checkPath[:idx-7]
			action = "enable"
		} else if checkPath[idx-8:] == "/disable" {
			checkName = checkPath[:idx-8]
			action = "disable"
		} else {
			checkName = checkPath
			action = ""
		}
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	switch action {
	case "enable":
		h.enableHealthCheckByName(w, r, checkName)
	case "disable":
		h.disableHealthCheckByName(w, r, checkName)
	default:
		http.Error(w, "Invalid action", http.StatusBadRequest)
	}
}

// handleConfigRoutes handles configuration routes
func (h *HealthHandler) handleConfigRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.GetMonitoringConfig(w, r)
	case "PUT":
		h.UpdateMonitoringConfig(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetAllAgentHealth returns health reports for all monitored agents
func (h *HealthHandler) GetAllAgentHealth(w http.ResponseWriter, r *http.Request) {
	reports, err := h.monitor.GetAllHealthReports()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get all health reports")
		http.Error(w, "Failed to get health reports", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	response := make(map[string]interface{})
	response["agents"] = reports
	response["total_agents"] = len(reports)
	response["timestamp"] = time.Now()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.WithError(err).Error("Failed to encode health reports response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// getAgentHealthByID returns the health status of a specific agent
func (h *HealthHandler) getAgentHealthByID(w http.ResponseWriter, _ *http.Request, agentID string) {
	h.monitor.mu.RLock()
	defer h.monitor.mu.RUnlock()

	report, exists := h.monitor.reports[agentID]
	if !exists {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GetAgentHealth returns the health status of a specific agent (original method for compatibility)
func (h *HealthHandler) GetAgentHealth(w http.ResponseWriter, r *http.Request) {
	// This would need proper routing in production
	http.Error(w, "Use /api/v1/health/agents/{agentId} endpoint", http.StatusBadRequest)
}

// startMonitoringForAgent starts health monitoring for a specific agent
func (h *HealthHandler) startMonitoringForAgent(w http.ResponseWriter, r *http.Request, agentID string) {
	err := h.monitor.StartMonitoring(r.Context(), agentID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start monitoring: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Monitoring started successfully",
		"agent_id": agentID,
	})
}

// StartMonitoring starts health monitoring for a specific agent (original method for compatibility)
func (h *HealthHandler) StartMonitoring(w http.ResponseWriter, r *http.Request) {
	// This would need proper routing in production
	http.Error(w, "Use /api/v1/health/agents/{agentId}/start endpoint", http.StatusBadRequest)
}

// stopMonitoringForAgent stops health monitoring for a specific agent
func (h *HealthHandler) stopMonitoringForAgent(w http.ResponseWriter, _ *http.Request, agentID string) {
	if agentID == "" {
		http.Error(w, "Agent ID is required", http.StatusBadRequest)
		return
	}

	if err := h.monitor.StopMonitoring(agentID); err != nil {
		h.logger.WithError(err).WithField("agent_id", agentID).Error("Failed to stop monitoring")
		http.Error(w, fmt.Sprintf("Failed to stop monitoring: %v", err), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"message":   "Monitoring stopped successfully",
		"agent_id":  agentID,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// StopMonitoring stops health monitoring for an agent (original method for compatibility)
func (h *HealthHandler) StopMonitoring(w http.ResponseWriter, r *http.Request) {
	// This would need proper routing in production
	http.Error(w, "Use /api/v1/health/agents/{agentId}/stop endpoint", http.StatusBadRequest)
}

// GetSystemMetrics returns system-wide health metrics
func (h *HealthHandler) GetSystemMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.monitor.GetMetrics()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get system metrics")
		http.Error(w, "Failed to get system metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		h.logger.WithError(err).Error("Failed to encode system metrics")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// GetSystemStatus returns overall system health status
func (h *HealthHandler) GetSystemStatus(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.monitor.GetMetrics()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get system metrics")
		http.Error(w, "Failed to get system status", http.StatusInternalServerError)
		return
	}

	// Determine overall system status
	overallStatus := h.calculateOverallSystemStatus(metrics)

	response := map[string]interface{}{
		"status":             overallStatus,
		"total_agents":       metrics.TotalAgents,
		"healthy_agents":     metrics.HealthyAgents,
		"degraded_agents":    metrics.DegradedAgents,
		"unhealthy_agents":   metrics.UnhealthyAgents,
		"critical_agents":    metrics.CriticalAgents,
		"unknown_agents":     metrics.UnknownAgents,
		"avg_response_time":  metrics.AvgResponseTime,
		"check_success_rate": h.calculateCheckSuccessRate(metrics),
		"timestamp":          time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.WithError(err).Error("Failed to encode system status")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// GetHealthChecks returns information about available health checks
func (h *HealthHandler) GetHealthChecks(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would query the monitor for registered checks
	// For now, return information about default checks
	checks := []map[string]interface{}{
		{
			"name":        "heartbeat",
			"type":        "heartbeat",
			"enabled":     true,
			"interval":    "30s",
			"timeout":     "5s",
			"description": "Monitors agent responsiveness via heartbeat",
		},
		{
			"name":        "resource",
			"type":        "resource",
			"enabled":     true,
			"interval":    "1m",
			"timeout":     "10s",
			"description": "Monitors system resource usage",
		},
		{
			"name":        "performance",
			"type":        "performance",
			"enabled":     true,
			"interval":    "2m",
			"timeout":     "15s",
			"description": "Monitors agent task execution performance",
		},
		{
			"name":        "connectivity",
			"type":        "connectivity",
			"enabled":     true,
			"interval":    "1m",
			"timeout":     "10s",
			"description": "Monitors database and network connectivity",
		},
	}

	response := map[string]interface{}{
		"health_checks": checks,
		"total_checks":  len(checks),
		"timestamp":     time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.WithError(err).Error("Failed to encode health checks")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// enableHealthCheckByName enables a specific health check
func (h *HealthHandler) enableHealthCheckByName(w http.ResponseWriter, _ *http.Request, checkName string) {
	if checkName == "" {
		http.Error(w, "Check name is required", http.StatusBadRequest)
		return
	}

	// In a real implementation, this would enable the check in the monitor
	response := map[string]interface{}{
		"message":    fmt.Sprintf("Health check '%s' enabled successfully", checkName),
		"check_name": checkName,
		"enabled":    true,
		"timestamp":  time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// EnableHealthCheck enables a specific health check (original method for compatibility)
func (h *HealthHandler) EnableHealthCheck(w http.ResponseWriter, r *http.Request) {
	// This would need proper routing in production
	http.Error(w, "Use /api/v1/health/checks/{checkName}/enable endpoint", http.StatusBadRequest)
}

// disableHealthCheckByName disables a specific health check
func (h *HealthHandler) disableHealthCheckByName(w http.ResponseWriter, _ *http.Request, checkName string) {
	if checkName == "" {
		http.Error(w, "Check name is required", http.StatusBadRequest)
		return
	}

	// In a real implementation, this would disable the check in the monitor
	response := map[string]interface{}{
		"message":    fmt.Sprintf("Health check '%s' disabled successfully", checkName),
		"check_name": checkName,
		"enabled":    false,
		"timestamp":  time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DisableHealthCheck disables a specific health check (original method for compatibility)
func (h *HealthHandler) DisableHealthCheck(w http.ResponseWriter, r *http.Request) {
	// This would need proper routing in production
	http.Error(w, "Use /api/v1/health/checks/{checkName}/disable endpoint", http.StatusBadRequest)
}

// GetMonitoringConfig returns current monitoring configuration
func (h *HealthHandler) GetMonitoringConfig(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would return the actual configuration
	config := map[string]interface{}{
		"check_interval":           "1m",
		"max_consecutive_failures": 3,
		"grace_period":             "2m",
		"recovery_threshold":       2,
		"escalation_threshold":     5,
		"auto_recovery_enabled":    true,
		"recovery_delay":           "30s",
		"enable_events":            true,
		"max_reports":              1000,
		"timestamp":                time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(config); err != nil {
		h.logger.WithError(err).Error("Failed to encode monitoring config")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// UpdateMonitoringConfig updates monitoring configuration
func (h *HealthHandler) UpdateMonitoringConfig(w http.ResponseWriter, r *http.Request) {
	var config map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid JSON in request body", http.StatusBadRequest)
		return
	}

	// In a real implementation, this would update the actual configuration
	// For now, just return success
	response := map[string]interface{}{
		"message":   "Configuration updated successfully",
		"config":    config,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// calculateOverallSystemStatus determines the overall system health status
func (h *HealthHandler) calculateOverallSystemStatus(metrics *SystemHealthMetrics) HealthStatus {
	if metrics.TotalAgents == 0 {
		return HealthStatusUnknown
	}

	criticalPercent := float64(metrics.CriticalAgents) / float64(metrics.TotalAgents)
	unhealthyPercent := float64(metrics.UnhealthyAgents) / float64(metrics.TotalAgents)
	degradedPercent := float64(metrics.DegradedAgents) / float64(metrics.TotalAgents)

	// Determine status based on percentages
	if criticalPercent > 0.1 { // More than 10% critical
		return HealthStatusCritical
	}
	if unhealthyPercent > 0.2 { // More than 20% unhealthy
		return HealthStatusUnhealthy
	}
	if degradedPercent > 0.3 { // More than 30% degraded
		return HealthStatusDegraded
	}

	return HealthStatusHealthy
}

// calculateCheckSuccessRate calculates the success rate of health checks
func (h *HealthHandler) calculateCheckSuccessRate(metrics *SystemHealthMetrics) float64 {
	if metrics.TotalChecksPerformed == 0 {
		return 0.0
	}

	successfulChecks := metrics.TotalChecksPerformed - metrics.TotalChecksFailed
	return float64(successfulChecks) / float64(metrics.TotalChecksPerformed) * 100
}
