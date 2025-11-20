# MVP-PUB-003: Publication Service Implementation

**Task ID**: MVP-PUB-003  
**Feature Branch**: `feature/MVP-PUB-003_publication_service`  
**Status**: 🔄 In Progress  
**Started**: 2025-11-20  
**Assigned To**: GitHub Copilot  
**Priority**: P0 (CRITICAL)

---

## 📋 Task Overview

### Objective
Implement the Publication Service layer for agency publishing and lifecycle management. This includes validation logic, publication workflow, deployment manifest generation, publication snapshot storage, and publication history tracking.

### Dependencies
- ✅ MVP-PUB-001: Agency State Machine & Data Models (completed)
- ✅ MVP-PUB-002: Tag Service Implementation (completed)
  - AgencyPublication model available
  - AgencyTag model available
  - DeploymentManifest model available
  - Database migration 006 created agency_publications collection
  - State machine with transition guards/actions

### Scope
**In Scope**:
- PublicationService interface and concrete implementation
- Pre-publish validation logic (specification, roles, workflows)
- Publication workflow (draft → validated → published)
- Deployment manifest generation (agent spawn plan, workflows, resources)
- Publication snapshot storage with versioning
- Publication history tracking
- HTTP handlers for publication endpoints
- Repository layer for agency_publications collection

**Out of Scope**:
- Agent spawning (handled in MVP-PUB-004)
- Workflow execution (handled in MVP-PUB-004)
- UI components (handled in MVP-PUB-005)
- Tag publishing integration (deferred)

---

## 🎯 Acceptance Criteria

### Core Functionality
- [ ] PublicationService interface defined with all required methods
- [ ] ValidateForPublish() checks specification completeness, roles, workflows
- [ ] Publish() creates immutable publication with snapshot and manifest
- [ ] Activate() transitions agency state (published → active) - stub for MVP-PUB-004
- [ ] Deactivate() transitions agency state (active → stopped) - stub for MVP-PUB-004
- [ ] GetPublicationHistory() retrieves all publications for agency
- [ ] Deployment manifest generation computes agent spawn plan from roles
- [ ] Publication snapshot includes specification, settings, metadata
- [ ] Semantic versioning support (v1.0.0, v1.1.0, etc.)
- [ ] PublicationRepository implements CRUD operations

### API Endpoints
- [ ] POST /api/v1/agencies/:id/validate (validate for publishing)
- [ ] POST /api/v1/agencies/:id/publish (publish agency)
- [ ] POST /api/v1/agencies/:id/activate (activate published agency)
- [ ] POST /api/v1/agencies/:id/deactivate (deactivate agency)
- [ ] GET /api/v1/agencies/:id/publications (get publication history)
- [ ] POST /api/v1/publications/:id/activate (activate specific publication)

### Validation Rules
- [ ] Agency must have specification with introduction
- [ ] Agency must have at least one goal defined
- [ ] Agency must have at least one role defined
- [ ] Agency must have at least one workflow defined
- [ ] Agency must be in draft or validated state to publish
- [ ] Version must follow semantic versioning (v1.0.0)
- [ ] Cannot publish if validation errors exist

### Testing
- [ ] Unit tests for PublicationService (with mock repository)
- [ ] Integration tests for PublicationRepository (with ArangoDB)
- [ ] Test validation logic (valid/invalid cases)
- [ ] Test manifest generation
- [ ] Test publication workflow
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

### 1. PublicationService Interface

**File**: `internal/agency/services/publication_service.go`

**Interface Definition**:
```go
package services

import (
    "context"
    "github.com/aosanya/CodeValdCortex/internal/agency/models"
)

type PublicationService interface {
    // ValidateForPublish checks if agency is ready to be published
    ValidateForPublish(ctx context.Context, agencyID string) (*ValidationResult, error)
    
    // Publish creates an immutable publication from current agency state
    Publish(ctx context.Context, agencyID string, req *PublishRequest) (*models.AgencyPublication, error)
    
    // Activate transitions agency from published to active (stub for MVP-PUB-004)
    Activate(ctx context.Context, publicationID string) (*ActivationResult, error)
    
    // Deactivate transitions agency from active to stopped
    Deactivate(ctx context.Context, agencyID string, graceful bool) error
    
    // GetPublicationHistory retrieves all publications for an agency
    GetPublicationHistory(ctx context.Context, agencyID string) ([]*models.AgencyPublication, error)
    
    // RepublishFromTag creates new publication from a tag (future)
    RepublishFromTag(ctx context.Context, tagID string) (*models.AgencyPublication, error)
}
```

