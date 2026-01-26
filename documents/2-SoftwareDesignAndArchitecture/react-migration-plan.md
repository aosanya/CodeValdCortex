# React Migration Plan - CodeValdFortex Frontend

**Date**: January 26, 2026  
**Status**: Planning Phase  
**Author**: Architecture Team

---

## Executive Summary

This document outlines the complete migration strategy from the current Templ+HTMX+Alpine.js frontend to a new React-based SPA called **CodeValdFortex**. The migration addresses critical architectural concerns around separation of concerns and scalability.

### Problem Statement
- **Current Issue**: Templ template rendering logic is entangled with business logic in Go handlers
- **Go Backend Bloat**: Presentation layer mixed with domain logic, making the codebase difficult to maintain
- **Limited Tooling**: Alpine.js/HTMX ecosystem lacks the robust tooling available in React

### Solution Overview
- **Thin Go API Backend**: Pure REST API exposing business logic only (CodeValdCortex)
- **React SPA Frontend**: All UI/presentation logic in CodeValdFortex
- **Clear Separation**: Backend = data/orchestration, Frontend = user experience
- **Incremental Migration**: Domain-by-domain rollout to minimize risk

---

## Architecture Overview

### System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Development Container                    │
│  /workspaces/                                               │
│  ├── CodeValdCortex/          (Go Backend - REST API)      │
│  │   ├── cmd/main.go                                       │
│  │   ├── internal/                                         │
│  │   │   ├── api/             (NEW: REST API handlers)    │
│  │   │   ├── agency/          (Business logic)            │
│  │   │   ├── agent/           (Business logic)            │
│  │   │   └── workflow/        (Business logic)            │
│  │   └── config.yaml                                       │
│  │                                                          │
│  └── CodeValdFortex/          (React Frontend - SPA)       │
│      ├── src/                                              │
│      │   ├── features/        (Redux slices by domain)    │
│      │   ├── components/      (React components)          │
│      │   ├── api/             (API client)                │
│      │   └── store/           (Redux store)               │
│      ├── package.json                                      │
│      └── vite.config.ts                                    │
└─────────────────────────────────────────────────────────────┘

Runtime Architecture:
┌──────────────┐          HTTP          ┌──────────────┐
│ CodeValdFortex│ ◄─────────────────────►│ CodeValdCortex│
│  (React SPA) │      REST API          │  (Go Backend) │
│  Port: 5173  │  /api/v1/agencies      │  Port: 8080   │
│              │  /api/v1/work-items    │               │
└──────────────┘  /api/v1/workflows     └──────────────┘
                                                │
                                                ▼
                                         ┌──────────────┐
                                         │   ArangoDB   │
                                         │  Port: 8529  │
                                         └──────────────┘
```

---

## Phase 1: Work Items Domain (Proof of Concept)

### Goals
1. Prove the React + Go API architecture works
2. Establish patterns for other domains to follow
3. Keep scope minimal to reduce risk
4. Side-by-side operation with existing Templ pages

### Work Items Domain Scope

**Current Templ Implementation** (to be migrated):
- List view of work items
- Create/edit forms
- Delete confirmation
- Search/filter functionality

**Data Model** (already exists in Go):
```go
// internal/agency/models/work_item.go
type WorkItem struct {
    ID                string    `json:"_key"`
    Title             string    `json:"title"`
    Description       string    `json:"description"`
    AgencyID          string    `json:"agency_id"`
    Type              string    `json:"type"`           // "goal", "task", etc.
    Status            string    `json:"status"`         // "active", "completed", etc.
    AutonomyLevel     string    `json:"autonomy_level"` // "0-human", "1-supervised", etc.
    CreatedAt         time.Time `json:"created_at"`
    UpdatedAt         time.Time `json:"updated_at"`
}
```

---

## REST API Specification - Work Items

### Base URL
```
http://localhost:8080/api/v1
```

### Authentication
- **Current**: Session-based (cookies)
- **Future**: JWT tokens (Phase 2)
- **Temporary**: CORS enabled for `http://localhost:5173`

### Endpoints

#### 1. List Work Items
```http
GET /api/v1/agencies/{agencyID}/work-items
```

