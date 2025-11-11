package lifecycle

import (
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	log "github.com/sirupsen/logrus"
)

// startAgentRuntime starts the agent's goroutine for task processing
func (m *Manager) startAgentRuntime(a *agent.Agent) error {
	// Start heartbeat goroutine
	go m.heartbeatLoop(a)

	// Start task processing goroutine
	go m.taskProcessingLoop(a)

	return nil
}

// stopAgentRuntime gracefully stops the agent's goroutines
func (m *Manager) stopAgentRuntime(a *agent.Agent) error {
	// Cancel the agent's context to signal shutdown
	a.Cancel()

	// Wait for goroutines to finish
	<-a.Done()
	return nil
}

// pauseAgentRuntime temporarily stops task processing
func (m *Manager) pauseAgentRuntime(_ *agent.Agent) error {
	// For now, pausing just changes state - tasks won't be processed
	// The task processing loop checks the state before processing
	return nil
}

// resumeAgentRuntime resumes task processing
func (m *Manager) resumeAgentRuntime(_ *agent.Agent) error {
	// State change is handled by caller - this is a no-op for now
	return nil
}

// heartbeatLoop periodically updates the agent's heartbeat
func (m *Manager) heartbeatLoop(a *agent.Agent) {
	ticker := NewTicker(a.Config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C():
			// Update heartbeat
			a.UpdateHeartbeat()

			// Log heartbeat
			log.WithFields(log.Fields{
				"agent_id": a.ID,
				"state":    a.GetState(),
			}).Trace("Agent heartbeat")

		case <-a.Context().Done():
			return
		}
	}
}

// taskProcessingLoop processes tasks from the agent's task channel
func (m *Manager) taskProcessingLoop(a *agent.Agent) {
	defer a.CloseDone()

	for {
		select {
		case task := <-a.TaskChan():
			// Only process if agent is running
			if a.GetState() != agent.StateRunning {
				log.WithFields(log.Fields{
					"agent_id": a.ID,
					"task_id":  task.ID,
					"state":    a.GetState(),
				}).Warn("Skipping task - agent not running")
				continue
			}

			// Process the task
			if err := m.processTask(a, task); err != nil {
				log.WithFields(log.Fields{
					"agent_id": a.ID,
					"task_id":  task.ID,
					"error":    err,
				}).Error("Task processing failed")
			}

		case <-a.Context().Done():
			return
		}
	}
}

// processTask executes a single task
func (m *Manager) processTask(a *agent.Agent, task agent.Task) error {
	log.WithFields(log.Fields{
		"agent_id":  a.ID,
		"task_id":   task.ID,
		"task_type": task.Type,
	}).Info("Processing task")

	// TODO: Actual task execution logic will be implemented in MVP-007
	// For now, just log the task

	return nil
}

// Ticker interface for testing
type Ticker interface {
	C() <-chan time.Time
	Stop()
}

// realTicker wraps time.Ticker
type realTicker struct {
	*time.Ticker
}

func (rt *realTicker) C() <-chan time.Time {
	return rt.Ticker.C
}

// NewTicker creates a new ticker
func NewTicker(d time.Duration) Ticker {
	return &realTicker{time.NewTicker(d)}
}
