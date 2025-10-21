package orchestration

import (
	"context"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
)

// WorkflowStatus represents the current state of workflow execution
type WorkflowStatus string

const (
	// WorkflowStatusPending indicates workflow is waiting to start
	WorkflowStatusPending WorkflowStatus = "pending"
	// WorkflowStatusRunning indicates workflow is currently executing
	WorkflowStatusRunning WorkflowStatus = "running"
	// WorkflowStatusCompleted indicates workflow finished successfully
	WorkflowStatusCompleted WorkflowStatus = "completed"
	// WorkflowStatusFailed indicates workflow execution failed
	WorkflowStatusFailed WorkflowStatus = "failed"
	// WorkflowStatusCancelled indicates workflow was cancelled
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
	// WorkflowStatusPaused indicates workflow execution is paused
	WorkflowStatusPaused WorkflowStatus = "paused"
)

// TaskStatus represents the current state of a workflow task
type TaskStatus string

const (
	// TaskStatusPending indicates task is waiting to be scheduled
	TaskStatusPending TaskStatus = "pending"
	// TaskStatusQueued indicates task is queued for execution
	TaskStatusQueued TaskStatus = "queued"
	// TaskStatusRunning indicates task is currently executing
	TaskStatusRunning TaskStatus = "running"
	// TaskStatusCompleted indicates task finished successfully
	TaskStatusCompleted TaskStatus = "completed"
	// TaskStatusFailed indicates task execution failed
	TaskStatusFailed TaskStatus = "failed"
	// TaskStatusSkipped indicates task was skipped due to conditions
	TaskStatusSkipped TaskStatus = "skipped"
	// TaskStatusRetrying indicates task is being retried after failure
	TaskStatusRetrying TaskStatus = "retrying"
)

// AgentSelectionStrategy defines how agents are selected for task execution
type AgentSelectionStrategy string

const (
	// AgentSelectionRoundRobin distributes tasks evenly across available agents
	AgentSelectionRoundRobin AgentSelectionStrategy = "round_robin"
	// AgentSelectionLeastLoaded assigns tasks to agents with lowest current load
	AgentSelectionLeastLoaded AgentSelectionStrategy = "least_loaded"
	// AgentSelectionSpecific assigns tasks to specific agents by ID
	AgentSelectionSpecific AgentSelectionStrategy = "specific"
	// AgentSelectionCapabilityBased selects agents based on required capabilities
	AgentSelectionCapabilityBased AgentSelectionStrategy = "capability_based"
	// AgentSelectionHealthAware prioritizes healthy agents for task assignment
	AgentSelectionHealthAware AgentSelectionStrategy = "health_aware"
)

// Workflow represents a multi-agent workflow definition
type Workflow struct {
	// ID is the unique workflow identifier
	ID string `json:"id"`

	// Name is the human-readable workflow name
	Name string `json:"name"`

	// Description provides workflow documentation
	Description string `json:"description"`

	// Version identifies the workflow version
	Version string `json:"version"`

	// Tasks defines the workflow tasks and their configuration
	Tasks []WorkflowTask `json:"tasks"`

	// Dependencies defines task execution dependencies
	Dependencies map[string][]string `json:"dependencies"`

	// Configuration holds workflow-wide settings
	Configuration WorkflowConfiguration `json:"configuration"`

	// Triggers define when the workflow should execute
	Triggers []WorkflowTrigger `json:"triggers"`

	// CreatedAt tracks workflow creation time
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt tracks last workflow modification
	UpdatedAt time.Time `json:"updated_at"`

	// CreatedBy identifies the workflow creator
	CreatedBy string `json:"created_by"`
}

