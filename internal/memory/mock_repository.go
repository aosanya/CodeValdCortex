package memory

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MockRepository is a mock implementation of MemoryRepository for testing
type MockRepository struct {
	mu sync.RWMutex

	// Storage maps
	workingMemory  map[string]*WorkingMemory  // key: agentID:key
	longtermMemory map[string]*LongtermMemory // key: agentID:key
	snapshots      map[string]*StateSnapshot  // key: snapshotID
	syncStatus     map[string]*SyncStatus     // key: agentID:instanceID

	// Call tracking
	calls map[string]int

	// Error injection
	errors map[string]error
}

// NewMockRepository creates a new mock repository
func NewMockRepository() *MockRepository {
	return &MockRepository{
		workingMemory:  make(map[string]*WorkingMemory),
		longtermMemory: make(map[string]*LongtermMemory),
		snapshots:      make(map[string]*StateSnapshot),
		syncStatus:     make(map[string]*SyncStatus),
		calls:          make(map[string]int),
		errors:         make(map[string]error),
	}
}

// SetError sets an error to be returned for a specific method
func (m *MockRepository) SetError(method string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[method] = err
}

// GetCallCount returns the number of times a method was called
func (m *MockRepository) GetCallCount(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.calls[method]
}

// Reset clears all data and call tracking
func (m *MockRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workingMemory = make(map[string]*WorkingMemory)
	m.longtermMemory = make(map[string]*LongtermMemory)
	m.snapshots = make(map[string]*StateSnapshot)
	m.syncStatus = make(map[string]*SyncStatus)
	m.calls = make(map[string]int)
	m.errors = make(map[string]error)
}

func (m *MockRepository) trackCall(method string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls[method]++
	if err, ok := m.errors[method]; ok {
		return err
	}
	return nil
}

// ============================================================================
// Working Memory Operations
// ============================================================================

func (m *MockRepository) StoreWorking(ctx context.Context, mem *WorkingMemory) error {
	if err := m.trackCall("StoreWorking"); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", mem.AgentID, mem.Key)
	now := time.Now()
	mem.CreatedAt = now
	mem.UpdatedAt = now
	mem.Version = 1
	mem.AccessCount = 0
	m.workingMemory[key] = mem
	return nil
}

func (m *MockRepository) GetWorking(ctx context.Context, agentID, key string) (*WorkingMemory, error) {
	if err := m.trackCall("GetWorking"); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	k := fmt.Sprintf("%s:%s", agentID, key)
	mem, ok := m.workingMemory[k]
	if !ok {
		return nil, fmt.Errorf("working memory not found")
	}

	// Update access tracking
	go func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		mem.AccessCount++
		mem.AccessedAt = time.Now()
	}()

	return mem, nil
}

func (m *MockRepository) UpdateWorking(ctx context.Context, mem *WorkingMemory) error {
	if err := m.trackCall("UpdateWorking"); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", mem.AgentID, mem.Key)
	existing, ok := m.workingMemory[key]
	if !ok {
		return fmt.Errorf("working memory not found")
	}

	// Check version for optimistic locking
	if mem.Version != existing.Version {
		return fmt.Errorf("version mismatch")
	}

	mem.Version++
	mem.UpdatedAt = time.Now()
	m.workingMemory[key] = mem
	return nil
}

func (m *MockRepository) DeleteWorking(ctx context.Context, agentID, key string) error {
	if err := m.trackCall("DeleteWorking"); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	k := fmt.Sprintf("%s:%s", agentID, key)
	delete(m.workingMemory, k)
	return nil
}

func (m *MockRepository) ListWorking(ctx context.Context, agentID string, filters MemoryFilters) ([]*WorkingMemory, error) {
	if err := m.trackCall("ListWorking"); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*WorkingMemory
	for _, mem := range m.workingMemory {
		if mem.AgentID != agentID {
			continue
		}
		// Apply filters if needed
		result = append(result, mem)
	}
	return result, nil
}

func (m *MockRepository) ClearWorking(ctx context.Context, agentID string) error {
	if err := m.trackCall("ClearWorking"); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for key, mem := range m.workingMemory {
		if mem.AgentID == agentID {
			delete(m.workingMemory, key)
		}
	}
	return nil
}

// ============================================================================
// Long-term Memory Operations
// ============================================================================

func (m *MockRepository) StoreLongterm(ctx context.Context, mem *LongtermMemory) error {
	if err := m.trackCall("StoreLongterm"); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", mem.AgentID, mem.Key)
	now := time.Now()
	mem.CreatedAt = now
	mem.UpdatedAt = now
	mem.Version = 1
	mem.AccessCount = 0
	m.longtermMemory[key] = mem
	return nil
}

func (m *MockRepository) GetLongterm(ctx context.Context, agentID, key string) (*LongtermMemory, error) {
	if err := m.trackCall("GetLongterm"); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	k := fmt.Sprintf("%s:%s", agentID, key)
	mem, ok := m.longtermMemory[k]
	if !ok {
		return nil, fmt.Errorf("longterm memory not found")
	}

	// Update access tracking
	go func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		mem.AccessCount++
		mem.LastAccessed = time.Now()
	}()

	return mem, nil
}

func (m *MockRepository) UpdateLongterm(ctx context.Context, mem *LongtermMemory) error {
	if err := m.trackCall("UpdateLongterm"); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", mem.AgentID, mem.Key)
	existing, ok := m.longtermMemory[key]
	if !ok {
		return fmt.Errorf("longterm memory not found")
	}

	// Check version for optimistic locking
	if mem.Version != existing.Version {
		return fmt.Errorf("version mismatch")
	}

	mem.Version++
	mem.UpdatedAt = time.Now()
	m.longtermMemory[key] = mem
	return nil
}

