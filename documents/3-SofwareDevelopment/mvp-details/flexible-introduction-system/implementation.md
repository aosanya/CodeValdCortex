# MVP-INTRO-001: Flexible Introduction System - Implementation

## Task Overview

**Task ID**: MVP-INTRO-001  
**Title**: Implement Flexible, Data-Driven Agency Introduction System  
**Priority**: P0 (Critical - Foundation for Agency Designer)  
**Status**: Not Started  
**Estimated Effort**: High (2-3 weeks)  
**Dependencies**: 
- MVP-FL-103 ✅ (Agency Designer Navigation - completed)
- Database schema for agencies collection
- AI Builder services foundation

**Related Tasks**:
- MVP-FL-104: Frontend Introduction Section (Flutter)
- MVP-025: Backend AI Agency Designer - Introduction (Go)
- MVP-042: AI-Powered Agency Creator

---

## Objective

Implement a complete flexible, data-driven introduction system that allows agencies to define custom introduction structures using templates or fully custom sections. Replace hard-coded 6-section introduction with schema-based approach.

---

## Architecture Reference

**Primary Specification**: [agency-introduction-schema.md](../../2-SoftwareDesignAndArchitecture/agency-introduction-schema.md)

**Key Design Principles**:
1. **Flexibility First** - Sections configurable via JSON, not hard-coded in UI
2. **Data-Driven Rendering** - Frontend renders based on schema
3. **Template Support** - Genesis (6-section), Minimal (3-section), Custom
4. **AI-Powered** - Generate and refine sections through AI services
5. **Extensibility** - New section types without code changes

---

## Phase 1: Database Schema & Models (Backend - Go)

### 1.1 Database Schema

**Collection**: `agency_introductions`

```javascript
{
  "_key": "agency_{uuid}_intro",
  "_id": "agency_introductions/agency_{uuid}_intro",
  "agency_id": "agency_{uuid}",
  "sections": [
    {
      "id": "intro_001",
      "title": "Problem Statement",
      "order": 1,
      "type": "text",
      "required": true,
      "content": "Markdown content...",
      "validation": {
        "min_length": 100,
        "max_length": 5000
      },
      "metadata": {
        "created_at": "2026-01-30T00:00:00Z",
        "updated_at": "2026-01-30T00:00:00Z",
        "created_by": "user_123",
        "version": "1.0.0"
      }
    }
  ],
  "template": "genesis|minimal|custom",
  "version": "1.0.0",
  "created_at": "2026-01-30T00:00:00Z",
  "updated_at": "2026-01-30T00:00:00Z",
  "created_by": "user_123"
}
```

### 1.2 Go Models

**File**: `internal/agency/models/introduction.go`

```go
package models

import "time"

// IntroductionSection represents a single section in the agency introduction
type IntroductionSection struct {
    ID         string                 `json:"id"`
    Title      string                 `json:"title"`
    Order      int                    `json:"order"`
    Type       SectionType            `json:"type"`
    Required   bool                   `json:"required"`
    Content    interface{}            `json:"content"` // Can be string, array, or nested object
    Validation *SectionValidation     `json:"validation,omitempty"`
    Metadata   SectionMetadata        `json:"metadata"`
}

// SectionType defines the type of section
type SectionType string

const (
    SectionTypeText   SectionType = "text"
    SectionTypeList   SectionType = "list"
    SectionTypeNested SectionType = "nested"
    SectionTypeTable  SectionType = "table"
)

// SectionValidation defines validation rules for a section
type SectionValidation struct {
    MinLength *int    `json:"min_length,omitempty"`
    MaxLength *int    `json:"max_length,omitempty"`
    Pattern   *string `json:"pattern,omitempty"`
    MinItems  *int    `json:"min_items,omitempty"`
    MaxItems  *int    `json:"max_items,omitempty"`
}

// SectionMetadata tracks section version history
type SectionMetadata struct {
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    CreatedBy string    `json:"created_by"`
    Version   string    `json:"version"`
}

// AgencyIntroduction represents the complete introduction structure
type AgencyIntroduction struct {
    Key       string                `json:"_key" arangodb:"_key"`
    ID        string                `json:"_id,omitempty" arangodb:"_id,omitempty"`
    AgencyID  string                `json:"agency_id"`
    Sections  []IntroductionSection `json:"sections"`
    Template  string                `json:"template"` // "genesis", "minimal", "custom"
    Version   string                `json:"version"`
    CreatedAt time.Time             `json:"created_at"`
    UpdatedAt time.Time             `json:"updated_at"`
    CreatedBy string                `json:"created_by"`
}

// ListItem represents an item in a list-type section
type ListItem struct {
    ID      string `json:"id"`
    Title   string `json:"title"`
    Content string `json:"content"`
}

// NestedContent represents nested section structure (e.g., in/out of scope)
type NestedContent struct {
    InScope  []ListItem `json:"in_scope,omitempty"`
    OutScope []ListItem `json:"out_of_scope,omitempty"`
}

// TableContent represents tabular data
type TableContent struct {
    Columns []string              `json:"columns"`
    Rows    []map[string]string   `json:"rows"`
}
```

