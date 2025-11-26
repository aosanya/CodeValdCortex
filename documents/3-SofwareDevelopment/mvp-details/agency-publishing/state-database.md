# Agency State Database Schema & Migration (MVP-PUB-001)

**Domain**: Agency Publishing & Tagging  
**Priority**: P0 (Critical)  
**Status**: ✅ Complete (2025-11-20)

## Overview

This document covers the database schema for agency lifecycle management:
- **ArangoDB Collections**: `agency_publications` and `agency_tags`
- **Indexes**: Optimized for common query patterns
- **Database Migration**: Collection creation script
- **Data Migration**: Status → State field conversion

Related documentation:
- [State Models](./state-models.md) - AgencyState enum, Publication model, Tag model
- [State Transitions](./state-transitions.md) - State machine logic, guards, and actions

## ArangoDB Collections

### 1. agency_publications Collection

**Purpose**: Store published versions of agencies with deployment manifests.

**Document Schema**:
```json
{
  "_key": "PUB-001-20250120-v1.0.0",
  "_id": "agency_publications/PUB-001-20250120-v1.0.0",
  "_rev": "_gVPKa1G---",
  
  "agency_id": "UC-INFRA-001",
  "version": "v1.0.0",
  "tag_id": "UC-INFRA-001-release-v1.0.0",
  "description": "Initial production release",
  
  "snapshot": {
    "specification": { /* complete agency spec */ },
    "ai_policy": { /* AI policy config */ },
    "settings": { /* agency settings */ },
    "metadata": { /* agency metadata */ }
  },
  
  "manifest": {
    "agent_spawn_plan": {
      "total_agents": 5,
      "agents": [
        {
          "role_code": "NW-MON",
          "name": "Network Monitor",
          "type": "autonomous",
          "autonomy_level": "high",
          "resource_limits": {
            "cpu_limit": "500m",
            "memory_limit": "256Mi",
            "token_budget": 10000
          },
          "configuration": {}
        }
      ]
    },
    "workflow_execution": {
      "workflows": [
        {
          "workflow_id": "wf-001",
          "name": "Leak Detection",
          "enabled": true,
          "auto_start": true
        }
      ]
    },
    "resource_allocation": {
      "total_cpu": "2000m",
      "total_memory": "1Gi",
      "max_agents": 10
    },
    "monitoring_config": {
      "enabled": true,
      "metrics_endpoint": "/metrics",
      "alerts": ["agent_spawn_failure", "workflow_timeout"]
    }
  },
  
  "published_at": "2025-01-20T10:00:00Z",
  "published_by": "admin@example.com",
  "activated_at": "2025-01-20T10:05:00Z",
  "deactivated_at": null,
  
  "metadata": {
    "deployment_environment": "production",
    "git_commit": "abc123"
  }
}
```

**Indexes**:

```javascript
// Index by agency_id (find all publications for an agency)
db.agency_publications.ensureIndex({
  type: "hash",
  fields: ["agency_id"]
});

// Index by published_at (chronological queries)
db.agency_publications.ensureIndex({
  type: "skiplist",
  fields: ["published_at"]
});

// Index by version (semantic version queries)
db.agency_publications.ensureIndex({
  type: "hash",
  fields: ["version"]
});

// Composite index for active publication lookup
db.agency_publications.ensureIndex({
  type: "hash",
  fields: ["agency_id", "deactivated_at"]
});
```

**Common Queries**:

```javascript
// Get active publication for agency
FOR pub IN agency_publications
  FILTER pub.agency_id == @agencyID
  FILTER pub.deactivated_at == null
  SORT pub.published_at DESC
  LIMIT 1
  RETURN pub

// Get publication history
FOR pub IN agency_publications
  FILTER pub.agency_id == @agencyID
  SORT pub.published_at DESC
  RETURN {
    version: pub.version,
    published_at: pub.published_at,
    published_by: pub.published_by,
    activated: pub.activated_at != null
  }
```

### 2. agency_tags Collection

**Purpose**: Store immutable snapshots of agencies for version control.

**Document Schema**:
```json
{
  "_key": "UC-INFRA-001-release-v1.0.0",
  "_id": "agency_tags/UC-INFRA-001-release-v1.0.0",
  "_rev": "_gVPKa1H---",
  
  "agency_id": "UC-INFRA-001",
  "name": "release-v1.0.0",
  "version": "v1.0.0",
  "description": "Production release version 1.0.0",
  "type": "release",
  "sha": "a3f7c2e9b1d4...",
  
  "snapshot": {
    "specification": { /* complete agency spec */ },
    "ai_policy": { /* AI policy config */ },
    "settings": { /* agency settings */ },
    "metadata": { /* agency metadata */ }
  },
  
  "metadata": {
    "git_commit": "abc123",
    "build_number": "42",
    "environment": "production",
    "custom_fields": {}
  },
  
  "created_at": "2025-01-20T09:30:00Z",
  "created_by": "admin@example.com"
}
```