**Request/Response Types**:
```go
type PublishRequest struct {
    Version      string            `json:"version" binding:"required"`
    Description  string            `json:"description" binding:"required"`
    AutoActivate bool              `json:"auto_activate,omitempty"`
    CreateTag    bool              `json:"create_tag,omitempty"`
    TagName      string            `json:"tag_name,omitempty"`
    Metadata     map[string]string `json:"metadata,omitempty"`
    PublishedBy  string            `json:"published_by,omitempty"`
}

type ValidationResult struct {
    Valid            bool                `json:"valid"`
    Errors           []ValidationError   `json:"errors"`
    Warnings         []ValidationWarning `json:"warnings"`
    Recommendations  []string            `json:"recommendations"`
}

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Code    string `json:"code"`
}

type ValidationWarning struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

type ActivationResult struct {
    AgentsSpawned        int       `json:"agents_spawned"`
    WorkflowsInitialized int       `json:"workflows_initialized"`
    MonitoringEnabled    bool      `json:"monitoring_enabled"`
    ActivatedAt          time.Time `json:"activated_at"`
}
```

**Concrete Implementation**:
```go
type publicationService struct {
    pubRepo      PublicationRepository
    agencyRepo   agency.Repository
    stateMachine *agency.AgencyStateMachine
    validator    *PublisherValidator
    logger       *slog.Logger
}

func NewPublicationService(
    pubRepo PublicationRepository,
    agencyRepo agency.Repository,
    stateMachine *agency.AgencyStateMachine,
    validator *PublisherValidator,
    logger *slog.Logger,
) PublicationService {
    return &publicationService{
        pubRepo:      pubRepo,
        agencyRepo:   agencyRepo,
        stateMachine: stateMachine,
        validator:    validator,
        logger:       logger,
    }
}
```

### 2. PublisherValidator

**File**: `internal/agency/validation/publisher_validator.go`

**Validation Checks**:
```go
package validation

type PublisherValidator struct {
    logger *slog.Logger
}

func NewPublisherValidator(logger *slog.Logger) *PublisherValidator {
    return &PublisherValidator{logger: logger}
}

// ValidateForPublish performs comprehensive pre-publish validation
func (v *PublisherValidator) ValidateForPublish(
    agencyDoc *models.Agency,
    spec *models.AgencySpecification,
) (*services.ValidationResult, error) {
    result := &services.ValidationResult{
        Valid:           true,
        Errors:          []services.ValidationError{},
        Warnings:        []services.ValidationWarning{},
        Recommendations: []string{},
    }
    
    // Check 1: Specification must exist
    if spec == nil {
        result.Errors = append(result.Errors, services.ValidationError{
            Field:   "specification",
            Message: "Agency must have a specification",
            Code:    "SPEC_MISSING",
        })
        result.Valid = false
    }
    
    // Check 2: Introduction must be non-empty
    if spec.Introduction == "" {
        result.Errors = append(result.Errors, services.ValidationError{
            Field:   "specification.introduction",
            Message: "Agency must have an introduction",
            Code:    "INTRO_MISSING",
        })
        result.Valid = false
    }
    
    // Check 3: At least one goal
    if len(spec.Goals) == 0 {
        result.Errors = append(result.Errors, services.ValidationError{
            Field:   "specification.goals",
            Message: "Agency must have at least one goal",
            Code:    "GOALS_MISSING",
        })
        result.Valid = false
    }
    
    // Check 4: At least one role
    if len(spec.Roles) == 0 {
        result.Errors = append(result.Errors, services.ValidationError{
            Field:   "specification.roles",
            Message: "Agency must have at least one role",
            Code:    "ROLES_MISSING",
        })
        result.Valid = false
    }
    
    // Check 5: At least one workflow
    if len(spec.Workflows) == 0 {
        result.Warnings = append(result.Warnings, services.ValidationWarning{
            Field:   "specification.workflows",
            Message: "Agency has no workflows defined",
        })
        result.Recommendations = append(result.Recommendations,
            "Consider adding workflows to orchestrate agent activities")
    }
    
    // Check 6: Agency state must be draft or validated
    if agencyDoc.State != models.AgencyStateDraft && 
       agencyDoc.State != models.AgencyStateValidated {
        result.Errors = append(result.Errors, services.ValidationError{
            Field:   "state",
            Message: fmt.Sprintf("Cannot publish agency in state: %s", agencyDoc.State),
            Code:    "INVALID_STATE",
        })
        result.Valid = false
    }
    
    return result, nil
}
```

