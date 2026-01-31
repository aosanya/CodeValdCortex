# MVP - UC-INFRA-001 Water Distribution Network Showcase

**Last Updated**: October 25, 2025  
**Current Status**: Phase 4 Complete (100%), Phase 6 In Progress (33%)

## Task Overview
- **Objective**: Demonstrate CodeValdCortex framework capabilities using the Water Distribution Network use case
- **Success Criteria**: Functional agent-based system that showcases autonomous infrastructure monitoring, agent collaboration, and real-time coordination
- **Approach**: Configuration-based roles + demonstration scenarios leveraging framework's messaging and coordination
- **Focus**: Demonstrate the design documented in `/documents/2-SoftwareDesignAndArchitecture/Usecases/UC-INFRA-001-water-distribution-network/`

## Current State Summary

### ✅ What's Working Now
1. **Role System**: 5 water infrastructure roles defined and loaded (pipe, sensor, pump, valve, zone coordinator) - See INFRA-001 to INFRA-005 in mvp_done.md
2. **Framework Core**: Complete agent runtime, messaging (direct + pub/sub), state persistence, REST API, Web UI - See INFRA-006, INFRA-012, INFRA-014, INFRA-016 in mvp_done.md
3. **Agent Instances**: 293 agent instances representing comprehensive Nairobi water distribution network - See INFRA-007, INFRA-008, INFRA-017 in mvp_done.md
4. **Demonstration Scenarios**: Leak detection, pressure optimization, and predictive maintenance implemented - See INFRA-009, INFRA-010, INFRA-011 in mvp_done.md
5. **Web Interface**: Bulma CSS-styled dashboard with agent management and geographic visualizer
6. **Network Topology Visualizer**: Deck.gl + MapLibre GL geographic visualization showing 293 agents on Nairobi map - See INFRA-017 in mvp_done.md

### 🎯 What's Next
1. **Data Storage**: Implement time-series storage for sensor readings and metrics (INFRA-013)
2. **Analytics**: Add water infrastructure-specific analytics queries and reports (INFRA-015)
3. **Advanced UI**: Alert management and performance metrics dashboards (INFRA-018, INFRA-019)

### 📊 Progress Metrics
- **Framework Foundation**: ✅ 100% Complete
- **Role Configuration**: ✅ 100% Complete (5/5 types)
- **Instance Creation**: ✅ 100% Complete (293/293 agents)
- **Scenario Implementation**: ✅ 100% Complete (3/3 scenarios)
- **Visualization & UI**: ⚠️ 67% Complete (base UI + topology visualizer done, advanced dashboards pending)
- **Overall MVP**: 44% Complete (12/27 tasks)

**Note**: Completed tasks (12 total) have been moved to mvp_done.md for detailed documentation. See that file for INFRA-001 through INFRA-017 details.

## Module Architecture

### Base Framework Module
**Location**: `/workspaces/CodeValdCortex/` (root)  
**Module**: `github.com/aosanya/CodeValdCortex`  
**Purpose**: Core agent runtime, communication system, and reusable infrastructure

**Implements**:
- Core agent runtime and lifecycle management
- ArangoDB-based message queuing and communication system (polling-based)
- Agent registry and state management
- Task scheduling system
- Configuration management
- Health monitoring
- Generic agent interfaces and base implementations

### Use Case Module (UC-INFRA-001)
**Location**: `/workspaces/CodeValdCortex/usecases/UC-INFRA-001-water-distribution-network/`  
**Module**: `github.com/aosanya/UC-INFRA-001-water-distribution-network`  
**Purpose**: Water Distribution Network specific implementation

**Imports**: `github.com/aosanya/CodeValdCortex` (base framework)

**Implements**:
- Domain-specific roles (Pipe, Sensor, Pump, Valve, Zone Coordinator)
- Water infrastructure business logic
- Leak detection scenarios
- Pressure optimization algorithms
- Predictive maintenance for water infrastructure
- Dashboard and visualization for water networks

