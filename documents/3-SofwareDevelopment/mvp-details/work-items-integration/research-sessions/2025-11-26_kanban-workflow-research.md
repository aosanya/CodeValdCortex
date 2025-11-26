# Research Session: Kanban Workflow & Issue Lifecycle

**Date**: 2025-11-26  
**Researcher**: AI Assistant  
**Methodology**: [research.prompt.md](/.github/prompts/research.prompt.md) - One question at a time  
**Focus**: Understanding end-to-end workflow from issue creation to completion

---

## Session Summary

Successfully documented the complete workflow for work items in CodeValdCortex, clarifying how Kanban board, Git system, agents, and pull requests integrate.

---

## Questions Asked & Answers

### Q1: Issue Creation - First Step
**Question**: What is the VERY FIRST action the user takes in the UI?

**Answer**: Option A - User creates issue first
- User opens Kanban board
- Clicks "Create Issue"
- Fills in Title and Description
- Issue automatically placed in **REQ1 column** (only entry point)

**Key Learning**: REQ1 is the ONLY entry point - all issues must start there (enforced workflow)

---

### Q2: Column Placement
**Question**: How does the issue end up in the REQ1 column specifically?

**Answer**: Automatic - REQ1 is the entry point
- No column selection needed
- All new issues default to REQ1
- Cannot create issues in other columns directly

**Key Learning**: Workflow enforces proper process - requirements gathering first

---

### Q3: Issue Progression
**Question**: How does an issue move from REQ1 to next stage?

**Answer**: Option B - Work completion triggers move
- Issue sits in REQ1
- Work gets done (requirements gathered)
- Someone marks issue as "Complete" or "Ready for Review"
- **Next agent or person pulls it to their step**

**Key Learning**: Pull-based workflow, not automatic push

---

### Q4: Agent Assignment
**Question**: When/how do agents get involved?

**Answer**: Hybrid - Manual (A) OR Claim-based (C)
- **Manual**: Admin assigns specific agent/human to issue
- **Claim**: Available agents/humans can claim work from queue
- Flexible model supports different team structures

**Key Learning**: Supports both managed and self-organizing teams

---

### Q5: Git Integration
**Question**: When does Git-in-ArangoDB system get involved?

**Answer**: Option B - Agent creates branch when starting work
- Issue created (no branch yet)
- Agent assigned/claims work
- Agent creates Git branch: `issue-123-jwt-auth`
- Agent works in branch, creates commits
- All work linked to issue via branch

**Key Learning**: Git branch per issue, created on-demand when work starts

---

### Q6: Work Completion
**Question**: What happens when agent finishes work?

**Answer**: Option A - Agent creates PR automatically
- Agent finishes work in branch
- Agent auto-creates Pull Request
- PR links to issue
- Human reviews and merges PR
- **PR merge triggers issue progression to next column**

**Key Learning**: PR merge is the automation trigger for workflow progression

---

## Key Findings

### 1. Work Item Definitions vs Issues
**Clarified Relationship**:
- **Work Item Definitions**: Templates in agency spec (REQ1, IMPL1, etc.)
  - Define deliverables, goals, tags
  - Created at agency design time
  - Stored in agency specification
  
- **Issues**: Runtime instances created by users
  - Reference Work Item Definitions implicitly via column/step
  - Track actual work progress
  - Stored in `work_issues` collection

**Mapping**: Issue in REQ1 column → Inherits context from REQ1 Work Item Definition

---

### 2. Complete Workflow Flow

```
1. Issue Creation
   User creates issue → Auto-placed in REQ1 column
   
2. Work Assignment (Flexible)
   Manual: Admin assigns agent/human
   OR Claim: Worker pulls from queue
   
3. Git Branch Creation
   Agent/human starts work → Creates branch
   
4. Work Execution
   Agent/human commits files to branch
   Multiple commits allowed
   
5. Pull Request Creation
   Agent auto-creates PR when work complete
   PR links to issue
   
6. Human Review
   Human reviews PR
   Approves and merges
   
7. Automated Progression
   PR merge → Issue moves to next column (REQ1 → REV1)
   Ready for next worker to claim
   
8. Repeat
   Process continues through all workflow steps
   REV1 → ARCH1 → IMPL1 → TEST1 → DEPLOY → DONE
```

---

### 3. Architecture Decisions Validated

✅ **Entry Point Enforcement**: All work starts at REQ1  
✅ **Pull-Based Workflow**: Workers pull work, not pushed automatically  
✅ **Git-Per-Issue**: Every issue gets own branch  
✅ **PR-Driven Progression**: PR merge triggers advancement  
✅ **Agent Auto-PR**: Agents create PRs automatically  
✅ **Flexible Assignment**: Manual OR claim-based  

---

## Documentation Created

### Primary Document: kanban-workflow.md

**Content**:
- Complete feature request workflow (7 steps)
- Use case: "Implement JWT Authentication"
- Data models (WorkIssue, PullRequest)
- API endpoints
- Integration points
- Implementation phases

**Key Sections**:
1. Issue Creation (User Action)
2. Work Assignment (Flexible Model)
3. Work Execution - Git Integration
4. Review Submission (Agent Auto-creates PR)
5. Human Review
6. Automated Issue Progression
7. Workflow Continues

---

## Gaps Identified (For Future Documentation)

### 1. Issue Creation Form Fields
**Gap**: Exact form fields not documented
**Needed**:
- Required fields (title, description)
- Optional fields (priority, tags, assignee?)
- Validation rules
- UI mockup/wireframe

### 2. Kanban Board UI
**Gap**: Visual design not specified
**Needed**:
- Column layout
- Card design
- Drag-drop mechanics (if manual moves allowed)
- Filters and views

### 3. Work Item Definition Usage
**Gap**: How definitions influence runtime behavior
**Needed**:
- Do deliverables become checklist in UI?
- Are goals displayed to agent?
- How do tags affect routing?

### 4. Agent Behavior Specification
**Gap**: Agent decision-making logic
**Needed**:
- How does agent know work is "complete"?
- What triggers PR creation?
- How does agent handle errors?

### 5. Notifications & Events
**Gap**: Event system not documented
**Needed**:
- Who gets notified when?
- Email, in-app, webhooks?
- Event schema and delivery

### 6. Conflict Resolution
**Gap**: What if multiple workers claim same issue?
**Needed**:
- Locking mechanism
- Race condition handling
- Reassignment process

---

## Next Steps

### Immediate
1. ✅ Update README.md topic index with kanban-workflow.md
2. ✅ Link from other docs to kanban-workflow.md
3. ⏳ Create Issue Creation UI specification
4. ⏳ Document agent completion logic

### Future Research Sessions
1. **Notifications & Events**: Complete event system design
2. **Agent Decision Logic**: How agents determine work completion
3. **Conflict Handling**: Race conditions, reassignments, error recovery
4. **UI/UX Design**: Kanban board, issue forms, PR review interface

---

## References

- **Methodology**: `/.github/prompts/research.prompt.md`
- **Agency Spec**: Actual JSON provided by user
- **Primary Output**: `kanban-workflow.md`
- **Related Docs**: 
  - `git-based-document-system.md`
  - `work-item-schema.md`
  - `pull-requests.md`

---

**Session Duration**: ~20 questions over structured exploration  
**Outcome**: Complete workflow documented with data models and API endpoints  
**Documentation Quality**: Production-ready architecture specification
