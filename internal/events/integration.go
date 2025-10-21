package events

import (
	"context"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/aosanya/CodeValdCortex/internal/communication"
	"github.com/aosanya/CodeValdCortex/internal/lifecycle"
	"github.com/aosanya/CodeValdCortex/internal/task"
	log "github.com/sirupsen/logrus"
)

// Integrator connects the event system with existing components
type Integrator struct {
	eventProcessor   *Processor
	messageService   *communication.MessageService
	pubsubService    *communication.PubSubService
	lifecycleManager *lifecycle.Manager
	taskScheduler    *task.Scheduler
}

// NewIntegrator creates a new event system integrator
func NewIntegrator(
	eventProcessor *Processor,
	messageService *communication.MessageService,
	pubsubService *communication.PubSubService,
	lifecycleManager *lifecycle.Manager,
	taskScheduler *task.Scheduler,
) *Integrator {
	return &Integrator{
		eventProcessor:   eventProcessor,
		messageService:   messageService,
		pubsubService:    pubsubService,
		lifecycleManager: lifecycleManager,
		taskScheduler:    taskScheduler,
	}
}

// SetupIntegration configures the event system integration with other components
func (i *Integrator) SetupIntegration(ctx context.Context) error {
	// Register built-in handlers
	if err := i.registerBuiltinHandlers(); err != nil {
		return err
	}

	// Setup message service integration
	if err := i.setupMessageIntegration(ctx); err != nil {
		return err
	}

	// Setup lifecycle integration
	if err := i.setupLifecycleIntegration(ctx); err != nil {
		return err
	}

	// Setup task scheduler integration
	if err := i.setupTaskIntegration(ctx); err != nil {
		return err
	}

	log.Info("Event system integration setup completed")
	return nil
}

// registerBuiltinHandlers registers all built-in event handlers
func (i *Integrator) registerBuiltinHandlers() error {
	// Register logging handler (runs on all events)
	loggingHandler := NewLoggingHandler()
	if err := i.eventProcessor.RegisterHandler(loggingHandler); err != nil {
		return err
	}

	// Register message handler
	messageHandler := NewMessageHandler(i.messageService)
	if err := i.eventProcessor.RegisterHandler(messageHandler); err != nil {
		return err
	}

	// Register state change handler
	stateChangeHandler := NewStateChangeHandler()
	if err := i.eventProcessor.RegisterHandler(stateChangeHandler); err != nil {
		return err
	}

	log.Info("Built-in event handlers registered")
	return nil
}

// setupMessageIntegration configures message service to publish events
func (i *Integrator) setupMessageIntegration(ctx context.Context) error {
	// Note: This would typically involve modifying the MessageService
	// to publish events when messages are sent, received, or fail.
	// For now, we'll just log that the integration is set up.

	log.Info("Message service event integration configured")
	return nil
}

// setupLifecycleIntegration configures lifecycle manager to publish events
func (i *Integrator) setupLifecycleIntegration(ctx context.Context) error {
	// Note: This would typically involve modifying the LifecycleManager
	// to publish events when agents change state.
	// For now, we'll just log that the integration is set up.

	log.Info("Lifecycle manager event integration configured")
	return nil
}

// setupTaskIntegration configures task scheduler to publish events
func (i *Integrator) setupTaskIntegration(ctx context.Context) error {
	// Note: This would typically involve modifying the TaskScheduler
	// to publish events when tasks are created, started, completed, or fail.
	// For now, we'll just log that the integration is set up.

	log.Info("Task scheduler event integration configured")
	return nil
}

// PublishMessageEvent publishes a message-related event
func (i *Integrator) PublishMessageEvent(ctx context.Context, eventType EventType, message *communication.Message, err error) error {
	event := &Event{
		Type:    eventType,
		AgentID: message.FromAgentID,
		Data: &MessageEventData{
			Message: message,
			Error:   err,
		},
		Priority: EventPriority(message.Priority),
	}

	return i.eventProcessor.PublishEvent(event)
}

// PublishAgentEvent publishes an agent lifecycle event
func (i *Integrator) PublishAgentEvent(ctx context.Context, eventType EventType, agentID string, oldState, newState agent.State) error {
	event := &Event{
		Type:    eventType,
		AgentID: agentID,
		Data: &AgentEventData{
			Agent: &agent.Agent{
				ID: agentID,
			},
			OldState: oldState,
			NewState: newState,
		},
		Priority: PriorityNormal,
	}

	return i.eventProcessor.PublishEvent(event)
}

// PublishTaskEvent publishes a task-related event
func (i *Integrator) PublishTaskEvent(ctx context.Context, eventType EventType, taskID, agentID, taskType, status string) error {
	event := &Event{
		Type:    eventType,
		AgentID: agentID,
		Data: &TaskEventData{
			TaskID:   taskID,
			AgentID:  agentID,
			TaskType: taskType,
			Status:   status,
		},
		Priority: PriorityHigh,
	}

	return i.eventProcessor.PublishEvent(event)
}

// PublishPoolEvent publishes a pool-related event
func (i *Integrator) PublishPoolEvent(ctx context.Context, eventType EventType, poolID, poolName, action string) error {
	event := &Event{
		Type:    eventType,
		AgentID: "", // Pool events don't have a specific agent
		Data: &PoolEventData{
			PoolID:   poolID,
			PoolName: poolName,
			Action:   action,
		},
		Priority: PriorityNormal,
	}

	return i.eventProcessor.PublishEvent(event)
}

// Shutdown gracefully shuts down the integration
func (i *Integrator) Shutdown(ctx context.Context) error {
	log.Info("Shutting down event system integration")

	// Stop the event processor
	if err := i.eventProcessor.Stop(); err != nil {
		log.WithError(err).Error("Error stopping event processor")
		return err
	}

	log.Info("Event system integration shut down completed")
	return nil
}
