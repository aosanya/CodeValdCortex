package orchestration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// Engine implements the WorkflowEngine interface
type Engine struct {
	// Configuration
	config OrchestrationConfig

	// Dependencies
	coordinator AgentCoordinator
	monitor     ExecutionMonitor
	repository  WorkflowRepository

	// Runtime state
	activeExecutions map[string]*WorkflowExecution
	executionMutex   sync.RWMutex

	// Channels for coordination
	taskQueue       chan *TaskExecution
	completionQueue chan *TaskExecution

	// Context and cancellation
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Logger
	logger *log.Logger
}

// NewEngine creates a new workflow engine instance
func NewEngine(config OrchestrationConfig, coordinator AgentCoordinator, monitor ExecutionMonitor, repository WorkflowRepository, logger *log.Logger) *Engine {
	ctx, cancel := context.WithCancel(context.Background())

	return &Engine{
		config:           config,
		coordinator:      coordinator,
		monitor:          monitor,
		repository:       repository,
		activeExecutions: make(map[string]*WorkflowExecution),
		taskQueue:        make(chan *TaskExecution, 1000),
		completionQueue:  make(chan *TaskExecution, 1000),
		ctx:              ctx,
		cancel:           cancel,
		logger:           logger,
	}
}

// Start starts the workflow engine
func (e *Engine) Start() error {
	e.logger.Info("Starting workflow engine")

	// Start task processing workers
	for i := 0; i < 10; i++ { // Configurable worker count
		e.wg.Add(1)
		go e.taskProcessorWorker(i)
	}

	// Start completion processing worker
	e.wg.Add(1)
	go e.completionProcessorWorker()

	// Start execution monitoring
	e.wg.Add(1)
	go e.executionMonitorWorker()

	e.logger.Info("Workflow engine started successfully")
	return nil
}

// Stop stops the workflow engine
func (e *Engine) Stop() error {
	e.logger.Info("Stopping workflow engine")

	// Cancel context to signal workers to stop
	e.cancel()

	// Wait for all workers to finish
	e.wg.Wait()

	e.logger.Info("Workflow engine stopped successfully")
	return nil
}

// ExecuteWorkflow starts execution of a workflow
func (e *Engine) ExecuteWorkflow(ctx context.Context, workflow *Workflow) (*WorkflowExecution, error) {
	e.logger.WithField("workflow_id", workflow.ID).Info("Starting workflow execution")

	// Validate workflow
	if err := e.validateWorkflow(workflow); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// Create execution instance
	execution := &WorkflowExecution{
		ID:             uuid.New().String(),
		WorkflowID:     workflow.ID,
		Status:         WorkflowStatusPending,
		StartTime:      time.Now(),
		TaskExecutions: make(map[string]*TaskExecution),
		Context:        make(map[string]interface{}),
		AgentsUsed:     make([]string, 0),
		TriggeredBy:    "api", // Could be extracted from context
		Metrics: ExecutionMetrics{
			TotalTasks: len(workflow.Tasks),
		},
	}

	// Initialize task executions
	for _, task := range workflow.Tasks {
		execution.TaskExecutions[task.ID] = &TaskExecution{
			TaskID:        task.ID,
			Status:        TaskStatusPending,
			Attempts:      0,
			Output:        make(map[string]interface{}),
			Logs:          make([]string, 0),
			ResourceUsage: ResourceUsage{},
		}
	}

	// Store execution
	if err := e.repository.StoreExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to store execution: %w", err)
	}

	// Register execution for monitoring
	e.executionMutex.Lock()
	e.activeExecutions[execution.ID] = execution
	e.executionMutex.Unlock()

	// Start monitoring
	if err := e.monitor.StartMonitoring(ctx, execution); err != nil {
		e.logger.WithError(err).Error("Failed to start execution monitoring")
	}

	// Start execution asynchronously
	e.wg.Add(1)
	go e.executeWorkflowAsync(ctx, workflow, execution)

	e.logger.WithFields(log.Fields{
		"execution_id": execution.ID,
		"workflow_id":  workflow.ID,
	}).Info("Workflow execution started")

	return execution, nil
}

