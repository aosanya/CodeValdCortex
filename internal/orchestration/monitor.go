package orchestration

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Monitor implements the ExecutionMonitor interface
type Monitor struct {
	// Configuration
	config MonitorConfig

	// Runtime state
	executions map[string]*WorkflowExecution
	execMutex  sync.RWMutex

	// Task tracking
	taskMetrics  map[string]*TaskMetrics
	metricsMutex sync.RWMutex

	// Event handling
	eventHandlers []ExecutionEventHandler
	handlerMutex  sync.RWMutex

	// Progress tracking (add fields not in WorkflowExecution)
	progressData  map[string]float64
	progressMutex sync.RWMutex

	// Context and cancellation
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Logger
	logger *log.Logger
}

// MonitorConfig configures the execution monitor
type MonitorConfig struct {
	// MetricsRetentionPeriod determines how long to keep task metrics
	MetricsRetentionPeriod time.Duration

	// ProgressUpdateInterval for periodic progress updates
	ProgressUpdateInterval time.Duration

	// EnableDetailedMetrics collects additional performance metrics
	EnableDetailedMetrics bool

	// MetricsCleanupInterval for cleaning up old metrics
	MetricsCleanupInterval time.Duration

	// MaxEventHandlers limits the number of concurrent event handlers
	MaxEventHandlers int
}

// DefaultMonitorConfig returns default monitor configuration
func DefaultMonitorConfig() MonitorConfig {
	return MonitorConfig{
		MetricsRetentionPeriod: 24 * time.Hour, // 24 hours
		ProgressUpdateInterval: 5 * time.Second,
		EnableDetailedMetrics:  true,
		MetricsCleanupInterval: 1 * time.Hour,
		MaxEventHandlers:       10,
	}
}

// TaskMetrics holds performance metrics for a task
type TaskMetrics struct {
	TaskID      string
	ExecutionID string
	AgentID     string
	StartTime   time.Time
	EndTime     *time.Time
	Duration    time.Duration
	MemoryUsage float64 // MB
	CPUUsage    float64 // percentage
	Status      TaskStatus
	ErrorCount  int
	RetryCount  int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ExecutionEventType defines types of execution events
type ExecutionEventType string

const (
	EventExecutionStarted   ExecutionEventType = "execution_started"
	EventExecutionCompleted ExecutionEventType = "execution_completed"
	EventExecutionFailed    ExecutionEventType = "execution_failed"
	EventExecutionCancelled ExecutionEventType = "execution_cancelled"
	EventTaskStarted        ExecutionEventType = "task_started"
	EventTaskCompleted      ExecutionEventType = "task_completed"
	EventTaskFailed         ExecutionEventType = "task_failed"
	EventTaskRetried        ExecutionEventType = "task_retried"
	EventAgentAssigned      ExecutionEventType = "agent_assigned"
	EventAgentReleased      ExecutionEventType = "agent_released"
	EventProgressUpdate     ExecutionEventType = "progress_update"
)

// ExecutionEventHandler processes execution events
type ExecutionEventHandler interface {
	HandleEvent(ctx context.Context, event *ExecutionEvent) error
}

// NewMonitor creates a new execution monitor
func NewMonitor(config MonitorConfig, logger *log.Logger) *Monitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &Monitor{
		config:        config,
		executions:    make(map[string]*WorkflowExecution),
		taskMetrics:   make(map[string]*TaskMetrics),
		eventHandlers: make([]ExecutionEventHandler, 0),
		progressData:  make(map[string]float64),
		ctx:           ctx,
		cancel:        cancel,
		logger:        logger,
	}
}

// Start starts the execution monitor
func (m *Monitor) Start() error {
	m.logger.Info("Starting execution monitor")

	// Start progress monitoring
	m.wg.Add(1)
	go m.progressMonitorWorker()

	// Start metrics cleanup
	m.wg.Add(1)
	go m.metricsCleanupWorker()

	m.logger.Info("Execution monitor started successfully")
	return nil
}

// Stop stops the execution monitor
func (m *Monitor) Stop() error {
	m.logger.Info("Stopping execution monitor")

	// Cancel context to signal workers to stop
	m.cancel()

	// Wait for all workers to finish
	m.wg.Wait()

	m.logger.Info("Execution monitor stopped successfully")
	return nil
}