**Indexes**:

```javascript
// Index by agency_id (find all tags for an agency)
db.agency_tags.ensureIndex({
  type: "hash",
  fields: ["agency_id"]
});

// Unique index on agency_id + name (prevent duplicate tag names)
db.agency_tags.ensureIndex({
  type: "hash",
  fields: ["agency_id", "name"],
  unique: true
});

// Index by semantic version (version-based queries)
db.agency_tags.ensureIndex({
  type: "skiplist",
  fields: ["version"]
});

// Index by creation time (chronological queries)
db.agency_tags.ensureIndex({
  type: "skiplist",
  fields: ["created_at"]
});

// Index by tag type (filter by release, snapshot, etc.)
db.agency_tags.ensureIndex({
  type: "hash",
  fields: ["type"]
});
```

**Common Queries**:

```javascript
// Get all tags for agency
FOR tag IN agency_tags
  FILTER tag.agency_id == @agencyID
  SORT tag.created_at DESC
  RETURN tag

// Get tag by name (unique per agency)
FOR tag IN agency_tags
  FILTER tag.agency_id == @agencyID
  FILTER tag.name == @tagName
  RETURN tag

// Get latest release tag
FOR tag IN agency_tags
  FILTER tag.agency_id == @agencyID
  FILTER tag.type == "release"
  SORT tag.created_at DESC
  LIMIT 1
  RETURN tag

// Get tags by semantic version range
FOR tag IN agency_tags
  FILTER tag.agency_id == @agencyID
  FILTER tag.version >= "v1.0.0"
  FILTER tag.version < "v2.0.0"
  SORT tag.version DESC
  RETURN tag
```

## Database Migration Script

**File**: `internal/database/migrations/006_agency_publishing.go`

```go
package migrations

import (
    "context"
    "fmt"
    
    "github.com/arangodb/go-driver"
)

// Migration006AgencyPublishing adds publishing/tagging collections
func Migration006AgencyPublishing(ctx context.Context, db driver.Database) error {
    // Create agency_publications collection
    if err := createPublicationsCollection(ctx, db); err != nil {
        return fmt.Errorf("failed to create publications collection: %w", err)
    }
    
    // Create agency_tags collection
    if err := createTagsCollection(ctx, db); err != nil {
        return fmt.Errorf("failed to create tags collection: %w", err)
    }
    
    return nil
}

func createPublicationsCollection(ctx context.Context, db driver.Database) error {
    // Create collection
    _, err := db.CreateCollection(ctx, "agency_publications", nil)
    if err != nil && !driver.IsConflict(err) {
        return err
    }
    
    col, err := db.Collection(ctx, "agency_publications")
    if err != nil {
        return err
    }
    
    // Index 1: agency_id (hash index for lookups)
    _, _, err = col.EnsureHashIndex(ctx, []string{"agency_id"}, nil)
    if err != nil {
        return fmt.Errorf("failed to create agency_id index: %w", err)
    }
    
    // Index 2: published_at (skiplist for time-based queries)
    _, _, err = col.EnsureSkipListIndex(ctx, []string{"published_at"}, nil)
    if err != nil {
        return fmt.Errorf("failed to create published_at index: %w", err)
    }
    
    // Index 3: version (hash index for version lookups)
    _, _, err = col.EnsureHashIndex(ctx, []string{"version"}, nil)
    if err != nil {
        return fmt.Errorf("failed to create version index: %w", err)
    }
    
    // Index 4: composite agency_id + deactivated_at (for active publication queries)
    _, _, err = col.EnsureHashIndex(ctx, []string{"agency_id", "deactivated_at"}, nil)
    if err != nil {
        return fmt.Errorf("failed to create composite index: %w", err)
    }
    
    return nil
}

func createTagsCollection(ctx context.Context, db driver.Database) error {
    // Create collection
    _, err := db.CreateCollection(ctx, "agency_tags", nil)
    if err != nil && !driver.IsConflict(err) {
        return err
    }
    
    col, err := db.Collection(ctx, "agency_tags")
    if err != nil {
        return err
    }
    
    // Index 1: agency_id (hash index for lookups)
    _, _, err = col.EnsureHashIndex(ctx, []string{"agency_id"}, nil)
    if err != nil {
        return fmt.Errorf("failed to create agency_id index: %w", err)
    }
    
    // Index 2: Unique index on agency_id + name (prevent duplicates)
    _, _, err = col.EnsureHashIndex(ctx, []string{"agency_id", "name"}, &driver.EnsureHashIndexOptions{
        Unique: true,
    })
    if err != nil {
        return fmt.Errorf("failed to create unique name index: %w", err)
    }
    
    // Index 3: version (skiplist for semantic version queries)
    _, _, err = col.EnsureSkipListIndex(ctx, []string{"version"}, nil)
    if err != nil {
        return fmt.Errorf("failed to create version index: %w", err)
    }
    
    // Index 4: created_at (skiplist for chronological queries)
    _, _, err = col.EnsureSkipListIndex(ctx, []string{"created_at"}, nil)
    if err != nil {
        return fmt.Errorf("failed to create created_at index: %w", err)
    }
    
    // Index 5: type (hash index for tag type filtering)
    _, _, err = col.EnsureHashIndex(ctx, []string{"type"}, nil)
    if err != nil {
        return fmt.Errorf("failed to create type index: %w", err)
    }
    
    return nil
}

// Rollback function (for migration testing/rollback)
func Rollback006AgencyPublishing(ctx context.Context, db driver.Database) error {
    // Drop agency_publications collection
    col, err := db.Collection(ctx, "agency_publications")
    if err == nil {
        if err := col.Remove(ctx); err != nil {
            return fmt.Errorf("failed to drop publications collection: %w", err)
        }
    }
    
    // Drop agency_tags collection
    col, err = db.Collection(ctx, "agency_tags")
    if err == nil {
        if err := col.Remove(ctx); err != nil {
            return fmt.Errorf("failed to drop tags collection: %w", err)
        }
    }
    
    return nil
}
```