### 1.3 Database Repository

**File**: `internal/agency/repository/introduction.go`

```go
package repository

import (
    "context"
    "fmt"
    
    "github.com/arangodb/go-driver"
    "your-module/internal/agency/models"
)

type IntroductionRepository struct {
    db         driver.Database
    collection driver.Collection
}

func NewIntroductionRepository(db driver.Database) (*IntroductionRepository, error) {
    collection, err := db.Collection(context.Background(), "agency_introductions")
    if err != nil {
        return nil, err
    }
    
    return &IntroductionRepository{
        db:         db,
        collection: collection,
    }, nil
}

// GetByAgencyID retrieves introduction for an agency
func (r *IntroductionRepository) GetByAgencyID(ctx context.Context, agencyID string) (*models.AgencyIntroduction, error) {
    query := `FOR intro IN agency_introductions
              FILTER intro.agency_id == @agency_id
              RETURN intro`
    
    cursor, err := r.db.Query(ctx, query, map[string]interface{}{
        "agency_id": agencyID,
    })
    if err != nil {
        return nil, err
    }
    defer cursor.Close()
    
    var intro models.AgencyIntroduction
    _, err = cursor.ReadDocument(ctx, &intro)
    if err != nil {
        return nil, err
    }
    
    return &intro, nil
}

// Create creates a new introduction
func (r *IntroductionRepository) Create(ctx context.Context, intro *models.AgencyIntroduction) error {
    _, err := r.collection.CreateDocument(ctx, intro)
    return err
}

// Update updates an existing introduction
func (r *IntroductionRepository) Update(ctx context.Context, intro *models.AgencyIntroduction) error {
    _, err := r.collection.UpdateDocument(ctx, intro.Key, intro)
    return err
}

// UpdateSection updates a specific section
func (r *IntroductionRepository) UpdateSection(ctx context.Context, agencyID, sectionID string, section *models.IntroductionSection) error {
    query := `FOR intro IN agency_introductions
              FILTER intro.agency_id == @agency_id
              LET sections = (
                  FOR s IN intro.sections
                  RETURN s.id == @section_id ? @new_section : s
              )
              UPDATE intro WITH { sections: sections, updated_at: DATE_ISO8601(DATE_NOW()) } IN agency_introductions`
    
    _, err := r.db.Query(ctx, query, map[string]interface{}{
        "agency_id":   agencyID,
        "section_id":  sectionID,
        "new_section": section,
    })
    return err
}

// AddSection adds a new section
func (r *IntroductionRepository) AddSection(ctx context.Context, agencyID string, section *models.IntroductionSection) error {
    query := `FOR intro IN agency_introductions
              FILTER intro.agency_id == @agency_id
              UPDATE intro WITH { 
                  sections: APPEND(intro.sections, [@new_section]),
                  updated_at: DATE_ISO8601(DATE_NOW())
              } IN agency_introductions`
    
    _, err := r.db.Query(ctx, query, map[string]interface{}{
        "agency_id":   agencyID,
        "new_section": section,
    })
    return err
}

// DeleteSection removes a section
func (r *IntroductionRepository) DeleteSection(ctx context.Context, agencyID, sectionID string) error {
    query := `FOR intro IN agency_introductions
              FILTER intro.agency_id == @agency_id
              LET sections = (
                  FOR s IN intro.sections
                  FILTER s.id != @section_id
                  RETURN s
              )
              UPDATE intro WITH { sections: sections, updated_at: DATE_ISO8601(DATE_NOW()) } IN agency_introductions`
    
    _, err := r.db.Query(ctx, query, map[string]interface{}{
        "agency_id":  agencyID,
        "section_id": sectionID,
    })
    return err
}

// ReorderSections updates section order
func (r *IntroductionRepository) ReorderSections(ctx context.Context, agencyID string, sectionIDs []string) error {
    // Implementation for reordering
    return nil
}
```

