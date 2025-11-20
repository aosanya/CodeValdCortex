---
mode: agent
---

# Start New Task

Follow the **mandatory task startup process** for project tasks:

## Task Startup Process (MANDATORY)

1. **Select next task from `mvp.md`**
   - Choose based on priority (P0 → P1 → P2)
   - Verify all dependencies are completed (marked with ~~MVP-XXX~~)
   - Check task is "Not Started" status

2. **Read detailed specification**
   - **Domain-based documentation**: Tasks are organized by problem domain in `documents/3-SofwareDevelopment/mvp-details/`
   - **🚨 CRITICAL FILE SIZE LIMITS**:
     - **Single file domains**: MAX 500 lines TOTAL
     - **Folder-based domains**: README.md MAX 300 lines, task files MAX 200 lines each
     - **⚠️ MANDATORY REFACTOR TRIGGER**: If domain file >500 lines OR README >300 lines:
       1. **STOP IMMEDIATELY** - Do not add more content
       2. **Create folder structure** first
       3. **Split existing content** into README + task files
       4. **Then proceed** with new task
   - **Find your task's domain**:
     - **Small domains** (2-4 tasks, <500 lines): Single file like `authentication.md`
     - **Large domains** (5+ tasks OR >500 lines): MUST use folder structure:
       ```
       work-items-integration/
       ├── README.md              # Overview, architecture (MAX 300 lines)
       ├── MVP-WI-001.md          # Gitea webhooks (MAX 200 lines)
       ├── MVP-WI-002.md          # API client (MAX 200 lines)
       ├── MVP-WI-003.md          # Agent-to-issue sync (MAX 200 lines)
       ├── MVP-WI-004.md          # PR automation (MAX 200 lines)
       ├── architecture/          # Optional: detailed designs
       │   ├── webhook-flow.md
       │   └── sync-architecture.md
       └── examples/              # Optional: code samples
           ├── webhook-payload.json
           └── pr-template.md
       ```
   - **Locate your task**: Search for `<!-- MVP-XXX -->` annotation
   - **Domain files are narrative documents**: Easy to read, straightforward, consumable
   - **🔄 REFACTOR WORKFLOW**:
     - If domain file >500 lines: Create folder, split into README + task files
     - If individual `MVP-XXX.md` files exist: Consolidate into domain folder
     - **Always refactor BEFORE adding new content**
   - **Folder structure template**:
     ```
     {domain-name}/
     ├── README.md              # Domain overview, architecture, task index
     ├── MVP-XXX.md             # Individual task specification
     ├── MVP-YYY.md             # Individual task specification
     ├── architecture/          # Optional: detailed technical designs
     │   └── *.md
     └── examples/              # Optional: code samples, configs
         └── *.{json,yaml,md}
     ```
   - **Key principles**:
     - **HARD LIMIT: 500 lines per file** - No exceptions
     - **README.md: MAX 300 lines** - Overview only, link to task files
     - **Task files: MAX 200 lines** - One task per file
     - Use subfolders (`architecture/`, `examples/`) to separate verbosity
     - Domain documentation must be narrative, not just task lists
   - Review all requirements, acceptance criteria, and technical specifications within the domain context
   - Understand how this task fits into the broader domain strategy

3. **Create feature branch**
   ```bash
   git checkout -b feature/MVP-XXX_description
   ```
   - Use exact format: `feature/MVP-XXX_description`
   - Description should be lowercase with underscores

4. **Read project guidelines**
   - Review `.github/instructions/rules.instructions.md`
   - Follow code quality standards (dart analyze, dart format)
   - Follow linting rules (const, final, no print statements)
   - Remember logging standards (MVP-XXX prefix)

5. **Create todo list**
   - Break down task into actionable steps
   - Use manage_todo_list tool to track progress
   - Mark items in-progress and completed as you work

## Pre-Implementation Checklist

