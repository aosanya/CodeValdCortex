package task

import (
	"context"
	"time"
)

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	// TaskStatusPending indicates task is waiting to be scheduled
	TaskStatusPending TaskStatus = "pending"
	// TaskStatusQueued indicates task is in the scheduler queue
	TaskStatusQueued TaskStatus = "queued"
	// TaskStatusRunning indicates task is currently executing
	TaskStatusRunning TaskStatus = "running"
	// TaskStatusCompleted indicates task finished successfully
	TaskStatusCompleted TaskStatus = "completed"
	// TaskStatusFailed indicates task execution failed
	TaskStatusFailed TaskStatus = "failed"
	// TaskStatusCancelled indicates task was cancelled before completion
	TaskStatusCancelled TaskStatus = "cancelled"
	// TaskStatusTimeout indicates task exceeded its timeout
	TaskStatusTimeout TaskStatus = "timeout"
)

// Task represents a unit of work to be executed by an agent
type Task struct {
	// ID is the unique task identifier
	ID string `json:"id"`

	// AgentID is the agent that will execute the task
	AgentID string `json:"agent_id"`

	// Type indicates the task type (determines which handler to use)
	Type string `json:"type"`

	// Name is a human-readable task name
	Name string `json:"name"`

	// Payload contains task-specific data
	Payload map[string]interface{} `json:"payload"`

	// Priority for task ordering (0-10, higher = more important)
	Priority int `json:"priority"`

	// Timeout for task execution (0 = no timeout)
	Timeout time.Duration `json:"timeout"`

	// Dependencies lists task IDs that must complete first
	Dependencies []string `json:"dependencies,omitempty"`

	// RetryPolicy defines retry behavior on failure
	RetryPolicy *RetryPolicy `json:"retry_policy,omitempty"`

	// Metadata contains additional task information
	Metadata map[string]string `json:"metadata,omitempty"`

	// Status is the current task state
	Status TaskStatus `json:"status"`

	// CreatedAt is when the task was created
	CreatedAt time.Time `json:"created_at"`

	// ScheduledAt is when the task was queued
	ScheduledAt time.Time `json:"scheduled_at,omitempty"`

	// StartedAt is when execution began
	StartedAt time.Time `json:"started_at,omitempty"`

	// CompletedAt is when execution finished
	CompletedAt time.Time `json:"completed_at,omitempty"`

	// Context for task execution (not serialized)
	ctx context.Context `json:"-"`

	// Cancel function for task cancellation (not serialized)
	cancel context.CancelFunc `json:"-"`
}

// TaskResult represents the outcome of task execution
type TaskResult struct {
	// TaskID is the ID of the executed task
	TaskID string `json:"task_id"`

	// AgentID is the agent that executed the task
	AgentID string `json:"agent_id"`

	// Status is the final task status
	Status TaskStatus `json:"status"`

	// Result contains the task output
	Result map[string]interface{} `json:"result,omitempty"`

	// Error contains error information if task failed
	Error string `json:"error,omitempty"`

	// StartedAt is when execution began
	StartedAt time.Time `json:"started_at"`

	// CompletedAt is when execution finished
	CompletedAt time.Time `json:"completed_at"`

	// Duration is the execution time
	Duration time.Duration `json:"duration"`

	// RetryCount tracks number of retry attempts
	RetryCount int `json:"retry_count"`

	// Metrics contains execution metrics
	Metrics TaskMetrics `json:"metrics"`
}

// TaskMetrics contains execution performance metrics
type TaskMetrics struct {
	// CPUTimeMs is the CPU time used in milliseconds
	CPUTimeMs int64 `json:"cpu_time_ms"`

	// MemoryBytes is the memory used in bytes
	MemoryBytes int64 `json:"memory_bytes"`

	// HandlerCalls is the number of handler invocations
	HandlerCalls int `json:"handler_calls"`

	// DatabaseQueries is the number of database queries
	DatabaseQueries int `json:"database_queries,omitempty"`

	// NetworkRequests is the number of network requests
	NetworkRequests int `json:"network_requests,omitempty"`
}

