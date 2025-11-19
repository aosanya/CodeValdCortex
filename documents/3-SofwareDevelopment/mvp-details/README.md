# MVP Task Documentation - Domain-Based Organization

## Overview

This directory contains **domain-based task documentation** where related tasks are grouped into cohesive narrative documents rather than isolated per-task files.

## Philosophy

**Traditional approach** (Deprecated):
- ❌ One file per task: `MVP-001.md`, `MVP-002.md`, `MVP-003.md`
- ❌ Fragmented context: each file stands alone
- ❌ Repetitive boilerplate: architecture repeated in multiple files
- ❌ Difficult to understand relationships between tasks

**Domain-based approach** (Current):
- ✅ One file per problem domain: `work-items-integration.md`, `agency-designer.md`
- ✅ Narrative flow: reads as cohesive story
- ✅ Shared context: architecture explained once for entire domain
- ✅ Clear relationships: dependencies and flow visible within document

**Keep it simple and consumable**:
- ✅ **Max 500 lines per file**: If domain file exceeds 500 lines, create a folder
- ✅ **Easy to read**: Straightforward language, clear structure, no unnecessary complexity
- ✅ **Separate verbosity**: Move detailed code examples to separate files in folder
- ✅ **Well-organized folders**: Use subfolders for architecture, examples, schemas

## File Size Guidelines

### Single File (Preferred for small domains)
**Use when**: Domain has 2-4 related tasks, total content < 500 lines

```
mvp-details/
├── work-items-integration.md  (350 lines - good!)
├── agency-designer.md         (280 lines - good!)
└── authentication.md          (150 lines - good!)
```

### Folder Structure (Required for large domains)
**Use when**: Domain has 5+ tasks OR single file would exceed 500 lines

```
mvp-details/
├── platform-infrastructure/
│   ├── README.md              (Overview, architecture, task list - MAX 300 lines)
│   ├── deployment.md          (MVP-037 Deployment Rollouts - concise)
│   ├── namespace-isolation.md (MVP-038 Namespace Isolation - concise)
│   ├── rbac.md               (MVP-039 Organization & RBAC - concise)
│   ├── billing.md            (MVP-040 Billing & Metering - concise)
│   ├── architecture/
│   │   ├── deployment-strategies.md  (Detailed diagrams, flows)
│   │   └── rbac-model.md            (Detailed permission matrix)
│   └── examples/
│       ├── deployment-yaml.md       (Sample Kubernetes manifests)
│       ├── rbac-policies.md         (Example RBAC configurations)
│       └── billing-queries.md       (Example metering queries)
```

### Folder Organization Rules

**README.md** (Entry point - MAX 300 lines):
- Domain overview and goals
- High-level architecture diagram
- Task list with links to individual task files
- Quick navigation to detailed docs

**Individual Task Files** (MAX 200 lines each):
- One task per file
- Concise description, objectives, acceptance criteria
- Link to detailed architecture/examples as needed
- Use annotations: `<!-- MVP-XXX -->`