// WorkflowTask represents a single task within a workflow
type WorkflowTask struct {
	// ID is the unique task identifier within the workflow
	ID string `json:"id"`

	// Name is the human-readable task name
	Name string `json:"name"`

	// Type specifies the task type (e.g., "http_request", "data_processing")
	Type string `json:"type"`

	// AgentSelector defines how to select agents for this task
	AgentSelector AgentSelector `json:"agent_selector"`

	// Parameters contains task-specific configuration
	Parameters map[string]interface{} `json:"parameters"`

	// Conditions define when this task should execute
	Conditions []TaskCondition `json:"conditions"`

	// RetryPolicy configures retry behavior for failed tasks
	RetryPolicy RetryPolicy `json:"retry_policy"`

	// Timeout specifies maximum task execution time
	Timeout time.Duration `json:"timeout"`

	// Priority affects task scheduling order (higher = more priority)
	Priority int `json:"priority"`

	// Resources specify required agent resources
	Resources TaskResources `json:"resources"`

	// OutputMapping defines how task outputs are captured
	OutputMapping map[string]string `json:"output_mapping"`
}

// AgentSelector defines criteria for selecting agents to execute tasks
type AgentSelector struct {
	// Strategy specifies the agent selection method
	Strategy AgentSelectionStrategy `json:"strategy"`

	// SpecificAgents lists specific agent IDs (for specific strategy)
	SpecificAgents []string `json:"specific_agents,omitempty"`

	// RequiredCapabilities lists required agent capabilities
	RequiredCapabilities []string `json:"required_capabilities,omitempty"`

	// PoolID specifies a specific agent pool
	PoolID string `json:"pool_id,omitempty"`

	// Tags define agent tags for selection
	Tags map[string]string `json:"tags,omitempty"`

	// HealthThreshold specifies minimum health score for selection
	HealthThreshold float64 `json:"health_threshold,omitempty"`
}

// TaskCondition defines when a task should execute
type TaskCondition struct {
	// Type specifies the condition type
	Type string `json:"type"`

	// Expression contains the condition logic
	Expression string `json:"expression"`

	// Parameters provide condition-specific data
	Parameters map[string]interface{} `json:"parameters"`
}

// RetryPolicy configures task retry behavior
type RetryPolicy struct {
	// MaxAttempts specifies maximum retry attempts
	MaxAttempts int `json:"max_attempts"`

	// InitialDelay is the delay before first retry
	InitialDelay time.Duration `json:"initial_delay"`

	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration `json:"max_delay"`

	// BackoffMultiplier for exponential backoff
	BackoffMultiplier float64 `json:"backoff_multiplier"`

	// RetryableErrors specifies which errors should trigger retries
	RetryableErrors []string `json:"retryable_errors"`
}

// TaskResources defines resource requirements for task execution
type TaskResources struct {
	// CPU in millicores (e.g., 100 = 0.1 CPU)
	CPU int `json:"cpu"`

	// Memory in megabytes
	Memory int `json:"memory"`

	// EstimatedDuration for resource planning
	EstimatedDuration time.Duration `json:"estimated_duration"`
}

// WorkflowConfiguration holds workflow-wide settings
type WorkflowConfiguration struct {
	// MaxConcurrentTasks limits parallel task execution
	MaxConcurrentTasks int `json:"max_concurrent_tasks"`

	// DefaultTimeout for tasks without explicit timeout
	DefaultTimeout time.Duration `json:"default_timeout"`

	// FailurePolicy defines workflow behavior on task failures
	FailurePolicy FailurePolicy `json:"failure_policy"`

	// NotificationSettings configure workflow notifications
	NotificationSettings NotificationSettings `json:"notification_settings"`

	// Variables define workflow-level variables
	Variables map[string]interface{} `json:"variables"`
}

// FailurePolicy defines workflow failure handling
type FailurePolicy struct {
	// OnTaskFailure specifies action when a task fails
	OnTaskFailure string `json:"on_task_failure"` // "stop", "continue", "retry_workflow"

	// MaxFailedTasks before stopping workflow
	MaxFailedTasks int `json:"max_failed_tasks"`

	// CriticalTasks that must succeed for workflow to continue
	CriticalTasks []string `json:"critical_tasks"`
}

