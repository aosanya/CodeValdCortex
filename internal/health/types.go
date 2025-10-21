package health

import (
	"context"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
)

// HealthStatus represents the overall health state of an agent or system component
type HealthStatus string

const (
	// HealthStatusHealthy indicates normal operation
	HealthStatusHealthy HealthStatus = "healthy"
	// HealthStatusDegraded indicates reduced performance but functional
	HealthStatusDegraded HealthStatus = "degraded"
	// HealthStatusUnhealthy indicates significant issues affecting operation
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	// HealthStatusCritical indicates severe issues requiring immediate attention
	HealthStatusCritical HealthStatus = "critical"
	// HealthStatusUnknown indicates health status cannot be determined
	HealthStatusUnknown HealthStatus = "unknown"
)

// CheckType defines the category of health check
type CheckType string

const (
	// CheckTypeHeartbeat monitors agent responsiveness
	CheckTypeHeartbeat CheckType = "heartbeat"
	// CheckTypeResource monitors resource usage (CPU, memory)
	CheckTypeResource CheckType = "resource"
	// CheckTypeConnectivity checks network and database connections
	CheckTypeConnectivity CheckType = "connectivity"
	// CheckTypePerformance monitors task execution performance
	CheckTypePerformance CheckType = "performance"
	// CheckTypeCustom allows for application-specific checks
	CheckTypeCustom CheckType = "custom"
)

// HealthCheckResult represents the outcome of a single health check
type HealthCheckResult struct {
	// CheckName identifies the specific check
	CheckName string

	// CheckType categorizes the health check
	CheckType CheckType

	// Status indicates the health state
	Status HealthStatus

	// Message provides human-readable details
	Message string

	// Details contains additional structured information
	Details map[string]interface{}

	// Duration is how long the check took to execute
	Duration time.Duration

	// Timestamp when the check was performed
	Timestamp time.Time

	// Error contains any error that occurred during the check
	Error error
}

// IsHealthy returns true if the check result indicates healthy status
func (r *HealthCheckResult) IsHealthy() bool {
	return r.Status == HealthStatusHealthy
}

// AgentHealthReport contains comprehensive health information for an agent
type AgentHealthReport struct {
	// AgentID identifies the agent
	AgentID string

	// OverallStatus is the aggregated health status
	OverallStatus HealthStatus

	// CheckResults contains individual check results
	CheckResults []*HealthCheckResult

	// Metrics contains current performance metrics
	Metrics *AgentMetrics

	// Timestamp when the report was generated
	Timestamp time.Time

	// LastHealthyTime indicates when the agent was last healthy
	LastHealthyTime time.Time

	// ConsecutiveFailures tracks the number of consecutive failed checks
	ConsecutiveFailures int
}

// IsHealthy returns true if the overall agent health is good
func (r *AgentHealthReport) IsHealthy() bool {
	return r.OverallStatus == HealthStatusHealthy || r.OverallStatus == HealthStatusDegraded
}

// GetFailedChecks returns all health checks that failed
func (r *AgentHealthReport) GetFailedChecks() []*HealthCheckResult {
	var failed []*HealthCheckResult
	for _, check := range r.CheckResults {
		if !check.IsHealthy() {
			failed = append(failed, check)
		}
	}
	return failed
}

// AgentMetrics contains performance and resource metrics for an agent
type AgentMetrics struct {
	// CPU usage percentage (0-100)
	CPUUsage float64

	// Memory usage in bytes
	MemoryUsage int64

	// Memory usage percentage (0-100)
	MemoryPercent float64

	// Number of active goroutines
	Goroutines int

	// Task execution metrics
	TasksCompleted int64
	TasksFailed    int64
	TasksInQueue   int64

	// Response time metrics (in milliseconds)
	AvgResponseTime float64
	P95ResponseTime float64
	P99ResponseTime float64

	// Network metrics
	ConnectionsActive int
	ConnectionsFailed int64

	// Uptime in seconds
	UptimeSeconds int64

	// Last heartbeat timestamp
	LastHeartbeat time.Time
}

// HealthCheck defines the interface for health check implementations
type HealthCheck interface {
	// Name returns the name of this health check
	Name() string

	// Type returns the type of this health check
	Type() CheckType

	// Check performs the health check and returns the result
	Check(ctx context.Context, agent *agent.Agent) *HealthCheckResult

	// Interval returns how often this check should be performed
	Interval() time.Duration

	// Timeout returns the maximum time allowed for this check
	Timeout() time.Duration

	// IsEnabled returns whether this check is currently enabled
	IsEnabled() bool
}