### Dependency Flow
```
┌─────────────────────────────────────────────────────────────┐
│  UC-INFRA-001 Water Distribution Network Module            │
│  (Use Case Specific Implementation)                         │
│  Location: Usecases/UC-INFRA-001-water-distribution-network│
│                                                             │
│  - Pipe Agent (extends base agent)                         │
│  - Sensor Agent (extends base agent)                       │
│  - Pump Agent (extends base agent)                         │
│  - Water-specific scenarios                                │
│  - Infrastructure dashboard                                │
└─────────────────────────────────────────────────────────────┘
                            ↓ imports
┌─────────────────────────────────────────────────────────────┐
│  CodeValdCortex Base Framework                              │
│  (Reusable Core Library)                                    │
│  Location: /workspaces/CodeValdCortex (root)               │
│                                                             │
│  - Agent Runtime Manager                                   │
│  - ArangoDB Communication System (INFRA-006)               │
│  - Agent Registry                                          │
│  - State Persistence (INFRA-010, INFRA-012)                │
│  - Task Scheduling                                         │
│  - Configuration Service                                   │
│  - Health Monitoring                                       │
└─────────────────────────────────────────────────────────────┘
                            ↓ uses
┌─────────────────────────────────────────────────────────────┐
│  External Dependencies                                      │
│  - ArangoDB (database)                                      │
│  - Go standard library                                      │
│  - Third-party packages (gin, etc.)                         │
│  - Templ (temporary dev UI - to be removed)                 │
└─────────────────────────────────────────────────────────────┘
```

### Task Classification by Module

**Framework Tasks (Base Module)**: Core functionality that should be in `github.com/aosanya/CodeValdCortex`
- INFRA-006: ArangoDB Message System (reusable communication layer)
- INFRA-010: ArangoDB Collections (base schema for agents)
- INFRA-011: Time-Series Data Storage (generic time-series support)
- INFRA-012: Agent State Persistence (base state management)

**Use Case Tasks (UC-INFRA-001 Module)**: Water distribution specific implementations
- INFRA-001: Pipe Agent (domain-specific)
- INFRA-002: Sensor Agent (domain-specific)
- INFRA-003: Pump Agent (domain-specific)
- INFRA-004: Valve Agent (domain-specific)
- INFRA-005: Zone Coordinator Agent (domain-specific)
- INFRA-007: Leak Detection Scenario (domain-specific)
- INFRA-008: Pressure Optimization (domain-specific)
- INFRA-009: Predictive Maintenance (domain-specific)
- INFRA-013: Historical Analytics (domain-specific queries)
- INFRA-014-017: Dashboard and UI (domain-specific visualization)
- INFRA-018-021: Advanced water infrastructure features
- INFRA-022-025: Deployment and integration

## Phase 3: Agent Runtime & Instance Management (P0 - Critical)

**Status**: ✅ Complete - All tasks moved to mvp_done.md

## Phase 4: Scenario Implementations (P1 - Critical)

**Status**: ✅ Complete - All tasks moved to mvp_done.md

## Phase 5: Data Storage & Analytics (P1 - Critical) ⚠️ PARTIALLY COMPLETE

**Status**: INFRA-012 and INFRA-014 complete (see mvp_done.md). Time-series storage patterns and analytics queries need optimization for water infrastructure metrics.

| Task ID | Title | Description | Module | Status | Priority | Effort | Skills Required | Dependencies |
|---------|-------|-------------|--------|--------|----------|--------|-----------------|--------------|
| INFRA-013 | Time-Series Data Storage | Implement efficient storage pattern for sensor readings: pressure logs, flow rates, temperature with date-based partitioning and retention policies using ArangoDB collections | **UC-INFRA-001** | Not Started | P1 | Medium | ArangoDB, AQL | INFRA-012 |
| INFRA-015 | Historical Analytics Queries | Implement AQL queries for water infrastructure analytics: trend analysis, efficiency reports, leak history, maintenance predictions | **UC-INFRA-001** | Not Started | P1 | Medium | AQL, Analytics | INFRA-013 |

