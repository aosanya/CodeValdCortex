# Agency Problem Definition and Work Items Documentation

## Overview

The CodeValdCortex platform introduces two key concepts for agency management:

1. **Problem Definition Module**: A structured way to define and catalog the specific problems that an agency aims to solve
2. **Work Items**: Discrete, manageable work packages that break down the problem-solving process into actionable tasks

## Problem Definition Module

### Purpose
The Problem Definition Module serves as the foundation for any agency's operational framework by clearly articulating:
- The specific challenges the agency addresses
- The scope and boundaries of each problem
- The relationship between different problems
- Success criteria for problem resolution

### Structure
Each problem definition includes:
- **Problem Code**: A unique identifier (e.g., PROB-001, RISK-ANALYSIS-01)
- **Problem Description**: A clear, detailed explanation of the challenge
- **Problem Scope**: Boundaries and constraints for the problem domain
- **Success Metrics**: Quantifiable measures of problem resolution

## Work Items (WI)

### Definition
Work Items are the fundamental building blocks for agency operations, representing discrete, manageable work packages that contribute to solving the defined problems. Each Work Item is designed to be:
- **Actionable**: Clear tasks that can be executed
- **Measurable**: Defined outcomes and deliverables
- **Bounded**: Limited scope with clear start and end points
- **Assignable**: Can be allocated to specific roles or agents

### Structure
Each Work Item includes:
- **WI Code**: Unique identifier (e.g., WI-001, TASK-ANALYSIS-01)
- **WI Description**: Detailed explanation of the work to be performed
- **Deliverables**: Expected outputs and outcomes
- **Dependencies**: Prerequisites and relationships to other WIs
- **Problem Relationships**: Mapping to one or more problems this work item addresses
- **RACI Matrix**: Role and responsibility assignments

## Problem-Work Item Relationship Mapping

### Purpose
The relationship mapping creates explicit connections between Work Items and the problems they solve, ensuring:
- **Traceability**: Clear links from problems to solutions
- **Coverage**: Verification that all problems have corresponding work items
- **Impact Analysis**: Understanding which work items affect which problems
- **Progress Tracking**: Monitoring problem resolution through work item completion

### Relationship Structure (Graph Database)
Each Work Item can have multiple problem relationships, modeled as graph edges in ArangoDB:

**Document Collections:**
- `problems` - Problem Definition documents
- `work_items` - Work Item documents  

**Edge Collection:**
- `problem_work_item_relationships` - Edges connecting problems to work items

**Edge Document Structure:**
```json
{
  "_from": "problems/PROB-001",
  "_to": "work_items/WI-001", 
  "relationship_type": "solves",
  "contribution_description": "Implements core data collection mechanism for risk analysis",
  "impact_level": "primary",
  "created_at": "2024-10-29T10:00:00Z",
  "updated_at": "2024-10-29T10:00:00Z"
}
```

**Graph Traversal Benefits:**
- **Multi-hop Queries**: Find all work items that solve problems related to a specific domain
- **Impact Analysis**: Trace which problems are affected by work item changes
- **Dependency Mapping**: Discover transitive relationships between problems and solutions
- **Coverage Analysis**: Identify problems without corresponding work items

### Relationship Types

| Type | Description | Example |
|------|-------------|---------|
| **solves** | Work item directly addresses and resolves the problem | "This work item solves PROB-001 by implementing automated data collection" |
| **supports** | Work item contributes to solving the problem but doesn't fully resolve it | "This work item supports PROB-002 by providing data validation capabilities" |
| **enables** | Work item creates prerequisites or foundations for solving the problem | "This work item enables PROB-003 by establishing the required infrastructure" |
| **mitigates** | Work item reduces the impact or likelihood of the problem | "This work item mitigates PROB-004 by implementing security controls" |

### Impact Levels

| Level | Description | Usage |
|-------|-------------|-------|
| **primary** | Work item is a main contributor to solving the problem | Core implementation work items |
| **secondary** | Work item provides important but not critical support | Supporting infrastructure, validation |
| **tertiary** | Work item has minimal but relevant impact | Documentation, minor enhancements |

### User Interface for Relationship Mapping

