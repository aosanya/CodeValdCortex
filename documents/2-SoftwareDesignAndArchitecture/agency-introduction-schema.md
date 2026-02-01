# Agency Introduction - Flexible Data-Driven Schema

## Overview

The Agency Introduction is a **flexible, data-driven structure** that allows agencies to define their context through customizable sections. While the Genesis Agency demonstrates a 6-section pattern (Problem Statement, Solution Approach, Scope, Context, Success Criteria, Stakeholders), agencies can define their own introduction structure based on their needs.

## Design Principles

1. **Flexibility First** - Sections are configurable, not hard-coded
2. **Data-Driven UI** - Frontend renders based on JSON schema, not fixed components
3. **Template Support** - Common patterns available as templates
4. **Extensibility** - Easy to add new section types without code changes
5. **Validation** - Schema-based validation ensures consistency

## JSON Schema

### Introduction Structure

```json
{
  "introduction": {
    "sections": [
      {
        "id": "intro_001",
        "title": "Section Title",
        "order": 1,
        "type": "text|list|nested|table",
        "required": true|false,
        "content": "...",
        "metadata": {
          "created_at": "2026-01-30T00:00:00Z",
          "updated_at": "2026-01-30T00:00:00Z",
          "created_by": "user_id",
          "version": "1.0.0"
        }
      }
    ],
    "template": "standard|genesis|custom",
    "version": "1.0.0"
  }
}
```

### Section Types

#### 1. Text Section
Simple markdown content

```json
{
  "id": "intro_001",
  "title": "Problem Statement",
  "order": 1,
  "type": "text",
  "required": true,
  "content": "Markdown formatted text content...",
  "validation": {
    "min_length": 100,
    "max_length": 5000
  }
}
```

#### 2. List Section
Structured list with items

```json
{
  "id": "intro_002",
  "title": "Success Criteria",
  "order": 2,
  "type": "list",
  "required": true,
  "content": [
    {
      "id": "success_001",
      "title": "Speed",
      "content": "Agency specification generated in < 2 hours"
    },
    {
      "id": "success_002",
      "title": "Quality",
      "content": "100% validation pass rate"
    }
  ],
  "validation": {
    "min_items": 1,
    "max_items": 20
  }
}
```

#### 3. Nested Section
Hierarchical content with subsections

```json
{
  "id": "intro_003",
  "title": "Scope",
  "order": 3,
  "type": "nested",
  "required": true,
  "content": {
    "in_scope": [
      {
        "id": "scope_in_001",
        "title": "Feature A",
        "content": "Description..."
      }
    ],
    "out_of_scope": [
      {
        "id": "scope_out_001",
        "title": "Feature B",
        "content": "Description..."
      }
    ]
  }
}
```

#### 4. Table Section
Tabular data with rows and columns

```json
{
  "id": "intro_004",
  "title": "Stakeholders",
  "order": 4,
  "type": "table",
  "required": false,
  "content": {
    "columns": ["Stakeholder", "Role", "Interest"],
    "rows": [
      {
        "id": "stakeholder_001",
        "data": ["Platform Admins", "Primary User", "Managing deployment"]
      }
    ]
  }
}
```

## Database Schema (ArangoDB)

