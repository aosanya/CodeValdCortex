package memory

import "time"

// MemoryType defines the type of memory storage
type MemoryType string

const (
	// MemoryTypeWorking represents short-term, task-specific memory
	MemoryTypeWorking MemoryType = "working"
	// MemoryTypeLongterm represents persistent knowledge storage
	MemoryTypeLongterm MemoryType = "longterm"
)

// WorkingMemory represents short-term memory for active task context
type WorkingMemory struct {
	// ID is the unique identifier for the memory entry
	ID string `json:"id"`

	// AgentID is the ID of the agent that owns this memory
	AgentID string `json:"agent_id"`

	// Key is the unique key within the agent's working memory space
	Key string `json:"key"`

	// Value is the memory content (stored as interface{} for flexibility)
	Value interface{} `json:"value"`

	// Metadata contains additional information about the memory
	Metadata map[string]interface{} `json:"metadata"`

	// CreatedAt is when the memory was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the memory was last modified
	UpdatedAt time.Time `json:"updated_at"`

	// AccessedAt is when the memory was last accessed
	AccessedAt time.Time `json:"accessed_at"`

	// AccessCount tracks how many times this memory has been accessed
	AccessCount int `json:"access_count"`

	// ExpiresAt defines when this memory should be automatically cleaned up
	ExpiresAt time.Time `json:"expires_at"`

	// Version is used for optimistic locking and conflict detection
	Version int `json:"version"`
}

// LongtermMemory represents persistent knowledge and experiences
type LongtermMemory struct {
	// ID is the unique identifier for the memory entry
	ID string `json:"id"`

	// AgentID is the ID of the agent that owns this memory
	AgentID string `json:"agent_id"`

	// Category classifies the type of knowledge (e.g., "skill", "fact", "experience")
	Category string `json:"category"`

	// Key is the unique key within the agent's long-term memory space
	Key string `json:"key"`

	// Value is the memory content
	Value interface{} `json:"value"`

	// Embedding is a vector representation for semantic search (future enhancement)
	Embedding []float64 `json:"embedding,omitempty"`

	// Metadata contains structured information about the memory
	Metadata MemoryMetadata `json:"metadata"`

	// CreatedAt is when the memory was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the memory was last modified
	UpdatedAt time.Time `json:"updated_at"`

	// LastAccessed is when the memory was last retrieved
	LastAccessed time.Time `json:"last_accessed"`

	// AccessCount tracks memory usage frequency
	AccessCount int `json:"access_count"`

	// Version for conflict detection
	Version int `json:"version"`
}

// MemoryMetadata contains structured metadata for long-term memories
type MemoryMetadata struct {
	// Source identifies where this memory came from
	Source string `json:"source"`

	// Importance ranks the memory's value (1-10, higher = more important)
	Importance int `json:"importance"`

	// Confidence indicates certainty of the information (0.0-1.0)
	Confidence float64 `json:"confidence"`

	// Tags enable categorization and search
	Tags []string `json:"tags"`

	// References link to related memory IDs
	References []string `json:"references,omitempty"`
}

// StateSnapshot represents a point-in-time capture of agent state
type StateSnapshot struct {
	// ID is the unique identifier for the snapshot
	ID string `json:"id"`

	// AgentID is the ID of the agent this snapshot belongs to
	AgentID string `json:"agent_id"`

	// SnapshotType classifies the snapshot (periodic, manual, pre-update, pre-shutdown)
	SnapshotType string `json:"snapshot_type"`

	// State contains the complete agent state
	State map[string]interface{} `json:"state"`

	// Checksum for integrity verification
	Checksum string `json:"checksum"`

	// Metadata contains additional snapshot information
	Metadata SnapshotMetadata `json:"metadata"`

	// CreatedAt is when the snapshot was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt defines when this snapshot should be deleted
	ExpiresAt time.Time `json:"expires_at"`

	// Version for tracking
	Version int `json:"version"`
}

