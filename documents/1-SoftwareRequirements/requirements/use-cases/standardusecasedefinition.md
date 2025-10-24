# Standard Use Case Definition Template for CodeValdCortex Agent Systems

**Version**: 1.0  
**Date**: October 23, 2025  
**Purpose**: Standard template for documenting agent-based use cases in the CodeValdCortex framework

## Document Structure

All use case documents in the CodeValdCortex system should follow this standardized structure to ensure consistency, completeness, and ease of understanding across different domains.

---

## Header Section

```markdown
# Use Case: [System Name] - [Brief Description]

**Use Case ID**: UC-[CATEGORY]-[NUMBER]  
**Use Case Name**: [Descriptive Name]  
**System**: [System Name]  
**Created**: [Date]  
**Status**: [Concept/Planning | Example/Reference | MVP Implementation | Production]
```

### Use Case ID Format
- **UC**: Use Case prefix
- **CATEGORY**: Domain category (e.g., CHAR for Charity, EVENT for Events, INFRA for Infrastructure, etc.)
- **NUMBER**: Sequential 3-digit number (001, 002, etc.)

### Status Values
- **Concept/Planning**: Initial ideation and requirements gathering
- **Example/Reference**: Template or reference implementation
- **MVP Implementation**: Minimum viable product in development
- **Production**: Deployed and operational

---

## 1. Overview Section

**Purpose**: Provide a high-level summary of the use case and its context within the CodeValdCortex framework.

**Required Content**:
- Brief description of what the system does
- How it uses the CodeValdCortex agent framework
- Key stakeholders or entities involved
- Optional: Etymology or meaning of the system name (especially for Swahili or culturally significant names)

**Template**:
```markdown
## Overview

[System Name] is an [example/production] agentic system built on the CodeValdCortex framework that demonstrates how [domain/problem area] can be [managed/facilitated/optimized] through autonomous agents. This use case focuses on [specific focus area] where [key entities] are modeled as intelligent agents that [primary capabilities].

**Note**: *"[System Name]" means "[meaning]" in [language], reflecting [significance].*
```

---

## 2. System Context Section

**Purpose**: Define the problem domain, business challenges, and proposed solution approach.

### 2.1 Domain
**Required Content**:
- Industry or problem domain
- Specific area of focus
- Target users or beneficiaries

**Template**:
```markdown
## System Context

### Domain
[Industry/sector description], [specific subdomain or focus area]
```

### 2.2 Business Problem
**Required Content**:
- List of current challenges or pain points
- Impact of these problems
- Why existing solutions are insufficient

**Template**:
```markdown
### Business Problem
Traditional [domain] systems suffer from:
- **[Problem Category 1]**: [Description of problem]
- **[Problem Category 2]**: [Description of problem]
- **[Problem Category 3]**: [Description of problem]
- **[Problem Category 4]**: [Description of problem]
- [Additional problems...]
```

### 2.3 Proposed Solution
**Required Content**:
- How the agent-based approach addresses the problems
- Key differentiators from traditional approaches
- High-level capabilities enabled by agents

**Template**:
```markdown
### Proposed Solution
An agentic system where each [element/entity] is an autonomous agent that:
- [Capability 1]: [How it solves problem]
- [Capability 2]: [How it solves problem]
- [Capability 3]: [How it solves problem]
- [Capability 4]: [How it solves problem]
- [Additional capabilities...]
```

---

## 3. Agent Types Section

**Purpose**: Detailed specification of each agent type in the system.

**Required Content**:
- Minimum 3-5 agent types (can be more for complex systems)
- Each agent must have: Represents, Attributes, Capabilities, State Machine, Example Behavior

### Agent Specification Template

```markdown
## Agent Types

### [Number]. [Agent Name] Agent

**Represents**: [What real-world entity, role, or concept this agent embodies]

**Attributes**:
- `attribute_id`: Unique identifier (e.g., "[PREFIX]-001")
- `attribute_name`: [Type] - [Description]
- `attribute_category`: [Enum values] - [Purpose]
- `related_agent_refs`: [References to other agents]
- [List 10-20 key attributes]

**Capabilities**:
- [Action/capability 1]
- [Action/capability 2]
- [Action/capability 3]
- [List 8-15 capabilities]

**State Machine**:
- `State_Name` - Description of state
- `Another_State` - Description of state
- `Final_State` - Description of state
- [List 4-8 states]

**Example Behavior**:
\```
IF [condition]
  THEN [action]
  AND [additional action]
  
IF [another condition]
  THEN [response]
  AND [coordination with other agents]
\```
```

