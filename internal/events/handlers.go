package events

import (
	"context"
	"fmt"

	"github.com/aosanya/CodeValdCortex/internal/communication"
	log "github.com/sirupsen/logrus"
)

// LoggingHandler logs all events for debugging and monitoring
type LoggingHandler struct {
	name     string
	priority int
}

// NewLoggingHandler creates a new logging handler
func NewLoggingHandler() *LoggingHandler {
	return &LoggingHandler{
		name:     "logging_handler",
		priority: 1, // Low priority, runs after other handlers
	}
}

func (h *LoggingHandler) Handle(ctx context.Context, event *Event) error {
	log.WithFields(log.Fields{
		"event_id":   event.ID,
		"event_type": event.Type,
		"agent_id":   event.AgentID,
		"priority":   event.Priority,
		"timestamp":  event.Timestamp,
		"metadata":   event.Metadata,
	}).Info("Event processed")
	return nil
}

func (h *LoggingHandler) CanHandle(eventType EventType) bool {
	return true // Handle all events
}

func (h *LoggingHandler) Priority() int {
	return h.priority
}

func (h *LoggingHandler) Name() string {
	return h.name
}

// MessageHandler processes message-related events
type MessageHandler struct {
	name           string
	priority       int
	messageService *communication.MessageService
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(messageService *communication.MessageService) *MessageHandler {
	return &MessageHandler{
		name:           "message_handler",
		priority:       10, // High priority for message processing
		messageService: messageService,
	}
}

func (h *MessageHandler) Handle(ctx context.Context, event *Event) error {
	switch event.Type {
	case EventTypeMessageReceived:
		return h.handleMessageReceived(ctx, event)
	case EventTypeMessageSent:
		return h.handleMessageSent(ctx, event)
	case EventTypeMessageFailed:
		return h.handleMessageFailed(ctx, event)
	default:
		return nil // Not a message event
	}
}

func (h *MessageHandler) CanHandle(eventType EventType) bool {
	return eventType == EventTypeMessageReceived ||
		eventType == EventTypeMessageSent ||
		eventType == EventTypeMessageFailed
}

func (h *MessageHandler) Priority() int {
	return h.priority
}

func (h *MessageHandler) Name() string {
	return h.name
}

func (h *MessageHandler) handleMessageReceived(_ context.Context, event *Event) error {
	data, ok := event.Data.(*MessageEventData)
	if !ok {
		return fmt.Errorf("invalid event data type for message received event")
	}

	// Process the received message
	log.WithFields(log.Fields{
		"message_id": data.Message.ID,
		"from":       data.Message.FromAgentID,
		"to":         data.Message.ToAgentID,
		"type":       data.Message.MessageType,
	}).Info("Processing received message")

	// Update message status or perform additional processing here

	return nil
}

func (h *MessageHandler) handleMessageSent(_ context.Context, event *Event) error {
	_, ok := event.Data.(*MessageEventData)
	if !ok {
		return fmt.Errorf("invalid event data type for message sent event")
	}

	log.WithField("event_id", event.ID).Info("Message sent successfully")

	return nil
}

func (h *MessageHandler) handleMessageFailed(_ context.Context, event *Event) error {
	data, ok := event.Data.(*MessageEventData)
	if !ok {
		return fmt.Errorf("invalid event data type for message failed event")
	}

	log.WithFields(log.Fields{
		"message_id": data.Message.ID,
		"from":       data.Message.FromAgentID,
		"to":         data.Message.ToAgentID,
		"error":      data.Error,
	}).Error("Message delivery failed")

	// Could implement retry logic or dead letter queue here

	return nil
}

// StateChangeHandler processes agent and system state changes
type StateChangeHandler struct {
	name     string
	priority int
}

// NewStateChangeHandler creates a new state change handler
func NewStateChangeHandler() *StateChangeHandler {
	return &StateChangeHandler{
		name:     "state_change_handler",
		priority: 8, // High priority for state changes
	}
}

func (h *StateChangeHandler) Handle(ctx context.Context, event *Event) error {
	switch event.Type {
	case EventTypeAgentCreated, EventTypeAgentStarted, EventTypeAgentStopped, EventTypeAgentFailed:
		return h.handleAgentStateChange(ctx, event)
	case EventTypePoolCreated, EventTypePoolUpdated, EventTypePoolDeleted:
		return h.handlePoolStateChange(ctx, event)
	case EventTypeTaskCreated, EventTypeTaskStarted, EventTypeTaskCompleted, EventTypeTaskFailed:
		return h.handleTaskStateChange(ctx, event)
	default:
		return nil // Not a state change event
	}
}

func (h *StateChangeHandler) CanHandle(eventType EventType) bool {
	stateChangeEvents := []EventType{
		EventTypeAgentCreated, EventTypeAgentStarted, EventTypeAgentStopped, EventTypeAgentFailed,
		EventTypePoolCreated, EventTypePoolUpdated, EventTypePoolDeleted,
		EventTypeTaskCreated, EventTypeTaskStarted, EventTypeTaskCompleted, EventTypeTaskFailed,
	}

	for _, et := range stateChangeEvents {
		if eventType == et {
			return true
		}
	}
	return false
}

func (h *StateChangeHandler) Priority() int {
	return h.priority
}

func (h *StateChangeHandler) Name() string {
	return h.name
}

func (h *StateChangeHandler) handleAgentStateChange(_ context.Context, event *Event) error {
	data, ok := event.Data.(*AgentEventData)
	if !ok {
		return fmt.Errorf("invalid event data type for agent state change event")
	}

	log.WithFields(log.Fields{
		"agent_id":  data.Agent.ID,
		"old_state": data.OldState,
		"new_state": data.NewState,
		"event":     event.Type,
	}).Info("Agent state changed")

	// Implement state transition logic, notifications, etc.

	return nil
}

func (h *StateChangeHandler) handlePoolStateChange(_ context.Context, event *Event) error {
	data, ok := event.Data.(*PoolEventData)
	if !ok {
		return fmt.Errorf("invalid event data type for pool state change event")
	}

	log.WithFields(log.Fields{
		"pool_id":   data.PoolID,
		"pool_name": data.PoolName,
		"action":    data.Action,
		"event":     event.Type,
	}).Info("Pool state changed")

	// Implement pool change notifications, metrics updates, etc.

	return nil
}

func (h *StateChangeHandler) handleTaskStateChange(_ context.Context, event *Event) error {
	data, ok := event.Data.(*TaskEventData)
	if !ok {
		return fmt.Errorf("invalid event data type for task state change event")
	}

	log.WithFields(log.Fields{
		"task_id":   data.TaskID,
		"agent_id":  data.AgentID,
		"task_type": data.TaskType,
		"status":    data.Status,
		"event":     event.Type,
	}).Info("Task state changed")

	// Implement task completion notifications, result processing, etc.

	return nil
}
