package pool

import (
	"context"
	"fmt"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

// Repository handles persistence of agent pools in ArangoDB
type Repository struct {
	// db is the ArangoDB database instance
	db driver.Database

	// collections hold references to ArangoDB collections
	poolsCollection       driver.Collection
	membershipsCollection driver.Collection
	metricsCollection     driver.Collection

	// logger for repository operations
	logger Logger
}

// PoolDocument represents a pool in ArangoDB
type PoolDocument struct {
	Key                   string                `json:"_key"`
	ID                    string                `json:"id"`
	Name                  string                `json:"name"`
	Description           string                `json:"description"`
	LoadBalancingStrategy LoadBalancingStrategy `json:"load_balancing_strategy"`
	MinAgents             int                   `json:"min_agents"`
	MaxAgents             int                   `json:"max_agents"`
	HealthCheckInterval   int64                 `json:"health_check_interval_ms"`
	ResourceLimits        ResourceLimits        `json:"resource_limits"`
	AutoScaling           AutoScalingConfig     `json:"auto_scaling"`
	Status                PoolStatus            `json:"status"`
	CreatedAt             time.Time             `json:"created_at"`
	UpdatedAt             time.Time             `json:"updated_at"`
}

// MembershipDocument represents pool membership in ArangoDB
type MembershipDocument struct {
	Key               string    `json:"_key"`
	PoolID            string    `json:"pool_id"`
	AgentID           string    `json:"agent_id"`
	Weight            int       `json:"weight"`
	JoinedAt          time.Time `json:"joined_at"`
	ActiveConnections int       `json:"active_connections"`
	LastHealthCheck   time.Time `json:"last_health_check"`
	Healthy           bool      `json:"healthy"`
}

// MetricsDocument represents pool metrics in ArangoDB
type MetricsDocument struct {
	Key                 string              `json:"_key"`
	PoolID              string              `json:"pool_id"`
	TotalRequests       int64               `json:"total_requests"`
	ActiveRequests      int64               `json:"active_requests"`
	FailedRequests      int64               `json:"failed_requests"`
	AverageResponseTime float64             `json:"average_response_time"`
	TotalAgents         int                 `json:"total_agents"`
	HealthyAgents       int                 `json:"healthy_agents"`
	ResourceUtilization ResourceUtilization `json:"resource_utilization"`
	Timestamp           time.Time           `json:"timestamp"`
}

// RepositoryConfig holds configuration for the pool repository
type RepositoryConfig struct {
	DatabaseURL  string
	DatabaseName string
	Username     string
	Password     string
}

// NewRepository creates a new pool repository
func NewRepository(config RepositoryConfig, logger Logger) (*Repository, error) {
	// Create HTTP connection
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{config.DatabaseURL},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ArangoDB connection: %w", err)
	}

	// Create client
	client, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(config.Username, config.Password),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ArangoDB client: %w", err)
	}

	// Get database
	db, err := client.Database(context.Background(), config.DatabaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to access database: %w", err)
	}

	repo := &Repository{
		db:     db,
		logger: logger,
	}

	// Initialize collections
	if err := repo.initializeCollections(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize collections: %w", err)
	}

	return repo, nil
}

