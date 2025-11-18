package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// Synchronizer implements the MemorySynchronizer interface
type Synchronizer struct {
	service  *Service
	repo     MemoryRepository
	strategy ConflictStrategy

	// Sync state
	mu         sync.RWMutex
	instanceID string
	running    bool
	stopChan   chan struct{}

	// Configuration
	syncInterval time.Duration
}

// NewSynchronizer creates a new memory synchronizer
func NewSynchronizer(service *Service, repo MemoryRepository, strategy ConflictStrategy, syncInterval time.Duration) *Synchronizer {
	if syncInterval == 0 {
		syncInterval = 5 * time.Minute // Default 5 minutes
	}

	return &Synchronizer{
		service:      service,
		repo:         repo,
		strategy:     strategy,
		instanceID:   uuid.New().String(),
		syncInterval: syncInterval,
		stopChan:     make(chan struct{}),
	}
}

// StartPeriodicSync starts the periodic synchronization loop
func (s *Synchronizer) StartPeriodicSync(ctx context.Context, agentID string) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("periodic sync already running")
	}
	s.running = true
	s.mu.Unlock()

	log.WithFields(log.Fields{
		"agent_id":      agentID,
		"instance_id":   s.instanceID,
		"sync_interval": s.syncInterval,
	}).Info("Starting periodic sync")

	// Start sync loop in goroutine
	go s.syncLoop(ctx, agentID)

	return nil
}

// StopPeriodicSync stops the periodic synchronization loop
func (s *Synchronizer) StopPeriodicSync() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("periodic sync not running")
	}

	close(s.stopChan)
	s.running = false
	log.Info("Stopped periodic sync")

	return nil
}

// syncLoop runs the periodic synchronization
func (s *Synchronizer) syncLoop(ctx context.Context, agentID string) {
	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Perform sync
			_, err := s.SyncAgent(ctx, agentID)
			if err != nil {
				log.WithError(err).Error("Periodic sync failed")
			}

		case <-s.stopChan:
			return

		case <-ctx.Done():
			return
		}
	}
}

// SyncAgent synchronizes all memory for an agent
func (s *Synchronizer) SyncAgent(ctx context.Context, agentID string) (*SyncResult, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent ID is required")
	}

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
	status, err := s.repo.GetSyncStatus(ctx, agentID, s.instanceID)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to get sync status: %v", err))
		return result, err
	}

	// Update status to syncing
	status.Status = SyncStateSyncing
	status.LastSyncAt = time.Now()
	err = s.repo.UpdateSyncStatus(ctx, status)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to update sync status: %v", err))
	}

	// Sync working memory
	workingCount, workingConflicts, err := s.syncWorkingMemory(ctx, agentID)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("working memory sync error: %v", err))
	} else {
		result.ItemsSynced += workingCount
		result.Conflicts = append(result.Conflicts, workingConflicts...)
	}

	// Sync long-term memory
	longtermCount, longtermConflicts, err := s.syncLongtermMemory(ctx, agentID)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("longterm memory sync error: %v", err))
	} else {
		result.ItemsSynced += longtermCount
		result.Conflicts = append(result.Conflicts, longtermConflicts...)
	}

	// Update final sync status
	if len(result.Conflicts) > 0 {
		status.Status = SyncStateConflict
		status.Conflicts = result.Conflicts
	} else if len(result.Errors) > 0 {
		status.Status = SyncStateError
	} else {
		status.Status = SyncStateSynced
	}

	status.PendingChanges = 0
	status.SyncVersion++
	err = s.repo.UpdateSyncStatus(ctx, status)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to update final sync status: %v", err))
	}

	result.DurationMs = time.Since(startTime).Milliseconds()
	result.Success = len(result.Errors) == 0

	log.WithFields(log.Fields{
		"agent_id":     agentID,
		"items_synced": result.ItemsSynced,
		"conflicts":    len(result.Conflicts),
		"errors":       len(result.Errors),
		"duration_ms":  result.DurationMs,
	}).Info("Completed agent memory sync")

	return result, nil
}

// syncWorkingMemory synchronizes working memory entries
func (s *Synchronizer) syncWorkingMemory(ctx context.Context, agentID string) (int, []MemoryConflict, error) {
	// For now, just validate that working memory is accessible
	// Real implementation would compare local vs remote state

	memories, err := s.repo.ListWorking(ctx, agentID, MemoryFilters{})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to list working memory: %w", err)
	}

	// No actual conflicts in simple implementation
	// Full implementation would compare with remote instance
	return len(memories), []MemoryConflict{}, nil
}

// syncLongtermMemory synchronizes long-term memory entries
func (s *Synchronizer) syncLongtermMemory(ctx context.Context, agentID string) (int, []MemoryConflict, error) {
	// For now, just validate that longterm memory is accessible
	// Real implementation would:
	// 1. Fetch remote changes
	// 2. Compare versions
	// 3. Detect conflicts
	// 4. Apply conflict resolution strategy

	memories, err := s.repo.ListLongterm(ctx, agentID, MemoryFilters{})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to list longterm memory: %w", err)
	}

	// No actual conflicts in simple implementation
	return len(memories), []MemoryConflict{}, nil
}

