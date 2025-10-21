package events

import (
	"testing"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEventProcessor_BasicOperations tests basic processor functionality
func TestEventProcessor_BasicOperations(t *testing.T) {
	// Create processor with default config
	config := DefaultProcessorConfig()
	config.WorkerCount = 2
	config.QueueSize = 10

	processor := NewProcessor(config)

	// Test starting processor
	err := processor.Start()
	require.NoError(t, err)

	// Test registering a handler
	handler := NewLoggingHandler()
	err = processor.RegisterHandler(handler)
	require.NoError(t, err)

	// Test publishing an event
	event := &Event{
		Type:     EventTypeAgentCreated,
		AgentID:  "test-agent-1",
		Priority: PriorityNormal,
		Data: &AgentEventData{
			Agent: &agent.Agent{
				ID: "test-agent-1",
			},
		},
	}

	err = processor.PublishEvent(event)
	require.NoError(t, err)

	// Wait a bit for processing
	time.Sleep(100 * time.Millisecond)

	// Check metrics
	metrics := processor.GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalEvents)

	// Test stopping processor
	err = processor.Stop()
	require.NoError(t, err)
}

// TestHandlerRegistry_BasicOperations tests handler registry functionality
func TestHandlerRegistry_BasicOperations(t *testing.T) {
	registry := NewHandlerRegistry()

	// Create test handlers
	loggingHandler := NewLoggingHandler()
	stateHandler := NewStateChangeHandler()

	// Test registering handlers
	err := registry.RegisterHandler(loggingHandler)
	require.NoError(t, err)

	err = registry.RegisterHandler(stateHandler)
	require.NoError(t, err)

	// Test getting handlers for event type
	handlers := registry.GetHandlers(EventTypeAgentCreated)
	assert.Len(t, handlers, 2) // Both handlers should match

	// Test unregistering handler
	err = registry.UnregisterHandler(loggingHandler)
	require.NoError(t, err)

	handlers = registry.GetHandlers(EventTypeAgentCreated)
	assert.Len(t, handlers, 1) // Only state handler should remain
}

// TestLoggingHandler tests logging event handling
func TestLoggingHandler(t *testing.T) {
	handler := NewLoggingHandler()

	// Test handler can handle all events
	assert.True(t, handler.CanHandle(EventTypeAgentCreated))
	assert.True(t, handler.CanHandle(EventTypeMessageReceived))
	assert.True(t, handler.CanHandle(EventTypeTaskCompleted))

	// Test handler properties
	assert.Equal(t, "logging_handler", handler.Name())
	assert.Equal(t, 1, handler.Priority())
}