---

## Phase 2: API Endpoints (Backend - Go)

### 2.1 API Routes

**File**: `internal/api/routes/introduction.go`

```go
package routes

import (
    "github.com/gin-gonic/gin"
    "your-module/internal/api/handlers"
)

func RegisterIntroductionRoutes(r *gin.RouterGroup, handler *handlers.IntroductionHandler) {
    intro := r.Group("/agencies/:agency_id/introduction")
    {
        intro.GET("", handler.GetIntroduction)
        intro.POST("", handler.CreateIntroduction)
        intro.PUT("", handler.UpdateIntroduction)
        
        // Section operations
        intro.GET("/sections/:section_id", handler.GetSection)
        intro.POST("/sections", handler.AddSection)
        intro.PUT("/sections/:section_id", handler.UpdateSection)
        intro.DELETE("/sections/:section_id", handler.DeleteSection)
        intro.PATCH("/sections/reorder", handler.ReorderSections)
        
        // AI generation
        intro.POST("/generate", handler.GenerateIntroduction)
        intro.POST("/sections/:section_id/refine", handler.RefineSection)
        
        // Templates
        intro.GET("/templates", handler.ListTemplates)
        intro.POST("/templates/:template_id/apply", handler.ApplyTemplate)
    }
}
```

### 2.2 API Handlers

**File**: `internal/api/handlers/introduction.go`

```go
package handlers

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "your-module/internal/agency/models"
    "your-module/internal/agency/service"
)

type IntroductionHandler struct {
    service *service.IntroductionService
}

func NewIntroductionHandler(service *service.IntroductionService) *IntroductionHandler {
    return &IntroductionHandler{service: service}
}

// GetIntroduction retrieves introduction for an agency
// GET /api/v1/agencies/:agency_id/introduction
func (h *IntroductionHandler) GetIntroduction(c *gin.Context) {
    agencyID := c.Param("agency_id")
    
    intro, err := h.service.GetByAgencyID(c.Request.Context(), agencyID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Introduction not found"})
        return
    }
    
    c.JSON(http.StatusOK, intro)
}

// UpdateSection updates a specific section
// PUT /api/v1/agencies/:agency_id/introduction/sections/:section_id
func (h *IntroductionHandler) UpdateSection(c *gin.Context) {
    agencyID := c.Param("agency_id")
    sectionID := c.Param("section_id")
    
    var section models.IntroductionSection
    if err := c.ShouldBindJSON(&section); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    section.ID = sectionID
    if err := h.service.UpdateSection(c.Request.Context(), agencyID, &section); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, section)
}

// AddSection adds a new section
// POST /api/v1/agencies/:agency_id/introduction/sections
func (h *IntroductionHandler) AddSection(c *gin.Context) {
    agencyID := c.Param("agency_id")
    
    var section models.IntroductionSection
    if err := c.ShouldBindJSON(&section); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if err := h.service.AddSection(c.Request.Context(), agencyID, &section); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, section)
}

// GenerateIntroduction generates introduction using AI
// POST /api/v1/agencies/:agency_id/introduction/generate
func (h *IntroductionHandler) GenerateIntroduction(c *gin.Context) {
    agencyID := c.Param("agency_id")
    
    var req struct {
        Keywords []string               `json:"keywords"`
        Template string                 `json:"template"`
        Context  map[string]interface{} `json:"context"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    intro, err := h.service.GenerateFromKeywords(c.Request.Context(), agencyID, req.Keywords, req.Template, req.Context)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, intro)
}

