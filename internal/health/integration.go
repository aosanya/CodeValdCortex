package health

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/communication"
	"github.com/aosanya/CodeValdCortex/internal/events"
	log "github.com/sirupsen/logrus"
)

// EventIntegrator connects health monitoring with the event processing system
type EventIntegrator struct {
	eventProcessor *events.Processor
	pubsubService  *communication.PubSubService
	logger         *log.Logger
}

// NewEventIntegrator creates a new health event integrator
func NewEventIntegrator(
	eventProcessor *events.Processor,
	pubsubService *communication.PubSubService,
	logger *log.Logger,
) *EventIntegrator {
	if logger == nil {
		logger = log.New()
	}

	return &EventIntegrator{
		eventProcessor: eventProcessor,
		pubsubService:  pubsubService,
		logger:         logger,
	}
}

// PublishHealthEvent publishes a health event to both the event system and pub/sub
func (ei *EventIntegrator) PublishHealthEvent(ctx context.Context, healthEvent *HealthEvent) error {
	// Convert health event to system event using simple data
	event := &events.Event{
		Type:    ei.mapHealthEventToSystemEvent(healthEvent.Type),
		AgentID: healthEvent.AgentID,
		Data: map[string]interface{}{
			"health_event_type": string(healthEvent.Type),
			"agent_id":          healthEvent.AgentID,
			"previous_status":   string(healthEvent.PreviousStatus),
			"current_status":    string(healthEvent.CurrentStatus),
			"message":           healthEvent.Message,
			"timestamp":         healthEvent.Timestamp,
		},
		Priority: ei.mapHealthStatusToPriority(healthEvent.CurrentStatus),
		Metadata: map[string]interface{}{
			"health_event_type": string(healthEvent.Type),
			"previous_status":   string(healthEvent.PreviousStatus),
			"current_status":    string(healthEvent.CurrentStatus),
			"source":            "health_monitor",
			"health_message":    healthEvent.Message,
		},
	}

	// Publish to event system
	if err := ei.eventProcessor.PublishEvent(event); err != nil {
		ei.logger.WithError(err).Error("Failed to publish health event to event system")
		return fmt.Errorf("failed to publish health event: %w", err)
	}

	// Publish to pub/sub system for broadcasting
	if ei.pubsubService != nil {
		if err := ei.publishToPubSub(ctx, healthEvent); err != nil {
			ei.logger.WithError(err).Warn("Failed to publish health event to pub/sub (non-critical)")
			// Don't fail the entire operation if pub/sub fails
		}
	}

	ei.logger.WithFields(log.Fields{
		"agent_id":        healthEvent.AgentID,
		"event_type":      healthEvent.Type,
		"current_status":  healthEvent.CurrentStatus,
		"previous_status": healthEvent.PreviousStatus,
	}).Info("Published health event")

	return nil
}

// publishToPubSub publishes health status to pub/sub for broadcasting
func (ei *EventIntegrator) publishToPubSub(ctx context.Context, healthEvent *HealthEvent) error {
	// Create payload for health status broadcasting
	payload := map[string]interface{}{
		"agent_id":        healthEvent.AgentID,
		"health_status":   healthEvent.CurrentStatus,
		"event_type":      healthEvent.Type,
		"message":         healthEvent.Message,
		"timestamp":       healthEvent.Timestamp,
		"previous_status": healthEvent.PreviousStatus,
		"metadata":        healthEvent.Metadata,
	}

	// Set routing patterns based on health status
	routingPatterns := ei.getRoutingPatterns(healthEvent)
	payload["routing_patterns"] = routingPatterns

	// Publish to agent-specific health channel
	agentEventName := fmt.Sprintf("agent.%s.health", healthEvent.AgentID)
	_, err := ei.pubsubService.Publish(ctx, "health_monitor", "system", agentEventName, payload, &communication.PublicationOptions{
		TTLSeconds: 300, // 5 minutes TTL
	})
	if err != nil {
		return fmt.Errorf("failed to publish agent health status to pub/sub: %w", err)
	}

	// Also publish to global health status channel
	globalPayload := map[string]interface{}{
		"agent_id":      healthEvent.AgentID,
		"health_status": healthEvent.CurrentStatus,
		"event_type":    healthEvent.Type,
		"timestamp":     healthEvent.Timestamp,
	}

	_, err = ei.pubsubService.Publish(ctx, "health_monitor", "system", "system.health.status", globalPayload, &communication.PublicationOptions{
		TTLSeconds: 300,
	})
	if err != nil {
		ei.logger.WithError(err).Warn("Failed to publish to global health channel")
		// Non-critical error
	}

	return nil
}