// SnapshotMetadata contains snapshot-specific metadata
type SnapshotMetadata struct {
	// Trigger identifies what caused the snapshot
	Trigger string `json:"trigger"`

	// Reason provides human-readable context
	Reason string `json:"reason"`

	// SizeBytes is the size of the snapshot data
	SizeBytes int64 `json:"size_bytes"`

	// Compressed indicates if the state data is compressed
	Compressed bool `json:"compressed"`
}

// SyncStatus tracks memory synchronization state
type SyncStatus struct {
	// AgentID is the agent being synchronized
	AgentID string `json:"agent_id"`

	// InstanceID identifies the specific agent instance
	InstanceID string `json:"instance_id"`

	// LastSyncAt is the timestamp of the last successful sync
	LastSyncAt time.Time `json:"last_sync_at"`

	// SyncVersion tracks the sync iteration number
	SyncVersion int `json:"sync_version"`

	// PendingChanges counts local changes not yet synced
	PendingChanges int `json:"pending_changes"`

	// Conflicts lists any detected conflicts
	Conflicts []MemoryConflict `json:"conflicts"`

	// Status indicates the current sync state
	Status SyncState `json:"status"`

	// Metadata contains sync-specific information
	Metadata map[string]interface{} `json:"metadata"`
}

// SyncState represents the synchronization status
type SyncState string

const (
	// SyncStateSynced indicates all changes are synchronized
	SyncStateSynced SyncState = "synced"
	// SyncStateSyncing indicates synchronization in progress
	SyncStateSyncing SyncState = "syncing"
	// SyncStateConflict indicates unresolved conflicts exist
	SyncStateConflict SyncState = "conflict"
	// SyncStateError indicates synchronization failure
	SyncStateError SyncState = "error"
)

// MemoryConflict represents a detected conflict during synchronization
type MemoryConflict struct {
	// Key is the memory key that has conflicting values
	Key string `json:"key"`

	// MemoryType indicates if it's working or long-term memory
	MemoryType MemoryType `json:"memory_type"`

	// LocalVersion is the version number of the local copy
	LocalVersion int `json:"local_version"`

	// RemoteVersion is the version number in the database
	RemoteVersion int `json:"remote_version"`

	// LocalValue is the local memory content
	LocalValue interface{} `json:"local_value"`

	// RemoteValue is the remote memory content
	RemoteValue interface{} `json:"remote_value"`

	// LocalTime is when the local version was modified
	LocalTime time.Time `json:"local_time"`

	// RemoteTime is when the remote version was modified
	RemoteTime time.Time `json:"remote_time"`

	// DetectedAt is when the conflict was discovered
	DetectedAt time.Time `json:"detected_at"`
}

// ConflictStrategy defines how to resolve memory conflicts
type ConflictStrategy string

const (
	// ConflictStrategyLastWriteWins uses timestamp to resolve conflicts
	ConflictStrategyLastWriteWins ConflictStrategy = "last_write_wins"
	// ConflictStrategyVersionBased rejects stale updates
	ConflictStrategyVersionBased ConflictStrategy = "version_based"
	// ConflictStrategyManual requires explicit resolution
	ConflictStrategyManual ConflictStrategy = "manual"
	// ConflictStrategyLocalWins always prefers local version
	ConflictStrategyLocalWins ConflictStrategy = "local_wins"
	// ConflictStrategyRemoteWins always prefers remote version
	ConflictStrategyRemoteWins ConflictStrategy = "remote_wins"
)

// SyncResult contains the outcome of a synchronization operation
type SyncResult struct {
	// AgentID is the agent that was synchronized
	AgentID string `json:"agent_id"`

	// SyncedAt is when the sync completed
	SyncedAt time.Time `json:"synced_at"`

	// DurationMs is how long the sync took
	DurationMs int64 `json:"duration_ms"`

	// ItemsSynced counts successfully synchronized items
	ItemsSynced int `json:"items_synced"`

	// Conflicts lists any conflicts encountered
	Conflicts []MemoryConflict `json:"conflicts"`

	// Errors lists any errors that occurred
	Errors []string `json:"errors"`

	// Success indicates if sync completed without errors
	Success bool `json:"success"`
}

