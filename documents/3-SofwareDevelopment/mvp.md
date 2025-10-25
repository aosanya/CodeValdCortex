# MVP - Minimum Viable Product Task Breakdown

## Task Overview
- **Objective**: Define and execute the minimum set of tasks required to launch a functional product that delivers core value to users
- **Success Criteria**: Deployable system with essential features that satisfies primary user needs and business objectives
- **Dependencies**: Infrastructure foundation and core technical architecture decisions

## Foundation Tasks (P0 - Blocking)

*All foundation tasks completed. See `mvp_done.md` for details.*

## Core Agent Mechanics (P0 - Blocking)

*All core agent mechanics tasks completed. See `mvp_done.md` for details.*

## Core Functionality Tasks (P1 - Critical)

| Task ID | Title                          | Description                                                                        | Status      | Priority | Effort | Skills Required         | Dependencies |
| ------- | ------------------------------ | ---------------------------------------------------------------------------------- | ----------- | -------- | ------ | ----------------------- | ------------ |
| MVP-021 | Agency Management System       | Create database schema and backend services for managing agencies (use cases). Store agency metadata, configurations, and settings in ArangoDB. Implement CRUD operations and API endpoints for agency lifecycle management | Not Started | P1       | High   | Go, ArangoDB, Backend Dev | MVP-010      |
| MVP-022 | Agency Selection Homepage      | Build homepage UI for selecting and switching between agencies. Display available agencies with descriptions, set selected agency as app context, persist selection in session. Integrate with dashboard navigation | Not Started | P1       | Medium | Go, Templ, Frontend Dev | MVP-021      |

## Platform Integration Tasks (P1 - Critical)

| Task ID | Title                 | Description                                                                        | Status      | Priority | Effort | Skills Required         | Dependencies |
| ------- | --------------------- | ---------------------------------------------------------------------------------- | ----------- | -------- | ------ | ----------------------- | ------------ |
| MVP-014 | Kubernetes Deployment | Create Kubernetes manifests and Helm charts for agent deployment                   | Not Started | P1       | High   | DevOps, Kubernetes      | MVP-010      |
| MVP-015 | Management Dashboard  | Build web interface with Templ+HTMX+Alpine.js for agent monitoring, real-time updates, and control | In Progress | P1       | Medium | Go, Frontend Dev, Templ | MVP-013      |

## Authentication & Security Tasks (P2 - Important)

| Task ID | Title                     | Description                                                | Status      | Priority | Effort | Skills Required       | Dependencies |
| ------- | ------------------------- | ---------------------------------------------------------- | ----------- | -------- | ------ | --------------------- | ------------ |
| MVP-024 | Basic User Authentication | Implement user registration, login, and session management | Not Started | P2       | Medium | Backend Dev, Security | MVP-014      |
| MVP-025 | Security Implementation   | Add input validation, HTTPS, and basic security headers    | Not Started | P2       | Medium | Security, Backend Dev | MVP-024      |
| MVP-026 | Access Control System     | Implement role-based access control for agent operations   | Not Started | P2       | Low    | Backend Dev, Security | MVP-025      |

## Agency Management Feature (P1 - Critical)

*Enables multi-tenant architecture where each use case operates as an independent agency with its own configuration, agents, and context*

### MVP-021: Agency Management System

**Objective**: Build the backend infrastructure for managing agencies (use cases) as first-class entities in the system, with database persistence and full CRUD operations.

**Key Deliverables**:

1. **Database Schema (ArangoDB)**:
   - `agencies` collection with fields:
     ```json
     {
       "_key": "UC-INFRA-001",
       "id": "UC-INFRA-001",
       "name": "Water Distribution Network",
       "display_name": "💧 Water Distribution",
       "description": "Smart water infrastructure monitoring and management",
       "category": "infrastructure",
       "icon": "💧",
       "status": "active",
       "config_path": "/usecases/UC-INFRA-001-water-distribution-network",
       "env_file": ".env",
       "metadata": {
         "location": "Nairobi, Kenya",
         "agent_types": ["pipe", "sensor", "pump", "valve", "coordinator"],
         "total_agents": 293,
         "zones": 5,
         "api_endpoint": "/api/v1/agencies/UC-INFRA-001"
       },
       "settings": {
         "auto_start": true,
         "monitoring_enabled": true,
         "dashboard_enabled": true,
         "visualizer_enabled": true
       },
       "created_at": "2025-10-20T10:00:00Z",
       "updated_at": "2025-10-25T14:30:00Z",
       "created_by": "system"
     }
     ```
   - Indexes on `id`, `category`, `status` for efficient queries
   - Unique constraint on `id` field

