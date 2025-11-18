package memory

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/config"
	"github.com/aosanya/CodeValdCortex/internal/database"
)

// skipIfNoDatabase skips the test if ArangoDB is not available
func skipIfNoDatabase(t *testing.T) *database.ArangoClient {
	host := os.Getenv("ARANGO_HOST")
	if host == "" {
		host = "localhost"
	}

	portStr := os.Getenv("ARANGO_PORT")
	port := 8529
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	dbName := os.Getenv("ARANGO_TEST_DB")
	if dbName == "" {
		dbName = "codeval_cortex_test"
	}

	user := os.Getenv("ARANGO_USER")
	if user == "" {
		user = "root"
	}

	password := os.Getenv("ARANGO_PASSWORD")
	if password == "" {
		password = ""
	}

	cfg := &config.DatabaseConfig{
		Host:     host,
		Port:     port,
		Database: dbName,
		Username: user,
		Password: password,
	}

	client, err := database.NewArangoClient(cfg)
	if err != nil {
		t.Skipf("Skipping test: ArangoDB not available: %v", err)
		return nil
	}

	return client
}

// cleanupTestData removes all test data from collections
func cleanupTestData(_ *testing.T, repo *Repository) {
	ctx := context.Background()

	// Delete all documents from test collections
	if repo.workingMemCol != nil {
		repo.workingMemCol.Truncate(ctx)
	}
	if repo.longtermMemCol != nil {
		repo.longtermMemCol.Truncate(ctx)
	}
	if repo.snapshotsCol != nil {
		repo.snapshotsCol.Truncate(ctx)
	}
	if repo.syncStatusCol != nil {
		repo.syncStatusCol.Truncate(ctx)
	}
}

// TestRepository_StoreAndGetWorkingMemory tests working memory creation and retrieval
func TestRepository_StoreAndGetWorkingMemory(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()
	expiresAt := now.Add(1 * time.Hour)

	mem := &WorkingMemory{
		AgentID: "agent-1",
		Key:     "current_task",
		Value: map[string]interface{}{
			"task_id": "task-123",
			"status":  "in_progress",
		},
		Metadata: map[string]interface{}{
			"tags": []string{"task", "active"},
		},
		ExpiresAt: expiresAt,
	}

	// Store working memory
	err = repo.StoreWorking(ctx, mem)
	if err != nil {
		t.Fatalf("Failed to store working memory: %v", err)
	}

	if mem.ID == "" {
		t.Error("Memory ID should be set after creation")
	}

	// Get working memory
	retrieved, err := repo.GetWorking(ctx, mem.AgentID, mem.Key)
	if err != nil {
		t.Fatalf("Failed to get working memory: %v", err)
	}

	if retrieved.ID != mem.ID {
		t.Errorf("ID = %v, want %v", retrieved.ID, mem.ID)
	}
	if retrieved.AgentID != mem.AgentID {
		t.Errorf("AgentID = %v, want %v", retrieved.AgentID, mem.AgentID)
	}
	if retrieved.Key != mem.Key {
		t.Errorf("Key = %v, want %v", retrieved.Key, mem.Key)
	}
	if retrieved.Version != 1 {
		t.Errorf("Version = %v, want 1", retrieved.Version)
	}
	if retrieved.AccessCount != 1 {
		t.Errorf("AccessCount = %v, want 1 (should increment on get)", retrieved.AccessCount)
	}
}

