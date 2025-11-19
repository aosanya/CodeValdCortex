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
   - **Find your task's domain file** (e.g., `work-items-integration.md`, `agency-designer.md`, `authentication.md`)
   - **Within the domain file**, locate your task using annotations like `<!-- MVP-XXX -->` or `## MVP-XXX: Task Title`
   - **Domain files are narrative documents** that read as cohesive guides, with tasks as annotated sections
   - **If domain file doesn't exist:**
     - Create new domain-based file: `{domain-name}.md` (e.g., `compliance-framework.md`)
     - Write as a **continuous narrative document** covering the entire domain
     - Annotate task sections with `<!-- MVP-XXX -->` comments for reference
     - Use this template structure:
       ```markdown
       # [Domain Name] (e.g., Work Items Integration)
       
       ## Overview
       [Narrative introduction to the entire domain]
       
       ## Architecture
       [Overall architecture for this domain]
       
       <!-- MVP-XXX -->
       ## [Task Title] (MVP-XXX)
       
       [Task description integrated into the narrative flow]
       
       **Priority**: [P0/P1/P2]  
       **Effort**: [Low/Medium/High]  
       **Dependencies**: [MVP-XXX, MVP-YYY]
       
       ### Objectives
       - [Objective 1]
       
       ### Requirements
       [Requirements in narrative form]
       
       ### Acceptance Criteria
       - [ ] [Criterion 1]
       
       ### Technical Details
       [Implementation details that flow with the document]
       
       <!-- /MVP-XXX -->
       
       [Continue narrative to next task or section...]
       ```
   - **Key principle**: Domain files should be readable as standalone documents, not just task lists
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

**Example Domain Files**:
- `work-items-integration.md` - Covers MVP-WI-001 through MVP-WI-004 (Gitea webhooks, API client, sync, PR automation)
- `agency-designer.md` - Covers MVP-046, MVP-047, MVP-042 (admin UI, export, AI creator)
- `authentication.md` - Covers MVP-026, MVP-027, MVP-028 (user auth, security, RBAC)
- `agent-lifecycle.md` - Covers MVP-033 through MVP-036 (FSM, runs, health, quarantine)
- `a2a-protocol.md` - Covers MVP-A2A-000 through MVP-A2A-009 (entire A2A integration)

**Task Annotations**:
```markdown
<!-- MVP-WI-001 -->
## Gitea Webhook Integration (MVP-WI-001)

The webhook integration forms the foundation of our work tracking system...

**Priority**: P0  
**Effort**: Medium  
**Dependencies**: None

[Narrative continues with objectives, requirements, technical details...]

<!-- /MVP-WI-001 -->
```

**Finding Tasks in Domain Files**:
1. Check `mvp.md` for task's domain category
2. Open corresponding domain file in `mvp-details/`
3. Search for `<!-- MVP-XXX -->` annotation
4. Read entire section for full context

**Creating New Domain Files**:
- Group related tasks by problem domain (not by tech stack or layer)
- Write as continuous narrative that explains the domain
- Include architecture diagrams, data flows, design decisions
- Annotate task boundaries with HTML comments
- Make it readable start-to-finish, not just searchable

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
