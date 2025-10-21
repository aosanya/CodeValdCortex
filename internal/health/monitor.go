package health

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	log "github.com/sirupsen/logrus"
)

// HealthMonitorConfig configures the health monitoring system
type HealthMonitorConfig struct {
	// CheckInterval defines how often to run health checks
	CheckInterval time.Duration

	// FailureDetection configures failure detection behavior
	FailureDetection FailureDetectionConfig

	// EnableEvents controls whether to publish health events
	EnableEvents bool

	// MaxReports limits the number of health reports to keep in memory
	MaxReports int
}

// DefaultHealthMonitorConfig returns default configuration
func DefaultHealthMonitorConfig() HealthMonitorConfig {
	return HealthMonitorConfig{
		CheckInterval: 1 * time.Minute,
		FailureDetection: FailureDetectionConfig{
			MaxConsecutiveFailures: 3,
			GracePeriod:            2 * time.Minute,
			RecoveryThreshold:      2,
			EscalationThreshold:    5,
			AutoRecoveryEnabled:    true,
			RecoveryDelay:          30 * time.Second,
		},
		EnableEvents: true,
		MaxReports:   1000,
	}
}

// Monitor implements the HealthMonitor interface
type Monitor struct {
	config         HealthMonitorConfig
	healthChecks   map[string]HealthCheck
	agents         map[string]*agent.Agent
	reports        map[string]*AgentHealthReport
	monitoring     map[string]context.CancelFunc
	eventPublisher HealthEventPublisher
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	logger         *log.Logger
}

// HealthEventPublisher defines the interface for publishing health events
type HealthEventPublisher interface {
	PublishHealthEvent(ctx context.Context, event *HealthEvent) error
}

// NewMonitor creates a new health monitor
func NewMonitor(config HealthMonitorConfig, eventPublisher HealthEventPublisher, logger *log.Logger) *Monitor {
	ctx, cancel := context.WithCancel(context.Background())

	if logger == nil {
		logger = log.New()
	}

	m := &Monitor{
		config:         config,
		healthChecks:   make(map[string]HealthCheck),
		agents:         make(map[string]*agent.Agent),
		reports:        make(map[string]*AgentHealthReport),
		monitoring:     make(map[string]context.CancelFunc),
		eventPublisher: eventPublisher,
		ctx:            ctx,
		cancel:         cancel,
		logger:         logger,
	}

	// Register default health checks
	for _, check := range GetDefaultHealthChecks() {
		m.RegisterHealthCheck(check)
	}

	return m
}

// StartMonitoring begins health monitoring for an agent
func (m *Monitor) StartMonitoring(ctx context.Context, agentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already monitoring
	if _, exists := m.monitoring[agentID]; exists {
		return fmt.Errorf("already monitoring agent %s", agentID)
	}

	// Get agent instance (in real implementation, this would retrieve from agent manager)
	agentInstance, exists := m.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	// Create monitoring context
	monitorCtx, cancel := context.WithCancel(m.ctx)
	m.monitoring[agentID] = cancel

	// Initialize health report
	m.reports[agentID] = &AgentHealthReport{
		AgentID:             agentID,
		OverallStatus:       HealthStatusUnknown,
		CheckResults:        []*HealthCheckResult{},
		Timestamp:           time.Now(),
		LastHealthyTime:     time.Now(),
		ConsecutiveFailures: 0,
	}

	// Start monitoring goroutine
	m.wg.Add(1)
	go m.monitorAgent(monitorCtx, agentInstance)

	m.logger.WithField("agent_id", agentID).Info("Started health monitoring")
	return nil
}

// StopMonitoring stops health monitoring for an agent
func (m *Monitor) StopMonitoring(agentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cancel, exists := m.monitoring[agentID]
	if !exists {
		return fmt.Errorf("not monitoring agent %s", agentID)
	}

	// Stop monitoring
	cancel()
	delete(m.monitoring, agentID)
	delete(m.agents, agentID)

	m.logger.WithField("agent_id", agentID).Info("Stopped health monitoring")
	return nil
}

// RegisterAgent registers an agent for potential monitoring
func (m *Monitor) RegisterAgent(agentInstance *agent.Agent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.agents[agentInstance.ID] = agentInstance
}

// UnregisterAgent removes an agent from the monitor
func (m *Monitor) UnregisterAgent(agentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.agents, agentID)
	delete(m.reports, agentID)
}