**architecture/** subfolder:
- Detailed architecture documents
- Complex diagrams and data flows
- Design decisions and rationale
- System integration patterns

**examples/** subfolder:
- Code snippets and templates
- Configuration examples
- API request/response samples
- Step-by-step tutorials

**schemas/** subfolder (if needed):
- JSON schemas
- Database schemas
- API specifications
- Data models

### Example: Platform Infrastructure Folder

```markdown
# platform-infrastructure/README.md (280 lines)

# Platform Infrastructure

## Overview
[Brief domain introduction - 50 lines]

## Architecture
[High-level diagram and concepts - 100 lines]

## Tasks in This Domain

1. [MVP-037: Deployment Rollouts](deployment.md)
2. [MVP-038: Namespace Isolation](namespace-isolation.md)
3. [MVP-039: Organization & RBAC](rbac.md)
4. [MVP-040: Billing & Metering](billing.md)
5. [MVP-041: Multi-tenancy Hardening](multi-tenancy.md)

## Detailed Documentation

- [Deployment Strategies](architecture/deployment-strategies.md)
- [RBAC Model](architecture/rbac-model.md)
- [Example Configurations](examples/)
```

```markdown
# platform-infrastructure/deployment.md (180 lines)

<!-- MVP-037 -->
# Deployment Rollouts (MVP-037)

[Concise task description]

**Priority**: P1  
**Effort**: High  
**Dependencies**: MVP-036  
**Status**: Not Started

## Objectives
[Brief list]

## Requirements
[Concise requirements]

## Acceptance Criteria
- [ ] [Criteria]

## Architecture Overview
See [Deployment Strategies](architecture/deployment-strategies.md) for detailed diagrams.

## Key Implementation Points
[High-level implementation notes]

**Key Files**: [List]

**Examples**: See [examples/deployment-yaml.md](examples/deployment-yaml.md)

<!-- /MVP-037 -->
```

## Document Structure

Domain files follow this narrative template:

```markdown
# [Domain Name]

## Overview
[Introduction to the problem domain and why it exists]

## Architecture
[Overall architecture for this domain - diagrams, data flows, key concepts]

## Goals
[High-level goals for this domain]

<!-- MVP-XXX -->
## [Task Title] (MVP-XXX)

[Task description that flows naturally from the previous narrative]

**Priority**: P0/P1/P2  
**Effort**: Low/Medium/High  
**Dependencies**: MVP-YYY, MVP-ZZZ  
**Status**: Not Started / In Progress / Completed

### Objectives
- [Specific objective 1]
- [Specific objective 2]

### Requirements
[Requirements written in narrative form, not bullet points if possible]

### Acceptance Criteria
- [ ] [Testable criterion 1]
- [ ] [Testable criterion 2]

### Technical Implementation
[Code examples, package structure, key decisions]

**Key Files**:
- `path/to/file.go` - Description
- `path/to/other.go` - Description

<!-- /MVP-XXX -->

[Narrative continues to next task...]
```

## Current Domain Files

### Work Items Integration (`work-items-integration.md`)
**Tasks**: MVP-WI-001, MVP-WI-002, MVP-WI-003, MVP-WI-004

Covers the entire work tracking system integration:
- MVP-WI-001: Gitea Webhook Integration
- MVP-WI-002: Gitea API Client
- MVP-WI-003: Agent-to-Issue Sync
- MVP-WI-004: Pull Request Automation

**Why one file?**  
These tasks form a complete bidirectional sync system between Gitea and CodeValdCortex agents. Understanding the webhook ingestion (WI-001) requires context about how agents will write back (WI-003), and how PRs tie everything together (WI-004).

### Agency Designer (`agency-designer.md`)
**Tasks**: MVP-046, MVP-047, MVP-042

Covers the agency design interface:
- MVP-046: Agency Admin & Configuration
- MVP-047: Export System (PDF, Markdown, JSON)
- MVP-042: AI-Powered Agency Creator

**Why one file?**  
These are all different aspects of the agency design experience. They share UI patterns, data models, and user workflows.

### Agent Lifecycle (`agent-lifecycle.md`)
**Tasks**: MVP-033, MVP-034, MVP-035, MVP-036

Covers agent operational states and health:
- MVP-033: Agent Lifecycle FSM
- MVP-034: Run Execution FSM
- MVP-035: Health & Circuit Breakers
- MVP-036: Quarantine System

**Why one file?**  
These form a progression: lifecycle states → run states → health monitoring → quarantine. Understanding quarantine requires understanding health checks, which requires understanding lifecycle states.

### Authentication & Security (`authentication.md`)
**Tasks**: MVP-026, MVP-027, MVP-028

Covers user authentication and authorization:
- MVP-026: Basic User Authentication
- MVP-027: Security Implementation
- MVP-028: Access Control System

**Why one file?**  
These build on each other: auth → secure headers → RBAC. They share session management, user models, and security patterns.

### A2A Protocol Integration (`a2a-protocol.md`)
**Tasks**: MVP-A2A-000 through MVP-A2A-009

Covers the complete Agent-to-Agent protocol:
- SDK integration, agent cards, registry, gateway, delegation, etc.

**Why one file?**  
This is a cohesive protocol integration. Each task builds on the previous, and understanding the gateway requires understanding agent cards, which requires understanding the SDK wrapper.

## Finding Tasks

### Method 1: Search by Task ID
```bash
grep -r "<!-- MVP-XXX -->" documents/3-SofwareDevelopment/mvp-details/
```

### Method 2: Check mvp.md for Domain Reference
Look at the "Details" column in `mvp.md` task table - it will reference the domain file.

### Method 3: Browse by Topic
Open the domain file that matches your topic area and read through the narrative.

## Creating New Domain Files

When creating a new domain file:

1. **Group related tasks** - minimum 2-3 tasks that share context
2. **Write narrative introduction** - explain the domain before diving into tasks
3. **Use HTML comment annotations** - `<!-- MVP-XXX -->` and `<!-- /MVP-XXX -->`
4. **Maintain flow** - each task should flow naturally from the previous
5. **Shared architecture** - put common diagrams/concepts at domain level, not per-task
6. **Cross-reference** - link to related tasks within the narrative

## Legacy Individual Task Files

Some tasks still have individual files (e.g., `MVP-WI-001.md`). These are:

1. **In-progress tasks** - Being converted to domain format
2. **Standalone tasks** - Don't fit naturally into a domain group
3. **Very complex tasks** - Warrant their own detailed documentation

**Migration strategy**: As we work on these tasks, we migrate them into domain files when appropriate.

## Benefits of Domain-Based Docs

1. **Better Understanding**: See how task fits into larger picture
2. **Less Duplication**: Architecture explained once, not repeated
3. **Easier Review**: Reviewers see full context, not isolated change
4. **Faster Onboarding**: New developers read entire domain, not piecemeal tasks
5. **Natural Dependencies**: Dependencies obvious from narrative flow
6. **Reduced File Count**: 10 tasks → 3 domain files instead of 10 individual files

## Example Reading Experience

**Old approach** (reading 4 separate files):
```
MVP-WI-001.md → "Persist webhooks to ArangoDB"
  (Why? What happens next? How does it fit?)
  
MVP-WI-002.md → "Build API client"
  (How does this relate to webhooks?)
  
MVP-WI-003.md → "Sync agents to issues"
  (How do webhooks connect to this?)
  
MVP-WI-004.md → "Automate PRs"
  (How does the whole system work together?)
```

**New approach** (reading one file):
```
work-items-integration.md:
  
  "Work items integration enables bidirectional sync between 
   external work tracking systems and CodeValdCortex agents...
   
   First, we ingest webhooks (MVP-WI-001), persisting to ArangoDB.
   
   Then, we build an API client (MVP-WI-002) for agents to write
   back progress, comments, and status updates...
   
   The sync layer (MVP-WI-003) ties these together, ensuring
   agent state changes flow back to issues...
   
   Finally, PR automation (MVP-WI-004) closes the loop by creating
   PRs, linking them to issues, and tracking completion..."
```

## Guidelines for Updates

When updating a task within a domain file:

1. **Find annotation**: Search for `<!-- MVP-XXX -->`
2. **Maintain narrative**: Updates should flow with surrounding text
3. **Update status**: Mark `**Status**: ✅ Completed YYYY-MM-DD` when done
4. **Add implementation notes**: Show what was actually built
5. **Link to code**: Reference key files created
6. **Add coding session link**: Reference the detailed coding session document
7. **Add implementation history table**: Track all coding sessions for this task
8. **Preserve flow**: Don't break the reading experience for other tasks

### Coding Session Reference Format

After completing a task, add a coding session reference and implementation history:

```markdown
**Status**: ✅ Completed 2025-11-19

**Coding Session**: [MVP-XXX_description](../coding_sessions/MVP-XXX_description.md)

### Implementation History

| Date | Session | Summary |
|------|---------|---------|
| 2025-11-19 | [MVP-XXX_description](../coding_sessions/MVP-XXX_description.md) | Brief summary of what was implemented |
| 2025-11-20 | [MVP-XXX_bugfix](../coding_sessions/MVP-XXX_bugfix.md) | Fixed issue with XYZ |
```

This allows tracking multiple coding sessions if a task requires iterations or bug fixes.

## Questions?

- **"My task doesn't fit any domain"** → Create individual file for now, we'll group later
- **"Domain file is getting too long"** → Consider splitting into sub-domains
- **"Task spans multiple domains"** → Pick primary domain, cross-reference others
- **"How do I update a task?"** → See `.github/prompts/finish-task.prompt.md`

---

**Remember**: Documentation is for humans. Optimize for reading comprehension and understanding, not just searchability.
