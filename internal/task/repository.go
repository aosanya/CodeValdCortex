package task

import (
	"context"
	"fmt"
	"time"

	"github.com/arangodb/go-driver"
	log "github.com/sirupsen/logrus"
)

const (
	// Collection names
	TasksCollection       = "agent_tasks"
	TaskResultsCollection = "agent_task_results"
	TaskMetricsCollection = "agent_task_metrics"
)

// Repository implements TaskRepository interface using ArangoDB
type Repository struct {
	db driver.Database

	// Collections
	tasks   driver.Collection
	results driver.Collection
	metrics driver.Collection
}

// NewRepository creates a new task repository
func NewRepository(db driver.Database) (*Repository, error) {
	repo := &Repository{
		db: db,
	}

	// Initialize collections
	if err := repo.initCollections(); err != nil {
		return nil, fmt.Errorf("failed to initialize collections: %w", err)
	}

	// Create indexes
	if err := repo.createIndexes(); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return repo, nil
}

// initCollections creates the required collections if they don't exist
func (r *Repository) initCollections() error {
	ctx := context.Background()

	// Tasks collection
	if exists, err := r.db.CollectionExists(ctx, TasksCollection); err != nil {
		return fmt.Errorf("failed to check tasks collection: %w", err)
	} else if !exists {
		collection, err := r.db.CreateCollection(ctx, TasksCollection, nil)
		if err != nil {
			return fmt.Errorf("failed to create tasks collection: %w", err)
		}
		r.tasks = collection
		log.WithField("collection", TasksCollection).Info("Created tasks collection")
	} else {
		collection, err := r.db.Collection(ctx, TasksCollection)
		if err != nil {
			return fmt.Errorf("failed to get tasks collection: %w", err)
		}
		r.tasks = collection
	}

	// Results collection
	if exists, err := r.db.CollectionExists(ctx, TaskResultsCollection); err != nil {
		return fmt.Errorf("failed to check results collection: %w", err)
	} else if !exists {
		collection, err := r.db.CreateCollection(ctx, TaskResultsCollection, nil)
		if err != nil {
			return fmt.Errorf("failed to create results collection: %w", err)
		}
		r.results = collection
		log.WithField("collection", TaskResultsCollection).Info("Created results collection")
	} else {
		collection, err := r.db.Collection(ctx, TaskResultsCollection)
		if err != nil {
			return fmt.Errorf("failed to get results collection: %w", err)
		}
		r.results = collection
	}

	// Metrics collection
	if exists, err := r.db.CollectionExists(ctx, TaskMetricsCollection); err != nil {
		return fmt.Errorf("failed to check metrics collection: %w", err)
	} else if !exists {
		collection, err := r.db.CreateCollection(ctx, TaskMetricsCollection, nil)
		if err != nil {
			return fmt.Errorf("failed to create metrics collection: %w", err)
		}
		r.metrics = collection
		log.WithField("collection", TaskMetricsCollection).Info("Created metrics collection")
	} else {
		collection, err := r.db.Collection(ctx, TaskMetricsCollection)
		if err != nil {
			return fmt.Errorf("failed to get metrics collection: %w", err)
		}
		r.metrics = collection
	}

	return nil
}

