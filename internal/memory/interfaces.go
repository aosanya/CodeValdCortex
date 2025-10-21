package memory

import (
	"context"
	"time"
)

// MemoryRepository defines the interface for memory persistence operations
type MemoryRepository interface {
	// Working Memory Operations
	StoreWorking(ctx context.Context, memory *WorkingMemory) error
	GetWorking(ctx context.Context, agentID, key string) (*WorkingMemory, error)
	UpdateWorking(ctx context.Context, memory *WorkingMemory) error
	DeleteWorking(ctx context.Context, agentID, key string) error
	ListWorking(ctx context.Context, agentID string, filters MemoryFilters) ([]*WorkingMemory, error)
	ClearWorking(ctx context.Context, agentID string) error

	// Long-term Memory Operations
	StoreLongterm(ctx context.Context, memory *LongtermMemory) error
	GetLongterm(ctx context.Context, agentID, key string) (*LongtermMemory, error)
	UpdateLongterm(ctx context.Context, memory *LongtermMemory) error
	DeleteLongterm(ctx context.Context, agentID, key string) error
	ListLongterm(ctx context.Context, agentID string, filters MemoryFilters) ([]*LongtermMemory, error)
	SearchLongterm(ctx context.Context, agentID string, query MemoryQuery) ([]*LongtermMemory, error)

	// State Snapshot Operations
	CreateSnapshot(ctx context.Context, snapshot *StateSnapshot) error
	GetSnapshot(ctx context.Context, snapshotID string) (*StateSnapshot, error)
	ListSnapshots(ctx context.Context, agentID string, filters SnapshotFilters) ([]*StateSnapshot, error)
	DeleteSnapshot(ctx context.Context, snapshotID string) error

	// Sync Status Operations
	GetSyncStatus(ctx context.Context, agentID, instanceID string) (*SyncStatus, error)
	UpdateSyncStatus(ctx context.Context, status *SyncStatus) error

	// Maintenance Operations
	CleanupExpired(ctx context.Context) (int, error)
	GetMemoryStats(ctx context.Context, agentID string) (*MemoryStats, error)
}

// MemoryService defines the high-level memory management interface
type MemoryService interface {
	// Working Memory
	StoreWorking(ctx context.Context, agentID, key string, value interface{}, ttl time.Duration) error
	RetrieveWorking(ctx context.Context, agentID, key string) (interface{}, error)
	UpdateWorking(ctx context.Context, agentID, key string, value interface{}) error
	DeleteWorking(ctx context.Context, agentID, key string) error
	ClearWorking(ctx context.Context, agentID string) error
	ListWorking(ctx context.Context, agentID string, filters MemoryFilters) ([]*WorkingMemory, error)

	// Long-term Memory
	Remember(ctx context.Context, agentID, key string, value interface{}, category string, metadata map[string]interface{}) error
	Recall(ctx context.Context, agentID, key string) (interface{}, error)
	Search(ctx context.Context, agentID string, query MemoryQuery) ([]*LongtermMemory, error)
	Forget(ctx context.Context, agentID, key string) error
	Archive(ctx context.Context, agentID string, criteria ArchiveCriteria) error

	// State Snapshots
	CreateSnapshot(ctx context.Context, agentID string, snapshotType, reason string) (*StateSnapshot, error)
	RestoreSnapshot(ctx context.Context, agentID, snapshotID string) error
	ListSnapshots(ctx context.Context, agentID string, filters SnapshotFilters) ([]*StateSnapshot, error)
	DeleteSnapshot(ctx context.Context, snapshotID string) error

	// Synchronization
	SyncMemory(ctx context.Context, agentID string) (*SyncResult, error)
	GetSyncStatus(ctx context.Context, agentID string) (*SyncStatus, error)
	ResolveConflict(ctx context.Context, conflict *MemoryConflict, strategy ConflictStrategy) error

	// Maintenance
	CleanupExpired(ctx context.Context) (int, error)
	GetMemoryStats(ctx context.Context, agentID string) (*MemoryStats, error)
}

// MemorySynchronizer defines the interface for memory synchronization operations
type MemorySynchronizer interface {
	// StartSync begins periodic memory synchronization for an agent
	StartSync(ctx context.Context, agentID, instanceID string, interval time.Duration) error

	// StopSync stops the synchronization process for an agent
	StopSync(ctx context.Context, agentID string) error

	// SyncNow performs an immediate synchronization
	SyncNow(ctx context.Context, agentID string) (*SyncResult, error)

	// ResolveConflict resolves a detected conflict using the specified strategy
	ResolveConflict(ctx context.Context, conflict *MemoryConflict, strategy ConflictStrategy) error

	// GetConflicts returns all unresolved conflicts for an agent
	GetConflicts(ctx context.Context, agentID string) ([]MemoryConflict, error)

	// GetSyncStatus returns the current synchronization status
	GetSyncStatus(ctx context.Context, agentID string) (*SyncStatus, error)
}