**Migration Registration**:

Add to `internal/database/migrations/migrations.go`:

```go
var Migrations = []Migration{
    // ... existing migrations ...
    {
        ID:          "006",
        Description: "Add agency publishing and tagging collections",
        Up:          Migration006AgencyPublishing,
        Down:        Rollback006AgencyPublishing,
    },
}
```

## Data Migration: Status → State

**Purpose**: Convert existing agencies from old `Status` field to new `State` field.

**File**: `scripts/migrate-agency-status-to-state.go`

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "internal/agency/models"
    "internal/database"
)

// Status → State Mapping Logic
var statusToStateMapping = map[models.AgencyStatus]models.AgencyState{
    // Simple mappings
    models.AgencyStatusPaused:   models.AgencyStatePaused,
    models.AgencyStatusArchived: models.AgencyStateArchived,
    
    // Context-dependent mappings (see migrateAgencyState)
    models.AgencyStatusActive:   "", // Depends on ActiveAgentCount
    models.AgencyStatusInactive: "", // Depends on PublicationID
}

func migrateAgencyState(agency *models.Agency) models.AgencyState {
    // Handle simple mappings
    if state, ok := statusToStateMapping[agency.Status]; ok && state != "" {
        return state
    }
    
    // Handle context-dependent mappings
    switch agency.Status {
    case models.AgencyStatusActive:
        // Active with agents → active state
        // Active without agents → published state (was activated but agents stopped)
        if agency.ActiveAgentCount > 0 {
            return models.AgencyStateActive
        }
        return models.AgencyStatePublished
        
    case models.AgencyStatusInactive:
        // Inactive with publication → stopped (was published but shut down)
        // Inactive without publication → draft (never published)
        if agency.PublicationID != "" {
            return models.AgencyStateStopped
        }
        return models.AgencyStateDraft
        
    default:
        // Fallback to draft for unknown states
        log.Printf("WARNING: Unknown status '%s' for agency %s, defaulting to draft", 
            agency.Status, agency.ID)
        return models.AgencyStateDraft
    }
}