// createIndexes creates indexes for better query performance
func (r *Repository) createIndexes() error {
	ctx := context.Background()

	// Tasks collection indexes
	taskIndexes := []struct {
		name   string
		fields []string
		unique bool
	}{
		{"agent_id_idx", []string{"agent_id"}, false},
		{"status_idx", []string{"status"}, false},
		{"type_idx", []string{"type"}, false},
		{"priority_idx", []string{"priority"}, false},
		{"created_at_idx", []string{"created_at"}, false},
		{"agent_status_idx", []string{"agent_id", "status"}, false},
		{"status_priority_idx", []string{"status", "priority"}, false},
	}

	for _, idx := range taskIndexes {
		if exists, err := r.tasks.IndexExists(ctx, idx.name); err != nil {
			log.WithError(err).WithField("index", idx.name).Warn("Failed to check index existence")
		} else if !exists {
			_, _, err := r.tasks.EnsurePersistentIndex(ctx, idx.fields, &driver.EnsurePersistentIndexOptions{
				Name:   idx.name,
				Unique: idx.unique,
			})
			if err != nil {
				log.WithError(err).WithField("index", idx.name).Warn("Failed to create index")
			} else {
				log.WithField("index", idx.name).Info("Created task index")
			}
		}
	}

	// Results collection indexes
	resultIndexes := []struct {
		name   string
		fields []string
		unique bool
	}{
		{"task_id_idx", []string{"task_id"}, true},
		{"agent_id_idx", []string{"agent_id"}, false},
		{"status_idx", []string{"status"}, false},
		{"completed_at_idx", []string{"completed_at"}, false},
		{"agent_status_idx", []string{"agent_id", "status"}, false},
	}

	for _, idx := range resultIndexes {
		if exists, err := r.results.IndexExists(ctx, idx.name); err != nil {
			log.WithError(err).WithField("index", idx.name).Warn("Failed to check index existence")
		} else if !exists {
			_, _, err := r.results.EnsurePersistentIndex(ctx, idx.fields, &driver.EnsurePersistentIndexOptions{
				Name:   idx.name,
				Unique: idx.unique,
			})
			if err != nil {
				log.WithError(err).WithField("index", idx.name).Warn("Failed to create index")
			} else {
				log.WithField("index", idx.name).Info("Created result index")
			}
		}
	}

	// Metrics collection indexes
	metricIndexes := []struct {
		name   string
		fields []string
		unique bool
	}{
		{"agent_id_idx", []string{"agent_id"}, true},
		{"last_updated_idx", []string{"last_updated"}, false},
	}

	for _, idx := range metricIndexes {
		if exists, err := r.metrics.IndexExists(ctx, idx.name); err != nil {
			log.WithError(err).WithField("index", idx.name).Warn("Failed to check index existence")
		} else if !exists {
			_, _, err := r.metrics.EnsurePersistentIndex(ctx, idx.fields, &driver.EnsurePersistentIndexOptions{
				Name:   idx.name,
				Unique: idx.unique,
			})
			if err != nil {
				log.WithError(err).WithField("index", idx.name).Warn("Failed to create index")
			} else {
				log.WithField("index", idx.name).Info("Created metric index")
			}
		}
	}

	return nil
}

// StoreTask saves a task to the database
func (r *Repository) StoreTask(ctx context.Context, task *Task) error {
	// Use task ID as document key
	meta, err := r.tasks.CreateDocument(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to store task: %w", err)
	}

	log.WithFields(log.Fields{
		"task_id":     task.ID,
		"agent_id":    task.AgentID,
		"document_id": meta.ID,
	}).Debug("Stored task")

	return nil
}

// GetTask retrieves a task by ID
func (r *Repository) GetTask(ctx context.Context, taskID string) (*Task, error) {
	var task Task
	_, err := r.tasks.ReadDocument(ctx, taskID, &task)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, fmt.Errorf("task not found: %s", taskID)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &task, nil
}

// UpdateTask updates an existing task
func (r *Repository) UpdateTask(ctx context.Context, task *Task) error {
	_, err := r.tasks.UpdateDocument(ctx, task.ID, task)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	log.WithFields(log.Fields{
		"task_id":  task.ID,
		"agent_id": task.AgentID,
		"status":   task.Status,
	}).Debug("Updated task")

	return nil
}