### Collection: `agency_introductions`

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
      "content": "...",
      "metadata": {
        "created_at": "2026-01-30T00:00:00Z",
        "updated_at": "2026-01-30T00:00:00Z",
        "created_by": "user_123",
        "version": "1.0.0"
      }
    }
  ],
  "template": "genesis",
  "version": "1.0.0",
  "created_at": "2026-01-30T00:00:00Z",
  "updated_at": "2026-01-30T00:00:00Z",
  "created_by": "user_123"
}
```

## API Endpoints

### Get Introduction
```
GET /api/v1/agencies/:agency_id/introduction
```

**Response:**
```json
{
  "agency_id": "agency_123",
  "sections": [...],
  "template": "genesis",
  "version": "1.0.0"
}
```

### Update Section
```
PUT /api/v1/agencies/:agency_id/introduction/sections/:section_id
```

**Request Body:**
```json
{
  "title": "Problem Statement",
  "content": "Updated content...",
  "type": "text"
}
```

### Add Section
```
POST /api/v1/agencies/:agency_id/introduction/sections
```

**Request Body:**
```json
{
  "title": "New Section",
  "order": 5,
  "type": "text",
  "required": false,
  "content": "..."
}
```

### Delete Section
```
DELETE /api/v1/agencies/:agency_id/introduction/sections/:section_id
```

### Reorder Sections
```
PATCH /api/v1/agencies/:agency_id/introduction/sections/reorder
```

**Request Body:**
```json
{
  "section_ids": ["intro_001", "intro_003", "intro_002"]
}
```

## Templates

### Genesis Template (6 Sections)

```json
{
  "template_id": "genesis",
  "name": "Genesis Agency Template",
  "description": "Standard 6-section introduction for agency builder",
  "sections": [
    {
      "id": "intro_001",
      "title": "Problem Statement",
      "order": 1,
      "type": "text",
      "required": true,
      "placeholder": "What problem does this agency solve?",
      "help_text": "Describe the business problem, current challenges, and pain points"
    },
    {
      "id": "intro_002",
      "title": "Solution Approach",
      "order": 2,
      "type": "text",
      "required": true,
      "placeholder": "How does this agency solve the problem?",
      "help_text": "Describe the approach, key features, and technical architecture"
    },
    {
      "id": "intro_003",
      "title": "Scope",
      "order": 3,
      "type": "nested",
      "required": true,
      "structure": {
        "in_scope": [],
        "out_of_scope": []
      }
    },
    {
      "id": "intro_004",
      "title": "Context",
      "order": 4,
      "type": "text",
      "required": true,
      "placeholder": "What context is important to understand?",
      "help_text": "Provide background, related systems, technology stack, and design decisions"
    },
    {
      "id": "intro_005",
      "title": "Success Criteria",
      "order": 5,
      "type": "list",
      "required": true,
      "help_text": "Define measurable success criteria"
    },
    {
      "id": "intro_006",
      "title": "Stakeholders",
      "order": 6,
      "type": "table",
      "required": false,
      "columns": ["Stakeholder", "Role", "Interest"]
    }
  ]
}
```

### Minimal Template (3 Sections)

```json
{
  "template_id": "minimal",
  "name": "Minimal Introduction",
  "description": "Simplified 3-section introduction for quick setup",
  "sections": [
    {
      "id": "intro_001",
      "title": "Overview",
      "order": 1,
      "type": "text",
      "required": true
    },
    {
      "id": "intro_002",
      "title": "Objectives",
      "order": 2,
      "type": "list",
      "required": true
    },
    {
      "id": "intro_003",
      "title": "Scope",
      "order": 3,
      "type": "nested",
      "required": true,
      "structure": {
        "in_scope": [],
        "out_of_scope": []
      }
    }
  ]
}
```

### Custom Template

Agencies can define completely custom structures with any number and type of sections.

## UI Components (Data-Driven)

### Flutter Component Structure

```dart
// Flutter Example - Data-Driven Section Renderer
class IntroductionSection extends StatelessWidget {
  final SectionData section;
  
  @override
  Widget build(BuildContext context) {
    switch (section.type) {
      case 'text':
        return TextSectionWidget(section);
      case 'list':
        return ListSectionWidget(section);
      case 'nested':
        return NestedSectionWidget(section);
      case 'table':
        return TableSectionWidget(section);
      default:
        return CustomSectionWidget(section);
    }
  }
}

// Section widgets are rendered based on data, not hard-coded
class AgencyIntroduction extends StatelessWidget {
  final IntroductionData introduction;
  