// Additional handlers...
```

---

## Phase 3: Template System (Backend - Go)

### 3.1 Template Definitions

**File**: `internal/agency/templates/introduction.go`

```go
package templates

import "your-module/internal/agency/models"

// IntroductionTemplate defines a template structure
type IntroductionTemplate struct {
    ID          string                          `json:"id"`
    Name        string                          `json:"name"`
    Description string                          `json:"description"`
    Sections    []IntroductionSectionTemplate   `json:"sections"`
}

// IntroductionSectionTemplate defines a section template
type IntroductionSectionTemplate struct {
    ID          string                 `json:"id"`
    Title       string                 `json:"title"`
    Order       int                    `json:"order"`
    Type        models.SectionType     `json:"type"`
    Required    bool                   `json:"required"`
    Placeholder string                 `json:"placeholder,omitempty"`
    HelpText    string                 `json:"help_text,omitempty"`
    Structure   interface{}            `json:"structure,omitempty"`
}

// GenesisTemplate returns the 6-section Genesis Agency template
func GenesisTemplate() *IntroductionTemplate {
    return &IntroductionTemplate{
        ID:          "genesis",
        Name:        "Genesis Agency Template",
        Description: "Standard 6-section introduction for agency builder",
        Sections: []IntroductionSectionTemplate{
            {
                ID:          "intro_001",
                Title:       "Problem Statement",
                Order:       1,
                Type:        models.SectionTypeText,
                Required:    true,
                Placeholder: "What problem does this agency solve?",
                HelpText:    "Describe the business problem, current challenges, and pain points",
            },
            {
                ID:          "intro_002",
                Title:       "Solution Approach",
                Order:       2,
                Type:        models.SectionTypeText,
                Required:    true,
                Placeholder: "How does this agency solve the problem?",
                HelpText:    "Describe the approach, key features, and technical architecture",
            },
            {
                ID:       "intro_003",
                Title:    "Scope",
                Order:    3,
                Type:     models.SectionTypeNested,
                Required: true,
                Structure: models.NestedContent{
                    InScope:  []models.ListItem{},
                    OutScope: []models.ListItem{},
                },
            },
            {
                ID:          "intro_004",
                Title:       "Context",
                Order:       4,
                Type:        models.SectionTypeText,
                Required:    true,
                Placeholder: "What context is important to understand?",
                HelpText:    "Provide background, related systems, technology stack, and design decisions",
            },
            {
                ID:       "intro_005",
                Title:    "Success Criteria",
                Order:    5,
                Type:     models.SectionTypeList,
                Required: true,
                HelpText: "Define measurable success criteria",
            },
            {
                ID:       "intro_006",
                Title:    "Stakeholders",
                Order:    6,
                Type:     models.SectionTypeTable,
                Required: false,
                Structure: models.TableContent{
                    Columns: []string{"Stakeholder", "Role", "Interest"},
                    Rows:    []map[string]string{},
                },
            },
        },
    }
}