#### Relationship Editor Interface
```
Work Item: WI-001 - Data Collection System

┌─ Problem Relationships ─────────────────────────────────────────────┐
│                                                                     │
│ ┌─ Relationship 1 ─────────────────────────────────────────────┐   │
│ │ Problem Code: [PROB-001        ▼] Search/Select             │   │
│ │ Relationship: [solves          ▼] solves/supports/enables   │   │
│ │ Impact Level: [primary         ▼] primary/secondary/tertiary│   │
│ │ Description:                                                │   │
│ │ ┌─────────────────────────────────────────────────────────┐ │   │
│ │ │ This work item solves PROB-001 by implementing         │ │   │
│ │ │ automated data collection from multiple financial       │ │   │
│ │ │ sources with real-time validation                      │ │   │
│ │ └─────────────────────────────────────────────────────────┘ │   │
│ │                                           [Remove] [Edit] │   │
│ └───────────────────────────────────────────────────────────────┘   │
│                                                                     │
│ ┌─ Relationship 2 ─────────────────────────────────────────────┐   │
│ │ Problem Code: [PROB-002        ▼] Search/Select             │   │
│ │ Relationship: [supports        ▼] solves/supports/enables   │   │
│ │ Impact Level: [secondary       ▼] primary/secondary/tertiary│   │
│ │ Description:                                                │   │
│ │ ┌─────────────────────────────────────────────────────────┐ │   │
│ │ │ This work item supports PROB-002 by providing data     │ │   │
│ │ │ quality validation that ensures accuracy of risk       │ │   │
│ │ │ calculations                                            │ │   │
│ │ └─────────────────────────────────────────────────────────┘ │   │
│ │                                           [Remove] [Edit] │   │
│ └───────────────────────────────────────────────────────────────┘   │
│                                                                     │
│                                              [+ Add Relationship] │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

#### Problem Validation
- **Problem Code Validation**: Ensure referenced problem codes exist in the agency
- **Duplicate Prevention**: Prevent multiple relationships to the same problem with the same type
- **Required Fields**: All relationship fields must be completed
- **Description Templates**: Provide templates like "This work item {relationship_type} {problem_code} by..."

## RACI Matrix for Work Items

### RACI Definition

The RACI matrix is a responsibility assignment matrix used to clarify roles and responsibilities for each Work Item. RACI stands for:

- **R - Responsible**: The role(s) that perform the work to complete the task
- **A - Accountable**: The role that is ultimately answerable for the completion and approval of the task
- **C - Consulted**: The role(s) that provide input and expertise during task execution
- **I - Informed**: The role(s) that need to be kept informed of progress and decisions

### RACI Matrix Template for Work Items

For each Work Item, the following roles should be considered in the RACI matrix:

#### Standard Agency Roles

| Role | Description |
|------|-------------|
| **Agency Lead** | Overall responsible for agency strategy and decisions |
| **Technical Lead** | Responsible for technical architecture and implementation |
| **Domain Expert** | Subject matter expert for the specific problem domain |
| **Quality Assurance** | Ensures deliverable quality and compliance |
| **Stakeholder Representative** | Represents end-user or client interests |
| **Agent Coordinator** | Manages AI agent assignments and orchestration |
| **Data Analyst** | Handles data requirements and analysis |
| **Security Officer** | Ensures security compliance and risk management |

#### RACI Matrix Template

```
Work Item: [WI Code] - [WI Description]

