# MVP-PUB-002: Tag Service Implementation

**Task ID**: MVP-PUB-002  
**Feature Branch**: `feature/MVP-PUB-002_tag_service`  
**Status**: 🔄 In Progress  
**Started**: 2025-01-19  
**Assigned To**: GitHub Copilot  
**Priority**: P0 (CRITICAL)

---

## 📋 Task Overview

### Objective
Implement the Tag Service layer for agency versioning and snapshots. This includes the TagService interface, TagRepository (ArangoDB), snapshot generation with SHA content hashing, tag comparison/diff logic, and HTTP handlers for tag CRUD operations.

### Dependencies
- ✅ MVP-PUB-001: Agency State Machine & Data Models (completed)
  - AgencyTag model available
  - TagType enum (release, snapshot, experimental, checkpoint)
  - Database migration 006 created agency_tags collection
  - Tag validation functions available

### Scope
**In Scope**:
- TagService interface and concrete implementation
- TagRepository interface and ArangoDB implementation
- Snapshot generation (deep copy of agency specification)
- SHA generation for content hashing (git-style)
- Tag CRUD operations (create, list, get, delete)
- Tag comparison logic (generate diffs between tags)
- Restore from tag functionality
- HTTP handlers for tag endpoints
- Unit tests for service layer
- Integration tests for repository layer

**Out of Scope**:
- Tag publishing (handled in MVP-PUB-003)
- UI components (handled in MVP-PUB-005)
- Agent spawn orchestration (handled in MVP-PUB-004)
- Publication workflow integration (handled in MVP-PUB-003)

---

## 🎯 Acceptance Criteria

### Core Functionality
- [ ] TagService interface defined with all required methods
- [ ] CreateTag() generates snapshot and SHA hash
- [ ] ListTags() supports filtering by type, name, date range
- [ ] GetTag() retrieves tag by agency ID and tag name
- [ ] CompareTags() generates path-based diff between two tags
- [ ] RestoreFromTag() overwrites agency draft with tag snapshot
- [ ] DeleteTag() removes tag from database
- [ ] TagRepository implements CRUD operations on agency_tags collection
- [ ] Snapshot generation creates deep copy of specification, policy, settings, metadata
- [ ] SHA generation produces consistent content hashes

### API Endpoints
- [ ] POST /api/v1/agencies/:id/tags (create tag)
- [ ] GET /api/v1/agencies/:id/tags (list tags with filters)
- [ ] GET /api/v1/agencies/:id/tags/:name (get specific tag)
- [ ] DELETE /api/v1/agencies/:id/tags/:name (delete tag)
- [ ] POST /api/v1/agencies/:id/tags/:name/restore (restore from tag)
- [ ] GET /api/v1/tags/:tag1/compare/:tag2 (compare two tags)

### Testing
- [ ] Unit tests for TagService (with mock repository)
- [ ] Integration tests for TagRepository (with ArangoDB)
- [ ] Test snapshot generation consistency
- [ ] Test SHA hashing consistency
- [ ] Test tag comparison diff generation
- [ ] Test restore from tag overwrites correctly
- [ ] All tests passing
- [ ] No compilation errors

### Code Quality
- [ ] Follow template-first architecture (no HTML in Go handlers)
- [ ] Keep file sizes under limits (services <400 lines, repositories <250 lines)
- [ ] Use functional programming principles (testable, pure functions)
- [ ] Consistent error handling
- [ ] Proper logging
- [ ] Clear separation of concerns

---

## 🏗️ Technical Approach

### 1. TagService Interface

**File**: `internal/agency/services/tag_service.go`

**Interface Definition**:
```go
package services

import (
    "context"
    "github.com/aosanya/CodeValdCortex/internal/agency/models"
)

type TagService interface {
    // Create tag from current agency state
    CreateTag(ctx context.Context, agencyID string, req *CreateTagRequest) (*models.AgencyTag, error)
    
    // List all tags for agency
    ListTags(ctx context.Context, agencyID string, filters *TagFilters) ([]*models.AgencyTag, error)
    
    // Get specific tag
    GetTag(ctx context.Context, agencyID string, tagName string) (*models.AgencyTag, error)
    
    // Compare two tags (generate diff)
    CompareTags(ctx context.Context, tagID1, tagID2 string) (*TagComparison, error)
    
    // Restore agency from tag (overwrite current draft)
    RestoreFromTag(ctx context.Context, agencyID string, tagName string) error
    
    // Delete tag
    DeleteTag(ctx context.Context, agencyID string, tagName string) error
}
```