// NotificationSettings configure workflow notifications
type NotificationSettings struct {
	// OnCompletion sends notification when workflow completes
	OnCompletion bool `json:"on_completion"`

	// OnFailure sends notification when workflow fails
	OnFailure bool `json:"on_failure"`

	// Recipients list notification recipients
	Recipients []string `json:"recipients"`

	// Channels specify notification channels (email, slack, etc.)
	Channels []string `json:"channels"`
}

// WorkflowTrigger defines when a workflow should execute
type WorkflowTrigger struct {
	// Type specifies the trigger type
	Type string `json:"type"` // "schedule", "event", "manual", "api"

	// Configuration holds trigger-specific settings
	Configuration map[string]interface{} `json:"configuration"`

	// Enabled indicates if trigger is active
	Enabled bool `json:"enabled"`
}

// WorkflowExecution represents an instance of workflow execution
type WorkflowExecution struct {
	// ID is the unique execution identifier
	ID string `json:"id"`

	// WorkflowID references the workflow definition
	WorkflowID string `json:"workflow_id"`

	// Status indicates current execution state
	Status WorkflowStatus `json:"status"`

	// StartTime tracks when execution began
	StartTime time.Time `json:"start_time"`

	// EndTime tracks when execution completed
	EndTime *time.Time `json:"end_time,omitempty"`

	// Duration is the total execution time
	Duration time.Duration `json:"duration"`

	// TaskExecutions tracks individual task executions
	TaskExecutions map[string]*TaskExecution `json:"task_executions"`

	// Context holds execution-wide variables and state
	Context map[string]interface{} `json:"context"`

	// Error contains execution error details
	Error string `json:"error,omitempty"`

	// Metrics captures execution performance data
	Metrics ExecutionMetrics `json:"metrics"`

	// TriggeredBy identifies what triggered this execution
	TriggeredBy string `json:"triggered_by"`

	// AgentsUsed tracks which agents participated in execution
	AgentsUsed []string `json:"agents_used"`
}

// TaskExecution represents the execution of a single workflow task
type TaskExecution struct {
	// TaskID references the workflow task
	TaskID string `json:"task_id"`

	// AgentID identifies the executing agent
	AgentID string `json:"agent_id"`

	// Status indicates current task state
	Status TaskStatus `json:"status"`

	// StartTime tracks when task execution began
	StartTime time.Time `json:"start_time"`

	// EndTime tracks when task execution completed
	EndTime *time.Time `json:"end_time,omitempty"`

	// Duration is the task execution time
	Duration time.Duration `json:"duration"`

	// Attempts tracks retry attempts
	Attempts int `json:"attempts"`

	// Output contains task execution results
	Output map[string]interface{} `json:"output"`

	// Error contains task error details
	Error string `json:"error,omitempty"`

	// Logs capture task execution logs
	Logs []string `json:"logs"`

	// ResourceUsage tracks actual resource consumption
	ResourceUsage ResourceUsage `json:"resource_usage"`
}

// ExecutionMetrics captures workflow execution performance data
type ExecutionMetrics struct {
	// TotalTasks in the workflow
	TotalTasks int `json:"total_tasks"`

	// CompletedTasks count
	CompletedTasks int `json:"completed_tasks"`

	// FailedTasks count
	FailedTasks int `json:"failed_tasks"`

	// SkippedTasks count
	SkippedTasks int `json:"skipped_tasks"`

	// AverageTaskDuration across all tasks
	AverageTaskDuration time.Duration `json:"average_task_duration"`

	// MaxConcurrentTasks achieved during execution
	MaxConcurrentTasks int `json:"max_concurrent_tasks"`

	// AgentsUtilized count of unique agents used
	AgentsUtilized int `json:"agents_utilized"`

	// TotalResourceUsage aggregated across all tasks
	TotalResourceUsage ResourceUsage `json:"total_resource_usage"`
}

