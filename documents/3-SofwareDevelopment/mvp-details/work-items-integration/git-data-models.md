# Git Data Models

**Related Tasks**: MVP-WI-005 (Git Core Layer), MVP-WI-007 (Pull Requests)  
**Status**: MVP-WI-005 ✅ Complete | MVP-WI-007 📋 Planned

## Overview

This document defines the data models for the Git implementation in ArangoDB. All Git objects are stored in content-addressable collections with SHA-1 hashing for deduplication and integrity.

**For Git operations (read/write)**, see [Git Core Operations](git-core-operations.md).  
**For merge algorithms**, see [Git Merge Strategies](git-merge-strategies.md).

---

## ArangoDB Collections

```go
// Agency database collections (Git subsystem)
collections := []string{
    // Content-addressable object store
    "git_objects",        // Blobs, trees, commits (immutable)
    "git_refs",           // Branches, tags, HEAD pointers
    
    // File system abstraction
    "git_artifacts",      // File/folder metadata (path → blob SHA mapping)
    
    // Repository management
    "repositories",       // Repository metadata (one per instance)
    "pull_requests",      // Pull request tracking and review
    
    // Work tracking integration
    "work_issues",        // Issues linked to file changes (Kanban)
    "workflows",          // Kanban workflow definitions
    "agents",             // Agent instances for work execution
}
```

**Key Principles**:
- **Content-addressable**: SHA-1 hash = document `_key`
- **Immutable objects**: Blobs, trees, commits never modified after creation
- **Deduplication**: Same content = same SHA = stored once
- **Repository-scoped**: All objects tagged with `repo_id` for isolation

---

## Core Git Objects

### GitObject (Base Storage)

```go
type GitObject struct {
    // ArangoDB document fields
    Key string `json:"_key"`      // SHA-1 hash of content (hex-encoded)
    ID  string `json:"_id"`       // Collection/key (git_objects/abc123...)
    Rev string `json:"_rev"`      // ArangoDB revision
    
    // Git object data
    Type    string `json:"type"`      // "commit", "tree", "blob"
    Size    int    `json:"size"`      // Content size in bytes
    Content string `json:"content"`   // Serialized object (plain text)
    
    // Metadata for queries and filtering
    RepoID  string    `json:"repo_id"`    // Repository identifier
    Created time.Time `json:"created"`    // Object creation timestamp
}
```

**Properties**:
- **SHA-1 as Key**: Document key is the content hash (ensures uniqueness)
- **Immutability**: Once created, never updated (append-only)
- **Plain Text Content**: No base64 encoding (readable in database, easier debugging)
- **Repository Isolation**: `repo_id` field for multi-instance support

**ArangoDB Indexes**:
```aql
CREATE INDEX idx_repo_type ON git_objects (repo_id, type)
```

---

### GitBlob (File Content)

```go
type GitBlob struct {
    SHA     string `json:"sha"`       // SHA-1 hash (hex-encoded)
    Content string `json:"content"`   // Plain text file content
    Size    int    `json:"size"`      // Content length in bytes
}
```

**Stored as GitObject**:
```json
{
  "_key": "abc123def456...",
  "type": "blob",
  "size": 34,
  "content": "# README\n\nThis is a test file.",
  "repo_id": "instance-001",
  "created": "2025-11-27T10:00:00Z"
}
```

**SHA-1 Calculation**:
```go
// Git blob format: "blob {size}\0{content}"
hash := sha1.Sum([]byte(fmt.Sprintf("blob %d\x00%s", len(content), content)))
sha := hex.EncodeToString(hash[:])
```

**Why Plain Text?**
- ✅ No base64 encoding overhead
- ✅ AQL queries can search content directly
- ✅ Easier debugging and inspection
- ✅ Smaller storage footprint
- ✅ Compatible with existing Git SHA-1 algorithm

---

### GitTree (Directory Structure)

```go
type GitTree struct {
    SHA     string      `json:"sha"`       // SHA-1 hash of tree
    Entries []TreeEntry `json:"entries"`   // Directory contents
}

type TreeEntry struct {
    Mode string `json:"mode"`   // File permissions (100644, 100755, 040000)
    Type string `json:"type"`   // "blob" (file) or "tree" (directory)
    SHA  string `json:"sha"`    // Content hash
    Path string `json:"path"`   // File/directory name (not full path)
}
```

**Mode Values**:
- `100644` - Regular file
- `100755` - Executable file
- `040000` - Directory (subdirectory)
- `120000` - Symbolic link

**Stored as GitObject**:
```json
{
  "_key": "def456ghi789...",
  "type": "tree",
  "content": "100644 blob abc123... README.md\n040000 tree xyz789... src\n",
  "repo_id": "instance-001",
  "created": "2025-11-27T10:05:00Z"
}
```

**Tree Content Format**:
```
{mode} {type} {sha}\t{path}\n
100644 blob abc123def456... README.md
040000 tree 789xyz123abc... documents
100644 blob def789ghi012... config.yaml
```

