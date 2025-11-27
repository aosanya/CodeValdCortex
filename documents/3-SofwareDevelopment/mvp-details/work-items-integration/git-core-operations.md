# Git Core Operations

**Related Tasks**: MVP-WI-005 (Git Core Layer), MVP-WI-007 (Pull Requests)  
**Status**: MVP-WI-005 ✅ Complete | MVP-WI-007 📋 Planned

## Overview

This document describes Git read/write operations implemented in ArangoDB, including blob storage, tree construction, commit creation, and reference management.

**For data model definitions**, see [Git Data Models](git-data-models.md).  
**For merge algorithms**, see [Git Merge Strategies](git-merge-strategies.md).

---

## Write Operations

### WriteBlob (Store File Content)

```go
package gitops

import (
    "crypto/sha1"
    "encoding/hex"
    "fmt"
    "time"
    "github.com/arangodb/go-driver"
)

type GitOps struct {
    db driver.Database
}

// WriteBlob stores file content as a blob object
func (g *GitOps) WriteBlob(ctx context.Context, repoID, content string) (string, error) {
    // Calculate SHA-1 hash (Git blob format)
    hash := sha1.Sum([]byte(fmt.Sprintf("blob %d\x00%s", len(content), content)))
    sha := hex.EncodeToString(hash[:])
    
    // Create blob object
    obj := &GitObject{
        Key:     sha,
        Type:    "blob",
        Size:    len(content),
        Content: content,
        RepoID:  repoID,
        Created: time.Now(),
    }
    
    // Insert into git_objects collection
    col, _ := g.db.Collection(ctx, "git_objects")
    _, err := col.CreateDocument(ctx, obj)
    
    // Ignore "document already exists" error (idempotent)
    if driver.IsConflict(err) {
        return sha, nil
    }
    
    return sha, err
}
```

**Idempotency**:
- Same content → same SHA-1 hash → same document key
- If document already exists, ArangoDB returns conflict error
- Return success (blob already stored, no action needed)
- **Result**: No duplicate storage, deduplication at database level

**Usage Example**:
```go
gitOps := NewGitOps(db)

// Store README.md content
content := "# My Project\n\nWelcome to my project."
sha, err := gitOps.WriteBlob(ctx, "instance-001", content)
// sha = "abc123def456..." (40-char hex string)
```

---

### WriteTree (Store Directory Structure)

```go
// WriteTree stores directory structure as a tree object
func (g *GitOps) WriteTree(ctx context.Context, repoID string, entries []TreeEntry) (string, error) {
    // Serialize tree entries
    var buf bytes.Buffer
    for _, e := range entries {
        fmt.Fprintf(&buf, "%s %s %s\t%s\n", e.Mode, e.Type, e.SHA, e.Path)
    }
    content := buf.String()
    
    // Calculate SHA-1 hash (Git tree format)
    hash := sha1.Sum([]byte(fmt.Sprintf("tree %d\x00%s", len(content), content)))
    sha := hex.EncodeToString(hash[:])
    
    // Create tree object
    obj := &GitObject{
        Key:     sha,
        Type:    "tree",
        Size:    len(content),
        Content: content,
        RepoID:  repoID,
        Created: time.Now(),
    }
    
    // Insert into git_objects collection
    col, _ := g.db.Collection(ctx, "git_objects")
    _, err := col.CreateDocument(ctx, obj)
    
    if driver.IsConflict(err) {
        return sha, nil
    }
    
    return sha, err
}
```

---

### Commit (Create Version Snapshot)

