package agent

import "errors"

var (
	// ErrAgentStopped is returned when attempting operations on a stopped agent
	ErrAgentStopped = errors.New("agent is stopped")

	// ErrTaskQueueFull is returned when task queue is at capacity
	ErrTaskQueueFull = errors.New("task queue is full")

	// ErrInvalidState is returned for invalid state transitions
	ErrInvalidState = errors.New("invalid state transition")

	// ErrAgentNotFound is returned when agent ID doesn't exist
	ErrAgentNotFound = errors.New("agent not found")

	// ErrTaskTimeout is returned when task execution exceeds timeout
	ErrTaskTimeout = errors.New("task execution timeout")
)