**Query Parameters:**
- `type` (optional): Filter by work item type
- `status` (optional): Filter by status
- `search` (optional): Search in title/description
- `limit` (optional): Pagination limit (default: 50)
- `offset` (optional): Pagination offset (default: 0)

**Response (200 OK):**
```json
{
  "data": [
    {
      "_key": "work-item-001",
      "title": "Implement user authentication",
      "description": "Add JWT-based auth system",
      "agency_id": "UC-INFRA-001",
      "type": "task",
      "status": "active",
      "autonomy_level": "1-supervised",
      "created_at": "2026-01-20T10:00:00Z",
      "updated_at": "2026-01-25T15:30:00Z"
    }
  ],
  "pagination": {
    "total": 42,
    "limit": 50,
    "offset": 0,
    "has_more": false
  }
}
```

#### 2. Get Single Work Item
```http
GET /api/v1/agencies/{agencyID}/work-items/{workItemID}
```

**Response (200 OK):**
```json
{
  "data": {
    "_key": "work-item-001",
    "title": "Implement user authentication",
    "description": "Add JWT-based auth system",
    "agency_id": "UC-INFRA-001",
    "type": "task",
    "status": "active",
    "autonomy_level": "1-supervised",
    "created_at": "2026-01-20T10:00:00Z",
    "updated_at": "2026-01-25T15:30:00Z"
  }
}
```

#### 3. Create Work Item
```http
POST /api/v1/agencies/{agencyID}/work-items
```

**Request Body:**
```json
{
  "title": "New work item",
  "description": "Detailed description",
  "type": "task",
  "status": "active",
  "autonomy_level": "1-supervised"
}
```

**Response (201 Created):**
```json
{
  "data": {
    "_key": "work-item-new-123",
    "title": "New work item",
    "description": "Detailed description",
    "agency_id": "UC-INFRA-001",
    "type": "task",
    "status": "active",
    "autonomy_level": "1-supervised",
    "created_at": "2026-01-26T12:00:00Z",
    "updated_at": "2026-01-26T12:00:00Z"
  }
}
```

#### 4. Update Work Item
```http
PUT /api/v1/agencies/{agencyID}/work-items/{workItemID}
```

**Request Body:**
```json
{
  "title": "Updated title",
  "description": "Updated description",
  "status": "completed"
}
```

**Response (200 OK):**
```json
{
  "data": {
    "_key": "work-item-001",
    "title": "Updated title",
    "description": "Updated description",
    "agency_id": "UC-INFRA-001",
    "type": "task",
    "status": "completed",
    "autonomy_level": "1-supervised",
    "created_at": "2026-01-20T10:00:00Z",
    "updated_at": "2026-01-26T12:05:00Z"
  }
}
```

#### 5. Delete Work Item
```http
DELETE /api/v1/agencies/{agencyID}/work-items/{workItemID}
```

**Response (204 No Content)**

#### Error Responses

**400 Bad Request:**
```json
{
  "error": {
    "code": "INVALID_INPUT",
    "message": "Title is required",
    "details": {
      "field": "title",
      "reason": "Field cannot be empty"
    }
  }
}
```

