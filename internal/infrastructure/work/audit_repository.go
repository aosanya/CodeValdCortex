package work

import (
	"context"
	"fmt"
	"time"

	driver "github.com/arangodb/go-driver"
	log "github.com/sirupsen/logrus"
)

// ArangoSyncAuditRepository implements SyncAuditRepository using ArangoDB
type ArangoSyncAuditRepository struct {
	db  driver.Database
	col driver.Collection
}

// NewSyncAuditRepository creates a new ArangoDB-backed sync audit repository
func NewSyncAuditRepository(db driver.Database) (SyncAuditRepository, error) {
	ctx := context.Background()

	// Ensure collection exists
	col, err := ensureCollection(ctx, db, CollectionSyncAudit)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure sync_audit collection: %w", err)
	}

	// Create indexes for efficient queries
	if err := ensureSyncAuditIndexes(ctx, col); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	log.WithField("collection", CollectionSyncAudit).Info("Sync audit repository initialized")

	return &ArangoSyncAuditRepository{
		db:  db,
		col: col,
	}, nil
}

// ensureSyncAuditIndexes creates necessary indexes
func ensureSyncAuditIndexes(ctx context.Context, col driver.Collection) error {
	// Index on agent_id for agent-specific audits
	_, _, err := col.EnsurePersistentIndex(ctx, []string{"agent_id"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_agent_id",
	})
	if err != nil {
		return fmt.Errorf("failed to create agent_id index: %w", err)
	}

	// Index on issue_id for issue-specific audits
	_, _, err = col.EnsurePersistentIndex(ctx, []string{"issue_id"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_issue_id",
	})
	if err != nil {
		return fmt.Errorf("failed to create issue_id index: %w", err)
	}

	// Index on sync_timestamp for time-based queries
	_, _, err = col.EnsurePersistentIndex(ctx, []string{"sync_timestamp"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_sync_timestamp",
	})
	if err != nil {
		return fmt.Errorf("failed to create sync_timestamp index: %w", err)
	}

	// Index on success for failure queries
	_, _, err = col.EnsurePersistentIndex(ctx, []string{"success"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_success",
	})
	if err != nil {
		return fmt.Errorf("failed to create success index: %w", err)
	}

	// Compound index for failure queries with time range
	_, _, err = col.EnsurePersistentIndex(ctx, []string{"success", "sync_timestamp"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_success_timestamp",
	})
	if err != nil {
		return fmt.Errorf("failed to create success+timestamp index: %w", err)
	}

	return nil
}

// Create creates a new audit record
func (r *ArangoSyncAuditRepository) Create(ctx context.Context, record *SyncAuditRecord) error {
	if record.SyncTimestamp.IsZero() {
		record.SyncTimestamp = time.Now()
	}

	meta, err := r.col.CreateDocument(ctx, record)
	if err != nil {
		return fmt.Errorf("failed to create sync audit record: %w", err)
	}

	// Set the Key from metadata
	record.Key = meta.Key

	log.WithFields(log.Fields{
		"agent_id":    record.AgentID,
		"event_type":  record.EventType,
		"sync_action": record.SyncAction,
		"success":     record.Success,
	}).Debug("Sync audit record created")

	return nil
}

// GetByAgentID retrieves all audit records for an agent
func (r *ArangoSyncAuditRepository) GetByAgentID(ctx context.Context, agentID string, limit int) ([]*SyncAuditRecord, error) {
	if limit <= 0 {
		limit = 100 // Default limit
	}

	query := `
		FOR record IN @@collection
		FILTER record.agent_id == @agent_id
		SORT record.sync_timestamp DESC
		LIMIT @limit
		RETURN record
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionSyncAudit,
		"agent_id":    agentID,
		"limit":       limit,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit records: %w", err)
	}
	defer cursor.Close()

	var records []*SyncAuditRecord
	for {
		var record SyncAuditRecord
		_, err := cursor.ReadDocument(ctx, &record)
		if driver.IsNoMoreDocuments(err) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read audit record: %w", err)
		}
		records = append(records, &record)
	}

	return records, nil
}

// GetByIssueID retrieves all audit records for an issue
func (r *ArangoSyncAuditRepository) GetByIssueID(ctx context.Context, issueID string, limit int) ([]*SyncAuditRecord, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		FOR record IN @@collection
		FILTER record.issue_id == @issue_id
		SORT record.sync_timestamp DESC
		LIMIT @limit
		RETURN record
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionSyncAudit,
		"issue_id":    issueID,
		"limit":       limit,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit records: %w", err)
	}
	defer cursor.Close()

	var records []*SyncAuditRecord
	for {
		var record SyncAuditRecord
		_, err := cursor.ReadDocument(ctx, &record)
		if driver.IsNoMoreDocuments(err) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read audit record: %w", err)
		}
		records = append(records, &record)
	}

	return records, nil
}

// GetFailures retrieves failed sync operations for retry/investigation
func (r *ArangoSyncAuditRepository) GetFailures(ctx context.Context, since int64, limit int) ([]*SyncAuditRecord, error) {
	if limit <= 0 {
		limit = 100
	}

	sinceTime := time.Unix(since, 0)

	query := `
		FOR record IN @@collection
		FILTER record.success == false
		FILTER record.sync_timestamp >= @since
		SORT record.sync_timestamp DESC
		LIMIT @limit
		RETURN record
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionSyncAudit,
		"since":       sinceTime,
		"limit":       limit,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query failed audits: %w", err)
	}
	defer cursor.Close()

	var records []*SyncAuditRecord
	for {
		var record SyncAuditRecord
		_, err := cursor.ReadDocument(ctx, &record)
		if driver.IsNoMoreDocuments(err) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read audit record: %w", err)
		}
		records = append(records, &record)
	}

	return records, nil
}

// GetStats calculates sync statistics for monitoring
func (r *ArangoSyncAuditRepository) GetStats(ctx context.Context, since int64) (*SyncStats, error) {
	sinceTime := time.Unix(since, 0)

	query := `
		LET records = (
			FOR record IN @@collection
			FILTER record.sync_timestamp >= @since
			RETURN record
		)
		
		LET total = LENGTH(records)
		LET successful = LENGTH(FOR r IN records FILTER r.success == true RETURN 1)
		LET failed = LENGTH(FOR r IN records FILTER r.success == false RETURN 1)
		LET retries = SUM(FOR r IN records RETURN r.retry_count)
		
		RETURN {
			total_events: total,
			successful_syncs: successful,
			failed_syncs: failed,
			retry_count: retries
		}
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionSyncAudit,
		"since":       sinceTime,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query stats: %w", err)
	}
	defer cursor.Close()

	if !cursor.HasMore() {
		return &SyncStats{}, nil
	}

	var result struct {
		TotalEvents      int64 `json:"total_events"`
		SuccessfulSyncs  int64 `json:"successful_syncs"`
		FailedSyncs      int64 `json:"failed_syncs"`
		RetryCount       int64 `json:"retry_count"`
	}

	_, err = cursor.ReadDocument(ctx, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to read stats: %w", err)
	}

	stats := &SyncStats{
		TotalEvents:     result.TotalEvents,
		SuccessfulSyncs: result.SuccessfulSyncs,
		FailedSyncs:     result.FailedSyncs,
		RetryCount:      result.RetryCount,
		// Note: Latency metrics would require additional tracking
		// For now, set to 0 and implement via metrics service later
		AverageLatencyMs: 0,
		P95LatencyMs:     0,
		P99LatencyMs:     0,
		CurrentQueueDepth: 0,
	}

	return stats, nil
}