**Request/Response Types**:
```go
type CreateTagRequest struct {
    Name        string            `json:"name" binding:"required"`
    Version     string            `json:"version,omitempty"`
    Description string            `json:"description" binding:"required"`
    Type        models.TagType    `json:"type" binding:"required"`
    Metadata    map[string]string `json:"metadata,omitempty"`
}

type TagFilters struct {
    Type      models.TagType `json:"type,omitempty"`
    NameLike  string         `json:"name_like,omitempty"`
    FromDate  *time.Time     `json:"from_date,omitempty"`
    ToDate    *time.Time     `json:"to_date,omitempty"`
    Limit     int            `json:"limit,omitempty"`
    Offset    int            `json:"offset,omitempty"`
}

type TagComparison struct {
    Tag1         *models.AgencyTag `json:"tag1"`
    Tag2         *models.AgencyTag `json:"tag2"`
    Differences  []TagDifference   `json:"differences"`
    Summary      string            `json:"summary"`
}

type TagDifference struct {
    Path     string      `json:"path"`
    Type     string      `json:"type"` // added, removed, modified
    OldValue interface{} `json:"old_value,omitempty"`
    NewValue interface{} `json:"new_value,omitempty"`
}
```

**Concrete Implementation**:
```go
type tagService struct {
    tagRepo    TagRepository
    agencyRepo AgencyRepository
    logger     *slog.Logger
}

func NewTagService(tagRepo TagRepository, agencyRepo AgencyRepository, logger *slog.Logger) TagService {
    return &tagService{
        tagRepo:    tagRepo,
        agencyRepo: agencyRepo,
        logger:     logger,
    }
}
```

### 2. TagRepository Interface

**File**: `internal/agency/arangodb/tag_repository.go`

**Interface Definition**:
```go
package arangodb

import (
    "context"
    "github.com/aosanya/CodeValdCortex/internal/agency/models"
    "github.com/aosanya/CodeValdCortex/internal/agency/services"
)

type TagRepository interface {
    Create(ctx context.Context, tag *models.AgencyTag) error
    GetByID(ctx context.Context, tagID string) (*models.AgencyTag, error)
    GetByAgencyAndName(ctx context.Context, agencyID, name string) (*models.AgencyTag, error)
    List(ctx context.Context, agencyID string, filters *services.TagFilters) ([]*models.AgencyTag, error)
    Delete(ctx context.Context, agencyID, name string) error
}
```

**ArangoDB Implementation**:
- Use `agency_tags` collection (created by migration 006)
- Leverage hash index on `agency_id`
- Leverage unique index on `agency_id` + `name`
- Leverage skiplist index on `created_at`
- Build AQL queries with filters (type, name LIKE, date ranges)
- Handle pagination with LIMIT/OFFSET

### 3. Snapshot Generation

**Function**: `generateSnapshot(agency *models.Agency) *models.AgencySnapshot`

**Responsibilities**:
- Deep copy AgencySpecification (all nested fields)
- Deep copy AIPolicy
- Deep copy AgencySettings
- Deep copy AgencyMetadata
- Generate SHA hash from serialized snapshot (JSON canonical form)

**SHA Generation Algorithm**:
```go
func generateContentHash(snapshot *models.AgencySnapshot) (string, error) {
    // 1. Serialize snapshot to JSON (canonical, sorted keys)
    data, err := json.Marshal(snapshot)
    if err != nil {
        return "", err
    }
    
    // 2. Compute SHA-256 hash
    hash := sha256.Sum256(data)
    
    // 3. Return hex-encoded hash (git-style)
    return hex.EncodeToString(hash[:]), nil
}
```

### 4. Tag Comparison Logic

**Function**: `compareTags(tag1, tag2 *models.AgencyTag) (*services.TagComparison, error)`

**Diff Algorithm**:
- Serialize both snapshots to JSON
- Use recursive JSON path traversal
- Detect added fields (in tag2, not in tag1)
- Detect removed fields (in tag1, not in tag2)
- Detect modified fields (different values)
- Generate path notation (e.g., "specification.roles[0].name")
- Generate summary string (e.g., "42 changes: 5 added, 3 removed, 34 modified")

### 5. HTTP Handlers

**File**: `internal/handlers/tag_handler.go`