**404 Not Found:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Work item not found",
    "details": {
      "work_item_id": "work-item-999"
    }
  }
}
```

**500 Internal Server Error:**
```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "An unexpected error occurred",
    "details": {}
  }
}
```

---

## CodeValdFortex Project Structure

### Directory Layout

```
/workspaces/covaldfortex/
├── public/
│   └── favicon.ico
├── src/
│   ├── api/                     # API client layer
│   │   ├── client.ts            # Axios instance, interceptors
│   │   └── workItems.ts         # Work items API calls
│   │
│   ├── app/                     # App-level setup
│   │   ├── App.tsx              # Root component
│   │   ├── store.ts             # Redux store configuration
│   │   └── routes.tsx           # React Router routes
│   │
│   ├── features/                # Feature-based organization
│   │   └── workItems/           # Work Items domain
│   │       ├── components/      # Work Items UI components
│   │       │   ├── WorkItemList.tsx
│   │       │   ├── WorkItemCard.tsx
│   │       │   ├── WorkItemForm.tsx
│   │       │   └── WorkItemFilters.tsx
│   │       ├── slices/          # Redux slice
│   │       │   └── workItemsSlice.ts
│   │       ├── types.ts         # TypeScript types
│   │       └── hooks.ts         # Custom hooks
│   │
│   ├── shared/                  # Shared utilities
│   │   ├── components/          # Reusable components
│   │   │   ├── Button.tsx
│   │   │   ├── Input.tsx
│   │   │   ├── Card.tsx
│   │   │   └── Modal.tsx
│   │   ├── hooks/               # Shared hooks
│   │   │   └── useDebounce.ts
│   │   └── utils/               # Utility functions
│   │       ├── date.ts
│   │       └── validation.ts
│   │
│   ├── styles/                  # Global styles
│   │   ├── bulma.scss           # Bulma CSS imports
│   │   └── custom.scss          # Custom overrides
│   │
│   ├── main.tsx                 # Entry point
│   └── vite-env.d.ts            # Vite type declarations
│
├── .env.development             # Dev environment vars
├── .env.production              # Prod environment vars
├── .gitignore
├── index.html                   # HTML template
├── package.json
├── tsconfig.json                # TypeScript config
├── tsconfig.node.json           # TypeScript config for Vite
├── vite.config.ts               # Vite configuration
└── README.md
```

---

## Redux Store Architecture

### Store Structure

```typescript
// src/app/store.ts
import { configureStore } from '@reduxjs/toolkit';
import workItemsReducer from '../features/workItems/slices/workItemsSlice';

export const store = configureStore({
  reducer: {
    workItems: workItemsReducer,
    // Future: agencies, agents, workflows, etc.
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        // Ignore these action types for date serialization
        ignoredActions: ['workItems/fetchWorkItems/fulfilled'],
      },
    }),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
```

### Work Items Slice

```typescript
// src/features/workItems/slices/workItemsSlice.ts
import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { WorkItem, WorkItemFilters } from '../types';
import * as workItemsApi from '../../../api/workItems';

interface WorkItemsState {
  items: WorkItem[];
  selectedItem: WorkItem | null;
  filters: WorkItemFilters;
  pagination: {
    total: number;
    limit: number;
    offset: number;
    hasMore: boolean;
  };
  loading: boolean;
  error: string | null;
}

const initialState: WorkItemsState = {
  items: [],
  selectedItem: null,
  filters: {
    search: '',
    type: undefined,
    status: undefined,
  },
  pagination: {
    total: 0,
    limit: 50,
    offset: 0,
    hasMore: false,
  },
  loading: false,
  error: null,
};

// Async thunks
export const fetchWorkItems = createAsyncThunk(
  'workItems/fetchWorkItems',
  async ({ agencyId, filters }: { agencyId: string; filters?: WorkItemFilters }) => {
    const response = await workItemsApi.listWorkItems(agencyId, filters);
    return response.data;
  }
);

export const createWorkItem = createAsyncThunk(
  'workItems/createWorkItem',
  async ({ agencyId, data }: { agencyId: string; data: Partial<WorkItem> }) => {
    const response = await workItemsApi.createWorkItem(agencyId, data);
    return response.data;
  }
);

export const updateWorkItem = createAsyncThunk(
  'workItems/updateWorkItem',
  async ({ agencyId, workItemId, data }: { agencyId: string; workItemId: string; data: Partial<WorkItem> }) => {
    const response = await workItemsApi.updateWorkItem(agencyId, workItemId, data);
    return response.data;
  }
);

export const deleteWorkItem = createAsyncThunk(
  'workItems/deleteWorkItem',
  async ({ agencyId, workItemId }: { agencyId: string; workItemId: string }) => {
    await workItemsApi.deleteWorkItem(agencyId, workItemId);
    return workItemId;
  }
);