// TestRepository_UpdateWorkingMemory tests updating working memory
func TestRepository_UpdateWorkingMemory(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()

	mem := &WorkingMemory{
		AgentID: "agent-1",
		Key:     "counter",
		Value:   1,
		Metadata: map[string]interface{}{
			"source": "test",
		},
		ExpiresAt: now.Add(1 * time.Hour),
	}

	// Store initial memory
	if err := repo.StoreWorking(ctx, mem); err != nil {
		t.Fatalf("Failed to store working memory: %v", err)
	}

	// Update value
	mem.Value = 2
	err = repo.UpdateWorking(ctx, mem)
	if err != nil {
		t.Fatalf("Failed to update working memory: %v", err)
	}

	// Retrieve and verify
	retrieved, err := repo.GetWorking(ctx, mem.AgentID, mem.Key)
	if err != nil {
		t.Fatalf("Failed to get working memory: %v", err)
	}

	// ArangoDB may return numbers as float64
	var retrievedValue float64
	switch v := retrieved.Value.(type) {
	case int:
		retrievedValue = float64(v)
	case float64:
		retrievedValue = v
	default:
		t.Fatalf("Unexpected value type: %T", retrieved.Value)
	}

	if retrievedValue != 2.0 {
		t.Errorf("Value = %v, want 2", retrievedValue)
	}
	if retrieved.Version < 2 {
		t.Errorf("Version = %v, want >= 2", retrieved.Version)
	}
}

// TestRepository_DeleteWorkingMemory tests deleting working memory
func TestRepository_DeleteWorkingMemory(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()

	mem := &WorkingMemory{
		AgentID:   "agent-1",
		Key:       "temp_data",
		Value:     "test",
		ExpiresAt: now.Add(1 * time.Hour),
		Metadata:  make(map[string]interface{}),
	}

	// Store memory
	if err := repo.StoreWorking(ctx, mem); err != nil {
		t.Fatalf("Failed to store working memory: %v", err)
	}

	// Delete memory
	err = repo.DeleteWorking(ctx, mem.AgentID, mem.Key)
	if err != nil {
		t.Fatalf("Failed to delete working memory: %v", err)
	}

	// Verify deletion
	_, err = repo.GetWorking(ctx, mem.AgentID, mem.Key)
	if err == nil {
		t.Error("Expected error when getting deleted memory")
	}
}

// TestRepository_ListWorkingMemory tests listing working memory with filters
func TestRepository_ListWorkingMemory(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()

	// Create test memories
	memories := []*WorkingMemory{
		{
			AgentID:   "agent-1",
			Key:       "task-1",
			Value:     "data-1",
			ExpiresAt: now.Add(1 * time.Hour),
			Metadata: map[string]interface{}{
				"tags": []interface{}{"task", "priority"},
			},
		},
		{
			AgentID:   "agent-1",
			Key:       "task-2",
			Value:     "data-2",
			ExpiresAt: now.Add(1 * time.Hour),
			Metadata: map[string]interface{}{
				"tags": []interface{}{"task"},
			},
		},
		{
			AgentID:   "agent-1",
			Key:       "metric-1",
			Value:     "data-3",
			ExpiresAt: now.Add(1 * time.Hour),
			Metadata: map[string]interface{}{
				"tags": []interface{}{"metric"},
			},
		},
		{
			AgentID:   "agent-2",
			Key:       "task-1",
			Value:     "data-4",
			ExpiresAt: now.Add(1 * time.Hour),
			Metadata:  map[string]interface{}{},
		},
	}

	for _, mem := range memories {
		if err := repo.StoreWorking(ctx, mem); err != nil {
			t.Fatalf("Failed to store working memory: %v", err)
		}
	}

	// List all memories for agent-1
	list, err := repo.ListWorking(ctx, "agent-1", MemoryFilters{})
	if err != nil {
		t.Fatalf("Failed to list working memory: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 memories for agent-1, got %d", len(list))
	}

	// List with tag filter
	list, err = repo.ListWorking(ctx, "agent-1", MemoryFilters{
		Tags: []string{"priority"},
	})
	if err != nil {
		t.Fatalf("Failed to list working memory with filter: %v", err)
	}

	if len(list) != 1 {
		t.Errorf("Expected 1 memory with 'priority' tag, got %d", len(list))
	}

	// List with limit
	list, err = repo.ListWorking(ctx, "agent-1", MemoryFilters{
		Limit: 2,
	})
	if err != nil {
		t.Fatalf("Failed to list working memory with limit: %v", err)
	}

	if len(list) > 2 {
		t.Errorf("Expected max 2 memories with limit, got %d", len(list))
	}
}