2. **Backend Services**:
   - `AgencyService` interface:
     ```go
     type AgencyService interface {
         CreateAgency(agency *Agency) error
         GetAgency(id string) (*Agency, error)
         ListAgencies(filters AgencyFilters) ([]*Agency, error)
         UpdateAgency(id string, updates AgencyUpdates) error
         DeleteAgency(id string) error
         SetActiveAgency(id string) error
         GetActiveAgency() (*Agency, error)
     }
     ```
   - `AgencyRepository` for ArangoDB operations
   - Configuration loader for agency-specific .env files
   - Validation service for agency configurations

3. **REST API Endpoints**:
   ```
   POST   /api/v1/agencies                    # Create new agency
   GET    /api/v1/agencies                    # List all agencies
   GET    /api/v1/agencies/{id}               # Get agency details
   PUT    /api/v1/agencies/{id}               # Update agency
   DELETE /api/v1/agencies/{id}               # Delete agency
   POST   /api/v1/agencies/{id}/activate      # Set as active agency
   GET    /api/v1/agencies/active             # Get current active agency
   GET    /api/v1/agencies/{id}/agents        # Get agency's agents
   GET    /api/v1/agencies/{id}/statistics    # Get agency statistics
   ```

4. **Agency Context Management**:
   - Middleware to inject current agency context into requests
   - Session storage for active agency selection
   - Environment variable management per agency
   - Scoped agent queries (filter by agency)

5. **Data Migration Script**:
   - Auto-discover existing use cases from `/usecases/` directory
   - Parse use case metadata from folders and .env files
   - Create initial agency records in database
   - Support for bulk import from JSON/YAML

**Acceptance Criteria**:
- Database schema created with proper indexes
- All CRUD operations functional via API
- Agency context correctly scopes agent queries
- Migration script successfully imports 10+ existing use cases
- Unit tests for service layer (>80% coverage)
- API endpoints documented with examples

**Technical Implementation**:
```
/workspaces/CodeValdCortex/
├── internal/
│   ├── agency/
│   │   ├── service.go              # AgencyService implementation
│   │   ├── repository.go           # ArangoDB repository
│   │   ├── types.go                # Agency, AgencyFilters, AgencyUpdates
│   │   ├── validator.go            # Configuration validation
│   │   └── context.go              # Agency context management
│   ├── handlers/
│   │   └── agency_handler.go       # REST API handlers
│   └── middleware/
│       └── agency_context.go       # Agency context middleware
└── scripts/
    └── migrate-agencies.go         # Discovery and migration script
```

### MVP-022: Agency Selection Homepage

**Objective**: Create an intuitive homepage that allows users to view, select, and switch between available agencies, establishing the agency as the primary application context.

**Key Deliverables**:

1. **Homepage UI (Templ + Bulma CSS)**:
   - Agency grid/card layout displaying:
     * Agency icon and display name
     * Category badge (infrastructure, agriculture, logistics, etc.)
     * Brief description
     * Key statistics (agents, zones, status)
     * "Open" button for selection
   - Responsive design (mobile, tablet, desktop)
   - Search and filter functionality by category/status
   - Sort options (name, category, recent, agents count)

2. **Agency Dashboard Integration**:
   - Replace current static dashboard with agency-aware version
   - Header shows current agency context:
     ```
     [Agency Icon] Agency Name | Dashboard
     [Switch Agency] button in navbar
     ```
   - Navigation links scoped to current agency:
     * `/agencies/{id}/dashboard`
     * `/agencies/{id}/agents`
     * `/agencies/{id}/visualizer`
     * `/agencies/{id}/settings`

3. **Session Management**:
   - Store selected agency ID in server-side session
   - Cookie-based session persistence
   - Session timeout handling (redirect to homepage)
   - "Remember my agency" option (localStorage)

4. **Agency Switching**:
   - Modal dialog for switching agencies
   - Confirmation prompt if unsaved changes exist
   - Graceful agent cleanup when switching
   - Preserve user preferences per agency