// StartTracking begins tracking a workflow execution
func (m *Monitor) StartTracking(ctx context.Context, execution *WorkflowExecution) error {
	m.logger.WithFields(log.Fields{
		"execution_id": execution.ID,
		"workflow_id":  execution.WorkflowID,
	}).Debug("Starting execution tracking")

	m.execMutex.Lock()
	m.executions[execution.ID] = execution
	m.execMutex.Unlock()

	// Initialize progress
	m.progressMutex.Lock()
	m.progressData[execution.ID] = 0.0
	m.progressMutex.Unlock()

	// Emit execution started event
	event := &ExecutionEvent{
		ExecutionID: execution.ID,
		Type:        string(EventExecutionStarted),
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"workflow_id": execution.WorkflowID,
			"total_tasks": len(execution.TaskExecutions),
			"message":     "Workflow execution started",
		},
	}

	if err := m.emitEvent(ctx, event); err != nil {
		m.logger.WithError(err).Warn("Failed to emit execution started event")
	}

	return nil
}

// StopTracking stops tracking a workflow execution
func (m *Monitor) StopTracking(ctx context.Context, executionID string) error {
	m.logger.WithField("execution_id", executionID).Debug("Stopping execution tracking")

	m.execMutex.Lock()
	execution, exists := m.executions[executionID]
	if exists {
		delete(m.executions, executionID)
	}
	m.execMutex.Unlock()

	// Clean up progress data
	m.progressMutex.Lock()
	delete(m.progressData, executionID)
	m.progressMutex.Unlock()

	if !exists {
		return nil
	}

	// Emit execution completed/failed event based on status
	eventType := EventExecutionCompleted
	message := "Workflow execution completed"

	if execution.Status == WorkflowStatusFailed {
		eventType = EventExecutionFailed
		message = "Workflow execution failed"
	} else if execution.Status == WorkflowStatusCancelled {
		eventType = EventExecutionCancelled
		message = "Workflow execution cancelled"
	}

	var duration time.Duration
	if execution.EndTime != nil {
		duration = execution.EndTime.Sub(execution.StartTime)
	}

	event := &ExecutionEvent{
		ExecutionID: executionID,
		Type:        string(eventType),
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"status":          string(execution.Status),
			"completion_time": execution.EndTime,
			"duration":        duration.Seconds(),
			"message":         message,
		},
	}

	if err := m.emitEvent(ctx, event); err != nil {
		m.logger.WithError(err).Warn("Failed to emit execution stopped event")
	}

	return nil
}

// UpdateProgress updates the execution progress
func (m *Monitor) UpdateProgress(ctx context.Context, executionID string, progress float64) error {
	m.execMutex.RLock()
	execution, exists := m.executions[executionID]
	m.execMutex.RUnlock()

	if !exists {
		return nil // Execution not being tracked
	}

	// Update progress in our tracking
	m.progressMutex.Lock()
	m.progressData[executionID] = progress
	m.progressMutex.Unlock()

	// Emit progress update event
	event := &ExecutionEvent{
		ExecutionID: executionID,
		Type:        string(EventProgressUpdate),
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"progress":        progress,
			"completed_tasks": m.countCompletedTasks(execution),
			"total_tasks":     len(execution.TaskExecutions),
			"message":         "Execution progress updated",
		},
	}

	if err := m.emitEvent(ctx, event); err != nil {
		m.logger.WithError(err).Warn("Failed to emit progress update event")
	}

	m.logger.WithFields(log.Fields{
		"execution_id": executionID,
		"progress":     progress,
	}).Debug("Execution progress updated")

	return nil
}

// RecordTaskStart records the start of a task execution
func (m *Monitor) RecordTaskStart(ctx context.Context, executionID, taskID, agentID string) error {
	m.logger.WithFields(log.Fields{
		"execution_id": executionID,
		"task_id":      taskID,
		"agent_id":     agentID,
	}).Debug("Recording task start")

	// Create task metrics
	metrics := &TaskMetrics{
		TaskID:      taskID,
		ExecutionID: executionID,
		AgentID:     agentID,
		StartTime:   time.Now(),
		Status:      TaskStatusRunning,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	m.metricsMutex.Lock()
	m.taskMetrics[taskID] = metrics
	m.metricsMutex.Unlock()

	// Emit task started event
	event := &ExecutionEvent{
		ExecutionID: executionID,
		Type:        string(EventTaskStarted),
		TaskID:      taskID,
		AgentID:     agentID,
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"message": "Task execution started",
		},
	}

	return m.emitEvent(ctx, event)
}

