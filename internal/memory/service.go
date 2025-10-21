package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// Service implements the MemoryService interface
type Service struct {
	repo MemoryRepository
}

// NewService creates a new memory service
func NewService(repo MemoryRepository) *Service {
	return &Service{
		repo: repo,
	}
}

// ============================================================================
// Working Memory Operations
// ============================================================================

// StoreWorking stores a value in working memory with TTL
func (s *Service) StoreWorking(ctx context.Context, agentID, key string, value interface{}, ttl time.Duration) error {
	if agentID == "" {
		return fmt.Errorf("agent ID is required")
	}
	if key == "" {
		return fmt.Errorf("key is required")
	}

	now := time.Now()
	mem := &WorkingMemory{
		AgentID:   agentID,
		Key:       key,
		Value:     value,
		Metadata:  make(map[string]interface{}),
		ExpiresAt: now.Add(ttl),
	}

	err := s.repo.StoreWorking(ctx, mem)
	if err != nil {
		return fmt.Errorf("failed to store working memory: %w", err)
	}

	log.WithFields(log.Fields{
		"agent_id": agentID,
		"key":      key,
		"ttl":      ttl,
	}).Debug("Stored working memory")

	return nil
}

// RetrieveWorking retrieves a value from working memory
func (s *Service) RetrieveWorking(ctx context.Context, agentID, key string) (interface{}, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent ID is required")
	}
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	mem, err := s.repo.GetWorking(ctx, agentID, key)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve working memory: %w", err)
	}

	// Check if expired
	if time.Now().After(mem.ExpiresAt) {
		// Clean up expired memory
		go s.repo.DeleteWorking(context.Background(), agentID, key)
		return nil, fmt.Errorf("memory expired")
	}

	return mem.Value, nil
}

// UpdateWorking updates an existing working memory value
func (s *Service) UpdateWorking(ctx context.Context, agentID, key string, value interface{}) error {
	if agentID == "" {
		return fmt.Errorf("agent ID is required")
	}
	if key == "" {
		return fmt.Errorf("key is required")
	}

	// Get existing memory
	mem, err := s.repo.GetWorking(ctx, agentID, key)
	if err != nil {
		return fmt.Errorf("failed to get working memory for update: %w", err)
	}

	// Update value
	mem.Value = value
	mem.UpdatedAt = time.Now()

	err = s.repo.UpdateWorking(ctx, mem)
	if err != nil {
		return fmt.Errorf("failed to update working memory: %w", err)
	}

	log.WithFields(log.Fields{
		"agent_id": agentID,
		"key":      key,
	}).Debug("Updated working memory")

	return nil
}

// DeleteWorking deletes a working memory entry
func (s *Service) DeleteWorking(ctx context.Context, agentID, key string) error {
	if agentID == "" {
		return fmt.Errorf("agent ID is required")
	}
	if key == "" {
		return fmt.Errorf("key is required")
	}

	err := s.repo.DeleteWorking(ctx, agentID, key)
	if err != nil {
		return fmt.Errorf("failed to delete working memory: %w", err)
	}

	log.WithFields(log.Fields{
		"agent_id": agentID,
		"key":      key,
	}).Debug("Deleted working memory")

	return nil
}

// ClearWorking removes all working memory for an agent
func (s *Service) ClearWorking(ctx context.Context, agentID string) error {
	if agentID == "" {
		return fmt.Errorf("agent ID is required")
	}

	err := s.repo.ClearWorking(ctx, agentID)
	if err != nil {
		return fmt.Errorf("failed to clear working memory: %w", err)
	}

	log.WithField("agent_id", agentID).Info("Cleared working memory")

	return nil
}

// ListWorking lists working memory entries with optional filters
func (s *Service) ListWorking(ctx context.Context, agentID string, filters MemoryFilters) ([]*WorkingMemory, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent ID is required")
	}

	memories, err := s.repo.ListWorking(ctx, agentID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list working memory: %w", err)
	}

	return memories, nil
}

// ============================================================================
// Long-term Memory Operations
// ============================================================================