**Handler Methods**:
```go
type TagHandler struct {
    tagService services.TagService
    logger     *slog.Logger
}

func NewTagHandler(tagService services.TagService, logger *slog.Logger) *TagHandler {
    return &TagHandler{
        tagService: tagService,
        logger:     logger,
    }
}

func (h *TagHandler) CreateTag(c *gin.Context) { /* POST /api/v1/agencies/:id/tags */ }
func (h *TagHandler) ListTags(c *gin.Context) { /* GET /api/v1/agencies/:id/tags */ }
func (h *TagHandler) GetTag(c *gin.Context) { /* GET /api/v1/agencies/:id/tags/:name */ }
func (h *TagHandler) DeleteTag(c *gin.Context) { /* DELETE /api/v1/agencies/:id/tags/:name */ }
func (h *TagHandler) RestoreFromTag(c *gin.Context) { /* POST /api/v1/agencies/:id/tags/:name/restore */ }
func (h *TagHandler) CompareTags(c *gin.Context) { /* GET /api/v1/tags/:tag1/compare/:tag2 */ }
```

**Route Registration** (in `internal/api/router.go`):
```go
// Tag management
agencyRoutes := v1.Group("/agencies")
{
    agencyRoutes.POST("/:id/tags", tagHandler.CreateTag)
    agencyRoutes.GET("/:id/tags", tagHandler.ListTags)
    agencyRoutes.GET("/:id/tags/:name", tagHandler.GetTag)
    agencyRoutes.DELETE("/:id/tags/:name", tagHandler.DeleteTag)
    agencyRoutes.POST("/:id/tags/:name/restore", tagHandler.RestoreFromTag)
}

tagRoutes := v1.Group("/tags")
{
    tagRoutes.GET("/:tag1/compare/:tag2", tagHandler.CompareTags)
}
```

---

## 🧪 Testing Strategy

### Unit Tests

**File**: `internal/agency/services/tag_service_test.go`

**Test Cases**:
1. **TestCreateTag_Success**
   - Create tag with valid request
   - Verify snapshot generated
   - Verify SHA hash created
   - Verify tag saved to repository

2. **TestCreateTag_DuplicateName**
   - Attempt to create tag with existing name
   - Verify error returned

3. **TestListTags_WithFilters**
   - List tags with type filter
   - List tags with name filter (LIKE)
   - List tags with date range filter
   - Verify pagination (limit/offset)

4. **TestGetTag_Success**
   - Retrieve existing tag
   - Verify all fields populated

5. **TestCompareTags_WithDifferences**
   - Compare two tags with different snapshots
   - Verify differences detected
   - Verify summary generated

6. **TestRestoreFromTag_Success**
   - Restore agency from tag
   - Verify agency specification updated
   - Verify state remains draft

7. **TestDeleteTag_Success**
   - Delete existing tag
   - Verify tag removed from repository

8. **TestGenerateSnapshot_Consistency**
   - Generate snapshot twice from same agency
   - Verify SHA hashes match

### Integration Tests

**File**: `internal/agency/arangodb/tag_repository_test.go`

**Test Cases**:
1. **TestTagRepository_Create**
   - Insert tag into ArangoDB
   - Verify tag stored correctly

2. **TestTagRepository_GetByID**
   - Retrieve tag by ID
   - Verify all fields match

3. **TestTagRepository_GetByAgencyAndName**
   - Retrieve tag by agency ID and name
   - Verify unique constraint works

4. **TestTagRepository_List**
   - Query with various filters
   - Verify results match filters

5. **TestTagRepository_Delete**
   - Delete tag by agency ID and name
   - Verify tag no longer exists

---

## 📁 Files to Create/Modify

### New Files
1. `documents/3-SofwareDevelopment/coding_sessions/MVP-PUB-002_tag_service.md` (this file)
2. `internal/agency/services/tag_service.go` (~400 lines)
3. `internal/agency/arangodb/tag_repository.go` (~250 lines)
4. `internal/handlers/tag_handler.go` (~300 lines)
5. `internal/agency/services/tag_service_test.go` (~400 lines)
6. `internal/agency/arangodb/tag_repository_test.go` (~200 lines)

### Modified Files
1. `internal/api/router.go` (+20 lines for tag routes)

**Total Estimated LOC**: ~1,570 new lines

---

## 🚀 Implementation Checklist

