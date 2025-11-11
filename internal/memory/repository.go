package memory

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/database"
	driver "github.com/arangodb/go-driver"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	// Collection names
	CollectionWorkingMemory  = "agent_working_memory"
	CollectionLongtermMemory = "agent_longterm_memory"
	CollectionSnapshots      = "agent_state_snapshots"
	CollectionSyncStatus     = "agent_memory_sync"
)

// Repository handles memory persistence in ArangoDB
type Repository struct {
	db                 *database.ArangoClient
	workingMemCol      driver.Collection
	longtermMemCol     driver.Collection
	snapshotsCol       driver.Collection
	syncStatusCol      driver.Collection
	ensuredCollections bool
}

// NewRepository creates a new memory repository
func NewRepository(db *database.ArangoClient) (*Repository, error) {
	if db == nil {
		return nil, fmt.Errorf("database client is required")
	}

	repo := &Repository{
		db: db,
	}

	// Ensure collections and indexes exist
	if err := repo.ensureCollections(context.Background()); err != nil {
		log.WithError(err).Error("Failed to ensure memory collections")
		return nil, err
	}

	return repo, nil
}

// ensureCollections creates collections and indexes if they don't exist
func (r *Repository) ensureCollections(ctx context.Context) error {
	if r.ensuredCollections {
		return nil
	}

	db := r.db.Database()

	// Ensure working memory collection
	var err error
	r.workingMemCol, err = r.ensureCollection(ctx, db, CollectionWorkingMemory)
	if err != nil {
		return fmt.Errorf("failed to ensure working memory collection: %w", err)
	}

	// Ensure long-term memory collection
	r.longtermMemCol, err = r.ensureCollection(ctx, db, CollectionLongtermMemory)
	if err != nil {
		return fmt.Errorf("failed to ensure longterm memory collection: %w", err)
	}

	// Ensure snapshots collection
	r.snapshotsCol, err = r.ensureCollection(ctx, db, CollectionSnapshots)
	if err != nil {
		return fmt.Errorf("failed to ensure snapshots collection: %w", err)
	}

	// Ensure sync status collection
	r.syncStatusCol, err = r.ensureCollection(ctx, db, CollectionSyncStatus)
	if err != nil {
		return fmt.Errorf("failed to ensure sync status collection: %w", err)
	}

	// Create indexes
	if err := r.ensureIndexes(ctx); err != nil {
		return fmt.Errorf("failed to ensure indexes: %w", err)
	}

	r.ensuredCollections = true
	log.Info("Memory collections and indexes ensured")
	return nil
}

// ensureCollection creates a collection if it doesn't exist
func (r *Repository) ensureCollection(ctx context.Context, db driver.Database, name string) (driver.Collection, error) {
	exists, err := db.CollectionExists(ctx, name)
	if err != nil {
		return nil, err
	}

	if !exists {
		col, err := db.CreateCollection(ctx, name, nil)
		if err != nil {
			return nil, err
		}
		log.WithField("collection", name).Info("Created collection")
		return col, nil
	}

	return db.Collection(ctx, name)
}

// ensureIndexes creates necessary indexes for efficient queries
func (r *Repository) ensureIndexes(ctx context.Context) error {
	// Working memory indexes
	if err := r.ensureWorkingMemoryIndexes(ctx); err != nil {
		return err
	}

	// Long-term memory indexes
	if err := r.ensureLongtermMemoryIndexes(ctx); err != nil {
		return err
	}

	// Snapshot indexes
	if err := r.ensureSnapshotIndexes(ctx); err != nil {
		return err
	}

	// Sync status indexes
	if err := r.ensureSyncStatusIndexes(ctx); err != nil {
		return err
	}

	return nil
}