// RecordTaskCompletion records the completion of a task execution
func (m *Monitor) RecordTaskCompletion(ctx context.Context, taskID string, status TaskStatus, error string) error {
	m.logger.WithFields(log.Fields{
		"task_id": taskID,
		"status":  status,
	}).Debug("Recording task completion")

	m.metricsMutex.Lock()
	metrics, exists := m.taskMetrics[taskID]
	if exists {
		endTime := time.Now()
		metrics.EndTime = &endTime
		metrics.Duration = endTime.Sub(metrics.StartTime)
		metrics.Status = status
		metrics.UpdatedAt = time.Now()

		if status == TaskStatusFailed {
			metrics.ErrorCount++
		}
	}
	m.metricsMutex.Unlock()

	if !exists {
		m.logger.WithField("task_id", taskID).Warn("Task metrics not found for completion")
		return nil
	}

	// Determine event type and message
	eventType := EventTaskCompleted
	message := "Task execution completed"

	if status == TaskStatusFailed {
		eventType = EventTaskFailed
		message = "Task execution failed"
	}

	// Emit task completion event
	event := &ExecutionEvent{
		ExecutionID: metrics.ExecutionID,
		Type:        string(eventType),
		TaskID:      taskID,
		AgentID:     metrics.AgentID,
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"duration":    metrics.Duration.Seconds(),
			"status":      string(status),
			"error":       error,
			"retry_count": metrics.RetryCount,
			"message":     message,
		},
	}

	return m.emitEvent(ctx, event)
}

// RecordTaskRetry records a task retry
func (m *Monitor) RecordTaskRetry(ctx context.Context, taskID string, retryCount int, reason string) error {
	m.logger.WithFields(log.Fields{
		"task_id":     taskID,
		"retry_count": retryCount,
		"reason":      reason,
	}).Debug("Recording task retry")

	m.metricsMutex.Lock()
	metrics, exists := m.taskMetrics[taskID]
	if exists {
		metrics.RetryCount = retryCount
		metrics.UpdatedAt = time.Now()
	}
	m.metricsMutex.Unlock()

	if !exists {
		m.logger.WithField("task_id", taskID).Warn("Task metrics not found for retry")
		return nil
	}

	// Emit task retry event
	event := &ExecutionEvent{
		ExecutionID: metrics.ExecutionID,
		Type:        string(EventTaskRetried),
		TaskID:      taskID,
		AgentID:     metrics.AgentID,
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"retry_count": retryCount,
			"reason":      reason,
			"message":     "Task execution retry",
		},
	}

	return m.emitEvent(ctx, event)
}

// GetExecutionStatus returns the current status of an execution
func (m *Monitor) GetExecutionStatus(ctx context.Context, executionID string) (*WorkflowExecution, error) {
	m.execMutex.RLock()
	execution, exists := m.executions[executionID]
	m.execMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("execution %s not found", executionID)
	}

	// Return a copy to avoid concurrent modification
	executionCopy := *execution
	return &executionCopy, nil
}

// GetTaskMetrics returns metrics for a specific task
func (m *Monitor) GetTaskMetrics(ctx context.Context, taskID string) (*TaskMetrics, error) {
	m.metricsMutex.RLock()
	metrics, exists := m.taskMetrics[taskID]
	m.metricsMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("task metrics %s not found", taskID)
	}

	// Return a copy to avoid concurrent modification
	metricsCopy := *metrics
	return &metricsCopy, nil
}

// GetExecutionMetrics returns aggregated metrics for an execution
func (m *Monitor) GetExecutionMetrics(ctx context.Context, executionID string) (*ExecutionMetrics, error) {
	m.execMutex.RLock()
	execution, exists := m.executions[executionID]
	m.execMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("execution %s not found", executionID)
	}

	// Collect task metrics
	var totalDuration time.Duration
	var avgMemoryUsage, avgCPUUsage float64
	var completedTasks, failedTasks, retriedTasks int

	m.metricsMutex.RLock()
	for taskID := range execution.TaskExecutions {
		if metrics, exists := m.taskMetrics[taskID]; exists {
			totalDuration += metrics.Duration
			avgMemoryUsage += metrics.MemoryUsage
			avgCPUUsage += metrics.CPUUsage

			if metrics.Status == TaskStatusCompleted {
				completedTasks++
			} else if metrics.Status == TaskStatusFailed {
				failedTasks++
			}

			if metrics.RetryCount > 0 {
				retriedTasks++
			}
		}
	}
	m.metricsMutex.RUnlock()

	taskCount := len(execution.TaskExecutions)
	var avgTaskDuration time.Duration
	if taskCount > 0 {
		avgMemoryUsage /= float64(taskCount)
		avgCPUUsage /= float64(taskCount)
		avgTaskDuration = totalDuration / time.Duration(taskCount)
	}

	// Use the actual ExecutionMetrics fields from types.go
	metrics := &ExecutionMetrics{
		TotalTasks:          taskCount,
		CompletedTasks:      completedTasks,
		FailedTasks:         failedTasks,
		SkippedTasks:        0, // Calculate if needed
		AverageTaskDuration: avgTaskDuration,
		MaxConcurrentTasks:  0, // Calculate if needed
		AgentsUtilized:      len(execution.AgentsUsed),
		TotalResourceUsage: ResourceUsage{
			CPU:       int(avgCPUUsage),
			Memory:    int(avgMemoryUsage),
			NetworkIO: 0, // Calculate if needed
			DiskIO:    0, // Calculate if needed
		},
	}

	return metrics, nil
}

