# Agent Memory System - Design Documentation

## 1. Overview

### 1.1 Purpose
The Agent Memory System provides persistent state management and memory synchronization for agents in the CodeValdCortex platform. It enables agents to maintain context across restarts, share knowledge, and build upon past experiences.

### 1.2 Key Features
- **Working Memory**: Short-term, fast-access memory for current tasks
- **Long-term Memory**: Persistent storage for knowledge and experiences
- **State Snapshots**: Point-in-time captures of agent state for recovery
- **Memory Synchronization**: Distributed state sync across agent instances
- **Semantic Search**: Context-aware memory retrieval
- **Memory Lifecycle**: Automatic cleanup and archival based on TTL and relevance

### 1.3 Dependencies
- **MVP-003**: Agent Registry System (agent identification)
- **MVP-004**: Agent Lifecycle Management (state persistence)
- **MVP-005**: Agent Communication System (sync notifications)
- **ArangoDB**: Database for memory persistence

## 2. Database Schema Design

### 2.1 Collections

#### `agent_working_memory` Collection (Document)
Stores short-term, task-specific memory with fast TTL.

**Document Structure**:
```json
{
  "_key": "mem-uuid-12345",
  "_rev": "...",
  "agent_id": "agent-abc123",
  "memory_type": "working",
  "key": "current_task_context",
  "value": {
    "task_id": "task-xyz789",
    "step": 3,
    "state": "processing",
    "partial_results": {...}
  },
  "metadata": {
    "source": "task_processor",
    "priority": 8,
    "tags": ["task", "active", "critical"]
  },
  "created_at": "2025-10-21T10:00:00.000Z",
  "updated_at": "2025-10-21T10:05:00.000Z",
  "accessed_at": "2025-10-21T10:05:30.000Z",
  "access_count": 15,
  "expires_at": "2025-10-21T11:00:00.000Z",
  "version": 3
}
```

**Indexes**:
```javascript
// Fast retrieval by agent and key
db._ensureIndex({
  type: "persistent",
  fields: ["agent_id", "key"],
  name: "idx_working_memory_lookup",
  unique: true
});

// Cleanup expired memories
db._ensureIndex({
  type: "persistent",
  fields: ["expires_at"],
  name: "idx_working_memory_expiration"
});

// Tag-based search
db._ensureIndex({
  type: "persistent",
  fields: ["agent_id", "metadata.tags[*]"],
  name: "idx_working_memory_tags"
});
```

#### `agent_longterm_memory` Collection (Document)
Stores persistent knowledge, experiences, and learned patterns.

**Document Structure**:
```json
{
  "_key": "mem-long-67890",
  "_rev": "...",
  "agent_id": "agent-abc123",
  "memory_type": "longterm",
  "category": "knowledge",
  "key": "skill_learned_data_validation",
  "value": {
    "skill_name": "data_validation",
    "proficiency": 0.85,
    "learned_at": "2025-10-15T14:30:00.000Z",
    "examples": [...],
    "patterns": [...]
  },
  "embedding": [0.123, 0.456, ...],  // Vector for semantic search (future)
  "metadata": {
    "source": "experience_aggregator",
    "importance": 9,
    "confidence": 0.92,
    "tags": ["skill", "data", "validation"],
    "references": ["mem-long-12345", "mem-long-67891"]
  },
  "created_at": "2025-10-15T14:30:00.000Z",
  "updated_at": "2025-10-21T10:00:00.000Z",
  "last_accessed": "2025-10-21T09:45:00.000Z",
  "access_count": 42,
  "version": 8
}
```

**Indexes**:
```javascript
// Agent-specific memory retrieval
db._ensureIndex({
  type: "persistent",
  fields: ["agent_id", "category"],
  name: "idx_longterm_memory_category"
});

// Key-based lookup
db._ensureIndex({
  type: "persistent",
  fields: ["agent_id", "key"],
  name: "idx_longterm_memory_key",
  unique: true
});

// Tag-based search
db._ensureIndex({
  type: "persistent",
  fields: ["agent_id", "metadata.tags[*]"],
  name: "idx_longterm_memory_tags"
});

// Importance-based retrieval
db._ensureIndex({
  type: "persistent",
  fields: ["agent_id", "metadata.importance"],
  name: "idx_longterm_memory_importance"
});

// Access patterns for cleanup
db._ensureIndex({
  type: "persistent",
  fields: ["last_accessed", "access_count"],
  name: "idx_longterm_memory_access"
});
```