// Remember stores a value in long-term memory with metadata
func (s *Service) Remember(ctx context.Context, agentID, key string, value interface{}, category string, metadata map[string]interface{}) error {
	if agentID == "" {
		return fmt.Errorf("agent ID is required")
	}
	if key == "" {
		return fmt.Errorf("key is required")
	}
	if category == "" {
		category = "general"
	}

	// Extract metadata fields
	memMetadata := MemoryMetadata{
		Source:     "manual",
		Importance: 5, // Default medium importance
		Confidence: 1.0,
		Tags:       []string{},
	}

	if metadata != nil {
		if source, ok := metadata["source"].(string); ok {
			memMetadata.Source = source
		}
		if importance, ok := metadata["importance"].(int); ok {
			memMetadata.Importance = importance
		}
		if confidence, ok := metadata["confidence"].(float64); ok {
			memMetadata.Confidence = confidence
		}
		if tags, ok := metadata["tags"].([]string); ok {
			memMetadata.Tags = tags
		}
		if refs, ok := metadata["references"].([]string); ok {
			memMetadata.References = refs
		}
	}

	mem := &LongtermMemory{
		AgentID:  agentID,
		Category: category,
		Key:      key,
		Value:    value,
		Metadata: memMetadata,
	}

	err := s.repo.StoreLongterm(ctx, mem)
	if err != nil {
		return fmt.Errorf("failed to remember: %w", err)
	}

	log.WithFields(log.Fields{
		"agent_id":   agentID,
		"category":   category,
		"key":        key,
		"importance": memMetadata.Importance,
	}).Info("Stored long-term memory")

	return nil
}

// Recall retrieves a value from long-term memory
func (s *Service) Recall(ctx context.Context, agentID, key string) (interface{}, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent ID is required")
	}
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	mem, err := s.repo.GetLongterm(ctx, agentID, key)
	if err != nil {
		return nil, fmt.Errorf("failed to recall: %w", err)
	}

	return mem.Value, nil
}

// Search searches long-term memory based on query criteria
func (s *Service) Search(ctx context.Context, agentID string, query MemoryQuery) ([]*LongtermMemory, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent ID is required")
	}

	// Use the repository's search (or list with filters)
	memories, err := s.repo.SearchLongterm(ctx, agentID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search memory: %w", err)
	}

	return memories, nil
}

// Forget removes a long-term memory entry
func (s *Service) Forget(ctx context.Context, agentID, key string) error {
	if agentID == "" {
		return fmt.Errorf("agent ID is required")
	}
	if key == "" {
		return fmt.Errorf("key is required")
	}

	err := s.repo.DeleteLongterm(ctx, agentID, key)
	if err != nil {
		return fmt.Errorf("failed to forget: %w", err)
	}

	log.WithFields(log.Fields{
		"agent_id": agentID,
		"key":      key,
	}).Info("Forgot long-term memory")

	return nil
}

// Archive moves old or low-importance memories to archive/deletion
func (s *Service) Archive(ctx context.Context, agentID string, criteria ArchiveCriteria) error {
	if agentID == "" {
		return fmt.Errorf("agent ID is required")
	}

	// Build filters based on criteria
	filters := MemoryFilters{}

	if len(criteria.Categories) > 0 {
		// Archive specific categories only
		for _, category := range criteria.Categories {
			filters.Category = category

			// Get memories matching criteria
			if criteria.OlderThan > 0 {
				olderThan := time.Now().Add(-criteria.OlderThan)
				filters.BeforeTime = &olderThan
			}

			if criteria.MaxImportance > 0 {
				filters.MinImportance = 0 // Will need custom query for max
			}

			memories, err := s.repo.ListLongterm(ctx, agentID, filters)
			if err != nil {
				return fmt.Errorf("failed to list memories for archival: %w", err)
			}

			// Process each memory
			archived := 0
			for _, mem := range memories {
				// Check access count if specified
				if criteria.MaxAccessCount > 0 && mem.AccessCount > criteria.MaxAccessCount {
					continue
				}

				// Check importance if specified
				if criteria.MaxImportance > 0 && mem.Metadata.Importance > criteria.MaxImportance {
					continue
				}

				if criteria.DryRun {
					log.WithFields(log.Fields{
						"agent_id": agentID,
						"key":      mem.Key,
						"category": mem.Category,
					}).Info("Would archive memory (dry run)")
				} else {
					// Actually delete/archive
					if err := s.repo.DeleteLongterm(ctx, agentID, mem.Key); err != nil {
						log.WithError(err).Warn("Failed to archive memory")
						continue
					}
				}
				archived++
			}

			log.WithFields(log.Fields{
				"agent_id": agentID,
				"category": category,
				"count":    archived,
				"dry_run":  criteria.DryRun,
			}).Info("Archived memories")
		}
	} else {
		// Archive all categories
		if criteria.OlderThan > 0 {
			olderThan := time.Now().Add(-criteria.OlderThan)
			filters.BeforeTime = &olderThan
		}

		memories, err := s.repo.ListLongterm(ctx, agentID, filters)
		if err != nil {
			return fmt.Errorf("failed to list memories for archival: %w", err)
		}

		archived := 0
		for _, mem := range memories {
			// Apply additional filters
			if criteria.MaxAccessCount > 0 && mem.AccessCount > criteria.MaxAccessCount {
				continue
			}
			if criteria.MaxImportance > 0 && mem.Metadata.Importance > criteria.MaxImportance {
				continue
			}

			if criteria.DryRun {
				log.WithFields(log.Fields{
					"agent_id": agentID,
					"key":      mem.Key,
				}).Info("Would archive memory (dry run)")
			} else {
				if err := s.repo.DeleteLongterm(ctx, agentID, mem.Key); err != nil {
					log.WithError(err).Warn("Failed to archive memory")
					continue
				}
			}
			archived++
		}

		log.WithFields(log.Fields{
			"agent_id": agentID,
			"count":    archived,
			"dry_run":  criteria.DryRun,
		}).Info("Archived memories")
	}

	return nil
}

