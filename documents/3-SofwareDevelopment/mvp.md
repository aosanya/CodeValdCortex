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

## Foundation Tasks (P0 - Blocking)

*All foundation tasks completed. See `mvp_done.md` for details.*

## Core Agent Mechanics (P0 - Blocking)

*All core agent mechanics tasks completed. See `mvp_done.md` for details.*

## Core Functionality Tasks (P1 - Critical)

*All core functionality tasks completed. See `mvp_done.md` for details.*

## Platform Integration Tasks (P1 - Critical)

| Task ID | Title                 | Description                                                                        | Status      | Priority | Effort | Skills Required         | Dependencies | Details |
| ------- | --------------------- | ---------------------------------------------------------------------------------- | ----------- | -------- | ------ | ----------------------- | ------------ | ------- |
| MVP-014 | Kubernetes Deployment | Create Kubernetes manifests and Helm charts for agent deployment                   | Not Started | P1       | High   | DevOps, Kubernetes      | MVP-010      |         |
| MVP-015 | Management Dashboard  | Build web interface with Templ+HTMX+Alpine.js for agent monitoring, real-time updates, and control | In Progress | P1       | Medium | Go, Frontend Dev, Templ | MVP-013      |         |
| MVP-023 | AI Agent Creator      | Implement AI-powered conversational interface for creating agents. AI asks questions, resolves details, and generates complete agent configurations through natural language dialogue | Not Started | P1       | Medium | Go, Templ, AI/LLM, Frontend Dev | MVP-025      | [MVP-023.md](mvp-details/MVP-023.md) |
| MVP-030 | Work Items Core Schema & Registry | Implement work item types registry, JSON schemas, and extend roles with taxonomy fields (autonomy, budget, safety, identity). **Architecture**: See [Work Items Documentation](../../2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items/) for complete GitOps + ArangoDB multi-graph design | Not Started | P1       | Medium | Go, ArangoDB, JSON Schema | MVP-029      | [MVP-030.md](mvp-details/MVP-030.md) |
| MVP-031 | Work Items Lifecycle & SLA | Implement state machine, timers, breach handlers, and SLA/SLO enforcement for work items | Not Started | P1       | Medium | Go, ArangoDB, Backend Dev | MVP-030      | [MVP-031.md](mvp-details/MVP-031.md) |
| MVP-032 | Work Items Assignment & Routing | Build declarative routing rules engine, skill matching, and agent selection algorithms | Not Started | P1       | Medium | Go, ArangoDB, Backend Dev | MVP-031      | [MVP-032.md](mvp-details/MVP-032.md) |
| MVP-033 | Agent Lifecycle FSM | Implement agent lifecycle states (Registered, Scheduled, Starting, Healthy, Degraded, Backoff, Draining, Quarantined, Stopped, Retired) with transitions, guards, and health probes | Not Started | P1       | High   | Go, Backend Dev, Health Checks | MVP-032      | [MVP-033.md](mvp-details/MVP-033.md) |
| MVP-034 | Run Execution FSM | Implement run states (Pending, Running, Waiting I/O, Waiting HITL, Succeeded, Failed, Compensating, Compensated, Orphaned) with retry/backoff logic | Not Started | P1       | High   | Go, Backend Dev, State Machine | MVP-033      | [MVP-034.md](mvp-details/MVP-034.md) |
| MVP-035 | Health & Circuit Breakers | Implement health probe framework (HTTP, TCP, exec, gRPC), circuit breaker integration, and degradation detection | Not Started | P1       | Medium | Go, Backend Dev, Monitoring | MVP-034      | [MVP-035.md](mvp-details/MVP-035.md) |
| MVP-036 | Quarantine System | Implement quarantine triggers, evidence capture, triage workflow, and re-enablement approval process | Not Started | P1       | Medium | Go, Security, Backend Dev | MVP-035      | [MVP-036.md](mvp-details/MVP-036.md) |
| MVP-037 | Deployment Rollouts | Implement blue-green, canary, and progressive delivery strategies with SLO-based rollback | Not Started | P1       | High   | Go, DevOps, Deployment | MVP-036      | [MVP-037.md](mvp-details/MVP-037.md) |
| MVP-038 | Namespace Isolation | Implement namespace hierarchy, resource quotas, network policies, and noisy neighbor protections | Not Started | P1       | High   | Go, Kubernetes, Networking | MVP-037      | [MVP-038.md](mvp-details/MVP-038.md) |
| MVP-039 | Organization & RBAC | Build org/BU/project hierarchy, role matrix, permission system, and approval chain engine | Not Started | P1       | High   | Go, Security, Backend Dev | MVP-038      | [MVP-039.md](mvp-details/MVP-039.md) |
| MVP-040 | Billing & Metering | Implement metering for all billing dimensions (agent-hours, storage, network, audit), cost allocation, and budget tracking | Not Started | P1       | Medium | Go, Backend Dev, Analytics | MVP-039      | [MVP-040.md](mvp-details/MVP-040.md) |
| MVP-041 | Multi-tenancy Hardening | Add advanced isolation (dedicated nodes, encryption), data residency controls, and compliance reporting | Not Started | P2       | Medium | Go, Security, Compliance | MVP-040      | [MVP-041.md](mvp-details/MVP-041.md) |

## Authentication & Security Tasks (P2 - Important)

