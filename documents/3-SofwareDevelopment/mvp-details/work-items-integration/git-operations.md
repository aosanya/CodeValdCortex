# Git Operations Implementation

**Related Tasks**: MVP-WI-005 (Git Core Layer), MVP-WI-007 (Pull Requests)  
**Status**: MVP-WI-005 ✅ Complete | MVP-WI-007 📋 Planned

## Overview

This document describes the low-level Git implementation in ArangoDB, including object storage, operations (commit, branch, merge), and data models.

---

## Git Object Model in ArangoDB

### Collections

```go
// Agency database collections
collections := []string{
    // Git object store (content-addressable)
    "git_objects",        // All Git objects (commits, trees, blobs)
    "git_refs",           // References (branches, tags, HEAD)
    
    // File system abstraction
    "git_artifacts",      // File/folder metadata (path → blob SHA mapping)
    
    // Higher-level abstractions
    "repositories",       // Repository metadata (one per instance/project)
    "pull_requests",      // Pull request tracking
    
    // Work tracking (Kanban integration)
    "work_issues",        // Issues linked to file changes
    "workflows",          // Kanban workflows
    "agents",             // Agent instances
}
```

---

## Data Models

### GitObject (Core Storage)

```go
type GitObject struct {
    // ArangoDB fields
    Key string `json:"_key"`      // SHA-1 hash of content
    ID  string `json:"_id"`
    Rev string `json:"_rev"`
    
    // Git object data
    Type    string `json:"type"`      // "commit", "tree", "blob"
    Size    int    `json:"size"`
    Content string `json:"content"`   // Serialized object
    
    // Metadata for queries
    RepoID  string    `json:"repo_id"`
    Created time.Time `json:"created"`
}
```

**Key Properties**:
- Content-addressable: SHA-1 hash = document key
- Immutable: Once created, never modified
- Deduplication: Same content = same SHA = stored once
- Repository-scoped: `repo_id` for multi-instance isolation

---

### GitBlob (File Content)

```go
// Blob = file content (plain text)
type GitBlob struct {
    SHA     string `json:"sha"`
    Content string `json:"content"`  // Plain text (code or Markdown)
    Size    int    `json:"size"`
}

// Stored as GitObject with type="blob"
```

**Storage Format**:
```
Content: "# README\n\nThis is a test file."
SHA-1: blob 34\0# README\n\nThis is a test file.
Key: "abc123..." (hex-encoded SHA-1)
```

**Why Plain Text?**
- ✅ No base64 encoding (readable in database)
- ✅ AQL queries can search content
- ✅ Easier debugging
- ✅ Smaller storage size

---

### GitTree (Directory Structure)

```go
type GitTree struct {
    SHA     string      `json:"sha"`
    Entries []TreeEntry `json:"entries"`
}

type TreeEntry struct {
    Mode string `json:"mode"`    // "100644" (file), "040000" (directory)
    Type string `json:"type"`    // "blob", "tree"
    SHA  string `json:"sha"`     // Object hash
    Path string `json:"path"`    // File/directory name
}
```

**Example Tree**:
```json
{
  "sha": "abc123",
  "entries": [
    {"mode": "040000", "type": "tree", "sha": "def456", "path": "documents"},
    {"mode": "100644", "type": "blob", "sha": "789abc", "path": "README.md"}
  ]
}
```

**Serialized Format** (Git-compatible):
```
040000 tree def456\tdocuments
100644 blob 789abc\tREADME.md
```

---

### GitCommit (Version Snapshot)

```go
type GitCommit struct {
    SHA       string    `json:"sha"`
    Tree      string    `json:"tree"`        // Root tree SHA
    Parents   []string  `json:"parents"`     // Parent commit SHAs (0-2 for merge)
    Author    Author    `json:"author"`
    Committer Author    `json:"committer"`
    Message   string    `json:"message"`
    Timestamp time.Time `json:"timestamp"`
}

type Author struct {
    Name  string    `json:"name"`   // "user-alice" or "agent-xyz"
    Email string    `json:"email"`
    When  time.Time `json:"when"`
}
```

**Example Commit**:
```json
{
  "sha": "commit-123",
  "tree": "tree-abc",
  "parents": ["commit-000"],
  "author": {"name": "user-alice", "when": "2025-11-26T10:00:00Z"},
  "committer": {"name": "codevaldcortex", "when": "2025-11-26T10:00:01Z"},
  "message": "Update requirements document",
  "timestamp": "2025-11-26T10:00:00Z"
}
```

**Merge Commits**:
- Regular commit: 1 parent
- Merge commit: 2 parents (source + target branches)

---

### GitRef (Branches & Tags)