// RetryPolicy defines retry behavior for failed tasks
type RetryPolicy struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int `json:"max_retries"`

	// InitialDelay is the delay before first retry
	InitialDelay time.Duration `json:"initial_delay"`

	// MaxDelay caps the retry delay
	MaxDelay time.Duration `json:"max_delay"`

	// Multiplier for exponential backoff
	Multiplier float64 `json:"multiplier"`

	// RetryableErrors defines which errors trigger retries
	RetryableErrors []string `json:"retryable_errors,omitempty"`
}

// TaskHandler defines the interface for task execution handlers
type TaskHandler interface {
	// Execute processes the task and returns the result
	Execute(ctx context.Context, task *Task) (*TaskResult, error)

	// Type returns the task type this handler supports
	Type() string

	// Validate checks if the task payload is valid
	Validate(task *Task) error
}

// TaskScheduler defines the interface for task scheduling
type TaskScheduler interface {
	// Schedule adds a task to the queue
	Schedule(task *Task) error

	// Next returns the next task to execute
	Next() (*Task, error)

	// Cancel removes a task from the queue
	Cancel(taskID string) error

	// Size returns the number of queued tasks
	Size() int

	// Clear removes all tasks from the queue
	Clear()
}

// TaskExecutor defines the interface for task execution
type TaskExecutor interface {
	// Execute runs a task using the registered handler
	Execute(ctx context.Context, task *Task) (*TaskResult, error)

	// RegisterHandler registers a handler for a task type
	RegisterHandler(handler TaskHandler) error

	// GetHandler retrieves the handler for a task type
	GetHandler(taskType string) (TaskHandler, error)
}

// TaskManager defines the comprehensive task management interface
type TaskManager interface {
	// Submit adds a task to the execution queue
	Submit(ctx context.Context, task *Task) error

	// RegisterHandler registers a task handler
	RegisterHandler(handler TaskHandler) error

	// GetTask retrieves a task by ID
	GetTask(ctx context.Context, taskID string) (*Task, error)

	// GetTaskResult retrieves task result
	GetTaskResult(ctx context.Context, taskID string) (*TaskResult, error)

	// CancelTask cancels a pending/running task
	CancelTask(ctx context.Context, taskID string) error

	// ListTasks lists tasks with filters
	ListTasks(ctx context.Context, filters TaskFilters) ([]*Task, error)

	// GetMetrics retrieves task execution metrics
	GetMetrics(ctx context.Context, agentID string) (*AgentTaskMetrics, error)

	// Start starts the task execution system
	Start() error

	// Stop stops the task execution system
	Stop() error
}

// TaskRepository defines the interface for task persistence
type TaskRepository interface {
	// StoreTask saves a task to the database
	StoreTask(ctx context.Context, task *Task) error

	// GetTask retrieves a task by ID
	GetTask(ctx context.Context, taskID string) (*Task, error)

	// UpdateTask updates an existing task
	UpdateTask(ctx context.Context, task *Task) error

	// ListTasks lists tasks with filters
	ListTasks(ctx context.Context, agentID string, filters TaskFilters) ([]*Task, error)

	// StoreResult saves a task result
	StoreResult(ctx context.Context, result *TaskResult) error

	// GetResult retrieves a task result
	GetResult(ctx context.Context, taskID string) (*TaskResult, error)

	// GetMetrics retrieves aggregated metrics
	GetMetrics(ctx context.Context, agentID string) (*AgentTaskMetrics, error)

	// UpdateMetrics updates aggregated metrics
	UpdateMetrics(ctx context.Context, metrics *AgentTaskMetrics) error

	// CleanupOldResults removes old task results
	CleanupOldResults(ctx context.Context, before time.Time) (int, error)
}

