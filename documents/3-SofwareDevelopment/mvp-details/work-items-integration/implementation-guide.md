# Git-Based Document System - Implementation Guide

**Architecture**: [git-based-document-system.md](./git-based-document-system.md)  
**Decisions**: [architecture/vcs-integration-decisions.md](./architecture/vcs-integration-decisions.md)  
**Status**: 📋 Ready for Implementation

## Implementation Phases

### Phase 1: Git Core Layer (Foundation)
**Duration**: 2-3 weeks  
**Dependencies**: None

#### Tasks
1. **Git Object Storage** (`internal/git/storage`)
   - [ ] Implement `GitObject` model and repository
   - [ ] Content-addressable storage (SHA-1 hashing)
   - [ ] Collections: `git_objects`, `git_refs`
   - [ ] Database initialization in agency setup

2. **Git Operations** (`internal/git/ops`)
   - [ ] `WriteBlob()` - Store file content
   - [ ] `WriteTree()` - Store directory structure
   - [ ] `Commit()` - Create version snapshot
   - [ ] `UpdateRef()` - Move branch pointers
   - [ ] `GetCommit()`, `GetTree()`, `GetBlob()` - Read operations

3. **Repository Management** (`internal/git/repository`)
   - [ ] `Repository` model and CRUD
   - [ ] Initialize repository with default branch (`main`)
   - [ ] Create initial commit (empty tree)

#### Acceptance Criteria
- Can create blob objects with SHA-1 keys
- Can create tree objects with file/directory entries
- Can create commits referencing trees and parents
- Can read back objects by SHA
- Content-addressable deduplication working

---

### Phase 2: File Explorer API (User Interface)
**Duration**: 2-3 weeks  
**Dependencies**: Phase 1

#### Tasks
1. **File Index** (`internal/git/fileindex`)
   - [ ] `FileIndex` model and repository
   - [ ] Build index from commit tree
   - [ ] Fast path lookups: `/documents/requirements.md` → blob SHA
   - [ ] Update index on commit

2. **File Browser API** (`internal/web/handlers/files`)
   - [ ] `GET /instances/{id}/files?path={path}` - List directory
   - [ ] `GET /instances/{id}/files/{path}` - Read file content
   - [ ] `PUT /instances/{id}/files/{path}` - Update file (creates commit)
   - [ ] `DELETE /instances/{id}/files/{path}` - Delete file
   - [ ] `POST /instances/{id}/files/{path}` - Create file

3. **UI Components** (`internal/web/pages/files`)
   - [ ] File explorer tree view (Alpine.js)
   - [ ] File editor (Monaco or CodeMirror)
   - [ ] "Save" button (triggers API)
   - [ ] File history timeline

#### Acceptance Criteria
- Can browse file tree in UI
- Can read file content
- Can edit and save files (creates Git commits)
- Can view file history

---

### Phase 3: Branching & Pull Requests (Collaboration)
**Duration**: 2-3 weeks  
**Dependencies**: Phase 2

#### Tasks
1. **Branch Operations** (`internal/git/branches`)
   - [ ] `CreateBranch(name, fromCommit)` - Create new branch
   - [ ] `ListBranches()` - Get all branches
   - [ ] `SwitchBranch(name)` - Change active branch
   - [ ] `DeleteBranch(name)` - Remove branch

2. **Pull Request Model** (`internal/git/pullrequest`)
   - [ ] `PullRequest` model and repository
   - [ ] Create PR from branch
   - [ ] List PRs (open, merged, closed)
   - [ ] Calculate diff between branches

3. **PR UI** (`internal/web/pages/pull-requests`)
   - [ ] "Request Review" button (creates PR from draft branch)
   - [ ] PR list view
   - [ ] PR detail view with diff
   - [ ] "Approve" button

#### Acceptance Criteria
- Can create draft branch automatically when editing
- Can create PR from draft branch
- Can view PR with file diffs
- Can approve PR (without merge yet)

---

### Phase 4: Merge & Conflict Resolution (Git Core)
**Duration**: 3-4 weeks  
**Dependencies**: Phase 3

#### Tasks
1. **Three-Way Merge** (`internal/git/merge`)
   - [ ] Find merge base (common ancestor)
   - [ ] Three-way merge algorithm (file-level)
   - [ ] Detect conflicts
   - [ ] Create merge commit on success

2. **Conflict Detection** (`internal/git/conflicts`)
   - [ ] Identify files modified in both branches
   - [ ] Extract conflict context (base, source, target)
   - [ ] Store conflicts in PR