// TestRepository_ClearWorkingMemory tests clearing all working memory for an agent
func TestRepository_ClearWorkingMemory(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()

	// Create test memories
	for i := 0; i < 5; i++ {
		mem := &WorkingMemory{
			AgentID:   "agent-1",
			Key:       "key-" + strconv.Itoa(i),
			Value:     "data",
			ExpiresAt: now.Add(1 * time.Hour),
			Metadata:  make(map[string]interface{}),
		}
		if err := repo.StoreWorking(ctx, mem); err != nil {
			t.Fatalf("Failed to store working memory: %v", err)
		}
	}

	// Clear all memories
	err = repo.ClearWorking(ctx, "agent-1")
	if err != nil {
		t.Fatalf("Failed to clear working memory: %v", err)
	}

	// Verify cleared
	list, err := repo.ListWorking(ctx, "agent-1", MemoryFilters{})
	if err != nil {
		t.Fatalf("Failed to list working memory: %v", err)
	}

	if len(list) != 0 {
		t.Errorf("Expected 0 memories after clear, got %d", len(list))
	}
}

// TestRepository_StoreAndGetLongtermMemory tests long-term memory creation and retrieval
func TestRepository_StoreAndGetLongtermMemory(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()

	mem := &LongtermMemory{
		AgentID:  "agent-1",
		Category: "skill",
		Key:      "data_processing",
		Value: map[string]interface{}{
			"proficiency": 0.85,
			"examples":    []string{"example1", "example2"},
		},
		Metadata: MemoryMetadata{
			Source:     "training",
			Importance: 8,
			Confidence: 0.92,
			Tags:       []string{"skill", "data"},
		},
	}

	// Store long-term memory
	err = repo.StoreLongterm(ctx, mem)
	if err != nil {
		t.Fatalf("Failed to store longterm memory: %v", err)
	}

	if mem.ID == "" {
		t.Error("Memory ID should be set after creation")
	}

	// Get long-term memory
	retrieved, err := repo.GetLongterm(ctx, mem.AgentID, mem.Key)
	if err != nil {
		t.Fatalf("Failed to get longterm memory: %v", err)
	}

	if retrieved.ID != mem.ID {
		t.Errorf("ID = %v, want %v", retrieved.ID, mem.ID)
	}
	if retrieved.AgentID != mem.AgentID {
		t.Errorf("AgentID = %v, want %v", retrieved.AgentID, mem.AgentID)
	}
	if retrieved.Category != mem.Category {
		t.Errorf("Category = %v, want %v", retrieved.Category, mem.Category)
	}
	if retrieved.Metadata.Importance != 8 {
		t.Errorf("Importance = %v, want 8", retrieved.Metadata.Importance)
	}
	if retrieved.AccessCount != 1 {
		t.Errorf("AccessCount = %v, want 1 (should increment on get)", retrieved.AccessCount)
	}
}

// TestRepository_UpdateLongtermMemory tests updating long-term memory
func TestRepository_UpdateLongtermMemory(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()

	mem := &LongtermMemory{
		AgentID:  "agent-1",
		Category: "knowledge",
		Key:      "fact-1",
		Value:    "initial value",
		Metadata: MemoryMetadata{
			Importance: 5,
			Confidence: 0.7,
			Tags:       []string{"fact"},
		},
	}

	// Store initial memory
	if err := repo.StoreLongterm(ctx, mem); err != nil {
		t.Fatalf("Failed to store longterm memory: %v", err)
	}

	// Update memory
	mem.Value = "updated value"
	mem.Metadata.Importance = 9
	mem.Metadata.Confidence = 0.95
	err = repo.UpdateLongterm(ctx, mem)
	if err != nil {
		t.Fatalf("Failed to update longterm memory: %v", err)
	}

	// Retrieve and verify
	retrieved, err := repo.GetLongterm(ctx, mem.AgentID, mem.Key)
	if err != nil {
		t.Fatalf("Failed to get longterm memory: %v", err)
	}

	if retrieved.Value != "updated value" {
		t.Errorf("Value = %v, want 'updated value'", retrieved.Value)
	}
	if retrieved.Metadata.Importance != 9 {
		t.Errorf("Importance = %v, want 9", retrieved.Metadata.Importance)
	}
	if retrieved.Version < 2 {
		t.Errorf("Version = %v, want >= 2", retrieved.Version)
	}
}