func (r *Repository) ensureWorkingMemoryIndexes(ctx context.Context) error {
	// Index for agent_id + key (unique lookup)
	_, _, err := r.workingMemCol.EnsurePersistentIndex(ctx, []string{"agent_id", "key"}, &driver.EnsurePersistentIndexOptions{
		Name:   "idx_working_memory_lookup",
		Unique: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create working memory lookup index: %w", err)
	}

	// Index for expiration cleanup
	_, _, err = r.workingMemCol.EnsurePersistentIndex(ctx, []string{"expires_at"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_working_memory_expiration",
	})
	if err != nil {
		return fmt.Errorf("failed to create working memory expiration index: %w", err)
	}

	// Index for tag-based search
	_, _, err = r.workingMemCol.EnsurePersistentIndex(ctx, []string{"agent_id", "metadata.tags[*]"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_working_memory_tags",
	})
	if err != nil {
		return fmt.Errorf("failed to create working memory tags index: %w", err)
	}

	return nil
}

func (r *Repository) ensureLongtermMemoryIndexes(ctx context.Context) error {
	// Index for agent_id + category
	_, _, err := r.longtermMemCol.EnsurePersistentIndex(ctx, []string{"agent_id", "category"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_longterm_memory_category",
	})
	if err != nil {
		return fmt.Errorf("failed to create longterm memory category index: %w", err)
	}

	// Index for agent_id + key (unique lookup)
	_, _, err = r.longtermMemCol.EnsurePersistentIndex(ctx, []string{"agent_id", "key"}, &driver.EnsurePersistentIndexOptions{
		Name:   "idx_longterm_memory_key",
		Unique: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create longterm memory key index: %w", err)
	}

	// Index for tag-based search
	_, _, err = r.longtermMemCol.EnsurePersistentIndex(ctx, []string{"agent_id", "metadata.tags[*]"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_longterm_memory_tags",
	})
	if err != nil {
		return fmt.Errorf("failed to create longterm memory tags index: %w", err)
	}

	// Index for importance-based retrieval
	_, _, err = r.longtermMemCol.EnsurePersistentIndex(ctx, []string{"agent_id", "metadata.importance"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_longterm_memory_importance",
	})
	if err != nil {
		return fmt.Errorf("failed to create longterm memory importance index: %w", err)
	}

	// Index for access patterns
	_, _, err = r.longtermMemCol.EnsurePersistentIndex(ctx, []string{"last_accessed", "access_count"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_longterm_memory_access",
	})
	if err != nil {
		return fmt.Errorf("failed to create longterm memory access index: %w", err)
	}

	return nil
}

func (r *Repository) ensureSnapshotIndexes(ctx context.Context) error {
	// Index for agent_id + created_at
	_, _, err := r.snapshotsCol.EnsurePersistentIndex(ctx, []string{"agent_id", "created_at"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_snapshots_agent_time",
	})
	if err != nil {
		return fmt.Errorf("failed to create snapshots agent_time index: %w", err)
	}

	// Index for expiration cleanup
	_, _, err = r.snapshotsCol.EnsurePersistentIndex(ctx, []string{"expires_at"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_snapshots_expiration",
	})
	if err != nil {
		return fmt.Errorf("failed to create snapshots expiration index: %w", err)
	}

	// Index for snapshot type
	_, _, err = r.snapshotsCol.EnsurePersistentIndex(ctx, []string{"agent_id", "snapshot_type"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_snapshots_type",
	})
	if err != nil {
		return fmt.Errorf("failed to create snapshots type index: %w", err)
	}

	return nil
}

func (r *Repository) ensureSyncStatusIndexes(ctx context.Context) error {
	// Index for agent_id + instance_id (unique)
	_, _, err := r.syncStatusCol.EnsurePersistentIndex(ctx, []string{"agent_id", "instance_id"}, &driver.EnsurePersistentIndexOptions{
		Name:   "idx_sync_status",
		Unique: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create sync status index: %w", err)
	}

	// Index for conflict detection
	_, _, err = r.syncStatusCol.EnsurePersistentIndex(ctx, []string{"agent_id", "status"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_sync_conflicts",
	})
	if err != nil {
		return fmt.Errorf("failed to create sync conflicts index: %w", err)
	}

	return nil
}

// ============================================================================
// Working Memory Operations
// ============================================================================

// StoreWorking creates a new working memory entry
func (r *Repository) StoreWorking(ctx context.Context, memory *WorkingMemory) error {
	if memory.ID == "" {
		memory.ID = uuid.New().String()
	}
	if memory.CreatedAt.IsZero() {
		memory.CreatedAt = time.Now()
	}
	memory.UpdatedAt = time.Now()
	memory.Version = 1

	doc := r.workingMemoryToDocument(memory)

	_, err := r.workingMemCol.CreateDocument(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to store working memory: %w", err)
	}


	return nil
}

// GetWorking retrieves a working memory entry by agent ID and key
func (r *Repository) GetWorking(ctx context.Context, agentID, key string) (*WorkingMemory, error) {
	query := `
		FOR m IN @@collection
		FILTER m.agent_id == @agent_id AND m.key == @key
		LIMIT 1
		RETURN m
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionWorkingMemory,
		"agent_id":    agentID,
		"key":         key,
	}

	cursor, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query working memory: %w", err)
	}
	defer cursor.Close()

	if !cursor.HasMore() {
		return nil, fmt.Errorf("working memory not found: %s/%s", agentID, key)
	}

	var doc map[string]interface{}
	_, err = cursor.ReadDocument(ctx, &doc)
	if err != nil {
		return nil, fmt.Errorf("failed to read working memory document: %w", err)
	}

	// Update access tracking
	memory := r.documentToWorkingMemory(doc)
	memory.AccessedAt = time.Now()
	memory.AccessCount++
	go r.updateAccessTracking(context.Background(), memory)

	return memory, nil
}

// UpdateWorking updates an existing working memory entry
func (r *Repository) UpdateWorking(ctx context.Context, memory *WorkingMemory) error {
	memory.UpdatedAt = time.Now()
	memory.Version++

	doc := r.workingMemoryToDocument(memory)

	query := `
		FOR m IN @@collection
		FILTER m.agent_id == @agent_id AND m.key == @key
		UPDATE m WITH @update IN @@collection
		RETURN NEW
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionWorkingMemory,
		"agent_id":    memory.AgentID,
		"key":         memory.Key,
		"update":      doc,
	}

	cursor, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to update working memory: %w", err)
	}
	defer cursor.Close()

	if !cursor.HasMore() {
		return fmt.Errorf("working memory not found for update: %s/%s", memory.AgentID, memory.Key)
	}


	return nil
}

// DeleteWorking removes a working memory entry
func (r *Repository) DeleteWorking(ctx context.Context, agentID, key string) error {
	query := `
		FOR m IN @@collection
		FILTER m.agent_id == @agent_id AND m.key == @key
		REMOVE m IN @@collection
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionWorkingMemory,
		"agent_id":    agentID,
		"key":         key,
	}

	_, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to delete working memory: %w", err)
	}


	return nil
}

// ListWorking retrieves all working memory entries for an agent
func (r *Repository) ListWorking(ctx context.Context, agentID string, filters MemoryFilters) ([]*WorkingMemory, error) {
	query := `
		FOR m IN @@collection
		FILTER m.agent_id == @agent_id
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionWorkingMemory,
		"agent_id":    agentID,
	}

	// Apply filters
	if len(filters.Tags) > 0 {
		query += ` FILTER LENGTH(INTERSECTION(m.metadata.tags, @tags)) > 0`
		bindVars["tags"] = filters.Tags
	}

	if filters.AfterTime != nil {
		query += ` FILTER m.created_at >= @after_time`
		bindVars["after_time"] = filters.AfterTime
	}

	if filters.BeforeTime != nil {
		query += ` FILTER m.created_at <= @before_time`
		bindVars["before_time"] = filters.BeforeTime
	}

	// Sorting
	if filters.SortBy != "" {
		direction := "ASC"
		if filters.SortDesc {
			direction = "DESC"
		}
		query += fmt.Sprintf(` SORT m.%s %s`, filters.SortBy, direction)
	} else {
		query += ` SORT m.created_at DESC`
	}

	// Pagination
	if filters.Limit > 0 {
		query += ` LIMIT @offset, @limit`
		bindVars["offset"] = filters.Offset
		bindVars["limit"] = filters.Limit
	}

	query += ` RETURN m`

	cursor, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to list working memory: %w", err)
	}
	defer cursor.Close()

	var memories []*WorkingMemory
	for cursor.HasMore() {
		var doc map[string]interface{}
		_, err := cursor.ReadDocument(ctx, &doc)
		if err != nil {
			log.WithError(err).Warn("Failed to read working memory document")
			continue
		}
		memories = append(memories, r.documentToWorkingMemory(doc))
	}

	return memories, nil
}

