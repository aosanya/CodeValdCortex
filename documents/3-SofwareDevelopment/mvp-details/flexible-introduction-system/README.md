# Flexible Introduction System Domain

**Domain**: Agency Introduction Management  
**Status**: Design Phase → Implementation  
**Priority**: P0 (Critical - Foundation for Agency Designer)

---

## Overview

The Flexible Introduction System implements a **data-driven, schema-based approach** to agency introductions, replacing hard-coded UI components with dynamic section rendering based on JSON schemas. This enables agencies to define custom introduction structures while providing helpful templates for quick start.

### Core Philosophy

1. **Flexibility First** - Sections are configurable, not hard-coded
2. **Data-Driven UI** - Frontend renders based on JSON schema
3. **Template Support** - Common patterns available (Genesis, Minimal, Custom)
4. **AI-Powered** - Generate and refine sections through AI services
5. **Extensibility** - Easy to add new section types without code changes

---

## Architecture

### System Components

**Backend (Go)**:
- Database schema (`agency_introductions` collection in ArangoDB)
- Models for all section types (Text, List, Nested, Table)
- Repository layer with full CRUD operations
- REST API endpoints for introduction management
- Template system (Genesis, Minimal, Custom)
- AI generation integration

**Frontend (Flutter)**:
- Data models matching backend schema
- Data-driven section renderers
- Template selector UI
- Drag-to-reorder sections
- Inline editing with auto-save
- AI generation dialogs

### Section Types

The system supports 4 extensible section types:

1. **Text** - Markdown content for narrative sections
2. **List** - Structured items (e.g., Success Criteria)
3. **Nested** - Hierarchical content (e.g., In Scope / Out of Scope)
4. **Table** - Tabular data (e.g., Stakeholders)

### Templates

**Genesis Template** (6 sections):
- Problem Statement
- Solution Approach
- Scope (In/Out of Scope)
- Context
- Success Criteria
- Stakeholders

**Minimal Template** (3 sections):
- Overview
- Objectives
- Scope

**Custom**: Build from scratch with any sections

---

## Task Index

### Implementation Tasks

| Task ID | Title | Status | Topic File |
|---------|-------|--------|------------|
| MVP-INTRO-001 | Flexible Introduction System | 📋 Not Started | [implementation.md](implementation.md) |

### Related Documentation

- **Architecture Spec**: [agency-introduction-schema.md](../../../2-SoftwareDesignAndArchitecture/agency-introduction-schema.md)
- **Platform Introduction**: [platform-introduction.md](../../../2-SoftwareDesignAndArchitecture/platform-introduction.md)
- **Research Summary**: [RESEARCH_SESSION_SUMMARY.md](../RESEARCH_SESSION_SUMMARY.md)
- **Genesis Agency Reference**: [genesis-agency/import/introduction.json](../genesis-agency/import/introduction.json)

---

## Data Model

### Database Schema (ArangoDB)

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
      "type": "text|list|nested|table",
      "required": true,
      "content": "...",
      "validation": {...},
      "metadata": {...}
    }
  ],
  "template": "genesis|minimal|custom",
  "version": "1.0.0",
  "created_at": "2026-01-30T00:00:00Z",
  "updated_at": "2026-01-30T00:00:00Z",
  "created_by": "user_123"
}
```

---

## API Endpoints

### Introduction Management

```
GET    /api/v1/agencies/:agency_id/introduction
POST   /api/v1/agencies/:agency_id/introduction
PUT    /api/v1/agencies/:agency_id/introduction
```

### Section Operations

```
GET    /api/v1/agencies/:agency_id/introduction/sections/:section_id
POST   /api/v1/agencies/:agency_id/introduction/sections
PUT    /api/v1/agencies/:agency_id/introduction/sections/:section_id
DELETE /api/v1/agencies/:agency_id/introduction/sections/:section_id
PATCH  /api/v1/agencies/:agency_id/introduction/sections/reorder
```

### AI Generation

```
POST   /api/v1/agencies/:agency_id/introduction/generate
POST   /api/v1/agencies/:agency_id/introduction/sections/:section_id/refine
```

### Templates

```
GET    /api/v1/agencies/:agency_id/introduction/templates
POST   /api/v1/agencies/:agency_id/introduction/templates/:template_id/apply
```

---

## Benefits

### For Agencies
- ✅ Customize introduction structure to fit needs
- ✅ Use templates for quick start
- ✅ AI-assisted content generation
- ✅ Easy to refine and iterate

### For Platform
- ✅ No UI changes needed for new section types
- ✅ Single data-driven component vs. multiple hard-coded components
- ✅ Easy to extend with new templates
- ✅ Consistent data model across agencies

### For Developers
- ✅ Schema-based validation
- ✅ Type-safe models
- ✅ Well-documented API
- ✅ Clear separation of concerns

---

## Success Metrics

- ✅ Agencies can create introduction with any template
- ✅ Users can add/edit/delete sections dynamically
- ✅ AI generation produces valid section content
- ✅ < 2 seconds to load/save introduction
- ✅ Zero validation errors with proper input
- ✅ 100% backwards compatibility with migrated agencies

---

## Integration Points

### Existing Systems

- **Agency Designer** (MVP-FL-103) - Navigation shell
- **AI Builder Services** - IntroductionBuilder for generation
- **Agency Management** - CreateAgency, InitializeAgencyDatabase
- **Research Framework** - `.github/prompts/research.prompt.md`

### Future Integration

- **Goals System** - Similar flexible structure
- **Roles System** - Similar flexible structure
- **Workflows System** - Similar flexible structure

---

## Timeline

**Estimated: 3-4 weeks**

- **Week 1**: Backend (Database, Models, Repository, API)
- **Week 2**: Templates & AI Integration
- **Week 3**: Frontend (UI Components, Data-Driven Rendering)
- **Week 4**: Testing & Deployment

---

**Last Updated**: January 30, 2026  
**Status**: Ready for Implementation