// TestRepository_ListLongtermMemory tests listing long-term memory with filters
func TestRepository_ListLongtermMemory(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()

	// Create test memories
	memories := []*LongtermMemory{
		{
			AgentID:  "agent-1",
			Category: "skill",
			Key:      "skill-1",
			Value:    "data",
			Metadata: MemoryMetadata{
				Importance: 9,
				Tags:       []string{"skill", "critical"},
			},
		},
		{
			AgentID:  "agent-1",
			Category: "skill",
			Key:      "skill-2",
			Value:    "data",
			Metadata: MemoryMetadata{
				Importance: 5,
				Tags:       []string{"skill"},
			},
		},
		{
			AgentID:  "agent-1",
			Category: "knowledge",
			Key:      "fact-1",
			Value:    "data",
			Metadata: MemoryMetadata{
				Importance: 7,
				Tags:       []string{"fact"},
			},
		},
		{
			AgentID:  "agent-2",
			Category: "skill",
			Key:      "skill-1",
			Value:    "data",
			Metadata: MemoryMetadata{
				Importance: 8,
				Tags:       []string{"skill"},
			},
		},
	}

	for _, mem := range memories {
		if err := repo.StoreLongterm(ctx, mem); err != nil {
			t.Fatalf("Failed to store longterm memory: %v", err)
		}
	}

	// List all memories for agent-1
	list, err := repo.ListLongterm(ctx, "agent-1", MemoryFilters{})
	if err != nil {
		t.Fatalf("Failed to list longterm memory: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 memories for agent-1, got %d", len(list))
	}

	// List by category
	list, err = repo.ListLongterm(ctx, "agent-1", MemoryFilters{
		Category: "skill",
	})
	if err != nil {
		t.Fatalf("Failed to list longterm memory by category: %v", err)
	}

	if len(list) != 2 {
		t.Errorf("Expected 2 skill memories, got %d", len(list))
	}

	// List by importance
	list, err = repo.ListLongterm(ctx, "agent-1", MemoryFilters{
		MinImportance: 7,
	})
	if err != nil {
		t.Fatalf("Failed to list longterm memory by importance: %v", err)
	}

	if len(list) < 2 {
		t.Errorf("Expected at least 2 memories with importance >= 7, got %d", len(list))
	}

	// List by tags
	list, err = repo.ListLongterm(ctx, "agent-1", MemoryFilters{
		Tags: []string{"critical"},
	})
	if err != nil {
		t.Fatalf("Failed to list longterm memory by tags: %v", err)
	}

	if len(list) != 1 {
		t.Errorf("Expected 1 memory with 'critical' tag, got %d", len(list))
	}
}

// TestRepository_CreateAndGetSnapshot tests snapshot creation and retrieval
func TestRepository_CreateAndGetSnapshot(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()

	snapshot := &StateSnapshot{
		AgentID:      "agent-1",
		SnapshotType: "manual",
		State: map[string]interface{}{
			"current_task": "task-123",
			"status":       "running",
			"memory": map[string]interface{}{
				"working": []string{"key1", "key2"},
			},
		},
		Metadata: SnapshotMetadata{
			Trigger:    "manual",
			Reason:     "backup before update",
			Compressed: false,
		},
		ExpiresAt: now.Add(30 * 24 * time.Hour),
	}

	// Create snapshot
	err = repo.CreateSnapshot(ctx, snapshot)
	if err != nil {
		t.Fatalf("Failed to create snapshot: %v", err)
	}

	if snapshot.ID == "" {
		t.Error("Snapshot ID should be set after creation")
	}
	if snapshot.Checksum == "" {
		t.Error("Checksum should be calculated")
	}
	if snapshot.Metadata.SizeBytes == 0 {
		t.Error("Size should be calculated")
	}

	// Get snapshot
	retrieved, err := repo.GetSnapshot(ctx, snapshot.ID)
	if err != nil {
		t.Fatalf("Failed to get snapshot: %v", err)
	}

	if retrieved.ID != snapshot.ID {
		t.Errorf("ID = %v, want %v", retrieved.ID, snapshot.ID)
	}
	if retrieved.AgentID != snapshot.AgentID {
		t.Errorf("AgentID = %v, want %v", retrieved.AgentID, snapshot.AgentID)
	}
	if retrieved.SnapshotType != snapshot.SnapshotType {
		t.Errorf("SnapshotType = %v, want %v", retrieved.SnapshotType, snapshot.SnapshotType)
	}
	if retrieved.Checksum != snapshot.Checksum {
		t.Errorf("Checksum = %v, want %v", retrieved.Checksum, snapshot.Checksum)
	}
}