#### `agent_state_snapshots` Collection (Document)
Point-in-time captures of complete agent state for recovery.

**Document Structure**:
```json
{
  "_key": "snapshot-uuid-99999",
  "_rev": "...",
  "agent_id": "agent-abc123",
  "snapshot_type": "periodic",
  "state": {
    "agent_state": "running",
    "current_task": "task-xyz789",
    "working_memory": {...},
    "active_subscriptions": [...],
    "pending_messages": [...]
  },
  "checksum": "sha256-...",
  "metadata": {
    "trigger": "scheduled",
    "reason": "periodic_backup",
    "size_bytes": 125000,
    "compressed": true
  },
  "created_at": "2025-10-21T10:00:00.000Z",
  "expires_at": "2025-11-21T10:00:00.000Z",
  "version": 1
}
```

**Indexes**:
```javascript
// Latest snapshots per agent
db._ensureIndex({
  type: "persistent",
  fields: ["agent_id", "created_at"],
  name: "idx_snapshots_agent_time"
});

// Cleanup old snapshots
db._ensureIndex({
  type: "persistent",
  fields: ["expires_at"],
  name: "idx_snapshots_expiration"
});

// Snapshot type filtering
db._ensureIndex({
  type: "persistent",
  fields: ["agent_id", "snapshot_type"],
  name: "idx_snapshots_type"
});
```

#### `agent_memory_sync` Collection (Document)
Tracks memory synchronization status across distributed agent instances.

**Document Structure**:
```json
{
  "_key": "sync-uuid-11111",
  "_rev": "...",
  "agent_id": "agent-abc123",
  "instance_id": "instance-node1-abc123",
  "last_sync_at": "2025-10-21T10:05:00.000Z",
  "sync_version": 156,
  "pending_changes": 0,
  "conflicts": [],
  "status": "synced",
  "metadata": {
    "node": "k8s-node-1",
    "sync_duration_ms": 45
  }
}
```

**Indexes**:
```javascript
// Agent sync status lookup
db._ensureIndex({
  type: "persistent",
  fields: ["agent_id", "instance_id"],
  name: "idx_sync_status",
  unique: true
});

// Conflict detection
db._ensureIndex({
  type: "persistent",
  fields: ["agent_id", "status"],
  name: "idx_sync_conflicts"
});
```

## 3. Memory Types and Operations

### 3.1 Working Memory
**Purpose**: Fast, ephemeral storage for active task context

**Characteristics**:
- TTL: 1-24 hours (configurable)
- Size: Limited per agent (e.g., 10 MB)
- Access: High frequency
- Persistence: Optional (can be memory-only with DB backup)

**Operations**:
- `Store(key, value, ttl)`: Save working memory entry
- `Retrieve(key)`: Get current value
- `Update(key, value)`: Modify existing entry
- `Delete(key)`: Remove entry
- `Clear()`: Remove all working memory

### 3.2 Long-term Memory
**Purpose**: Persistent knowledge and experience storage

**Characteristics**:
- TTL: Indefinite or years
- Size: Unlimited (within reason)
- Access: Lower frequency
- Persistence: Always persisted

**Operations**:
- `Remember(key, value, category)`: Store long-term memory
- `Recall(key)`: Retrieve specific memory
- `Search(query, filters)`: Semantic/tag-based search
- `Forget(key)`: Remove memory
- `Archive(criteria)`: Move to cold storage

### 3.3 State Snapshots
**Purpose**: Point-in-time recovery checkpoints

**Characteristics**:
- TTL: 7-90 days (configurable)
- Size: Full agent state (compressed)
- Access: Infrequent (recovery only)
- Persistence: Always persisted

**Operations**:
- `CreateSnapshot(reason)`: Capture current state
- `RestoreSnapshot(snapshotID)`: Restore to checkpoint
- `ListSnapshots(filters)`: Browse available snapshots
- `DeleteSnapshot(snapshotID)`: Remove old snapshot