// ============================================================================
// State Snapshot Operations
// ============================================================================

// CreateSnapshot creates a point-in-time snapshot of agent state
func (s *Service) CreateSnapshot(ctx context.Context, agentID string, snapshotType, reason string) (*StateSnapshot, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent ID is required")
	}
	if snapshotType == "" {
		snapshotType = "manual"
	}
	if reason == "" {
		reason = "manual snapshot"
	}

	// Build state from current memories
	state := make(map[string]interface{})

	// Include working memory keys
	workingMems, err := s.repo.ListWorking(ctx, agentID, MemoryFilters{})
	if err == nil {
		workingKeys := make([]string, 0, len(workingMems))
		for _, mem := range workingMems {
			workingKeys = append(workingKeys, mem.Key)
		}
		state["working_memory_keys"] = workingKeys
		state["working_memory_count"] = len(workingKeys)
	}

	// Include long-term memory summary
	longtermMems, err := s.repo.ListLongterm(ctx, agentID, MemoryFilters{})
	if err == nil {
		categoryCounts := make(map[string]int)
		for _, mem := range longtermMems {
			categoryCounts[mem.Category]++
		}
		state["longterm_memory_categories"] = categoryCounts
		state["longterm_memory_count"] = len(longtermMems)
	}

	// Add timestamp
	state["snapshot_time"] = time.Now()

	// Determine expiration based on snapshot type
	var expiresAt time.Time
	switch snapshotType {
	case "periodic":
		expiresAt = time.Now().Add(7 * 24 * time.Hour) // 7 days
	case "manual":
		expiresAt = time.Now().Add(30 * 24 * time.Hour) // 30 days
	case "pre-update", "pre-shutdown":
		expiresAt = time.Now().Add(90 * 24 * time.Hour) // 90 days
	default:
		expiresAt = time.Now().Add(30 * 24 * time.Hour) // Default 30 days
	}

	snapshot := &StateSnapshot{
		AgentID:      agentID,
		SnapshotType: snapshotType,
		State:        state,
		Metadata: SnapshotMetadata{
			Trigger: "service",
			Reason:  reason,
		},
		ExpiresAt: expiresAt,
	}

	err = s.repo.CreateSnapshot(ctx, snapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	log.WithFields(log.Fields{
		"agent_id":      agentID,
		"snapshot_id":   snapshot.ID,
		"snapshot_type": snapshotType,
	}).Info("Created state snapshot")

	return snapshot, nil
}

// RestoreSnapshot restores agent state from a snapshot
func (s *Service) RestoreSnapshot(ctx context.Context, agentID, snapshotID string) error {
	if agentID == "" {
		return fmt.Errorf("agent ID is required")
	}
	if snapshotID == "" {
		return fmt.Errorf("snapshot ID is required")
	}

	// Get snapshot
	snapshot, err := s.repo.GetSnapshot(ctx, snapshotID)
	if err != nil {
		return fmt.Errorf("failed to get snapshot: %w", err)
	}

	// Verify agent ID matches
	if snapshot.AgentID != agentID {
		return fmt.Errorf("snapshot does not belong to agent %s", agentID)
	}

	// For now, we'll just log the restoration
	// Full restoration would require more complex logic to rebuild state
	log.WithFields(log.Fields{
		"agent_id":    agentID,
		"snapshot_id": snapshotID,
	}).Info("Snapshot restoration requested (not yet fully implemented)")

	// TODO: Implement full state restoration
	// This would involve:
	// 1. Clearing current working memory
	// 2. Potentially restoring working memory from snapshot
	// 3. Ensuring long-term memory consistency

	return fmt.Errorf("snapshot restoration not yet fully implemented")
}