5. **Routes and Navigation**:
   ```
   GET  /                               # Homepage (agency selection)
   GET  /agencies/{id}                  # Redirect to agency dashboard
   GET  /agencies/{id}/dashboard        # Agency-specific dashboard
   POST /agencies/{id}/select           # Set active agency (session)
   GET  /agencies/switch                # Show agency switcher modal
   ```

6. **Agency Statistics Widget**:
   - Real-time agent count
   - Health status indicator
   - Recent activity summary
   - Quick actions (start/stop all agents)

**UI Mockup Structure**:
```html
<!-- Homepage: Agency Selection -->
<section class="section">
  <div class="container">
    <h1 class="title">Select Agency</h1>
    <p class="subtitle">Choose a use case to begin</p>
    
    <!-- Filters -->
    <div class="field is-grouped">
      <div class="control">
        <div class="select">
          <select id="category-filter">
            <option value="">All Categories</option>
            <option value="infrastructure">Infrastructure</option>
            <option value="agriculture">Agriculture</option>
            <!-- ... -->
          </select>
        </div>
      </div>
      <div class="control is-expanded">
        <input class="input" type="text" placeholder="Search agencies...">
      </div>
    </div>
    
    <!-- Agency Grid -->
    <div class="columns is-multiline">
      <!-- Agency Card -->
      <div class="column is-one-third">
        <div class="card agency-card">
          <div class="card-content">
            <div class="media">
              <div class="media-left">
                <span class="icon is-large">💧</span>
              </div>
              <div class="media-content">
                <p class="title is-4">Water Distribution</p>
                <p class="subtitle is-6">Infrastructure</p>
              </div>
            </div>
            <div class="content">
              <p>Smart water infrastructure monitoring and management</p>
              <div class="tags">
                <span class="tag is-success">Active</span>
                <span class="tag is-info">293 Agents</span>
              </div>
            </div>
          </div>
          <footer class="card-footer">
            <a href="/agencies/UC-INFRA-001/dashboard" 
               class="card-footer-item">Open</a>
          </footer>
        </div>
      </div>
      <!-- Repeat for other agencies -->
    </div>
  </div>
</section>
```

**Acceptance Criteria**:
- Homepage displays all available agencies from database
- Users can search/filter agencies by category
- Clicking "Open" sets agency as active context
- Dashboard and all pages respect agency context
- Agency switcher allows changing agencies mid-session
- Session persists agency selection across page refreshes
- Responsive design works on mobile and desktop
- Integration tests for full navigation flow

**Technical Implementation**:
```
/workspaces/CodeValdCortex/
├── internal/web/
│   ├── templates/
│   │   ├── homepage.templ          # Agency selection page
│   │   ├── agency_switcher.templ   # Modal for switching
│   │   └── dashboard.templ         # Updated with agency context
│   ├── handlers/
│   │   └── homepage_handler.go     # Homepage routes
│   └── middleware/
│       └── require_agency.go       # Require agency selection
└── static/
    ├── css/
    │   └── agencies.css            # Agency-specific styles
    └── js/
        └── agency-switcher.js      # Client-side interactions
```

**Dependencies**:
- MVP-021 must be completed first (agency backend)
- Requires session management implementation
- Integrates with existing dashboard and visualizer

---

## Agent Property Broadcasting Feature (P1 - Critical)

*Enables UC-TRACK-001 (Safiri Salama) and other real-time tracking/monitoring use cases*

| Task ID | Title                                    | Description                                                                                                      | Status      | Priority | Effort | Skills Required            | Dependencies |
| ------- | ---------------------------------------- | ---------------------------------------------------------------------------------------------------------------- | ----------- | -------- | ------ | -------------------------- | ------------ |
| MVP-016 | Core Broadcasting Infrastructure         | Implement BroadcastConfiguration, PropertyBroadcaster service, ContextEvaluator, and integration with PubSub    | Not Started | P1       | High   | Go, Backend Dev, PubSub    | MVP-013      |
| MVP-017 | Subscription Management                  | Build SubscriptionManager, subscriber filtering, favorite functionality, and subscription API endpoints          | Not Started | P1       | Medium | Go, Backend Dev, REST API  | MVP-016      |
| MVP-018 | Privacy & Security Controls              | Implement geofencing, property masking, permission model, audit logging, and encryption for sensitive properties | Not Started | P1       | Medium | Security, Backend Dev      | MVP-017      |
| MVP-019 | Performance Optimization & Scale         | Performance tuning, caching, load balancing for broadcasters, message broker optimization, monitoring & alerting | Not Started | P1       | Medium | Performance, DevOps        | MVP-018      |
| MVP-020 | UC-TRACK-001 Integration & Testing       | Implement Vehicle & Passenger agents, build mobile/web UI, SACCO management portal, end-to-end testing          | Not Started | P1       | High   | Full-stack, Mobile Dev     | MVP-019      |

