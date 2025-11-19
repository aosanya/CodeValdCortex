package work

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/events"
	log "github.com/sirupsen/logrus"
)

// SyncEventHandler implements events.EventHandler to process agent events
// and sync them to work tracking systems
type SyncEventHandler struct {
	syncService SyncService
	name        string
	priority    int
}

// NewSyncEventHandler creates a new event handler that syncs agent events to work items
func NewSyncEventHandler(syncService SyncService) *SyncEventHandler {
	return &SyncEventHandler{
		syncService: syncService,
		name:        "work-sync-handler",
		priority:    100, // High priority for sync operations
	}
}

// Handle processes an event and syncs it to the work tracking system
func (h *SyncEventHandler) Handle(ctx context.Context, event *events.Event) error {
	// Convert event to SyncEventPayload
	payload, err := h.convertEventToPayload(event)
	if err != nil {
		log.WithError(err).WithField("event_type", event.Type).Debug("Skipping non-sync event")
		return nil // Not an error - just not a sync event
	}

	// Handle the sync
	if err := h.syncService.HandleAgentEvent(ctx, payload); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"agent_id":   payload.AgentID,
			"event_type": payload.EventType,
			"event_name": payload.EventName,
		}).Error("Failed to sync agent event")
		return fmt.Errorf("failed to sync agent event: %w", err)
	}

	return nil
}

// CanHandle returns true if this handler can process the given event type
func (h *SyncEventHandler) CanHandle(eventType events.EventType) bool {
	switch eventType {
	case events.EventTypeAgentCreated,
		events.EventTypeAgentStarted,
		events.EventTypeAgentStopped,
		events.EventTypeAgentFailed,
		events.EventTypeAgentHealthChanged,
		events.EventTypeTaskCreated,
		events.EventTypeTaskStarted,
		events.EventTypeTaskCompleted,
		events.EventTypeTaskFailed:
		return true
	default:
		return false
	}
}

// Priority returns the priority of this handler (higher = higher priority)
func (h *SyncEventHandler) Priority() int {
	return h.priority
}

// Name returns a descriptive name for this handler
func (h *SyncEventHandler) Name() string {
	return h.name
}

