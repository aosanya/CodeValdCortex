package events

import (
	"context"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/aosanya/CodeValdCortex/internal/communication"
)

// EventType defines the type of event being processed
type EventType string

const (
	// Message events
	EventTypeMessageReceived EventType = "message_received"
	EventTypeMessageSent     EventType = "message_sent"
	EventTypeMessageFailed   EventType = "message_failed"

	// Agent lifecycle events
	EventTypeAgentCreated       EventType = "agent_created"
	EventTypeAgentStarted       EventType = "agent_started"
	EventTypeAgentStopped       EventType = "agent_stopped"
	EventTypeAgentFailed        EventType = "agent_failed"
	EventTypeAgentHealthChanged EventType = "agent_health_changed"

	// Pool events
	EventTypePoolCreated          EventType = "pool_created"
	EventTypePoolUpdated          EventType = "pool_updated"
	EventTypePoolDeleted          EventType = "pool_deleted"
	EventTypeAgentAddedToPool     EventType = "agent_added_to_pool"
	EventTypeAgentRemovedFromPool EventType = "agent_removed_from_pool"

	// Task events
	EventTypeTaskCreated   EventType = "task_created"
	EventTypeTaskStarted   EventType = "task_started"
	EventTypeTaskCompleted EventType = "task_completed"
	EventTypeTaskFailed    EventType = "task_failed"

	// System events
	EventTypeSystemStartup  EventType = "system_startup"
	EventTypeSystemShutdown EventType = "system_shutdown"
	EventTypeConfigChanged  EventType = "config_changed"
)

// EventPriority defines the priority level for event processing
type EventPriority int

const (
	PriorityLow EventPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// Event represents a system event that can be processed by handlers
type Event struct {
	// ID is a unique identifier for this event instance
	ID string

	// Type identifies what kind of event this is
	Type EventType

	// Priority determines processing order
	Priority EventPriority

	// AgentID is the agent associated with this event (if applicable)
	AgentID string

	// Data contains event-specific payload
	Data interface{}

	// Metadata contains additional context information
	Metadata map[string]interface{}

	// Timestamp when the event was created
	Timestamp time.Time

	// Context for cancellation and timeouts
	Context context.Context
}

// EventHandler defines the interface for processing events
type EventHandler interface {
	// Handle processes the event and returns an error if processing fails
	Handle(ctx context.Context, event *Event) error

	// CanHandle returns true if this handler can process the given event type
	CanHandle(eventType EventType) bool

	// Priority returns the priority of this handler (higher values = higher priority)
	Priority() int

	// Name returns a descriptive name for this handler
	Name() string
}

// HandlerRegistration contains information about a registered handler
type HandlerRegistration struct {
	Handler   EventHandler
	EventType EventType
	Priority  int
	AgentID   string // Empty for global handlers
}

// EventResult represents the result of event processing
type EventResult struct {
	EventID      string
	Processed    bool
	Error        error
	Duration     time.Duration
	HandlerCount int
}

// MessageEventData contains data for message-related events
type MessageEventData struct {
	Message *communication.Message
	Error   error
}

// AgentEventData contains data for agent lifecycle events
type AgentEventData struct {
	Agent    *agent.Agent
	OldState agent.State
	NewState agent.State
	Error    error
}

// PoolEventData contains data for pool-related events
type PoolEventData struct {
	PoolID   string
	AgentID  string
	PoolName string
	Action   string
	Error    error
}

// TaskEventData contains data for task-related events
type TaskEventData struct {
	TaskID   string
	AgentID  string
	TaskType string
	Status   string
	Error    error
	Result   interface{}
}

// SystemEventData contains data for system-level events
type SystemEventData struct {
	Component string
	Action    string
	Config    map[string]interface{}
	Error     error
}