// TestRepository_ListSnapshots tests listing snapshots with filters
func TestRepository_ListSnapshots(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()

	// Create test snapshots
	snapshots := []*StateSnapshot{
		{
			AgentID:      "agent-1",
			SnapshotType: "periodic",
			State:        map[string]interface{}{"data": "1"},
			Metadata: SnapshotMetadata{
				Trigger: "scheduled",
				Reason:  "hourly backup",
			},
			ExpiresAt: now.Add(30 * 24 * time.Hour),
		},
		{
			AgentID:      "agent-1",
			SnapshotType: "manual",
			State:        map[string]interface{}{"data": "2"},
			Metadata: SnapshotMetadata{
				Trigger: "user",
				Reason:  "before update",
			},
			ExpiresAt: now.Add(30 * 24 * time.Hour),
		},
		{
			AgentID:      "agent-1",
			SnapshotType: "pre-shutdown",
			State:        map[string]interface{}{"data": "3"},
			Metadata: SnapshotMetadata{
				Trigger: "system",
				Reason:  "graceful shutdown",
			},
			ExpiresAt: now.Add(30 * 24 * time.Hour),
		},
		{
			AgentID:      "agent-2",
			SnapshotType: "periodic",
			State:        map[string]interface{}{"data": "4"},
			Metadata: SnapshotMetadata{
				Trigger: "scheduled",
				Reason:  "hourly backup",
			},
			ExpiresAt: now.Add(30 * 24 * time.Hour),
		},
	}

	for _, snap := range snapshots {
		if err := repo.CreateSnapshot(ctx, snap); err != nil {
			t.Fatalf("Failed to create snapshot: %v", err)
		}
	}

	// List all snapshots for agent-1
	list, err := repo.ListSnapshots(ctx, "agent-1", SnapshotFilters{})
	if err != nil {
		t.Fatalf("Failed to list snapshots: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 snapshots for agent-1, got %d", len(list))
	}

	// List by type
	list, err = repo.ListSnapshots(ctx, "agent-1", SnapshotFilters{
		SnapshotType: "manual",
	})
	if err != nil {
		t.Fatalf("Failed to list snapshots by type: %v", err)
	}

	if len(list) != 1 {
		t.Errorf("Expected 1 manual snapshot, got %d", len(list))
	}

	// List with limit
	list, err = repo.ListSnapshots(ctx, "agent-1", SnapshotFilters{
		Limit: 2,
	})
	if err != nil {
		t.Fatalf("Failed to list snapshots with limit: %v", err)
	}

	if len(list) > 2 {
		t.Errorf("Expected max 2 snapshots with limit, got %d", len(list))
	}
}