### 3. PublicationRepository Interface

**File**: `internal/agency/arangodb/publication_repository.go`

**Interface Definition**:
```go
package arangodb

import (
    "context"
    "github.com/aosanya/CodeValdCortex/internal/agency/models"
)

type PublicationRepository interface {
    Create(ctx context.Context, pub *models.AgencyPublication) error
    GetByID(ctx context.Context, pubID string) (*models.AgencyPublication, error)
    ListByAgency(ctx context.Context, agencyID string) ([]*models.AgencyPublication, error)
    GetLatest(ctx context.Context, agencyID string) (*models.AgencyPublication, error)
    Update(ctx context.Context, pub *models.AgencyPublication) error
}
```

**ArangoDB Implementation**:
- Use `agency_publications` collection (created by migration 006)
- Leverage hash index on `agency_id`
- Leverage skiplist index on `published_at`
- Build AQL queries for history retrieval
- Handle semantic versioning sorting

### 4. Deployment Manifest Generation

**Function**: `generateDeploymentManifest(spec *models.AgencySpecification) *models.DeploymentManifest`

**Responsibilities**:
- Generate AgentSpawnPlan from roles
- Create WorkflowExecution graph
- Allocate resource quotas (CPU, memory, token budgets)
- Generate monitoring configuration

**Agent Spawn Plan Algorithm**:
```go
func (s *publicationService) generateAgentSpawnPlan(spec *models.AgencySpecification) models.AgentSpawnPlan {
    agents := []models.AgentDefinition{}
    
    for _, role := range spec.Roles {
        agent := models.AgentDefinition{
            RoleCode:      role.Code,
            Name:          role.Name,
            Type:          role.Type,
            AutonomyLevel: role.AutonomyLevel,
            ResourceLimits: models.ResourceLimits{
                CPULimit:    "500m",  // Default: 0.5 CPU cores
                MemoryLimit: "512Mi", // Default: 512 MB
                TokenBudget: role.TokenBudget,
            },
            Configuration: map[string]interface{}{
                "role_description": role.Description,
                "responsibilities": role.Responsibilities,
            },
        }
        agents = append(agents, agent)
    }
    
    return models.AgentSpawnPlan{
        TotalAgents: len(agents),
        Agents:      agents,
    }
}
```

### 5. Publication Workflow

**State Transitions**:
```
1. Validate:     draft → validated
2. Publish:      validated → published
3. Activate:     published → active
4. Deactivate:   active → stopped
```

**Publish Algorithm**:
```go
func (s *publicationService) Publish(ctx context.Context, agencyID string, req *PublishRequest) (*models.AgencyPublication, error) {
    // 1. Retrieve current agency
    agency, err := s.agencyRepo.GetByID(ctx, agencyID)
    if err != nil {
        return nil, err
    }
    
    // 2. Validate for publishing
    spec, err := s.agencyRepo.GetSpecification(ctx, agencyID)
    if err != nil {
        return nil, err
    }
    
    validation, err := s.validator.ValidateForPublish(agency, spec)
    if err != nil {
        return nil, err
    }
    
    if !validation.Valid {
        return nil, fmt.Errorf("validation failed: %d errors", len(validation.Errors))
    }
    
    // 3. Transition to validated if in draft
    if agency.State == models.AgencyStateDraft {
        if err := s.stateMachine.Transition(ctx, agency, models.AgencyStateValidated); err != nil {
            return nil, err
        }
    }
    
    // 4. Generate snapshot
    snapshot := models.AgencySnapshot{
        Specification: *spec,
        Settings:      agency.Settings,
        Metadata:      agency.Metadata,
    }
    
    // 5. Generate deployment manifest
    manifest := s.generateDeploymentManifest(spec)
    
    // 6. Create publication
    publication := &models.AgencyPublication{
        AgencyID:    agencyID,
        Version:     req.Version,
        Description: req.Description,
        Snapshot:    snapshot,
        Manifest:    manifest,
        PublishedAt: time.Now(),
        PublishedBy: req.PublishedBy,
        Metadata:    req.Metadata,
    }
    
    // 7. Save publication
    if err := s.pubRepo.Create(ctx, publication); err != nil {
        return nil, err
    }
    
    // 8. Update agency state to published
    if err := s.stateMachine.Transition(ctx, agency, models.AgencyStatePublished); err != nil {
        return nil, err
    }
    
    // 9. Update agency metadata
    agency.PublishedAt = &publication.PublishedAt
    agency.PublishedBy = req.PublishedBy
    agency.PublicationID = publication.Key
    if err := s.agencyRepo.Update(ctx, agency); err != nil {
        return nil, err
    }
    
    // 10. Auto-activate if requested
    if req.AutoActivate {
        _, err := s.Activate(ctx, publication.Key)
        if err != nil {
            s.logger.Warn("auto-activation failed", "error", err)
        }
    }
    
    return publication, nil
}
```

