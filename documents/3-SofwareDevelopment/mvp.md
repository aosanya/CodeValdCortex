# MVP - Minimum Viable Product Task Breakdown

## Task Overview
- **Objective**: Define and execute the minimum set of tasks required to launch a functional product that delivers core value to users
- **Success Criteria**: Deployable system with essential features that satisfies primary user needs and business objectives
- **Dependencies**: Infrastructure foundation and core technical architecture decisions

## Documentation Structure
- **High-Level Overview**: This file (`mvp.md`) provides task tables, priorities, dependencies, and brief descriptions
- **Detailed Specifications**: Each task with detailed requirements is documented in `/documents/3-SofwareDevelopment/mvp-details/{TASK_ID}.md`
- **Reference Pattern**: Tasks reference their detail files using the format `See: mvp-details/{TASK_ID}.md`

## Workflow Integration

### Task Management Process
1. **Task Assignment**: Pick tasks based on priority (P0 first) and dependencies
2. **Implementation**: Update "Status" column as work progresses (Not Started → In Progress → Testing → Complete)
3. **Completion Process** (MANDATORY):
   - Create detailed coding session document in `coding_sessions/` using format: `{TaskID}_{description}.md`
   - Add completed task to summary table in `mvp_done.md` with completion date
   - Remove completed task from this active `mvp.md` file
   - Update any dependent task references
   - Merge feature branch to main
4. **Dependencies**: Ensure prerequisite tasks are completed before starting dependent work

### Branch Management (MANDATORY)
```bash
# Create feature branch
git checkout -b feature/MVP-XXX_description

# Work, build validation, test
# ... development work ...

# Merge when complete and tested
git checkout main
git merge feature/MVP-XXX_description --no-ff
git branch -d feature/MVP-XXX_description
```

---

## Status Legend
- ✅ **Completed** - Task done, merged to main (see `mvp_done.md`)
- 🚀 **In Progress** - Currently being worked on
- 📋 **Not Started** - Ready to begin (dependencies met)
- ⏸️ **Blocked** - Waiting on dependencies
- ⚠️ **Deprecated** - Superseded by other work

---

## P1: React Frontend API Support (PREREQUISITE FOR REACT MIGRATION)

*REST API infrastructure to support the new React SPA frontend (CodeValdFortex)*

| Task ID | Title | Description | Status | Priority | Effort | Skills | Dependencies | Details |
|---------|-------|-------------|--------|----------|--------|--------|--------------|---------|
| MVP-RM-003 | Work Items REST API | Implement 5 REST endpoints for work items (list with pagination/filtering, get single, create, update, delete). Enforce agency-specific data isolation | ✅ Completed | P1 | Medium | Go, Gin, ArangoDB | ~~MVP-RM-002~~ ✅ | [react-migration/MVP-RM-003_work_items_api.md](mvp-details/react-migration/MVP-RM-003_work_items_api.md) |

---

## P0: Agency Designer - Core Features (CRITICAL)

*Foundation for agency configuration and design*

| Task ID | Title | Description | Status | Priority | Effort | Skills | Dependencies | Details |
|---------|-------|-------------|--------|----------|--------|--------|--------------|---------|
| MVP-046 | Agency Admin & Configuration Page | Build comprehensive admin interface for agency-wide settings: token budgets (role & individual agent levels), rate limits, resource quotas, AI model selection, cost controls, monitoring dashboards, and operational parameters | 📋 Not Started | P0 | Medium | Go, Templ, Frontend Dev, Analytics | MVP-044 ✅ | [MVP-046.md](mvp-details/MVP-046.md) |
| MVP-047 | Agency Designer Export System | Implement comprehensive export functionality for entire agency design (all sections) to PDF, Markdown, and JSON formats with customizable templates and branding | 📋 Not Started | P0 | Medium | Go, PDF Generation, File Export | MVP-044 ✅ | [MVP-047.md](mvp-details/MVP-047.md) |
| MVP-049 | AI Policy Layer - Runtime Enforcement | Build action authorization, approval workflows, risk scoring, budget tracking, and policy violation handling with real-time feedback and audit logging | 📋 Not Started | P0 | High | Go, Security, Backend Dev | ~~MVP-048~~ ✅ | [AI Policy Layer](../../2-SoftwareDesignAndArchitecture/ai-policy-layer.md) |
| MVP-050 | AI Policy Layer - Advanced Features | Implement data classification engine, PII detection/masking, compliance reporting, policy versioning, and multi-policy inheritance | 📋 Not Started | P0 | Medium | Go, Security, ML, Backend Dev | MVP-049 | [AI Policy Layer](../../2-SoftwareDesignAndArchitecture/ai-policy-layer.md) |
| MVP-042 | AI-Powered Agency Creator | Implement AI-driven agency creation flow with text upload, selective generation (introduction, goals, work items, roles, RACI), and batch AI generation | 📋 Not Started | P0 | High | Go, Templ, AI/LLM, Frontend Dev | MVP-047 | [MVP-042.md](mvp-details/MVP-042.md) |