**Completed Tasks**: See mvp_done.md for INFRA-012 (Collections Schema) and INFRA-014 (State Persistence)

## Phase 6: Visualization & UI (P1 - Critical) ⚠️ PARTIALLY COMPLETE

**Status**: INFRA-016 and INFRA-017 complete (see mvp_done.md). Alert management and performance metrics dashboards remain.

| Task ID | Title | Description | Module | Status | Priority | Effort | Skills Required | Dependencies |
|---------|-------|-------------|--------|--------|----------|--------|-----------------|--------------|
| INFRA-018 | Alert Management UI | Enhance framework alert system with water-specific alerts: leak detection, pressure anomalies, maintenance schedules with priority indicators | **UC-INFRA-001** | Not Started | P1 | Medium | Go, Frontend | INFRA-016, INFRA-009 |
| INFRA-019 | Performance Metrics Dashboard | Add infrastructure metrics view: flow rates, pressure trends, energy consumption, leak detection stats, agent health for zones | **UC-INFRA-001** | Not Started | P1 | Medium | Frontend, Chart.js | INFRA-016, INFRA-015 |

**Completed Tasks**: See mvp_done.md for INFRA-016 (Framework Web UI) and INFRA-017 (Network Topology Visualizer)

## Phase 7: Advanced Features & Scenarios (P2 - Enhancement)

**Status**: Advanced water infrastructure features for extended demonstrations.

| Task ID | Title | Description | Module | Status | Priority | Effort | Skills Required | Dependencies |
|---------|-------|-------------|--------|--------|----------|--------|-----------------|--------------|
| INFRA-020 | Emergency Response Coordination | Implement emergency scenario: Fire hydrant request → Zone increases pressure → Pumps coordinate → Valves reroute → Control room notified | **UC-INFRA-001** | Not Started | P2 | High | Go, Complex Events | INFRA-010 |
| INFRA-021 | Energy Optimization | Implement smart pump scheduling to minimize energy costs while maintaining service levels based on time-of-day pricing and demand forecasting | **UC-INFRA-001** | Not Started | P2 | Medium | Go, Scheduling | INFRA-010 |
| INFRA-022 | Water Quality Monitoring | Add water quality dimensions to sensor agents: contamination detection, temperature, pH levels with automatic response protocols | **UC-INFRA-001** | Not Started | P2 | Medium | Go, Analytics | INFRA-008 |
| INFRA-023 | Customer Meter Integration | Add customer meter role for consumption tracking, billing data, and customer-level leak detection | **UC-INFRA-001** | Not Started | P2 | Low | Go, APIs | INFRA-008 |

## Phase 8: Integration & Deployment (P2 - Enhancement)

**Status**: Production deployment and external system integration features.

| Task ID | Title | Description | Module | Status | Priority | Effort | Skills Required | Dependencies |
|---------|-------|-------------|--------|--------|----------|--------|-----------------|--------------|
| INFRA-024 | Docker Compose Setup | Enhance docker-compose.yml for UC-INFRA-001 with ArangoDB, monitoring stack, and proper networking configuration | **UC-INFRA-001** | Not Started | P2 | Low | DevOps, Docker | INFRA-007 |
| INFRA-025 | IoT Gateway Integration | Implement MQTT/Modbus/OPC UA protocol adapters for connecting real physical sensors to agent system | **UC-INFRA-001** | Not Started | P2 | High | IoT, Protocols | INFRA-008 |
| INFRA-026 | SCADA System Integration | Build integration layer with existing SCADA systems for bi-directional data exchange and control commands | **UC-INFRA-001** | Not Started | P2 | High | SCADA, APIs | INFRA-010 |
| INFRA-027 | GIS System Integration | Integrate with Geographic Information Systems for spatial data, mapping infrastructure assets, and location-based analysis | **UC-INFRA-001** | Not Started | P2 | Medium | GIS, APIs | INFRA-017 |

