# Genesis Agency - Agency Builder

**Agency ID**: `agency_genesis`  
**Status**: Design Phase  
**Category**: Meta-Agency (Platform Core)  
**Icon**: 🏭

---

## Overview

The Genesis Agency is the **foundational meta-agency** of the CodeValdCortex platform. Its sole purpose is to design, build, and deploy other agencies through collaborative human-AI interaction.

This is not just a tool—it's a **self-contained operational agency** with its own agents, workflows, and deliverables that happen to produce other agencies as outputs.

---

## Core Concept

**Genesis Agency = Agency Builder Agency**

- **Input**: Keywords, requirements, documents (RFPs, SOWs, specs)
- **Process**: Iterative research and specification refinement
- **Output**: Complete agency artifacts ready for deployment

---

## Directory Structure

```
genesis-agency/
├── README.md                          # This file
├── SPECIFICATION_SUMMARY.md           # Summary and statistics
├── IMPLEMENTATION_ROADMAP.md          # Development roadmap
├── import/                            # Importable JSON specifications
│   ├── introduction.json              # Agency introduction (6 sections)
│   ├── goals.json                     # Goals catalog (10 goals)
│   ├── roles.json                     # Role specifications (5 roles)
│   ├── work-items.json                # Work items catalog (16 items)
│   └── workflows.json                 # Workflow definitions (3 workflows)
└── artifacts/                         # Output artifact schemas
    ├── agency-specification.schema.json
    ├── database-init.schema.json
    └── deployment-manifest.schema.json
```

---

## Navigation

- [Specification Summary](./SPECIFICATION_SUMMARY.md) - Overview and statistics
- [Implementation Roadmap](./IMPLEMENTATION_ROADMAP.md) - Development plan
- [Import Files](./import/) - Importable JSON specifications
  - [introduction.json](./import/introduction.json) - Problem statement, context, scope (6 sections)
  - [goals.json](./import/goals.json) - 10 SMART goals
  - [roles.json](./import/roles.json) - 5 agent roles with work item mappings
  - [work-items.json](./import/work-items.json) - 16 work items with deliverables
  - [workflows.json](./import/workflows.json) - 3 workflows (Creation, Refinement, Validation)
- [Artifacts](./artifacts/) - Output schemas for deployable artifacts

---

## Quick Start

### How to Use Genesis Agency

1. **Enter Keywords/Requirements**
   ```
   User: "water distribution network monitoring in Nairobi"
   ```

2. **Iterative Research Session** (using research.prompt.md)
   - Agent asks targeted questions
   - Human provides domain knowledge
   - Specification emerges through dialogue

3. **Review & Refine**
   - Introduction drafted
   - Goals structured
   - Work items decomposed
   - Roles defined
   - RACI matrix generated

4. **Generate Artifacts**
   - Agency specification (JSON/YAML)
   - Database initialization scripts
   - Deployment manifest
   - Documentation templates

5. **Deploy New Agency**
   - Create agency record in database
   - Initialize agency-specific database
   - Set up initial specification
   - Ready for further customization

---

## Integration Points

### Leverages Existing Platform Capabilities

- **AI Builder Services** (`internal/builder/ai/*`)
  - GoalsBuilder - Generate goals from requirements
  - WorkItemsBuilder - Generate work items from goals
  - RolesBuilder - Define agent roles
  - WorkflowsBuilder - Create workflows
  - IntroductionBuilder - Draft introduction sections

- **Agency Management** (`internal/agency/*`)
  - CreateAgency - Provision new agency
  - InitializeAgencyDatabase - Set up database
  - CreateSpecification - Initialize specification

- **Research Framework** (`.github/prompts/research.prompt.md`)
  - Structured Q&A methodology
  - Gap tracking
  - One-question-at-a-time exploration

---

## Development Status

- [x] Architecture design
- [x] Role definitions
- [ ] Work item specifications
- [ ] Workflow implementations
- [ ] Artifact schemas
- [ ] Database initialization
- [ ] End-to-end testing

---

## Related Documentation

- [MVP-042: AI-Powered Agency Creator](../MVP-042.md)
- [Research Prompt Framework](../../../../.github/prompts/research.prompt.md)
- [Agency Specifications](../../../2-SoftwareDesignAndArchitecture/agency-operation-framework/)
- [Builder Services](../../../2-SoftwareDesignAndArchitecture/builder-services/)