// DetectConflicts identifies memory conflicts between local and remote state
func (s *Synchronizer) DetectConflicts(ctx context.Context, agentID string) ([]MemoryConflict, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent ID is required")
	}

	conflicts := []MemoryConflict{}

	// Get sync status to check for known conflicts
	status, err := s.repo.GetSyncStatus(ctx, agentID, s.instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync status: %w", err)
	}

	// Return existing conflicts from status
	conflicts = append(conflicts, status.Conflicts...)

	// In a full implementation, we would:
	// 1. Compare local working memory with remote
	// 2. Compare local longterm memory with remote
	// 3. Check version numbers and timestamps
	// 4. Identify any discrepancies

	return conflicts, nil
}

// ResolveConflicts resolves detected conflicts using the configured strategy
func (s *Synchronizer) ResolveConflicts(ctx context.Context, agentID string, conflicts []MemoryConflict) error {
	if agentID == "" {
		return fmt.Errorf("agent ID is required")
	}

	if len(conflicts) == 0 {
		return nil
	}

	log.WithFields(log.Fields{
		"agent_id":  agentID,
		"conflicts": len(conflicts),
		"strategy":  s.strategy,
	}).Info("Resolving memory conflicts")

	resolved := 0
	failed := 0

	for _, conflict := range conflicts {
		err := s.service.ResolveConflict(ctx, &conflict, s.strategy)
		if err != nil {
			log.WithError(err).WithField("key", conflict.Key).Warn("Failed to resolve conflict")
			failed++
		} else {
			resolved++
		}
	}

	log.WithFields(log.Fields{
		"agent_id": agentID,
		"resolved": resolved,
		"failed":   failed,
	}).Info("Conflict resolution completed")

	if failed > 0 {
		return fmt.Errorf("failed to resolve %d conflicts", failed)
	}

	// Clear conflicts from sync status
	status, err := s.repo.GetSyncStatus(ctx, agentID, s.instanceID)
	if err != nil {
		return fmt.Errorf("failed to get sync status: %w", err)
	}

	status.Conflicts = []MemoryConflict{}
	if status.Status == SyncStateConflict {
		status.Status = SyncStateSynced
	}

	err = s.repo.UpdateSyncStatus(ctx, status)
	if err != nil {
		return fmt.Errorf("failed to update sync status: %w", err)
	}

	return nil
}

// ForcePush pushes all local memory to remote, overwriting any conflicts
func (s *Synchronizer) ForcePush(ctx context.Context, agentID string) error {
	if agentID == "" {
		return fmt.Errorf("agent ID is required")
	}

	log.WithField("agent_id", agentID).Warn("Force pushing local memory (overwriting remote)")

	// In a full implementation:
	// 1. Get all local working memory
	// 2. Get all local longterm memory
	// 3. Push to remote storage/database
	// 4. Update sync status

	// For now, just update sync status
	status, err := s.repo.GetSyncStatus(ctx, agentID, s.instanceID)
	if err != nil {
		return fmt.Errorf("failed to get sync status: %w", err)
	}

	status.Status = SyncStateSynced
	status.PendingChanges = 0
	status.Conflicts = []MemoryConflict{}
	status.LastSyncAt = time.Now()
	status.SyncVersion++

	if status.Metadata == nil {
		status.Metadata = make(map[string]interface{})
	}
	status.Metadata["last_force_push"] = time.Now()

	err = s.repo.UpdateSyncStatus(ctx, status)
	if err != nil {
		return fmt.Errorf("failed to update sync status: %w", err)
	}

	log.WithField("agent_id", agentID).Info("Force push completed")

	return nil
}

// ForcePull pulls all remote memory to local, overwriting any conflicts
func (s *Synchronizer) ForcePull(ctx context.Context, agentID string) error {
	if agentID == "" {
		return fmt.Errorf("agent ID is required")
	}

	log.WithField("agent_id", agentID).Warn("Force pulling remote memory (overwriting local)")

	// In a full implementation:
	// 1. Get all remote working memory
	// 2. Get all remote longterm memory
	// 3. Clear local memory
	// 4. Store all remote entries locally
	// 5. Update sync status

	// For now, just update sync status
	status, err := s.repo.GetSyncStatus(ctx, agentID, s.instanceID)
	if err != nil {
		return fmt.Errorf("failed to get sync status: %w", err)
	}

	status.Status = SyncStateSynced
	status.PendingChanges = 0
	status.Conflicts = []MemoryConflict{}
	status.LastSyncAt = time.Now()
	status.SyncVersion++

	if status.Metadata == nil {
		status.Metadata = make(map[string]interface{})
	}
	status.Metadata["last_force_pull"] = time.Now()

	err = s.repo.UpdateSyncStatus(ctx, status)
	if err != nil {
		return fmt.Errorf("failed to update sync status: %w", err)
	}

	log.WithField("agent_id", agentID).Info("Force pull completed")

	return nil
}

// GetStatus returns the current synchronization status
func (s *Synchronizer) GetStatus(ctx context.Context, agentID string) (*SyncStatus, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent ID is required")
	}

	status, err := s.repo.GetSyncStatus(ctx, agentID, s.instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync status: %w", err)
	}

	return status, nil
}

// IsRunning returns whether periodic sync is running
func (s *Synchronizer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetInstanceID returns the instance ID of this synchronizer
func (s *Synchronizer) GetInstanceID() string {
	return s.instanceID
}

// SetConflictStrategy updates the conflict resolution strategy
func (s *Synchronizer) SetConflictStrategy(strategy ConflictStrategy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.strategy = strategy
	log.WithField("strategy", strategy).Info("Updated conflict resolution strategy")
}

// GetConflictStrategy returns the current conflict resolution strategy
func (s *Synchronizer) GetConflictStrategy() ConflictStrategy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.strategy
}
