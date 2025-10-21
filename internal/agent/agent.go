package agent

import (
	"context"
	"sync"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/communication"
	"github.com/google/uuid"
)

// State represents the current state of an agent
type State string

const (
	// StateCreated indicates agent has been created but not started
	StateCreated State = "created"
	// StateRunning indicates agent is actively processing tasks
	StateRunning State = "running"
	// StatePaused indicates agent is paused and not processing tasks
	StatePaused State = "paused"
	// StateStopped indicates agent has been stopped gracefully
	StateStopped State = "stopped"
	// StateFailed indicates agent has encountered an error
	StateFailed State = "failed"
)

// Agent represents a single agent instance
type Agent struct {
	// ID is the unique identifier for the agent
	ID string

	// Name is a human-readable name for the agent
	Name string

	// Type indicates the agent type (e.g., "worker", "coordinator")
	Type string

	// State is the current state of the agent
	State State

	// Metadata contains additional agent information
	Metadata map[string]string

	// Config holds agent configuration
	Config Config

	// CreatedAt is the timestamp when agent was created
	CreatedAt time.Time

	// UpdatedAt is the timestamp of last state update
	UpdatedAt time.Time

	// LastHeartbeat is the timestamp of last health check
	LastHeartbeat time.Time

	// ctx is the agent's context for cancellation
	ctx context.Context

	// cancel is the function to cancel agent context
	cancel context.CancelFunc

	// mu protects concurrent access to agent state
	mu sync.RWMutex

	// taskChan is the channel for receiving tasks
	taskChan chan Task

	// done signals agent shutdown completion
	done chan struct{}

	// errChan reports errors from the agent
	errChan chan error

	// doneClosed tracks if done channel is already closed
	doneClosed bool

	// Communication services
	messageService *communication.MessageService
	pubSubService  *communication.PubSubService
	commPoller     *communication.CommunicationPoller
}

// Config holds agent configuration
type Config struct {
	// MaxConcurrentTasks limits concurrent task execution
	MaxConcurrentTasks int

	// TaskQueueSize sets the task channel buffer size
	TaskQueueSize int

	// HeartbeatInterval defines health check frequency
	HeartbeatInterval time.Duration

	// TaskTimeout sets the default task execution timeout
	TaskTimeout time.Duration

	// Resources defines agent resource allocations
	Resources Resources
}

// Resources defines agent resource requirements
type Resources struct {
	// CPU in millicores (e.g., 100 = 0.1 CPU)
	CPU int

	// Memory in megabytes
	Memory int

	// MaxTasks is the maximum number of tasks in queue
	MaxTasks int
}

// Task represents a unit of work for an agent
type Task struct {
	// ID is the unique task identifier
	ID string

	// Type indicates the task type
	Type string

	// Payload contains task-specific data
	Payload interface{}

	// Priority for task ordering (higher = more important)
	Priority int

	// Timeout for task execution
	Timeout time.Duration

	// CreatedAt is when the task was created
	CreatedAt time.Time
}

// TaskResult represents the outcome of task execution
type TaskResult struct {
	// TaskID is the ID of the executed task
	TaskID string

	// AgentID is the ID of the agent that executed the task
	AgentID string

	// Success indicates if task completed successfully
	Success bool

	// Result contains the task output
	Result interface{}

	// Error contains error information if task failed
	Error error

	// StartedAt is when task execution began
	StartedAt time.Time

	// CompletedAt is when task execution finished
	CompletedAt time.Time

	// Duration is the execution time
	Duration time.Duration
}

// New creates a new agent with the given configuration
func New(name, agentType string, config Config) *Agent {
	ctx, cancel := context.WithCancel(context.Background())

	// Set default configuration values
	if config.MaxConcurrentTasks == 0 {
		config.MaxConcurrentTasks = 5
	}
	if config.TaskQueueSize == 0 {
		config.TaskQueueSize = 100
	}
	if config.HeartbeatInterval == 0 {
		config.HeartbeatInterval = 30 * time.Second
	}
	if config.TaskTimeout == 0 {
		config.TaskTimeout = 5 * time.Minute
	}

	return &Agent{
		ID:            uuid.New().String(),
		Name:          name,
		Type:          agentType,
		State:         StateCreated,
		Metadata:      make(map[string]string),
		Config:        config,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
		LastHeartbeat: time.Now().UTC(),
		ctx:           ctx,
		cancel:        cancel,
		taskChan:      make(chan Task, config.TaskQueueSize),
		done:          make(chan struct{}),
		errChan:       make(chan error, 10),
	}
}

// GetState returns the current agent state (thread-safe)
func (a *Agent) GetState() State {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.State
}