```go
type GitRef struct {
    Key    string `json:"_key"`      // "refs/heads/main", "refs/heads/feature-xyz"
    RepoID string `json:"repo_id"`
    Type   string `json:"type"`      // "branch", "tag", "HEAD"
    Target string `json:"target"`    // Commit SHA
}
```

**Examples**:
```
refs/heads/main              → "commit-sha-abc"
refs/heads/draft-alice       → "commit-sha-def"
refs/heads/agent-work-123    → "commit-sha-789"
HEAD                         → "refs/heads/main"
```

**Reference Naming**:
- Branches: `refs/heads/{branch-name}`
- Tags: `refs/tags/{tag-name}`
- HEAD: `HEAD` (symbolic ref)

---

### Repository (Project Container)

```go
type Repository struct {
    Key         string `json:"_key"`
    InstanceID  string `json:"instance_id"`  // One repo per instance
    Name        string `json:"name"`
    Description string `json:"description"`
    DefaultRef  string `json:"default_ref"`  // "refs/heads/main"
    
    // File structure
    RootPath    string `json:"root_path"`    // "/" for root
    
    Created time.Time `json:"created"`
    Updated time.Time `json:"updated"`
}
```

**One Repository Per Instance**:
- Each agency instance has its own repository
- Repository ID = Instance ID
- Default branch: `main`

---

### PullRequest (Review Workflow)

```go
type PullRequest struct {
    Key string `json:"_key"`
    
    // Context
    RepoID      string `json:"repo_id"`
    InstanceID  string `json:"instance_id"`
    
    // Metadata
    Title       string `json:"title"`
    Description string `json:"description"`
    
    // Git references
    SourceBranch string `json:"source_branch"`  // "refs/heads/draft-alice"
    TargetBranch string `json:"target_branch"`  // "refs/heads/main"
    BaseCommit   string `json:"base_commit"`    // Merge base (common ancestor)
    HeadCommit   string `json:"head_commit"`    // Latest commit on source
    MergeCommit  string `json:"merge_commit,omitempty"` // Created on merge
    
    // Status
    Status      string   `json:"status"`       // "open", "merged", "closed", "conflict"
    Conflicts   []string `json:"conflicts"`    // File paths with conflicts
    
    // Review
    CreatedBy   string    `json:"created_by"`   // User or agent
    Reviewers   []string  `json:"reviewers"`
    ApprovedBy  []string  `json:"approved_by"`
    
    // AI assistance
    AIResolution *AIConflictResolution `json:"ai_resolution,omitempty"`
    
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type AIConflictResolution struct {
    ConflictFiles []string  `json:"conflict_files"`
    ProposedMerge string    `json:"proposed_merge"`  // AI-generated merge
    Confidence    float64   `json:"confidence"`      // 0.0-1.0
    Reasoning     string    `json:"reasoning"`       // AI explanation
    AppliedAt     time.Time `json:"applied_at,omitempty"`
}
```

---

## Write Operations

### WriteBlob (Store File Content)

```go
package gitops

type GitOps struct {
    db *arangodb.Database
}

// WriteBlob stores file content
func (g *GitOps) WriteBlob(repoID, content string) (string, error) {
    // Calculate SHA-1
    hash := sha1.Sum([]byte(fmt.Sprintf("blob %d\x00%s", len(content), content)))
    sha := hex.EncodeToString(hash[:])
    
    obj := &GitObject{
        Key:     sha,
        Type:    "blob",
        Size:    len(content),
        Content: content,
        RepoID:  repoID,
        Created: time.Now(),
    }
    
    // Insert (idempotent - same content = same SHA)
    g.db.Collection("git_objects").CreateDocument(ctx, obj)
    return sha, nil
}
```

**Idempotency**:
- Same content → same SHA → already exists
- CreateDocument fails with conflict error (ignore)
- No duplicate storage

---

### WriteTree (Store Directory)

```go
// WriteTree stores directory structure
func (g *GitOps) WriteTree(repoID string, entries []TreeEntry) (string, error) {
    // Serialize tree
    var buf bytes.Buffer
    for _, e := range entries {
        fmt.Fprintf(&buf, "%s %s %s\t%s\n", e.Mode, e.Type, e.SHA, e.Path)
    }
    content := buf.String()
    
    hash := sha1.Sum([]byte(fmt.Sprintf("tree %d\x00%s", len(content), content)))
    sha := hex.EncodeToString(hash[:])
    
    obj := &GitObject{
        Key:     sha,
        Type:    "tree",
        Content: content,
        RepoID:  repoID,
        Created: time.Now(),
    }
    
    g.db.Collection("git_objects").CreateDocument(ctx, obj)
    return sha, nil
}
```

---

### Commit (Create Version Snapshot)