// Slice
const workItemsSlice = createSlice({
  name: 'workItems',
  initialState,
  reducers: {
    setFilters: (state, action: PayloadAction<WorkItemFilters>) => {
      state.filters = action.payload;
    },
    clearSelectedItem: (state) => {
      state.selectedItem = null;
    },
    selectItem: (state, action: PayloadAction<WorkItem>) => {
      state.selectedItem = action.payload;
    },
  },
  extraReducers: (builder) => {
    // Fetch work items
    builder.addCase(fetchWorkItems.pending, (state) => {
      state.loading = true;
      state.error = null;
    });
    builder.addCase(fetchWorkItems.fulfilled, (state, action) => {
      state.loading = false;
      state.items = action.payload.data;
      state.pagination = action.payload.pagination;
    });
    builder.addCase(fetchWorkItems.rejected, (state, action) => {
      state.loading = false;
      state.error = action.error.message || 'Failed to fetch work items';
    });
    
    // Create work item
    builder.addCase(createWorkItem.fulfilled, (state, action) => {
      state.items.unshift(action.payload.data);
      state.pagination.total += 1;
    });
    
    // Update work item
    builder.addCase(updateWorkItem.fulfilled, (state, action) => {
      const index = state.items.findIndex(item => item._key === action.payload.data._key);
      if (index !== -1) {
        state.items[index] = action.payload.data;
      }
      if (state.selectedItem?._key === action.payload.data._key) {
        state.selectedItem = action.payload.data;
      }
    });
    
    // Delete work item
    builder.addCase(deleteWorkItem.fulfilled, (state, action) => {
      state.items = state.items.filter(item => item._key !== action.payload);
      state.pagination.total -= 1;
      if (state.selectedItem?._key === action.payload) {
        state.selectedItem = null;
      }
    });
  },
});

export const { setFilters, clearSelectedItem, selectItem } = workItemsSlice.actions;
export default workItemsSlice.reducer;
```

---

## React Component Examples

### Work Item List Component

```tsx
// src/features/workItems/components/WorkItemList.tsx
import React, { useEffect } from 'react';
import { useAppDispatch, useAppSelector } from '../../../app/hooks';
import { fetchWorkItems, selectItem } from '../slices/workItemsSlice';
import WorkItemCard from './WorkItemCard';
import WorkItemFilters from './WorkItemFilters';

interface Props {
  agencyId: string;
}