// SetState updates the agent state (thread-safe)
func (a *Agent) SetState(state State) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.State = state
	a.UpdatedAt = time.Now().UTC()
}

// UpdateHeartbeat updates the last heartbeat timestamp
func (a *Agent) UpdateHeartbeat() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.LastHeartbeat = time.Now().UTC()
}

// IsHealthy checks if agent is responsive
func (a *Agent) IsHealthy() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Agent is unhealthy if no heartbeat within 2x interval
	threshold := a.Config.HeartbeatInterval * 2
	return time.Since(a.LastHeartbeat) < threshold
}

// Context returns the agent's context
func (a *Agent) Context() context.Context {
	return a.ctx
}

// SubmitTask adds a task to the agent's queue
func (a *Agent) SubmitTask(task Task) error {
	select {
	case a.taskChan <- task:
		return nil
	case <-a.ctx.Done():
		return ErrAgentStopped
	default:
		return ErrTaskQueueFull
	}
}

// Done returns a channel that closes when agent shuts down
func (a *Agent) Done() <-chan struct{} {
	return a.done
}

// Errors returns a channel for receiving agent errors
func (a *Agent) Errors() <-chan error {
	return a.errChan
}

// TaskChan returns the task channel for the agent (for runtime manager use)
func (a *Agent) TaskChan() <-chan Task {
	return a.taskChan
}

// Cancel cancels the agent's context (for runtime manager use)
func (a *Agent) Cancel() {
	a.cancel()
}

// Communication methods

// SetupCommunication initializes communication services for the agent
func (a *Agent) SetupCommunication(messageService *communication.MessageService, pubSubService *communication.PubSubService) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.messageService = messageService
	a.pubSubService = pubSubService
}

// StartCommunicationPolling starts polling for messages and publications
func (a *Agent) StartCommunicationPolling(messageInterval, publicationInterval time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.messageService == nil || a.pubSubService == nil {
		return // Communication not set up
	}

	if a.commPoller != nil && a.commPoller.IsRunning() {
		return // Already running
	}

	// Create combined poller
	a.commPoller = communication.NewCommunicationPoller(
		communication.MessagePollerConfig{
			AgentID:   a.ID,
			Interval:  messageInterval,
			BatchSize: 100,
		},
		communication.PublicationPollerConfig{
			AgentID:  a.ID,
			Interval: publicationInterval,
		},
		a.messageService,
		a.pubSubService,
		a.handleIncomingMessage,
		a.handleIncomingPublication,
	)

	a.commPoller.Start()
}

// StopCommunicationPolling stops polling for messages and publications
func (a *Agent) StopCommunicationPolling() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.commPoller != nil {
		a.commPoller.Stop()
	}
}

// SendMessage sends a direct message to another agent
func (a *Agent) SendMessage(toAgentID string, msgType communication.MessageType, payload map[string]interface{}, opts *communication.MessageOptions) (string, error) {
	if a.messageService == nil {
		return "", ErrCommunicationNotSetup
	}

	return a.messageService.SendMessage(a.ctx, a.ID, toAgentID, msgType, payload, opts)
}

// Subscribe creates a subscription to events matching the pattern
func (a *Agent) Subscribe(eventPattern string, filters *communication.SubscriptionFilters) (string, error) {
	if a.pubSubService == nil {
		return "", ErrCommunicationNotSetup
	}

	return a.pubSubService.Subscribe(a.ctx, a.ID, a.Type, eventPattern, filters)
}

// Unsubscribe deactivates a subscription
func (a *Agent) Unsubscribe(subscriptionID string) error {
	if a.pubSubService == nil {
		return ErrCommunicationNotSetup
	}

	return a.pubSubService.Unsubscribe(a.ctx, subscriptionID)
}

// Publish publishes an event to all subscribers
func (a *Agent) Publish(eventName string, payload map[string]interface{}, opts *communication.PublicationOptions) (string, error) {
	if a.pubSubService == nil {
		return "", ErrCommunicationNotSetup
	}

	return a.pubSubService.Publish(a.ctx, a.ID, a.Type, eventName, payload, opts)
}

// handleIncomingMessage processes received messages (can be overridden by custom handlers)
func (a *Agent) handleIncomingMessage(msg *communication.Message) error {
	// Default implementation: log the message
	// In a real implementation, this would route to appropriate handlers based on message type
	// For now, we acknowledge the message
	if a.messageService != nil {
		return a.messageService.AcknowledgeMessage(a.ctx, msg.ID)
	}
	return nil
}

// handleIncomingPublication processes received publications (can be overridden by custom handlers)
func (a *Agent) handleIncomingPublication(pub *communication.Publication) error {
	// Default implementation: log the publication
	// In a real implementation, this would route to appropriate handlers based on event type
	return nil
}