// ClearWorking removes all working memory entries for an agent
func (r *Repository) ClearWorking(ctx context.Context, agentID string) error {
	query := `
		FOR m IN @@collection
		FILTER m.agent_id == @agent_id
		REMOVE m IN @@collection
		COLLECT WITH COUNT INTO count
		RETURN count
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionWorkingMemory,
		"agent_id":    agentID,
	}

	cursor, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to clear working memory: %w", err)
	}
	defer cursor.Close()

	var count int
	if cursor.HasMore() {
		_, err := cursor.ReadDocument(ctx, &count)
		if err != nil {
			return fmt.Errorf("failed to read clear count: %w", err)
		}
	}

	log.WithFields(log.Fields{
		"agent_id": agentID,
		"count":    count,
	}).Info("Cleared working memory")

	return nil
}

// ============================================================================
// Long-term Memory Operations
// ============================================================================

// StoreLongterm creates a new long-term memory entry
func (r *Repository) StoreLongterm(ctx context.Context, memory *LongtermMemory) error {
	if memory.ID == "" {
		memory.ID = uuid.New().String()
	}
	if memory.CreatedAt.IsZero() {
		memory.CreatedAt = time.Now()
	}
	memory.UpdatedAt = time.Now()
	memory.Version = 1

	doc := r.longtermMemoryToDocument(memory)

	_, err := r.longtermMemCol.CreateDocument(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to store longterm memory: %w", err)
	}


	return nil
}

// GetLongterm retrieves a long-term memory entry by agent ID and key
func (r *Repository) GetLongterm(ctx context.Context, agentID, key string) (*LongtermMemory, error) {
	query := `
		FOR m IN @@collection
		FILTER m.agent_id == @agent_id AND m.key == @key
		LIMIT 1
		RETURN m
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionLongtermMemory,
		"agent_id":    agentID,
		"key":         key,
	}

	cursor, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query longterm memory: %w", err)
	}
	defer cursor.Close()

	if !cursor.HasMore() {
		return nil, fmt.Errorf("longterm memory not found: %s/%s", agentID, key)
	}

	var doc map[string]interface{}
	_, err = cursor.ReadDocument(ctx, &doc)
	if err != nil {
		return nil, fmt.Errorf("failed to read longterm memory document: %w", err)
	}

	// Update access tracking
	memory := r.documentToLongtermMemory(doc)
	memory.LastAccessed = time.Now()
	memory.AccessCount++
	go r.updateLongtermAccessTracking(context.Background(), memory)

	return memory, nil
}