// MemoryStats provides insights into agent memory usage
type MemoryStats struct {
	// AgentID is the agent these stats belong to
	AgentID string `json:"agent_id"`

	// WorkingMemoryCount is the number of working memory entries
	WorkingMemoryCount int `json:"working_memory_count"`

	// WorkingMemorySizeBytes is the approximate size of working memory
	WorkingMemorySizeBytes int64 `json:"working_memory_size_bytes"`

	// LongtermMemoryCount is the number of long-term memory entries
	LongtermMemoryCount int `json:"longterm_memory_count"`

	// LongtermMemorySizeBytes is the approximate size of long-term memory
	LongtermMemorySizeBytes int64 `json:"longterm_memory_size_bytes"`

	// SnapshotCount is the number of state snapshots
	SnapshotCount int `json:"snapshot_count"`

	// TotalSizeBytes is the total memory footprint
	TotalSizeBytes int64 `json:"total_size_bytes"`

	// LastSyncAt is when memory was last synchronized
	LastSyncAt time.Time `json:"last_sync_at"`

	// LastSnapshotAt is when the last snapshot was created
	LastSnapshotAt time.Time `json:"last_snapshot_at"`
}

// MemoryFilters defines filtering options for memory queries
type MemoryFilters struct {
	// Tags filters by memory tags
	Tags []string `json:"tags,omitempty"`

	// Category filters long-term memory by category
	Category string `json:"category,omitempty"`

	// MinImportance filters by minimum importance score
	MinImportance int `json:"min_importance,omitempty"`

	// AfterTime filters by creation/update time
	AfterTime *time.Time `json:"after_time,omitempty"`

	// BeforeTime filters by creation/update time
	BeforeTime *time.Time `json:"before_time,omitempty"`

	// Limit restricts the number of results
	Limit int `json:"limit,omitempty"`

	// Offset for pagination
	Offset int `json:"offset,omitempty"`

	// SortBy specifies the sort field
	SortBy string `json:"sort_by,omitempty"`

	// SortDesc indicates descending sort order
	SortDesc bool `json:"sort_desc,omitempty"`
}

// MemoryQuery defines search parameters for memory retrieval
type MemoryQuery struct {
	// Query is the search text (future: semantic search)
	Query string `json:"query"`

	// Filters apply additional constraints
	Filters MemoryFilters `json:"filters"`
}

// SnapshotFilters defines filtering options for snapshot queries
type SnapshotFilters struct {
	// SnapshotType filters by snapshot type
	SnapshotType string `json:"snapshot_type,omitempty"`

	// AfterTime filters by creation time
	AfterTime *time.Time `json:"after_time,omitempty"`

	// BeforeTime filters by creation time
	BeforeTime *time.Time `json:"before_time,omitempty"`

	// Limit restricts the number of results
	Limit int `json:"limit,omitempty"`

	// Offset for pagination
	Offset int `json:"offset,omitempty"`
}

// ArchiveCriteria defines rules for memory archival
type ArchiveCriteria struct {
	// OlderThan archives memories older than this duration
	OlderThan time.Duration `json:"older_than,omitempty"`

	// MaxAccessCount archives memories accessed less than this
	MaxAccessCount int `json:"max_access_count,omitempty"`

	// MaxImportance archives memories with importance below this
	MaxImportance int `json:"max_importance,omitempty"`

	// Categories limits archival to specific categories
	Categories []string `json:"categories,omitempty"`

	// DryRun simulates archival without actual deletion
	DryRun bool `json:"dry_run,omitempty"`
}