// MinimalTemplate returns a simplified 3-section template
func MinimalTemplate() *IntroductionTemplate {
    return &IntroductionTemplate{
        ID:          "minimal",
        Name:        "Minimal Introduction",
        Description: "Simplified 3-section introduction for quick setup",
        Sections: []IntroductionSectionTemplate{
            {
                ID:       "intro_001",
                Title:    "Overview",
                Order:    1,
                Type:     models.SectionTypeText,
                Required: true,
            },
            {
                ID:       "intro_002",
                Title:    "Objectives",
                Order:    2,
                Type:     models.SectionTypeList,
                Required: true,
            },
            {
                ID:       "intro_003",
                Title:    "Scope",
                Order:    3,
                Type:     models.SectionTypeNested,
                Required: true,
            },
        },
    }
}

// GetAllTemplates returns all available templates
func GetAllTemplates() []*IntroductionTemplate {
    return []*IntroductionTemplate{
        GenesisTemplate(),
        MinimalTemplate(),
    }
}
```

---

## Phase 4: AI Integration (Backend - Go)

### 4.1 AI Generation Service

**File**: `internal/builder/ai/introduction_builder.go`

```go
package ai

import (
    "context"
    "fmt"
    
    "your-module/internal/agency/models"
)

type IntroductionBuilder struct {
    llmClient LLMClient
}

func NewIntroductionBuilder(llmClient LLMClient) *IntroductionBuilder {
    return &IntroductionBuilder{llmClient: llmClient}
}

// GenerateFromKeywords generates introduction sections from keywords
func (b *IntroductionBuilder) GenerateFromKeywords(ctx context.Context, keywords []string, template string, context map[string]interface{}) (*models.AgencyIntroduction, error) {
    prompt := fmt.Sprintf(`Generate an agency introduction based on the following keywords: %v
    
Template: %s
Additional Context: %v

Generate each section according to the template structure.`, keywords, template, context)
    
    // Call LLM to generate sections
    response, err := b.llmClient.Complete(ctx, prompt)
    if err != nil {
        return nil, err
    }
    
    // Parse response into AgencyIntroduction structure
    intro := parseIntroductionResponse(response, template)
    
    return intro, nil
}

// RefineSection refines an existing section based on instructions
func (b *IntroductionBuilder) RefineSection(ctx context.Context, section *models.IntroductionSection, instructions string) (*models.IntroductionSection, error) {
    prompt := fmt.Sprintf(`Refine the following section:

Title: %s
Current Content: %v

Refinement Instructions: %s

Return the refined content maintaining the same structure.`, section.Title, section.Content, instructions)
    
    response, err := b.llmClient.Complete(ctx, prompt)
    if err != nil {
        return nil, err
    }
    
    // Parse refined content
    section.Content = response
    
    return section, nil
}
```

---

## Phase 5: Frontend Implementation (Flutter)

### 5.1 Data Models

**File**: `lib/features/agency_designer/models/introduction_model.dart`

```dart
class IntroductionSection {
  final String id;
  final String title;
  final int order;
  final SectionType type;
  final bool required;
  final dynamic content;
  final SectionValidation? validation;
  final SectionMetadata metadata;

  IntroductionSection({
    required this.id,
    required this.title,
    required this.order,
    required this.type,
    required this.required,
    required this.content,
    this.validation,
    required this.metadata,
  });

  factory IntroductionSection.fromJson(Map<String, dynamic> json) {
    return IntroductionSection(
      id: json['id'],
      title: json['title'],
      order: json['order'],
      type: SectionType.fromString(json['type']),
      required: json['required'] ?? false,
      content: json['content'],
      validation: json['validation'] != null
          ? SectionValidation.fromJson(json['validation'])
          : null,
      metadata: SectionMetadata.fromJson(json['metadata']),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'title': title,
      'order': order,
      'type': type.toString(),
      'required': required,
      'content': content,
      'validation': validation?.toJson(),
      'metadata': metadata.toJson(),
    };
  }
}

enum SectionType {
  text,
  list,
  nested,
  table;

  static SectionType fromString(String value) {
    return SectionType.values.firstWhere(
      (e) => e.toString().split('.').last == value,
    );
  }
}

class AgencyIntroduction {
  final String agencyId;
  final List<IntroductionSection> sections;
  final String template;
  final String version;