### Agent Attribute Guidelines
- **ID Attributes**: Follow pattern PREFIX-XXX (e.g., "AGT-001", "SEN-P-045")
- **Enums**: Clearly define all possible values
- **Relationships**: Reference other agent types by ID
- **Measurements**: Include units (meters, liters/min, kg, etc.)
- **Temporal**: Use ISO 8601 format for dates/times

### Capabilities Guidelines
- Use active verbs (Monitor, Detect, Calculate, Coordinate, etc.)
- Be specific and actionable
- Include both autonomous and coordinated capabilities
- Cover normal operations, exception handling, and optimization

### State Machine Guidelines
- Use clear, descriptive state names (PascalCase or Snake_Case)
- Include normal operational states
- Include error/exception states
- Include maintenance/offline states
- Keep to 4-8 states for simplicity

### Example Behavior Guidelines
- Use pseudo-code format
- Show conditional logic (IF-THEN-AND patterns)
- Demonstrate agent autonomy
- Show inter-agent communication
- Include both routine and exception scenarios

---

## 4. Agent Interaction Scenarios Section

**Purpose**: Demonstrate how agents work together to achieve system goals through concrete scenarios.

**Required Content**:
- Minimum 3 realistic scenarios
- Each scenario shows multi-agent coordination
- Include both normal and exceptional situations

### Scenario Template

```markdown
## Agent Interaction Scenarios

### Scenario [Number]: [Scenario Name]

**Trigger**: [What initiates this scenario]

**Agent Interaction Flow**:

1. **[Agent Type] ([Agent ID])** [initial action]
   \```
   State: [From] → [To]
   Action: [What it does]
   Decision: [What it decides]
   \```

2. **[Agent Type]** [responds or coordinates]
   \```
   → [Target Agent]: "[Communication message]"
   → [Another Agent]: "[Communication message]"
   Action: [Coordinated action]
   \```

3. **[Next Agent]** [continues the flow]
   \```
   [Details of agent action]
   [Coordination details]
   \```

[Continue with 5-10 steps showing complete scenario resolution]
```

### Scenario Guidelines
- **Scenario 1**: Normal operations or common use case
- **Scenario 2**: Exception handling or emergency response
- **Scenario 3**: Optimization or learning scenario
- Show realistic agent IDs (e.g., PIPE-045, SENS-P-12, etc.)
- Include timing information when relevant
- Show both autonomous decisions and coordinated actions
- Demonstrate value of agent-based approach

---

## 5. Technical Architecture Section

**Purpose**: Define how agents communicate, data flows, and deployment structure.

### 5.1 Agent Communication Patterns

**Required Content**:
- Communication protocols used
- Message types and formats
- Synchronous vs asynchronous patterns

**Template**:
```markdown
## Technical Architecture

### Agent Communication Patterns

1. **[Pattern Type]** (via [mechanism]):
   - [Use case for this pattern]
   - [Message types]
   - Protocol: [Technology/protocol used]

2. **[Another Pattern]**:
   - [Use case]
   - [Details]
   - Protocol: [Technology]

[List 3-5 communication patterns]
```

### 5.2 Data Flow

**Required Content**:
- ASCII diagram or structured text showing data flow
- Direction of information flow
- Key transformation or aggregation points

**Template**:
```markdown
### Data Flow

\```
[Source Agents]
  ↓ ([data/event type])
[Processing Agents]
  ↓ ([transformed data])
[Coordination Agents]
  ↓ ([aggregated data])
[Presentation/Action Layer]
\```
```

### 5.3 Agent Deployment Model

**Required Content**:
- CodeValdCortex framework components used
- Deployment architecture
- Database and storage considerations
- External integrations

**Template**:
```markdown
### Agent Deployment Model

**CodeValdCortex Framework Components Used**:
1. **Runtime Manager**: [How it's used]
2. **Agent Registry**: [What's tracked]
3. **Task System**: [What's scheduled]
4. **Memory Service**: [What's stored]
5. **Communication System**: [How agents communicate]
6. **Configuration Service**: [What's configured]
7. **Health Monitor**: [What's monitored]

**Deployment Architecture**:
\```
[User/Client Layer]
  ↓
[API Gateway]
  ↓
[Application Servers (CodeValdCortex Runtime)]
  ├─ [Agent Type 1] Agents
  ├─ [Agent Type 2] Agents
  ├─ [Agent Type 3] Agents
  └─ [Agent Type N] Agents
  ↓
[Data Layer]
  ├─ [Database 1] ([Technology])
  ├─ [Database 2] ([Technology])
  ├─ [Cache] ([Technology])
  └─ [File Storage] ([Technology])
  ↓
[External Integrations]
  ├─ [Integration 1] ([Purpose])
  ├─ [Integration 2] ([Purpose])
  └─ [Integration N] ([Purpose])
\```
```

---

## 6. Integration Points Section