### MVP-016: Core Broadcasting Infrastructure

**Objective**: Build the foundational broadcasting system that enables agents to automatically publish properties at configurable intervals.

**Key Deliverables**:
1. **Data Structures**:
   - `BroadcastConfiguration` type with rules, intervals, privacy controls
   - `BroadcastRule` with condition evaluation logic
   - `PropertyUpdateMessage` format
   - `BroadcastMetrics` for monitoring

2. **Core Services**:
   - `PropertyBroadcaster` service implementing lifecycle management
   - `ContextEvaluator` for rule matching and interval determination
   - `BroadcastConfigRepository` for persistent storage

3. **Agent Integration**:
   - Extend Agent base class with broadcasting methods
   - Add `EnableBroadcasting()`, `StartBroadcasting()`, `StopBroadcasting()`
   - Add `UpdateBroadcastInterval()`, `PauseBroadcasting()`, `BroadcastNow()`
   - Implement property collection from agent state

4. **PubSub Integration**:
   - Extend PubSubService with `PublishPropertyUpdate()`
   - Add topic routing for property updates
   - Implement message formatting and priority handling

**Acceptance Criteria**:
- Agent can configure and start broadcasting
- Context-aware interval adjustment works correctly
- Properties are published to message bus
- Basic metrics collection functional
- Unit tests for all core components (>80% coverage)

**Technical Details**: See `documents/3-SofwareDevelopment/core-systems/agent-property-broadcasting.md`

### MVP-017: Subscription Management

**Objective**: Enable agents to subscribe to property updates from other agents with filtering and notification preferences.

**Key Deliverables**:
1. **Subscription Service**:
   - `SubscriptionManager` interface and implementation
   - Subscribe/Unsubscribe operations
   - Favorite/Unfavorite functionality
   - Subscription approval workflow

2. **Filtering Logic**:
   - Property-based filtering
   - Priority-based filtering
   - Geofence-based filtering
   - Time window filtering
   - Subscriber type restrictions

3. **Data Models**:
   - `Subscriber` type with preferences
   - `SubscriptionFilters` configuration
   - `NotificationPreferences` settings

4. **REST API Endpoints**:
   ```
   POST   /api/v1/agents/{agentId}/broadcasting/subscribe
   DELETE /api/v1/agents/{agentId}/broadcasting/unsubscribe
   POST   /api/v1/agents/{agentId}/broadcasting/favorite
   DELETE /api/v1/agents/{agentId}/broadcasting/favorite
   GET    /api/v1/agents/{agentId}/broadcasting/subscribers
   GET    /api/v1/agents/{agentId}/subscribers/subscriptions
   ```

5. **Notification Delivery**:
   - Push notification integration
   - SMS notification integration
   - Email notification integration
   - In-app notification handling

**Acceptance Criteria**:
- Subscribers receive filtered property updates
- Favorite notifications work with priority
- API endpoints functional and documented
- Subscription persistence works correctly
- Integration tests for subscribe/receive flow

### MVP-018: Privacy & Security Controls

**Objective**: Implement comprehensive privacy and security features to protect sensitive location and property data.

**Key Deliverables**:
1. **Geofencing Service**:
   - Geofence zone definition and storage
   - Point-in-polygon detection
   - Integration with broadcast pause logic
   - Admin UI for geofence management

2. **Property Masking**:
   - Masking rule configuration per property
   - Context-aware masking (e.g., mask exact location, show only vicinity)
   - Redaction of sensitive fields

3. **Permission Model**:
   - `BroadcastPermissions` service
   - Subscriber authorization checks
   - Property-level access control
   - Approval workflow for sensitive publishers

4. **Audit Logging**:
   - Log all subscription requests
   - Track property access patterns
   - Record broadcast pause/resume events
   - Generate compliance reports