## Showcase Deliverables

### Priority Demo Scenarios

1. **Role Management** ✅ (INFRA-001 to INFRA-005)
   - Show 5 water infrastructure roles loaded from JSON
   - Demonstrate type registration via Web UI
   - Display JSON schema validation

2. **Agent Instance Creation & Management** (INFRA-007, INFRA-008)
   - Create 27 agent instances representing a water distribution zone
   - Show agent state and metadata through UI
   - Demonstrate agent lifecycle (create, start, pause, stop)

3. **Leak Detection & Isolation** (INFRA-009)
   - Simulated pipe burst with sensor anomaly detection
   - Multi-agent collaboration using pub/sub messaging
   - Valve isolation and zone coordinator alerting
   - Real-time visualization of response

4. **Pressure Optimization** (INFRA-010)
   - Collaborative pump control based on sensor feedback
   - Zone-wide pressure balancing
   - Dashboard showing pressure trends and pump adjustments

5. **Predictive Maintenance Alert** (INFRA-011)
   - Pump degradation pattern detection
   - Maintenance prediction and work order generation
   - Historical efficiency data visualization

6. **Emergency Response** (INFRA-020) - Optional
   - Fire hydrant high-pressure request
   - Coordinated pump and valve response
   - Real-time network adjustment visualization

### Current Demo Capabilities ✅

**Available Now**:
- ✅ Web UI at http://localhost:8083 with Bulma CSS styling
- ✅ Role registry (5 water infrastructure types loaded)
- ✅ Agent instance management (293 agents representing Nairobi water network)
- ✅ Three complete demonstration scenarios:
  - ✅ Leak Detection (INFRA-009) - 4-step workflow with multi-agent coordination
  - ✅ Pressure Optimization (INFRA-010) - 3-cycle continuous optimization
  - ✅ Predictive Maintenance (INFRA-011) - 4-week degradation monitoring
- ✅ Geographic network visualizer (INFRA-017) - Deck.gl + MapLibre GL showing 293 agents on Nairobi map
- ✅ ArangoDB message and pub/sub infrastructure
- ✅ Real-time health monitoring and status display
- ✅ Configuration-based deployment with startup scripts

**Needs Implementation**:
- ⚠️ Time-series data storage (INFRA-013)
- ⚠️ Water-specific analytics queries (INFRA-015)
- ⚠️ Alert management UI (INFRA-018)
- ⚠️ Performance metrics dashboard (INFRA-019)

### Documentation Deliverables

**Completed** (See mvp_done.md):
- ✅ Agent type JSON schemas with detailed property definitions (INFRA-001 to INFRA-005)
- ✅ Configuration-based architecture documentation
- ✅ Environment variable configuration guide
- ✅ Six detailed coding session documents:
  - INFRA-001: Pipe Agent Implementation
  - INFRA-007: Fix Agent Instance Data Loading Path
  - INFRA-009: Leak Detection Scenario
  - INFRA-010: Pressure Optimization Scenario
  - INFRA-011: Predictive Maintenance Scenario
  - INFRA-017: Network Topology Visualizer
- ✅ Startup and deployment scripts
- ✅ Five detailed coding session documents:
  - INFRA-001: Pipe Agent Implementation
  - INFRA-007: Fix Agent Instance Data Loading Path
  - INFRA-009: Leak Detection Scenario
  - INFRA-010: Pressure Optimization Scenario
  - INFRA-011: Predictive Maintenance Scenario

**Remaining**:
- Architecture alignment: design → implementation mapping document
- Performance benchmarks (agent response times, message throughput)
- REST API usage guide for agent operations
- Visualization implementation guides

## Resource Requirements

