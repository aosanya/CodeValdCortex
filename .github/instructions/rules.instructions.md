---
applyTo: '**'
---

# Code Structure Rules

## Template-First Architecture

**Always prefer `.templ` files over JavaScript for HTML generation.**

- HTML markup should be defined in `.templ` files (using the templ template engine)
- JavaScript should only handle:
  - Event handling
  - Data fetching
  - DOM manipulation (show/hide, updates)
  - State management
- Do NOT generate HTML strings in JavaScript files
- Pre-render all content sections in templates, then toggle visibility with JavaScript
- Benefits:
  - Type safety
  - Server-side rendering capability
  - Better maintainability
  - Separation of concerns

## CSS and Styling

**Minimize custom CSS by leveraging Bulma CSS framework.**

- Use Bulma's built-in classes whenever possible
- Only create custom CSS when Bulma doesn't provide the needed styling
- Keep custom CSS files minimal and focused
- Prefer Bulma utility classes over custom styles
- Benefits:
  - Consistent design language
  - Less CSS to maintain
  - Faster development
  - Better responsive design out of the box

## Task Management and Workflow

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

### Task Completion Process (MANDATORY)
1. **Task Assignment**: Pick tasks based on priority (P0 first) and dependencies
2. **Implementation**: Update "Status" column as work progresses (Not Started → In Progress → Testing → Complete)
3. **Completion Process**:
   - Create detailed coding session document in `coding_sessions/` using format: `{TaskID}_{description}.md`
   - Add completed task to summary table in `mvp_done.md` with completion date
   - Remove completed task from active `mvp.md` file
   - Update any dependent task references
   - **Update README.md roadmap**: For milestone/major feature completion, update the roadmap section with brief, concise progress notes (move items from planned to completed, update current focus)
   - Merge feature branch to main (see Branch Management above)
4. **Dependencies**: Ensure prerequisite tasks are completed before starting dependent work

### README.md Update Guidelines
- **When to Update**: Major milestones, completed feature groups, significant architecture changes
- **Keep Changes Brief**: Update roadmap status, current focus, and key capabilities only
- **Areas to Update**: 
  - Roadmap section (✅ completed items, 🔄 current focus)
  - Key Features (if new major capabilities added)
  - Current development focus and next milestones
- **Avoid**: Detailed technical changes, minor bug fixes, work-in-progress updates

### Repository Structure
```
/workspaces/CodeValdCortex/
├── documents/3-SofwareDevelopment/
│   ├── mvp.md                    # Active tasks only
│   ├── mvp_done.md              # Completed tasks archive
│   └── coding_sessions/         # Detailed implementation logs
├── [project code structure]     # Implementation code
└── [other project folders]      # Additional project resources
```