// convertEventToPayload converts an events.Event to a SyncEventPayload
func (h *SyncEventHandler) convertEventToPayload(event *events.Event) (*SyncEventPayload, error) {
	if event == nil {
		return nil, fmt.Errorf("event is nil")
	}

	payload := &SyncEventPayload{
		AgentID:        event.AgentID,
		EventTimestamp: event.Timestamp,
		EventData:      make(map[string]interface{}),
	}

	// Copy metadata to event data
	if event.Metadata != nil {
		for k, v := range event.Metadata {
			payload.EventData[k] = v
		}
	}

	// Map event types to sync event types
	switch event.Type {
	case events.EventTypeAgentCreated:
		payload.EventType = "lifecycle"
		payload.EventName = "agent.lifecycle.registered"
		if data, ok := event.Data.(*events.AgentEventData); ok {
			payload.AgentType = getAgentType(data)
			payload.NewState = string(data.NewState)
		}

	case events.EventTypeAgentStarted:
		payload.EventType = "lifecycle"
		payload.EventName = "agent.lifecycle.starting"
		if data, ok := event.Data.(*events.AgentEventData); ok {
			payload.AgentType = getAgentType(data)
			payload.NewState = string(data.NewState)
		}

	case events.EventTypeAgentHealthChanged:
		payload.EventType = "lifecycle"
		if data, ok := event.Data.(*events.AgentEventData); ok {
			payload.AgentType = getAgentType(data)
			payload.OldState = string(data.OldState)
			payload.NewState = string(data.NewState)

			// Map health states to event names
			switch data.NewState {
			case "healthy":
				payload.EventName = "agent.lifecycle.healthy"
			case "degraded":
				payload.EventName = "agent.lifecycle.degraded"
			case "quarantined":
				payload.EventName = "agent.lifecycle.quarantined"
			default:
				payload.EventName = "agent.lifecycle.healthy"
			}
		}

	case events.EventTypeAgentStopped:
		payload.EventType = "lifecycle"
		payload.EventName = "agent.lifecycle.stopped"
		if data, ok := event.Data.(*events.AgentEventData); ok {
			payload.AgentType = getAgentType(data)
			payload.OldState = string(data.OldState)
		}

	case events.EventTypeAgentFailed:
		payload.EventType = "lifecycle"
		payload.EventName = "agent.lifecycle.failed"
		if data, ok := event.Data.(*events.AgentEventData); ok {
			payload.AgentType = getAgentType(data)
			if data.Error != nil {
				payload.ErrorMessage = data.Error.Error()
				payload.ErrorType = "agent_failure"
			}
		}

	case events.EventTypeTaskStarted:
		payload.EventType = "run"
		payload.EventName = "run.execution.running"
		if data, ok := event.Data.(*events.TaskEventData); ok {
			payload.TaskID = data.TaskID
			payload.TaskName = data.TaskType
			payload.TaskStatus = data.Status
		}

	case events.EventTypeTaskCompleted:
		payload.EventType = "run"
		payload.EventName = "run.execution.succeeded"
		if data, ok := event.Data.(*events.TaskEventData); ok {
			payload.TaskID = data.TaskID
			payload.TaskName = data.TaskType
			payload.TaskStatus = "succeeded"
			if data.Result != nil {
				if summary, ok := data.Result.(string); ok {
					payload.TaskSummary = summary
				}
			}
		}

	case events.EventTypeTaskFailed:
		payload.EventType = "run"
		payload.EventName = "run.execution.failed"
		if data, ok := event.Data.(*events.TaskEventData); ok {
			payload.TaskID = data.TaskID
			payload.TaskName = data.TaskType
			payload.TaskStatus = "failed"
			if data.Error != nil {
				payload.ErrorMessage = data.Error.Error()
				payload.ErrorType = "task_failure"
			}
		}

	default:
		return nil, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	return payload, nil
}

// getAgentType extracts agent type from AgentEventData
func getAgentType(data *events.AgentEventData) string {
	if data == nil || data.Agent == nil {
		return "unknown"
	}
	return data.Agent.Type
}

// RegisterSyncHandler registers the sync event handler with the event registry
func RegisterSyncHandler(registry *events.HandlerRegistry, syncService SyncService) error {
	handler := NewSyncEventHandler(syncService)

	// Register for all agent and task events
	eventTypes := []events.EventType{
		events.EventTypeAgentCreated,
		events.EventTypeAgentStarted,
		events.EventTypeAgentStopped,
		events.EventTypeAgentFailed,
		events.EventTypeAgentHealthChanged,
		events.EventTypeTaskStarted,
		events.EventTypeTaskCompleted,
		events.EventTypeTaskFailed,
	}

	if err := registry.RegisterHandler(handler, eventTypes...); err != nil {
		return fmt.Errorf("failed to register sync handler: %w", err)
	}

	log.WithFields(log.Fields{
		"handler":     handler.Name(),
		"event_types": len(eventTypes),
	}).Info("Sync event handler registered")

	return nil
}

// SyncEventPublisher provides helper methods to publish sync-specific events
type SyncEventPublisher struct {
	processor *events.Processor
}

// NewSyncEventPublisher creates a new sync event publisher
func NewSyncEventPublisher(processor *events.Processor) *SyncEventPublisher {
	return &SyncEventPublisher{
		processor: processor,
	}
}

// PublishProgressUpdate publishes a progress update event
func (p *SyncEventPublisher) PublishProgressUpdate(ctx context.Context, agentID string, percentage int, message string) error {
	event := &events.Event{
		ID:        fmt.Sprintf("progress-%s-%d", agentID, time.Now().UnixNano()),
		Type:      "progress_update", // Custom event type
		Priority:  events.PriorityNormal,
		AgentID:   agentID,
		Timestamp: time.Now(),
		Context:   ctx,
		Metadata: map[string]interface{}{
			"percentage": percentage,
			"message":    message,
		},
	}

	return p.processor.PublishEvent(event)
}

// PublishMilestoneComplete publishes a milestone completion event
func (p *SyncEventPublisher) PublishMilestoneComplete(ctx context.Context, agentID string, summary string) error {
	event := &events.Event{
		ID:        fmt.Sprintf("milestone-%s-%d", agentID, time.Now().UnixNano()),
		Type:      "milestone_complete", // Custom event type
		Priority:  events.PriorityHigh,
		AgentID:   agentID,
		Timestamp: time.Now(),
		Context:   ctx,
		Metadata: map[string]interface{}{
			"summary": summary,
		},
	}

	return p.processor.PublishEvent(event)
}