Before starting implementation:
- [ ] Task selected from mvp.md based on priority and dependencies
- [ ] All dependency tasks are completed (~~MVP-XXX~~)
- [ ] Read domain documentation file (e.g., `work-items-integration.md`) and located task section
- [ ] Feature branch created: `feature/MVP-XXX_description`
- [ ] Reviewed code quality standards in rules.instructions.md
- [ ] Todo list created with implementation steps
- [ ] Understand acceptance criteria and validation requirements

## Domain Documentation Approach

**Philosophy**: Tasks are documented within domain-based narrative documents, not individual files per task.

**Benefits**:
- ✅ **Context**: Understand how task fits into broader domain strategy
- ✅ **Coherence**: Related tasks read as cohesive story, not isolated tickets
- ✅ **Efficiency**: Reduce documentation fragmentation
- ✅ **Onboarding**: New developers understand entire domain, not just one task

**Example Domain Structures**:
- **Small domain** (single file): `authentication.md` - 3 tasks, 450 lines total
- **Large domain** (folder):
  ```
  work-items-integration/
  ├── README.md           # Overview, architecture (280 lines)
  ├── MVP-WI-001.md       # Gitea webhooks (180 lines)
  ├── MVP-WI-002.md       # API client (150 lines)
  ├── MVP-WI-003.md       # Agent-to-issue sync (200 lines)
  ├── MVP-WI-004.md       # PR automation (190 lines)
  └── architecture/
      └── sync-flow.md    # Detailed flow diagrams
  ```
- **Another example**:
  ```
  agent-lifecycle/
  ├── README.md           # FSM overview, state diagrams (250 lines)
  ├── MVP-033.md          # Agent lifecycle FSM (200 lines)
  ├── MVP-034.md          # Run execution FSM (200 lines)
  ├── MVP-035.md          # Health & circuit breakers (180 lines)
  └── MVP-036.md          # Quarantine system (200 lines)
  ```

**Task Annotations in Folder Structure**:

Each task file starts with metadata:
```markdown
# MVP-WI-004: Pull Request Automation

**Domain**: Work Items Integration  
**Priority**: P0  
**Effort**: High  
**Dependencies**: MVP-WI-003 ✅

## Overview
[Task description...]

## Requirements
[Detailed requirements...]
```

README.md contains domain overview and task index:
```markdown
# Work Items Integration Domain

## Overview
[Domain narrative...]

## Architecture
[System design...]

## Task Index
- [MVP-WI-001: Gitea Webhooks](MVP-WI-001.md) - ✅ Complete
- [MVP-WI-002: API Client](MVP-WI-002.md) - ✅ Complete
- [MVP-WI-003: Agent Sync](MVP-WI-003.md) - ✅ Complete
- [MVP-WI-004: PR Automation](MVP-WI-004.md) - 📋 Not Started
```

**Finding Tasks**:
1. Check `mvp.md` for task's domain
2. Navigate to domain folder in `mvp-details/`
3. Open task file directly: `MVP-XXX.md`
4. Or read README.md for domain context first

## Development Standards

**Code Quality:**
- Run `dart analyze` before committing
- Run `dart format .` for consistent formatting
- Use const constructors where possible
- Mark fields as final if never reassigned
- Remove all print() statements (use logging framework)

**Logging (if needed):**
- Prefix all logs with task ID: `MVP-XXX-INFO:`, `MVP-XXX-ERROR:`
- Remove debug logs before committing

**Testing:**
- Write tests for business logic
- Run `flutter test` before completion
- Test on iOS and Android if UI-related

## Git Workflow

```bash
# Start new task
git checkout main
git pull  # Ensure latest changes
git checkout -b feature/MVP-XXX_description

# Regular development commits
git add .
git commit -m "Descriptive message"

# Continue until complete, then use "Complete Branch" prompt
```

## Success Criteria
- ✅ Next priority task identified with all dependencies met
- ✅ Specification document reviewed and understood
- ✅ Feature branch created with correct naming
- ✅ Todo list created for tracking implementation
- ✅ Ready to begin implementation following project standards