// GetHealthReport returns the current health report for an agent
func (m *Monitor) GetHealthReport(agentID string) (*AgentHealthReport, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	report, exists := m.reports[agentID]
	if !exists {
		return nil, fmt.Errorf("no health report found for agent %s", agentID)
	}

	// Return a copy of the report
	reportCopy := *report
	return &reportCopy, nil
}

// GetAllHealthReports returns health reports for all monitored agents
func (m *Monitor) GetAllHealthReports() (map[string]*AgentHealthReport, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	reports := make(map[string]*AgentHealthReport)
	for agentID, report := range m.reports {
		reportCopy := *report
		reports[agentID] = &reportCopy
	}

	return reports, nil
}

// RegisterHealthCheck adds a new health check to the system
func (m *Monitor) RegisterHealthCheck(check HealthCheck) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if check == nil {
		return fmt.Errorf("health check cannot be nil")
	}

	name := check.Name()
	if _, exists := m.healthChecks[name]; exists {
		return fmt.Errorf("health check %s already registered", name)
	}

	m.healthChecks[name] = check
	m.logger.WithField("check_name", name).Info("Registered health check")
	return nil
}

// UnregisterHealthCheck removes a health check from the system
func (m *Monitor) UnregisterHealthCheck(checkName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.healthChecks[checkName]; !exists {
		return fmt.Errorf("health check %s not found", checkName)
	}

	delete(m.healthChecks, checkName)
	m.logger.WithField("check_name", checkName).Info("Unregistered health check")
	return nil
}

// GetMetrics returns current system-wide health metrics
func (m *Monitor) GetMetrics() (*SystemHealthMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := &SystemHealthMetrics{
		TotalAgents:          len(m.reports),
		HealthyAgents:        0,
		DegradedAgents:       0,
		UnhealthyAgents:      0,
		CriticalAgents:       0,
		UnknownAgents:        0,
		AvgResponseTime:      0,
		TotalChecksPerformed: 0,
		TotalChecksFailed:    0,
		Timestamp:            time.Now(),
	}

	// Calculate aggregate metrics
	var totalResponseTime float64
	var responseTimeCount int64

	for _, report := range m.reports {
		switch report.OverallStatus {
		case HealthStatusHealthy:
			metrics.HealthyAgents++
		case HealthStatusDegraded:
			metrics.DegradedAgents++
		case HealthStatusUnhealthy:
			metrics.UnhealthyAgents++
		case HealthStatusCritical:
			metrics.CriticalAgents++
		default:
			metrics.UnknownAgents++
		}

		// Aggregate check metrics
		for _, checkResult := range report.CheckResults {
			metrics.TotalChecksPerformed++
			if !checkResult.IsHealthy() {
				metrics.TotalChecksFailed++
			}

			// Add response time data
			totalResponseTime += float64(checkResult.Duration.Milliseconds())
			responseTimeCount++
		}
	}

	// Calculate average response time
	if responseTimeCount > 0 {
		metrics.AvgResponseTime = totalResponseTime / float64(responseTimeCount)
	}

	return metrics, nil
}

// Shutdown gracefully shuts down the health monitor
func (m *Monitor) Shutdown() error {
	m.logger.Info("Shutting down health monitor")

	// Stop all monitoring
	m.mu.Lock()
	for agentID, cancel := range m.monitoring {
		cancel()
		m.logger.WithField("agent_id", agentID).Debug("Stopped monitoring during shutdown")
	}
	m.monitoring = make(map[string]context.CancelFunc)
	m.mu.Unlock()

	// Cancel main context
	m.cancel()

	// Wait for all monitoring goroutines to finish
	m.wg.Wait()

	m.logger.Info("Health monitor shutdown complete")
	return nil
}

// monitorAgent runs health checks for a specific agent
func (m *Monitor) monitorAgent(ctx context.Context, agentInstance *agent.Agent) {
	defer m.wg.Done()

	agentID := agentInstance.ID
	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	m.logger.WithField("agent_id", agentID).Debug("Started agent monitoring loop")

	// Grace period before starting health checks
	select {
	case <-ctx.Done():
		return
	case <-time.After(m.config.FailureDetection.GracePeriod):
		// Continue with monitoring
	}

	for {
		select {
		case <-ctx.Done():
			m.logger.WithField("agent_id", agentID).Debug("Agent monitoring stopped")
			return
		case <-ticker.C:
			m.performHealthChecks(ctx, agentInstance)
		}
	}
}

