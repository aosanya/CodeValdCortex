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
   - Task details are found in `documents/3-SofwareDevelopment/mvp-details/`
   - Open `documents/3-SofwareDevelopment/mvp-details/MVP-XXX.md`
   - **If details file doesn't exist:**
     - Search `documents/2-SoftwareDesignAndArchitecture/` for relevant context
     - Review similar tasks in `documents/3-SofwareDevelopment/coding_sessions/`
     - Create new `MVP-XXX.md` in `mvp-details/` folder using the template:
       ```markdown
       # MVP-XXX: [Task Title]
       
       ## Overview
       **Priority**: [P0/P1/P2]  
       **Effort**: [Low/Medium/High]  
       **Skills Required**: [List skills]  
       **Dependencies**: [MVP-XXX, MVP-YYY]  
       **Status**: Not Started
       
       ## Description
       [Detailed description from mvp.md or architecture docs]
       
       ## Objectives
       - [Key objective 1]
       - [Key objective 2]
       
       ## Requirements
       [Functional and technical requirements]
       
       ## Acceptance Criteria
       - [ ] [Criterion 1]
       - [ ] [Criterion 2]
       
       ## Technical Specifications
       [Implementation details, architecture decisions]
       ```
   - Review all requirements, acceptance criteria, and technical specifications
   - Understand objectives and constraints

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
- [ ] Read MVP-XXX.md specification document thoroughly
- [ ] Feature branch created: `feature/MVP-XXX_description`
- [ ] Reviewed code quality standards in rules.instructions.md
- [ ] Todo list created with implementation steps
- [ ] Understand acceptance criteria and validation requirements

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