```go
// Commit creates a version snapshot
func (g *GitOps) Commit(repoID, treeSHA string, parents []string, author, message string) (string, error) {
    commit := &GitCommit{
        Tree:      treeSHA,
        Parents:   parents,
        Author:    Author{Name: author, When: time.Now()},
        Committer: Author{Name: "codevaldcortex", When: time.Now()},
        Message:   message,
        Timestamp: time.Now(),
    }
    
    // Serialize
    content := serializeCommit(commit)
    hash := sha1.Sum([]byte(fmt.Sprintf("commit %d\x00%s", len(content), content)))
    sha := hex.EncodeToString(hash[:])
    
    obj := &GitObject{
        Key:     sha,
        Type:    "commit",
        Content: content,
        RepoID:  repoID,
        Created: time.Now(),
    }
    
    g.db.Collection("git_objects").CreateDocument(ctx, obj)
    return sha, nil
}
```

---

### UpdateRef (Move Branch Pointer)

```go
// UpdateRef moves branch pointer
func (g *GitOps) UpdateRef(repoID, refName, commitSHA string) error {
    ref := &GitRef{
        Key:    refName,
        RepoID: repoID,
        Type:   "branch",
        Target: commitSHA,
    }
    
    g.db.Collection("git_refs").UpdateDocument(ctx, refName, ref)
    return nil
}
```

**Atomic Branch Updates**:
- ArangoDB document updates are atomic
- No race conditions on branch moves
- Failed update rolls back automatically

---

## Read Operations

### GetCommit

```go
// GetCommit retrieves commit object
func (g *GitOps) GetCommit(sha string) (*GitCommit, error) {
    var obj GitObject
    g.db.Collection("git_objects").ReadDocument(ctx, sha, &obj)
    
    return deserializeCommit(obj.Content), nil
}
```

### GetTree

```go
// GetTree retrieves directory structure
func (g *GitOps) GetTree(sha string) (*GitTree, error) {
    var obj GitObject
    g.db.Collection("git_objects").ReadDocument(ctx, sha, &obj)
    
    return deserializeTree(obj.Content), nil
}
```

### GetBlob

```go
// GetBlob retrieves file content
func (g *GitOps) GetBlob(sha string) (string, error) {
    var obj GitObject
    g.db.Collection("git_objects").ReadDocument(ctx, sha, &obj)
    
    return obj.Content, nil
}
```

### ReadFile (Path-Based Lookup)

```go
// ReadFile reads file at specific path and commit
func (g *GitOps) ReadFile(repoID, commitSHA, filePath string) (string, error) {
    // 1. Get commit
    commit, _ := g.GetCommit(commitSHA)
    
    // 2. Get root tree
    tree, _ := g.GetTree(commit.Tree)
    
    // 3. Navigate path (e.g., "documents/requirements.md")
    parts := strings.Split(filePath, "/")
    currentTree := tree
    
    for i, part := range parts {
        if i == len(parts)-1 {
            // Last part = file
            entry := findEntry(currentTree, part)
            return g.GetBlob(entry.SHA)
        }
        
        // Directory - navigate deeper
        entry := findEntry(currentTree, part)
        currentTree, _ = g.GetTree(entry.SHA)
    }
    
    return "", errors.New("file not found")
}
```

**Path Navigation**:
- Start at root tree
- Walk path segments
- Lookup each directory as tree
- Final segment is file blob

---

## Merge Operations

### Three-Way Merge

```go
// Merge performs three-way merge
func (g *GitOps) Merge(repoID, sourceBranch, targetBranch string) (*MergeResult, error) {
    // 1. Get branch heads
    sourceRef := g.getRef(repoID, sourceBranch)
    targetRef := g.getRef(repoID, targetBranch)
    
    sourceCommit, _ := g.GetCommit(sourceRef.Target)
    targetCommit, _ := g.GetCommit(targetRef.Target)
    
    // 2. Find merge base (common ancestor)
    baseCommit := g.findMergeBase(sourceCommit, targetCommit)
    
    // 3. Get trees
    baseTree, _ := g.GetTree(baseCommit.Tree)
    sourceTree, _ := g.GetTree(sourceCommit.Tree)
    targetTree, _ := g.GetTree(targetCommit.Tree)
    
    // 4. Three-way merge
    mergedTree, conflicts := g.mergeTreesThreeWay(baseTree, sourceTree, targetTree)
    
    if len(conflicts) > 0 {
        return &MergeResult{
            Status:    "conflict",
            Conflicts: conflicts,
        }, nil
    }
    
    // 5. Create merge commit
    treeSHA, _ := g.WriteTree(repoID, mergedTree.Entries)
    commitSHA, _ := g.Commit(
        repoID,
        treeSHA,
        []string{sourceCommit.SHA, targetCommit.SHA},
        "system",
        fmt.Sprintf("Merge %s into %s", sourceBranch, targetBranch),
    )
    
    // 6. Update target branch
    g.UpdateRef(repoID, targetBranch, commitSHA)
    
    return &MergeResult{
        Status:      "merged",
        CommitSHA:   commitSHA,
        MergeCommit: commitSHA,
    }, nil
}
```