```go
// Commit creates a commit object (version snapshot)
func (g *GitOps) Commit(ctx context.Context, repoID, treeSHA string, parents []string, author Author, message string) (string, error) {
    // Create commit struct
    commit := &GitCommit{
        Tree:      treeSHA,
        Parents:   parents,
        Author:    author,
        Committer: Author{Name: "codevaldcortex", Email: "system@codevaldcortex", When: time.Now()},
        Message:   message,
        Timestamp: time.Now(),
    }
    
    // Serialize commit
    content := serializeCommit(commit)
    
    // Calculate SHA-1 hash (Git commit format)
    hash := sha1.Sum([]byte(fmt.Sprintf("commit %d\x00%s", len(content), content)))
    sha := hex.EncodeToString(hash[:])
    
    // Create commit object
    obj := &GitObject{
        Key:     sha,
        Type:    "commit",
        Size:    len(content),
        Content: content,
        RepoID:  repoID,
        Created: time.Now(),
    }
    
    // Insert into git_objects collection
    col, _ := g.db.Collection(ctx, "git_objects")
    _, err := col.CreateDocument(ctx, obj)
    
    if driver.IsConflict(err) {
        return sha, nil
    }
    
    return sha, err
}

// serializeCommit converts GitCommit to Git commit format
func serializeCommit(c *GitCommit) string {
    var buf bytes.Buffer
    
    fmt.Fprintf(&buf, "tree %s\n", c.Tree)
    
    for _, parent := range c.Parents {
        fmt.Fprintf(&buf, "parent %s\n", parent)
    }
    
    fmt.Fprintf(&buf, "author %s <%s> %d\n", c.Author.Name, c.Author.Email, c.Author.When.Unix())
    fmt.Fprintf(&buf, "committer %s <%s> %d\n", c.Committer.Name, c.Committer.Email, c.Committer.When.Unix())
    
    fmt.Fprintf(&buf, "\n%s", c.Message)
    
    return buf.String()
}
```

---

### UpdateRef (Move Branch Pointer)

```go
// UpdateRef moves branch pointer to target commit
func (g *GitOps) UpdateRef(ctx context.Context, repoID, refName, commitSHA string) error {
    ref := &GitRef{
        Key:     refName,
        RepoID:  repoID,
        Type:    "branch",
        Target:  commitSHA,
        Updated: time.Now(),
    }
    
    col, _ := g.db.Collection(ctx, "git_refs")
    
    // Upsert: Update if exists, create if not
    _, err := col.UpdateDocument(ctx, refName, ref)
    if driver.IsNotFound(err) {
        ref.Created = time.Now()
        _, err = col.CreateDocument(ctx, ref)
    }
    
    return err
}
```

**Atomic Branch Updates**:
- ArangoDB document updates are atomic
- No race conditions on branch pointer moves
- Failed update automatically rolls back
- Concurrent updates handled by ArangoDB's MVCC

**Usage Example**:
```go
// Create main branch
err := gitOps.UpdateRef(ctx, "instance-001", "refs/heads/main", commitSHA)

// Create feature branch
err = gitOps.UpdateRef(ctx, "instance-001", "refs/heads/issue-123-jwt-auth", commitSHA2)

// Fast-forward main branch
err = gitOps.UpdateRef(ctx, "instance-001", "refs/heads/main", commitSHA2)
```

---

## Read Operations

### GetCommit

```go
// GetCommit retrieves commit object by SHA
func (g *GitOps) GetCommit(ctx context.Context, sha string) (*GitCommit, error) {
    var obj GitObject
    
    col, _ := g.db.Collection(ctx, "git_objects")
    _, err := col.ReadDocument(ctx, sha, &obj)
    if err != nil {
        return nil, err
    }
    
    // Deserialize commit content
    return deserializeCommit(obj.Content), nil
}

// deserializeCommit parses Git commit format
func deserializeCommit(content string) *GitCommit {
    commit := &GitCommit{
        Parents: []string{},
    }
    
    lines := strings.Split(content, "\n")
    messageStart := 0
    
    for i, line := range lines {
        if line == "" {
            messageStart = i + 1
            break
        }
        
        parts := strings.SplitN(line, " ", 2)
        key := parts[0]
        value := parts[1]
        
        switch key {
        case "tree":
            commit.Tree = value
        case "parent":
            commit.Parents = append(commit.Parents, value)
        case "author":
            commit.Author = parseAuthor(value)
        case "committer":
            commit.Committer = parseAuthor(value)
        }
    }
    
    commit.Message = strings.Join(lines[messageStart:], "\n")
    
    return commit
}
```

---

### GetTree

```go
// GetTree retrieves tree object by SHA
func (g *GitOps) GetTree(ctx context.Context, sha string) (*GitTree, error) {
    var obj GitObject
    
    col, _ := g.db.Collection(ctx, "git_objects")
    _, err := col.ReadDocument(ctx, sha, &obj)
    if err != nil {
        return nil, err
    }
    
    // Deserialize tree content
    return deserializeTree(obj.Content), nil
}

// deserializeTree parses Git tree format
func deserializeTree(content string) *GitTree {
    tree := &GitTree{
        Entries: []TreeEntry{},
    }
    
    lines := strings.Split(strings.TrimSpace(content), "\n")
    
    for _, line := range lines {
        parts := strings.Fields(line)
        if len(parts) < 4 {
            continue
        }
        
        entry := TreeEntry{
            Mode: parts[0],
            Type: parts[1],
            SHA:  parts[2],
            Path: parts[3],
        }
        
        tree.Entries = append(tree.Entries, entry)
    }
    
    return tree
}
```