// TestRepository_DeleteSnapshot tests deleting a snapshot
func TestRepository_DeleteSnapshot(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()

	snapshot := &StateSnapshot{
		AgentID:      "agent-1",
		SnapshotType: "manual",
		State:        map[string]interface{}{"data": "test"},
		Metadata: SnapshotMetadata{
			Trigger: "test",
			Reason:  "test",
		},
		ExpiresAt: now.Add(1 * time.Hour),
	}

	// Create snapshot
	if err := repo.CreateSnapshot(ctx, snapshot); err != nil {
		t.Fatalf("Failed to create snapshot: %v", err)
	}

	// Delete snapshot
	err = repo.DeleteSnapshot(ctx, snapshot.ID)
	if err != nil {
		t.Fatalf("Failed to delete snapshot: %v", err)
	}

	// Verify deletion
	_, err = repo.GetSnapshot(ctx, snapshot.ID)
	if err == nil {
		t.Error("Expected error when getting deleted snapshot")
	}
}

// TestRepository_SyncStatus tests sync status operations
func TestRepository_SyncStatus(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()

	status := &SyncStatus{
		AgentID:        "agent-1",
		InstanceID:     "instance-1",
		LastSyncAt:     now,
		SyncVersion:    1,
		PendingChanges: 0,
		Conflicts:      []MemoryConflict{},
		Status:         SyncStateSynced,
		Metadata: map[string]interface{}{
			"node": "k8s-node-1",
		},
	}

	// Update (upsert) sync status
	err = repo.UpdateSyncStatus(ctx, status)
	if err != nil {
		t.Fatalf("Failed to update sync status: %v", err)
	}

	// Get sync status
	retrieved, err := repo.GetSyncStatus(ctx, status.AgentID, status.InstanceID)
	if err != nil {
		t.Fatalf("Failed to get sync status: %v", err)
	}

	if retrieved.AgentID != status.AgentID {
		t.Errorf("AgentID = %v, want %v", retrieved.AgentID, status.AgentID)
	}
	if retrieved.Status != SyncStateSynced {
		t.Errorf("Status = %v, want %v", retrieved.Status, SyncStateSynced)
	}

	// Update with conflict
	status.Status = SyncStateConflict
	status.PendingChanges = 2
	status.Conflicts = []MemoryConflict{
		{
			Key:           "test-key",
			MemoryType:    MemoryTypeWorking,
			LocalVersion:  2,
			RemoteVersion: 3,
			LocalValue:    "local",
			RemoteValue:   "remote",
			DetectedAt:    now,
		},
	}

	err = repo.UpdateSyncStatus(ctx, status)
	if err != nil {
		t.Fatalf("Failed to update sync status with conflict: %v", err)
	}

	// Verify conflict
	retrieved, err = repo.GetSyncStatus(ctx, status.AgentID, status.InstanceID)
	if err != nil {
		t.Fatalf("Failed to get sync status after conflict: %v", err)
	}

	if retrieved.Status != SyncStateConflict {
		t.Errorf("Status = %v, want conflict", retrieved.Status)
	}
	if len(retrieved.Conflicts) != 1 {
		t.Errorf("Conflicts count = %v, want 1", len(retrieved.Conflicts))
	}
	if retrieved.PendingChanges != 2 {
		t.Errorf("PendingChanges = %v, want 2", retrieved.PendingChanges)
	}
}

