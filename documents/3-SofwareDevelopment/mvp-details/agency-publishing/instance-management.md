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

### 🎨 [UI Templates](instance-ui-templates.md)
**658 lines** - Complete Templ template specifications: hybrid list view (by-tag + flat table tabs), 5-panel dashboard, start instance dialog

### 💻 [JavaScript & Routing](instance-ui-javascript.md)
**335 lines** - Tab switching, client-side filtering, auto-refresh control, dialog management, instance control functions, UI routes

### 🔬 [Research Session](instance-research-session.md)
**Full architectural Q&A** - 11 questions answered covering navigation structure, filtering approach, data loading strategy, polling mechanism, panel design

---

## Key Design Decisions (Research-Driven)

### Instance Management Architecture
1. **Storage**: All instances in `agency_instances` collection (agency DB)
2. **Agent References**: Lazy initialization - agents spawn on-demand when jobs arrive
3. **Optimistic Start**: Instances immediately enter "running" state without validation
4. **Instance Isolation**: All collections use `instance_id` field for filtering/querying

### UI/UX Architecture
5. **Navigation**: Dual routes - `/instances` (top-level list) + `/agencies/:id/instances/:id` (dashboard)
6. **List View**: Hybrid approach - Tab 1 (by-tag grouping), Tab 2 (flat table with filters)
7. **Filtering**: Standard approach - state dropdown, tag dropdown, search box, sort dropdown
8. **Data Loading**: Full server render on page load (future: pagination for 500+ instances)
9. **Dashboard Polling**: Opt-in auto-refresh with staggered intervals per panel
   - Overview: 30s, Agents: 30s, Workflows: 20s, Activity: 60s
10. **Panel Independence**: Each panel is separate component with own HTMX endpoint

### Operational Patterns
11. **Health**: On-demand calculation (not stored)
12. **Shutdown**: Graceful (30s timeout, rejects new jobs)
13. **Naming**: Unique per agency
14. **Delete**: Soft delete only (preserves audit trail)
15. **Uptime**: Real-time calculation (`current_time - started_at`)

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