### Team Members
- **Backend Developer (Go)**: Agent implementation, business logic, framework integration
- **DevOps Engineer**: Kubernetes deployment, CI/CD, infrastructure automation
- **IoT Specialist**: Sensor integration, protocol implementation, hardware interfacing
- **Flutter Developer**: Cross-platform dashboard UI (CodeValdFortex), real-time visualizations, user experience
- **Data Engineer**: Database design, time-series optimization, analytics queries

### Tools and Platforms
- **Development**: Go 1.21+, CodeValdCortex Framework, Docker, Git
- **Backend**: ArangoDB 3.11+ (multi-model database for documents, graphs, key-value, time-series)
- **IoT**: MQTT Broker (Mosquitto), Modbus libraries, OPC UA toolkit
- **Frontend (Primary)**: Flutter (CodeValdFortex) - Web, iOS, Android, Desktop
- **Frontend (Temporary Dev UI)**: Templ + HTMX + Alpine.js (to be removed via MVP-CLEANUP-001-014)
- **Visualization**: Chart.js, Deck.gl, MapLibre GL, SVG/D3.js
- **CI/CD**: GitHub Actions, automated testing, Docker builds
- **Monitoring**: Prometheus, Grafana, ELK Stack

### Infrastructure
- **Development**: Local Docker Compose environment
- **Staging**: Kubernetes cluster (3 nodes minimum)
- **Production**: Kubernetes cluster with auto-scaling
- **Edge Devices**: Raspberry Pi or industrial edge gateways for field deployment
- **Cloud**: AWS/GCP/Azure for central coordination (optional)

## Risk Assessment

### Technical Risks
- **Agent Complexity**: Implementing autonomous agents with complex state machines may be challenging
  - *Mitigation*: Start with simplified agent behaviors, iterate to add complexity
- **Real-Time Performance**: Meeting sub-second response time requirements for critical events
  - *Mitigation*: Use ArangoDB polling with optimized indexes, implement caching for hot data
- **IoT Integration**: Connecting to diverse sensor protocols (MQTT, Modbus, OPC UA)
  - *Mitigation*: Use abstraction layer for protocol handling, test with simulated sensors first
- **State Consistency**: Maintaining agent state consistency across failures and restarts
  - *Mitigation*: Leverage ArangoDB ACID transactions, implement proper state persistence

### Implementation Risks
- **Framework Learning Curve**: Team may need time to learn CodeValdCortex patterns and ArangoDB
  - *Mitigation*: Provide training sessions on CodeValdCortex and ArangoDB AQL, code examples, pair programming
- **Scope Expansion**: Adding too many roles or features beyond MVP
  - *Mitigation*: Strict adherence to phased approach, focus on 5 core roles first
- **Time-Series Data Volume**: High-frequency sensor data may overwhelm storage
  - *Mitigation*: Implement data aggregation in ArangoDB, use date-based collection partitioning and retention policies early

### Showcase Risks
- **Demo Reliability**: Live demos may fail due to timing issues or bugs
  - *Mitigation*: Prepare recorded demo videos as backup, rehearse extensively
- **Performance Under Load**: System may not scale as expected during demos
  - *Mitigation*: Load test before demos, have fallback to smaller dataset

## Success Metrics

### Technical Metrics
- **Agent Response Time**: <500ms for critical events (leak detection, pressure alerts) - ⏳ Pending scenario implementation
- **Message Throughput**: Support 1000+ messages/second between agents - ✅ Framework capable, needs load testing
- **Data Ingestion**: Handle 10,000+ sensor readings per minute - ⏳ Pending time-series implementation
- **System Uptime**: 99%+ during development, 99.9%+ for staging demos - ✅ Framework stable