  AgencyIntroduction({
    required this.agencyId,
    required this.sections,
    required this.template,
    required this.version,
  });

  factory AgencyIntroduction.fromJson(Map<String, dynamic> json) {
    return AgencyIntroduction(
      agencyId: json['agency_id'],
      sections: (json['sections'] as List)
          .map((s) => IntroductionSection.fromJson(s))
          .toList(),
      template: json['template'],
      version: json['version'],
    );
  }
}
```

### 5.2 Data-Driven Section Renderer

**File**: `lib/features/agency_designer/widgets/introduction/section_renderer.dart`

```dart
import 'package:flutter/material.dart';

class SectionRenderer extends StatelessWidget {
  final IntroductionSection section;
  final Function(IntroductionSection)? onUpdate;
  final Function()? onDelete;

  const SectionRenderer({
    Key? key,
    required this.section,
    this.onUpdate,
    this.onDelete,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.symmetric(vertical: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildHeader(context),
          _buildContent(context),
        ],
      ),
    );
  }

  Widget _buildHeader(BuildContext context) {
    return ListTile(
      title: Text(
        section.title,
        style: Theme.of(context).textTheme.titleLarge,
      ),
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          IconButton(
            icon: const Icon(Icons.edit),
            onPressed: () => _handleEdit(context),
          ),
          if (!section.required)
            IconButton(
              icon: const Icon(Icons.delete),
              onPressed: onDelete,
            ),
        ],
      ),
    );
  }

  Widget _buildContent(BuildContext context) {
    switch (section.type) {
      case SectionType.text:
        return TextSectionWidget(
          content: section.content as String,
          onUpdate: (newContent) => _updateContent(newContent),
        );
      case SectionType.list:
        return ListSectionWidget(
          items: section.content as List,
          onUpdate: (newItems) => _updateContent(newItems),
        );
      case SectionType.nested:
        return NestedSectionWidget(
          content: section.content as Map<String, dynamic>,
          onUpdate: (newContent) => _updateContent(newContent),
        );
      case SectionType.table:
        return TableSectionWidget(
          content: section.content as Map<String, dynamic>,
          onUpdate: (newContent) => _updateContent(newContent),
        );
    }
  }

  void _updateContent(dynamic newContent) {
    if (onUpdate != null) {
      final updatedSection = IntroductionSection(
        id: section.id,
        title: section.title,
        order: section.order,
        type: section.type,
        required: section.required,
        content: newContent,
        validation: section.validation,
        metadata: section.metadata,
      );
      onUpdate!(updatedSection);
    }
  }

  void _handleEdit(BuildContext context) {
    // Show edit dialog
  }
}
```

### 5.3 Template Selector

**File**: `lib/features/agency_designer/widgets/introduction/template_selector.dart`

```dart
class TemplateSelector extends StatelessWidget {
  final Function(String templateId) onTemplateSelected;

