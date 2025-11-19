package work

import "time"

// AgentIssueLink tracks the bidirectional relationship between agents and work items (issues)
// This enables syncing agent state back to the originating work tracking system
type AgentIssueLink struct {
	Key string `json:"_key" arangodb:"_key"`

	// Agent identification
	AgentID string `json:"agent_id"`

	// Work item identification
	IssueID       string `json:"issue_id"`        // Work tracking system issue ID (string for multi-provider)
	IssueNumber   int64  `json:"issue_number"`    // Numeric issue number (Gitea, GitHub)
	RepositoryURL string `json:"repository_url"`  // Full URL to repository
	ProviderType  string `json:"provider_type"`   // "gitea", "github", "gitlab", "jira"

	// Workflow context
	WorkflowID    string `json:"workflow_id"`
	ColumnID      string `json:"column_id"`
	WorkItemDefID string `json:"work_item_def_id"` // Work item definition used to create agent

	// Lifecycle tracking
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Status      string     `json:"status"` // "active", "completed", "failed", "cancelled"

	// Sync metadata
	LastSyncAt      time.Time `json:"last_sync_at"`
	SyncCount       int       `json:"sync_count"`        // Number of sync operations performed
	LastEventType   string    `json:"last_event_type"`   // Last agent event type synced
	LastCommentID   string    `json:"last_comment_id"`   // ID of last comment posted
	CurrentMilestone string   `json:"current_milestone"` // Current milestone/column name
}

// SyncAuditRecord logs all sync operations for compliance and debugging
// Provides complete audit trail of what agent events triggered what work tracking updates
type SyncAuditRecord struct {
	Key string `json:"_key" arangodb:"_key"`

	// Agent context
	AgentID string `json:"agent_id"`
	IssueID string `json:"issue_id"`

	// Event details
	EventType string `json:"event_type"` // "lifecycle", "run", "progress", "custom"
	EventName string `json:"event_name"` // e.g., "agent.lifecycle.healthy", "run.execution.succeeded"

	// Sync operation performed
	SyncAction    string                 `json:"sync_action"`    // "create_comment", "add_label", "remove_label", "update_milestone"
	ActionDetails map[string]interface{} `json:"action_details"` // Provider-specific details (comment ID, label names, etc.)

	// Timing
	EventTimestamp time.Time `json:"event_timestamp"` // When agent event occurred
	SyncTimestamp  time.Time `json:"sync_timestamp"`  // When sync operation was performed

	// Result
	Success      bool   `json:"success"`
	ErrorMessage string `json:"error_message,omitempty"`
	RetryCount   int    `json:"retry_count"` // Number of retry attempts

	// Provider context
	ProviderType string `json:"provider_type"` // "gitea", "github", etc.
	RepositoryURL string `json:"repository_url"`
}

// SyncEventPayload wraps agent events for sync processing
// This is the internal representation used by the sync service
type SyncEventPayload struct {
	// Agent identification
	AgentID   string `json:"agent_id"`
	AgentType string `json:"agent_type"`

	// Event details
	EventType      string                 `json:"event_type"` // "lifecycle", "run", "progress"
	EventName      string                 `json:"event_name"` // Full event name
	EventTimestamp time.Time              `json:"event_timestamp"`
	EventData      map[string]interface{} `json:"event_data"` // Event-specific data

	// State information (for lifecycle events)
	OldState string `json:"old_state,omitempty"`
	NewState string `json:"new_state,omitempty"`

	// Progress information (for progress events)
	ProgressPercentage int    `json:"progress_percentage,omitempty"`
	ProgressMessage    string `json:"progress_message,omitempty"`

	// Error information (for failure events)
	ErrorType    string `json:"error_type,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	StackTrace   string `json:"stack_trace,omitempty"`

	// Task information (for run events)
	TaskID          string `json:"task_id,omitempty"`
	TaskName        string `json:"task_name,omitempty"`
	TaskDescription string `json:"task_description,omitempty"`
	TaskStatus      string `json:"task_status,omitempty"`
	TaskDuration    string `json:"task_duration,omitempty"`
	TaskSummary     string `json:"task_summary,omitempty"`
}

// CommentTemplateData holds data for rendering issue comments
type CommentTemplateData struct {
	// Agent details
	AgentID   string `json:"agent_id"`
	AgentType string `json:"agent_type"`

	// Timing
	Timestamp     time.Time `json:"timestamp"`
	FormattedTime string    `json:"formatted_time"`

	// Workflow context
	ColumnName     string `json:"column_name,omitempty"`
	WorkflowName   string `json:"workflow_name,omitempty"`
	CurrentColumn  string `json:"current_column,omitempty"`
	NextColumn     string `json:"next_column,omitempty"`

	// Status and progress
	Status             string `json:"status,omitempty"`
	StatusMessage      string `json:"status_message,omitempty"`
	ProgressPercentage int    `json:"progress_percentage,omitempty"`
	DetailedDescription string `json:"detailed_description,omitempty"`

	// Task details
	TaskName      string `json:"task_name,omitempty"`
	TaskDuration  string `json:"task_duration,omitempty"`
	TaskSummary   string `json:"task_summary,omitempty"`
	Deliverables  string `json:"deliverables,omitempty"`

	// Error details
	ErrorType        string `json:"error_type,omitempty"`
	ErrorMessage     string `json:"error_message,omitempty"`
	Severity         string `json:"severity,omitempty"`
	StackTrace       string `json:"stack_trace,omitempty"`
	RemediationGuide string `json:"remediation_guide,omitempty"`

	// Work summary (for milestone completion)
	WorkSummary string `json:"work_summary,omitempty"`

	// Custom fields
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// SyncStats tracks sync service performance metrics
type SyncStats struct {
	TotalEvents       int64   `json:"total_events"`
	SuccessfulSyncs   int64   `json:"successful_syncs"`
	FailedSyncs       int64   `json:"failed_syncs"`
	AverageLatencyMs  float64 `json:"average_latency_ms"`
	P95LatencyMs      float64 `json:"p95_latency_ms"`
	P99LatencyMs      float64 `json:"p99_latency_ms"`
	RetryCount        int64   `json:"retry_count"`
	CurrentQueueDepth int     `json:"current_queue_depth"`
}