// HealthMonitor defines the interface for the health monitoring system
type HealthMonitor interface {
	// StartMonitoring begins health monitoring for an agent
	StartMonitoring(ctx context.Context, agentID string) error

	// StopMonitoring stops health monitoring for an agent
	StopMonitoring(agentID string) error

	// GetHealthReport returns the current health report for an agent
	GetHealthReport(agentID string) (*AgentHealthReport, error)

	// GetAllHealthReports returns health reports for all monitored agents
	GetAllHealthReports() (map[string]*AgentHealthReport, error)

	// RegisterHealthCheck adds a new health check to the system
	RegisterHealthCheck(check HealthCheck) error

	// UnregisterHealthCheck removes a health check from the system
	UnregisterHealthCheck(checkName string) error

	// GetMetrics returns current system-wide health metrics
	GetMetrics() (*SystemHealthMetrics, error)
}

// SystemHealthMetrics contains aggregated health metrics across all agents
type SystemHealthMetrics struct {
	// Total number of monitored agents
	TotalAgents int

	// Number of healthy agents
	HealthyAgents int

	// Number of degraded agents
	DegradedAgents int

	// Number of unhealthy agents
	UnhealthyAgents int

	// Number of critical agents
	CriticalAgents int

	// Number of unknown status agents
	UnknownAgents int

	// Average response time across all agents
	AvgResponseTime float64

	// Total number of health checks performed
	TotalChecksPerformed int64

	// Total number of failed health checks
	TotalChecksFailed int64

	// Timestamp when metrics were collected
	Timestamp time.Time
}

// HealthEventType defines the types of health-related events
type HealthEventType string

const (
	// HealthEventAgentHealthy indicates an agent became healthy
	HealthEventAgentHealthy HealthEventType = "agent_healthy"
	// HealthEventAgentDegraded indicates an agent became degraded
	HealthEventAgentDegraded HealthEventType = "agent_degraded"
	// HealthEventAgentUnhealthy indicates an agent became unhealthy
	HealthEventAgentUnhealthy HealthEventType = "agent_unhealthy"
	// HealthEventAgentCritical indicates an agent became critical
	HealthEventAgentCritical HealthEventType = "agent_critical"
	// HealthEventCheckFailed indicates a specific health check failed
	HealthEventCheckFailed HealthEventType = "health_check_failed"
	// HealthEventRecovery indicates an agent recovered from unhealthy state
	HealthEventRecovery HealthEventType = "agent_recovery"
)

// HealthEvent represents a health-related event that can be published
type HealthEvent struct {
	// Type of health event
	Type HealthEventType

	// AgentID identifies the affected agent
	AgentID string

	// PreviousStatus is the previous health status
	PreviousStatus HealthStatus

	// CurrentStatus is the current health status
	CurrentStatus HealthStatus

	// CheckResult contains details if related to a specific check
	CheckResult *HealthCheckResult

	// Message provides human-readable description
	Message string

	// Metadata contains additional event context
	Metadata map[string]interface{}

	// Timestamp when the event occurred
	Timestamp time.Time
}

// FailureDetectionConfig configures failure detection behavior
type FailureDetectionConfig struct {
	// MaxConsecutiveFailures before marking agent as unhealthy
	MaxConsecutiveFailures int

	// GracePeriod before starting failure detection after agent startup
	GracePeriod time.Duration

	// RecoveryThreshold consecutive successful checks needed for recovery
	RecoveryThreshold int

	// EscalationThreshold consecutive failures before marking as critical
	EscalationThreshold int

	// AutoRecoveryEnabled allows automatic recovery attempts
	AutoRecoveryEnabled bool

	// RecoveryDelay between recovery attempts
	RecoveryDelay time.Duration
}

// RecoveryAction defines an action that can be taken to recover an unhealthy agent
type RecoveryAction interface {
	// Name returns the name of this recovery action
	Name() string

	// CanRecover returns true if this action can be applied to the given agent
	CanRecover(agent *agent.Agent, report *AgentHealthReport) bool

	// Recover attempts to recover the agent and returns success status
	Recover(ctx context.Context, agent *agent.Agent) error

	// Priority returns the priority of this recovery action (higher = more preferred)
	Priority() int
}
