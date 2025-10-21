package memory

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestSynchronizer_NewSynchronizer(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	sync := NewSynchronizer(service, repo, ConflictStrategyLastWriteWins, time.Minute)

	if sync == nil {
		t.Fatal("Expected synchronizer, got nil")
	}

	if sync.instanceID == "" {
		t.Error("Expected instance ID to be set")
	}

	if sync.strategy != ConflictStrategyLastWriteWins {
		t.Errorf("Expected strategy %s, got %s", ConflictStrategyLastWriteWins, sync.strategy)
	}

	if sync.syncInterval != time.Minute {
		t.Errorf("Expected sync interval %v, got %v", time.Minute, sync.syncInterval)
	}
}

func TestSynchronizer_NewSynchronizerDefaults(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	// Test default sync interval
	sync := NewSynchronizer(service, repo, ConflictStrategyVersionBased, 0)

	if sync.syncInterval != 5*time.Minute {
		t.Errorf("Expected default sync interval 5m, got %v", sync.syncInterval)
	}
}

func TestSynchronizer_GetSetters(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	sync := NewSynchronizer(service, repo, ConflictStrategyLastWriteWins, time.Minute)

	// Test GetInstanceID
	instanceID := sync.GetInstanceID()
	if instanceID == "" {
		t.Error("Expected instance ID, got empty string")
	}

	// Test IsRunning (should be false initially)
	if sync.IsRunning() {
		t.Error("Expected IsRunning to be false initially")
	}

	// Test GetConflictStrategy
	strategy := sync.GetConflictStrategy()
	if strategy != ConflictStrategyLastWriteWins {
		t.Errorf("Expected strategy %s, got %s", ConflictStrategyLastWriteWins, strategy)
	}

	// Test SetConflictStrategy
	sync.SetConflictStrategy(ConflictStrategyRemoteWins)
	newStrategy := sync.GetConflictStrategy()
	if newStrategy != ConflictStrategyRemoteWins {
		t.Errorf("Expected strategy %s after set, got %s", ConflictStrategyRemoteWins, newStrategy)
	}
}

func TestSynchronizer_SyncAgent(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	sync := NewSynchronizer(service, repo, ConflictStrategyLastWriteWins, time.Minute)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Store some test data
	err := service.StoreWorking(ctx, agentID, "key1", "value1", time.Hour)
	if err != nil {
		t.Fatalf("Failed to store working memory: %v", err)
	}

	err = service.Remember(ctx, agentID, "key2", "value2", "category", nil)
	if err != nil {
		t.Fatalf("Failed to remember: %v", err)
	}

	// Sync agent
	result, err := sync.SyncAgent(ctx, agentID)
	if err != nil {
		t.Fatalf("Failed to sync agent: %v", err)
	}

	if result == nil {
		t.Fatal("Expected sync result, got nil")
	}

	if result.AgentID != agentID {
		t.Errorf("Expected agent ID %s, got %s", agentID, result.AgentID)
	}

	if !result.Success {
		t.Error("Expected successful sync")
	}

	if result.ItemsSynced != 2 {
		t.Errorf("Expected 2 items synced, got %d", result.ItemsSynced)
	}

	if len(result.Conflicts) != 0 {
		t.Errorf("Expected 0 conflicts, got %d", len(result.Conflicts))
	}
}

func TestSynchronizer_SyncAgentValidation(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	sync := NewSynchronizer(service, repo, ConflictStrategyLastWriteWins, time.Minute)
	ctx := context.Background()

	// Test empty agent ID
	_, err := sync.SyncAgent(ctx, "")
	if err == nil {
		t.Error("Expected error for empty agent ID, got nil")
	}
}

func TestSynchronizer_DetectConflicts(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	sync := NewSynchronizer(service, repo, ConflictStrategyLastWriteWins, time.Minute)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Detect conflicts (should be none initially)
	conflicts, err := sync.DetectConflicts(ctx, agentID)
	if err != nil {
		t.Fatalf("Failed to detect conflicts: %v", err)
	}

	if len(conflicts) != 0 {
		t.Errorf("Expected 0 conflicts, got %d", len(conflicts))
	}
}