export const WorkItemList: React.FC<Props> = ({ agencyId }) => {
  const dispatch = useAppDispatch();
  const { items, loading, error, filters } = useAppSelector(state => state.workItems);

  useEffect(() => {
    dispatch(fetchWorkItems({ agencyId, filters }));
  }, [dispatch, agencyId, filters]);

  if (loading) {
    return (
      <div className="has-text-centered p-6">
        <div className="is-loading">Loading work items...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="notification is-danger">
        <p>Error: {error}</p>
      </div>
    );
  }

  return (
    <div className="work-items-container">
      <div className="level mb-4">
        <div className="level-left">
          <div className="level-item">
            <h2 className="title is-4">Work Items</h2>
          </div>
        </div>
        <div className="level-right">
          <div className="level-item">
            <button className="button is-primary">
              <span className="icon">
                <i className="fas fa-plus"></i>
              </span>
              <span>New Work Item</span>
            </button>
          </div>
        </div>
      </div>

      <WorkItemFilters />

      {items.length === 0 ? (
        <div className="notification is-info is-light">
          No work items found. Create one to get started!
        </div>
      ) : (
        <div className="columns is-multiline">
          {items.map(item => (
            <div key={item._key} className="column is-one-third">
              <WorkItemCard 
                workItem={item} 
                onClick={() => dispatch(selectItem(item))}
              />
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
```

### Work Item Card Component (Bulma CSS)

```tsx
// src/features/workItems/components/WorkItemCard.tsx
import React from 'react';
import { WorkItem } from '../types';

interface Props {
  workItem: WorkItem;
  onClick: () => void;
}

export const WorkItemCard: React.FC<Props> = ({ workItem, onClick }) => {
  const getStatusClass = (status: string) => {
    switch (status) {
      case 'active': return 'is-success';
      case 'completed': return 'is-info';
      case 'blocked': return 'is-danger';
      default: return 'is-light';
    }
  };

  return (
    <div className="card work-item-card" onClick={onClick} style={{ cursor: 'pointer' }}>
      <div className="card-content">
        <div className="media">
          <div className="media-left">
            <span className="icon is-large has-text-primary">
              <i className="fas fa-tasks fa-2x"></i>
            </span>
          </div>
          <div className="media-content">
            <p className="title is-6">{workItem.title}</p>
            <p className="subtitle is-7 has-text-grey">
              {workItem.description.substring(0, 80)}
              {workItem.description.length > 80 && '...'}
            </p>
          </div>
        </div>

        <div className="tags mt-3">
          <span className={`tag ${getStatusClass(workItem.status)}`}>
            {workItem.status}
          </span>
          <span className="tag is-light">
            {workItem.type}
          </span>
          <span className="tag is-warning is-light">
            {workItem.autonomy_level}
          </span>
        </div>
      </div>

      <footer className="card-footer">
        <a className="card-footer-item" onClick={(e) => { e.stopPropagation(); /* edit */ }}>
          <span className="icon is-small">
            <i className="fas fa-edit"></i>
          </span>
          <span>Edit</span>
        </a>
        <a className="card-footer-item" onClick={(e) => { e.stopPropagation(); /* delete */ }}>
          <span className="icon is-small">
            <i className="fas fa-trash"></i>
          </span>
          <span>Delete</span>
        </a>
      </footer>
    </div>
  );
};
```

---

## Development Workflow

### Running Both Systems Simultaneously

**Terminal 1 - Go Backend:**
```bash
cd /workspaces/CodeValdCortex
make run
# Runs on http://localhost:8080
```

**Terminal 2 - React Frontend:**
```bash
cd /workspaces/CodeValdFortex
npm run dev
# Runs on http://localhost:5173
```

### CORS Configuration (Go Backend)

```go
// internal/api/middleware/cors.go
package middleware

import (
    "github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}
```

### Vite Proxy Configuration (Development)

```typescript
// vite.config.ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
});
```

### Environment Variables

**`.env.development` (CodeValdFortex):**
```bash
VITE_API_BASE_URL=http://localhost:8080
VITE_APP_NAME=CodeValdFortex
```

**`.env.production` (CodeValdFortex):**
```bash
VITE_API_BASE_URL=https://api.codevaldcortex.com
VITE_APP_NAME=CodeValdFortex
```

---

## Deployment Strategy

### Development (Current)
- **Backend**: `make run` → http://localhost:8080
- **Frontend**: `npm run dev` → http://localhost:5173
- **CORS**: Enabled for localhost

### Staging
- **Backend**: Docker container → https://api-staging.codevaldcortex.com
- **Frontend**: Nginx serving static build → https://app-staging.codevaldcortex.com
- **CORS**: Enabled for staging domain

### Production
- **Backend**: Kubernetes deployment → https://api.codevaldcortex.com
- **Frontend**: CDN (CloudFlare/Vercel) → https://app.codevaldcortex.com
- **CORS**: Enabled for production domain only

### Build Process

**CodeValdFortex Production Build:**
```bash
cd /workspaces/CodeValdFortex
npm run build
# Outputs to dist/ directory
# Contains optimized JS, CSS, HTML
```

**Deployment:**
```bash
# Option 1: Static hosting (Vercel, Netlify)
vercel deploy dist/

# Option 2: Self-hosted Nginx
rsync -avz dist/ server:/var/www/CodeValdFortex/
```

---

## Migration Timeline & Milestones

### Phase 1: Foundation (Weeks 1-2)

**Week 1 - Setup & Infrastructure**
- [ ] Create CodeValdFortex repository in `/workspaces/CodeValdFortex/`
- [ ] Initialize Vite + React + TypeScript project
- [ ] Install dependencies (Redux Toolkit, Bulma, Axios, React Router)
- [ ] Set up Bulma CSS with custom theming
- [ ] Configure ESLint, Prettier, TypeScript strict mode
- [ ] Create basic app structure (routing, store, API client)

**Week 2 - Work Items REST API (Go Backend)**
- [ ] Create `internal/api/` package structure
- [ ] Implement Work Items API handlers
  - [ ] GET /api/v1/agencies/:id/work-items
  - [ ] GET /api/v1/agencies/:id/work-items/:workItemId
  - [ ] POST /api/v1/agencies/:id/work-items
  - [ ] PUT /api/v1/agencies/:id/work-items/:workItemId
  - [ ] DELETE /api/v1/agencies/:id/work-items/:workItemId
- [ ] Add CORS middleware
- [ ] Write API integration tests
- [ ] Document API with examples

### Phase 2: Work Items React Implementation (Weeks 3-4)

**Week 3 - Redux & Components**
- [ ] Create Work Items Redux slice with async thunks
- [ ] Build reusable Bulma components (Button, Input, Card, Modal)
- [ ] Implement WorkItemList component
- [ ] Implement WorkItemCard component
- [ ] Implement WorkItemFilters component
- [ ] Add loading states and error handling

**Week 4 - Forms & CRUD Operations**
- [ ] Implement WorkItemForm component (create/edit)
- [ ] Add form validation
- [ ] Implement delete confirmation modal
- [ ] Add success/error notifications
- [ ] Write component tests (React Testing Library)
- [ ] End-to-end testing with backend

### Phase 3: Testing & Refinement (Week 5)

- [ ] Complete test coverage (>80%)
- [ ] Performance optimization (lazy loading, memoization)
- [ ] Accessibility audit (WCAG 2.1 AA compliance)
- [ ] Browser compatibility testing
- [ ] Mobile responsiveness testing
- [ ] Security review (XSS, CSRF protection)

### Phase 4: Deployment & Monitoring (Week 6)

- [ ] Set up staging environment
- [ ] Deploy backend API to staging
- [ ] Deploy React app to staging
- [ ] User acceptance testing
- [ ] Performance monitoring setup
- [ ] Error tracking (Sentry/Rollbar)
- [ ] Production deployment
- [ ] Rollback plan documentation

### Phase 5: Gradual Rollout (Weeks 7-8)

- [ ] A/B testing (50% traffic to React, 50% to Templ)
- [ ] Monitor error rates, performance metrics
- [ ] Gather user feedback
- [ ] Fix critical bugs
- [ ] 100% traffic to React Work Items pages
- [ ] Decommission Templ Work Items pages

---

## Future Phases - Domain Migration Order

### Phase 6: Agent Cards Domain
**Complexity**: Medium  
**Estimated**: 3 weeks
- Agent list/grid view
- Agent detail cards
- Start/stop/restart actions
- Real-time status updates (WebSocket integration)

### Phase 7: Agency Dashboard
**Complexity**: Medium-High  
**Estimated**: 4 weeks
- Agency overview
- Statistics widgets
- Quick actions
- Chart.js integration for metrics

### Phase 8: Workflow Designer
**Complexity**: Very High  
**Estimated**: 6-8 weeks
- Drag-and-drop canvas (React DnD or react-beautiful-dnd)
- Complex state management
- Real-time collaboration (future)
- Visual workflow builder

### Phase 9: AI Policy Wizard
**Complexity**: Medium  
**Estimated**: 3 weeks
- Multi-step wizard
- Policy configuration forms
- Validation logic
- Preview/test functionality

### Phase 10: Authentication & User Management
**Complexity**: High  
**Estimated**: 4 weeks
- JWT-based authentication
- Login/logout flows
- User profile management
- Role-based access control

---

## Success Criteria

### Phase 1 (Work Items) Success Metrics

**Technical:**
- [ ] API response time <100ms (95th percentile)
- [ ] Frontend bundle size <500KB (gzipped)
- [ ] Lighthouse performance score >90
- [ ] Zero critical security vulnerabilities
- [ ] Test coverage >80%
- [ ] Zero console errors in production

**User Experience:**
- [ ] Page load time <2 seconds
- [ ] Time to interactive <3 seconds
- [ ] Smooth animations (60fps)
- [ ] Mobile responsive (works on tablets/phones)
- [ ] Accessible (keyboard navigation, screen readers)

**Business:**
- [ ] Feature parity with Templ version
- [ ] Zero data loss during migration
- [ ] <1% error rate in production
- [ ] Positive user feedback
- [ ] Team velocity maintained or improved

---

## Risk Assessment & Mitigation

### High Risks

**Risk 1: Performance Degradation**
- **Impact**: Users experience slower load times
- **Likelihood**: Medium
- **Mitigation**: 
  - Implement code splitting
  - Use React.lazy() for route-based code splitting
  - Monitor bundle size with CI checks
  - Set up performance budgets

**Risk 2: Data Consistency Issues**
- **Impact**: Work items not synced properly between systems
- **Likelihood**: Low
- **Mitigation**:
  - Thorough integration testing
  - Database transaction handling
  - Rollback procedures documented
  - Side-by-side validation during rollout

**Risk 3: Authentication/Session Handling**
- **Impact**: Users logged out unexpectedly, CORS issues
- **Likelihood**: Medium
- **Mitigation**:
  - Extensive CORS testing
  - Cookie-based session initially (JWT later)
  - Graceful error handling for auth failures
  - Clear error messages

### Medium Risks

**Risk 4: Developer Learning Curve**
- **Impact**: Slower development velocity
- **Likelihood**: Medium
- **Mitigation**:
  - Training sessions on React/Redux
  - Pair programming for first features
  - Comprehensive documentation
  - Code review process

**Risk 5: Increased Deployment Complexity**
- **Impact**: Harder to deploy, more moving parts
- **Likelihood**: Low-Medium
- **Mitigation**:
  - Automated deployment scripts
  - Docker containerization
  - CI/CD pipeline (GitHub Actions)
  - Rollback procedures

---

## Open Questions & Decisions Needed

1. **Authentication**: Should we migrate to JWT immediately or stay with session cookies?
   - **Recommendation**: Stay with sessions for Phase 1, migrate to JWT in Phase 10

2. **Routing**: Use React Router hash mode or history mode?
   - **Recommendation**: History mode (requires server-side rewrites)

3. **Styling**: Import full Bulma or only used components?
   - **Recommendation**: Full Bulma for consistency, optimize later with PurgeCSS

4. **API Versioning**: Start with `/api/v1/` or just `/api/`?
   - **Recommendation**: `/api/v1/` for future-proofing

5. **Error Tracking**: Sentry, Rollbar, or self-hosted?
   - **Recommendation**: Sentry (free tier sufficient for MVP)

6. **Testing**: Jest + React Testing Library or Vitest?
   - **Recommendation**: Vitest (faster, better Vite integration)

---

## Next Steps

### Immediate Actions (This Week)

1. **Review & Approve Plan**: Architecture team sign-off
2. **Create CodeValdFortex Repository**: Initialize Vite project
3. **Set Up Development Environment**: Ensure both repos run in container
4. **Create Feature Branch**: `feature/react-migration-phase1`
5. **Begin Week 1 Tasks**: Project scaffolding

### Communication Plan

- **Weekly Status Updates**: Every Friday, migration progress report
- **Blocker Discussions**: Daily standup mention of blockers
- **Demo Sessions**: End of each phase, demo to stakeholders
- **Documentation Updates**: Update this plan as decisions are made

---

## Appendix

### Technology Stack Summary

| Category | Technology | Version | Purpose |
|----------|-----------|---------|---------|
| **Backend** | Go | 1.21+ | Business logic, REST API |
| **Backend Framework** | Gin | 1.9+ | HTTP routing, middleware |
| **Database** | ArangoDB | 3.11+ | Multi-model database |
| **Frontend Framework** | React | 18.2+ | UI library |
| **Build Tool** | Vite | 5.x | Fast build tool |
| **Language** | TypeScript | 5.x | Type safety |
| **State Management** | Redux Toolkit | 2.x | Global state |
| **CSS Framework** | Bulma | 1.0+ | Styling |
| **HTTP Client** | Axios | 1.6+ | API requests |
| **Routing** | React Router | 6.x | Client-side routing |
| **Testing** | Vitest | 1.x | Unit/integration tests |
| **Testing** | React Testing Library | 14.x | Component tests |
| **Linting** | ESLint | 8.x | Code quality |
| **Formatting** | Prettier | 3.x | Code formatting |

### Resources

- [React Documentation](https://react.dev/)
- [Redux Toolkit Documentation](https://redux-toolkit.js.org/)
- [Bulma Documentation](https://bulma.io/documentation/)
- [Vite Documentation](https://vitejs.dev/)
- [TypeScript Documentation](https://www.typescriptlang.org/docs/)
- [Gin Framework](https://gin-gonic.com/docs/)

---

**Document Status**: Draft  
**Last Updated**: January 26, 2026  
**Next Review**: End of Week 1 (Project setup complete)