### 6. HTTP Handlers

**File**: `internal/handlers/publication_handler.go`

**Handler Methods**:
```go
type PublicationHandler struct {
    pubService services.PublicationService
    logger     *logrus.Logger
}

func NewPublicationHandler(pubService services.PublicationService, logger *logrus.Logger) *PublicationHandler {
    return &PublicationHandler{
        pubService: pubService,
        logger:     logger,
    }
}

func (h *PublicationHandler) ValidateForPublish(c *gin.Context) { /* POST /api/v1/agencies/:id/validate */ }
func (h *PublicationHandler) Publish(c *gin.Context) { /* POST /api/v1/agencies/:id/publish */ }
func (h *PublicationHandler) Activate(c *gin.Context) { /* POST /api/v1/agencies/:id/activate */ }
func (h *PublicationHandler) Deactivate(c *gin.Context) { /* POST /api/v1/agencies/:id/deactivate */ }
func (h *PublicationHandler) GetPublicationHistory(c *gin.Context) { /* GET /api/v1/agencies/:id/publications */ }
func (h *PublicationHandler) ActivatePublication(c *gin.Context) { /* POST /api/v1/publications/:id/activate */ }
```

**Route Registration** (in `internal/app/app.go`):
```go
// Publication endpoints
if a.publicationService != nil {
    pubHandler := handlers.NewPublicationHandler(*a.publicationService, a.logger)
    v1.POST("/agencies/:id/validate", pubHandler.ValidateForPublish)
    v1.POST("/agencies/:id/publish", pubHandler.Publish)
    v1.POST("/agencies/:id/activate", pubHandler.Activate)
    v1.POST("/agencies/:id/deactivate", pubHandler.Deactivate)
    v1.GET("/agencies/:id/publications", pubHandler.GetPublicationHistory)
    v1.POST("/publications/:id/activate", pubHandler.ActivatePublication)
    a.logger.Info("Publication endpoints registered")
}
```

---

## 🧪 Testing Strategy

### Unit Tests

**File**: `internal/agency/services/publication_service_test.go`

**Test Cases**:
1. **TestValidateForPublish_Valid** - All checks pass
2. **TestValidateForPublish_MissingIntroduction** - Validation error
3. **TestValidateForPublish_NoGoals** - Validation error
4. **TestValidateForPublish_NoRoles** - Validation error
5. **TestPublish_Success** - Creates publication
6. **TestPublish_InvalidState** - Cannot publish from active state
7. **TestGenerateDeploymentManifest** - Correct agent spawn plan
8. **TestGetPublicationHistory** - Retrieves all publications

### Integration Tests

**File**: `internal/agency/arangodb/publication_repository_test.go`

**Test Cases**:
1. **TestPublicationRepository_Create** - Insert publication
2. **TestPublicationRepository_GetByID** - Retrieve by ID
3. **TestPublicationRepository_ListByAgency** - Query by agency_id
4. **TestPublicationRepository_GetLatest** - Get most recent

---

## 📁 Files to Create/Modify

### New Files
1. `documents/3-SofwareDevelopment/coding_sessions/MVP-PUB-003_publication_service.md` (this file)
2. `internal/agency/services/publication_service.go` (~500 lines)
3. `internal/agency/validation/publisher_validator.go` (~200 lines)
4. `internal/agency/arangodb/publication_repository.go` (~250 lines)
5. `internal/handlers/publication_handler.go` (~350 lines)