| Task ID | Title                     | Description                                                | Status      | Priority | Effort | Skills Required       | Dependencies |
| ------- | ------------------------- | ---------------------------------------------------------- | ----------- | -------- | ------ | --------------------- | ------------ |
| MVP-026 | Basic User Authentication | Implement user registration, login, and session management | Not Started | P2       | Medium | Backend Dev, Security | MVP-014      |
| MVP-027 | Security Implementation   | Add input validation, HTTPS, and basic security headers    | Not Started | P2       | Medium | Security, Backend Dev | MVP-026      |
| MVP-028 | Access Control System     | Implement role-based access control for agent operations   | Not Started | P2       | Low    | Backend Dev, Security | MVP-027      |

## Agency Designer (P1 - Critical)

| Task ID | Title                     | Description                                                | Status      | Priority | Effort | Skills Required       | Dependencies | Details |
| ------- | ------------------------- | ---------------------------------------------------------- | ----------- | -------- | ------ | --------------------- | ------------ | ------- |
| MVP-046 | Agency Admin & Configuration Page | Build comprehensive admin interface for agency-wide settings: token budgets (role & individual agent levels), rate limits, resource quotas, AI model selection, cost controls, monitoring dashboards, and operational parameters | Not Started | P1       | Medium | Go, Templ, Frontend Dev, Analytics | MVP-044      | [MVP-046.md](mvp-details/MVP-046.md) |
| MVP-047 | Agency Designer Export System | Implement comprehensive export functionality for entire agency design (all sections) to PDF, Markdown, and JSON formats with customizable templates and branding | Not Started | P1       | Medium | Go, PDF Generation, File Export | MVP-044      | [MVP-047.md](mvp-details/MVP-047.md) |
| MVP-048 | AI Policy Layer - Foundation | Implement organizational AI governance: first-run policy wizard (stance, model approval, autonomy levels, data classification), policy schema, basic enforcement engine, and UI indicators. Addresses DORA requirement for clear AI stance and runtime guardrails | Not Started | P0       | High   | Go, Security, Backend Dev, Templ | MVP-044      | [AI Policy Layer](../../2-SoftwareDesignAndArchitecture/ai-policy-layer.md) |
| MVP-049 | AI Policy Layer - Runtime Enforcement | Build action authorization, approval workflows, risk scoring, budget tracking, and policy violation handling with real-time feedback and audit logging | Not Started | P1       | High   | Go, Security, Backend Dev | MVP-048      | [AI Policy Layer](../../2-SoftwareDesignAndArchitecture/ai-policy-layer.md) |
| MVP-050 | AI Policy Layer - Advanced Features | Implement data classification engine, PII detection/masking, compliance reporting, policy versioning, and multi-policy inheritance | Not Started | P1       | Medium | Go, Security, ML, Backend Dev | MVP-049      | [AI Policy Layer](../../2-SoftwareDesignAndArchitecture/ai-policy-layer.md) |
| MVP-052 | Workflow Visual Designer | Build drag-and-drop workflow designer using xyflow (vanilla JS) to visually connect and orchestrate work items. Features: node editor, connections/dependencies, conditional routing, parallel paths, validation, save/load workflows as JSON, execution visualization. **Architecture**: See [Work Items Documentation](../../2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items/) for GitOps workflow integration | Not Started | P1       | High   | Go, Templ, Frontend Dev (xyflow/Alpine.js) | MVP-051 (Complete)      | [MVP-052.md](mvp-details/MVP-052.md) |
| MVP-042 | AI-Powered Agency Creator | Implement AI-driven agency creation flow with text upload, selective generation (introduction, goals, work items, roles, RACI), and batch AI generation | Not Started | P1       | High   | Go, Templ, AI/LLM, Frontend Dev | MVP-047      | [MVP-042.md](mvp-details/MVP-042.md) |

## Agent Property Broadcasting Feature (P1 - Critical)

*Enables UC-TRACK-001 (Safiri Salama) and other real-time tracking/monitoring use cases*

| Task ID | Title                                    | Description                                                                                                      | Status      | Priority | Effort | Skills Required            | Dependencies |
| ------- | ---------------------------------------- | ---------------------------------------------------------------------------------------------------------------- | ----------- | -------- | ------ | -------------------------- | ------------ |
| MVP-016 | Core Broadcasting Infrastructure         | Implement BroadcastConfiguration, PropertyBroadcaster service, ContextEvaluator, and integration with PubSub    | Not Started | P1       | High   | Go, Backend Dev, PubSub    | MVP-013      |
| MVP-017 | Subscription Management                  | Build SubscriptionManager, subscriber filtering, favorite functionality, and subscription API endpoints          | Not Started | P1       | Medium | Go, Backend Dev, REST API  | MVP-016      |
| MVP-018 | Privacy & Security Controls              | Implement geofencing, property masking, permission model, audit logging, and encryption for sensitive properties | Not Started | P1       | Medium | Security, Backend Dev      | MVP-017      |
| MVP-019 | Performance Optimization & Scale         | Performance tuning, caching, load balancing for broadcasters, message broker optimization, monitoring & alerting | Not Started | P1       | Medium | Performance, DevOps        | MVP-018      | [📄](mvp-details/MVP-019.md) |
| MVP-020 | UC-TRACK-001 Integration & Testing       | Implement Vehicle & Passenger agents, build mobile/web UI, SACCO management portal, end-to-end testing          | Not Started | P1       | High   | Full-stack, Mobile Dev     | MVP-019      | [📄](mvp-details/MVP-020.md) |

**Note**: Complete technical specifications available in `/documents/3-SofwareDevelopment/core-systems/agent-property-broadcasting.md`

---

**Note**: This document contains only active and pending tasks. All completed tasks are moved to `mvp_done.md` to maintain a clean, actionable backlog.