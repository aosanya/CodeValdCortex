package memory

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestService_StoreAndRetrieveWorking(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	agentID := "test-agent-1"
	key := "current_task"
	value := "processing-data"
	ttl := 5 * time.Minute

	// Store working memory
	err := service.StoreWorking(ctx, agentID, key, value, ttl)
	if err != nil {
		t.Fatalf("Failed to store working memory: %v", err)
	}

	// Verify repository was called
	if repo.GetCallCount("StoreWorking") != 1 {
		t.Errorf("Expected 1 call to StoreWorking, got %d", repo.GetCallCount("StoreWorking"))
	}

	// Retrieve working memory
	retrieved, err := service.RetrieveWorking(ctx, agentID, key)
	if err != nil {
		t.Fatalf("Failed to retrieve working memory: %v", err)
	}

	if retrieved != value {
		t.Errorf("Expected value %v, got %v", value, retrieved)
	}
}

func TestService_StoreWorkingValidation(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	tests := []struct {
		name    string
		agentID string
		key     string
		wantErr bool
	}{
		{"Empty agent ID", "", "key1", true},
		{"Empty key", "agent1", "", true},
		{"Valid inputs", "agent1", "key1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.StoreWorking(ctx, tt.agentID, tt.key, "value", time.Minute)
			if (err != nil) != tt.wantErr {
				t.Errorf("StoreWorking() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_UpdateWorking(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	agentID := "test-agent-1"
	key := "task_status"

	// Store initial value
	err := service.StoreWorking(ctx, agentID, key, "pending", time.Hour)
	if err != nil {
		t.Fatalf("Failed to store working memory: %v", err)
	}

	// Update value
	err = service.UpdateWorking(ctx, agentID, key, "in_progress")
	if err != nil {
		t.Fatalf("Failed to update working memory: %v", err)
	}

	// Retrieve and verify
	retrieved, err := service.RetrieveWorking(ctx, agentID, key)
	if err != nil {
		t.Fatalf("Failed to retrieve working memory: %v", err)
	}

	if retrieved != "in_progress" {
		t.Errorf("Expected updated value 'in_progress', got %v", retrieved)
	}
}

func TestService_DeleteWorking(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	agentID := "test-agent-1"
	key := "temp_data"

	// Store memory
	err := service.StoreWorking(ctx, agentID, key, "temporary", time.Minute)
	if err != nil {
		t.Fatalf("Failed to store working memory: %v", err)
	}

	// Delete memory
	err = service.DeleteWorking(ctx, agentID, key)
	if err != nil {
		t.Fatalf("Failed to delete working memory: %v", err)
	}

	// Verify deletion
	_, err = service.RetrieveWorking(ctx, agentID, key)
	if err == nil {
		t.Error("Expected error retrieving deleted memory, got nil")
	}
}

func TestService_ClearWorking(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Store multiple memories
	keys := []string{"key1", "key2", "key3"}
	for _, key := range keys {
		err := service.StoreWorking(ctx, agentID, key, fmt.Sprintf("value-%s", key), time.Hour)
		if err != nil {
			t.Fatalf("Failed to store working memory: %v", err)
		}
	}

	// Clear all
	err := service.ClearWorking(ctx, agentID)
	if err != nil {
		t.Fatalf("Failed to clear working memory: %v", err)
	}

	// Verify all cleared
	if repo.GetCallCount("ClearWorking") != 1 {
		t.Errorf("Expected 1 call to ClearWorking, got %d", repo.GetCallCount("ClearWorking"))
	}
}

func TestService_ListWorking(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Store multiple memories
	for i := 0; i < 5; i++ {
		err := service.StoreWorking(ctx, agentID, fmt.Sprintf("key%d", i), i, time.Hour)
		if err != nil {
			t.Fatalf("Failed to store working memory: %v", err)
		}
	}

	// List memories
	memories, err := service.ListWorking(ctx, agentID, MemoryFilters{})
	if err != nil {
		t.Fatalf("Failed to list working memory: %v", err)
	}

	if len(memories) != 5 {
		t.Errorf("Expected 5 memories, got %d", len(memories))
	}
}

func TestService_RememberAndRecall(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	agentID := "test-agent-1"
	key := "learned_skill"
	value := "data_analysis"
	category := "skills"

	metadata := map[string]interface{}{
		"importance": 8,
		"confidence": 0.95,
		"tags":       []string{"skill", "analytics"},
	}

	// Remember
	err := service.Remember(ctx, agentID, key, value, category, metadata)
	if err != nil {
		t.Fatalf("Failed to remember: %v", err)
	}

	// Verify repository was called
	if repo.GetCallCount("StoreLongterm") != 1 {
		t.Errorf("Expected 1 call to StoreLongterm, got %d", repo.GetCallCount("StoreLongterm"))
	}

	// Recall
	recalled, err := service.Recall(ctx, agentID, key)
	if err != nil {
		t.Fatalf("Failed to recall: %v", err)
	}

	if recalled != value {
		t.Errorf("Expected value %v, got %v", value, recalled)
	}
}

func TestService_RememberValidation(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	tests := []struct {
		name    string
		agentID string
		key     string
		wantErr bool
	}{
		{"Empty agent ID", "", "key1", true},
		{"Empty key", "agent1", "", true},
		{"Valid inputs", "agent1", "key1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Remember(ctx, tt.agentID, tt.key, "value", "category", nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Remember() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Forget(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	agentID := "test-agent-1"
	key := "old_info"

	// Remember something
	err := service.Remember(ctx, agentID, key, "outdated", "general", nil)
	if err != nil {
		t.Fatalf("Failed to remember: %v", err)
	}

	// Forget it
	err = service.Forget(ctx, agentID, key)
	if err != nil {
		t.Fatalf("Failed to forget: %v", err)
	}

	// Verify deletion
	_, err = service.Recall(ctx, agentID, key)
	if err == nil {
		t.Error("Expected error recalling forgotten memory, got nil")
	}
}

func TestService_Search(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Store multiple memories in different categories
	memories := []struct {
		key      string
		value    string
		category string
	}{
		{"skill1", "coding", "skills"},
		{"skill2", "testing", "skills"},
		{"fact1", "golang", "facts"},
	}

	for _, mem := range memories {
		err := service.Remember(ctx, agentID, mem.key, mem.value, mem.category, nil)
		if err != nil {
			t.Fatalf("Failed to remember: %v", err)
		}
	}

	// Search by category
	query := MemoryQuery{
		Filters: MemoryFilters{
			Category: "skills",
		},
	}

	results, err := service.Search(ctx, agentID, query)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'skills' category, got %d", len(results))
	}
}

func TestService_Archive(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Store some old memories
	for i := 0; i < 3; i++ {
		err := service.Remember(ctx, agentID, fmt.Sprintf("old%d", i), "old-data", "general", map[string]interface{}{
			"importance": 2, // Low importance
		})
		if err != nil {
			t.Fatalf("Failed to remember: %v", err)
		}
	}

	// Archive with dry run
	criteria := ArchiveCriteria{
		MaxImportance: 3,
		DryRun:        true,
	}

	err := service.Archive(ctx, agentID, criteria)
	if err != nil {
		t.Fatalf("Failed to archive (dry run): %v", err)
	}

	// Verify nothing was actually deleted (dry run)
	memories, err := service.Search(ctx, agentID, MemoryQuery{})
	if err != nil {
		t.Fatalf("Failed to search after dry run: %v", err)
	}

	if len(memories) != 3 {
		t.Errorf("Expected 3 memories after dry run, got %d", len(memories))
	}

	// Archive for real
	criteria.DryRun = false
	err = service.Archive(ctx, agentID, criteria)
	if err != nil {
		t.Fatalf("Failed to archive: %v", err)
	}
}

func TestService_CreateSnapshot(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Store some memories first
	err := service.StoreWorking(ctx, agentID, "task", "current", time.Hour)
	if err != nil {
		t.Fatalf("Failed to store working memory: %v", err)
	}

	err = service.Remember(ctx, agentID, "knowledge", "important", "facts", nil)
	if err != nil {
		t.Fatalf("Failed to remember: %v", err)
	}

	// Create snapshot
	snapshot, err := service.CreateSnapshot(ctx, agentID, "manual", "test snapshot")
	if err != nil {
		t.Fatalf("Failed to create snapshot: %v", err)
	}

	if snapshot == nil {
		t.Fatal("Expected snapshot, got nil")
	}

	if snapshot.AgentID != agentID {
		t.Errorf("Expected agent ID %s, got %s", agentID, snapshot.AgentID)
	}

	if snapshot.SnapshotType != "manual" {
		t.Errorf("Expected snapshot type 'manual', got %s", snapshot.SnapshotType)
	}

	// Verify state contains memory info
	if snapshot.State == nil {
		t.Error("Expected snapshot state, got nil")
	}
}

func TestService_ListSnapshots(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Create multiple snapshots
	types := []string{"periodic", "manual", "pre-update"}
	for _, snapshotType := range types {
		_, err := service.CreateSnapshot(ctx, agentID, snapshotType, "test")
		if err != nil {
			t.Fatalf("Failed to create snapshot: %v", err)
		}
	}

	// List all snapshots
	snapshots, err := service.ListSnapshots(ctx, agentID, SnapshotFilters{})
	if err != nil {
		t.Fatalf("Failed to list snapshots: %v", err)
	}

	if len(snapshots) != 3 {
		t.Errorf("Expected 3 snapshots, got %d", len(snapshots))
	}

	// List with filter
	filteredSnapshots, err := service.ListSnapshots(ctx, agentID, SnapshotFilters{
		SnapshotType: "manual",
	})
	if err != nil {
		t.Fatalf("Failed to list filtered snapshots: %v", err)
	}

	if len(filteredSnapshots) != 1 {
		t.Errorf("Expected 1 manual snapshot, got %d", len(filteredSnapshots))
	}
}

func TestService_DeleteSnapshot(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Create snapshot
	snapshot, err := service.CreateSnapshot(ctx, agentID, "manual", "test")
	if err != nil {
		t.Fatalf("Failed to create snapshot: %v", err)
	}

	// Delete snapshot
	err = service.DeleteSnapshot(ctx, snapshot.ID)
	if err != nil {
		t.Fatalf("Failed to delete snapshot: %v", err)
	}

	// Verify deletion
	if repo.GetCallCount("DeleteSnapshot") != 1 {
		t.Errorf("Expected 1 call to DeleteSnapshot, got %d", repo.GetCallCount("DeleteSnapshot"))
	}
}

func TestService_SyncMemory(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Perform sync
	result, err := service.SyncMemory(ctx, agentID)
	if err != nil {
		t.Fatalf("Failed to sync memory: %v", err)
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

	// Verify repository calls
	if repo.GetCallCount("GetSyncStatus") == 0 {
		t.Error("Expected GetSyncStatus to be called")
	}
}

func TestService_GetSyncStatus(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Get sync status
	status, err := service.GetSyncStatus(ctx, agentID)
	if err != nil {
		t.Fatalf("Failed to get sync status: %v", err)
	}

	if status == nil {
		t.Fatal("Expected sync status, got nil")
	}

	if status.AgentID != agentID {
		t.Errorf("Expected agent ID %s, got %s", agentID, status.AgentID)
	}
}

func TestService_ResolveConflict(t *testing.T) {
	service := NewService(NewMockRepository())
	ctx := context.Background()

	conflict := &MemoryConflict{
		Key:           "test-key",
		MemoryType:    MemoryTypeLongterm,
		LocalVersion:  1,
		RemoteVersion: 2,
		LocalValue:    "local",
		RemoteValue:   "remote",
		LocalTime:     time.Now().Add(-1 * time.Hour),
		RemoteTime:    time.Now(),
		DetectedAt:    time.Now(),
	}

	tests := []struct {
		name     string
		strategy ConflictStrategy
		wantErr  bool
	}{
		{"Last write wins", ConflictStrategyLastWriteWins, false},
		{"Local wins", ConflictStrategyLocalWins, false},
		{"Remote wins", ConflictStrategyRemoteWins, false},
		{"Version based", ConflictStrategyVersionBased, false},
		{"Manual", ConflictStrategyManual, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ResolveConflict(ctx, conflict, tt.strategy)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveConflict() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_CleanupExpired(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	// Cleanup
	count, err := service.CleanupExpired(ctx)
	if err != nil {
		t.Fatalf("Failed to cleanup expired: %v", err)
	}

	if count < 0 {
		t.Errorf("Expected non-negative count, got %d", count)
	}

	// Verify repository was called
	if repo.GetCallCount("CleanupExpired") != 1 {
		t.Errorf("Expected 1 call to CleanupExpired, got %d", repo.GetCallCount("CleanupExpired"))
	}
}

func TestService_GetMemoryStats(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Store some memories
	err := service.StoreWorking(ctx, agentID, "key1", "value1", time.Hour)
	if err != nil {
		t.Fatalf("Failed to store working memory: %v", err)
	}

	err = service.Remember(ctx, agentID, "key2", "value2", "category", nil)
	if err != nil {
		t.Fatalf("Failed to remember: %v", err)
	}

	// Get stats
	stats, err := service.GetMemoryStats(ctx, agentID)
	if err != nil {
		t.Fatalf("Failed to get memory stats: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected stats, got nil")
	}

	if stats.AgentID != agentID {
		t.Errorf("Expected agent ID %s, got %s", agentID, stats.AgentID)
	}

	if stats.WorkingMemoryCount != 1 {
		t.Errorf("Expected 1 working memory, got %d", stats.WorkingMemoryCount)
	}

	if stats.LongtermMemoryCount != 1 {
		t.Errorf("Expected 1 longterm memory, got %d", stats.LongtermMemoryCount)
	}
}

func TestService_ErrorHandling(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	agentID := "test-agent-1"

	// Inject error
	expectedErr := fmt.Errorf("repository error")
	repo.SetError("StoreWorking", expectedErr)

	// Try to store
	err := service.StoreWorking(ctx, agentID, "key", "value", time.Minute)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Reset error
	repo.Reset()

	// Should work now
	err = service.StoreWorking(ctx, agentID, "key", "value", time.Minute)
	if err != nil {
		t.Errorf("Expected no error after reset, got %v", err)
	}
}

func TestService_RetrieveWorkingExpired(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	agentID := "test-agent-1"
	key := "expired_key"

	// Store with very short TTL
	err := service.StoreWorking(ctx, agentID, key, "value", 1*time.Nanosecond)
	if err != nil {
		t.Fatalf("Failed to store working memory: %v", err)
	}

	// Wait a bit to ensure expiration
	time.Sleep(10 * time.Millisecond)

	// Try to retrieve - should fail due to expiration
	_, err = service.RetrieveWorking(ctx, agentID, key)
	if err == nil {
		t.Error("Expected error for expired memory, got nil")
	}
}