// ResourceUsage tracks actual resource consumption
type ResourceUsage struct {
	// CPU usage in millicores
	CPU int `json:"cpu"`

	// Memory usage in megabytes
	Memory int `json:"memory"`

	// NetworkIO in bytes
	NetworkIO int64 `json:"network_io"`

	// DiskIO in bytes
	DiskIO int64 `json:"disk_io"`
}

// WorkflowEngine defines the interface for workflow orchestration
type WorkflowEngine interface {
	// ExecuteWorkflow starts execution of a workflow
	ExecuteWorkflow(ctx context.Context, workflow *Workflow) (*WorkflowExecution, error)

	// GetExecution retrieves a workflow execution by ID
	GetExecution(ctx context.Context, executionID string) (*WorkflowExecution, error)

	// ListExecutions returns workflow executions with filtering
	ListExecutions(ctx context.Context, filters ExecutionFilters) ([]*WorkflowExecution, error)

	// CancelExecution stops a running workflow execution
	CancelExecution(ctx context.Context, executionID string) error

	// PauseExecution temporarily stops workflow execution
	PauseExecution(ctx context.Context, executionID string) error

	// ResumeExecution continues a paused workflow execution
	ResumeExecution(ctx context.Context, executionID string) error

	// RetryExecution restarts a failed workflow execution
	RetryExecution(ctx context.Context, executionID string) (*WorkflowExecution, error)
}

// AgentCoordinator defines the interface for agent coordination
type AgentCoordinator interface {
	// SelectAgents chooses agents for task execution based on criteria
	SelectAgents(ctx context.Context, selector AgentSelector, count int) ([]*agent.Agent, error)

	// AssignTask assigns a task to a specific agent
	AssignTask(ctx context.Context, agentID string, task *WorkflowTask, execution *WorkflowExecution) error

	// GetAgentLoad returns current load information for an agent
	GetAgentLoad(ctx context.Context, agentID string) (*AgentLoad, error)

	// GetAvailableAgents returns agents available for task assignment
	GetAvailableAgents(ctx context.Context) ([]*agent.Agent, error)

	// RebalanceLoad redistributes tasks across agents
	RebalanceLoad(ctx context.Context) error
}

// ExecutionMonitor defines the interface for monitoring workflow executions
type ExecutionMonitor interface {
	// StartMonitoring begins monitoring a workflow execution
	StartMonitoring(ctx context.Context, execution *WorkflowExecution) error

	// StopMonitoring stops monitoring a workflow execution
	StopMonitoring(ctx context.Context, executionID string) error

	// GetMetrics returns execution metrics
	GetMetrics(ctx context.Context, executionID string) (*ExecutionMetrics, error)

	// GetProgress returns execution progress information
	GetProgress(ctx context.Context, executionID string) (*ExecutionProgress, error)

	// WatchExecution provides real-time execution updates
	WatchExecution(ctx context.Context, executionID string) (<-chan *ExecutionEvent, error)
}

// WorkflowRepository defines the interface for workflow persistence
type WorkflowRepository interface {
	// StoreWorkflow saves a workflow definition
	StoreWorkflow(ctx context.Context, workflow *Workflow) error

	// GetWorkflow retrieves a workflow by ID
	GetWorkflow(ctx context.Context, workflowID string) (*Workflow, error)

	// ListWorkflows returns workflows with filtering
	ListWorkflows(ctx context.Context, filters WorkflowFilters) ([]*Workflow, error)

	// UpdateWorkflow modifies an existing workflow
	UpdateWorkflow(ctx context.Context, workflow *Workflow) error

	// DeleteWorkflow removes a workflow
	DeleteWorkflow(ctx context.Context, workflowID string) error

	// StoreExecution saves workflow execution data
	StoreExecution(ctx context.Context, execution *WorkflowExecution) error

	// GetExecution retrieves workflow execution data
	GetExecution(ctx context.Context, executionID string) (*WorkflowExecution, error)

	// UpdateExecution modifies execution data
	UpdateExecution(ctx context.Context, execution *WorkflowExecution) error
}