// initializeCollections creates and configures ArangoDB collections
func (r *Repository) initializeCollections(ctx context.Context) error {
	// Create agent_pools collection
	poolsExists, err := r.db.CollectionExists(ctx, "agent_pools")
	if err != nil {
		return fmt.Errorf("failed to check agent_pools collection: %w", err)
	}

	if !poolsExists {
		r.poolsCollection, err = r.db.CreateCollection(ctx, "agent_pools", nil)
		if err != nil {
			return fmt.Errorf("failed to create agent_pools collection: %w", err)
		}
	} else {
		r.poolsCollection, err = r.db.Collection(ctx, "agent_pools")
		if err != nil {
			return fmt.Errorf("failed to get agent_pools collection: %w", err)
		}
	}

	// Create agent_pool_memberships collection
	membershipsExists, err := r.db.CollectionExists(ctx, "agent_pool_memberships")
	if err != nil {
		return fmt.Errorf("failed to check agent_pool_memberships collection: %w", err)
	}

	if !membershipsExists {
		r.membershipsCollection, err = r.db.CreateCollection(ctx, "agent_pool_memberships", nil)
		if err != nil {
			return fmt.Errorf("failed to create agent_pool_memberships collection: %w", err)
		}
	} else {
		r.membershipsCollection, err = r.db.Collection(ctx, "agent_pool_memberships")
		if err != nil {
			return fmt.Errorf("failed to get agent_pool_memberships collection: %w", err)
		}
	}

	// Create agent_pool_metrics collection
	metricsExists, err := r.db.CollectionExists(ctx, "agent_pool_metrics")
	if err != nil {
		return fmt.Errorf("failed to check agent_pool_metrics collection: %w", err)
	}

	if !metricsExists {
		r.metricsCollection, err = r.db.CreateCollection(ctx, "agent_pool_metrics", nil)
		if err != nil {
			return fmt.Errorf("failed to create agent_pool_metrics collection: %w", err)
		}
	} else {
		r.metricsCollection, err = r.db.Collection(ctx, "agent_pool_metrics")
		if err != nil {
			return fmt.Errorf("failed to get agent_pool_metrics collection: %w", err)
		}
	}

	// Create indexes
	return r.createIndexes(ctx)
}

// createIndexes creates necessary database indexes
func (r *Repository) createIndexes(ctx context.Context) error {
	// Index on pool status
	_, _, err := r.poolsCollection.EnsurePersistentIndex(ctx, []string{"status"}, nil)
	if err != nil {
		r.logger.Warn("Failed to create status index", "error", err)
	}

	// Index on pool name
	_, _, err = r.poolsCollection.EnsurePersistentIndex(ctx, []string{"name"}, &driver.EnsurePersistentIndexOptions{
		Unique: true,
	})
	if err != nil {
		r.logger.Warn("Failed to create name index", "error", err)
	}

	// Index on membership pool_id and agent_id
	_, _, err = r.membershipsCollection.EnsurePersistentIndex(ctx, []string{"pool_id"}, nil)
	if err != nil {
		r.logger.Warn("Failed to create pool_id index", "error", err)
	}

	_, _, err = r.membershipsCollection.EnsurePersistentIndex(ctx, []string{"agent_id"}, nil)
	if err != nil {
		r.logger.Warn("Failed to create agent_id index", "error", err)
	}

	// Compound index on pool_id and agent_id for memberships
	_, _, err = r.membershipsCollection.EnsurePersistentIndex(ctx, []string{"pool_id", "agent_id"}, &driver.EnsurePersistentIndexOptions{
		Unique: true,
	})
	if err != nil {
		r.logger.Warn("Failed to create compound index", "error", err)
	}

	// Index on metrics pool_id and timestamp
	_, _, err = r.metricsCollection.EnsurePersistentIndex(ctx, []string{"pool_id"}, nil)
	if err != nil {
		r.logger.Warn("Failed to create metrics pool_id index", "error", err)
	}

	_, _, err = r.metricsCollection.EnsurePersistentIndex(ctx, []string{"timestamp"}, nil)
	if err != nil {
		r.logger.Warn("Failed to create timestamp index", "error", err)
	}

	return nil
}

// StorePool saves a pool configuration to ArangoDB
func (r *Repository) StorePool(ctx context.Context, pool *AgentPool) error {
	doc := &PoolDocument{
		Key:                   pool.ID,
		ID:                    pool.ID,
		Name:                  pool.Config.Name,
		Description:           pool.Config.Description,
		LoadBalancingStrategy: pool.Config.LoadBalancingStrategy,
		MinAgents:             pool.Config.MinAgents,
		MaxAgents:             pool.Config.MaxAgents,
		HealthCheckInterval:   pool.Config.HealthCheckInterval.Milliseconds(),
		ResourceLimits:        pool.Config.ResourceLimits,
		AutoScaling:           pool.Config.AutoScaling,
		Status:                pool.Status,
		CreatedAt:             pool.CreatedAt,
		UpdatedAt:             pool.UpdatedAt,
	}

	_, err := r.poolsCollection.CreateDocument(ctx, doc)
	if err != nil {
		if driver.IsConflict(err) {
			// Update existing document
			_, err = r.poolsCollection.ReplaceDocument(ctx, pool.ID, doc)
		}
		if err != nil {
			return fmt.Errorf("failed to store pool: %w", err)
		}
	}

	r.logger.Info("Stored pool configuration",
		"pool_id", pool.ID,
		"name", pool.Config.Name)

	return nil
}

