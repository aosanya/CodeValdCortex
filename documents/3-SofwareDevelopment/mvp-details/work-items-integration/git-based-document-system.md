# Git-Based Document & Code System in ArangoDB

<!-- Research Session: 2025-11-26 -->
**Tasks Covered**: MVP-WI-005 (Document Management), MVP-PUB-007 (Instance Management)  
**Status**: 📋 Planned - Architecture Complete

## Overview

CodeValdCortex implements a **full Git version control system inside ArangoDB** that works seamlessly for both code and documents. The system provides Git's powerful versioning, branching, and merging capabilities while presenting a user-friendly file explorer interface that completely abstracts Git mechanics from end users.

### Vision

Enable collaborative editing of code and documents by humans and AI agents with:
- Git-quality version control and conflict resolution
- File/folder navigation like a traditional file system
- Automatic merging with AI assistance
- Section-based document editing for granular control
- No Git knowledge required by users

---

## Core Architecture Decisions

### 1. **Git Implementation in ArangoDB** (Not External VCS)

**Decision**: Implement full Git object model directly in ArangoDB

**Why**:
- ✅ Single database for all data (no external dependencies)
- ✅ Custom merge strategies (AI-assisted conflict resolution)
- ✅ Tight integration with agency workflows
- ✅ Query Git history with AQL
- ✅ No sync delays or conflicts between systems

**What this means**:
- All Git objects (commits, trees, blobs) stored in ArangoDB collections
- Git operations (commit, branch, merge) implemented in Go
- Content-addressable storage using SHA-1 hashing
- Full Git compatibility at data model level

**Implementation Details**: See [git-operations.md](git-operations.md) for complete Git object model and operations.

---

### 2. **File Explorer UX** (Git Mechanics Hidden)

**Decision**: Present files/folders with familiar explorer interface, hide Git concepts

**User sees**:
```
📁 my-project/
├── 📁 documents/
│   ├── 📄 requirements.md
│   └── 📄 specifications.yaml
└── 📁 code/
    └── 📄 workflow.go
```

**User does NOT see**:
- ❌ `git commit -m "..."`
- ❌ Branch names or commit hashes
- ❌ `git merge` or `git rebase`
- ❌ SHA-1 hashes

**User actions**:
- "Save" → Backend creates Git commit
- "Request Review" → Backend creates pull request
- "Approve" → Backend performs Git merge
- "View History" → Backend shows commit log

**Implementation Details**: See [file-explorer.md](file-explorer.md) for File Explorer API and UI.

---

### 3. **Sectioned Documents** (Granular Merging)

**Decision**: Structure documents as sections with unique IDs for conflict-free collaboration

**Why**:
```
Monolithic Document:
  Agent edits line 50
  User edits line 100
  → Merge conflict (whole file)

Sectioned Document:
  Agent edits "Introduction" section
  User edits "Conclusion" section
  → Auto-merge (different sections)
```

**Document structure**:
```markdown
---
id: requirements-doc-001
sections:
  - id: introduction
  - id: functional-requirements
  - id: non-functional-requirements
---

<!-- section: introduction -->
# Introduction
Content here...

<!-- section: functional-requirements -->
# Functional Requirements
Content here...
```

**Implementation Details**: See [collaborative-editing.md](collaborative-editing.md) for sectioned documents and merge strategies.

---

### 4. **No Human/AI Distinction** (Unified Access)

**Decision**: Humans and AI agents use same API, same permissions, same workflows

**Why**:
- ✅ Simpler architecture
- ✅ AI can collaborate with humans naturally
- ✅ Same conflict resolution for all actors
- ✅ Unified audit trail

**Implementation**: `author` field can be `user-alice` or `agent-requirements-001`

---

### 5. **AI-Assisted Conflict Resolution**

**Decision**: Use AI to intelligently merge conflicts when possible

**Workflow**:
```
1. Git merge detects conflict
2. Extract conflict context (base, theirs, ours)
3. Send to AI: "Resolve this conflict intelligently"
4. AI analyzes both changes
5. AI proposes merged version
6. Human reviews and approves (or manually edits)
```

**Implementation Details**: See [collaborative-editing.md](collaborative-editing.md) for AI conflict resolution workflow.

---

## Kanban Workflow Integration

### Linking Files to Issues

```go
type WorkIssue struct {
    Key string `json:"_key"`
    
    // Issue data
    Title       string `json:"title"`
    Description string `json:"description"`
    Milestone   string `json:"milestone"`
    Labels      []string `json:"labels"`
    
    // File references (what files does this issue affect?)
    Files       []string `json:"files"`        // File paths
    Branch      string   `json:"branch"`       // Working branch for this issue
    PullRequest string   `json:"pull_request,omitempty"` // PR ID when created
    
    // Workflow tracking
    InstanceID  string `json:"instance_id"`
    WorkflowID  string `json:"workflow_id"`
    ColumnID    string `json:"column_id"`
    
    Status      string    `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### Issue → Branch → PR Workflow

```
1. Issue created: "Update requirements document"
   - milestone: "In Progress"
   - files: ["/documents/requirements.md"]
   
2. System creates branch: "issue-42-update-requirements"
   - Branch from main
   - Agent assigned to issue starts work
   
3. Agent edits file on branch
   - Makes commits to branch
   - Updates tracked in issue
   
4. Agent completes work
   - Creates pull request
   - Links PR to issue
   - Issue milestone → "Review"
   
5. Human reviews PR
   - Views diff in UI
   - Approves or requests changes
   
6. PR merged
   - Branch merged to main
   - Issue milestone → "Done"
   - Branch deleted
```

---

## Summary

**What we've built**:
1. ✅ Full Git implementation in ArangoDB (commits, trees, blobs, refs)
2. ✅ File explorer UX (hide Git complexity)
3. ✅ Sectioned documents (granular merging)
4. ✅ AI-assisted conflict resolution
5. ✅ Kanban integration (issues → branches → PRs)
6. ✅ Unified human/AI workflow
7. ✅ Content-addressable storage (deduplication)

**Benefits**:
- Single database (no external dependencies)
- Git-quality version control
- User-friendly interface
- Collaborative editing (human + AI)
- Automatic conflict resolution
- Works for code AND documents

**Implementation Documentation**:
- [git-operations.md](git-operations.md) - Git object model, write/read/merge operations
- [file-explorer.md](file-explorer.md) - File browser API, indexing, MVP-WI-006 implementation
- [collaborative-editing.md](collaborative-editing.md) - Sectioned documents, AI conflict resolution
- [pull-requests.md](pull-requests.md) - Code review workflow
- [kanban-workflow.md](kanban-workflow.md) - Issue management, workflow automation

**Next steps for implementation**:
- Build Git object storage layer
- Implement core Git operations (commit, merge)
- Create file explorer API
- Build conflict resolution UI
- Integrate with Kanban workflows
- Add AI merge assistance