// TaskFilters defines filtering options for task queries
type TaskFilters struct {
	// Status filters by task status
	Status []TaskStatus `json:"status,omitempty"`

	// Type filters by task type
	Type string `json:"type,omitempty"`

	// Priority filters by minimum priority
	MinPriority int `json:"min_priority,omitempty"`

	// CreatedAfter filters by creation time
	CreatedAfter *time.Time `json:"created_after,omitempty"`

	// CreatedBefore filters by creation time
	CreatedBefore *time.Time `json:"created_before,omitempty"`

	// Limit restricts the number of results
	Limit int `json:"limit,omitempty"`

	// Offset for pagination
	Offset int `json:"offset,omitempty"`

	// SortBy specifies the sort field
	SortBy string `json:"sort_by,omitempty"`

	// SortDesc indicates descending sort order
	SortDesc bool `json:"sort_desc,omitempty"`
}

// AgentTaskMetrics contains aggregated task execution metrics for an agent
type AgentTaskMetrics struct {
	// AgentID is the agent these metrics belong to
	AgentID string `json:"agent_id"`

	// TotalTasks is the total number of tasks
	TotalTasks int64 `json:"total_tasks"`

	// CompletedTasks is the number of completed tasks
	CompletedTasks int64 `json:"completed_tasks"`

	// FailedTasks is the number of failed tasks
	FailedTasks int64 `json:"failed_tasks"`

	// CancelledTasks is the number of cancelled tasks
	CancelledTasks int64 `json:"cancelled_tasks"`

	// TimeoutTasks is the number of timed out tasks
	TimeoutTasks int64 `json:"timeout_tasks"`

	// AvgDurationMs is the average execution time
	AvgDurationMs int64 `json:"avg_duration_ms"`

	// TotalDurationMs is the total execution time
	TotalDurationMs int64 `json:"total_duration_ms"`

	// TasksByType is a breakdown by task type
	TasksByType map[string]int64 `json:"tasks_by_type"`

	// LastUpdated is when metrics were last updated
	LastUpdated time.Time `json:"last_updated"`
}

// SchedulingPolicy defines how tasks are ordered in the queue
type SchedulingPolicy string

const (
	// SchedulingPolicyPriorityFirst prioritizes high-priority tasks
	SchedulingPolicyPriorityFirst SchedulingPolicy = "priority_first"
	// SchedulingPolicyFIFO processes tasks in order of arrival
	SchedulingPolicyFIFO SchedulingPolicy = "fifo"
	// SchedulingPolicyFairShare balances across task types
	SchedulingPolicyFairShare SchedulingPolicy = "fair_share"
	// SchedulingPolicyDeadline prioritizes tasks approaching timeout
	SchedulingPolicyDeadline SchedulingPolicy = "deadline"
)

// WorkerPoolConfig defines worker pool configuration
type WorkerPoolConfig struct {
	// MaxWorkers is the maximum number of concurrent workers
	MaxWorkers int

	// MinWorkers is the minimum number of workers to maintain
	MinWorkers int

	// IdleTimeout is how long idle workers wait before terminating
	IdleTimeout time.Duration

	// QueueSize is the task queue buffer size
	QueueSize int

	// SchedulingPolicy determines task ordering
	SchedulingPolicy SchedulingPolicy
}

// ExecutorConfig defines task executor configuration
type ExecutorConfig struct {
	// DefaultTimeout is the default task timeout
	DefaultTimeout time.Duration

	// MaxTimeout is the maximum allowed timeout
	MaxTimeout time.Duration

	// DefaultRetryPolicy is the default retry policy
	DefaultRetryPolicy *RetryPolicy

	// MetricsEnabled enables metrics collection
	MetricsEnabled bool

	// MetricsInterval is how often to update metrics
	MetricsInterval time.Duration
}