**SHA-1 Calculation**:
```go
// Serialize tree entries
var buf bytes.Buffer
for _, e := range entries {
    fmt.Fprintf(&buf, "%s %s %s\t%s\n", e.Mode, e.Type, e.SHA, e.Path)
}
content := buf.String()

// Git tree format: "tree {size}\0{content}"
hash := sha1.Sum([]byte(fmt.Sprintf("tree %d\x00%s", len(content), content)))
sha := hex.EncodeToString(hash[:])
```

---

### GitCommit (Version Snapshot)

```go
type GitCommit struct {
    SHA       string    `json:"sha"`         // SHA-1 hash of commit
    Tree      string    `json:"tree"`        // Root tree SHA
    Parents   []string  `json:"parents"`     // Parent commit SHAs (0+ commits)
    Author    Author    `json:"author"`      // Author details
    Committer Author    `json:"committer"`   // Committer details
    Message   string    `json:"message"`     // Commit message
    Timestamp time.Time `json:"timestamp"`   // Commit creation time
}

type Author struct {
    Name  string    `json:"name"`    // "John Doe" or "agent-impl-001"
    Email string    `json:"email"`   // "john@example.com" or "agent@codevaldcortex"
    When  time.Time `json:"when"`    // Timestamp
}
```

**Parent Commits**:
- **0 parents**: Initial commit (root)
- **1 parent**: Normal commit
- **2+ parents**: Merge commit

**Stored as GitObject**:
```json
{
  "_key": "789abc012def...",
  "type": "commit",
  "content": "tree def456ghi789...\nparent abc123...\nauthor John Doe <john@example.com> 1732700000\ncommitter codevaldcortex <system@codevaldcortex> 1732700000\n\nAdd JWT authentication implementation",
  "repo_id": "instance-001",
  "created": "2025-11-27T10:10:00Z"
}
```

**Commit Content Format**:
```
tree {tree_sha}
parent {parent_sha}
parent {parent_sha}  (if merge commit)
author {name} <{email}> {timestamp}
committer {name} <{email}> {timestamp}

{commit message}
```

**SHA-1 Calculation**:
```go
content := serializeCommit(commit)  // Format above
hash := sha1.Sum([]byte(fmt.Sprintf("commit %d\x00%s", len(content), content)))
sha := hex.EncodeToString(hash[:])
```

---

## References & Metadata

### GitRef (Branches & Tags)

```go
type GitRef struct {
    // ArangoDB fields
    Key string `json:"_key"`      // Reference name (refs/heads/main, refs/tags/v1.0)
    ID  string `json:"_id"`
    Rev string `json:"_rev"`
    
    // Reference data
    RepoID string `json:"repo_id"`    // Repository identifier
    Type   string `json:"type"`       // "branch" or "tag"
    Target string `json:"target"`     // Commit SHA this ref points to
    
    // Metadata
    Created time.Time `json:"created"`
    Updated time.Time `json:"updated"`
}
```

**Reference Naming**:
- **Branches**: `refs/heads/{branch_name}` (e.g., `refs/heads/main`, `refs/heads/feature-jwt`)
- **Tags**: `refs/tags/{tag_name}` (e.g., `refs/tags/v1.0.0`)
- **HEAD**: `refs/HEAD` (symbolic ref to current branch)

**Example Refs**:
```json
[
  {
    "_key": "refs/heads/main",
    "repo_id": "instance-001",
    "type": "branch",
    "target": "789abc012def...",
    "updated": "2025-11-27T10:10:00Z"
  },
  {
    "_key": "refs/heads/issue-123-jwt-auth",
    "repo_id": "instance-001",
    "type": "branch",
    "target": "456def789ghi...",
    "updated": "2025-11-27T10:15:00Z"
  }
]
```

**ArangoDB Indexes**:
```aql
CREATE INDEX idx_repo_ref ON git_refs (repo_id, type)
```

---

### Repository (Project Container)

```go
type Repository struct {
    // ArangoDB fields
    Key string `json:"_key"`       // Repository identifier (instance ID)
    ID  string `json:"_id"`
    Rev string `json:"_rev"`
    
    // Repository metadata
    InstanceID  string `json:"instance_id"`  // Agency instance ID (1:1 mapping)
    Name        string `json:"name"`         // Repository display name
    Description string `json:"description"`  // Repository description
    DefaultRef  string `json:"default_ref"`  // Default branch (refs/heads/main)
    
    // File structure
    RootPath    string `json:"root_path"`    // Root directory path (usually "/")
    
    // Timestamps
    Created time.Time `json:"created"`
    Updated time.Time `json:"updated"`
}
```

**One Repository Per Instance**:
- Each agency instance has **exactly one repository**
- Repository ID = Instance ID
- Default branch: `main`
- All work items create branches in this repository

**Example Repository**:
```json
{
  "_key": "instance-001",
  "instance_id": "instance-001",
  "name": "Water Distribution Network",
  "description": "UC-INFRA-001 - Smart water distribution network management",
  "default_ref": "refs/heads/main",
  "root_path": "/",
  "created": "2025-11-20T09:00:00Z",
  "updated": "2025-11-27T10:10:00Z"
}
```