### Functional Metrics
- **Roles**: All 5 core roles (Pipe, Sensor, Pump, Valve, Zone Coordinator) fully functional - ✅ Complete (JSON configurations)
- **Agent Instances**: Create and manage 27+ agent instances representing water network - ⏳ Pending instance creation
- **Communication Success**: 99%+ message delivery rate between agents - ✅ Framework tested
- **Leak Detection**: Identify and isolate simulated leaks within 30 seconds - ⏳ Pending scenario implementation
- **Predictive Accuracy**: >80% accuracy in predicting pump failures - ⏳ Pending ML implementation

### Showcase Metrics
- **Demo Completion**: Successfully demonstrate core scenarios - ⏳ 1/6 scenarios ready (agent management)
- **Dashboard Responsiveness**: Real-time updates displayed within 1 second - ✅ Framework UI responsive
- **Visualization Quality**: Clear, intuitive network topology with status indicators - ⏳ Pending topology visualizer
- **Documentation Coverage**: Complete mapping from design docs to implementation - ⚠️ Partial (JSON schemas documented)

### Business Value Metrics
- **Framework Validation**: Prove CodeValdCortex can handle complex IoT/agent systems - ✅ Architecture validated
- **Reusability**: Agent patterns can be adapted for other use cases - ✅ Configuration-only approach proven
- **Performance Proof**: Demonstrate sub-second agent coordination at scale - ⏳ Pending scenario testing
- **Market Readiness**: Showcase quality sufficient for demos and presentations - ⚠️ Needs scenario implementations

## Current Progress Summary

### ✅ Completed (41% of MVP)
1. **Phase 1**: All 5 role configurations (INFRA-001 to INFRA-005)
2. **Phase 2**: Complete framework communication system (INFRA-006)
3. **Phase 4**: Three demonstration scenarios (INFRA-009, INFRA-010, INFRA-011)
4. **Phase 5**: ArangoDB collections and agent state persistence (INFRA-012, INFRA-014)
5. **Phase 6**: Base web UI with agent management (INFRA-016), Geographic network visualizer (INFRA-017)
6. **Infrastructure**: Environment configuration, startup scripts, database auto-creation, 293-agent Nairobi water network

### ⏳ In Progress (0 tasks)
- None - Ready for next phase

### 🎯 Next Priorities
1. **INFRA-013**: Time-series data storage for sensor readings
2. **INFRA-015**: Historical analytics queries for infrastructure metrics
3. **INFRA-018**: Alert management UI

### 📊 Overall Statistics
- **Total MVP Tasks**: 27 tasks
- **Completed**: 12 tasks (44%)
- **Framework-Provided**: 5 tasks (19%)
- **Remaining**: 10 tasks (37%)
- **P0 (Critical) Complete**: 7/10 (70%)
- **P1 (High) Complete**: 5/11 (45%)
- **P2 (Enhancement) Complete**: 0/6 (0%)

## Workflow Integration

### Task Management Process
1. **Task Assignment**: Pick tasks based on phase and priority, following dependencies
2. **Implementation**: Update "Status" column as work progresses (Not Started → In Progress → Complete)
3. **Design Alignment**: Reference corresponding design documentation sections during implementation
4. **Completion Process** (MANDATORY):
   - Create detailed coding session document in `coding_sessions/` using format: `INFRA-{TaskID}_{description}.md`
   - Document implementation approach and key decisions
   - Include examples demonstrating functionality
   - Add completed task to `mvp_done.md` with completion date
   - Update any dependent task references in this file
   - Merge feature branch to main:
     ```bash
     # Merge when complete and tested
     git checkout main
     git merge feature/MVP-XXX_description
     git branch -d feature/MVP-XXX_description
     git push origin main
     ```
5. **Dependencies**: Ensure prerequisite tasks are completed before starting dependent work

### Branch Management
For significant new features:
```bash
# Create feature branch
git checkout -b feature/INFRA-XXX_description

# Work on implementation
# ... development work ...

# Validation before merge
# - Verify implementation works as expected
# - Run tests and validation
# - Check performance

# Merge when complete
git checkout main
git merge feature/INFRA-XXX_description
git branch -d feature/INFRA-XXX_description
git push origin main
```