  const TemplateSelector({
    Key? key,
    required this.onTemplateSelected,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Text('Select Introduction Template'),
        ListTile(
          title: Text('Genesis Template'),
          subtitle: Text('6-section comprehensive introduction'),
          onTap: () => onTemplateSelected('genesis'),
        ),
        ListTile(
          title: Text('Minimal Template'),
          subtitle: Text('3-section quick setup'),
          onTap: () => onTemplateSelected('minimal'),
        ),
        ListTile(
          title: Text('Custom'),
          subtitle: Text('Build from scratch'),
          onTap: () => onTemplateSelected('custom'),
        ),
      ],
    );
  }
}
```

---

## Acceptance Criteria

### Backend

- [ ] Database schema created for `agency_introductions` collection
- [ ] Go models defined for all section types (Text, List, Nested, Table)
- [ ] Repository layer with full CRUD operations
- [ ] API endpoints implemented and tested
  - [ ] GET /agencies/:id/introduction
  - [ ] POST /agencies/:id/introduction/sections
  - [ ] PUT /agencies/:id/introduction/sections/:section_id
  - [ ] DELETE /agencies/:id/introduction/sections/:section_id
  - [ ] PATCH /agencies/:id/introduction/sections/reorder
  - [ ] POST /agencies/:id/introduction/generate (AI)
- [ ] Template system with Genesis and Minimal templates
- [ ] AI generation integration for sections
- [ ] Validation logic for section content
- [ ] Unit tests for all services
- [ ] Integration tests for API endpoints

### Frontend

- [ ] Flutter models match backend schema
- [ ] Data-driven section renderer implemented
- [ ] Support for all 4 section types (Text, List, Nested, Table)
- [ ] Template selector UI
- [ ] Drag-to-reorder sections
- [ ] Add/edit/delete section operations
- [ ] Inline editing with auto-save
- [ ] Live preview mode
- [ ] AI generation dialog
- [ ] Section refinement dialog
- [ ] Validation feedback
- [ ] Responsive layout (mobile/tablet/desktop)
- [ ] Unit tests for widgets
- [ ] Integration tests for complete flow

### Documentation

- [ ] API documentation (Swagger/OpenAPI)
- [ ] Architecture decision records
- [ ] User guide for template system
- [ ] Developer guide for adding new section types
- [ ] Migration guide from old hard-coded structure

---

## Testing Strategy

### Unit Tests

**Backend**:
- Section validation logic
- Template generation
- AI integration (mocked)
- Repository operations

**Frontend**:
- Widget rendering for each section type
- State management
- Form validation

### Integration Tests

- Complete flow: Create agency → Apply template → Edit sections → Save
- API endpoint tests with real database
- AI generation end-to-end
- Template application

### E2E Tests

- User creates agency with Genesis template
- User customizes sections
- User generates content with AI
- User publishes agency

---

## Migration Strategy

### Existing Agencies

```sql
FOR agency IN agencies
  LET legacy_intro = agency.introduction
  INSERT {
    _key: CONCAT("agency_", agency._key, "_intro"),
    agency_id: agency._id,
    sections: [
      {
        id: "intro_001",
        title: "Overview",
        order: 1,
        type: "text",
        required: true,
        content: legacy_intro,
        metadata: {
          created_at: agency.created_at,
          updated_at: agency.updated_at,
          created_by: agency.created_by,
          version: "1.0.0"
        }
      }
    ],
    template: "custom",
    version: "1.0.0",
    created_at: agency.created_at,
    updated_at: agency.updated_at,
    created_by: agency.created_by
  } INTO agency_introductions
```

---

## Deployment Plan

### Phase 1: Backend (Week 1)
- Day 1-2: Database schema and models
- Day 3-4: Repository and service layer
- Day 5: API endpoints and handlers

### Phase 2: Templates & AI (Week 2)
- Day 1-2: Template system implementation
- Day 3-5: AI generation integration

### Phase 3: Frontend (Week 3)
- Day 1-2: Flutter models and API client
- Day 3-5: UI components and data-driven rendering

### Phase 4: Testing & Polish (Week 4)
- Day 1-3: Testing (unit, integration, E2E)
- Day 4-5: Bug fixes, documentation, deployment

---

## Success Metrics

- ✅ Agencies can create introduction with any template
- ✅ Users can add/edit/delete sections dynamically
- ✅ AI generation produces valid section content
- ✅ < 2 seconds to load/save introduction
- ✅ Zero validation errors with proper input
- ✅ 100% backwards compatibility with migrated agencies

---

## Related Documentation

- [Agency Introduction Schema](../../2-SoftwareDesignAndArchitecture/agency-introduction-schema.md)
- [Platform Introduction](../../2-SoftwareDesignAndArchitecture/platform-introduction.md)
- [Research Session Summary](./RESEARCH_SESSION_SUMMARY.md)
- Genesis Agency Reference: `mvp-details/genesis-agency/import/introduction.json`

---

**Created**: January 30, 2026  
**Status**: Ready for Implementation  
**Estimated Delivery**: 3-4 weeks