func TestSynchronizer_ResolveConflicts(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	sync := NewSynchronizer(service, repo, ConflictStrategyLastWriteWins, time.Minute)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Create test conflicts
	conflicts := []MemoryConflict{
		{
			Key:           "key1",
			MemoryType:    MemoryTypeLongterm,
			LocalVersion:  1,
			RemoteVersion: 2,
			LocalValue:    "local",
			RemoteValue:   "remote",
			LocalTime:     time.Now().Add(-1 * time.Hour),
			RemoteTime:    time.Now(),
			DetectedAt:    time.Now(),
		},
	}

	// Resolve conflicts
	err := sync.ResolveConflicts(ctx, agentID, conflicts)
	if err != nil {
		t.Fatalf("Failed to resolve conflicts: %v", err)
	}

	// Verify status was updated
	status, err := sync.GetStatus(ctx, agentID)
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if len(status.Conflicts) != 0 {
		t.Errorf("Expected conflicts to be cleared, got %d", len(status.Conflicts))
	}
}

func TestSynchronizer_ResolveConflictsEmpty(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	sync := NewSynchronizer(service, repo, ConflictStrategyLastWriteWins, time.Minute)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Resolve empty conflicts list (should succeed)
	err := sync.ResolveConflicts(ctx, agentID, []MemoryConflict{})
	if err != nil {
		t.Errorf("Expected no error for empty conflicts, got %v", err)
	}
}

func TestSynchronizer_ForcePush(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	sync := NewSynchronizer(service, repo, ConflictStrategyLastWriteWins, time.Minute)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Force push
	err := sync.ForcePush(ctx, agentID)
	if err != nil {
		t.Fatalf("Failed to force push: %v", err)
	}

	// Verify status
	status, err := sync.GetStatus(ctx, agentID)
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if status.Status != SyncStateSynced {
		t.Errorf("Expected status %s, got %s", SyncStateSynced, status.Status)
	}

	if status.PendingChanges != 0 {
		t.Errorf("Expected 0 pending changes, got %d", status.PendingChanges)
	}

	if len(status.Conflicts) != 0 {
		t.Errorf("Expected 0 conflicts, got %d", len(status.Conflicts))
	}

	// Verify metadata
	if status.Metadata == nil {
		t.Error("Expected metadata to be set")
	} else {
		if _, ok := status.Metadata["last_force_push"]; !ok {
			t.Error("Expected last_force_push in metadata")
		}
	}
}

func TestSynchronizer_ForcePull(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	sync := NewSynchronizer(service, repo, ConflictStrategyLastWriteWins, time.Minute)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Force pull
	err := sync.ForcePull(ctx, agentID)
	if err != nil {
		t.Fatalf("Failed to force pull: %v", err)
	}

	// Verify status
	status, err := sync.GetStatus(ctx, agentID)
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if status.Status != SyncStateSynced {
		t.Errorf("Expected status %s, got %s", SyncStateSynced, status.Status)
	}

	// Verify metadata
	if status.Metadata == nil {
		t.Error("Expected metadata to be set")
	} else {
		if _, ok := status.Metadata["last_force_pull"]; !ok {
			t.Error("Expected last_force_pull in metadata")
		}
	}
}

func TestSynchronizer_GetStatus(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	sync := NewSynchronizer(service, repo, ConflictStrategyLastWriteWins, time.Minute)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Get status
	status, err := sync.GetStatus(ctx, agentID)
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if status == nil {
		t.Fatal("Expected status, got nil")
	}

	if status.AgentID != agentID {
		t.Errorf("Expected agent ID %s, got %s", agentID, status.AgentID)
	}

	if status.InstanceID != sync.GetInstanceID() {
		t.Errorf("Expected instance ID %s, got %s", sync.GetInstanceID(), status.InstanceID)
	}
}