// UpdateLongterm updates an existing long-term memory entry
func (r *Repository) UpdateLongterm(ctx context.Context, memory *LongtermMemory) error {
	memory.UpdatedAt = time.Now()
	memory.Version++

	doc := r.longtermMemoryToDocument(memory)

	query := `
		FOR m IN @@collection
		FILTER m.agent_id == @agent_id AND m.key == @key
		UPDATE m WITH @update IN @@collection
		RETURN NEW
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionLongtermMemory,
		"agent_id":    memory.AgentID,
		"key":         memory.Key,
		"update":      doc,
	}

	cursor, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to update longterm memory: %w", err)
	}
	defer cursor.Close()

	if !cursor.HasMore() {
		return fmt.Errorf("longterm memory not found for update: %s/%s", memory.AgentID, memory.Key)
	}


	return nil
}

// DeleteLongterm removes a long-term memory entry
func (r *Repository) DeleteLongterm(ctx context.Context, agentID, key string) error {
	query := `
		FOR m IN @@collection
		FILTER m.agent_id == @agent_id AND m.key == @key
		REMOVE m IN @@collection
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionLongtermMemory,
		"agent_id":    agentID,
		"key":         key,
	}

	_, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to delete longterm memory: %w", err)
	}


	return nil
}

// ListLongterm retrieves long-term memory entries with filtering
func (r *Repository) ListLongterm(ctx context.Context, agentID string, filters MemoryFilters) ([]*LongtermMemory, error) {
	query := `
		FOR m IN @@collection
		FILTER m.agent_id == @agent_id
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionLongtermMemory,
		"agent_id":    agentID,
	}

	// Apply filters
	if filters.Category != "" {
		query += ` FILTER m.category == @category`
		bindVars["category"] = filters.Category
	}

	if len(filters.Tags) > 0 {
		query += ` FILTER LENGTH(INTERSECTION(m.metadata.tags, @tags)) > 0`
		bindVars["tags"] = filters.Tags
	}

	if filters.MinImportance > 0 {
		query += ` FILTER m.metadata.importance >= @min_importance`
		bindVars["min_importance"] = filters.MinImportance
	}

	if filters.AfterTime != nil {
		query += ` FILTER m.created_at >= @after_time`
		bindVars["after_time"] = filters.AfterTime
	}

	if filters.BeforeTime != nil {
		query += ` FILTER m.created_at <= @before_time`
		bindVars["before_time"] = filters.BeforeTime
	}

	// Sorting
	if filters.SortBy != "" {
		direction := "ASC"
		if filters.SortDesc {
			direction = "DESC"
		}
		query += fmt.Sprintf(` SORT m.%s %s`, filters.SortBy, direction)
	} else {
		query += ` SORT m.metadata.importance DESC, m.created_at DESC`
	}

	// Pagination
	if filters.Limit > 0 {
		query += ` LIMIT @offset, @limit`
		bindVars["offset"] = filters.Offset
		bindVars["limit"] = filters.Limit
	}

	query += ` RETURN m`

	cursor, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to list longterm memory: %w", err)
	}
	defer cursor.Close()

	var memories []*LongtermMemory
	for cursor.HasMore() {
		var doc map[string]interface{}
		_, err := cursor.ReadDocument(ctx, &doc)
		if err != nil {
			log.WithError(err).Warn("Failed to read longterm memory document")
			continue
		}
		memories = append(memories, r.documentToLongtermMemory(doc))
	}

	return memories, nil
}

// SearchLongterm performs text-based search (basic implementation, can be enhanced)
func (r *Repository) SearchLongterm(ctx context.Context, agentID string, query MemoryQuery) ([]*LongtermMemory, error) {
	// For now, use tag-based and category filtering
	// Future: implement full-text search or vector similarity search
	return r.ListLongterm(ctx, agentID, query.Filters)
}

// ============================================================================
// State Snapshot Operations
// ============================================================================

// CreateSnapshot creates a new state snapshot
func (r *Repository) CreateSnapshot(ctx context.Context, snapshot *StateSnapshot) error {
	if snapshot.ID == "" {
		snapshot.ID = uuid.New().String()
	}
	if snapshot.CreatedAt.IsZero() {
		snapshot.CreatedAt = time.Now()
	}
	snapshot.Version = 1

	// Calculate checksum
	stateJSON, err := json.Marshal(snapshot.State)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot state: %w", err)
	}
	hash := sha256.Sum256(stateJSON)
	snapshot.Checksum = fmt.Sprintf("%x", hash)
	snapshot.Metadata.SizeBytes = int64(len(stateJSON))

	doc := r.snapshotToDocument(snapshot)

	_, err = r.snapshotsCol.CreateDocument(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	log.WithFields(log.Fields{
		"agent_id":      snapshot.AgentID,
		"snapshot_id":   snapshot.ID,
		"snapshot_type": snapshot.SnapshotType,
		"size_bytes":    snapshot.Metadata.SizeBytes,
	}).Info("Created state snapshot")

	return nil
}

// GetSnapshot retrieves a specific snapshot
func (r *Repository) GetSnapshot(ctx context.Context, snapshotID string) (*StateSnapshot, error) {
	query := `
		FOR s IN @@collection
		FILTER s.id == @snapshot_id
		LIMIT 1
		RETURN s
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionSnapshots,
		"snapshot_id": snapshotID,
	}

	cursor, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query snapshot: %w", err)
	}
	defer cursor.Close()

	if !cursor.HasMore() {
		return nil, fmt.Errorf("snapshot not found: %s", snapshotID)
	}

	var doc map[string]interface{}
	_, err = cursor.ReadDocument(ctx, &doc)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot document: %w", err)
	}

	return r.documentToSnapshot(doc), nil
}