// ManagerConfig defines task manager configuration
type ManagerConfig struct {
	// WorkerPool configuration
	WorkerPool WorkerPoolConfig

	// Executor configuration
	Executor ExecutorConfig

	// PersistTasks enables task persistence
	PersistTasks bool

	// PersistResults enables result persistence
	PersistResults bool

	// ResultRetention is how long to keep results
	ResultRetention time.Duration
}

// NewTask creates a new task with defaults
func NewTask(agentID, taskType, name string, payload map[string]interface{}) *Task {
	return &Task{
		ID:        generateTaskID(),
		AgentID:   agentID,
		Type:      taskType,
		Name:      name,
		Payload:   payload,
		Priority:  5, // Default medium priority
		Status:    TaskStatusPending,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]string),
	}
}

// WithPriority sets the task priority
func (t *Task) WithPriority(priority int) *Task {
	if priority < 0 {
		priority = 0
	}
	if priority > 10 {
		priority = 10
	}
	t.Priority = priority
	return t
}

// WithTimeout sets the task timeout
func (t *Task) WithTimeout(timeout time.Duration) *Task {
	t.Timeout = timeout
	return t
}

// WithRetryPolicy sets the retry policy
func (t *Task) WithRetryPolicy(policy *RetryPolicy) *Task {
	t.RetryPolicy = policy
	return t
}

// WithDependencies sets task dependencies
func (t *Task) WithDependencies(deps ...string) *Task {
	t.Dependencies = deps
	return t
}

// WithMetadata adds metadata to the task
func (t *Task) WithMetadata(key, value string) *Task {
	if t.Metadata == nil {
		t.Metadata = make(map[string]string)
	}
	t.Metadata[key] = value
	return t
}

// SetContext sets the execution context for the task
func (t *Task) SetContext(ctx context.Context) {
	t.ctx, t.cancel = context.WithCancel(ctx)
	if t.Timeout > 0 {
		t.ctx, t.cancel = context.WithTimeout(ctx, t.Timeout)
	}
}

// Context returns the task execution context
func (t *Task) Context() context.Context {
	if t.ctx == nil {
		t.ctx = context.Background()
	}
	return t.ctx
}

// Cancel cancels the task execution
func (t *Task) Cancel() {
	if t.cancel != nil {
		t.cancel()
	}
}

// IsTerminal returns true if the task is in a terminal state
func (t *Task) IsTerminal() bool {
	return t.Status == TaskStatusCompleted ||
		t.Status == TaskStatusFailed ||
		t.Status == TaskStatusCancelled ||
		t.Status == TaskStatusTimeout
}

// ShouldRetry determines if a task should be retried based on the result and policy
func (t *Task) ShouldRetry(result *TaskResult) bool {
	if t.RetryPolicy == nil {
		return false
	}

	if result.RetryCount >= t.RetryPolicy.MaxRetries {
		return false
	}

	// Don't retry completed or cancelled tasks
	if result.Status == TaskStatusCompleted || result.Status == TaskStatusCancelled {
		return false
	}

	// Retry failed and timeout tasks
	return result.Status == TaskStatusFailed || result.Status == TaskStatusTimeout
}

// GetRetryDelay calculates the delay before next retry
func (t *Task) GetRetryDelay(retryCount int) time.Duration {
	if t.RetryPolicy == nil {
		return 0
	}

	// Calculate exponential backoff: InitialDelay * (2^retryCount) * Multiplier
	exponentialFactor := float64(int(1) << uint(retryCount))
	delay := time.Duration(float64(t.RetryPolicy.InitialDelay) * exponentialFactor * t.RetryPolicy.Multiplier)

	if delay > t.RetryPolicy.MaxDelay {
		delay = t.RetryPolicy.MaxDelay
	}

	return delay
}

// DefaultRetryPolicy returns a sensible default retry policy
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:   3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     60 * time.Second,
		Multiplier:   2.0,
	}
}

// generateTaskID generates a unique task ID
func generateTaskID() string {
	// This will be replaced with proper UUID generation
	return "task_" + time.Now().Format("20060102150405")
}