---

### PullRequest (Review Workflow)

```go
type PullRequest struct {
    // ArangoDB fields
    Key string `json:"_key"`       // PR identifier (pr-123)
    ID  string `json:"_id"`
    Rev string `json:"_rev"`
    
    // Context
    RepoID      string `json:"repo_id"`      // Repository ID
    InstanceID  string `json:"instance_id"`  // Agency instance ID
    
    // Metadata
    Title       string `json:"title"`        // PR title
    Description string `json:"description"`  // PR description (Markdown)
    
    // Git references
    SourceBranch string `json:"source_branch"`  // refs/heads/issue-123-jwt-auth
    TargetBranch string `json:"target_branch"`  // refs/heads/main
    BaseCommit   string `json:"base_commit"`    // Merge base (common ancestor)
    HeadCommit   string `json:"head_commit"`    // Latest commit on source branch
    MergeCommit  string `json:"merge_commit,omitempty"` // Created on successful merge
    
    // Status
    Status      string   `json:"status"`       // "open", "merged", "closed", "conflict"
    Conflicts   []string `json:"conflicts"`    // File paths with conflicts
    
    // Review
    CreatedBy   string    `json:"created_by"`   // User ID or agent ID
    Reviewers   []string  `json:"reviewers"`    // Assigned reviewers
    ApprovedBy  []string  `json:"approved_by"`  // Users who approved
    
    // AI assistance (optional)
    AIResolution *AIConflictResolution `json:"ai_resolution,omitempty"`
    
    // Timestamps
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type AIConflictResolution struct {
    ConflictFiles []string  `json:"conflict_files"`  // Files with conflicts
    ProposedMerge string    `json:"proposed_merge"`  // AI-generated merge result
    Confidence    float64   `json:"confidence"`      // Confidence score (0.0-1.0)
    Reasoning     string    `json:"reasoning"`       // AI explanation
    AppliedAt     time.Time `json:"applied_at,omitempty"` // When resolution applied
}
```

**PR Status Values**:
- `open` - Created, awaiting review
- `merged` - Successfully merged to target
- `closed` - Closed without merging
- `conflict` - Merge conflicts detected

**Example Pull Request**:
```json
{
  "_key": "pr-456",
  "repo_id": "instance-001",
  "instance_id": "instance-001",
  "title": "Implement JWT authentication requirements",
  "description": "This PR adds the requirements document for JWT authentication...",
  "source_branch": "refs/heads/issue-123-jwt-auth",
  "target_branch": "refs/heads/main",
  "base_commit": "abc123...",
  "head_commit": "def456...",
  "status": "open",
  "conflicts": [],
  "created_by": "agent-req-001",
  "reviewers": ["user-alice"],
  "approved_by": [],
  "created_at": "2025-11-27T10:20:00Z",
  "updated_at": "2025-11-27T10:20:00Z"
}
```

---

## Performance Considerations

### Content-Addressable Benefits

- **Deduplication**: Same content stored once (same SHA = same document)
- **Fast Lookups**: SHA-1 is document key (O(1) retrieval by hash)
- **Immutability**: No updates needed, only inserts (simpler concurrency)
- **Cache-Friendly**: Objects never change (safe to cache indefinitely)
- **Integrity**: SHA-1 hash verifies content hasn't been corrupted

### Query Optimization

**Recommended Indexes**:
```aql
// git_objects collection
CREATE INDEX idx_repo_type ON git_objects (repo_id, type)
CREATE INDEX idx_created ON git_objects (created)

// git_refs collection
CREATE INDEX idx_repo_ref ON git_refs (repo_id, type)

// pull_requests collection
CREATE INDEX idx_repo_status ON pull_requests (repo_id, status)
CREATE INDEX idx_instance ON pull_requests (instance_id)
```

**Fast Queries**:
```aql
// Get all commits for repository (chronological)
FOR obj IN git_objects
  FILTER obj.repo_id == @repoID AND obj.type == "commit"
  SORT obj.created DESC
  LIMIT 100
  RETURN obj

// Get current branch head
FOR ref IN git_refs
  FILTER ref.repo_id == @repoID AND ref._key == "refs/heads/main"
  RETURN ref.target

// Get open PRs for instance
FOR pr IN pull_requests
  FILTER pr.instance_id == @instanceID AND pr.status == "open"
  SORT pr.created_at DESC
  RETURN pr
```

---

## Related Documentation

- **Git Core Operations**: [git-core-operations.md](git-core-operations.md) - Read/write operations
- **Git Merge Strategies**: [git-merge-strategies.md](git-merge-strategies.md) - Three-way merge, conflict handling
- **File Explorer**: [file-explorer.md](file-explorer.md) - High-level file operations
- **Pull Requests**: [pull-requests.md](pull-requests.md) - Code review workflow
- **AI Conflict Resolution**: [ai-conflict-resolution.md](ai-conflict-resolution.md) - AI-assisted merging

---

**Last Updated**: 2025-11-27  
**Status**: Architecture Defined - MVP-WI-005 ✅ Complete