// ListTasks lists tasks with filters
func (r *Repository) ListTasks(ctx context.Context, agentID string, filters TaskFilters) ([]*Task, error) {
	// Build AQL query
	query := "FOR task IN " + TasksCollection
	bindVars := make(map[string]interface{})
	conditions := make([]string, 0)

	// Agent ID filter
	if agentID != "" {
		conditions = append(conditions, "task.agent_id == @agent_id")
		bindVars["agent_id"] = agentID
	}

	// Status filter
	if len(filters.Status) > 0 {
		conditions = append(conditions, "task.status IN @statuses")
		bindVars["statuses"] = filters.Status
	}

	// Type filter
	if filters.Type != "" {
		conditions = append(conditions, "task.type == @type")
		bindVars["type"] = filters.Type
	}

	// Priority filter
	if filters.MinPriority > 0 {
		conditions = append(conditions, "task.priority >= @min_priority")
		bindVars["min_priority"] = filters.MinPriority
	}

	// Time filters
	if filters.CreatedAfter != nil {
		conditions = append(conditions, "task.created_at >= @created_after")
		bindVars["created_after"] = filters.CreatedAfter.Format(time.RFC3339)
	}
	if filters.CreatedBefore != nil {
		conditions = append(conditions, "task.created_at <= @created_before")
		bindVars["created_before"] = filters.CreatedBefore.Format(time.RFC3339)
	}

	// Add WHERE clause if conditions exist
	if len(conditions) > 0 {
		query += " FILTER " + fmt.Sprintf("(%s)", fmt.Sprintf("(%s)", fmt.Sprintf("%s", conditions[0])))
		for i := 1; i < len(conditions); i++ {
			query += " AND " + fmt.Sprintf("(%s)", conditions[i])
		}
	}

	// Add sorting
	query += " SORT task.created_at DESC"

	// Add limit
	if filters.Limit > 0 {
		query += " LIMIT @limit"
		bindVars["limit"] = filters.Limit
	}

	query += " RETURN task"

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer cursor.Close()

	var tasks []*Task
	for {
		var task Task
		_, err := cursor.ReadDocument(ctx, &task)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to read task from cursor: %w", err)
		}
		tasks = append(tasks, &task)
	}

	return tasks, nil
}

// StoreResult saves a task result
func (r *Repository) StoreResult(ctx context.Context, result *TaskResult) error {
	// Use task ID as document key for easy lookup
	meta, err := r.results.CreateDocument(ctx, result)
	if err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	log.WithFields(log.Fields{
		"task_id":     result.TaskID,
		"agent_id":    result.AgentID,
		"status":      result.Status,
		"document_id": meta.ID,
	}).Debug("Stored task result")

	return nil
}

// GetResult retrieves a task result
func (r *Repository) GetResult(ctx context.Context, taskID string) (*TaskResult, error) {
	var result TaskResult
	_, err := r.results.ReadDocument(ctx, taskID, &result)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, fmt.Errorf("result not found for task: %s", taskID)
		}
		return nil, fmt.Errorf("failed to get result: %w", err)
	}
	return &result, nil
}

// GetMetrics retrieves aggregated metrics
func (r *Repository) GetMetrics(ctx context.Context, agentID string) (*AgentTaskMetrics, error) {
	var metrics AgentTaskMetrics
	_, err := r.metrics.ReadDocument(ctx, agentID, &metrics)
	if err != nil {
		if driver.IsNotFound(err) {
			// Return zero metrics for agents without any tasks
			return &AgentTaskMetrics{
				AgentID:     agentID,
				TasksByType: make(map[string]int64),
				LastUpdated: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}
	return &metrics, nil
}

// UpdateMetrics updates aggregated metrics
func (r *Repository) UpdateMetrics(ctx context.Context, metrics *AgentTaskMetrics) error {
	// Use agent ID as document key
	_, err := r.metrics.UpdateDocument(ctx, metrics.AgentID, metrics)
	if err != nil {
		// If document doesn't exist, create it
		if driver.IsNotFound(err) {
			_, err = r.metrics.CreateDocument(ctx, metrics)
			if err != nil {
				return fmt.Errorf("failed to create metrics: %w", err)
			}
		} else {
			return fmt.Errorf("failed to update metrics: %w", err)
		}
	}

	log.WithFields(log.Fields{
		"agent_id":    metrics.AgentID,
		"total_tasks": metrics.TotalTasks,
		"completed":   metrics.CompletedTasks,
		"failed":      metrics.FailedTasks,
	}).Debug("Updated task metrics")

	return nil
}

// CleanupOldResults removes old task results
func (r *Repository) CleanupOldResults(ctx context.Context, before time.Time) (int, error) {
	query := `
		FOR result IN ` + TaskResultsCollection + `
		FILTER result.completed_at < @before
		REMOVE result IN ` + TaskResultsCollection + `
		RETURN OLD
	`

	bindVars := map[string]interface{}{
		"before": before.Format(time.RFC3339),
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old results: %w", err)
	}
	defer cursor.Close()

	count := 0
	for {
		var result interface{}
		_, err := cursor.ReadDocument(ctx, &result)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return count, fmt.Errorf("failed to read cleanup result: %w", err)
		}
		count++
	}

	log.WithFields(log.Fields{
		"count":  count,
		"before": before,
	}).Info("Cleaned up old task results")

	return count, nil
}