┌─────────────────────────┬─────────────┬─────────────┬──────────────┬─────────────────────┬─────────────────┬─────────────┬─────────────────┐
│ Role                    │ Agency Lead │ Tech Lead   │ Domain Expert │ QA                  │ Stakeholder Rep │ Agent Coord │ Security Officer │
├─────────────────────────┼─────────────┼─────────────┼──────────────┼─────────────────────┼─────────────────┼─────────────┼─────────────────┤
│ Task Definition         │     A       │     C       │      R       │          I          │        C        │      I      │        I        │
│ Technical Implementation│     I       │     A       │      C       │          C          │        I        │      R      │        C        │
│ Quality Review          │     I       │     C       │      C       │          A          │        I        │      I      │        C        │
│ Stakeholder Approval    │     C       │     I       │      I       │          C          │        A        │      I      │        I        │
│ Deployment Decision     │     A       │     R       │      C       │          C          │        C        │      R      │        C        │
└─────────────────────────┴─────────────┴─────────────┴──────────────┴─────────────────────┴─────────────────┴─────────────┴─────────────────┘
```

### RACI Guidelines for Work Items

#### Best Practices

1. **Single Accountable (A)**: Each activity should have exactly one Accountable role
2. **Clear Responsible (R)**: At least one role must be Responsible for each activity
3. **Appropriate Consultation (C)**: Include roles that provide essential input
4. **Relevant Information (I)**: Keep stakeholders informed without overwhelming them

#### Common RACI Patterns by Work Type

##### 1. **Analysis and Research WIs**
- **Domain Expert**: Usually Responsible (R) for conducting analysis
- **Agency Lead**: Accountable (A) for outcomes and decisions
- **Technical Lead**: Consulted (C) for technical feasibility
- **Stakeholder Rep**: Informed (I) of findings

##### 2. **Technical Implementation WIs**
- **Technical Lead**: Accountable (A) for technical delivery
- **Agent Coordinator**: Responsible (R) for agent orchestration
- **Domain Expert**: Consulted (C) for domain-specific requirements
- **Quality Assurance**: Consulted (C) for testing requirements

##### 3. **Quality Assurance WIs**
- **Quality Assurance**: Accountable (A) and Responsible (R) for testing
- **Technical Lead**: Consulted (C) for technical criteria
- **Stakeholder Rep**: Consulted (C) for acceptance criteria
- **Agency Lead**: Informed (I) of quality status

##### 4. **Stakeholder Communication WIs**
- **Stakeholder Rep**: Accountable (A) for stakeholder satisfaction
- **Agency Lead**: Responsible (R) for communication strategy
- **Domain Expert**: Consulted (C) for technical explanations
- **All Roles**: Informed (I) of stakeholder feedback

### Implementation in CodeValdCortex

#### Data Model Extension (Graph Database)

The system uses ArangoDB's graph capabilities to model relationships:

**Work Item Document:**
```go
type WorkItem struct {
    Key         string           `json:"_key,omitempty"`
    ID          string           `json:"_id,omitempty"`
    AgencyID    string           `json:"agency_id"`
    Number      int              `json:"number"`
    Code        string           `json:"code"`
    Description string           `json:"description"`
    RACI        RACIMatrix       `json:"raci"`
    CreatedAt   time.Time        `json:"created_at"`
    UpdatedAt   time.Time        `json:"updated_at"`
}
```

**Problem Document:**
```go
type Problem struct {
    Key         string    `json:"_key,omitempty"`
    ID          string    `json:"_id,omitempty"`
    AgencyID    string    `json:"agency_id"`
    Number      int       `json:"number"`
    Code        string    `json:"code"`
    Description string    `json:"description"`
    Scope       string    `json:"scope"`
    SuccessMetrics []string `json:"success_metrics"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

**Relationship Edge:**
```go
type ProblemWorkItemRelationship struct {
    Key                     string            `json:"_key,omitempty"`
    ID                      string            `json:"_id,omitempty"`
    From                    string            `json:"_from"` // problems/{problem_key}
    To                      string            `json:"_to"`   // work_items/{work_item_key}
    RelationshipType        RelationshipType  `json:"relationship_type"`
    ContributionDescription string            `json:"contribution_description"`
    ImpactLevel            ImpactLevel       `json:"impact_level"`
    CreatedAt              time.Time         `json:"created_at"`
    UpdatedAt              time.Time         `json:"updated_at"`
}

type RelationshipType string

const (
    RelationshipSolves    RelationshipType = "solves"
    RelationshipSupports  RelationshipType = "supports"
    RelationshipEnables   RelationshipType = "enables"
    RelationshipMitigates RelationshipType = "mitigates"
)

type ImpactLevel string

const (
    ImpactPrimary   ImpactLevel = "primary"
    ImpactSecondary ImpactLevel = "secondary"
    ImpactTertiary  ImpactLevel = "tertiary"
)

type RACIMatrix struct {
    Activities []RACIActivity `json:"activities"`
}

type RACIActivity struct {
    Name        string              `json:"name"`
    Description string              `json:"description"`
    Assignments map[string]RACIRole `json:"assignments"`
}

type RACIRole string

const (
    RACIResponsible RACIRole = "R"
    RACIAccountable RACIRole = "A"
    RACIConsulted   RACIRole = "C"
    RACIInformed    RACIRole = "I"
)
```

**Graph Queries (AQL Examples):**

*Find all work items solving a specific problem:*
```aql
FOR v, e IN 1..1 OUTBOUND "problems/PROB-001" problem_work_item_relationships
  FILTER e.relationship_type == "solves"
  RETURN {work_item: v, relationship: e}
```

*Find all problems addressed by a work item:*
```aql
FOR v, e IN 1..1 INBOUND "work_items/WI-001" problem_work_item_relationships
  RETURN {problem: v, relationship: e}
```

*Coverage analysis - problems without work items:*
```aql
FOR p IN problems
  LET work_items = (
    FOR v IN 1..1 OUTBOUND p._id problem_work_item_relationships
      RETURN v
  )
  FILTER LENGTH(work_items) == 0
  RETURN p
```

type RACIMatrix struct {
    Activities []RACIActivity `json:"activities"`
}

type RACIActivity struct {
    Name        string              `json:"name"`
    Description string              `json:"description"`
    Assignments map[string]RACIRole `json:"assignments"`
}

type RACIRole string

const (
    RACIResponsible RACIRole = "R"
    RACIAccountable RACIRole = "A"
    RACIConsulted   RACIRole = "C"
    RACIInformed    RACIRole = "I"
)
```

#### User Interface Integration

The Agency Designer should include:
1. **Graph Relationship Editor**: Visual interface for creating edges between problems and work items
2. **Graph Visualization**: Interactive diagram showing problem-work item relationships
3. **RACI Matrix Editor**: Visual interface for defining role assignments
4. **Role Templates**: Pre-defined RACI patterns for common WI types
5. **Relationship Templates**: Pre-defined relationship descriptions and patterns
6. **Graph Analytics**: Coverage analysis, impact analysis, and dependency mapping
7. **Validation Rules**: Ensure valid graph references and appropriate RACI assignments
8. **Traceability Views**: Multi-hop traversals and relationship paths
9. **Export Capabilities**: Generate RACI and relationship documentation for stakeholder review

**Graph Database Benefits:**
- **Performance**: Efficient traversal of complex relationship networks
- **Flexibility**: Easy addition of new relationship types and multi-hop queries
- **Scalability**: Handle large networks of problems and work items
- **Analytics**: Built-in graph algorithms for network analysis
- **Consistency**: ACID transactions for relationship integrity

### Benefits of RACI Implementation

1. **Clear Accountability**: Eliminates confusion about who is responsible for what
2. **Improved Communication**: Defines who needs to be consulted and informed
3. **Risk Mitigation**: Prevents tasks from falling through gaps
4. **Efficiency**: Reduces unnecessary meetings and communications
5. **Quality Assurance**: Ensures proper review and approval processes
6. **Scalability**: Enables consistent role assignment across multiple agencies

### Example RACI Matrix for Financial Risk Analysis

```
Work Item: FRA-WI-001 - Financial Data Collection and Validation

Problem Relationships:
- Solves PROB-001 (Data Quality Issues) by implementing automated validation
- Supports PROB-003 (Real-time Analysis) by providing clean data feeds

┌─────────────────────────┬─────────────┬─────────────┬──────────────┬─────────────────────┬─────────────────┬─────────────┐
│ Activity                │ Agency Lead │ Tech Lead   │ Risk Analyst │ QA Engineer         │ Client Rep      │ Data Agent  │
├─────────────────────────┼─────────────┼─────────────┼──────────────┼─────────────────────┼─────────────────┼─────────────┤
│ Data Source Definition  │     A       │     C       │      R       │          I          │        C        │      I      │
│ Data Collection Setup   │     I       │     A       │      C       │          C          │        I        │      R      │
│ Data Validation Rules   │     I       │     C       │      A       │          R          │        I        │      C      │
│ Quality Review          │     I       │     I       │      C       │          A          │        I        │      I      │
│ Client Approval         │     C       │     I       │      I       │          C          │        A        │      I      │
└─────────────────────────┴─────────────┴─────────────┴──────────────┴─────────────────────┴─────────────────┴─────────────┘
```

This comprehensive example shows:
- **Problem Relationships**: Clear mapping to specific problems being addressed
- **RACI Assignments**: Detailed responsibility matrix for each activity
- **Traceability**: Direct links from work execution to problem resolution

### Complete Work Item Example with Graph Relationships

**Work Item Document:**
```json
{
  "_key": "WI-001",
  "_id": "work_items/WI-001",
  "code": "FRA-WI-001",
  "description": "Financial Data Collection and Validation System",
  "agency_id": "financial_risk_agency",
  "raci": {
    "activities": [
      {
        "name": "Data Source Definition",
        "assignments": {
          "Agency Lead": "A",
          "Risk Analyst": "R", 
          "Tech Lead": "C",
          "Client Rep": "C"
        }
      }
    ]
  }
}
```

**Relationship Edges:**
```json
[
  {
    "_from": "problems/PROB-001",
    "_to": "work_items/WI-001",
    "relationship_type": "solves",
    "contribution_description": "This work item solves PROB-001 by implementing automated data collection with real-time validation that eliminates manual data entry errors and ensures 99.9% data accuracy",
    "impact_level": "primary"
  },
  {
    "_from": "problems/PROB-003", 
    "_to": "work_items/WI-001",
    "relationship_type": "supports",
    "contribution_description": "This work item supports PROB-003 by providing clean, validated data feeds that enable real-time risk analysis calculations",
    "impact_level": "secondary"
  }
]
```

**Graph Visualization:**
```
PROB-001 ──[solves/primary]──┐
                               ├──► WI-001
PROB-003 ──[supports/secondary]┘
```

## Conclusion

The integration of Problem Definition and Work Items with RACI matrices provides a comprehensive framework for agency operations within CodeValdCortex. This approach ensures:

- **Clarity**: Clear problem definitions and work breakdown
- **Accountability**: Explicit role assignments and responsibilities
- **Efficiency**: Streamlined coordination and communication
- **Quality**: Structured review and approval processes
- **Scalability**: Reusable patterns across multiple agencies and use cases

This documentation should be referenced when creating new agencies, defining problems, or establishing Work Items to ensure consistent and effective operational practices.