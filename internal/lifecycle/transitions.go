package lifecycle

import (
	"fmt"

	"github.com/aosanya/CodeValdCortex/internal/agent"
)

// StateTransitionError represents an invalid state transition
type StateTransitionError struct {
	From agent.State
	To   agent.State
}

func (e *StateTransitionError) Error() string {
	return fmt.Sprintf("invalid state transition from %s to %s", e.From, e.To)
}

// validateTransition checks if a state transition is valid
func (m *Manager) validateTransition(from, to agent.State) error {
	// Define valid transitions
	validTransitions := map[agent.State][]agent.State{
		// From empty state (agent doesn't exist yet)
		agent.State(""): {agent.StateCreated},

		// From created state
		agent.StateCreated: {agent.StateRunning, agent.StateStopped},

		// From running state
		agent.StateRunning: {agent.StatePaused, agent.StateStopped, agent.StateFailed},

		// From paused state
		agent.StatePaused: {agent.StateRunning, agent.StateStopped},

		// From stopped state
		agent.StateStopped: {agent.StateRunning},

		// From failed state
		agent.StateFailed: {agent.StateRunning, agent.StateStopped},
	}

	// Check if transition is allowed
	allowedStates, exists := validTransitions[from]
	if !exists {
		return &StateTransitionError{From: from, To: to}
	}

	for _, allowed := range allowedStates {
		if allowed == to {
			return nil
		}
	}

	return &StateTransitionError{From: from, To: to}
}