5. **Encryption**:
   - Encrypt sensitive properties in transit
   - Secure storage of subscriber credentials
   - API authentication and authorization

**Acceptance Criteria**:
- Broadcasting pauses in restricted geofences
- Property masking works correctly
- Unauthorized subscribers cannot access data
- Complete audit trail for all operations
- Security tests pass (no critical vulnerabilities)

### MVP-019: Performance Optimization & Scale

**Objective**: Ensure the broadcasting system can handle production-scale loads with low latency and high reliability.

**Key Deliverables**:
1. **Caching Strategy**:
   - Cache subscriber lists in memory with TTL
   - Cache broadcast configurations
   - Cache geofence definitions
   - Redis integration for distributed caching

2. **Performance Optimizations**:
   - Batch property updates where possible
   - Async message publishing
   - Connection pooling for message broker
   - Efficient subscriber filtering at source

3. **Load Balancing**:
   - Distribute agents across broadcaster instances
   - Sharding strategy for high-volume agents
   - Health checks and failover

4. **Monitoring & Alerting**:
   - Prometheus metrics integration
   - Grafana dashboards for broadcast metrics
   - Alert rules for anomalies (high latency, failures)
   - Performance profiling tools

5. **Resource Limits**:
   - Implement `BroadcastResourceLimits`
   - Rate limiting per agent
   - Throttling for misbehaving agents
   - Circuit breakers for external services

**Acceptance Criteria**:
- Support 10,000+ concurrent broadcasting agents
- Sub-second delivery latency (p99 < 500ms)
- Handle 100,000+ messages per minute
- Memory usage stays within bounds under load
- Passes load and stress tests

**Performance Targets**:
- Broadcast-to-delivery latency: <500ms (p99)
- Concurrent agents: 10,000+
- Subscribers per agent: 1,000+
- Message throughput: 100,000/min
- System uptime: 99.9%

### MVP-020: UC-TRACK-001 Integration & Testing

**Objective**: Complete end-to-end implementation of the Safiri Salama tracking system as the reference implementation.

**Key Deliverables**:
1. **Vehicle Agent Implementation**:
   - GPS location tracking integration
   - Context-aware broadcast configuration
   - Status state machine (idle, en_route, at_stop, etc.)
   - Driver dashboard (mobile/tablet)

2. **Passenger Agent Implementation**:
   - Favorite vehicle management
   - Smart notification handling
   - Trip rating and feedback
   - Commute pattern learning

3. **Parent Agent Implementation**:
   - Child bus tracking
   - Emergency alert handling
   - Proximity notifications
   - School communication portal

4. **Fleet Operator Portal**:
   - SACCO manager dashboard
   - Loyalty analytics
   - Vehicle performance metrics
   - Passenger engagement insights

5. **Mobile Applications**:
   - Passenger app (iOS/Android) - track favorites
   - Parent app (iOS/Android) - track school bus
   - Driver app (iOS/Android) - status control

6. **Web Portal**:
   - Admin dashboard for fleet operators
   - Real-time fleet map view
   - Analytics and reporting
   - Configuration management

7. **Use Case Deployment**:
   - UC-TRACK-001 folder structure
   - Agent JSON schemas (Vehicle, Passenger, Parent, etc.)
   - Environment configuration (.env)
   - Deployment scripts (start.sh)
   - Documentation and user guides

**Acceptance Criteria**:
- All 5 agent types functional (Vehicle, Parent, Passenger, Route Manager, Fleet Operator)
- Mobile apps published to app stores (beta)
- End-to-end tracking works: Vehicle → Broadcast → Passenger receives
- Favorite notification feature works
- Real-world pilot with 5+ vehicles and 50+ subscribers
- User feedback positive (>4.0 rating)
- System stable under real usage patterns

**Pilot Program**:
1. **School Pilot**: 2-3 schools, 5-10 buses, 200-300 parents
2. **Matatu Pilot**: 1-2 SACCOs, 10-15 vehicles, 500-1000 passengers
3. **Duration**: 4 weeks
4. **Success Metrics**: >70% active usage, <5% error rate, positive feedback

---

## Resource Requirements

### Team Members
- **Backend Developer**: API development, database design, security implementation
- **Frontend Developer**: UI/UX implementation, responsive design, user experience
- **DevOps Engineer**: Infrastructure setup, CI/CD, production deployment
- **QA Engineer**: Testing strategy, test automation, quality assurance