## 4. Memory Synchronization

### 4.1 Synchronization Strategy

**Periodic Sync**:
- Every 30 seconds (configurable)
- Push local changes to database
- Pull remote changes from database
- Merge with conflict resolution

**Event-driven Sync**:
- On critical state changes
- Before agent shutdown
- After task completion
- Manual trigger via API

### 4.2 Conflict Resolution

**Resolution Strategies**:

1. **Last-Write-Wins (Default)**:
   - Compare timestamps
   - Keep most recent version
   - Log conflict for audit

2. **Version-based**:
   - Track version numbers
   - Reject stale updates
   - Force manual resolution

3. **Custom Merge**:
   - Application-specific logic
   - Intelligent merging of values
   - Preserve both versions with metadata

**Conflict Detection**:
```go
type MemoryConflict struct {
    Key           string
    LocalVersion  int
    RemoteVersion int
    LocalValue    interface{}
    RemoteValue   interface{}
    LocalTime     time.Time
    RemoteTime    time.Time
}
```

### 4.3 Distributed Synchronization

**Multi-instance Coordination**:
- Each agent instance maintains local cache
- Periodic sync to shared database
- Pub/sub notifications for real-time updates
- Optimistic locking for concurrent writes

**Sync Flow**:
```
1. Agent modifies working memory (local cache)
2. Change marked as "dirty" for sync
3. Background sync process:
   a. Fetch current DB version
   b. Compare with local version
   c. Resolve conflicts if any
   d. Persist merged version
   e. Update sync metadata
4. Publish sync event (via MVP-005)
5. Other instances receive notification
6. Pull updated memory from DB
```

## 5. API Design

### 5.1 Memory Service Interface

```go
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
```

### 5.2 Agent Memory Methods

```go
// Add to Agent struct in internal/agent/agent.go
type Agent struct {
    // ... existing fields ...
    memoryService *memory.MemoryService
}

// Working Memory methods
func (a *Agent) Remember(key string, value interface{}, ttl time.Duration) error
func (a *Agent) Recall(key string) (interface{}, error)
func (a *Agent) Forget(key string) error

// Long-term Memory methods
func (a *Agent) LearnKnowledge(key string, value interface{}, category string) error
func (a *Agent) RecallKnowledge(key string) (interface{}, error)
func (a *Agent) SearchMemory(query string, tags []string) ([]*memory.LongtermMemory, error)

// Snapshot methods
func (a *Agent) CreateMemorySnapshot(reason string) error
func (a *Agent) RestoreFromSnapshot(snapshotID string) error

// Synchronization
func (a *Agent) SyncMemory() error
func (a *Agent) GetMemoryStats() (*memory.MemoryStats, error)
```

## 6. Implementation Components

### 6.1 File Structure
```
internal/memory/
├── repository.go          # ArangoDB persistence layer
├── repository_test.go     # Repository integration tests
├── service.go            # High-level memory operations
├── service_test.go       # Service unit tests
├── synchronizer.go       # Memory sync coordination
├── synchronizer_test.go  # Sync unit tests
├── types.go             # Data structures
├── interfaces.go        # Interface definitions
└── TESTING.md          # Testing documentation
```

### 6.2 Core Types

