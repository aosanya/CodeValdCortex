package events

import (
	"fmt"
	"sync"
)

// HandlerRegistry manages event handler registration and lookup
type HandlerRegistry struct {
	mu       sync.RWMutex
	handlers map[EventType][]EventHandler
	global   []EventHandler // Global handlers that process all events
}

// NewHandlerRegistry creates a new handler registry
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers: make(map[EventType][]EventHandler),
		global:   make([]EventHandler, 0),
	}
}

// RegisterHandler registers a handler for specific event types
func (r *HandlerRegistry) RegisterHandler(handler EventHandler, eventTypes ...EventType) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if len(eventTypes) == 0 {
		// Register as global handler if no specific types provided
		r.global = append(r.global, handler)
		return nil
	}

	// Register for specific event types
	for _, eventType := range eventTypes {
		if r.handlers[eventType] == nil {
			r.handlers[eventType] = make([]EventHandler, 0)
		}

		// Check if handler is already registered for this event type
		for _, existing := range r.handlers[eventType] {
			if existing == handler {
				return fmt.Errorf("handler already registered for event type %s", eventType)
			}
		}

		r.handlers[eventType] = append(r.handlers[eventType], handler)
	}

	return nil
}

// UnregisterHandler removes a handler from all event types
func (r *HandlerRegistry) UnregisterHandler(handler EventHandler) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	removed := false

	// Remove from global handlers
	for i, h := range r.global {
		if h == handler {
			r.global = append(r.global[:i], r.global[i+1:]...)
			removed = true
			break
		}
	}

	// Remove from specific event type handlers
	for eventType, handlers := range r.handlers {
		for i, h := range handlers {
			if h == handler {
				r.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
				removed = true
				break
			}
		}
	}

	if !removed {
		return fmt.Errorf("handler not found in registry")
	}

	return nil
}

// GetHandlers returns all handlers that can process the given event type
func (r *HandlerRegistry) GetHandlers(eventType EventType) []EventHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []EventHandler

	// Add global handlers
	for _, handler := range r.global {
		if handler.CanHandle(eventType) {
			result = append(result, handler)
		}
	}

	// Add specific handlers for this event type
	if handlers, exists := r.handlers[eventType]; exists {
		for _, handler := range handlers {
			if handler.CanHandle(eventType) {
				result = append(result, handler)
			}
		}
	}

	return result
}

// GetHandlerCount returns the total number of registered handlers
func (r *HandlerRegistry) GetHandlerCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := len(r.global)
	for _, handlers := range r.handlers {
		count += len(handlers)
	}
	return count
}

// GetEventTypes returns all event types that have registered handlers
func (r *HandlerRegistry) GetEventTypes() []EventType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var eventTypes []EventType
	for eventType := range r.handlers {
		eventTypes = append(eventTypes, eventType)
	}
	return eventTypes
}

// Clear removes all registered handlers
func (r *HandlerRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.handlers = make(map[EventType][]EventHandler)
	r.global = make([]EventHandler, 0)
}