### Tools and Platforms
- **Development**: Git, Docker, VS Code/IDE of choice
- **Backend**: Node.js/Python/Go (TBD), REST/GraphQL APIs
- **Frontend**: React/Vue/Angular (TBD), CSS frameworks
- **Database**: PostgreSQL/MongoDB (TBD)
- **CI/CD**: GitHub Actions/GitLab CI (TBD)
- **Monitoring**: Basic logging and health checks

### Infrastructure
- **Hosting**: Cloud provider (AWS/GCP/Azure TBD)
- **Environments**: Development, staging, production
- **CDN**: Basic content delivery for static assets
- **SSL**: Certificate management and HTTPS enforcement

## Risk Assessment

### Identified Risks
- **Technical Debt**: Rushing MVP features may compromise code quality
- **Scope Creep**: Adding non-essential features that delay launch
- **Performance Issues**: Scalability problems under load
- **Security Vulnerabilities**: Inadequate security implementation
- **Integration Challenges**: Third-party service dependencies

### Mitigation Strategies
- **Code Reviews**: Mandatory peer review for all code changes
- **Feature Freeze**: Strict adherence to MVP scope definition
- **Load Testing**: Early performance testing with realistic data volumes
- **Security Audits**: Regular security reviews and penetration testing
- **Fallback Plans**: Alternative solutions for critical third-party dependencies

### Contingency Plans
- **MVP Scope Reduction**: Remove P2/P3 features if timeline is at risk
- **Technical Alternatives**: Backup technology choices for critical components
- **Extended Timeline**: Buffer time for unexpected complications
- **Resource Scaling**: Option to add temporary team members if needed

## MVP Success Metrics

### Technical Metrics
- **Uptime**: 99%+ availability during business hours
- **Response Time**: <2 seconds for critical user actions
- **Security**: Zero critical vulnerabilities at launch
- **Performance**: Support for 100+ concurrent users

### User Metrics
- **Registration**: >80% completion rate for sign-up flow
- **Workflow Completion**: >70% completion rate for primary user journey
- **User Retention**: >50% of users return within first week
- **Error Rate**: <5% user-facing errors

### Business Metrics
- **Timeline**: Launch within planned development window
- **Budget**: Stay within allocated development resources
- **Value Validation**: Demonstrate core value proposition to target users
- **Market Readiness**: Receive positive feedback from beta users

## Workflow Integration

### Task Management Process
1. **Task Assignment**: Pick tasks based on priority (P0 first) and dependencies
2. **Implementation**: Update "Status" column as work progresses (Not Started → In Progress → Testing → Complete)
3. **Completion Process** (MANDATORY):
   - Create detailed coding session document in `coding_sessions/` using format: `{TaskID}_{description}.md`
   - Add completed task to summary table in `mvp_done.md` with completion date
   - Remove completed task from this active `mvp.md` file
   - Update any dependent task references
   - Merge feature branch to main:
     ```bash
     # Merge when complete and tested
     git checkout main
     git merge feature/MVP-XXX_description
     git branch -d feature/MVP-XXX_description
     git push origin main
     ```
4. **Dependencies**: Ensure prerequisite tasks are completed before starting dependent work

### Branch Management (MANDATORY)
For each new task:
```bash
# Create feature branch
git checkout -b feature/MVP-XXX_description

# Work on task implementation
# ... development work ...

# Build validation before merge
# - Follow coding standards
# - Run linting and validation tools
# - Verify code follows established patterns
# - Check for deprecated API usage
# - Remove unused code/imports/variables
# - Run build processes and tests
# - Fix any build errors or warnings

# Merge when complete and tested
git checkout main
git merge feature/MVP-XXX_description
git branch -d feature/MVP-XXX_description
```

### Repository Structure
```
/workspaces/CodeValdCortex/
├── documents/3-SofwareDevelopment/
│   ├── mvp.md                    # This file - Active tasks only
│   ├── mvp_done.md              # Completed tasks archive
│   └── coding_sessions/         # Detailed implementation logs
├── [project code structure]     # Implementation code
└── [other project folders]      # Additional project resources
```

---

**Note**: This document contains only active and pending tasks. All completed tasks are moved to `mvp_done.md` to maintain a clean, actionable backlog.