### Modified Files
1. `internal/app/app.go` (+30 lines for publication service init and routes)

**Total Estimated LOC**: ~1,330 new lines + documentation

---

## 🚀 Implementation Checklist

### Phase 1: Validation Layer
- [ ] Create `publisher_validator.go` with validation logic
- [ ] Implement ValidateForPublish() with all checks
- [ ] Add validation error/warning types
- [ ] Test validation rules

### Phase 2: Repository Layer
- [ ] Create `publication_repository.go` with interface
- [ ] Implement Create() with ArangoDB insert
- [ ] Implement GetByID(), ListByAgency(), GetLatest()
- [ ] Test repository operations

### Phase 3: Service Layer
- [ ] Create `publication_service.go` with interface
- [ ] Implement ValidateForPublish()
- [ ] Implement Publish() with workflow
- [ ] Implement generateDeploymentManifest()
- [ ] Implement Activate() stub
- [ ] Implement Deactivate() stub
- [ ] Implement GetPublicationHistory()

### Phase 4: HTTP Handlers
- [ ] Create `publication_handler.go`
- [ ] Implement all 6 handler methods
- [ ] Add route registration in app.go
- [ ] Initialize publication service

### Phase 5: Testing
- [ ] Write unit tests for service
- [ ] Write integration tests for repository
- [ ] Test validation scenarios
- [ ] Verify all tests passing
- [ ] Verify no compilation errors

---

## 📊 Progress Tracking

### Session Log

**2025-11-20 - Session 1: Task Setup**
- Created feature branch `feature/MVP-PUB-003_publication_service`
- Created task specification document
- Defined todo list with 8 items
- Ready to begin implementation

---

## 🔗 References

- **Architecture Document**: `/documents/2-SoftwareDesignAndArchitecture/agency-publishing-tagging-architecture.md`
- **MVP Task List**: `/documents/3-SofwareDevelopment/mvp.md` (line 72)
- **Dependency (MVP-PUB-001)**: `/documents/3-SofwareDevelopment/coding_sessions/MVP-PUB-001_state_machine_data_models.md`
- **Dependency (MVP-PUB-002)**: `/documents/3-SofwareDevelopment/coding_sessions/MVP-PUB-002_tag_service.md`
- **AgencyPublication Model**: `/internal/agency/models/publication.go`
- **Database Migration**: `/internal/database/migrations/006_agency_publishing.go`

---

## 💡 Design Decisions

### 1. Validation-First Approach
- Validate before allowing publish
- Transition draft → validated → published
- Clear validation errors vs warnings

### 2. Immutable Publications
- Once published, cannot be modified
- Version-controlled snapshots
- Deployment manifests frozen at publish time

### 3. Semantic Versioning
- Require v1.0.0 format
- Track major.minor.patch versions
- Sort by version for history

### 4. Activation Stubs
- Activate/Deactivate implemented in MVP-PUB-004
- This task creates stubs that return success
- State transitions work, but no agents spawned yet

### 5. Manifest Generation
- Compute from specification at publish time
- Default resource limits (500m CPU, 512Mi memory)
- Token budgets from role definitions

---

## ⚠️ Known Limitations

1. **Activation Stub**: Activate/Deactivate don't spawn agents (MVP-PUB-004)
2. **No Version Validation**: Doesn't check if version already exists
3. **No Rollback**: Cannot undo publication (future enhancement)
4. **Single Active Publication**: No support for multiple concurrent versions

---

## 🔜 Future Enhancements

1. **Version Validation**: Check for duplicate versions
2. **Publication Rollback**: Revert to previous publication
3. **Blue-Green Deployment**: Support multiple active versions
4. **Approval Workflow**: Require approval before publish
5. **Publication Templates**: Predefined deployment configurations

---

## 📝 Notes

- Following template-first architecture (no HTML in handlers)
- Using functional programming principles (testable, pure functions)
- Keeping file sizes under limits (services <500 lines, repositories <250 lines)
- All API endpoints return JSON (no HTML rendering in this task)
- UI components will be added in MVP-PUB-005
- Agent spawning deferred to MVP-PUB-004
