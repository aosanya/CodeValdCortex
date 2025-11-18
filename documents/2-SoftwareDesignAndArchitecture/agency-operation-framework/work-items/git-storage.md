# Git Storage in ArangoDB

This document details how git objects are stored in ArangoDB, enabling both traditional git operations and advanced graph-based analytics.

## Overview

Every git object (blob, tree, commit, ref) is stored as an ArangoDB document with:
- **Document properties**: Native git metadata
- **Full-text indexing**: Searchable content
- **Graph edges**: Relationships between objects
- **Multi-model**: Both document queries and graph traversals

## Git Object Collections

### GitBlob Collection

Stores file content with searchable text.

```typescript
interface GitBlob {
  _key: string;           // Hash of content (SHA-1)
  _id: string;            // Full document ID: "git_objects/{hash}"
  type: "blob";
  repo_id: string;        // Repository identifier
  size: number;           // Size in bytes
  content_raw: Buffer;    // Original binary content
  content_text?: string;  // Extracted text (if text file)
  path: string;           // Current path (from tree)
  language?: string;      // Detected programming language
  encoding: string;       // File encoding (utf-8, binary, etc.)
  stored_at: ISODateTime;
}
```

**Indexes**:
- Primary: `_key` (hash)
- Full-text: `content_text` with analyzer `text_en_code`
- Hash: `repo_id`
- Persistent: `[repo_id, path]`

**Example Document**:
```json
{
  "_key": "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3",
  "_id": "git_objects/a94a8fe5ccb19ba61c4c0873d391e987982fbbd3",
  "type": "blob",
  "repo_id": "codevaldcortex",
  "size": 1247,
  "content_text": "package main\n\nimport \"fmt\"\n\nfunc main() {...}",
  "path": "cmd/main.go",
  "language": "go",
  "encoding": "utf-8",
  "stored_at": "2025-11-07T10:30:00Z"
}
```

### GitTree Collection

Stores directory snapshots.

```typescript
interface GitTree {
  _key: string;           // Hash of tree
  _id: string;
  type: "tree";
  repo_id: string;
  entries: TreeEntry[];   // Directory contents
  stored_at: ISODateTime;
}

interface TreeEntry {
  mode: string;           // File mode (100644, 100755, 040000)
  type: "blob" | "tree";
  name: string;           // File or directory name
  hash: string;           // SHA-1 of blob or subtree
}
```

**Example Document**:
```json
{
  "_key": "3c4e9cd789d88d8d89c1073191ef",
  "_id": "git_objects/3c4e9cd789d88d8d89c1073191ef",
  "type": "tree",
  "repo_id": "codevaldcortex",
  "entries": [
    {"mode": "100644", "type": "blob", "name": "main.go", "hash": "a94a8fe5..."},
    {"mode": "040000", "type": "tree", "name": "internal", "hash": "f3b5c8d2..."}
  ],
  "stored_at": "2025-11-07T10:30:00Z"
}
```

### GitCommit Collection

Stores commit metadata and authorship.

```typescript
interface GitCommit {
  _key: string;           // Commit hash (SHA-1)
  _id: string;
  type: "commit";
  repo_id: string;
  tree: string;           // Hash of tree object
  parents: string[];      // Parent commit hashes
  author: GitSignature;
  committer: GitSignature;
  message: string;        // Commit message
  work_item_id?: string;  // Link to work item (if automated)
  committed_at: ISODateTime;
  stored_at: ISODateTime;
}

interface GitSignature {
  name: string;
  email: string;
  timestamp: ISODateTime;
}
```

**Indexes**:
- Primary: `_key` (hash)
- Hash: `repo_id`
- Persistent: `work_item_id`
- Persistent: `author.email`
- Full-text: `message`

**Example Document**:
```json
{
  "_key": "e83c5163316f89bfbde7d9ab23ca2e25604af290",
  "_id": "git_commits/e83c5163316f89bfbde7d9ab23ca2e25604af290",
  "type": "commit",
  "repo_id": "codevaldcortex",
  "tree": "3c4e9cd789d88d8d89c1073191ef",
  "parents": ["b5d2d4c7f3a8e9c1d2b3a4f5e6d7c8a9"],
  "author": {
    "name": "LLM Agent",
    "email": "llm-agent@codevaldcortex.ai",
    "timestamp": "2025-11-07T10:30:00Z"
  },
  "committer": {
    "name": "Work Item Executor",
    "email": "executor@codevaldcortex.ai",
    "timestamp": "2025-11-07T10:30:00Z"
  },
  "message": "feat: implement user authentication API\n\nCloses #87",
  "work_item_id": "work_items/WI-001-87",
  "committed_at": "2025-11-07T10:30:00Z",
  "stored_at": "2025-11-07T10:30:05Z"
}
```

### GitRef Collection

Stores branches, tags, and HEAD.

```typescript
interface GitRef {
  _key: string;           // Ref name (e.g., "refs/heads/main")
  _id: string;
  type: "ref";
  repo_id: string;
  name: string;           // Full ref name
  target: string;         // Hash of commit
  ref_type: "branch" | "tag" | "HEAD";
  updated_at: ISODateTime;
}
```

**Example Documents**:
```json
[
  {
    "_key": "refs_heads_main",
    "_id": "git_refs/refs_heads_main",
    "type": "ref",
    "repo_id": "codevaldcortex",
    "name": "refs/heads/main",
    "target": "e83c5163316f89bfbde7d9ab23ca2e25604af290",
    "ref_type": "branch",
    "updated_at": "2025-11-07T10:30:00Z"
  },
  {
    "_key": "refs_tags_v1.0.0",
    "_id": "git_refs/refs_tags_v1.0.0",
    "type": "ref",
    "repo_id": "codevaldcortex",
    "name": "refs/tags/v1.0.0",
    "target": "a1b2c3d4e5f6...",
    "ref_type": "tag",
    "updated_at": "2025-11-01T14:00:00Z"
  }
]
```