// TestRepository_CleanupExpired tests cleanup of expired memories and snapshots
func TestRepository_CleanupExpired(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	// Create expired working memory
	expiredWorking := &WorkingMemory{
		AgentID:   "agent-1",
		Key:       "expired-1",
		Value:     "test",
		ExpiresAt: past,
		Metadata:  make(map[string]interface{}),
	}
	if err := repo.StoreWorking(ctx, expiredWorking); err != nil {
		t.Fatalf("Failed to store expired working memory: %v", err)
	}

	// Create non-expired working memory
	validWorking := &WorkingMemory{
		AgentID:   "agent-1",
		Key:       "valid-1",
		Value:     "test",
		ExpiresAt: future,
		Metadata:  make(map[string]interface{}),
	}
	if err := repo.StoreWorking(ctx, validWorking); err != nil {
		t.Fatalf("Failed to store valid working memory: %v", err)
	}

	// Create expired snapshot
	expiredSnapshot := &StateSnapshot{
		AgentID:      "agent-1",
		SnapshotType: "test",
		State:        map[string]interface{}{"data": "test"},
		Metadata:     SnapshotMetadata{Trigger: "test", Reason: "test"},
		ExpiresAt:    past,
	}
	if err := repo.CreateSnapshot(ctx, expiredSnapshot); err != nil {
		t.Fatalf("Failed to create expired snapshot: %v", err)
	}

	// Create valid snapshot
	validSnapshot := &StateSnapshot{
		AgentID:      "agent-1",
		SnapshotType: "test",
		State:        map[string]interface{}{"data": "test"},
		Metadata:     SnapshotMetadata{Trigger: "test", Reason: "test"},
		ExpiresAt:    future,
	}
	if err := repo.CreateSnapshot(ctx, validSnapshot); err != nil {
		t.Fatalf("Failed to create valid snapshot: %v", err)
	}

	// Run cleanup
	count, err := repo.CleanupExpired(ctx)
	if err != nil {
		t.Fatalf("Failed to cleanup expired: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 items cleaned up, got %d", count)
	}

	// Verify expired items are gone
	_, err = repo.GetWorking(ctx, expiredWorking.AgentID, expiredWorking.Key)
	if err == nil {
		t.Error("Expected expired working memory to be deleted")
	}

	_, err = repo.GetSnapshot(ctx, expiredSnapshot.ID)
	if err == nil {
		t.Error("Expected expired snapshot to be deleted")
	}

	// Verify valid items remain
	_, err = repo.GetWorking(ctx, validWorking.AgentID, validWorking.Key)
	if err != nil {
		t.Error("Valid working memory should still exist")
	}

	_, err = repo.GetSnapshot(ctx, validSnapshot.ID)
	if err != nil {
		t.Error("Valid snapshot should still exist")
	}
}

// TestRepository_GetMemoryStats tests memory statistics retrieval
func TestRepository_GetMemoryStats(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()

	// Create working memories
	for i := 0; i < 3; i++ {
		mem := &WorkingMemory{
			AgentID:   "agent-1",
			Key:       "work-" + strconv.Itoa(i),
			Value:     "data",
			ExpiresAt: now.Add(1 * time.Hour),
			Metadata:  make(map[string]interface{}),
		}
		if err := repo.StoreWorking(ctx, mem); err != nil {
			t.Fatalf("Failed to store working memory: %v", err)
		}
	}

	// Create long-term memories
	for i := 0; i < 5; i++ {
		mem := &LongtermMemory{
			AgentID:  "agent-1",
			Category: "test",
			Key:      "long-" + strconv.Itoa(i),
			Value:    "data",
			Metadata: MemoryMetadata{
				Importance: 5,
				Tags:       []string{"test"},
			},
		}
		if err := repo.StoreLongterm(ctx, mem); err != nil {
			t.Fatalf("Failed to store longterm memory: %v", err)
		}
	}

	// Create snapshots
	for i := 0; i < 2; i++ {
		snap := &StateSnapshot{
			AgentID:      "agent-1",
			SnapshotType: "test",
			State:        map[string]interface{}{"data": "test"},
			Metadata:     SnapshotMetadata{Trigger: "test", Reason: "test"},
			ExpiresAt:    now.Add(1 * time.Hour),
		}
		if err := repo.CreateSnapshot(ctx, snap); err != nil {
			t.Fatalf("Failed to create snapshot: %v", err)
		}
	}

	// Get stats
	stats, err := repo.GetMemoryStats(ctx, "agent-1")
	if err != nil {
		t.Fatalf("Failed to get memory stats: %v", err)
	}

	if stats.AgentID != "agent-1" {
		t.Errorf("AgentID = %v, want agent-1", stats.AgentID)
	}
	if stats.WorkingMemoryCount != 3 {
		t.Errorf("WorkingMemoryCount = %v, want 3", stats.WorkingMemoryCount)
	}
	if stats.LongtermMemoryCount != 5 {
		t.Errorf("LongtermMemoryCount = %v, want 5", stats.LongtermMemoryCount)
	}
	if stats.SnapshotCount != 2 {
		t.Errorf("SnapshotCount = %v, want 2", stats.SnapshotCount)
	}
}