**Note**: For configuration-only changes (roles, environment variables), direct commits to main branch are acceptable.

### Implementation Patterns Established

**Note**: Detailed implementation documentation for INFRA-001 through INFRA-011 is available in mvp_done.md.

**Configuration-Based Roles** ✅:
- Agent types defined in JSON files (`config/agents/*.json`)
- Framework auto-loads from `USECASE_CONFIG_DIR/config/agents/`
- No Go code needed for basic roles
- JSON schema validation enforced by framework
- Types persist to ArangoDB `agent_types` collection

**Agent Instance Management** ✅:
- Instances created via REST API POST `/api/v1/agents`
- Instances visible in Web UI at http://localhost:8083
- Full CRUD operations available (Create, Read, Update, Delete)
- State persisted to ArangoDB `agents` collection

**Communication Patterns** ✅:
- Direct messaging: `MessageService.SendMessage(fromID, toID, type, payload)`
- Pub/sub: `PubSubService.Publish(topic, payload)` and `Subscribe(pattern, handler)`
- Polling: `CommunicationPoller` automatically checks for new messages
- Topic patterns: `zone.*.alert`, `sensor.pressure.#` (wildcards supported)

**Scenario Patterns** ✅:
- Event-driven: Leak Detection (INFRA-009) - 4-step reactive workflow
- Continuous optimization: Pressure Management (INFRA-010) - 3-cycle adaptive loop
- Time-series predictive: Maintenance (INFRA-011) - 4-week degradation analysis

**Environment Configuration** ✅:
- All settings in `.env` file with `CVXC_` prefix
- Use case-specific settings with `USECASE_` prefix
- Water infrastructure thresholds (pressure, flow, conditions)
- Monitoring intervals configurable per role

### Repository Structure (Actual Implementation)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ Base Framework Module: github.com/aosanya/CodeValdCortex                   │
│ Location: /workspaces/CodeValdCortex/ (root)                               │
│ Status: CORE FUNCTIONALITY COMPLETE ✅                                      │
└─────────────────────────────────────────────────────────────────────────────┘
/workspaces/CodeValdCortex/
├── go.mod                               # Module: github.com/aosanya/CodeValdCortex
├── cmd/
│   └── main.go                          # ✅ Framework server entry point
├── internal/                            # FRAMEWORK IMPLEMENTATIONS
│   ├── agent/                           # ✅ Core agent runtime (lifecycle, tasks)
│   ├── communication/                   # ✅ INFRA-006: Message & pub/sub systems
│   ├── database/                        # ✅ INFRA-012: ArangoDB integration
│   ├── memory/                          # ✅ INFRA-014: State persistence
│   ├── registry/                        # ✅ Agent & role registries
│   ├── task/                            # ✅ Task scheduling system
│   ├── config/                          # ✅ Configuration management
│   ├── api/                             # ✅ REST API server
│   ├── web/                             # ✅ INFRA-016: Web UI (Bulma CSS)
│   └── app/                             # ✅ Application initialization
├── static/                              # ✅ CSS, JS, images (self-hosted)
│   ├── css/bulma.min.css                # ✅ Bulma CSS framework
│   └── js/                              # ✅ Deck.gl, MapLibre GL, Chart.js (Templ dev UI - temporary)
├── bin/
│   └── codevaldcortex                   # ✅ Compiled binary
└── documents/                           # Framework documentation