// ListSnapshots retrieves snapshots for an agent with filtering
func (r *Repository) ListSnapshots(ctx context.Context, agentID string, filters SnapshotFilters) ([]*StateSnapshot, error) {
	query := `
		FOR s IN @@collection
		FILTER s.agent_id == @agent_id
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionSnapshots,
		"agent_id":    agentID,
	}

	// Apply filters
	if filters.SnapshotType != "" {
		query += ` FILTER s.snapshot_type == @snapshot_type`
		bindVars["snapshot_type"] = filters.SnapshotType
	}

	if filters.AfterTime != nil {
		query += ` FILTER s.created_at >= @after_time`
		bindVars["after_time"] = filters.AfterTime
	}

	if filters.BeforeTime != nil {
		query += ` FILTER s.created_at <= @before_time`
		bindVars["before_time"] = filters.BeforeTime
	}

	// Sort by creation time (newest first)
	query += ` SORT s.created_at DESC`

	// Pagination
	if filters.Limit > 0 {
		query += ` LIMIT @offset, @limit`
		bindVars["offset"] = filters.Offset
		bindVars["limit"] = filters.Limit
	}

	query += ` RETURN s`

	cursor, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}
	defer cursor.Close()

	var snapshots []*StateSnapshot
	for cursor.HasMore() {
		var doc map[string]interface{}
		_, err := cursor.ReadDocument(ctx, &doc)
		if err != nil {
			log.WithError(err).Warn("Failed to read snapshot document")
			continue
		}
		snapshots = append(snapshots, r.documentToSnapshot(doc))
	}

	return snapshots, nil
}

// DeleteSnapshot removes a snapshot
func (r *Repository) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	query := `
		FOR s IN @@collection
		FILTER s.id == @snapshot_id
		REMOVE s IN @@collection
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionSnapshots,
		"snapshot_id": snapshotID,
	}

	_, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}


	return nil
}

// ============================================================================
// Sync Status Operations
// ============================================================================

// GetSyncStatus retrieves the sync status for an agent instance
func (r *Repository) GetSyncStatus(ctx context.Context, agentID, instanceID string) (*SyncStatus, error) {
	query := `
		FOR s IN @@collection
		FILTER s.agent_id == @agent_id AND s.instance_id == @instance_id
		LIMIT 1
		RETURN s
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionSyncStatus,
		"agent_id":    agentID,
		"instance_id": instanceID,
	}

	cursor, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query sync status: %w", err)
	}
	defer cursor.Close()

	if !cursor.HasMore() {
		// Return default sync status if not found
		return &SyncStatus{
			AgentID:        agentID,
			InstanceID:     instanceID,
			Status:         SyncStateSynced,
			SyncVersion:    0,
			PendingChanges: 0,
			Conflicts:      []MemoryConflict{},
			Metadata:       make(map[string]interface{}),
		}, nil
	}

	var doc map[string]interface{}
	_, err = cursor.ReadDocument(ctx, &doc)
	if err != nil {
		return nil, fmt.Errorf("failed to read sync status document: %w", err)
	}

	return r.documentToSyncStatus(doc), nil
}

// UpdateSyncStatus updates the sync status for an agent instance
func (r *Repository) UpdateSyncStatus(ctx context.Context, status *SyncStatus) error {
	doc := r.syncStatusToDocument(status)

	query := `
		UPSERT { agent_id: @agent_id, instance_id: @instance_id }
		INSERT @insert
		UPDATE @update
		IN @@collection
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionSyncStatus,
		"agent_id":    status.AgentID,
		"instance_id": status.InstanceID,
		"insert":      doc,
		"update":      doc,
	}

	_, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to update sync status: %w", err)
	}

	return nil
}

// ============================================================================
// Maintenance Operations
// ============================================================================