```go
// Working Memory
type WorkingMemory struct {
    ID          string
    AgentID     string
    Key         string
    Value       interface{}
    Metadata    map[string]interface{}
    CreatedAt   time.Time
    UpdatedAt   time.Time
    AccessedAt  time.Time
    AccessCount int
    ExpiresAt   time.Time
    Version     int
}

// Long-term Memory
type LongtermMemory struct {
    ID          string
    AgentID     string
    Category    string
    Key         string
    Value       interface{}
    Embedding   []float64  // For semantic search (future)
    Metadata    MemoryMetadata
    CreatedAt   time.Time
    UpdatedAt   time.Time
    LastAccessed time.Time
    AccessCount int
    Version     int
}

type MemoryMetadata struct {
    Source      string
    Importance  int      // 1-10
    Confidence  float64  // 0.0-1.0
    Tags        []string
    References  []string // Related memory IDs
}

// State Snapshot
type StateSnapshot struct {
    ID           string
    AgentID      string
    SnapshotType string  // periodic, manual, pre-update, pre-shutdown
    State        map[string]interface{}
    Checksum     string
    Metadata     SnapshotMetadata
    CreatedAt    time.Time
    ExpiresAt    time.Time
    Version      int
}

// Sync Status
type SyncStatus struct {
    AgentID         string
    InstanceID      string
    LastSyncAt      time.Time
    SyncVersion     int
    PendingChanges  int
    Conflicts       []MemoryConflict
    Status          SyncState  // synced, syncing, conflict, error
}

type SyncState string
const (
    SyncStateSynced    SyncState = "synced"
    SyncStateSyncing   SyncState = "syncing"
    SyncStateConflict  SyncState = "conflict"
    SyncStateError     SyncState = "error"
)
```

## 7. Performance Considerations

### 7.1 Caching Strategy
- **Local cache**: In-memory working memory cache per agent instance
- **TTL-based eviction**: Automatic cleanup of expired entries
- **LRU eviction**: Remove least-recently-used entries when cache full
- **Write-through**: Updates go to both cache and DB
- **Read-through**: Cache miss triggers DB lookup

### 7.2 Optimization Techniques
- **Batch operations**: Group multiple memory updates
- **Lazy persistence**: Defer non-critical writes
- **Compression**: Compress large snapshots
- **Indexed queries**: Leverage ArangoDB indexes
- **Connection pooling**: Reuse database connections

### 7.3 Scalability
- **Horizontal scaling**: Multiple agent instances share DB
- **Sharding**: Partition memory by agent_id (future)
- **Read replicas**: Distribute read load (future)
- **Async sync**: Non-blocking synchronization

## 8. Security and Privacy

### 8.1 Access Control
- Memory accessible only by owning agent
- Admin override for debugging/support
- Audit log for sensitive memory access

### 8.2 Data Protection
- Encryption at rest (ArangoDB feature)
- Encryption in transit (TLS)
- Sensitive data masking in logs
- Configurable retention policies

### 8.3 Data Lifecycle
- **Retention**: Configurable per memory type
- **Archival**: Move old memories to cold storage
- **Deletion**: Secure deletion with audit trail
- **GDPR compliance**: Right to be forgotten support

## 9. Monitoring and Observability

### 9.1 Metrics
- Memory size per agent (working + long-term)
- Memory operations per second
- Sync latency and frequency
- Conflict rate
- Cache hit/miss ratio
- Expired memory cleanup count

### 9.2 Logging
- Memory operations (INFO level)
- Sync conflicts (WARN level)
- Persistence failures (ERROR level)
- Performance anomalies (WARN level)

### 9.3 Alerting
- High conflict rate (>5% of syncs)
- Sync failures (>3 consecutive)
- Memory size exceeds quota
- Snapshot creation failures

## 10. Testing Strategy

### 10.1 Unit Tests
- Repository CRUD operations
- Service layer logic
- Conflict resolution algorithms
- Cache behavior
- TTL expiration

### 10.2 Integration Tests
- End-to-end memory workflows
- Multi-instance synchronization
- Snapshot creation and restoration
- Database failover scenarios

### 10.3 Performance Tests
- High-frequency memory operations
- Large memory payloads
- Concurrent access patterns
- Sync performance under load

## 11. Future Enhancements

### 11.1 Semantic Search
- Vector embeddings for memory content
- Similarity-based memory retrieval
- Context-aware recommendations
- Knowledge graph integration

### 11.2 Memory Optimization
- Automatic importance scoring
- Intelligent archival
- Memory consolidation
- Deduplication

### 11.3 Advanced Sync
- Distributed consensus (Raft)
- Multi-master replication
- Conflict-free replicated data types (CRDTs)
- Event sourcing for memory changes

### 11.4 Analytics
- Memory usage patterns
- Knowledge growth tracking
- Agent learning curves
- Memory effectiveness metrics

---

**Document Version**: 1.0
**Last Updated**: 2025-10-21
**Status**: Design Complete - Ready for Implementation
