---
agent: agent
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
   
   - **🔄 MANDATORY REFACTOR WORKFLOW** (ENFORCE BEFORE ADDING ANY NEW CONTENT):
     ```
     BEFORE writing new task documentation:
     1. CHECK file size: wc -l {domain-file}.md
     2. IF >500 lines OR individual MVP-XXX.md files exist:
        a. CREATE folder: mvp-details/{domain-name}/
        b. CREATE README.md (overview, architecture, task index)
        c. SPLIT content by TOPIC (NOT by task ID):
           - Group related tasks into topic files
           - Examples: webhooks.md, authentication.md, state-machines.md
        d. MOVE architecture diagrams → architecture/ subfolder
        e. MOVE examples → examples/ subfolder
     3. ONLY THEN add new task content to appropriate topic file
     ```
   
   - **🛑 STOP CONDITIONS** (Do NOT proceed until these are fixed):
     - ❌ Domain file exceeds 500 lines → **MUST refactor first**
     - ❌ README.md exceeds 300 lines → **MUST split content**
     - ❌ Individual `MVP-XXX.md` files exist → **MUST consolidate by topic**
     - ❌ Task file exceeds 200 lines → **MUST split into subtopics**
   
   - **Find your task's domain**:
     - **Small domains** (2-4 tasks, <500 lines): Single file like `authentication.md`
     - **Large domains** (5+ tasks OR >500 lines): MUST use folder structure
   
   - **📁 FOLDER STRUCTURE TEMPLATE** (Use when domain >500 lines):
     ```
     {domain-name}/
     ├── README.md              # Domain overview, architecture, task index (MAX 300 lines)
     ├── {topic-1}.md           # Topic-based file grouping related tasks (MAX 500 lines)
     ├── {topic-2}.md           # Topic-based file grouping related tasks (MAX 500 lines)
     ├── {topic-3}.md           # Topic-based file grouping related tasks (MAX 500 lines)
     ├── architecture/          # Optional: detailed technical designs
     │   ├── flow-diagrams.md
     │   ├── data-models.md
     │   └── state-machines.md
     └── examples/              # Optional: code samples, configs
         ├── sample-configs.yaml
         ├── example-payloads.json
         └── api-examples.md
     ```
   
   - **🏷️ FILE NAMING PRINCIPLES** (CRITICAL):
     ```
     ✅ CORRECT (topic-based):
        - webhooks.md            (all webhook tasks together)
        - authentication.md      (login + OAuth + RBAC)
        - state-machines.md      (agent FSM + run FSM)
        - api-client.md          (all API client tasks)
     
     ❌ WRONG (task ID-based):
        - MVP-001.md             (DON'T use task IDs as filenames)
        - task-123.md            (DON'T use generic task numbers)
        - feature-auth.md        (DON'T use vague names)
     
     📋 GROUPING RULES:
        - If 2+ tasks cover same TOPIC → ONE file
        - If tasks share same COMPONENT → ONE file
        - If tasks are sequential PHASES → ONE file
        - Group by CONCEPT, not by task ID
     ```
   - **Locate your task**: Search for `<!-- MVP-XXX -->` annotation or check task index in README.md
   - **Domain files are narrative documents**: Easy to read, straightforward, consumable - NOT just task lists
   - Review all requirements, acceptance criteria, and technical specifications within the domain context
   - Understand how this task fits into the broader domain strategy

3. **Create feature branch**
   - **🚨 ALWAYS branch from `dev`, NOT `main`**
   - Use exact format: `feature/MVP-XXX_description`
   - Description should be lowercase with underscores
   - **🚨 MULTI-REPO WORKSPACE**: Create feature branch in EACH repository from `dev`:
     ```bash
     # For CodeValdCortex
     cd /workspaces/CodeValdCortex
     git checkout dev
     git pull origin dev  # Ensure latest dev changes
     git checkout -b feature/MVP-XXX_description
     
     # For CodeValdFortex
     cd /workspaces/CodeValdFortex
     git checkout dev
     git pull origin dev  # Ensure latest dev changes
     git checkout -b feature/MVP-XXX_description
     ```
   - Keep branches synchronized across repos when working on integrated features
   - All repos must be on feature branch before starting implementation
   - **Branch hierarchy**: `main` (stable) ← `dev` (integration) ← `feature/MVP-XXX` (your work)

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
- [ ] **🔄 REFACTOR CHECK COMPLETED** (MANDATORY):
  - [ ] Checked domain file size: `wc -l mvp-details/{domain}.md`
  - [ ] If >500 lines: Created folder structure and split by topic
  - [ ] If individual MVP-XXX.md files exist: Consolidated into topic-based files
  - [ ] All files comply with limits (README <300, topic files <500)
  - [ ] File names use topics (webhooks.md) NOT task IDs (MVP-001.md)
- [ ] Read domain documentation file (e.g., `work-items-integration/`) and located task section
- [ ] Feature branch created in ALL workspace repos: `feature/MVP-XXX_description`
  - [ ] CodeValdCortex: `feature/MVP-XXX_description` created
  - [ ] CodeValdFortex: `feature/MVP-XXX_description` created
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
- [Webhooks](webhooks.md) - MVP-WI-001 ✅ Complete
- [API Client](api-client.md) - MVP-WI-002 ✅ Complete  
- [Synchronization](synchronization.md) - MVP-WI-003 ✅ Complete
- [Pull Requests](pull-requests.md) - MVP-WI-004 📋 Not Started
```

**Finding Tasks**:
1. Check `mvp.md` for task's domain
2. Navigate to domain folder in `mvp-details/`
3. Open topic file (e.g., `webhooks.md`, `authentication.md`)
4. Search for `<!-- MVP-XXX -->` annotation within the topic file
5. Or read README.md for domain overview first

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

**Branch Strategy:**
- `main` - Stable production branch (merge from `dev` periodically)
- `dev` - Active development integration branch (merge feature branches here)
- `feature/MVP-XXX` - Individual task branches (created from `dev`)

```bash
# Start new task - Create feature branches from dev in ALL repos
# CodeValdCortex
cd /workspaces/CodeValdCortex
git checkout dev
git pull origin dev  # Ensure latest dev changes
git checkout -b feature/MVP-XXX_description

# CodeValdFortex
cd /workspaces/CodeValdFortex
git checkout dev
git pull origin dev  # Ensure latest dev changes
git checkout -b feature/MVP-XXX_description

# Regular development commits (in appropriate repo)
git add .
git commit -m "Descriptive message"

# When task complete - merge to dev (use "Complete Branch" prompt)
# Feature branches merge to dev, not main
git checkout dev
git merge feature/MVP-XXX_description --no-ff
git push origin dev

# Periodically: dev → main (for releases/milestones)
# This is done separately, not per task
git checkout main
git merge dev --no-ff
git push origin main
```

**Multi-Repo Notes:**
- Create feature branches with same name in both repos for consistency
- **ALWAYS branch from `dev`**, never from `main`
- Commit changes to the appropriate repository based on what you're modifying
- Frontend changes → CodeValdFortex
- Backend changes → CodeValdCortex
- Keep branches synchronized if task spans both repositories
- Merge completed features to `dev` branch
- `dev` → `main` merges happen at milestones (not per task)

## Success Criteria
- ✅ Next priority task identified with all dependencies met
- ✅ Specification document reviewed and understood
- ✅ Feature branch created with correct naming
- ✅ Todo list created for tracking implementation
- ✅ Ready to begin implementation following project standards