// executeWorkflowAsync handles the actual workflow execution
func (e *Engine) executeWorkflowAsync(ctx context.Context, workflow *Workflow, execution *WorkflowExecution) {
	defer e.wg.Done()

	e.logger.WithField("execution_id", execution.ID).Debug("Starting async workflow execution")

	// Update status to running
	execution.Status = WorkflowStatusRunning
	e.updateExecution(ctx, execution)

	// Build dependency graph
	depGraph, err := e.buildDependencyGraph(workflow)
	if err != nil {
		e.failExecution(ctx, execution, fmt.Errorf("failed to build dependency graph: %w", err))
		return
	}

	// Execute tasks in dependency order
	if err := e.executeTasks(ctx, workflow, execution, depGraph); err != nil {
		e.failExecution(ctx, execution, err)
		return
	}

	// Mark as completed
	e.completeExecution(ctx, execution)
}

// buildDependencyGraph creates a dependency graph for task execution order
func (e *Engine) buildDependencyGraph(workflow *Workflow) (*DependencyGraph, error) {
	graph := NewDependencyGraph()

	// Add all tasks as nodes
	for _, task := range workflow.Tasks {
		graph.AddNode(task.ID)
	}

	// Add dependencies as edges
	for taskID, deps := range workflow.Dependencies {
		for _, depID := range deps {
			if err := graph.AddEdge(depID, taskID); err != nil {
				return nil, fmt.Errorf("invalid dependency %s -> %s: %w", depID, taskID, err)
			}
		}
	}

	// Validate graph (check for cycles)
	if err := graph.ValidateAcyclic(); err != nil {
		return nil, fmt.Errorf("workflow contains circular dependencies: %w", err)
	}

	return graph, nil
}