// CleanupExpired removes expired memories and snapshots
func (r *Repository) CleanupExpired(ctx context.Context) (int, error) {
	totalDeleted := 0

	// Clean up expired working memory
	workingQuery := `
		FOR m IN @@collection
		FILTER m.expires_at < @now
		REMOVE m IN @@collection
		COLLECT WITH COUNT INTO count
		RETURN count
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionWorkingMemory,
		"now":         time.Now(),
	}

	cursor, err := r.db.Database().Query(ctx, workingQuery, bindVars)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup working memory: %w", err)
	}
	defer cursor.Close()

	var count int
	if cursor.HasMore() {
		cursor.ReadDocument(ctx, &count)
		totalDeleted += count
	}

	// Clean up expired snapshots
	snapshotQuery := `
		FOR s IN @@collection
		FILTER s.expires_at < @now
		REMOVE s IN @@collection
		COLLECT WITH COUNT INTO count
		RETURN count
	`

	bindVars["@collection"] = CollectionSnapshots

	cursor, err = r.db.Database().Query(ctx, snapshotQuery, bindVars)
	if err != nil {
		return totalDeleted, fmt.Errorf("failed to cleanup snapshots: %w", err)
	}
	defer cursor.Close()

	if cursor.HasMore() {
		cursor.ReadDocument(ctx, &count)
		totalDeleted += count
	}

	if totalDeleted > 0 {
		log.WithField("count", totalDeleted).Info("Cleaned up expired memories and snapshots")
	}

	return totalDeleted, nil
}

// GetMemoryStats retrieves memory usage statistics for an agent
func (r *Repository) GetMemoryStats(ctx context.Context, agentID string) (*MemoryStats, error) {
	stats := &MemoryStats{
		AgentID: agentID,
	}

	// Count working memory
	workingQuery := `
		FOR m IN @@collection
		FILTER m.agent_id == @agent_id
		COLLECT WITH COUNT INTO count
		RETURN count
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionWorkingMemory,
		"agent_id":    agentID,
	}

	cursor, err := r.db.Database().Query(ctx, workingQuery, bindVars)
	if err == nil {
		defer cursor.Close()
		if cursor.HasMore() {
			cursor.ReadDocument(ctx, &stats.WorkingMemoryCount)
		}
	}

	// Count long-term memory
	longtermQuery := `
		FOR m IN @@collection
		FILTER m.agent_id == @agent_id
		COLLECT WITH COUNT INTO count
		RETURN count
	`

	bindVars["@collection"] = CollectionLongtermMemory

	cursor, err = r.db.Database().Query(ctx, longtermQuery, bindVars)
	if err == nil {
		defer cursor.Close()
		if cursor.HasMore() {
			cursor.ReadDocument(ctx, &stats.LongtermMemoryCount)
		}
	}

	// Count snapshots
	snapshotQuery := `
		FOR s IN @@collection
		FILTER s.agent_id == @agent_id
		COLLECT WITH COUNT INTO count
		RETURN count
	`

	bindVars["@collection"] = CollectionSnapshots

	cursor, err = r.db.Database().Query(ctx, snapshotQuery, bindVars)
	if err == nil {
		defer cursor.Close()
		if cursor.HasMore() {
			cursor.ReadDocument(ctx, &stats.SnapshotCount)
		}
	}

	// Get last sync time
	syncQuery := `
		FOR s IN @@collection
		FILTER s.agent_id == @agent_id
		SORT s.last_sync_at DESC
		LIMIT 1
		RETURN s.last_sync_at
	`

	bindVars["@collection"] = CollectionSyncStatus

	cursor, err = r.db.Database().Query(ctx, syncQuery, bindVars)
	if err == nil {
		defer cursor.Close()
		if cursor.HasMore() {
			cursor.ReadDocument(ctx, &stats.LastSyncAt)
		}
	}

	// Get last snapshot time
	lastSnapshotQuery := `
		FOR s IN @@collection
		FILTER s.agent_id == @agent_id
		SORT s.created_at DESC
		LIMIT 1
		RETURN s.created_at
	`

	bindVars["@collection"] = CollectionSnapshots

	cursor, err = r.db.Database().Query(ctx, lastSnapshotQuery, bindVars)
	if err == nil {
		defer cursor.Close()
		if cursor.HasMore() {
			cursor.ReadDocument(ctx, &stats.LastSnapshotAt)
		}
	}

	return stats, nil
}

// ============================================================================
// Helper Methods - Document Conversion
// ============================================================================

func (r *Repository) workingMemoryToDocument(m *WorkingMemory) map[string]interface{} {
	return map[string]interface{}{
		"id":           m.ID,
		"agent_id":     m.AgentID,
		"key":          m.Key,
		"value":        m.Value,
		"metadata":     m.Metadata,
		"created_at":   m.CreatedAt,
		"updated_at":   m.UpdatedAt,
		"accessed_at":  m.AccessedAt,
		"access_count": m.AccessCount,
		"expires_at":   m.ExpiresAt,
		"version":      m.Version,
	}
}