func (m *MockRepository) DeleteLongterm(ctx context.Context, agentID, key string) error {
	if err := m.trackCall("DeleteLongterm"); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	k := fmt.Sprintf("%s:%s", agentID, key)
	delete(m.longtermMemory, k)
	return nil
}

func (m *MockRepository) ListLongterm(ctx context.Context, agentID string, filters MemoryFilters) ([]*LongtermMemory, error) {
	if err := m.trackCall("ListLongterm"); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*LongtermMemory
	for _, mem := range m.longtermMemory {
		if mem.AgentID != agentID {
			continue
		}

		// Apply filters
		if filters.Category != "" && mem.Category != filters.Category {
			continue
		}
		if filters.BeforeTime != nil && mem.CreatedAt.After(*filters.BeforeTime) {
			continue
		}
		if filters.MinImportance > 0 && mem.Metadata.Importance < filters.MinImportance {
			continue
		}

		result = append(result, mem)
	}
	return result, nil
}

func (m *MockRepository) SearchLongterm(ctx context.Context, agentID string, query MemoryQuery) ([]*LongtermMemory, error) {
	if err := m.trackCall("SearchLongterm"); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*LongtermMemory
	for _, mem := range m.longtermMemory {
		if mem.AgentID != agentID {
			continue
		}

		// Apply query filters
		if query.Filters.Category != "" && mem.Category != query.Filters.Category {
			continue
		}

		result = append(result, mem)
	}
	return result, nil
}

// ============================================================================
// Snapshot Operations
// ============================================================================

func (m *MockRepository) CreateSnapshot(ctx context.Context, snapshot *StateSnapshot) error {
	if err := m.trackCall("CreateSnapshot"); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate ID if not set
	if snapshot.ID == "" {
		snapshot.ID = fmt.Sprintf("snap-%d", time.Now().UnixNano())
	}

	now := time.Now()
	snapshot.CreatedAt = now
	snapshot.Version = 1
	m.snapshots[snapshot.ID] = snapshot
	return nil
}

func (m *MockRepository) GetSnapshot(ctx context.Context, snapshotID string) (*StateSnapshot, error) {
	if err := m.trackCall("GetSnapshot"); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshot, ok := m.snapshots[snapshotID]
	if !ok {
		return nil, fmt.Errorf("snapshot not found")
	}
	return snapshot, nil
}

func (m *MockRepository) ListSnapshots(ctx context.Context, agentID string, filters SnapshotFilters) ([]*StateSnapshot, error) {
	if err := m.trackCall("ListSnapshots"); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*StateSnapshot
	for _, snapshot := range m.snapshots {
		if snapshot.AgentID != agentID {
			continue
		}

		// Apply filters
		if filters.SnapshotType != "" && snapshot.SnapshotType != filters.SnapshotType {
			continue
		}

		result = append(result, snapshot)
	}
	return result, nil
}

func (m *MockRepository) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	if err := m.trackCall("DeleteSnapshot"); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.snapshots, snapshotID)
	return nil
}

// ============================================================================
// Sync Operations
// ============================================================================

func (m *MockRepository) GetSyncStatus(ctx context.Context, agentID, instanceID string) (*SyncStatus, error) {
	if err := m.trackCall("GetSyncStatus"); err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", agentID, instanceID)
	status, ok := m.syncStatus[key]
	if !ok {
		// Create new status
		status = &SyncStatus{
			AgentID:    agentID,
			InstanceID: instanceID,
			Status:     SyncStateSynced,
			LastSyncAt: time.Now(),
		}
		m.syncStatus[key] = status
	}
	return status, nil
}

func (m *MockRepository) UpdateSyncStatus(ctx context.Context, status *SyncStatus) error {
	if err := m.trackCall("UpdateSyncStatus"); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", status.AgentID, status.InstanceID)
	m.syncStatus[key] = status
	return nil
}

// ============================================================================
// Maintenance Operations
// ============================================================================

func (m *MockRepository) CleanupExpired(ctx context.Context) (int, error) {
	if err := m.trackCall("CleanupExpired"); err != nil {
		return 0, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	now := time.Now()

	// Clean working memory
	for key, mem := range m.workingMemory {
		if now.After(mem.ExpiresAt) {
			delete(m.workingMemory, key)
			count++
		}
	}

	// Clean snapshots
	for key, snapshot := range m.snapshots {
		if now.After(snapshot.ExpiresAt) {
			delete(m.snapshots, key)
			count++
		}
	}

	return count, nil
}

func (m *MockRepository) GetMemoryStats(ctx context.Context, agentID string) (*MemoryStats, error) {
	if err := m.trackCall("GetMemoryStats"); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &MemoryStats{
		AgentID:             agentID,
		WorkingMemoryCount:  0,
		LongtermMemoryCount: 0,
		SnapshotCount:       0,
	}

	// Count working memory
	for _, mem := range m.workingMemory {
		if mem.AgentID == agentID {
			stats.WorkingMemoryCount++
		}
	}

	// Count longterm memory
	for _, mem := range m.longtermMemory {
		if mem.AgentID == agentID {
			stats.LongtermMemoryCount++
		}
	}

	// Count snapshots
	for _, snapshot := range m.snapshots {
		if snapshot.AgentID == agentID {
			stats.SnapshotCount++
		}
	}

	return stats, nil
}