// executeTasks executes workflow tasks respecting dependencies
func (e *Engine) executeTasks(ctx context.Context, workflow *Workflow, execution *WorkflowExecution, depGraph *DependencyGraph) error {
	taskMap := make(map[string]*WorkflowTask)
	for i := range workflow.Tasks {
		taskMap[workflow.Tasks[i].ID] = &workflow.Tasks[i]
	}

	// Get execution batches (tasks that can run in parallel)
	batches := depGraph.GetExecutionBatches()

	for batchIndex, batch := range batches {
		e.logger.WithFields(log.Fields{
			"execution_id": execution.ID,
			"batch_index":  batchIndex,
			"batch_size":   len(batch),
		}).Debug("Executing task batch")

		// Execute all tasks in this batch concurrently
		var batchWg sync.WaitGroup
		batchErrors := make(chan error, len(batch))

		for _, taskID := range batch {
			task, exists := taskMap[taskID]
			if !exists {
				return fmt.Errorf("task %s not found in workflow", taskID)
			}

			batchWg.Add(1)
			go func(t *WorkflowTask) {
				defer batchWg.Done()
				if err := e.executeTask(ctx, t, execution); err != nil {
					batchErrors <- err
				}
			}(task)
		}

		// Wait for batch completion
		batchWg.Wait()
		close(batchErrors)

		// Check for batch errors
		var batchErr error
		for err := range batchErrors {
			if err != nil {
				batchErr = err
				break
			}
		}

		if batchErr != nil {
			// Handle failure policy
			if e.shouldStopOnFailure(workflow, execution) {
				return batchErr
			}
		}

		// Check if context was cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	return nil
}

// executeTask executes a single workflow task
func (e *Engine) executeTask(ctx context.Context, task *WorkflowTask, execution *WorkflowExecution) error {
	taskExecution := execution.TaskExecutions[task.ID]

	e.logger.WithFields(log.Fields{
		"execution_id": execution.ID,
		"task_id":      task.ID,
		"task_type":    task.Type,
	}).Debug("Executing task")

	// Check task conditions
	if !e.shouldExecuteTask(task, execution) {
		taskExecution.Status = TaskStatusSkipped
		e.updateExecution(ctx, execution)
		return nil
	}

	// Update task status
	taskExecution.Status = TaskStatusQueued
	taskExecution.StartTime = time.Now()
	e.updateExecution(ctx, execution)

	// Select agent for task execution
	agents, err := e.coordinator.SelectAgents(ctx, task.AgentSelector, 1)
	if err != nil {
		return fmt.Errorf("failed to select agent for task %s: %w", task.ID, err)
	}

	if len(agents) == 0 {
		return fmt.Errorf("no available agents for task %s", task.ID)
	}

	selectedAgent := agents[0]
	taskExecution.AgentID = selectedAgent.ID

	// Add agent to execution agents list
	e.addAgentToExecution(execution, selectedAgent.ID)

	// Update task status to running
	taskExecution.Status = TaskStatusRunning
	e.updateExecution(ctx, execution)

	// Execute task with retry logic
	err = e.executeTaskWithRetry(ctx, task, taskExecution, selectedAgent, execution)

	// Update end time and duration
	now := time.Now()
	taskExecution.EndTime = &now
	taskExecution.Duration = now.Sub(taskExecution.StartTime)

	if err != nil {
		taskExecution.Status = TaskStatusFailed
		taskExecution.Error = err.Error()
	} else {
		taskExecution.Status = TaskStatusCompleted
	}

	e.updateExecution(ctx, execution)
	return err
}

// executeTaskWithRetry executes a task with retry logic
func (e *Engine) executeTaskWithRetry(ctx context.Context, task *WorkflowTask, taskExecution *TaskExecution, agent *agent.Agent, execution *WorkflowExecution) error {
	var lastError error

	maxAttempts := task.RetryPolicy.MaxAttempts
	if maxAttempts == 0 {
		maxAttempts = 1 // At least one attempt
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		taskExecution.Attempts = attempt

		// Create task context with timeout
		taskCtx, cancel := context.WithTimeout(ctx, task.Timeout)

		// Execute the task
		err := e.assignAndExecuteTask(taskCtx, task, taskExecution, agent, execution)
		cancel()

		if err == nil {
			// Task succeeded
			return nil
		}

		lastError = err
		taskExecution.Logs = append(taskExecution.Logs, fmt.Sprintf("Attempt %d failed: %s", attempt, err.Error()))

		// Check if we should retry
		if attempt < maxAttempts && e.shouldRetryTask(task, err) {
			// Calculate retry delay
			delay := e.calculateRetryDelay(task.RetryPolicy, attempt)

			e.logger.WithFields(log.Fields{
				"execution_id": execution.ID,
				"task_id":      task.ID,
				"attempt":      attempt,
				"delay":        delay,
			}).Debug("Retrying task after delay")

			// Update status to retrying
			taskExecution.Status = TaskStatusRetrying
			e.updateExecution(ctx, execution)

			// Wait for retry delay
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("task failed after %d attempts: %w", maxAttempts, lastError)
}

// assignAndExecuteTask assigns and executes a task on an agent
func (e *Engine) assignAndExecuteTask(ctx context.Context, task *WorkflowTask, taskExecution *TaskExecution, agent *agent.Agent, execution *WorkflowExecution) error {
	// Assign task to agent
	if err := e.coordinator.AssignTask(ctx, agent.ID, task, execution); err != nil {
		return fmt.Errorf("failed to assign task to agent: %w", err)
	}

	// For now, simulate task execution
	// In a real implementation, this would delegate to the agent's task system
	return e.simulateTaskExecution(ctx, task, taskExecution, agent, execution)
}

// simulateTaskExecution simulates task execution (placeholder for real implementation)
func (e *Engine) simulateTaskExecution(ctx context.Context, task *WorkflowTask, taskExecution *TaskExecution, agent *agent.Agent, execution *WorkflowExecution) error {
	// This is a simulation - in real implementation, this would:
	// 1. Convert WorkflowTask to agent.Task
	// 2. Submit to agent's task system
	// 3. Wait for completion
	// 4. Capture results and outputs

	e.logger.WithFields(log.Fields{
		"task_id":  task.ID,
		"agent_id": agent.ID,
		"type":     task.Type,
	}).Debug("Simulating task execution")

	// Simulate processing time
	select {
	case <-time.After(100 * time.Millisecond):
	case <-ctx.Done():
		return ctx.Err()
	}

	// Simulate task results
	taskExecution.Output["result"] = "success"
	taskExecution.Output["processed_at"] = time.Now()
	taskExecution.Logs = append(taskExecution.Logs, fmt.Sprintf("Task %s executed successfully on agent %s", task.ID, agent.ID))

	// Update resource usage
	taskExecution.ResourceUsage = ResourceUsage{
		CPU:       task.Resources.CPU,
		Memory:    task.Resources.Memory,
		NetworkIO: 1024, // Simulated
		DiskIO:    512,  // Simulated
	}

	return nil
}

// Helper methods

func (e *Engine) validateWorkflow(workflow *Workflow) error {
	if workflow.ID == "" {
		return fmt.Errorf("workflow ID is required")
	}

	if len(workflow.Tasks) == 0 {
		return fmt.Errorf("workflow must have at least one task")
	}

	// Validate task IDs are unique
	taskIDs := make(map[string]bool)
	for _, task := range workflow.Tasks {
		if task.ID == "" {
			return fmt.Errorf("task ID is required")
		}
		if taskIDs[task.ID] {
			return fmt.Errorf("duplicate task ID: %s", task.ID)
		}
		taskIDs[task.ID] = true
	}

	// Validate dependencies reference existing tasks
	for taskID, deps := range workflow.Dependencies {
		if !taskIDs[taskID] {
			return fmt.Errorf("dependency references unknown task: %s", taskID)
		}
		for _, depID := range deps {
			if !taskIDs[depID] {
				return fmt.Errorf("dependency references unknown task: %s", depID)
			}
		}
	}

	return nil
}

func (e *Engine) shouldExecuteTask(task *WorkflowTask, execution *WorkflowExecution) bool {
	// Evaluate task conditions
	for _, condition := range task.Conditions {
		if !e.evaluateCondition(condition, execution) {
			return false
		}
	}
	return true
}

func (e *Engine) evaluateCondition(condition TaskCondition, execution *WorkflowExecution) bool {
	// Simple condition evaluation - in real implementation would be more sophisticated
	switch condition.Type {
	case "always":
		return true
	case "never":
		return false
	case "context_equals":
		key, ok := condition.Parameters["key"].(string)
		if !ok {
			return false
		}
		expectedValue := condition.Parameters["value"]
		actualValue := execution.Context[key]
		return actualValue == expectedValue
	default:
		return true
	}
}

func (e *Engine) shouldStopOnFailure(workflow *Workflow, execution *WorkflowExecution) bool {
	policy := workflow.Configuration.FailurePolicy

	switch policy.OnTaskFailure {
	case "stop":
		return true
	case "continue":
		return false
	case "retry_workflow":
		// Could implement workflow retry logic here
		return true
	default:
		return true
	}
}

func (e *Engine) shouldRetryTask(task *WorkflowTask, err error) bool {
	// Check if error type is retryable
	errMsg := err.Error()
	for _, retryableError := range task.RetryPolicy.RetryableErrors {
		if retryableError == errMsg || retryableError == "*" {
			return true
		}
	}

	// Default retry behavior
	return len(task.RetryPolicy.RetryableErrors) == 0
}

func (e *Engine) calculateRetryDelay(policy RetryPolicy, attempt int) time.Duration {
	delay := policy.InitialDelay

	// Apply exponential backoff
	for i := 1; i < attempt; i++ {
		delay = time.Duration(float64(delay) * policy.BackoffMultiplier)
		if delay > policy.MaxDelay {
			delay = policy.MaxDelay
			break
		}
	}

	return delay
}

func (e *Engine) addAgentToExecution(execution *WorkflowExecution, agentID string) {
	for _, id := range execution.AgentsUsed {
		if id == agentID {
			return // Already in list
		}
	}
	execution.AgentsUsed = append(execution.AgentsUsed, agentID)
}

func (e *Engine) updateExecution(ctx context.Context, execution *WorkflowExecution) {
	if err := e.repository.UpdateExecution(ctx, execution); err != nil {
		e.logger.WithError(err).Error("Failed to update execution")
	}
}

func (e *Engine) failExecution(ctx context.Context, execution *WorkflowExecution, err error) {
	execution.Status = WorkflowStatusFailed
	execution.Error = err.Error()
	now := time.Now()
	execution.EndTime = &now
	execution.Duration = now.Sub(execution.StartTime)

	e.updateExecution(ctx, execution)

	// Remove from active executions
	e.executionMutex.Lock()
	delete(e.activeExecutions, execution.ID)
	e.executionMutex.Unlock()

	// Stop monitoring
	if err := e.monitor.StopMonitoring(ctx, execution.ID); err != nil {
		e.logger.WithError(err).Error("Failed to stop execution monitoring")
	}

	e.logger.WithField("execution_id", execution.ID).WithError(err).Error("Workflow execution failed")
}

func (e *Engine) completeExecution(ctx context.Context, execution *WorkflowExecution) {
	execution.Status = WorkflowStatusCompleted
	now := time.Now()
	execution.EndTime = &now
	execution.Duration = now.Sub(execution.StartTime)

	// Update metrics
	execution.Metrics.CompletedTasks = e.countTasksByStatus(execution, TaskStatusCompleted)
	execution.Metrics.FailedTasks = e.countTasksByStatus(execution, TaskStatusFailed)
	execution.Metrics.SkippedTasks = e.countTasksByStatus(execution, TaskStatusSkipped)
	execution.Metrics.AgentsUtilized = len(execution.AgentsUsed)

	e.updateExecution(ctx, execution)

	// Remove from active executions
	e.executionMutex.Lock()
	delete(e.activeExecutions, execution.ID)
	e.executionMutex.Unlock()

	// Stop monitoring
	if err := e.monitor.StopMonitoring(ctx, execution.ID); err != nil {
		e.logger.WithError(err).Error("Failed to stop execution monitoring")
	}

	e.logger.WithField("execution_id", execution.ID).Info("Workflow execution completed successfully")
}

func (e *Engine) countTasksByStatus(execution *WorkflowExecution, status TaskStatus) int {
	count := 0
	for _, taskExec := range execution.TaskExecutions {
		if taskExec.Status == status {
			count++
		}
	}
	return count
}

// Worker methods

func (e *Engine) taskProcessorWorker(workerID int) {
	defer e.wg.Done()

	e.logger.WithField("worker_id", workerID).Debug("Task processor worker started")

	for {
		select {
		case <-e.ctx.Done():
			return
		case taskExecution := <-e.taskQueue:
			// Process task execution
			e.logger.WithFields(log.Fields{
				"worker_id":    workerID,
				"task_id":      taskExecution.TaskID,
				"execution_id": "unknown", // Would need to pass execution context
			}).Debug("Processing task")

			// Task processing logic would go here
			// For now, this is handled directly in executeTask
		}
	}
}

func (e *Engine) completionProcessorWorker() {
	defer e.wg.Done()

	e.logger.Debug("Completion processor worker started")

	for {
		select {
		case <-e.ctx.Done():
			return
		case taskExecution := <-e.completionQueue:
			// Process task completion
			e.logger.WithField("task_id", taskExecution.TaskID).Debug("Processing task completion")

			// Completion processing logic would go here
		}
	}
}

func (e *Engine) executionMonitorWorker() {
	defer e.wg.Done()

	e.logger.Debug("Execution monitor worker started")

	ticker := time.NewTicker(10 * time.Second) // Monitor every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.monitorActiveExecutions()
		}
	}
}

func (e *Engine) monitorActiveExecutions() {
	e.executionMutex.RLock()
	executions := make([]*WorkflowExecution, 0, len(e.activeExecutions))
	for _, execution := range e.activeExecutions {
		executions = append(executions, execution)
	}
	e.executionMutex.RUnlock()

	for _, execution := range executions {
		// Check for timeouts, stuck tasks, etc.
		e.checkExecutionHealth(execution)
	}
}

func (e *Engine) checkExecutionHealth(execution *WorkflowExecution) {
	// Check for execution timeout
	if e.config.DefaultWorkflowTimeout > 0 {
		if time.Since(execution.StartTime) > e.config.DefaultWorkflowTimeout {
			e.logger.WithField("execution_id", execution.ID).Warn("Workflow execution timeout")
			// Could implement timeout handling here
		}
	}

	// Check for stuck tasks
	for taskID, taskExec := range execution.TaskExecutions {
		if taskExec.Status == TaskStatusRunning {
			// Check task timeout
			if time.Since(taskExec.StartTime) > 5*time.Minute { // Configurable
				e.logger.WithFields(log.Fields{
					"execution_id": execution.ID,
					"task_id":      taskID,
				}).Warn("Task execution appears stuck")
			}
		}
	}
}

// Interface implementation methods

func (e *Engine) GetExecution(ctx context.Context, executionID string) (*WorkflowExecution, error) {
	return e.repository.GetExecution(ctx, executionID)
}

func (e *Engine) ListExecutions(ctx context.Context, filters ExecutionFilters) ([]*WorkflowExecution, error) {
	// For now, return empty list - would implement repository query
	return []*WorkflowExecution{}, nil
}

func (e *Engine) CancelExecution(ctx context.Context, executionID string) error {
	e.executionMutex.Lock()
	execution, exists := e.activeExecutions[executionID]
	if !exists {
		e.executionMutex.Unlock()
		return fmt.Errorf("execution not found: %s", executionID)
	}

	execution.Status = WorkflowStatusCancelled
	delete(e.activeExecutions, executionID)
	e.executionMutex.Unlock()

	now := time.Now()
	execution.EndTime = &now
	execution.Duration = now.Sub(execution.StartTime)

	e.updateExecution(ctx, execution)

	e.logger.WithField("execution_id", executionID).Info("Workflow execution cancelled")
	return nil
}

func (e *Engine) PauseExecution(ctx context.Context, executionID string) error {
	e.executionMutex.Lock()
	execution, exists := e.activeExecutions[executionID]
	if !exists {
		e.executionMutex.Unlock()
		return fmt.Errorf("execution not found: %s", executionID)
	}

	execution.Status = WorkflowStatusPaused
	e.executionMutex.Unlock()

	e.updateExecution(ctx, execution)

	e.logger.WithField("execution_id", executionID).Info("Workflow execution paused")
	return nil
}

func (e *Engine) ResumeExecution(ctx context.Context, executionID string) error {
	e.executionMutex.Lock()
	execution, exists := e.activeExecutions[executionID]
	if !exists {
		e.executionMutex.Unlock()
		return fmt.Errorf("execution not found: %s", executionID)
	}

	execution.Status = WorkflowStatusRunning
	e.executionMutex.Unlock()

	e.updateExecution(ctx, execution)

	e.logger.WithField("execution_id", executionID).Info("Workflow execution resumed")
	return nil
}

func (e *Engine) RetryExecution(ctx context.Context, executionID string) (*WorkflowExecution, error) {
	// Get original execution
	originalExecution, err := e.repository.GetExecution(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original execution: %w", err)
	}

	// Get workflow definition
	workflow, err := e.repository.GetWorkflow(ctx, originalExecution.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	// Start new execution
	return e.ExecuteWorkflow(ctx, workflow)
}