// AddEventHandler adds an event handler
func (m *Monitor) AddEventHandler(handler ExecutionEventHandler) error {
	m.handlerMutex.Lock()
	defer m.handlerMutex.Unlock()

	if len(m.eventHandlers) >= m.config.MaxEventHandlers {
		return fmt.Errorf("maximum number of event handlers (%d) reached", m.config.MaxEventHandlers)
	}

	m.eventHandlers = append(m.eventHandlers, handler)
	m.logger.WithField("handler_count", len(m.eventHandlers)).Debug("Event handler added")

	return nil
}

// Helper methods

func (m *Monitor) emitEvent(ctx context.Context, event *ExecutionEvent) error {
	m.handlerMutex.RLock()
	handlers := make([]ExecutionEventHandler, len(m.eventHandlers))
	copy(handlers, m.eventHandlers)
	m.handlerMutex.RUnlock()

	// Process events asynchronously to avoid blocking
	for _, handler := range handlers {
		go func(h ExecutionEventHandler) {
			if err := h.HandleEvent(ctx, event); err != nil {
				m.logger.WithError(err).WithField("event_type", event.Type).Warn("Event handler failed")
			}
		}(handler)
	}

	return nil
}

func (m *Monitor) countCompletedTasks(execution *WorkflowExecution) int {
	count := 0
	m.metricsMutex.RLock()
	for taskID := range execution.TaskExecutions {
		if metrics, exists := m.taskMetrics[taskID]; exists && metrics.Status == TaskStatusCompleted {
			count++
		}
	}
	m.metricsMutex.RUnlock()
	return count
}

// Worker methods

func (m *Monitor) progressMonitorWorker() {
	defer m.wg.Done()

	m.logger.Debug("Progress monitor worker started")

	ticker := time.NewTicker(m.config.ProgressUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.updateAllProgress()
		}
	}
}

func (m *Monitor) updateAllProgress() {
	m.execMutex.RLock()
	executions := make([]*WorkflowExecution, 0, len(m.executions))
	for _, execution := range m.executions {
		if execution.Status == WorkflowStatusRunning {
			executions = append(executions, execution)
		}
	}
	m.execMutex.RUnlock()

	for _, execution := range executions {
		progress := m.calculateProgress(execution)
		if err := m.UpdateProgress(context.Background(), execution.ID, progress); err != nil {
			m.logger.WithError(err).WithField("execution_id", execution.ID).Warn("Failed to update progress")
		}
	}
}

func (m *Monitor) calculateProgress(execution *WorkflowExecution) float64 {
	if len(execution.TaskExecutions) == 0 {
		return 0.0
	}

	completedTasks := m.countCompletedTasks(execution)
	return float64(completedTasks) / float64(len(execution.TaskExecutions))
}

func (m *Monitor) metricsCleanupWorker() {
	defer m.wg.Done()

	m.logger.Debug("Metrics cleanup worker started")

	ticker := time.NewTicker(m.config.MetricsCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.cleanupOldMetrics()
		}
	}
}

func (m *Monitor) cleanupOldMetrics() {
	cutoff := time.Now().Add(-m.config.MetricsRetentionPeriod)

	m.metricsMutex.Lock()
	toDelete := make([]string, 0)
	for taskID, metrics := range m.taskMetrics {
		if metrics.CreatedAt.Before(cutoff) {
			toDelete = append(toDelete, taskID)
		}
	}

	for _, taskID := range toDelete {
		delete(m.taskMetrics, taskID)
	}
	m.metricsMutex.Unlock()

	if len(toDelete) > 0 {
		m.logger.WithField("cleaned_metrics", len(toDelete)).Debug("Cleaned up old task metrics")
	}
}