**Purpose**: Define external systems and services that integrate with the agent system.

**Template**:
```markdown
## Integration Points

### 1. [Integration Name]
- [Purpose/functionality]
- [Data exchanged]
- [Frequency of interaction]
- Integration: [Technology/API used]

### 2. [Another Integration]
- [Details]
- Integration: [Technology]

[List 5-10 key integrations]
```

---

## 7. Benefits Demonstrated Section

**Purpose**: Quantify the value and improvements from the agent-based approach.

**Template**:
```markdown
## Benefits Demonstrated

### 1. [Benefit Category]
- **Before**: [Traditional approach]
- **With Agents**: [Agent-based approach]
- **Metric**: [Quantifiable improvement]

### 2. [Another Benefit]
- **Before**: [Problem]
- **With Agents**: [Solution]
- **Metric**: [Measurement]

[List 6-10 key benefits with metrics]
```

### Benefit Categories (Examples)
- Efficiency improvements
- Cost reductions
- Response time improvements
- Quality improvements
- Scalability gains
- User satisfaction
- Risk reduction
- Sustainability impact

---

## 8. Implementation Phases Section

**Purpose**: Define the roadmap for building the system.

**Template**:
```markdown
## Implementation Phases

### Phase 1: [Phase Name] (Months X-Y)
- [Deliverable 1]
- [Deliverable 2]
- [Deliverable 3]
- **Deliverable**: [Key milestone]

### Phase 2: [Phase Name] (Months X-Y)
- [Activities]
- **Deliverable**: [Milestone]

### Phase 3: [Phase Name] (Months X-Y)
- [Activities]
- **Deliverable**: [Milestone]

### Phase 4: [Phase Name] (Months X-Y)
- [Activities]
- **Deliverable**: [Milestone]
```

### Phase Guidelines
- Plan for 3-5 phases
- Each phase 2-4 months
- Build core infrastructure first
- Add intelligence/optimization last
- Define clear deliverables

---

## 9. Success Criteria Section

**Purpose**: Define measurable success metrics across technical, operational, and business dimensions.

**Template**:
```markdown
## Success Criteria

### Technical Metrics
- ✅ [Uptime percentage]
- ✅ [Response time]
- ✅ [Scalability metric]
- ✅ [Quality metric]

### Operational Metrics
- ✅ [Efficiency metric]
- ✅ [Cost metric]
- ✅ [Performance metric]
- ✅ [Reliability metric]

### Business Metrics
- ✅ [ROI metric]
- ✅ [Revenue/savings metric]
- ✅ [Customer satisfaction]
- ✅ [Market metric]

### Impact Metrics (if applicable)
- ✅ [Social impact]
- ✅ [Environmental impact]
- ✅ [Community benefit]
```

---

## 10. Conclusion Section

**Purpose**: Summarize the use case value and broader applicability.

**Template**:
```markdown
## Conclusion

[System Name] demonstrates the power of the CodeValdCortex agent framework applied to [domain]. By treating [entities] as intelligent, autonomous agents, the system achieves:

- **[Key Achievement 1]**: [Description]
- **[Key Achievement 2]**: [Description]
- **[Key Achievement 3]**: [Description]
- **[Key Achievement 4]**: [Description]
- **[Key Achievement 5]**: [Description]

This use case serves as a reference implementation for applying agentic principles to other [domain] areas such as [related domain 1], [related domain 2], [related domain 3], and [related domain 4].

---

**Related Documents**:
- System Architecture: `documents/2-SoftwareDesignAndArchitecture/`
- Agent Implementation: `internal/agent/`
- Communication System: `internal/communication/`
- [Domain-specific system]: `internal/[system]/`
- API Documentation: `documents/4-QA/`
- Dashboard: MVP-015 Management Dashboard

**Related Use Cases**:
- [UC-XXX-001]: [Related use case name]
- [UC-YYY-001]: [Related use case name]
```

---

## Appendices (Optional)

### A. Glossary
Define domain-specific terms and acronyms

### B. References
External standards, research papers, or industry guidelines

### C. Compliance Considerations
Regulatory or compliance requirements specific to the domain

### D. Security & Privacy
Security model and privacy protections for sensitive use cases

---

## References and Framework Rules

### Framework Topology Visualizer

When designing use cases that require visualization of agent networks and topologies, refer to the **Framework Topology Visualizer** documentation:

**Location**: `/workspaces/CodeValdCortex/documents/2-SoftwareDesignAndArchitecture/framework-topology-visualizer/`