### Merge Algorithm (File-Level)

```go
// mergeTreesThreeWay performs file-level three-way merge
func (g *GitOps) mergeTreesThreeWay(base, source, target *GitTree) (*GitTree, []Conflict) {
    merged := &GitTree{Entries: []TreeEntry{}}
    conflicts := []Conflict{}
    
    // Build maps for comparison
    baseMap := toEntryMap(base)
    sourceMap := toEntryMap(source)
    targetMap := toEntryMap(target)
    
    // Get all file paths
    allPaths := getAllPaths(base, source, target)
    
    for _, path := range allPaths {
        baseEntry := baseMap[path]
        sourceEntry := sourceMap[path]
        targetEntry := targetMap[path]
        
        // Case 1: File unchanged in both
        if baseEntry != nil && sourceEntry != nil && targetEntry != nil {
            if sourceEntry.SHA == targetEntry.SHA {
                merged.Entries = append(merged.Entries, sourceEntry)
                continue
            }
        }
        
        // Case 2: File modified in source only
        if baseEntry != nil && targetEntry != nil && targetEntry.SHA == baseEntry.SHA && sourceEntry != nil {
            merged.Entries = append(merged.Entries, sourceEntry)
            continue
        }
        
        // Case 3: File modified in target only
        if baseEntry != nil && sourceEntry != nil && sourceEntry.SHA == baseEntry.SHA && targetEntry != nil {
            merged.Entries = append(merged.Entries, targetEntry)
            continue
        }
        
        // Case 4: File added in source only
        if baseEntry == nil && targetEntry == nil && sourceEntry != nil {
            merged.Entries = append(merged.Entries, sourceEntry)
            continue
        }
        
        // Case 5: File added in target only
        if baseEntry == nil && sourceEntry == nil && targetEntry != nil {
            merged.Entries = append(merged.Entries, targetEntry)
            continue
        }
        
        // Case 6: CONFLICT - File modified in both
        if sourceEntry != nil && targetEntry != nil && sourceEntry.SHA != targetEntry.SHA {
            conflicts = append(conflicts, Conflict{
                Path:       path,
                BaseSHA:    baseEntry.SHA,
                SourceSHA:  sourceEntry.SHA,
                TargetSHA:  targetEntry.SHA,
            })
            continue
        }
    }
    
    return merged, conflicts
}

type Conflict struct {
    Path      string `json:"path"`
    BaseSHA   string `json:"base_sha"`
    SourceSHA string `json:"source_sha"`
    TargetSHA string `json:"target_sha"`
}
```

**Merge Cases**:
1. ✅ Unchanged → Keep either version
2. ✅ Modified in source only → Use source
3. ✅ Modified in target only → Use target
4. ✅ Added in source only → Include from source
5. ✅ Added in target only → Include from target
6. ⚠️ Modified in both → **CONFLICT**

---

## Performance Considerations

### Content-Addressable Benefits

- **Deduplication**: Same content stored once
- **Fast Lookups**: SHA-1 is document key (O(1) retrieval)
- **Immutability**: No updates, only inserts
- **Cache-Friendly**: Objects never change

### Query Optimization

**AQL Index Recommendations**:
```aql
// Index on git_objects
CREATE INDEX idx_repo_type ON git_objects (repo_id, type)

// Index on git_refs
CREATE INDEX idx_repo_ref ON git_refs (repo_id, type)
```

**Fast Queries**:
```aql
// Get all commits for repository
FOR obj IN git_objects
  FILTER obj.repo_id == @repoID AND obj.type == "commit"
  SORT obj.created DESC
  RETURN obj

// Get current branch head
FOR ref IN git_refs
  FILTER ref.repo_id == @repoID AND ref._key == "refs/heads/main"
  RETURN ref.target
```

---

## Future Enhancements

1. **Garbage Collection**: Remove unreachable objects
2. **Pack Files**: Compress object storage
3. **Delta Compression**: Store diffs instead of full content
4. **Shallow Clones**: Partial history
5. **Large File Support**: External blob storage
6. **Concurrent Writes**: Optimistic locking on refs
7. **Reflog**: Track reference history
8. **Hooks**: Pre-commit, post-commit events

---

## Related Documentation

- [File Explorer](file-explorer.md) - High-level file operations
- [Collaborative Editing](collaborative-editing.md) - Sectioned documents and AI merging
- [Pull Requests](pull-requests.md) - Code review workflow
- [Git-Based Document System](git-based-document-system.md) - Architecture overview