// ListSnapshots lists snapshots for an agent with filters
func (s *Service) ListSnapshots(ctx context.Context, agentID string, filters SnapshotFilters) ([]*StateSnapshot, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent ID is required")
	}

	snapshots, err := s.repo.ListSnapshots(ctx, agentID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}

	return snapshots, nil
}

// DeleteSnapshot deletes a specific snapshot
func (s *Service) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	if snapshotID == "" {
		return fmt.Errorf("snapshot ID is required")
	}

	err := s.repo.DeleteSnapshot(ctx, snapshotID)
	if err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	log.WithField("snapshot_id", snapshotID).Debug("Deleted snapshot")

	return nil
}

// ============================================================================
// Synchronization Operations
// ============================================================================

// SyncMemory performs a basic synchronization (full implementation requires Synchronizer)
func (s *Service) SyncMemory(ctx context.Context, agentID string) (*SyncResult, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent ID is required")
	}

	// This is a simplified version
	// Full sync logic is in the Synchronizer component

	startTime := time.Now()
	result := &SyncResult{
		AgentID:     agentID,
		SyncedAt:    time.Now(),
		ItemsSynced: 0,
		Conflicts:   []MemoryConflict{},
		Errors:      []string{},
		Success:     false,
	}

	// Get current sync status
	instanceID := uuid.New().String()
	status, err := s.repo.GetSyncStatus(ctx, agentID, instanceID)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to get sync status: %v", err))
		return result, err
	}

	// Update sync status
	status.LastSyncAt = time.Now()
	status.Status = SyncStateSynced
	status.PendingChanges = 0

	err = s.repo.UpdateSyncStatus(ctx, status)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to update sync status: %v", err))
		return result, err
	}

	result.DurationMs = time.Since(startTime).Milliseconds()
	result.Success = true

	log.WithFields(log.Fields{
		"agent_id":    agentID,
		"duration_ms": result.DurationMs,
	}).Debug("Basic sync completed")

	return result, nil
}

// GetSyncStatus retrieves the current synchronization status
func (s *Service) GetSyncStatus(ctx context.Context, agentID string) (*SyncStatus, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent ID is required")
	}

	// Use a default instance ID for now
	instanceID := "default"
	status, err := s.repo.GetSyncStatus(ctx, agentID, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync status: %w", err)
	}

	return status, nil
}

// ResolveConflict resolves a memory conflict using the specified strategy
func (s *Service) ResolveConflict(ctx context.Context, conflict *MemoryConflict, strategy ConflictStrategy) error {
	if conflict == nil {
		return fmt.Errorf("conflict is required")
	}

	log.WithFields(log.Fields{
		"key":      conflict.Key,
		"strategy": strategy,
	}).Info("Resolving memory conflict")

	// Apply resolution strategy
	switch strategy {
	case ConflictStrategyLastWriteWins:
		// Use the version with the latest timestamp
		if conflict.RemoteTime.After(conflict.LocalTime) {
			log.Debug("Remote version is newer, using remote")
			// Would update local with remote value
		} else {
			log.Debug("Local version is newer, keeping local")
			// Keep local value
		}

	case ConflictStrategyLocalWins:
		log.Debug("Using local version (local wins strategy)")
		// Keep local value, push to remote

	case ConflictStrategyRemoteWins:
		log.Debug("Using remote version (remote wins strategy)")
		// Update local with remote value

	case ConflictStrategyVersionBased:
		// Compare version numbers
		if conflict.RemoteVersion > conflict.LocalVersion {
			log.Debug("Remote version number is higher, using remote")
		} else {
			log.Debug("Local version number is higher or equal, keeping local")
		}

	case ConflictStrategyManual:
		// Requires manual intervention
		return fmt.Errorf("manual conflict resolution required for key: %s", conflict.Key)

	default:
		return fmt.Errorf("unknown conflict strategy: %s", strategy)
	}

	// Note: Actual implementation would update the memory here
	// For now, this is a placeholder

	return nil
}

// ============================================================================
// Maintenance Operations
// ============================================================================

// CleanupExpired removes expired memories and snapshots
func (s *Service) CleanupExpired(ctx context.Context) (int, error) {
	count, err := s.repo.CleanupExpired(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired items: %w", err)
	}

	if count > 0 {
		log.WithField("count", count).Info("Cleaned up expired memories and snapshots")
	}

	return count, nil
}

// GetMemoryStats retrieves memory usage statistics for an agent
func (s *Service) GetMemoryStats(ctx context.Context, agentID string) (*MemoryStats, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent ID is required")
	}

	stats, err := s.repo.GetMemoryStats(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory stats: %w", err)
	}

	return stats, nil
}
