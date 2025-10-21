package task

import (
	"container/heap"
	"context"
	"errors"
	"sync"
	"time"
)

var (
	// ErrSchedulerStopped is returned when scheduler is not running
	ErrSchedulerStopped = errors.New("scheduler is stopped")
	// ErrTaskNotFound is returned when task is not in queue
	ErrTaskNotFound = errors.New("task not found in queue")
	// ErrQueueFull is returned when queue is at capacity
	ErrQueueFull = errors.New("task queue is full")
)

// priorityQueue implements heap.Interface for task scheduling
type priorityQueue []*Task

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	// Higher priority comes first
	if pq[i].Priority != pq[j].Priority {
		return pq[i].Priority > pq[j].Priority
	}
	// Same priority: FIFO (earlier created comes first)
	return pq[i].CreatedAt.Before(pq[j].CreatedAt)
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *priorityQueue) Push(x interface{}) {
	task := x.(*Task)
	*pq = append(*pq, task)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	task := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return task
}

// Scheduler implements task scheduling with priority queue
type Scheduler struct {
	queue  priorityQueue
	queueM sync.RWMutex
	tasks  map[string]*Task // taskID -> task for quick lookup
	config WorkerPoolConfig

	// Worker pool
	workers      []*worker
	taskChan     chan *Task
	resultChan   chan *TaskResult
	shutdownChan chan struct{}
	wg           sync.WaitGroup

	executor TaskExecutor
	repo     TaskRepository

	started bool
	mu      sync.RWMutex
}

// worker represents a task execution worker
type worker struct {
	id       int
	taskChan <-chan *Task
	executor TaskExecutor
	wg       *sync.WaitGroup
	shutdown <-chan struct{}
}

// NewScheduler creates a new task scheduler
func NewScheduler(config WorkerPoolConfig, executor TaskExecutor, repo TaskRepository) *Scheduler {
	if config.MaxWorkers <= 0 {
		config.MaxWorkers = 10
	}
	if config.MinWorkers <= 0 {
		config.MinWorkers = 1
	}
	if config.QueueSize <= 0 {
		config.QueueSize = 100
	}
	if config.IdleTimeout <= 0 {
		config.IdleTimeout = 5 * time.Minute
	}

	s := &Scheduler{
		queue:        make(priorityQueue, 0),
		tasks:        make(map[string]*Task),
		config:       config,
		taskChan:     make(chan *Task, config.QueueSize),
		resultChan:   make(chan *TaskResult, config.MaxWorkers),
		shutdownChan: make(chan struct{}),
		executor:     executor,
		repo:         repo,
	}

	heap.Init(&s.queue)
	return s
}

// Start starts the scheduler and worker pool
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return nil
	}

	// Start minimum workers
	s.workers = make([]*worker, 0, s.config.MaxWorkers)
	for i := 0; i < s.config.MinWorkers; i++ {
		s.startWorker(i)
	}

	// Start dispatcher
	s.wg.Add(1)
	go s.dispatcher()

	// Start result processor
	s.wg.Add(1)
	go s.resultProcessor()

	s.started = true
	return nil
}

// Stop stops the scheduler and all workers
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return nil
	}

	close(s.shutdownChan)
	s.wg.Wait()

	s.started = false
	return nil
}

// Schedule adds a task to the queue
func (s *Scheduler) Schedule(task *Task) error {
	s.mu.RLock()
	if !s.started {
		s.mu.RUnlock()
		return ErrSchedulerStopped
	}
	s.mu.RUnlock()

	s.queueM.Lock()
	defer s.queueM.Unlock()

	// Check if queue is full
	if len(s.queue) >= s.config.QueueSize {
		return ErrQueueFull
	}

	// Check for duplicate
	if _, exists := s.tasks[task.ID]; exists {
		return errors.New("task already in queue")
	}

	// Update task status
	task.Status = TaskStatusQueued
	task.ScheduledAt = time.Now()

	// Add to queue
	heap.Push(&s.queue, task)
	s.tasks[task.ID] = task

	// Persist if repository is available
	if s.repo != nil {
		if err := s.repo.UpdateTask(context.Background(), task); err != nil {
			// Log error but don't fail scheduling
			_ = err
		}
	}

	return nil
}