### Phase 1: Service Layer
- [ ] Create `tag_service.go` with interface definition
- [ ] Implement CreateTag() with snapshot generation
- [ ] Implement generateSnapshot() with deep copy logic
- [ ] Implement generateContentHash() with SHA-256
- [ ] Implement ListTags() with filter support
- [ ] Implement GetTag() retrieval
- [ ] Implement CompareTags() with diff generation
- [ ] Implement RestoreFromTag() with agency update
- [ ] Implement DeleteTag() with validation

### Phase 2: Repository Layer
- [ ] Create `tag_repository.go` with interface
- [ ] Implement Create() with ArangoDB insert
- [ ] Implement GetByID() with AQL query
- [ ] Implement GetByAgencyAndName() with compound query
- [ ] Implement List() with filter builder
- [ ] Implement Delete() with AQL delete

### Phase 3: HTTP Handlers
- [ ] Create `tag_handler.go` with handler struct
- [ ] Implement CreateTag handler (POST)
- [ ] Implement ListTags handler (GET with query params)
- [ ] Implement GetTag handler (GET)
- [ ] Implement DeleteTag handler (DELETE)
- [ ] Implement RestoreFromTag handler (POST)
- [ ] Implement CompareTags handler (GET)
- [ ] Add route registration in router.go

### Phase 4: Testing
- [ ] Write TagService unit tests
- [ ] Write TagRepository integration tests
- [ ] Test snapshot generation consistency
- [ ] Test SHA hashing consistency
- [ ] Test tag comparison diff generation
- [ ] Test restore from tag functionality
- [ ] Verify all tests passing
- [ ] Verify no compilation errors

### Phase 5: Integration
- [ ] Run `go build ./...`
- [ ] Run `go test ./internal/agency/...`
- [ ] Fix any issues
- [ ] Commit changes
- [ ] Update this document with results

---

## 📊 Progress Tracking

### Session Log

**2025-01-19 - Session 1: Task Setup**
- Created feature branch `feature/MVP-PUB-002_tag_service`
- Created task specification document
- Defined todo list with 10 items
- Ready to begin implementation

---

## 🔗 References

- **Architecture Document**: `/documents/2-SoftwareDesignAndArchitecture/agency-publishing-tagging-architecture.md`
- **MVP Task List**: `/documents/3-SofwareDevelopment/mvp.md` (line 71)
- **Dependency (MVP-PUB-001)**: `/documents/3-SofwareDevelopment/coding_sessions/MVP-PUB-001_state_machine_data_models.md`
- **AgencyTag Model**: `/internal/agency/models/tag.go`
- **Database Migration**: `/internal/database/migrations/006_agency_publishing.go`

---

## 💡 Design Decisions

### 1. Snapshot Immutability
- Tags create immutable snapshots (deep copies)
- Original agency can be modified without affecting tags
- SHA hashing ensures content integrity

### 2. SHA Generation Algorithm
- Using SHA-256 for content hashing (git-style)
- JSON canonical form (sorted keys) for consistency
- Hex-encoded output for readability

### 3. Tag Comparison
- Path-based diff notation (JSON path)
- Three diff types: added, removed, modified
- Summary string for quick overview

### 4. Restore from Tag
- Overwrites agency specification, policy, settings, metadata
- Agency state remains "draft" (not published)
- Creates audit trail entry

### 5. Repository Pattern
- TagRepository abstraction for testability
- ArangoDB implementation leverages existing indexes
- Filter builder pattern for flexible queries

---

## ⚠️ Known Limitations

1. **Tag Size**: No limit on snapshot size (could be large for complex agencies)
2. **Comparison Performance**: Diff generation may be slow for large snapshots
3. **Concurrent Updates**: No optimistic locking for agency updates during restore
4. **Soft Delete**: Tags are hard-deleted (no soft delete/archive)

---

## 🔜 Future Enhancements

1. **Tag Compression**: Compress snapshots to reduce storage
2. **Incremental Snapshots**: Store only diffs from previous tag
3. **Tag Expiration**: Auto-delete old experimental/snapshot tags
4. **Tag Locking**: Prevent deletion of published tags
5. **Tag Metadata Search**: Full-text search on tag metadata

---

## 📝 Notes

- Following template-first architecture (no HTML in handlers)
- Using functional programming principles (testable, pure functions)
- Keeping file sizes under limits (services <400 lines, repositories <250 lines)
- All API endpoints return JSON (no HTML rendering in this task)
- UI components will be added in MVP-PUB-005