3. **Manual Conflict Resolution UI**
   - [ ] Display conflict markers in UI
   - [ ] Allow manual editing of conflicted files
   - [ ] "Resolve Conflict" action

#### Acceptance Criteria
- Can merge PR when no conflicts
- Can detect merge conflicts
- Can manually resolve conflicts in UI
- Creates merge commit with two parents

---

### Phase 5: Sectioned Documents (Advanced Merging)
**Duration**: 2-3 weeks  
**Dependencies**: Phase 4

#### Tasks
1. **Document Parser** (`internal/git/documents`)
   - [ ] Parse YAML frontmatter
   - [ ] Extract sections by ID markers
   - [ ] Reconstruct document from sections

2. **Section-Level Merge** (`internal/git/merge/sectioned`)
   - [ ] Three-way merge at section granularity
   - [ ] Auto-merge different sections
   - [ ] Conflict only when same section modified
   - [ ] Preserve section order

3. **Document Templates**
   - [ ] Requirements document template
   - [ ] Design document template
   - [ ] Work item template

#### Acceptance Criteria
- Documents structured with section markers
- Can merge documents with different section edits
- Conflicts only occur on same-section modifications
- Section order preserved

---

### Phase 6: AI-Assisted Conflict Resolution (Intelligence Layer)
**Duration**: 2-3 weeks  
**Dependencies**: Phase 5

#### Tasks
1. **AI Conflict Resolver** (`internal/git/ai`)
   - [ ] Extract conflict context (base, source, target)
   - [ ] Generate AI prompt for conflict resolution
   - [ ] Parse AI response
   - [ ] Store AI resolution in PR

2. **AI Resolution UI**
   - [ ] Display AI-proposed merge in PR
   - [ ] Show AI reasoning/confidence
   - [ ] "Accept AI Resolution" button
   - [ ] "Edit AI Resolution" option

3. **Prompt Engineering**
   - [ ] Context-aware prompts (code vs docs)
   - [ ] Section-level conflict prompts
   - [ ] Handle multi-file conflicts

#### Acceptance Criteria
- AI generates merge proposal for conflicts
- Human can review and approve AI merge
- AI reasoning visible in UI
- Can override AI decision

---

### Phase 7: Kanban Integration (Workflows)
**Duration**: 2-3 weeks  
**Dependencies**: Phase 3

#### Tasks
1. **Issue → Branch Linkage** (`internal/git/workitems`)
   - [ ] `WorkIssue.Files` - Track affected files
   - [ ] `WorkIssue.Branch` - Associated branch
   - [ ] Auto-create branch when issue assigned to agent

2. **Workflow Automation**
   - [ ] Issue milestone "In Progress" → Create branch
   - [ ] Branch commit → Update issue activity
   - [ ] PR created → Issue milestone "Review"
   - [ ] PR merged → Issue milestone "Done"

3. **Agent Integration**
   - [ ] Agent can read/write files via API
   - [ ] Agent commits attributed correctly
   - [ ] Agent can create PRs

#### Acceptance Criteria
- Issues linked to file changes
- Branch auto-created for agent work
- PR auto-created when agent completes
- Issue status synced with PR state

---

## Technical Specifications

### Database Collections

```go
// Add to internal/agency/database_initializer.go
func initializeGitCollections(db *arangodb.Database) error {
    collections := []string{
        "git_objects",      // Content-addressable Git objects
        "git_refs",         // Branches, tags, HEAD
        "repositories",     // Repository metadata
        "pull_requests",    // PR tracking
        "file_index",       // Fast file path lookups
    }
    
    for _, coll := range collections {
        if err := createCollection(db, coll); err != nil {
            return err
        }
    }
    
    return nil
}
```

### API Endpoints

```
# File Operations
GET    /api/v1/instances/{id}/files?path={path}        # List directory
GET    /api/v1/instances/{id}/files/{path}             # Read file
PUT    /api/v1/instances/{id}/files/{path}             # Update file
DELETE /api/v1/instances/{id}/files/{path}             # Delete file
POST   /api/v1/instances/{id}/files/{path}             # Create file

# Repository Management
GET    /api/v1/instances/{id}/repository                # Get repo info
GET    /api/v1/instances/{id}/repository/branches       # List branches
POST   /api/v1/instances/{id}/repository/branches       # Create branch
DELETE /api/v1/instances/{id}/repository/branches/{name} # Delete branch

# Pull Requests
GET    /api/v1/instances/{id}/pull-requests             # List PRs
POST   /api/v1/instances/{id}/pull-requests             # Create PR
GET    /api/v1/instances/{id}/pull-requests/{prID}      # Get PR details
POST   /api/v1/instances/{id}/pull-requests/{prID}/approve # Approve PR
POST   /api/v1/instances/{id}/pull-requests/{prID}/merge  # Merge PR

# History
GET    /api/v1/instances/{id}/files/{path}/history      # File commit history
GET    /api/v1/instances/{id}/commits/{sha}             # Get commit details
```