  @override
  Widget build(BuildContext context) {
    return Column(
      children: introduction.sections
          .map((section) => IntroductionSection(section: section))
          .toList(),
    );
  }
}
```

### Editor Features

1. **Add Section Button** - Dropdown to select section type
2. **Drag-to-Reorder** - Drag handles for section reordering
3. **Inline Editing** - Click-to-edit section titles and content
4. **Template Selector** - Choose from predefined templates
5. **Section Actions** - Edit, Delete, Duplicate, Move Up/Down
6. **Live Preview** - Preview mode showing rendered introduction
7. **Version History** - Track changes to each section

## Validation Rules

### Section-Level Validation

```json
{
  "validation": {
    "required": true,
    "min_length": 100,
    "max_length": 5000,
    "pattern": "regex_pattern",
    "custom_validators": ["no_placeholder_text", "grammar_check"]
  }
}
```

### Introduction-Level Validation

```javascript
{
  "rules": [
    {
      "rule": "at_least_one_required_section",
      "message": "Introduction must have at least one required section"
    },
    {
      "rule": "unique_section_ids",
      "message": "Section IDs must be unique"
    },
    {
      "rule": "sequential_ordering",
      "message": "Section order must be sequential starting from 1"
    },
    {
      "rule": "no_duplicate_titles",
      "message": "Section titles must be unique"
    }
  ]
}
```

## AI-Powered Generation

### Generate Introduction from Keywords

```
POST /api/v1/agencies/:agency_id/introduction/generate
```

**Request:**
```json
{
  "keywords": ["water distribution", "Nairobi", "monitoring"],
  "template": "genesis",
  "context": {
    "domain": "infrastructure",
    "location": "Kenya",
    "stakeholders": ["utilities", "government"]
  }
}
```

**Response:**
```json
{
  "sections": [
    {
      "id": "intro_001",
      "title": "Problem Statement",
      "content": "AI-generated problem statement...",
      "confidence": 0.92
    }
  ]
}
```

### Refine Section

```
POST /api/v1/agencies/:agency_id/introduction/sections/:section_id/refine
```

**Request:**
```json
{
  "instructions": "Make it more concise and focus on measurable outcomes",
  "tone": "professional|casual|technical"
}
```

## Migration from Static Structure

For existing agencies with hardcoded introduction sections:

```javascript
// Migration Script
function migrateIntroduction(agency) {
  const legacyIntro = agency.introduction; // String or single field
  
  return {
    sections: [
      {
        id: "intro_001",
        title: "Overview",
        order: 1,
        type: "text",
        required: true,
        content: legacyIntro,
        metadata: {
          created_at: agency.created_at,
          updated_at: agency.updated_at,
          created_by: agency.created_by,
          version: "1.0.0",
          migrated_from: "legacy_introduction"
        }
      }
    ],
    template: "custom",
    version: "1.0.0"
  };
}
```

## Export Formats

### Markdown Export

```markdown
# Agency Introduction

## Problem Statement

Content...

## Solution Approach

Content...

## Scope

### In Scope
- Item 1
- Item 2

### Out of Scope
- Item 1
- Item 2
```

### PDF Export

Rendered with proper formatting, section numbering, and styling based on template.

### JSON Export

Raw JSON structure for programmatic access.

## Benefits of Flexible Structure

1. **Adaptability** - Different agencies need different introduction structures
2. **Evolution** - Easy to add new section types without code changes
3. **Consistency** - Templates ensure common patterns are followed
4. **Scalability** - No UI changes needed for new section types
5. **Maintainability** - Single data-driven component vs. multiple hard-coded components
6. **AI-Friendly** - Easy to generate and refine sections programmatically

## Implementation Tasks

See MVP task breakdown for phased implementation:
- MVP-FL-104: Introduction Section (Frontend - Flutter)
- MVP-025: AI Agency Designer - Introduction (Backend - Go)
- API endpoints for CRUD operations on introduction sections
- Template management system
- Section type registry and renderers

---

**Document Version**: 1.0.0  
**Last Updated**: January 30, 2026  
**Status**: Architecture Specification