func main() {
    ctx := context.Background()
    
    // Connect to database
    client, err := database.NewClient(ctx, database.Config{
        Host:     "localhost",
        Port:     8529,
        Database: "codevaldcortex",
    })
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    
    db := client.Database()
    col, err := db.Collection(ctx, "agencies")
    if err != nil {
        log.Fatalf("Failed to get agencies collection: %v", err)
    }
    
    // Query all agencies
    query := "FOR a IN agencies RETURN a"
    cursor, err := db.Query(ctx, query, nil)
    if err != nil {
        log.Fatalf("Failed to query agencies: %v", err)
    }
    defer cursor.Close()
    
    updated := 0
    skipped := 0
    
    // Process each agency
    for cursor.HasMore() {
        var agency models.Agency
        _, err := cursor.ReadDocument(ctx, &agency)
        if err != nil {
            log.Printf("Failed to read agency: %v", err)
            continue
        }
        
        // Skip if State already set
        if agency.State != "" {
            skipped++
            continue
        }
        
        // Compute new State from Status
        newState := migrateAgencyState(&agency)
        
        // Update agency with new State field
        updateQuery := `
            UPDATE @key WITH { state: @state } IN agencies
            RETURN NEW
        `
        bindVars := map[string]interface{}{
            "key":   agency.Key,
            "state": newState,
        }
        
        _, err = db.Query(ctx, updateQuery, bindVars)
        if err != nil {
            log.Printf("Failed to update agency %s: %v", agency.ID, err)
            continue
        }
        
        log.Printf("Migrated %s: %s → %s", agency.ID, agency.Status, newState)
        updated++
    }
    
    fmt.Printf("\nMigration complete:\n")
    fmt.Printf("  Updated: %d agencies\n", updated)
    fmt.Printf("  Skipped: %d agencies (already have State)\n", skipped)
}
```

**Migration Command**:

```bash
# Add to Makefile
migrate-state:
    go run scripts/migrate-agency-status-to-state.go

# Run migration
make migrate-state
```

**Validation Query**:

After migration, verify all agencies have State field:

```javascript
// Count agencies by state
FOR a IN agencies
  COLLECT state = a.state WITH COUNT INTO count
  RETURN { state: state, count: count }

// Find agencies missing State field
FOR a IN agencies
  FILTER a.state == null OR a.state == ""
  RETURN { id: a._id, status: a.status }
```

## Testing Strategy

**Integration Tests** (`database/migrations_test.go`):

```go
func TestMigration006AgencyPublishing(t *testing.T) {
    ctx := context.Background()
    db := setupTestDatabase(t)
    defer teardownTestDatabase(t, db)
    
    // Run migration
    err := Migration006AgencyPublishing(ctx, db)
    assert.NoError(t, err)
    
    // Verify agency_publications collection exists
    exists, err := db.CollectionExists(ctx, "agency_publications")
    assert.NoError(t, err)
    assert.True(t, exists)
    
    // Verify agency_tags collection exists
    exists, err = db.CollectionExists(ctx, "agency_tags")
    assert.NoError(t, err)
    assert.True(t, exists)
    
    // Verify indexes created
    col, _ := db.Collection(ctx, "agency_publications")
    indexes, _ := col.Indexes(ctx)
    assert.GreaterOrEqual(t, len(indexes), 4) // 4 custom indexes + primary
    
    col, _ = db.Collection(ctx, "agency_tags")
    indexes, _ = col.Indexes(ctx)
    assert.GreaterOrEqual(t, len(indexes), 5) // 5 custom indexes + primary
}

func TestStatusToStateMigration(t *testing.T) {
    tests := []struct {
        name             string
        status           models.AgencyStatus
        activeAgentCount int
        publicationID    string
        expectedState    models.AgencyState
    }{
        {
            name:          "active with agents",
            status:        models.AgencyStatusActive,
            activeAgentCount: 5,
            expectedState: models.AgencyStateActive,
        },
        {
            name:          "active without agents",
            status:        models.AgencyStatusActive,
            activeAgentCount: 0,
            expectedState: models.AgencyStatePublished,
        },
        {
            name:          "inactive never published",
            status:        models.AgencyStatusInactive,
            publicationID: "",
            expectedState: models.AgencyStateDraft,
        },
        {
            name:          "inactive was published",
            status:        models.AgencyStatusInactive,
            publicationID: "PUB-001",
            expectedState: models.AgencyStateStopped,
        },
        {
            name:          "paused",
            status:        models.AgencyStatusPaused,
            expectedState: models.AgencyStatePaused,
        },
        {
            name:          "archived",
            status:        models.AgencyStatusArchived,
            expectedState: models.AgencyStateArchived,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            agency := &models.Agency{
                Status:           tt.status,
                ActiveAgentCount: tt.activeAgentCount,
                PublicationID:    tt.publicationID,
            }
            
            state := migrateAgencyState(agency)
            assert.Equal(t, tt.expectedState, state)
        })
    }
}
```

## Acceptance Criteria

- [x] ArangoDB collections created: `agency_publications`, `agency_tags`
- [x] Proper indexes created on both collections
- [x] Database migration script created (Migration006)
- [x] Migration script handles collection conflicts (already exists)
- [x] Rollback function defined for migration testing
- [x] Data migration script created for Status→State conversion
- [x] Status→State mapping logic handles all edge cases
- [x] Validation queries defined for post-migration verification
- [x] Integration tests verify collection and index creation
- [x] Unit tests verify Status→State migration logic
- [x] Migration is idempotent (safe to run multiple times)