func TestSynchronizer_PeriodicSync(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	// Use short interval for testing
	sync := NewSynchronizer(service, repo, ConflictStrategyLastWriteWins, 50*time.Millisecond)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Store some test data
	err := service.StoreWorking(ctx, agentID, "key1", "value1", time.Hour)
	if err != nil {
		t.Fatalf("Failed to store working memory: %v", err)
	}

	// Start periodic sync
	err = sync.StartPeriodicSync(ctx, agentID)
	if err != nil {
		t.Fatalf("Failed to start periodic sync: %v", err)
	}

	// Verify it's running
	if !sync.IsRunning() {
		t.Error("Expected IsRunning to be true after start")
	}

	// Let it run for a bit
	time.Sleep(150 * time.Millisecond)

	// Stop periodic sync
	err = sync.StopPeriodicSync()
	if err != nil {
		t.Fatalf("Failed to stop periodic sync: %v", err)
	}

	// Verify it's stopped
	if sync.IsRunning() {
		t.Error("Expected IsRunning to be false after stop")
	}
}

func TestSynchronizer_PeriodicSyncDoubleStart(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	sync := NewSynchronizer(service, repo, ConflictStrategyLastWriteWins, time.Minute)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Start periodic sync
	err := sync.StartPeriodicSync(ctx, agentID)
	if err != nil {
		t.Fatalf("Failed to start periodic sync: %v", err)
	}
	defer sync.StopPeriodicSync()

	// Try to start again (should fail)
	err = sync.StartPeriodicSync(ctx, agentID)
	if err == nil {
		t.Error("Expected error starting periodic sync twice, got nil")
	}
}

func TestSynchronizer_PeriodicSyncStopWithoutStart(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	sync := NewSynchronizer(service, repo, ConflictStrategyLastWriteWins, time.Minute)

	// Try to stop without starting (should fail)
	err := sync.StopPeriodicSync()
	if err == nil {
		t.Error("Expected error stopping periodic sync without start, got nil")
	}
}

func TestSynchronizer_ConcurrentOperations(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	sync := NewSynchronizer(service, repo, ConflictStrategyLastWriteWins, time.Minute)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Perform multiple operations concurrently
	done := make(chan bool, 3)

	// Goroutine 1: Sync agent
	go func() {
		_, err := sync.SyncAgent(ctx, agentID)
		if err != nil {
			t.Errorf("SyncAgent failed: %v", err)
		}
		done <- true
	}()

	// Goroutine 2: Get status
	go func() {
		_, err := sync.GetStatus(ctx, agentID)
		if err != nil {
			t.Errorf("GetStatus failed: %v", err)
		}
		done <- true
	}()

	// Goroutine 3: Detect conflicts
	go func() {
		_, err := sync.DetectConflicts(ctx, agentID)
		if err != nil {
			t.Errorf("DetectConflicts failed: %v", err)
		}
		done <- true
	}()

	// Wait for all operations
	timeout := time.After(2 * time.Second)
	for i := 0; i < 3; i++ {
		select {
		case <-done:
			// Operation completed
		case <-timeout:
			t.Fatal("Concurrent operations timed out")
		}
	}
}

func TestSynchronizer_ErrorHandling(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	sync := NewSynchronizer(service, repo, ConflictStrategyLastWriteWins, time.Minute)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Inject error into repository
	repo.SetError("GetSyncStatus", fmt.Errorf("repository error"))

	// Try to sync (should handle error gracefully)
	result, err := sync.SyncAgent(ctx, agentID)
	if err == nil {
		t.Error("Expected error from sync with repository error")
	}

	if result == nil {
		t.Fatal("Expected result even with error, got nil")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected errors in result")
	}

	if result.Success {
		t.Error("Expected sync to not be successful with errors")
	}
}

func TestSynchronizer_StrategyChange(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	sync := NewSynchronizer(service, repo, ConflictStrategyLastWriteWins, time.Minute)

	// Test all strategies
	strategies := []ConflictStrategy{
		ConflictStrategyLastWriteWins,
		ConflictStrategyVersionBased,
		ConflictStrategyManual,
		ConflictStrategyLocalWins,
		ConflictStrategyRemoteWins,
	}

	for _, strategy := range strategies {
		sync.SetConflictStrategy(strategy)
		got := sync.GetConflictStrategy()
		if got != strategy {
			t.Errorf("Expected strategy %s, got %s", strategy, got)
		}
	}
}