// ExecutionFilters defines filters for querying workflow executions
type ExecutionFilters struct {
	// WorkflowID filters by workflow
	WorkflowID string

	// Status filters by execution status
	Status []WorkflowStatus

	// StartTimeAfter filters executions started after time
	StartTimeAfter *time.Time

	// StartTimeBefore filters executions started before time
	StartTimeBefore *time.Time

	// TriggeredBy filters by trigger source
	TriggeredBy string

	// Limit limits number of results
	Limit int

	// Offset for pagination
	Offset int
}

// WorkflowFilters defines filters for querying workflows
type WorkflowFilters struct {
	// Name filters by workflow name pattern
	Name string

	// CreatedBy filters by creator
	CreatedBy string

	// Tags filters by workflow tags
	Tags map[string]string

	// Limit limits number of results
	Limit int

	// Offset for pagination
	Offset int
}

// AgentLoad represents current agent load information
type AgentLoad struct {
	// AgentID identifies the agent
	AgentID string

	// ActiveTasks count of currently executing tasks
	ActiveTasks int

	// QueuedTasks count of tasks waiting for execution
	QueuedTasks int

	// CPUUsage current CPU utilization percentage
	CPUUsage float64

	// MemoryUsage current memory utilization percentage
	MemoryUsage float64

	// HealthScore from health monitoring system
	HealthScore float64

	// Capabilities available on this agent
	Capabilities []string

	// LastUpdated timestamp of load information
	LastUpdated time.Time
}

// ExecutionProgress represents workflow execution progress
type ExecutionProgress struct {
	// ExecutionID identifies the execution
	ExecutionID string

	// OverallProgress percentage (0-100)
	OverallProgress float64

	// CompletedTasks count
	CompletedTasks int

	// TotalTasks count
	TotalTasks int

	// CurrentPhase describes current execution phase
	CurrentPhase string

	// EstimatedTimeRemaining until completion
	EstimatedTimeRemaining time.Duration

	// TaskProgress tracks individual task progress
	TaskProgress map[string]float64
}

// ExecutionEvent represents real-time execution events
type ExecutionEvent struct {
	// ExecutionID identifies the execution
	ExecutionID string

	// Type specifies the event type
	Type string

	// TaskID for task-specific events
	TaskID string

	// AgentID for agent-specific events
	AgentID string

	// Data contains event-specific information
	Data map[string]interface{}

	// Timestamp when event occurred
	Timestamp time.Time
}

// OrchestrationConfig configures the orchestration system
type OrchestrationConfig struct {
	// MaxConcurrentWorkflows limits parallel workflow executions
	MaxConcurrentWorkflows int

	// DefaultWorkflowTimeout for workflows without explicit timeout
	DefaultWorkflowTimeout time.Duration

	// TaskDistributionStrategy for assigning tasks to agents
	TaskDistributionStrategy AgentSelectionStrategy

	// HealthCheckInterval for monitoring agent health
	HealthCheckInterval time.Duration

	// MetricsCollection configuration
	MetricsCollection MetricsConfig

	// EventPublishing configuration
	EventPublishing EventConfig
}

// MetricsConfig configures metrics collection
type MetricsConfig struct {
	// Enabled indicates if metrics collection is active
	Enabled bool

	// RetentionPeriod for metric data
	RetentionPeriod time.Duration

	// CollectionInterval for gathering metrics
	CollectionInterval time.Duration
}

// EventConfig configures event publishing
type EventConfig struct {
	// Enabled indicates if event publishing is active
	Enabled bool

	// PublishToEventSystem sends events to the event processing system
	PublishToEventSystem bool

	// PublishToPubSub sends events to pub/sub system
	PublishToPubSub bool

	// EventTopics specifies which topics to publish to
	EventTopics []string
}
