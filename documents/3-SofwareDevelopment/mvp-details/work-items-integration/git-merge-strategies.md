# Git Merge Strategies

**Related Tasks**: MVP-WI-007 (Pull Requests), MVP-WI-011 (AI-Assisted Merging)  
**Status**: 📋 Planned

## Overview

This document describes merge algorithms and conflict handling for the Git operations layer. It covers three-way merge implementation, conflict detection, and integration with AI-assisted resolution.

**For core Git operations** (object storage, commits, branches), see [Git Operations](git-operations.md).

---

## Three-Way Merge

### Merge Workflow

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

type MergeResult struct {
    Status      string     `json:"status"`       // "merged", "conflict"
    CommitSHA   string     `json:"commit_sha"`
    MergeCommit string     `json:"merge_commit"`
    Conflicts   []Conflict `json:"conflicts,omitempty"`
}
```

---

## File-Level Merge Algorithm

### mergeTreesThreeWay Implementation

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
```

### Helper Functions

```go
// toEntryMap converts tree entries to path→entry map
func toEntryMap(tree *GitTree) map[string]*TreeEntry {
    m := make(map[string]*TreeEntry)
    if tree == nil {
        return m
    }
    
    for i := range tree.Entries {
        m[tree.Entries[i].Path] = &tree.Entries[i]
    }
    return m
}

// getAllPaths returns union of all paths from 3 trees
func getAllPaths(base, source, target *GitTree) []string {
    pathSet := make(map[string]bool)
    
    if base != nil {
        for _, e := range base.Entries {
            pathSet[e.Path] = true
        }
    }
    if source != nil {
        for _, e := range source.Entries {
            pathSet[e.Path] = true
        }
    }
    if target != nil {
        for _, e := range target.Entries {
            pathSet[e.Path] = true
        }
    }
    
    paths := []string{}
    for p := range pathSet {
        paths = append(paths, p)
    }
    sort.Strings(paths)
    return paths
}
```

---

## Merge Cases Explained

| Case | Base | Source | Target | Result | Explanation |
|------|------|--------|--------|--------|-------------|
| 1 | ✅ | ✅ (same) | ✅ (same) | **Auto-merge** | File unchanged in both branches |
| 2 | ✅ | ✅ (modified) | ✅ (unchanged) | **Use source** | Only source branch modified the file |
| 3 | ✅ | ✅ (unchanged) | ✅ (modified) | **Use target** | Only target branch modified the file |
| 4 | ❌ | ✅ (new) | ❌ | **Use source** | File added in source branch only |
| 5 | ❌ | ❌ | ✅ (new) | **Use target** | File added in target branch only |
| 6 | ✅ | ✅ (modified) | ✅ (modified) | **CONFLICT** | Both branches modified the same file |

### Case 6 Conflict Handling

When both branches modify the same file, the system:

1. **Detects conflict**: SHA comparison shows different versions
2. **Records conflict data**:
   ```go
   Conflict{
       Path:      "documents/requirements.md",
       BaseSHA:   "abc123",  // Original version
       SourceSHA: "def456",  // Draft branch version
       TargetSHA: "789abc",  // Main branch version
   }
   ```
3. **Triggers AI resolution workflow** (see [AI Conflict Resolution](ai-conflict-resolution.md))
4. **Awaits human approval** before completing merge

---

## Conflict Data Model

```go
type Conflict struct {
    Path      string `json:"path"`        // File path with conflict
    BaseSHA   string `json:"base_sha"`    // Common ancestor version
    SourceSHA string `json:"source_sha"`  // Source branch version
    TargetSHA string `json:"target_sha"`  // Target branch version
}
```

**Usage in Pull Request**:
```go
type PullRequest struct {
    // ... (other fields)
    
    Status      string     `json:"status"`       // "conflict" if conflicts exist
    Conflicts   []Conflict `json:"conflicts"`    // List of conflicting files
    AIResolution *AIConflictResolution `json:"ai_resolution,omitempty"`
}
```

---

## Finding Merge Base

### findMergeBase Algorithm

```go
// findMergeBase finds common ancestor of two commits
func (g *GitOps) findMergeBase(commit1, commit2 *GitCommit) *GitCommit {
    // Simple implementation: walk back from both commits
    ancestors1 := g.getAncestors(commit1)
    ancestors2 := g.getAncestors(commit2)
    
    // Find first common ancestor
    for _, sha := range ancestors1 {
        if contains(ancestors2, sha) {
            commit, _ := g.GetCommit(sha)
            return commit
        }
    }
    
    return nil // No common ancestor (separate histories)
}

// getAncestors walks commit history
func (g *GitOps) getAncestors(commit *GitCommit) []string {
    ancestors := []string{commit.SHA}
    
    for _, parentSHA := range commit.Parents {
        parent, _ := g.GetCommit(parentSHA)
        ancestors = append(ancestors, g.getAncestors(parent)...)
    }
    
    return ancestors
}
```

**Example**:
```
main:    A---B---C---E
                 \
draft:            D---F

Merge base = C (last common ancestor)
```

---

## Performance Optimization

### Merge Performance Characteristics

| Operation | Complexity | Optimization |
|-----------|-----------|--------------|
| Get tree objects | O(1) per tree | Content-addressable storage |
| Compare file SHAs | O(n) files | Hash comparison only (no content) |
| Find merge base | O(d) depth | Depth typically < 100 commits |
| Create merge commit | O(n) files | Parallel tree writes |

### Fast Conflict Detection

**Key insight**: Only SHA comparison, no content diffing needed.

```go
// Fast check: are files different?
if sourceEntry.SHA != targetEntry.SHA {
    // CONFLICT - don't read content yet
    conflicts = append(conflicts, Conflict{...})
}
```

**Content only loaded when**:
- AI resolution needed (see [AI Conflict Resolution](ai-conflict-resolution.md))
- Human manual resolution requested
- Displaying conflict details in UI

---

## Integration with Sectioned Documents

For Markdown files using sectioned document format:

1. **File-level merge detects conflict** (this layer)
2. **Section-level merge attempted** (see [Sectioned Documents](sectioned-documents.md))
3. **AI resolution if section merge fails** (see [AI Conflict Resolution](ai-conflict-resolution.md))

**Workflow**:
```
File conflict detected
    ↓
Check if file is sectioned document (.md with --- sections)
    ↓
YES → Section-level merge
    ↓
Section conflict?
    ↓
YES → AI resolution
    ↓
Human review
```

---

## Future Enhancements

1. **Octopus Merge**: Merge >2 branches simultaneously
2. **Recursive Merge**: Better merge base selection for complex histories
3. **Rename Detection**: Track file renames across branches
4. **Binary File Merging**: Strategy for non-text files
5. **Merge Commit Metadata**: Record merge strategy used
6. **Cherry-Pick**: Apply individual commits across branches
7. **Rebase**: Replay commits on top of another branch
8. **Conflict Caching**: Remember conflict resolutions for similar cases

---

## Related Documentation

- [Git Operations](git-operations.md) - Core Git implementation (object storage, commits, branches)
- [Sectioned Documents](sectioned-documents.md) - Section-level merging for Markdown files
- [AI Conflict Resolution](ai-conflict-resolution.md) - AI-powered merge conflict resolution
- [Pull Requests](pull-requests.md) - Code review workflow
- [Git-Based Document System](git-based-document-system.md) - Architecture overview