---

### GetBlob

```go
// GetBlob retrieves file content by SHA
func (g *GitOps) GetBlob(ctx context.Context, sha string) (string, error) {
    var obj GitObject
    
    col, _ := g.db.Collection(ctx, "git_objects")
    _, err := col.ReadDocument(ctx, sha, &obj)
    if err != nil {
        return "", err
    }
    
    return obj.Content, nil
}
```

---

### ReadFile (Path-Based Lookup)

```go
// ReadFile reads file at specific path and commit
func (g *GitOps) ReadFile(ctx context.Context, repoID, commitSHA, filePath string) (string, error) {
    // 1. Get commit object
    commit, err := g.GetCommit(ctx, commitSHA)
    if err != nil {
        return "", err
    }
    
    // 2. Get root tree
    tree, err := g.GetTree(ctx, commit.Tree)
    if err != nil {
        return "", err
    }
    
    // 3. Navigate path (e.g., "documents/requirements.md")
    parts := strings.Split(filePath, "/")
    currentTree := tree
    
    for i, part := range parts {
        if i == len(parts)-1 {
            // Last part = file (blob)
            entry := findEntry(currentTree, part)
            if entry == nil {
                return "", fmt.Errorf("file not found: %s", filePath)
            }
            return g.GetBlob(ctx, entry.SHA)
        }
        
        // Directory - navigate deeper
        entry := findEntry(currentTree, part)
        if entry == nil || entry.Type != "tree" {
            return "", fmt.Errorf("directory not found: %s", part)
        }
        
        currentTree, err = g.GetTree(ctx, entry.SHA)
        if err != nil {
            return "", err
        }
    }
    
    return "", fmt.Errorf("file not found: %s", filePath)
}

// findEntry finds tree entry by path name
func findEntry(tree *GitTree, name string) *TreeEntry {
    for _, entry := range tree.Entries {
        if entry.Path == name {
            return &entry
        }
    }
    return nil
}
```

**Path Navigation**:
1. Start at commit's root tree
2. Split file path by `/` separator
3. For each path segment:
   - If last segment: lookup blob (file)
   - Otherwise: lookup tree (directory) and navigate into it
4. Return blob content

**Usage Example**:
```go
// Read file at specific commit
content, err := gitOps.ReadFile(
    ctx,
    "instance-001",
    "789abc012def...",              // Commit SHA
    "documents/requirements.md",    // File path
)
// content = "# Requirements\n\n..."
```

---

### ListRefs (List All Branches/Tags)

```go
// ListRefs lists all references (branches and tags)
func (g *GitOps) ListRefs(ctx context.Context, repoID string) ([]GitRef, error) {
    query := `
        FOR ref IN git_refs
          FILTER ref.repo_id == @repoID
          SORT ref._key ASC
          RETURN ref
    `
    
    bindVars := map[string]interface{}{
        "repoID": repoID,
    }
    
    cursor, err := g.db.Query(ctx, query, bindVars)
    if err != nil {
        return nil, err
    }
    defer cursor.Close()
    
    var refs []GitRef
    for cursor.HasMore() {
        var ref GitRef
        _, err := cursor.ReadDocument(ctx, &ref)
        if err != nil {
            return nil, err
        }
        refs = append(refs, ref)
    }
    
    return refs, nil
}
```

---

## Related Documentation

- **Git Data Models**: [git-data-models.md](git-data-models.md) - Object schemas, collections
- **Git Merge Strategies**: [git-merge-strategies.md](git-merge-strategies.md) - Three-way merge, conflict resolution
- **File Explorer**: [file-explorer.md](file-explorer.md) - High-level file operations
- **Pull Requests**: [pull-requests.md](pull-requests.md) - Code review workflow
- **Sectioned Documents**: [sectioned-documents.md](sectioned-documents.md) - Section-level merging

---

**Last Updated**: 2025-11-27  
**Status**: Architecture Defined - MVP-WI-005 ✅ Complete