// mapHealthEventToSystemEvent maps health event types to system event types
func (ei *EventIntegrator) mapHealthEventToSystemEvent(healthEventType HealthEventType) events.EventType {
	switch healthEventType {
	case HealthEventAgentHealthy:
		return events.EventTypeAgentStarted
	case HealthEventAgentDegraded:
		return events.EventTypeAgentStarted // Could be a custom health event type
	case HealthEventAgentUnhealthy:
		return events.EventTypeAgentFailed
	case HealthEventAgentCritical:
		return events.EventTypeAgentFailed
	case HealthEventRecovery:
		return events.EventTypeAgentStarted
	case HealthEventCheckFailed:
		return events.EventTypeAgentFailed
	default:
		return events.EventTypeAgentFailed
	}
}

// mapHealthStatusToPriority maps health status to event priority
func (ei *EventIntegrator) mapHealthStatusToPriority(status HealthStatus) events.EventPriority {
	switch status {
	case HealthStatusCritical:
		return events.PriorityCritical
	case HealthStatusUnhealthy:
		return events.PriorityHigh
	case HealthStatusDegraded:
		return events.PriorityNormal
	case HealthStatusHealthy:
		return events.PriorityLow
	default:
		return events.PriorityNormal
	}
}

// getRoutingPatterns determines pub/sub routing patterns based on health event
func (ei *EventIntegrator) getRoutingPatterns(healthEvent *HealthEvent) []string {
	patterns := []string{
		fmt.Sprintf("agent.%s.*", healthEvent.AgentID),
		"system.health.*",
	}

	// Add status-specific patterns
	switch healthEvent.CurrentStatus {
	case HealthStatusCritical:
		patterns = append(patterns, "alerts.critical.*", "notifications.urgent.*")
	case HealthStatusUnhealthy:
		patterns = append(patterns, "alerts.warning.*", "notifications.warning.*")
	case HealthStatusDegraded:
		patterns = append(patterns, "monitoring.degraded.*")
	case HealthStatusHealthy:
		patterns = append(patterns, "monitoring.healthy.*")
	}

	// Add event type specific patterns
	switch healthEvent.Type {
	case HealthEventRecovery:
		patterns = append(patterns, "recovery.*", "notifications.recovery.*")
	case HealthEventCheckFailed:
		patterns = append(patterns, "failures.*", "diagnostics.*")
	}

	return patterns
}

// HealthMetricsCollector provides advanced metrics collection for health monitoring
type HealthMetricsCollector struct {
	monitor *Monitor
	logger  *log.Logger
}

// NewHealthMetricsCollector creates a new health metrics collector
func NewHealthMetricsCollector(monitor *Monitor, logger *log.Logger) *HealthMetricsCollector {
	if logger == nil {
		logger = log.New()
	}

	return &HealthMetricsCollector{
		monitor: monitor,
		logger:  logger,
	}
}