// Next returns the next task to execute (for manual scheduling)
func (s *Scheduler) Next() (*Task, error) {
	s.queueM.Lock()
	defer s.queueM.Unlock()

	if len(s.queue) == 0 {
		return nil, errors.New("queue is empty")
	}

	task := heap.Pop(&s.queue).(*Task)
	delete(s.tasks, task.ID)

	return task, nil
}

// Cancel removes a task from the queue
func (s *Scheduler) Cancel(taskID string) error {
	s.queueM.Lock()
	defer s.queueM.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return ErrTaskNotFound
	}

	// Find and remove from queue
	for i, t := range s.queue {
		if t.ID == taskID {
			heap.Remove(&s.queue, i)
			delete(s.tasks, taskID)

			// Update task status
			task.Status = TaskStatusCancelled
			task.CompletedAt = time.Now()

			// Persist if repository is available
			if s.repo != nil {
				if err := s.repo.UpdateTask(context.Background(), task); err != nil {
					_ = err
				}
			}

			return nil
		}
	}

	return ErrTaskNotFound
}

// Size returns the number of queued tasks
func (s *Scheduler) Size() int {
	s.queueM.RLock()
	defer s.queueM.RUnlock()
	return len(s.queue)
}

// Clear removes all tasks from the queue
func (s *Scheduler) Clear() {
	s.queueM.Lock()
	defer s.queueM.Unlock()

	s.queue = make(priorityQueue, 0)
	s.tasks = make(map[string]*Task)
	heap.Init(&s.queue)
}

// dispatcher dispatches tasks from queue to workers
func (s *Scheduler) dispatcher() {
	defer s.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.shutdownChan:
			return

		case <-ticker.C:
			// Check if we need more workers
			s.scaleWorkers()

			// Dispatch tasks
			s.dispatchTasks()
		}
	}
}

// dispatchTasks sends tasks from queue to task channel
func (s *Scheduler) dispatchTasks() {
	for {
		s.queueM.Lock()
		if len(s.queue) == 0 {
			s.queueM.Unlock()
			return
		}

		task := heap.Pop(&s.queue).(*Task)
		delete(s.tasks, task.ID)
		s.queueM.Unlock()

		// Try to send task (non-blocking)
		select {
		case s.taskChan <- task:
			// Task dispatched successfully
		default:
			// Channel full, put task back
			s.queueM.Lock()
			heap.Push(&s.queue, task)
			s.tasks[task.ID] = task
			s.queueM.Unlock()
			return
		}
	}
}

// scaleWorkers adjusts the number of workers based on queue size
func (s *Scheduler) scaleWorkers() {
	s.queueM.RLock()
	queueSize := len(s.queue)
	s.queueM.RUnlock()

	currentWorkers := len(s.workers)

	// Scale up if queue is large
	if queueSize > currentWorkers*2 && currentWorkers < s.config.MaxWorkers {
		s.startWorker(currentWorkers)
	}

	// Note: Scale down is handled by worker idle timeout
}

// startWorker starts a new worker
func (s *Scheduler) startWorker(id int) {
	w := &worker{
		id:       id,
		taskChan: s.taskChan,
		executor: s.executor,
		wg:       &s.wg,
		shutdown: s.shutdownChan,
	}

	s.workers = append(s.workers, w)
	s.wg.Add(1)
	go w.run(s.resultChan, s.config.IdleTimeout)
}