func (r *Repository) documentToWorkingMemory(doc map[string]interface{}) *WorkingMemory {
	m := &WorkingMemory{}
	m.ID, _ = doc["id"].(string)
	m.AgentID, _ = doc["agent_id"].(string)
	m.Key, _ = doc["key"].(string)
	m.Value = doc["value"]
	m.Metadata, _ = doc["metadata"].(map[string]interface{})
	m.CreatedAt, _ = parseTime(doc["created_at"])
	m.UpdatedAt, _ = parseTime(doc["updated_at"])
	m.AccessedAt, _ = parseTime(doc["accessed_at"])

	// Handle numeric types from ArangoDB (always float64)
	if accessCount, ok := doc["access_count"].(float64); ok {
		m.AccessCount = int(accessCount)
	}

	m.ExpiresAt, _ = parseTime(doc["expires_at"])

	if version, ok := doc["version"].(float64); ok {
		m.Version = int(version)
	}

	if m.Metadata == nil {
		m.Metadata = make(map[string]interface{})
	}
	return m
}

func (r *Repository) longtermMemoryToDocument(m *LongtermMemory) map[string]interface{} {
	return map[string]interface{}{
		"id":            m.ID,
		"agent_id":      m.AgentID,
		"category":      m.Category,
		"key":           m.Key,
		"value":         m.Value,
		"embedding":     m.Embedding,
		"metadata":      m.Metadata,
		"created_at":    m.CreatedAt,
		"updated_at":    m.UpdatedAt,
		"last_accessed": m.LastAccessed,
		"access_count":  m.AccessCount,
		"version":       m.Version,
	}
}

func (r *Repository) documentToLongtermMemory(doc map[string]interface{}) *LongtermMemory {
	m := &LongtermMemory{}
	m.ID, _ = doc["id"].(string)
	m.AgentID, _ = doc["agent_id"].(string)
	m.Category, _ = doc["category"].(string)
	m.Key, _ = doc["key"].(string)
	m.Value = doc["value"]

	// Parse embedding if present
	if embeddingData, ok := doc["embedding"].([]interface{}); ok {
		m.Embedding = make([]float64, len(embeddingData))
		for i, v := range embeddingData {
			if f, ok := v.(float64); ok {
				m.Embedding[i] = f
			}
		}
	}

	// Parse metadata
	if metadataDoc, ok := doc["metadata"].(map[string]interface{}); ok {
		m.Metadata.Source, _ = metadataDoc["source"].(string)
		if importance, ok := metadataDoc["importance"].(float64); ok {
			m.Metadata.Importance = int(importance)
		}
		m.Metadata.Confidence, _ = metadataDoc["confidence"].(float64)
		if tagsData, ok := metadataDoc["tags"].([]interface{}); ok {
			m.Metadata.Tags = make([]string, len(tagsData))
			for i, v := range tagsData {
				if s, ok := v.(string); ok {
					m.Metadata.Tags[i] = s
				}
			}
		}
		if refsData, ok := metadataDoc["references"].([]interface{}); ok {
			m.Metadata.References = make([]string, len(refsData))
			for i, v := range refsData {
				if s, ok := v.(string); ok {
					m.Metadata.References[i] = s
				}
			}
		}
	}

	m.CreatedAt, _ = parseTime(doc["created_at"])
	m.UpdatedAt, _ = parseTime(doc["updated_at"])
	m.LastAccessed, _ = parseTime(doc["last_accessed"])
	if accessCount, ok := doc["access_count"].(float64); ok {
		m.AccessCount = int(accessCount)
	}
	if version, ok := doc["version"].(float64); ok {
		m.Version = int(version)
	}

	return m
}

func (r *Repository) snapshotToDocument(s *StateSnapshot) map[string]interface{} {
	return map[string]interface{}{
		"id":            s.ID,
		"agent_id":      s.AgentID,
		"snapshot_type": s.SnapshotType,
		"state":         s.State,
		"checksum":      s.Checksum,
		"metadata":      s.Metadata,
		"created_at":    s.CreatedAt,
		"expires_at":    s.ExpiresAt,
		"version":       s.Version,
	}
}

func (r *Repository) documentToSnapshot(doc map[string]interface{}) *StateSnapshot {
	s := &StateSnapshot{}
	s.ID, _ = doc["id"].(string)
	s.AgentID, _ = doc["agent_id"].(string)
	s.SnapshotType, _ = doc["snapshot_type"].(string)
	s.State, _ = doc["state"].(map[string]interface{})
	s.Checksum, _ = doc["checksum"].(string)

	// Parse metadata
	if metadataDoc, ok := doc["metadata"].(map[string]interface{}); ok {
		s.Metadata.Trigger, _ = metadataDoc["trigger"].(string)
		s.Metadata.Reason, _ = metadataDoc["reason"].(string)
		if sizeBytes, ok := metadataDoc["size_bytes"].(float64); ok {
			s.Metadata.SizeBytes = int64(sizeBytes)
		}
		s.Metadata.Compressed, _ = metadataDoc["compressed"].(bool)
	}

	s.CreatedAt, _ = parseTime(doc["created_at"])
	s.ExpiresAt, _ = parseTime(doc["expires_at"])
	if version, ok := doc["version"].(float64); ok {
		s.Version = int(version)
	}

	if s.State == nil {
		s.State = make(map[string]interface{})
	}

	return s
}