// CollectAgentMetrics collects detailed metrics for a specific agent
func (hmc *HealthMetricsCollector) CollectAgentMetrics(agentID string) (*AgentMetrics, error) {
	report, err := hmc.monitor.GetHealthReport(agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get health report for agent %s: %w", agentID, err)
	}

	// Extract metrics from health report
	metrics := &AgentMetrics{
		LastHeartbeat: report.LastHealthyTime,
		UptimeSeconds: int64(time.Since(report.LastHealthyTime).Seconds()),
	}

	// Collect metrics from health check results
	for _, checkResult := range report.CheckResults {
		switch checkResult.CheckType {
		case CheckTypeHeartbeat:
			if lastHeartbeat, ok := checkResult.Details["last_heartbeat"].(time.Time); ok {
				metrics.LastHeartbeat = lastHeartbeat
			}
		case CheckTypeResource:
			if memUsage, ok := checkResult.Details["memory_usage_bytes"].(int64); ok {
				metrics.MemoryUsage = memUsage
			}
			if memPercent, ok := checkResult.Details["memory_usage_percent"].(float64); ok {
				metrics.MemoryPercent = memPercent
			}
			if goroutines, ok := checkResult.Details["goroutines"].(int); ok {
				metrics.Goroutines = goroutines
			}
		case CheckTypePerformance:
			// Would extract task metrics in a real implementation
			metrics.TasksCompleted = 0 // Placeholder
			metrics.TasksFailed = 0    // Placeholder
			metrics.TasksInQueue = 0   // Placeholder
		case CheckTypeConnectivity:
			metrics.ConnectionsActive = 1 // Simplified
		}

		// Calculate response time metrics
		duration := float64(checkResult.Duration.Milliseconds())
		if metrics.AvgResponseTime == 0 {
			metrics.AvgResponseTime = duration
		} else {
			metrics.AvgResponseTime = (metrics.AvgResponseTime + duration) / 2
		}
	}

	return metrics, nil
}

// HealthStatusBroadcaster manages broadcasting of health status changes
type HealthStatusBroadcaster struct {
	eventIntegrator *EventIntegrator
	subscribers     map[string][]chan *HealthEvent
	mu              sync.RWMutex
	logger          *log.Logger
}

// NewHealthStatusBroadcaster creates a new health status broadcaster
func NewHealthStatusBroadcaster(eventIntegrator *EventIntegrator, logger *log.Logger) *HealthStatusBroadcaster {
	if logger == nil {
		logger = log.New()
	}

	return &HealthStatusBroadcaster{
		eventIntegrator: eventIntegrator,
		subscribers:     make(map[string][]chan *HealthEvent),
		logger:          logger,
	}
}

// Subscribe registers a subscriber for health events of a specific agent
func (hsb *HealthStatusBroadcaster) Subscribe(agentID string) <-chan *HealthEvent {
	hsb.mu.Lock()
	defer hsb.mu.Unlock()

	eventChan := make(chan *HealthEvent, 100) // Buffered channel
	hsb.subscribers[agentID] = append(hsb.subscribers[agentID], eventChan)

	return eventChan
}

// SubscribeAll registers a subscriber for all health events
func (hsb *HealthStatusBroadcaster) SubscribeAll() <-chan *HealthEvent {
	return hsb.Subscribe("*") // Use "*" for all agents
}

// BroadcastHealthEvent broadcasts a health event to all relevant subscribers
func (hsb *HealthStatusBroadcaster) BroadcastHealthEvent(healthEvent *HealthEvent) {
	hsb.mu.RLock()
	defer hsb.mu.RUnlock()

	// Broadcast to agent-specific subscribers
	if subscribers, exists := hsb.subscribers[healthEvent.AgentID]; exists {
		hsb.broadcastToSubscribers(subscribers, healthEvent)
	}

	// Broadcast to global subscribers
	if allSubscribers, exists := hsb.subscribers["*"]; exists {
		hsb.broadcastToSubscribers(allSubscribers, healthEvent)
	}
}

// broadcastToSubscribers sends the event to a list of subscribers
func (hsb *HealthStatusBroadcaster) broadcastToSubscribers(subscribers []chan *HealthEvent, event *HealthEvent) {
	for _, subscriber := range subscribers {
		select {
		case subscriber <- event:
			// Event sent successfully
		default:
			hsb.logger.Warn("Subscriber channel full, dropping health event")
		}
	}
}
