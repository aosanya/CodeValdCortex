# React Migration Domain

**Domain**: Frontend Architecture Migration  
**Status**: Planning Phase  
**Priority**: P1 (Strategic Foundation)  
**Estimated Timeline**: 8-10 weeks

---

## Domain Overview

This domain covers the complete migration from the current Templ+HTMX+Alpine.js frontend to a React-based Single Page Application (SPA) called **CodeValdFortex**. The migration addresses critical architectural concerns around separation of concerns, scalability, and maintainability.

### Problem Statement

**Current Issues:**
- **Backend Bloat**: Presentation layer logic mixed with business logic in Go handlers
- **Tight Coupling**: Templ templates tightly coupled to Go backend
- **Limited Tooling**: Alpine.js/HTMX ecosystem lacks robust React tooling
- **Scalability Concerns**: Difficult to maintain as complexity grows

**Solution:**
- **Thin Go API Backend**: Pure REST API exposing business logic (CodeValdCortex)
- **React SPA Frontend**: All UI/presentation logic in CodeValdFortex
- **Clear Separation**: Backend = data/orchestration, Frontend = user experience
- **Incremental Migration**: Domain-by-domain rollout to minimize risk

---

## Architecture Overview

### System Architecture

```
Development Container:
/workspaces/
├── CodeValdCortex/          (Go Backend - REST API)
│   ├── internal/api/        (NEW: REST API handlers)
│   └── internal/{domain}/   (Business logic)
│
└── CodeValdFortex/          (React Frontend - SPA)
    ├── src/features/        (Redux slices by domain)
    ├── src/components/      (React components)
    └── src/api/             (API client)

Runtime:
┌──────────────┐     HTTP     ┌──────────────┐
│CodeValdFortex│ ◄───────────►│CodeValdCortex│
│  (React SPA) │  REST API    │  (Go Backend)│
│  Port: 5173  │              │  Port: 8080  │
└──────────────┘              └──────────────┘
                                     │
                                     ▼
                              ┌──────────────┐
                              │   ArangoDB   │
                              └──────────────┘
```

---

## Migration Phases

### Phase 1: Foundation & Work Items PoC (Weeks 1-5)
**Goal**: Prove React + Go API architecture works  
**Scope**: Work Items domain (list, create, edit, delete)

### Phase 2: Agent Cards (Weeks 6-8)
**Goal**: Migrate agent management UI  
**Scope**: Agent list, detail cards, actions, real-time status

### Phase 3: Agency Dashboard (Weeks 9-12)
**Goal**: Migrate agency overview  
**Scope**: Statistics, metrics, quick actions, charts

### Phase 4+: Advanced Domains
- Workflow Designer (drag-drop canvas)
- AI Policy Wizard
- Authentication & User Management

---

## Technology Stack

| Category | Technology | Purpose |
|----------|-----------|---------|
| **Frontend** | React 18.2+ | UI library |
| **Build Tool** | Vite 5.x | Fast build tool |
| **Language** | TypeScript 5.x | Type safety |
| **State** | Redux Toolkit 2.x | Global state |
| **CSS** | Bulma 1.0+ | Styling framework |
| **HTTP** | Axios 1.6+ | API requests |
| **Routing** | React Router 6.x | Client-side routing |
| **Backend** | Go 1.21+ | REST API |
| **API Framework** | Gin 1.9+ | HTTP routing |

---

## Task Index

### Phase 1: Foundation & Work Items PoC

**Infrastructure Setup:**
- [MVP-RM-001: Project Setup](./MVP-RM-001_project_setup.md) - Initialize CodeValdFortex repository
- [MVP-RM-002: Backend API Infrastructure](./MVP-RM-002_backend_api.md) - Create internal/api package structure

**Work Items Domain:**
- [MVP-RM-003: Work Items REST API](./MVP-RM-003_work_items_api.md) - Implement Work Items API endpoints
- [MVP-RM-004: Work Items Redux Store](./MVP-RM-004_work_items_redux.md) - Redux slice with async thunks
- [MVP-RM-005: Work Items UI Components](./MVP-RM-005_work_items_ui.md) - React components (list, card, form)

**Testing & Deployment:**
- [MVP-RM-006: Testing Suite](./MVP-RM-006_testing.md) - Unit, integration, E2E tests
- [MVP-RM-007: Deployment Pipeline](./MVP-RM-007_deployment.md) - CI/CD, staging, production

### Phase 2: Agent Cards (Future)
- MVP-RM-008: Agent REST API
- MVP-RM-009: Agent UI Components
- MVP-RM-010: Real-time Status Updates (WebSockets)

### Phase 3: Agency Dashboard (Future)
- MVP-RM-011: Dashboard API
- MVP-RM-012: Dashboard UI
- MVP-RM-013: Charts & Metrics Integration

---

## Success Metrics

### Technical Metrics
- API response time <100ms (95th percentile)
- Frontend bundle size <500KB (gzipped)
- Lighthouse performance score >90
- Test coverage >80%
- Zero critical security vulnerabilities

### User Experience Metrics
- Page load time <2 seconds
- Time to interactive <3 seconds
- Mobile responsive (works on tablets/phones)
- Accessible (WCAG 2.1 AA compliance)

### Business Metrics
- Feature parity with Templ version
- Zero data loss during migration
- <1% error rate in production
- Positive user feedback
- Team velocity maintained or improved

---

## Risk Mitigation

**High Risks:**
1. **Performance Degradation** → Code splitting, performance budgets, monitoring
2. **Data Consistency** → Thorough testing, transaction handling, rollback procedures
3. **Auth/Session Handling** → CORS testing, graceful error handling

**Medium Risks:**
1. **Developer Learning Curve** → Training, pair programming, documentation
2. **Deployment Complexity** → Automation, Docker, CI/CD pipeline

---

## Development Workflow

### Running Both Systems
```bash
# Terminal 1 - Go Backend
cd /workspaces/CodeValdCortex
make run  # http://localhost:8080

# Terminal 2 - React Frontend
cd /workspaces/CodeValdFortex
npm run dev  # http://localhost:5173
```

### Branch Management
```bash
# Create feature branch in BOTH repos
cd /workspaces/CodeValdCortex
git checkout dev
git checkout -b feature/MVP-RM-XXX_description

cd /workspaces/CodeValdFortex
git checkout dev
git checkout -b feature/MVP-RM-XXX_description
```

---

## References

- **Main Plan**: `/documents/2-SoftwareDesignAndArchitecture/react-migration-plan.md`
- **Backend Architecture**: `/documents/2-SoftwareDesignAndArchitecture/backend-architecture.md`
- **Current Frontend**: `/documents/2-SoftwareDesignAndArchitecture/frontend-architecture-updated.md`

---

**Last Updated**: January 26, 2026  
**Next Review**: End of Phase 1 (Week 5)