// GetPool retrieves a pool configuration from ArangoDB
func (r *Repository) GetPool(ctx context.Context, poolID string) (*PoolDocument, error) {
	var doc PoolDocument
	_, err := r.poolsCollection.ReadDocument(ctx, poolID, &doc)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, fmt.Errorf("pool not found: %s", poolID)
		}
		return nil, fmt.Errorf("failed to get pool: %w", err)
	}

	return &doc, nil
}

// ListPools returns all pools with optional filtering
func (r *Repository) ListPools(ctx context.Context, status PoolStatus) ([]*PoolDocument, error) {
	var query string
	var bindVars map[string]interface{}

	if status != "" {
		query = "FOR p IN agent_pools FILTER p.status == @status RETURN p"
		bindVars = map[string]interface{}{
			"status": status,
		}
	} else {
		query = "FOR p IN agent_pools RETURN p"
		bindVars = map[string]interface{}{}
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query pools: %w", err)
	}
	defer cursor.Close()

	var pools []*PoolDocument
	for cursor.HasMore() {
		var doc PoolDocument
		_, err := cursor.ReadDocument(ctx, &doc)
		if err != nil {
			return nil, fmt.Errorf("failed to read pool document: %w", err)
		}
		pools = append(pools, &doc)
	}

	return pools, nil
}

// DeletePool removes a pool configuration from ArangoDB
func (r *Repository) DeletePool(ctx context.Context, poolID string) error {
	// Delete pool document
	_, err := r.poolsCollection.RemoveDocument(ctx, poolID)
	if err != nil && !driver.IsNotFound(err) {
		return fmt.Errorf("failed to delete pool: %w", err)
	}

	// Delete associated memberships
	query := "FOR m IN agent_pool_memberships FILTER m.pool_id == @pool_id REMOVE m IN agent_pool_memberships"
	bindVars := map[string]interface{}{
		"pool_id": poolID,
	}

	_, err = r.db.Query(ctx, query, bindVars)
	if err != nil {
		r.logger.Warn("Failed to delete pool memberships", "pool_id", poolID, "error", err)
	}

	r.logger.Info("Deleted pool", "pool_id", poolID)

	return nil
}

// StoreMembership saves a pool membership to ArangoDB
func (r *Repository) StoreMembership(ctx context.Context, poolID string, member *AgentPoolMember) error {
	membershipKey := fmt.Sprintf("%s_%s", poolID, member.Agent.ID)

	doc := &MembershipDocument{
		Key:               membershipKey,
		PoolID:            poolID,
		AgentID:           member.Agent.ID,
		Weight:            member.Weight,
		JoinedAt:          member.JoinedAt,
		ActiveConnections: member.ActiveConnections,
		LastHealthCheck:   member.LastHealthCheck,
		Healthy:           member.Healthy,
	}

	_, err := r.membershipsCollection.CreateDocument(ctx, doc)
	if err != nil {
		if driver.IsConflict(err) {
			// Update existing membership
			_, err = r.membershipsCollection.ReplaceDocument(ctx, membershipKey, doc)
		}
		if err != nil {
			return fmt.Errorf("failed to store membership: %w", err)
		}
	}

	r.logger.Debug("Stored pool membership",
		"pool_id", poolID,
		"agent_id", member.Agent.ID)

	return nil
}

// GetMemberships retrieves all memberships for a pool
func (r *Repository) GetMemberships(ctx context.Context, poolID string) ([]*MembershipDocument, error) {
	query := "FOR m IN agent_pool_memberships FILTER m.pool_id == @pool_id RETURN m"
	bindVars := map[string]interface{}{
		"pool_id": poolID,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query memberships: %w", err)
	}
	defer cursor.Close()

	var memberships []*MembershipDocument
	for cursor.HasMore() {
		var doc MembershipDocument
		_, err := cursor.ReadDocument(ctx, &doc)
		if err != nil {
			return nil, fmt.Errorf("failed to read membership document: %w", err)
		}
		memberships = append(memberships, &doc)
	}

	return memberships, nil
}