## Storing Git Objects from Gitea

### Sync Strategy

Two approaches for populating ArangoDB:

**1. Webhook-Triggered Sync** (Real-time)
```go
func (s *GitSync) OnPushEvent(payload *gitea.PushPayload) error {
    for _, commit := range payload.Commits {
        // 1. Fetch commit from Gitea
        commitObj := s.gitea.GetCommit(commit.ID)
        
        // 2. Store commit in ArangoDB
        s.storeCommit(commitObj)
        
        // 3. Fetch tree recursively
        tree := s.gitea.GetTree(commitObj.Tree.SHA)
        s.storeTree(tree, commitObj.Tree.SHA)
        
        // 4. Store all blobs
        for _, entry := range tree.Entries {
            if entry.Type == "blob" {
                blob := s.gitea.GetBlob(entry.SHA)
                s.storeBlob(blob, entry.Path)
            }
        }
    }
    return nil
}
```

**2. Periodic Full Sync** (Batch)
```go
func (s *GitSync) FullSync(repoID string) error {
    // Get all refs
    refs := s.gitea.ListRefs(repoID)
    
    for _, ref := range refs {
        s.storeRef(ref)
        
        // Walk commit history
        commits := s.gitea.ListCommits(ref.Target, 1000)
        for _, commit := range commits {
            s.storeCommit(commit)
            s.storeTree(commit.Tree, commit.Tree.SHA)
        }
    }
    
    return nil
}
```

### Extracting Text Content

```go
func (s *GitSync) storeBlob(blob *gitea.Blob, path string) error {
    doc := GitBlob{
        Key:        blob.SHA,
        Type:       "blob",
        RepoID:     s.repoID,
        Size:       blob.Size,
        ContentRaw: blob.Content,
        Path:       path,
        Encoding:   detectEncoding(blob.Content),
        StoredAt:   time.Now(),
    }
    
    // Extract text if readable
    if isTextFile(path) && doc.Encoding == "utf-8" {
        doc.ContentText = string(blob.Content)
        doc.Language = detectLanguage(path, blob.Content)
    }
    
    _, err := s.db.Collection("git_objects").CreateDocument(ctx, doc)
    return err
}

func isTextFile(path string) bool {
    ext := filepath.Ext(path)
    textExts := []string{".go", ".js", ".py", ".md", ".txt", ".json", ".yaml"}
    for _, te := range textExts {
        if ext == te {
            return true
        }
    }
    return false
}
```

## Full-Text Search Configuration

### Analyzers

```json
{
  "name": "text_en_code",
  "type": "text",
  "properties": {
    "locale": "en",
    "case": "lower",
    "stopwords": [],
    "accent": false,
    "stemming": true
  },
  "features": ["frequency", "norm", "position"]
}
```

### Search View

```json
{
  "name": "git_content_search",
  "type": "arangosearch",
  "links": {
    "git_objects": {
      "analyzers": ["text_en_code"],
      "fields": {
        "content_text": {
          "analyzers": ["text_en_code"]
        },
        "path": {
          "analyzers": ["identity"]
        },
        "language": {
          "analyzers": ["identity"]
        }
      },
      "includeAllFields": false
    }
  }
}
```

### Search Query Example

```go
func (r *GitRepo) SearchCode(term string, language string) ([]SearchResult, error) {
    query := `
        FOR doc IN git_content_search
            SEARCH ANALYZER(
                doc.content_text IN TOKENS(@term, "text_en_code"),
                "text_en_code"
            )
            FILTER doc.repo_id == @repoID
            FILTER @language == null OR doc.language == @language
            LET score = BM25(doc)
            SORT score DESC
            LIMIT 50
            RETURN {
                path: doc.path,
                language: doc.language,
                score: score,
                snippet: SUBSTRING(doc.content_text, 0, 200)
            }
    `
    
    cursor, err := r.db.Query(ctx, query, map[string]interface{}{
        "term":     term,
        "repoID":   r.repoID,
        "language": language,
    })
    
    var results []SearchResult
    for cursor.HasMore() {
        var result SearchResult
        cursor.ReadDocument(ctx, &result)
        results = append(results, result)
    }
    
    return results, err
}
```

## Storage Optimization

### Delta Storage

For large repositories, consider storing only diffs:

```go
type GitBlobDelta struct {
    Key        string `json:"_key"`
    Type       string `json:"type"`        // "blob-delta"
    BaseHash   string `json:"base_hash"`   // Reference to base blob
    DiffPatch  string `json:"diff_patch"`  // Binary diff
    Size       int64  `json:"size"`
}
```

### Compression

```go
func (s *GitSync) storeBlob(blob *gitea.Blob, path string) error {
    // Compress large text files
    if blob.Size > 10*1024 && isTextFile(path) {
        compressed := gzip.Compress(blob.Content)
        doc.ContentRaw = compressed
        doc.Compressed = true
    }
    
    // ...rest of storage logic
}
```

### Deduplication

Git's content-addressable storage naturally deduplicates via SHA-1 hashing. Identical files across commits share the same blob document in ArangoDB.

---

**See Also**:
- [Graph Queries](./graph-queries.md) - Querying git relationships
- [GitOps Workflow](./gitops-workflow.md) - How work items use git storage