┌─────────────────────────────────────────────────────────────────────────────┐
│ Use Case Module: UC-INFRA-001 Water Distribution Network                   │
│ Location: /workspaces/CodeValdCortex/usecases/UC-INFRA-001-*/              │
│ Status: CONFIGURATION COMPLETE, SCENARIOS PENDING ⚠️                        │
└─────────────────────────────────────────────────────────────────────────────┘
/workspaces/CodeValdCortex/usecases/UC-INFRA-001-water-distribution-network/
├── .env                                 # ✅ Environment configuration
├── start.sh                             # ✅ Startup script
├── cmd/
│   └── main.go                          # ✅ Usage instructions (runs via framework)
├── config/
│   └── agents/                          # ✅ Agent type definitions
│       ├── pipe.json                    # ✅ INFRA-001: 221 lines, complete schema
│       ├── sensor.json                  # ✅ INFRA-002: 171 lines, complete schema
│       ├── pump.json                    # ✅ INFRA-003: 189 lines, complete schema
│       ├── valve.json                   # ✅ INFRA-004: 198 lines, complete schema
│       └── zone_coordinator.json        # ✅ INFRA-005: 324 lines, complete schema
├── bin/                                 # (not used - framework runs the show)
└── (scenarios/)                         # ⚠️ TODO: Scenario implementations
    ├── (leak_detection/)                # ⚠️ INFRA-009: To be implemented
    ├── (pressure_optimization/)         # ⚠️ INFRA-010: To be implemented
    └── (predictive_maintenance/)        # ⚠️ INFRA-011: To be implemented

Related Documentation:
/workspaces/CodeValdCortex/documents/
└── 3-SofwareDevelopment/Usecases/UC-INFRA-001-water-distribution-network/
    ├── mvp.md                           # ✅ This file - Active task list (UPDATED)
    ├── mvp_done.md                      # ✅ Completed tasks (1 task documented)
    └── coding_sessions/
        └── INFRA-001_pipe-agent.md      # ✅ Initial implementation session log
```

**Key Differences from Original Plan**:
- ✅ **Configuration-Only Approach**: No Go code in use case directory - framework loads JSON
- ✅ **Centralized Execution**: Use case runs through framework binary, not standalone
- ✅ **Environment-Driven**: All settings configured via .env file
- ⚠️ **Scenarios as Scripts**: Scenario implementations can be standalone scripts or framework extensions
- ✅ **No Separate Module**: Use case is configuration data, not a separate Go module

### Quality Gates

Before marking a task complete:
- [ ] Implementation works as expected (tested manually or automated)
- [ ] Changes don't break existing functionality
- [ ] Configuration follows established patterns
- [ ] Documentation updated (if applicable)
- [ ] Coding session document created (for significant work)
- [ ] Changes committed to git

**Note**: Given the configuration-only approach, many quality gates from the original plan are simplified or not applicable (no custom agent code, no state machines to test, etc.)

---

## Quick Start (Current State)

To run UC-INFRA-001 Water Distribution Network showcase:

```bash
# 1. Ensure ArangoDB is running (default: localhost:8529)
# Check docker-compose.yml in root or start manually

# 2. Navigate to use case directory
cd /workspaces/CodeValdCortex/usecases/UC-INFRA-001-water-distribution-network

# 3. Run the start script (builds framework if needed, loads environment, starts server)
./start.sh

# 4. Access the Web UI
open http://localhost:8083
# or if in dev container:
$BROWSER http://localhost:8083

# 5. Verify roles are loaded
# Navigate to "Roles" page - should show 5 infrastructure types + 5 core types
```

**What You'll See**:
- ✅ Dashboard with agent statistics
- ✅ Roles page listing all 10 types (5 core + 5 water infrastructure)
- ✅ Agent instances page (currently empty - INFRA-007 will populate)
- ✅ Health monitoring status

**Next Steps After This**:
1. Create agent instances (INFRA-007) using Web UI or REST API
2. Implement scenario scripts (INFRA-009, INFRA-010, INFRA-011)
3. Add topology visualizer to dashboard (INFRA-017)

---

**Note**: This MVP document has been updated to reflect the actual state of the UC-INFRA-001 implementation as of October 23, 2025. The configuration-based approach means many originally planned "implementation" tasks are complete via JSON configuration rather than custom Go code. Focus has shifted to creating agent instances and implementing demonstration scenarios that leverage the framework's messaging and coordination capabilities.