### Package Structure

```
internal/
├── git/
│   ├── models/               # Git data models
│   │   ├── object.go         # GitObject, GitBlob, GitTree, GitCommit
│   │   ├── ref.go            # GitRef
│   │   ├── repository.go     # Repository
│   │   └── pullrequest.go    # PullRequest
│   ├── storage/              # Low-level storage
│   │   ├── objects.go        # Object storage operations
│   │   └── refs.go           # Ref storage operations
│   ├── ops/                  # Git operations
│   │   ├── commit.go         # WriteBlob, WriteTree, Commit
│   │   ├── read.go           # GetCommit, GetTree, GetBlob
│   │   └── branch.go         # Branch operations
│   ├── merge/                # Merge logic
│   │   ├── threeway.go       # Three-way merge
│   │   ├── sectioned.go      # Section-level merge
│   │   └── conflicts.go      # Conflict detection
│   ├── documents/            # Document parsing
│   │   ├── parser.go         # Parse sectioned documents
│   │   └── templates/        # Document templates
│   ├── fileindex/            # File index
│   │   └── index.go          # Path → blob mapping
│   ├── ai/                   # AI conflict resolution
│   │   └── resolver.go       # AI merge proposals
│   └── workitems/            # Kanban integration
│       └── integration.go    # Issue → branch → PR
└── web/
    ├── handlers/
    │   ├── files/            # File API handlers
    │   ├── pull_requests/    # PR API handlers
    │   └── repository/       # Repository API handlers
    └── pages/
        ├── files/            # File explorer UI
        └── pull_requests/    # PR UI
```

---

## Testing Strategy

### Unit Tests
- Git object serialization/deserialization
- SHA-1 hash calculation
- Three-way merge algorithm
- Section parsing
- Conflict detection

### Integration Tests
- End-to-end file operations
- Branch creation and switching
- PR creation and merge
- AI conflict resolution

### UI Tests
- File explorer navigation
- File editing and saving
- PR creation workflow
- Conflict resolution UI

---

## Performance Considerations

1. **Content-Addressable Deduplication**
   - Same content = same SHA = stored once
   - Reduces storage for unchanged files

2. **File Index Caching**
   - Cache file index in memory
   - Rebuild only on commit

3. **Lazy Tree Loading**
   - Load tree nodes only when accessed
   - Don't load entire repository tree

4. **Diff Optimization**
   - Calculate diffs on demand
   - Cache diff results in PR

---

## Migration Path

### Existing Work Items
```go
// Migrate existing work_issues to use Git
func migrateWorkIssues() {
    issues := getAllIssues()
    for _, issue := range issues {
        // Create repository if not exists
        repo := getOrCreateRepository(issue.InstanceID)
        
        // Create initial commit if empty
        if repo.DefaultRef == "" {
            initializeRepository(repo)
        }
        
        // Create branch for in-progress issues
        if issue.Status == "in_progress" {
            createBranchForIssue(issue)
        }
    }
}
```

---

## Success Metrics

- ✅ **Core Git operations working** (commit, branch, merge)
- ✅ **File explorer UI functional** (browse, edit, save)
- ✅ **Pull requests created and merged**
- ✅ **Conflicts detected and resolvable**
- ✅ **Section-based merging reduces conflicts by 70%+**
- ✅ **AI resolves 60%+ of conflicts automatically**
- ✅ **Issues linked to file changes in Kanban**

---

## References

- **Architecture**: [git-based-document-system.md](./git-based-document-system.md)
- **Decisions**: [architecture/vcs-integration-decisions.md](./architecture/vcs-integration-decisions.md)
- **Git Internals**: https://git-scm.com/book/en/v2/Git-Internals-Git-Objects
- **Three-Way Merge**: https://en.wikipedia.org/wiki/Merge_(version_control)#Three-way_merge