// performHealthChecks executes all registered health checks for an agent
func (m *Monitor) performHealthChecks(ctx context.Context, agentInstance *agent.Agent) {
	agentID := agentInstance.ID

	m.mu.RLock()
	checks := make([]HealthCheck, 0, len(m.healthChecks))
	for _, check := range m.healthChecks {
		if check.IsEnabled() {
			checks = append(checks, check)
		}
	}
	m.mu.RUnlock()

	// Perform all health checks
	var results []*HealthCheckResult
	for _, check := range checks {
		// Create timeout context for individual check
		checkCtx, cancel := context.WithTimeout(ctx, check.Timeout())
		result := check.Check(checkCtx, agentInstance)
		cancel()

		results = append(results, result)
	}

	// Update health report
	m.updateHealthReport(agentID, results)
}

// updateHealthReport updates the health report for an agent and publishes events
func (m *Monitor) updateHealthReport(agentID string, results []*HealthCheckResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	report, exists := m.reports[agentID]
	if !exists {
		return
	}

	previousStatus := report.OverallStatus
	report.CheckResults = results
	report.Timestamp = time.Now()

	// Determine overall health status
	report.OverallStatus = m.calculateOverallStatus(results)

	// Update failure tracking
	if report.OverallStatus == HealthStatusHealthy || report.OverallStatus == HealthStatusDegraded {
		if report.ConsecutiveFailures > 0 {
			// Agent recovered
			if m.config.EnableEvents && m.eventPublisher != nil {
				event := &HealthEvent{
					Type:           HealthEventRecovery,
					AgentID:        agentID,
					PreviousStatus: previousStatus,
					CurrentStatus:  report.OverallStatus,
					Message:        fmt.Sprintf("Agent %s recovered from %s to %s", agentID, previousStatus, report.OverallStatus),
					Timestamp:      time.Now(),
				}
				m.eventPublisher.PublishHealthEvent(context.Background(), event)
			}
		}
		report.ConsecutiveFailures = 0
		report.LastHealthyTime = time.Now()
	} else {
		report.ConsecutiveFailures++
	}

	// Publish status change events
	if previousStatus != report.OverallStatus && m.config.EnableEvents && m.eventPublisher != nil {
		eventType := m.getHealthEventType(report.OverallStatus)
		event := &HealthEvent{
			Type:           eventType,
			AgentID:        agentID,
			PreviousStatus: previousStatus,
			CurrentStatus:  report.OverallStatus,
			Message:        fmt.Sprintf("Agent %s health changed from %s to %s", agentID, previousStatus, report.OverallStatus),
			Timestamp:      time.Now(),
		}
		m.eventPublisher.PublishHealthEvent(context.Background(), event)
	}
}

// calculateOverallStatus determines the overall health status from individual check results
func (m *Monitor) calculateOverallStatus(results []*HealthCheckResult) HealthStatus {
	if len(results) == 0 {
		return HealthStatusUnknown
	}

	criticalCount := 0
	unhealthyCount := 0
	degradedCount := 0
	healthyCount := 0

	for _, result := range results {
		switch result.Status {
		case HealthStatusCritical:
			criticalCount++
		case HealthStatusUnhealthy:
			unhealthyCount++
		case HealthStatusDegraded:
			degradedCount++
		case HealthStatusHealthy:
			healthyCount++
		}
	}

	// Determine overall status based on worst case
	if criticalCount > 0 {
		return HealthStatusCritical
	}
	if unhealthyCount > 0 {
		return HealthStatusUnhealthy
	}
	if degradedCount > 0 {
		return HealthStatusDegraded
	}
	if healthyCount > 0 {
		return HealthStatusHealthy
	}

	return HealthStatusUnknown
}

// getHealthEventType maps health status to event type
func (m *Monitor) getHealthEventType(status HealthStatus) HealthEventType {
	switch status {
	case HealthStatusHealthy:
		return HealthEventAgentHealthy
	case HealthStatusDegraded:
		return HealthEventAgentDegraded
	case HealthStatusUnhealthy:
		return HealthEventAgentUnhealthy
	case HealthStatusCritical:
		return HealthEventAgentCritical
	default:
		return HealthEventAgentUnhealthy
	}
}