func (r *Repository) syncStatusToDocument(s *SyncStatus) map[string]interface{} {
	return map[string]interface{}{
		"agent_id":        s.AgentID,
		"instance_id":     s.InstanceID,
		"last_sync_at":    s.LastSyncAt,
		"sync_version":    s.SyncVersion,
		"pending_changes": s.PendingChanges,
		"conflicts":       s.Conflicts,
		"status":          string(s.Status),
		"metadata":        s.Metadata,
	}
}

func (r *Repository) documentToSyncStatus(doc map[string]interface{}) *SyncStatus {
	s := &SyncStatus{}
	s.AgentID, _ = doc["agent_id"].(string)
	s.InstanceID, _ = doc["instance_id"].(string)
	s.LastSyncAt, _ = parseTime(doc["last_sync_at"])
	if syncVersion, ok := doc["sync_version"].(float64); ok {
		s.SyncVersion = int(syncVersion)
	}
	if pendingChanges, ok := doc["pending_changes"].(float64); ok {
		s.PendingChanges = int(pendingChanges)
	}

	// Parse conflicts
	if conflictsData, ok := doc["conflicts"].([]interface{}); ok {
		s.Conflicts = make([]MemoryConflict, len(conflictsData))
		for i, v := range conflictsData {
			if conflictDoc, ok := v.(map[string]interface{}); ok {
				s.Conflicts[i] = r.documentToMemoryConflict(conflictDoc)
			}
		}
	}

	if statusStr, ok := doc["status"].(string); ok {
		s.Status = SyncState(statusStr)
	}
	s.Metadata, _ = doc["metadata"].(map[string]interface{})

	if s.Conflicts == nil {
		s.Conflicts = []MemoryConflict{}
	}
	if s.Metadata == nil {
		s.Metadata = make(map[string]interface{})
	}

	return s
}

func (r *Repository) documentToMemoryConflict(doc map[string]interface{}) MemoryConflict {
	c := MemoryConflict{}
	c.Key, _ = doc["key"].(string)
	if memType, ok := doc["memory_type"].(string); ok {
		c.MemoryType = MemoryType(memType)
	}
	if localVersion, ok := doc["local_version"].(float64); ok {
		c.LocalVersion = int(localVersion)
	}
	if remoteVersion, ok := doc["remote_version"].(float64); ok {
		c.RemoteVersion = int(remoteVersion)
	}
	c.LocalValue = doc["local_value"]
	c.RemoteValue = doc["remote_value"]
	c.LocalTime, _ = parseTime(doc["local_time"])
	c.RemoteTime, _ = parseTime(doc["remote_time"])
	c.DetectedAt, _ = parseTime(doc["detected_at"])
	return c
}

// updateAccessTracking asynchronously updates access tracking for working memory
func (r *Repository) updateAccessTracking(ctx context.Context, memory *WorkingMemory) {
	query := `
		FOR m IN @@collection
		FILTER m.agent_id == @agent_id AND m.key == @key
		UPDATE m WITH { accessed_at: @accessed_at, access_count: @access_count } IN @@collection
	`

	bindVars := map[string]interface{}{
		"@collection":  CollectionWorkingMemory,
		"agent_id":     memory.AgentID,
		"key":          memory.Key,
		"accessed_at":  memory.AccessedAt,
		"access_count": memory.AccessCount,
	}

	_, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		log.WithError(err).Warn("Failed to update access tracking for working memory")
	}
}

// updateLongtermAccessTracking asynchronously updates access tracking for long-term memory
func (r *Repository) updateLongtermAccessTracking(ctx context.Context, memory *LongtermMemory) {
	query := `
		FOR m IN @@collection
		FILTER m.agent_id == @agent_id AND m.key == @key
		UPDATE m WITH { last_accessed: @last_accessed, access_count: @access_count } IN @@collection
	`

	bindVars := map[string]interface{}{
		"@collection":   CollectionLongtermMemory,
		"agent_id":      memory.AgentID,
		"key":           memory.Key,
		"last_accessed": memory.LastAccessed,
		"access_count":  memory.AccessCount,
	}

	_, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		log.WithError(err).Warn("Failed to update access tracking for longterm memory")
	}
}

// parseTime helper to parse time from various formats
func parseTime(val interface{}) (time.Time, bool) {
	switch v := val.(type) {
	case time.Time:
		return v, true
	case string:
		t, err := time.Parse(time.RFC3339, v)
		if err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