// run executes the worker loop
func (w *worker) run(resultChan chan<- *TaskResult, idleTimeout time.Duration) {
	defer w.wg.Done()

	timer := time.NewTimer(idleTimeout)
	defer timer.Stop()

	for {
		timer.Reset(idleTimeout)

		select {
		case <-w.shutdown:
			return

		case task := <-w.taskChan:
			// Execute task
			task.Status = TaskStatusRunning
			task.StartedAt = time.Now()

			result, err := w.executor.Execute(task.Context(), task)
			if err != nil {
				if result == nil {
					result = &TaskResult{
						TaskID:      task.ID,
						AgentID:     task.AgentID,
						Status:      TaskStatusFailed,
						Error:       err.Error(),
						StartedAt:   task.StartedAt,
						CompletedAt: time.Now(),
					}
				}
			}

			// Update task status
			task.Status = result.Status
			task.CompletedAt = result.CompletedAt

			// Send result
			select {
			case resultChan <- result:
			case <-w.shutdown:
				return
			}

		case <-timer.C:
			// Idle timeout - worker can exit if above minimum
			// (The scheduler will recreate workers as needed)
			return
		}
	}
}

// resultProcessor processes task results
func (s *Scheduler) resultProcessor() {
	defer s.wg.Done()

	for {
		select {
		case <-s.shutdownChan:
			return

		case result := <-s.resultChan:
			// Persist result if repository is available
			if s.repo != nil {
				ctx := context.Background()
				if err := s.repo.StoreResult(ctx, result); err != nil {
					// Log error but continue
					_ = err
				}

				// Update metrics
				if err := s.updateMetrics(ctx, result); err != nil {
					_ = err
				}
			}

			// Handle retry if needed
			s.handleRetry(result)
		}
	}
}

// handleRetry checks if task should be retried and reschedules if needed
func (s *Scheduler) handleRetry(result *TaskResult) {
	if s.repo == nil {
		return
	}

	ctx := context.Background()
	task, err := s.repo.GetTask(ctx, result.TaskID)
	if err != nil {
		return
	}

	if !task.ShouldRetry(result) {
		return
	}

	// Calculate retry delay
	delay := task.GetRetryDelay(result.RetryCount)

	// Schedule retry after delay
	time.AfterFunc(delay, func() {
		task.Status = TaskStatusPending
		if err := s.Schedule(task); err != nil {
			// Log error
			_ = err
		}
	})
}

// updateMetrics updates aggregated task metrics
func (s *Scheduler) updateMetrics(ctx context.Context, result *TaskResult) error {
	metrics, err := s.repo.GetMetrics(ctx, result.AgentID)
	if err != nil {
		// Initialize new metrics
		metrics = &AgentTaskMetrics{
			AgentID:     result.AgentID,
			TasksByType: make(map[string]int64),
		}
	}

	// Update counters
	metrics.TotalTasks++
	switch result.Status {
	case TaskStatusCompleted:
		metrics.CompletedTasks++
	case TaskStatusFailed:
		metrics.FailedTasks++
	case TaskStatusCancelled:
		metrics.CancelledTasks++
	case TaskStatusTimeout:
		metrics.TimeoutTasks++
	}

	// Update duration
	durationMs := result.Duration.Milliseconds()
	metrics.TotalDurationMs += durationMs
	if metrics.CompletedTasks > 0 {
		metrics.AvgDurationMs = metrics.TotalDurationMs / metrics.CompletedTasks
	}

	// Update task type count
	task, err := s.repo.GetTask(ctx, result.TaskID)
	if err == nil {
		metrics.TasksByType[task.Type]++
	}

	metrics.LastUpdated = time.Now()

	return s.repo.UpdateMetrics(ctx, metrics)
}

// GetQueuedTasks returns all tasks currently in the queue
func (s *Scheduler) GetQueuedTasks() []*Task {
	s.queueM.RLock()
	defer s.queueM.RUnlock()

	tasks := make([]*Task, 0, len(s.queue))
	for _, task := range s.queue {
		tasks = append(tasks, task)
	}
	return tasks
}

// GetTask retrieves a task from the queue by ID
func (s *Scheduler) GetTask(taskID string) (*Task, error) {
	s.queueM.RLock()
	defer s.queueM.RUnlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, ErrTaskNotFound
	}
	return task, nil
}
