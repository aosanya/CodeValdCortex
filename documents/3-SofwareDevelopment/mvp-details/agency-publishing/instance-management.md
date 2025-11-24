# MVP-PUB-007: Agency Instance Management System

**Task ID**: MVP-PUB-007  
**Priority**: P0 (Critical)  
**Effort**: High  
**Skills**: Go, ArangoDB, Templ, Frontend Dev  
**Dependencies**: MVP-PUB-006 (Agency-Specific Tag Storage) ✅

---

## Overview

Implement a comprehensive instance management system that enables running multiple independent instances of an agency from any tag snapshot.

## Documentation Structure

This task is split across topic-based files (refactored from original 884-line file):

### 📊 [Data Models & Database Schema](instance-data-models.md)
**182 lines** - `AgencyInstance` model, database collections, indexes

### ⚙️ [Service Layer & Repository](instance-services.md)
**296 lines** - Service interface, graceful shutdown, health calculation

### 🌐 [HTTP API Endpoints](instance-api.md)
**353 lines** - 9 REST endpoints, handlers, error handling

### 🎨 [UI Components](instance-ui.md)
**382 lines** - Tags list, instance dialog, dashboard (5 panels)

---

## Key Design Decisions

1. **Storage**: All instances in `agency_instances` collection (agency DB)
2. **Agents**: References to tag configs (not physical)
3. **State**: "Running" immediately on creation
4. **Health**: On-demand calculation
5. **Shutdown**: Graceful (30s timeout, rejects new jobs)
6. **Naming**: Unique per agency
7. **Delete**: Soft delete only
8. **Uptime**: Real-time (`current_time - started_at`)

---

## Implementation Checklist

- [ ] Data models & collections
- [ ] Service & repository layers
- [ ] 9 API endpoints
- [ ] UI (tags list, dialog, dashboard)
- [ ] Integration & tests

See detailed checklists in topic files above.

---

## Timeline

**Estimated**: 2-3 days (16-24 hours)  
**Complexity**: High