**Key Documents**:
- `README.md` - Overview and file structure
- `01-overview.md` - Executive summary, goals, MVP scope
- `02-architecture.md` - Component architecture and determinism contract
- `03-data-source-and-inference.md` - Agent API integration, edge inference, JSON Patch semantics
- `04-rendering.md` - Renderer selection, layouts, basemap behavior
- `05-security-and-testing.md` - RBAC, expression sandboxing, testing strategy
- `06-delivery-mvp.md` - Delivery phases and acceptance criteria
- `07-canonical_types.json` - Standard relationship taxonomy
- `00-full.md` - Complete canonical reference document

**When to Apply**:
Use the Framework Topology Visualizer specifications when your use case involves:
- Multiple agent types with spatial or network relationships
- Real-time visualization of agent networks and topologies
- Cross-domain relationship modeling (supply, observe, route, command, host, depends_on)
- Geographic or force-directed layout of agent systems
- Interactive topology exploration and analysis

**Mandatory Rules**:

1. **Canonical Relationship Types**: Use the standardized relationship taxonomy defined in `07-canonical_types.json`:
   - `supply` - Resource/service provision
   - `observe` - Monitoring/sensing relationships
   - `route` - Path/flow connections
   - `command` - Control/authority relationships
   - `host` - Container/hosting relationships
   - `depends_on` - Dependency relationships

2. **Agent Data Structure**: Ensure all agents include:
   - `id` - Unique identifier
   - `agentType` - Type classification
   - `coordinates` - Geographic location (if applicable)
   - `metadata` - Additional attributes for visualization

3. **Schema Versioning**: All visualization configurations must:
   - Include mandatory `schemaVersion` field
   - Follow semantic versioning (MAJOR.MINOR.PATCH)
   - Validate against `visualization-config.schema.json`

4. **Deterministic Rendering**:
   - Use stable, deterministic IDs for agents and relationships
   - Apply seeded layout algorithms for consistent positioning
   - Document layout seed values in use case specifications

5. **Security and Access Control**:
   - Define RBAC policies for agent data access
   - Use server-side expression sandboxing for dynamic filters
   - Document security requirements in use case

6. **Real-time Updates**:
   - Use JSON Patch (RFC 6902) for incremental updates
   - Include sequence numbers for replay and consistency
   - Handle update conflicts and network failures

7. **Agent API Integration**:
   - Use the Agent API as the single source of truth
   - Document API endpoints used for agent data
   - Specify polling intervals or WebSocket configuration

**Integration with Use Case Template**:

When documenting a use case with visualization requirements:

- **Section 3 (Agent Types)**: Include visualization-relevant attributes in agent specifications
  - Add `coordinates` attribute for geographic positioning
  - Define `visualization_metadata` for display properties
  - Specify relationship types using canonical taxonomy

- **Section 4 (Agent Interaction Scenarios)**: Show how topology changes during scenarios
  - Document edge creation/deletion triggers
  - Illustrate network state before/after interactions

- **Section 5 (Technical Architecture)**: 
  - Reference Framework Topology Visualizer in deployment model
  - Specify visualization configuration schema version
  - Document renderer choice (Canvas, SVG, WebGL, MapLibre-GL)

- **Section 6 (Integration Points)**:
  - List Agent API endpoints for topology data
  - Specify visualization update mechanisms (polling/WebSocket)
  - Document basemap services if using geographic rendering

- **Section 9 (Success Criteria)**:
  - Add visualization-specific metrics (render time, frame rate, update latency)
  - Define topology complexity limits (max nodes, max edges)

**Example Reference**:
```markdown
**Visualization**: This use case uses the Framework Topology Visualizer 
(schema version 1.0.0) with Canvas rendering and Force-Directed layout. 
Relationships follow the canonical taxonomy defined in 07-canonical_types.json, 
using `supply`, `route`, and `observe` edge types. See visualization 
configuration in `/usecases/UC-INFRA-001-water-distribution-network/viz-config.json`.
```

---

## Document Guidelines

### Writing Style
- Use clear, concise language
- Define acronyms on first use
- Use consistent terminology throughout
- Include examples to illustrate concepts

### Formatting
- Use markdown consistently
- Number all agent types
- Use code blocks for pseudo-code
- Use bullet points for lists
- Use tables for complex comparisons

### Completeness Checklist
- ✅ All 10 main sections present
- ✅ Minimum 3-5 agent types defined
- ✅ Each agent has all required subsections
- ✅ Minimum 3 interaction scenarios
- ✅ Technical architecture diagram included
- ✅ Quantifiable metrics provided
- ✅ Implementation phases defined
- ✅ Success criteria measurable
- ✅ Related documents linked

### Maintenance
- Update status as project progresses
- Revise metrics based on actual results
- Add new scenarios as they emerge
- Keep agent specifications current with implementation
- Document lessons learned

---

**Document Template Version**: 1.0  
**Last Updated**: October 23, 2025  
**Maintained By**: System Architecture Team
