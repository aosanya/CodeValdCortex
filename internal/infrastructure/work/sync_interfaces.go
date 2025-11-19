package work

import (
	"context"
)

// SyncService handles bidirectional synchronization between agents and work tracking systems
// It transforms agent lifecycle and execution events into work tracking updates (comments, labels, milestones)
type SyncService interface {
	// HandleAgentEvent processes an agent event and syncs it to the work tracking system
	HandleAgentEvent(ctx context.Context, event *SyncEventPayload) error

	// PostComment creates an issue comment from an agent event
	PostComment(ctx context.Context, link *AgentIssueLink, template string, data *CommentTemplateData) error

	// UpdateLabels adds or removes labels from an issue based on agent state
	UpdateLabels(ctx context.Context, link *AgentIssueLink, add []string, remove []string) error

	// ProgressMilestone moves an issue to the next workflow column/milestone
	ProgressMilestone(ctx context.Context, link *AgentIssueLink) error

	// RecordAudit logs a sync operation to the audit trail
	RecordAudit(ctx context.Context, record *SyncAuditRecord) error

	// GetStats returns sync service performance metrics
	GetStats(ctx context.Context) (*SyncStats, error)
}

// AgentIssueLinkRepository manages agent-to-issue link persistence
type AgentIssueLinkRepository interface {
	// Create creates a new agent-issue link
	Create(ctx context.Context, link *AgentIssueLink) error

	// GetByAgentID retrieves the link for a specific agent
	GetByAgentID(ctx context.Context, agentID string) (*AgentIssueLink, error)

	// GetByIssueID retrieves the link for a specific issue
	GetByIssueID(ctx context.Context, issueID string) (*AgentIssueLink, error)

	// UpdateStatus updates the link status (active, completed, failed)
	UpdateStatus(ctx context.Context, agentID string, status string) error

	// UpdateLastSync updates the last sync timestamp and metadata
	UpdateLastSync(ctx context.Context, agentID string, eventType string, commentID string) error

	// MarkCompleted marks the link as completed with timestamp
	MarkCompleted(ctx context.Context, agentID string) error

	// ListActive returns all active agent-issue links
	ListActive(ctx context.Context) ([]*AgentIssueLink, error)

	// Delete removes a link
	Delete(ctx context.Context, agentID string) error
}

// SyncAuditRepository manages sync audit record persistence
type SyncAuditRepository interface {
	// Create creates a new audit record
	Create(ctx context.Context, record *SyncAuditRecord) error

	// GetByAgentID retrieves all audit records for an agent
	GetByAgentID(ctx context.Context, agentID string, limit int) ([]*SyncAuditRecord, error)

	// GetByIssueID retrieves all audit records for an issue
	GetByIssueID(ctx context.Context, issueID string, limit int) ([]*SyncAuditRecord, error)

	// GetFailures retrieves failed sync operations for retry/investigation
	GetFailures(ctx context.Context, since int64, limit int) ([]*SyncAuditRecord, error)

	// GetStats calculates sync statistics for monitoring
	GetStats(ctx context.Context, since int64) (*SyncStats, error)
}

// TemplateRenderer renders comment templates with data
type TemplateRenderer interface {
	// Render renders a template with the provided data
	Render(ctx context.Context, templateName string, data *CommentTemplateData) (string, error)

	// RegisterTemplate registers a new template
	RegisterTemplate(name string, template string) error

	// GetTemplate retrieves a template by name
	GetTemplate(name string) (string, error)
}

// WorkflowService provides workflow and column information for milestone progression
type WorkflowService interface {
	// GetWorkflow retrieves workflow by ID
	GetWorkflow(ctx context.Context, workflowID string) (*Workflow, error)

	// GetNextColumn determines the next column in workflow sequence
	GetNextColumn(ctx context.Context, workflowID string, currentColumnID string) (*WorkflowColumn, error)

	// GetMilestoneMapping gets the milestone name for a workflow column
	GetMilestoneMapping(ctx context.Context, workflowID string, columnID string) (string, error)
}

// Workflow represents a Kanban workflow with columns
type Workflow struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	AgencyID    string            `json:"agency_id"`
	Columns     []*WorkflowColumn `json:"columns"`
	Description string            `json:"description,omitempty"`
}

// WorkflowColumn represents a column/stage in a workflow
type WorkflowColumn struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Position      int    `json:"position"`
	WorkItemDefID string `json:"work_item_def_id"`
	MilestoneName string `json:"milestone_name"` // Mapped milestone name in work tracking system
	AutoAssign    bool   `json:"auto_assign"`
	MaxConcurrent int    `json:"max_concurrent"`
	NextColumnID  string `json:"next_column_id,omitempty"`
}