// RemoveMembership removes an agent from a pool
func (r *Repository) RemoveMembership(ctx context.Context, poolID, agentID string) error {
	membershipKey := fmt.Sprintf("%s_%s", poolID, agentID)

	_, err := r.membershipsCollection.RemoveDocument(ctx, membershipKey)
	if err != nil && !driver.IsNotFound(err) {
		return fmt.Errorf("failed to remove membership: %w", err)
	}

	r.logger.Debug("Removed pool membership",
		"pool_id", poolID,
		"agent_id", agentID)

	return nil
}

// StoreMetrics saves pool metrics to ArangoDB
func (r *Repository) StoreMetrics(ctx context.Context, poolID string, metrics *PoolMetrics) error {
	// Use timestamp-based key for metrics
	metricsKey := fmt.Sprintf("%s_%d", poolID, time.Now().UnixNano())

	doc := &MetricsDocument{
		Key:                 metricsKey,
		PoolID:              poolID,
		TotalRequests:       metrics.TotalRequests,
		ActiveRequests:      metrics.ActiveRequests,
		FailedRequests:      metrics.FailedRequests,
		AverageResponseTime: metrics.AverageResponseTime,
		TotalAgents:         metrics.TotalAgents,
		HealthyAgents:       metrics.HealthyAgents,
		ResourceUtilization: metrics.ResourceUtilization,
		Timestamp:           time.Now(),
	}

	_, err := r.metricsCollection.CreateDocument(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to store metrics: %w", err)
	}

	r.logger.Debug("Stored pool metrics", "pool_id", poolID)

	return nil
}

// GetLatestMetrics retrieves the most recent metrics for a pool
func (r *Repository) GetLatestMetrics(ctx context.Context, poolID string) (*MetricsDocument, error) {
	query := `
		FOR m IN agent_pool_metrics 
		FILTER m.pool_id == @pool_id 
		SORT m.timestamp DESC 
		LIMIT 1 
		RETURN m
	`
	bindVars := map[string]interface{}{
		"pool_id": poolID,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer cursor.Close()

	if !cursor.HasMore() {
		return nil, fmt.Errorf("no metrics found for pool: %s", poolID)
	}

	var doc MetricsDocument
	_, err = cursor.ReadDocument(ctx, &doc)
	if err != nil {
		return nil, fmt.Errorf("failed to read metrics document: %w", err)
	}

	return &doc, nil
}

// GetMetricsHistory retrieves historical metrics for a pool
func (r *Repository) GetMetricsHistory(ctx context.Context, poolID string, since time.Time, limit int) ([]*MetricsDocument, error) {
	query := `
		FOR m IN agent_pool_metrics 
		FILTER m.pool_id == @pool_id AND m.timestamp >= @since
		SORT m.timestamp DESC 
		LIMIT @limit 
		RETURN m
	`
	bindVars := map[string]interface{}{
		"pool_id": poolID,
		"since":   since,
		"limit":   limit,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics history: %w", err)
	}
	defer cursor.Close()

	var metrics []*MetricsDocument
	for cursor.HasMore() {
		var doc MetricsDocument
		_, err := cursor.ReadDocument(ctx, &doc)
		if err != nil {
			return nil, fmt.Errorf("failed to read metrics document: %w", err)
		}
		metrics = append(metrics, &doc)
	}

	return metrics, nil
}

// CleanupOldMetrics removes metrics older than the specified duration
func (r *Repository) CleanupOldMetrics(ctx context.Context, olderThan time.Duration) error {
	cutoffTime := time.Now().Add(-olderThan)

	query := "FOR m IN agent_pool_metrics FILTER m.timestamp < @cutoff REMOVE m IN agent_pool_metrics"
	bindVars := map[string]interface{}{
		"cutoff": cutoffTime,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to cleanup old metrics: %w", err)
	}
	defer cursor.Close()

	r.logger.Info("Cleaned up old pool metrics", "cutoff_time", cutoffTime)

	return nil
}