---

## P0: Work Items & Document Management System (FOUNDATIONAL)

*Git-in-ArangoDB with Kanban workflow automation and agent instantiation*

### Core Architecture

| Task ID | Title | Description | Status | Priority | Effort | Skills | Dependencies | Details |
|---------|-------|-------------|--------|----------|--------|--------|--------------|---------|
| MVP-WI-007 | Pull Request System | Implement PR creation/management, diff generation, three-way merge algorithm, conflict detection, PR review workflow, approval tracking, merge operations | 📋 Not Started | P0 | High | Go, ArangoDB, Git Merge Logic | ~~MVP-WI-006~~ ✅ | [work-items-integration/pull-requests.md](mvp-details/work-items-integration/pull-requests.md) |
| MVP-WI-009 | Issue-Git Integration | Implement branch creation from issues, issue-branch linking, PR-issue association, automatic issue progression on PR merge, workflow orchestration service | 📋 Not Started | P0 | Medium | Go, ArangoDB, Event System | MVP-WI-007, ~~MVP-WI-008~~ ✅, ~~MVP-054~~ ✅ | [work-items-integration/kanban-workflow.md](mvp-details/work-items-integration/kanban-workflow.md) |
| MVP-WI-010 | Agent Work Assignment | Implement flexible assignment model (manual + claim-based), agent-issue linking, work queue API, notification system, agent instance creation from work item definitions | 📋 Not Started | P0 | Medium | Go, ArangoDB, Backend Dev | MVP-WI-009, MVP-032 | [work-items-integration/kanban-workflow.md](mvp-details/work-items-integration/kanban-workflow.md) |
| MVP-WI-011 | AI-Assisted Conflict Resolution | Build conflict detection in merges, AI-powered merge suggestions, conflict resolution UI, three-way merge with AI assistance, merge conflict analytics | 📋 Not Started | P1 | High | Go, AI/LLM, Git Merge, Frontend Dev | MVP-WI-007 | [work-items-integration/git-based-document-system.md](mvp-details/work-items-integration/git-based-document-system.md#ai-assisted-conflict-resolution) |
| MVP-WI-012 | Workbench Chat Panel & Workflows Section | Integrate chat panel and workflows section into the Workbench UI, reusing agency designer components and backend wiring. Ensure HTMX dynamic updates, multi-tenancy, and context-aware issue creation. See: mvp-details/work-items-integration/workbench-chat-panel.md | 🗒️ Not Started | P0 | Medium | Go, Templ, HTMX, Frontend Dev | MVP-WI-008 | [work-items-integration/workbench-chat-panel.md](mvp-details/work-items-integration/workbench-chat-panel.md) |

### Work Item Definitions & Workflows

| Task ID | Title | Description | Status | Priority | Effort | Skills | Dependencies | Details |
|---------|-------|-------------|--------|----------|--------|--------|--------------|---------|
| MVP-031 | Work Item Lifecycle & SLA | Build work item state machines (open→assigned→in_progress→ready_for_review→completed), SLA tracking, deadline management, state transition validation | 📋 Not Started | P0 | Medium | Go, ArangoDB, State Machine | ~~MVP-030~~ ✅ | [work-items-integration/work-item-schema.md](mvp-details/work-items-integration/work-item-schema.md) |
| MVP-032 | Agent Factory & Orchestration | Implement AgentFactory.CreateFromWorkItemDefinition(), agent lifecycle FSM (created→running→completed), agent→issue linking, workflow orchestrator that monitors issue events and spawns agents | 📋 Not Started | P0 | Medium | Go, ArangoDB, Backend Dev | MVP-031 | [MVP-032.md](mvp-details/MVP-032.md) |

**Architecture Reference**: `/documents/3-SofwareDevelopment/mvp-details/work-items-integration/`

**Key Documents**:
- [Git-Based Document System](mvp-details/work-items-integration/git-based-document-system.md) - Complete Git implementation in ArangoDB
- [Kanban Workflow](mvp-details/work-items-integration/kanban-workflow.md) - End-to-end workflow from issue creation to completion
- [Deliverables Structure](mvp-details/work-items-integration/deliverables-structure-research.md) - Enhanced deliverable architecture with folder trees and AI prompt instructions
- [Architectural Decisions](mvp-details/work-items-integration/architecture/vcs-integration-decisions.md) - Why internal Git vs external VCS

---

## P1: Agent Lifecycle & Runtime (CRITICAL)

*Production-ready agent deployment and operation*

| Task ID | Title | Description | Status | Priority | Effort | Skills | Dependencies | Details |
|---------|-------|-------------|--------|----------|--------|--------|--------------|---------|
| MVP-033 | Agent Lifecycle FSM | Implement agent lifecycle states (Registered, Scheduled, Starting, Healthy, Degraded, Backoff, Draining, Quarantined, Stopped, Retired) with transitions, guards, and health probes | 📋 Not Started | P1 | High | Go, Backend Dev, Health Checks | MVP-032 | [MVP-033.md](mvp-details/MVP-033.md) |
| MVP-034 | Run Execution FSM | Implement run states (Pending, Running, Waiting I/O, Waiting HITL, Succeeded, Failed, Compensating, Compensated, Orphaned) with retry/backoff logic | 📋 Not Started | P1 | High | Go, Backend Dev, State Machine | MVP-033 | [MVP-034.md](mvp-details/MVP-034.md) |
| MVP-035 | Health & Circuit Breakers | Implement health probe framework (HTTP, TCP, exec, gRPC), circuit breaker integration, and degradation detection | 📋 Not Started | P1 | Medium | Go, Backend Dev, Monitoring | MVP-034 | [MVP-035.md](mvp-details/MVP-035.md) |
| MVP-036 | Quarantine System | Implement quarantine triggers, evidence capture, triage workflow, and re-enablement approval process | 📋 Not Started | P1 | Medium | Go, Security, Backend Dev | MVP-035 | [MVP-036.md](mvp-details/MVP-036.md) |

---

## P1: Platform Infrastructure (CRITICAL)

*Deployment, monitoring, and platform capabilities*

| Task ID | Title | Description | Status | Priority | Effort | Skills | Dependencies | Details |
|---------|-------|-------------|--------|----------|--------|--------|--------------|---------|
| MVP-014 | Kubernetes Deployment | Create Kubernetes manifests and Helm charts for agent deployment | 📋 Not Started | P1 | High | DevOps, Kubernetes | MVP-010 ✅ | - |
| MVP-015 | Management Dashboard | Build web interface with Templ+HTMX+Alpine.js for agent monitoring, real-time updates, and control | 🚀 In Progress | P1 | Medium | Go, Frontend Dev, Templ | MVP-013 ✅ | - |
| MVP-023 | AI Agent Creator | Implement AI-powered conversational interface for creating agents. AI asks questions, resolves details, and generates complete agent configurations through natural language dialogue | 📋 Not Started | P1 | Medium | Go, Templ, AI/LLM, Frontend Dev | MVP-025 ✅ | [MVP-023.md](mvp-details/MVP-023.md) |
| MVP-037 | Deployment Rollouts | Implement blue-green, canary, and progressive delivery strategies with SLO-based rollback | 📋 Not Started | P1 | High | Go, DevOps, Deployment | MVP-036 | [MVP-037.md](mvp-details/MVP-037.md) |
| MVP-038 | Namespace Isolation | Implement namespace hierarchy, resource quotas, network policies, and noisy neighbor protections | 📋 Not Started | P1 | High | Go, Kubernetes, Networking | MVP-037 | [MVP-038.md](mvp-details/MVP-038.md) |
| MVP-039 | Organization & RBAC | Build org/BU/project hierarchy, role matrix, permission system, and approval chain engine | 📋 Not Started | P1 | High | Go, Security, Backend Dev | MVP-038 | [MVP-039.md](mvp-details/MVP-039.md) |
| MVP-040 | Billing & Metering | Implement metering for all billing dimensions (agent-hours, storage, network, audit), cost allocation, and budget tracking | 📋 Not Started | P1 | Medium | Go, Backend Dev, Analytics | MVP-039 | [MVP-040.md](mvp-details/MVP-040.md) |

---

## P1: Real-Time Property Broadcasting (CRITICAL)

*Enables UC-TRACK-001 (Safiri Salama) and other real-time tracking/monitoring use cases*

| Task ID | Title | Description | Status | Priority | Effort | Skills | Dependencies | Details |
|---------|-------|-------------|--------|----------|--------|--------|--------------|---------|
| MVP-016 | Core Broadcasting Infrastructure | Implement BroadcastConfiguration, PropertyBroadcaster service, ContextEvaluator, and integration with PubSub | 📋 Not Started | P1 | High | Go, Backend Dev, PubSub | MVP-013 ✅ | - |
| MVP-017 | Subscription Management | Build SubscriptionManager, subscriber filtering, favorite functionality, and subscription API endpoints | 📋 Not Started | P1 | Medium | Go, Backend Dev, REST API | MVP-016 | - |
| MVP-018 | Privacy & Security Controls | Implement geofencing, property masking, permission model, audit logging, and encryption for sensitive properties | 📋 Not Started | P1 | Medium | Security, Backend Dev | MVP-017 | - |
| MVP-019 | Performance Optimization & Scale | Performance tuning, caching, load balancing for broadcasters, message broker optimization, monitoring & alerting | 📋 Not Started | P1 | Medium | Performance, DevOps | MVP-018 | [MVP-019.md](mvp-details/MVP-019.md) |
| MVP-020 | UC-TRACK-001 Integration & Testing | Implement Vehicle & Passenger agents, build mobile/web UI, SACCO management portal, end-to-end testing | 📋 Not Started | P1 | High | Full-stack, Mobile Dev | MVP-019 | [MVP-020.md](mvp-details/MVP-020.md) |

**Complete Technical Specification**: `/documents/3-SofwareDevelopment/core-systems/agent-property-broadcasting.md`

---

## P1: A2A Protocol Integration (STRATEGIC)

*Transform CodeValdCortex into the "Kubernetes of Multi-Vendor AI Agents" - enabling seamless interoperability with external A2A-compatible agents*

**SDK Integration**: Uses official `a2a-go` SDK (https://github.com/a2aproject/a2a-go) for protocol compliance and faster implementation.

| Task ID | Title | Description | Status | Priority | Effort | Skills | Dependencies | Details |
|---------|-------|-------------|--------|----------|--------|--------|--------------|---------|
| MVP-A2A-000 | a2a-go SDK Integration | Add `github.com/a2aproject/a2a-go` to project, create wrapper interfaces, implement protocol translation layer | 📋 Not Started | P0 | Low | Go, SDK Integration | None | - |
| MVP-A2A-001 | A2A Agent Card Generator | Implement A2A-compliant Agent Card generation using a2a-go types, validation, and HTTP endpoint to serve cards | 📋 Not Started | P0 | Medium | Go, Backend Dev, REST API | MVP-A2A-000 | - |
| MVP-A2A-002 | External Agent Registry | Build external agent registration API using a2a-go client, health check system (60s intervals), discovery UI | 📋 Not Started | P0 | Medium | Go, Backend Dev, Frontend Dev | MVP-A2A-001 | - |
| MVP-A2A-003 | A2A Gateway Service | Implement bidirectional gateway using a2a-go server/client for protocol translation between internal and external agents | 📋 Not Started | P0 | Medium | Go, a2a-go SDK, Backend Dev | MVP-A2A-002 | - |
| MVP-A2A-004 | Task Delegation System | Build sync/async task execution with a2a-go client, retry logic (exponential backoff), timeout enforcement, audit logging | 📋 Not Started | P0 | High | Go, Backend Dev, Orchestration | MVP-A2A-003 | - |
| MVP-A2A-005 | Enhanced Orchestration | Implement intelligent agent selection algorithm (capability, trust, cost, latency), configurable weights, A/B testing support | 📋 Not Started | P1 | High | Go, Backend Dev, Algorithms | MVP-A2A-004 | - |
| MVP-A2A-006 | Security & Compliance | Implement OAuth 2.0 using a2a-go auth helpers, JWT validation, API key management, RBAC integration, compliance audit reports | 📋 Not Started | P0 | Medium | Go, Security, OAuth 2.0 | MVP-A2A-005 | - |
| MVP-A2A-007 | Monitoring & Observability | Add Prometheus metrics, Grafana dashboards, distributed tracing (OpenTelemetry), structured logging, and alerting rules | 📋 Not Started | P1 | Medium | Go, Prometheus, Grafana, DevOps | MVP-A2A-006 | - |
| MVP-A2A-008 | Performance Optimization | Implement connection pooling, Redis caching for agent cards, serialization optimization, load testing, and tuning | 📋 Not Started | P1 | Medium | Go, Performance, Redis | MVP-A2A-007 | - |
| MVP-A2A-009 | Documentation & Developer Experience | Create OpenAPI/Swagger specs, integration guide, sample implementations (Python, Node.js), troubleshooting playbook, video tutorials | 📋 Not Started | P1 | Medium | Technical Writing, Developer Relations | MVP-A2A-008 | - |

**Strategic Value**: 
- 40% reduction in custom integration costs (60% with SDK)
- 60% faster time-to-value for new capabilities
- 3x expansion of addressable agent ecosystem
- Aligns with Linux Foundation open standards
- **SDK Benefits**: Protocol compliance guaranteed, upstream security updates, reduced maintenance burden

**Complete Technical Specification**: [`/documents/2-SoftwareDesignAndArchitecture/a2a-protocol-integration.md`](../../2-SoftwareDesignAndArchitecture/a2a-protocol-integration.md)

---

## P0: Authentication & User Management (CRITICAL - FRONTEND PREREQUISITE)

*Core authentication endpoints required by Flutter frontend (CodeValdFortex)*

| Task ID | Title | Description | Status | Priority | Effort | Skills | Dependencies | Details |
|---------|-------|-------------|--------|----------|--------|--------|--------------|---------|
| MVP-AUTH-005 | Protected Routes Integration | Apply authentication middleware to protected routes, update handlers to use real user context (replace "system" with actual user_id from context), add permission checks for agency/instance operations | 📋 Not Started | P0 | Low | Go, Backend Dev | ~~MVP-AUTH-004~~ ✅ | [authentication.md](mvp-details/authentication.md) |

**Frontend Integration**: CodeValdFortex Flutter app (MVP-FL-009, MVP-FL-010, MVP-FL-011) depends on these endpoints

**API Endpoints Summary**:
- `POST /api/v1/auth/register` - Create new user account
- `POST /api/v1/auth/login` - Login with email/password, returns JWT tokens
- `POST /api/v1/auth/refresh` - Refresh access token using refresh token
- `POST /api/v1/auth/logout` - Invalidate tokens
- `GET /api/v1/auth/me` - Get current authenticated user info

---

## P2: Security & Authentication (IMPORTANT)

*Advanced security features and hardening*

| Task ID | Title | Description | Status | Priority | Effort | Skills | Dependencies | Details |
|---------|-------|-------------|--------|----------|--------|--------|--------------|---------|
| MVP-026 | Password Reset & Email Verification | Implement forgot password flow, email verification on registration, password reset tokens, email service integration | 📋 Not Started | P2 | Medium | Backend Dev, Email, Security | MVP-AUTH-005 | - |
| MVP-027 | Security Implementation | Add input validation, HTTPS enforcement, security headers (CORS, CSP, HSTS), rate limiting on auth endpoints, brute force protection | 📋 Not Started | P2 | Medium | Security, Backend Dev | MVP-026 | - |
| MVP-028 | Access Control System | Implement role-based access control (RBAC) for agent operations, permission matrix, admin/user/viewer roles, resource-level permissions | 📋 Not Started | P2 | Low | Backend Dev, Security | MVP-027 | - |
| MVP-041 | Multi-tenancy Hardening | Add advanced isolation (dedicated nodes, encryption), data residency controls, and compliance reporting | 📋 Not Started | P2 | Medium | Go, Security, Compliance | MVP-040 | [MVP-041.md](mvp-details/MVP-041.md) |

---

## P2: Agent-Based Compliance Frameworks (FUTURE ENHANCEMENT)

*Intelligent compliance agents that provide context-aware, dynamic enforcement vs static configuration*

| Task ID | Title | Description | Status | Priority | Effort | Skills | Dependencies | Details |
|---------|-------|-------------|--------|----------|--------|--------|--------------|---------|
| MVP-051 | Compliance Agent Architecture | Design and implement base ComplianceAgent interface, knowledge base system, context analyzer, reasoning engine, and agent-to-workflow bridge service | 📋 Not Started | P2 | High | Go, AI/LLM, Security, Architecture | MVP-050 | [MVP-051.md](mvp-details/MVP-051-Compliance-Agent-Architecture.md) |
| MVP-052-CMP | GDPR Compliance Agent | Implement GDPR-specific agent with article knowledge base, lawful basis engine, jurisdiction analyzer, cross-border transfer rules, and explainable compliance plan generation | 📋 Not Started | P2 | High | Go, AI/LLM, GDPR Expertise, Security | MVP-051 | - |
| MVP-053 | Multi-Framework Compliance System | Implement SOC2Agent (trust service criteria), HIPAAAgent (PHI classification, safeguards), ISO27001Agent (risk assessment, controls), and unified compliance dashboard | 📋 Not Started | P2 | High | Go, AI/LLM, Compliance Expertise | MVP-052-CMP | - |

**Note**: MVP-052 is already used for Workflow Visual Designer (completed). Compliance GDPR task renamed to MVP-052-CMP to avoid conflict.

**Key Benefits**:
- ✅ **Context-aware**: Different compliance steps based on data type, jurisdiction, use case
- ✅ **Explainable**: AI provides reasoning for each requirement (e.g., "satisfies GDPR Article 6(1)(a)")
- ✅ **Adaptive**: Automatically updates when regulations change
- ✅ **Intelligent**: Understands lawful basis, data sensitivity, risk levels
- ✅ **Auditable**: Complete reasoning trail for compliance officers

**Architecture**: Compliance agents analyze context → generate compliance plans → bridge converts plans to executable workflows → workflow designer executes and monitors

---

## P1: React Migration - Frontend Modernization

*Strategic migration from Templ+HTMX to React SPA for better separation of concerns and scalability*

### Phase 1: Foundation & Work Items PoC (Weeks 1-5)

| Task ID | Title | Description | Status | Priority | Effort | Skills | Dependencies | Details |
|---------|-------|-------------|--------|----------|--------|--------|--------------|---------|
| MVP-RM-003 | Work Items REST API | Implement 5 REST endpoints for work items (list with pagination/filtering, get single, create, update, delete). Enforce agency-specific data isolation | 📋 Not Started | P1 | Medium | Go, Gin, ArangoDB | ~~MVP-RM-002~~ ✅ | [react-migration/MVP-RM-003_work_items_api.md](mvp-details/react-migration/MVP-RM-003_work_items_api.md) |
| MVP-RM-004 | Work Items Redux Store | Create Redux slice with async thunks for fetching, creating, updating, deleting work items. Implement filtering, pagination state, error handling | 📋 Not Started | P1 | Medium | React, Redux Toolkit, TypeScript | ~~MVP-RM-001~~ ✅, MVP-RM-003 | [react-migration/](mvp-details/react-migration/) |
| MVP-RM-005 | Work Items UI Components | Build React components using Bulma CSS: WorkItemList, WorkItemCard, WorkItemForm, WorkItemFilters. Implement CRUD operations, form validation, loading states | 📋 Not Started | P1 | High | React, TypeScript, Bulma CSS | MVP-RM-004 | [react-migration/](mvp-details/react-migration/) |
| MVP-RM-006 | Testing Suite | Write unit tests (Vitest), component tests (React Testing Library), integration tests for work items feature. Achieve >80% coverage | 📋 Not Started | P1 | Medium | Vitest, React Testing Library, Go testing | MVP-RM-005 | [react-migration/](mvp-details/react-migration/) |
| MVP-RM-007 | Deployment Pipeline | Set up staging/production environments, CI/CD pipeline (GitHub Actions), Docker containerization, monitoring, rollback procedures | 📋 Not Started | P1 | Medium | Docker, CI/CD, Nginx | MVP-RM-006 | [react-migration/](mvp-details/react-migration/) |

**Architecture Reference**: [react-migration/README.md](mvp-details/react-migration/README.md)  
**Migration Plan**: `/documents/2-SoftwareDesignAndArchitecture/react-migration-plan.md`  
**Tasks Summary**: [react-migration/TASKS_SUMMARY.md](mvp-details/react-migration/TASKS_SUMMARY.md)

**Key Goals**:
- Prove React + Go API architecture works
- Establish clear separation: Backend = business logic, Frontend = presentation
- Create patterns for future domain migrations
- Side-by-side operation with existing Templ pages during rollout

---

## P2: UI Migration & Cleanup (FUTURE)

*Remove Templ/HTMX UI code after Flutter frontend (CodeValdFortex) reaches production. Transform Cortex into pure REST API backend.*

**Prerequisites**: 
- CodeValdFortex must reach production with feature parity
- All UI features must be validated in Flutter app
- API contracts must be stable and documented

| Task ID | Title | Description | Status | Priority | Effort | Skills | Dependencies | Details |
|---------|-------|-------------|--------|----------|--------|--------|--------------|---------|
| MVP-CLEANUP-001 | Remove Agency Selection UI | Remove Templ templates/HTMX from MVP-022 (Agency Selection Homepage). Keep REST API endpoints: GET /api/v1/agencies (listing), GET /api/v1/agencies/:id/status | 📋 Not Started | P2 | Low | Go, Refactoring | Fortex MVP-FL-101 in production | - |
| MVP-CLEANUP-002 | Remove Create Agency UI | Remove Templ templates/HTMX from MVP-024 (Create Agency Form). Keep POST /api/v1/agencies endpoint with JSON body | 📋 Not Started | P2 | Low | Go, Refactoring | Fortex MVP-FL-102 in production | - |
| MVP-CLEANUP-003 | Remove Introduction UI | Remove Templ templates from MVP-025 (AI Agency Designer introduction section). Keep AI generation API endpoints, text processing logic | 📋 Not Started | P2 | Low | Go, Refactoring | Fortex MVP-FL-104 in production | - |
| MVP-CLEANUP-004 | Remove Goals UI | Remove Templ templates/HTMX from MVP-029 (Goals Module). Keep CRUD API endpoints (GET/POST/PUT/DELETE /api/v1/agencies/:id/goals), AI generation endpoints | 📋 Not Started | P2 | Low | Go, Refactoring | Fortex MVP-FL-105 in production | - |
| MVP-CLEANUP-005 | Remove Work Items UI | Remove Templ templates from MVP-030/043/054. Keep Work Items CRUD API, deliverables tree API, AI generation endpoints. Preserve business logic in service layer | 📋 Not Started | P2 | Medium | Go, Refactoring | Fortex MVP-FL-106 in production | - |
| MVP-CLEANUP-006 | Remove Roles UI | Remove Templ templates/HTMX from MVP-044 (Roles UI Module). Keep Roles CRUD API, autonomy level validation, tags system, AI generation endpoints | 📋 Not Started | P2 | Low | Go, Refactoring | Fortex MVP-FL-107 in production | - |
| MVP-CLEANUP-007 | Remove RACI Matrix UI | Remove Templ templates from MVP-045 (RACI Matrix UI Editor). Keep RACI assignment API endpoints (GET/POST /api/v1/agencies/:id/raci), grid data serialization | 📋 Not Started | P2 | Low | Go, Refactoring | Fortex MVP-FL-108 in production | - |
| MVP-CLEANUP-008 | Remove Workflows UI | Remove Templ templates/Alpine.js from MVP-051/052 (Workflow Visual Designer). Keep Workflow CRUD API, visual designer state API (nodes/connections), validation logic | 📋 Not Started | P2 | Medium | Go, Refactoring | Fortex MVP-FL-109 in production | - |
| MVP-CLEANUP-009 | Remove AI Policy UI | Remove Templ templates from MVP-048 (AI Policy Layer). Keep Policy CRUD API, validation endpoints, compliance framework logic | 📋 Not Started | P2 | Low | Go, Refactoring | Fortex MVP-FL-110 in production | - |
| MVP-CLEANUP-010 | Remove Publishing UI | Remove Templ templates/HTMX from MVP-PUB-005 (Publishing UI). Keep Publishing API endpoints (POST /api/v1/agencies/:id/validate, /publish, /activate, /tags). Preserve state machine logic | 📋 Not Started | P2 | Low | Go, Refactoring | Fortex MVP-FL-112-115 in production | - |
| MVP-CLEANUP-011 | Remove Instance Management UI | Remove Templ templates from MVP-PUB-007 (Instance Management). Keep Instance API endpoints (POST/GET/DELETE /api/v1/agencies/:id/instances, lifecycle controls). Preserve multi-database orchestration | 📋 Not Started | P2 | Medium | Go, Refactoring | Fortex MVP-FL-117-119 in production | - |
| MVP-CLEANUP-012 | Remove Kanban/Workbench UI | Remove Templ templates/HTMX from MVP-WI-008 (Workbench Kanban). Keep Issue CRUD API, board state API, workflow integration. Preserve issue lifecycle logic | 📋 Not Started | P2 | Medium | Go, Refactoring | Fortex MVP-FL-120-121 in production | - |
| MVP-CLEANUP-013 | Remove File Explorer UI | Remove Templ templates from MVP-WI-006 (File Explorer). Keep File/folder CRUD API endpoints (GET/POST/PUT/DELETE /api/v1/files), Git operations API | 📋 Not Started | P2 | Low | Go, Refactoring | Fortex MVP-FL-122-123 in production | - |
| MVP-CLEANUP-014 | API Documentation & Cleanup | Generate OpenAPI/Swagger specifications for all remaining REST endpoints. Remove unused handler functions, cleanup routing, remove static file serving (CSS/JS), update documentation to reflect API-only architecture | 📋 Not Started | P2 | Medium | Go, OpenAPI, Documentation | All cleanup tasks complete | - |

**Cleanup Scope**:
- ❌ Remove: All Templ templates (`*.templ` files)
- ❌ Remove: HTMX attributes and JavaScript
- ❌ Remove: Alpine.js components
- ❌ Remove: Static CSS/SCSS files for UI
- ❌ Remove: Frontend JavaScript files
- ❌ Remove: HTML rendering handlers
- ✅ Keep: All REST API endpoints
- ✅ Keep: Go service layer (business logic)
- ✅ Keep: ArangoDB repositories
- ✅ Keep: WebSocket/SSE for real-time updates
- ✅ Keep: AI integration logic
- ✅ Keep: State machines and validators

**Post-Cleanup Architecture**:
- Pure REST API backend (Gin framework)
- JSON-only request/response (no HTML rendering)
- OpenAPI documented endpoints
- WebSocket/SSE for real-time features
- Improved performance (no template compilation)
- Smaller binary size
- Simplified deployment (no static assets)

**Total P2 Cleanup**: 14 tasks

---

## Bugs and Issues

### Resolved Bugs

| Bug ID | Title | Description | Affected Area | Priority | Resolution |
|--------|-------|-------------|---------------|----------|------------|
| BUG-001 | Instance list not displaying when navigating from versions page | When clicking "View Instances" from tag dropdown on versions page, navigation to `/agencies/{id}/instances?tag_key={tagID}` occurs but no instances were displayed. Root cause: (1) Instances in database had empty `tag_id` field (created before field validation), (2) Tab switching code used wrong tab name ('table' instead of 'all-instances'). | Instance Management UI | P0 | ✅ Fixed - Updated existing instance `tag_id` manually, fixed tab name in instances.js, verified filter logic working correctly |

### Active Bugs

_(None)_

---

## Deprecated / Removed Tasks

The following tasks are no longer applicable due to architectural changes:

| Task ID | Original Title | Reason | Replaced By |
|---------|---------------|--------|-------------|
| ~~MVP-WI-001~~ | ~~Gitea Webhook Integration~~ | ❌ **Removed** - External VCS integration not implemented | MVP-WI-005 through MVP-WI-011 (Git-in-ArangoDB) |
| ~~MVP-WI-002~~ | ~~Gitea API Client~~ | ❌ **Removed** - No external VCS integration | MVP-WI-006 (File Explorer API) |
| ~~MVP-WI-003~~ | ~~Bidirectional Sync~~ | ❌ **Removed** - No external system synchronization | MVP-WI-009 (Issue-Git Integration) |
| ~~MVP-WI-004~~ | ~~PR Automation for External VCS~~ | ❌ **Removed** - Internal PR system instead | MVP-WI-007 (Pull Request System) |

**Architectural Decision**: After research, decided to implement Git directly in ArangoDB rather than integrate with external VCS (Gitea/GitHub/GitLab). See [vcs-integration-decisions.md](mvp-details/work-items-integration/architecture/vcs-integration-decisions.md) for rationale.

---

## Task Summary by Priority

### P0 (Blocking - Must Complete First)
- **Authentication & User Management**: 5 tasks (MVP-AUTH-001 through MVP-AUTH-005)
- **Agency Designer**: 5 tasks (MVP-046, MVP-047, MVP-049, MVP-050, MVP-042)
- **Work Items & Document Management**: 5 tasks (MVP-WI-007 through MVP-WI-010, ~~MVP-030~~ ✅)
- **Work Item Definitions**: 2 tasks (MVP-031, MVP-032)
- **A2A Protocol**: 6 foundational tasks (MVP-A2A-000 through MVP-A2A-004, MVP-A2A-006)

**Total P0**: 23 tasks

### P1 (Critical - Core Features)
- **Git AI Features**: 1 task (MVP-WI-011 - AI-Assisted Conflict Resolution)
- **Agent Lifecycle**: 4 tasks (MVP-033 through MVP-036)
- **Platform Infrastructure**: 7 tasks (MVP-014, MVP-015, MVP-023, MVP-037 through MVP-040)
- **Property Broadcasting**: 5 tasks (MVP-016 through MVP-020)
- **A2A Advanced**: 4 tasks (MVP-A2A-005, MVP-A2A-007, MVP-A2A-008, MVP-A2A-009)

**Total P1**: 21 tasks

### P2 (Important - Quality & Enhancement)
- **Advanced Security**: 4 tasks (MVP-026 through MVP-028, MVP-041)
- **Compliance Agents**: 3 tasks (MVP-051, MVP-052-CMP, MVP-053)
- **UI Migration & Cleanup**: 14 tasks (MVP-CLEANUP-001 through MVP-CLEANUP-014)

**Total P2**: 21 tasks

**Grand Total Active Tasks**: 66 tasks

---

**Note**: This document contains only active and pending tasks. All completed tasks are moved to `mvp_done.md` to maintain a clean, actionable backlog.

Follow This sequence!!!

Phase 1 (P0 Core):

1. ~~MVP-WI-005 - Git Core Layer~~ ✅
2. ~~MVP-WI-006 - File Explorer~~ ✅
3. ~~MVP-030 - Work Item Definitions~~ ✅
4. ~~MVP-WI-008 - Kanban Board~~ ✅
5. MVP-WI-007 - Pull Requests
6. MVP-031 - Work Item Lifecycle
7. MVP-WI-009 - Issue-Git Integration
8. MVP-032 - Agent Factory
9. MVP-WI-010 - Agent Assignment

Phase 2 (P1 Enhancement):
10. MVP-WI-011 - AI Conflict Resolution