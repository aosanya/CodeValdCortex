# Graph Queries

This document provides comprehensive examples of multi-dimensional graph queries across the four named graphs in ArangoDB.

## Named Graphs Overview

CodeValdCortex uses four interconnected graphs:

1. **commit_graph** - Git commit history and ancestry
2. **code_graph** - Code dependencies and relationships  
3. **workflow_graph** - Work items → commits → agents
4. **knowledge_graph** - Agent expertise tracking

## Graph Definitions

### 1. Commit Graph

```json
{
  "name": "commit_graph",
  "edgeDefinitions": [
    {
      "collection": "commit_parents",
      "from": ["git_commits"],
      "to": ["git_commits"]
    }
  ]
}
```

**Purpose**: Navigate commit history, find ancestors, detect merge commits

### 2. Code Graph

```json
{
  "name": "code_graph",
  "edgeDefinitions": [
    {
      "collection": "code_dependencies",
      "from": ["git_objects"],
      "to": ["git_objects"]
    }
  ]
}
```

**Purpose**: Analyze import/include relationships, detect circular dependencies

### 3. Workflow Graph

```json
{
  "name": "workflow_graph",
  "edgeDefinitions": [
    {
      "collection": "work_item_commits",
      "from": ["work_items"],
      "to": ["git_commits"]
    },
    {
      "collection": "commit_work_items",
      "from": ["git_commits"],
      "to": ["work_items"]
    },
    {
      "collection": "agent_work_items",
      "from": ["agents"],
      "to": ["work_items"]
    }
  ]
}
```

**Purpose**: Trace work items to code changes, find agent contributions

### 4. Knowledge Graph

```json
{
  "name": "knowledge_graph",
  "edgeDefinitions": [
    {
      "collection": "agent_code_expertise",
      "from": ["agents"],
      "to": ["git_objects"]
    }
  ]
}
```

**Purpose**: Find code experts, track agent specialization

## Commit Graph Queries

### Find Commit Ancestors

```go
// Get all ancestors of a commit (full history)
query := `
    FOR v, e, p IN 1..100 OUTBOUND
        @commitHash
        GRAPH 'commit_graph'
        RETURN {
            commit: v._key,
            message: v.message,
            author: v.author.name,
            depth: LENGTH(p.edges)
        }
`

// Parameters
params := map[string]interface{}{
    "commitHash": "git_commits/e83c5163316f89bfbde7d9ab23ca2e25604af290",
}
```

### Find Common Ancestor

```go
// Find merge base of two branches
query := `
    LET commit1Ancestors = (
        FOR v IN 1..100 OUTBOUND @commit1 GRAPH 'commit_graph'
            RETURN v._key
    )
    
    LET commit2Ancestors = (
        FOR v IN 1..100 OUTBOUND @commit2 GRAPH 'commit_graph'
            RETURN v._key
    )
    
    LET common = INTERSECTION(commit1Ancestors, commit2Ancestors)
    
    FOR commitKey IN common
        LET commit = DOCUMENT(CONCAT('git_commits/', commitKey))
        SORT commit.committed_at DESC
        LIMIT 1
        RETURN commit
`

// Parameters
params := map[string]interface{}{
    "commit1": "git_commits/abc123...",
    "commit2": "git_commits/def456...",
}
```

### Detect Merge Commits

```go
// Find all merge commits (multiple parents)
query := `
    FOR commit IN git_commits
        FILTER LENGTH(commit.parents) > 1
        SORT commit.committed_at DESC
        RETURN {
            commit: commit._key,
            message: commit.message,
            parents: commit.parents,
            merged_at: commit.committed_at
        }
`
```

## Code Graph Queries

### Find Direct Dependencies

```go
// What files does this file import?
query := `
    FOR v, e IN 1..1 OUTBOUND
        @fileHash
        GRAPH 'code_graph'
        RETURN {
            file: v.path,
            language: v.language,
            import_type: e.type
        }
`

// Parameters
params := map[string]interface{}{
    "fileHash": "git_objects/a94a8fe5ccb19ba61c4c0873d391e987982fbbd3",
}
```

### Find Dependents (Reverse)

```go
// What files import this file?
query := `
    FOR v, e IN 1..1 INBOUND
        @fileHash
        GRAPH 'code_graph'
        RETURN {
            file: v.path,
            language: v.language
        }
`
```

### Impact Analysis (Transitive)

```go
// What breaks if I change this file?
query := `
    FOR v, e, p IN 1..10 INBOUND
        @fileHash
        GRAPH 'code_graph'
        OPTIONS {uniqueVertices: "global"}
        RETURN {
            file: v.path,
            distance: LENGTH(p.edges),
            path: p.vertices[* RETURN CURRENT.path]
        }
`
```

### Circular Dependency Detection

```go
// Find circular dependencies
query := `
    FOR file IN git_objects
        FILTER file.type == "blob"
        
        LET cycles = (
            FOR v, e, p IN 2..10 OUTBOUND file._id
                GRAPH 'code_graph'
                FILTER v._id == file._id
                LIMIT 1
                RETURN p.vertices[* RETURN CURRENT.path]
        )
        
        FILTER LENGTH(cycles) > 0
        RETURN {
            file: file.path,
            cycle: cycles[0]
        }
`
```

### Dependency Depth Analysis

```go
// How deep is the dependency tree?
query := `
    FOR file IN git_objects
        FILTER file.type == "blob"
        FILTER file.language == "go"
        
        LET maxDepth = MAX(
            FOR v, e, p IN 1..20 OUTBOUND file._id
                GRAPH 'code_graph'
                RETURN LENGTH(p.edges)
        )
        
        SORT maxDepth DESC
        LIMIT 20
        RETURN {
            file: file.path,
            max_dependency_depth: maxDepth
        }
`
```

## Workflow Graph Queries

### Trace Work Item to Code

```go
// What commits came from this work item?
query := `
    FOR commit, e IN 1..1 OUTBOUND
        @workItemID
        GRAPH 'workflow_graph'
        RETURN {
            commit: commit._key,
            message: commit.message,
            author: commit.author,
            files_changed: LENGTH(commit.tree.entries)
        }
`

// Parameters
params := map[string]interface{}{
    "workItemID": "work_items/WI-001-87",
}
```

### Find Work Items for File

```go
// Which work items modified this file?
query := `
    // Step 1: Find commits that touched this file
    FOR commit IN git_commits
        FILTER commit.repo_id == @repoID
        LET tree = DOCUMENT(CONCAT('git_objects/', commit.tree))
        FILTER @fileHash IN tree.entries[* RETURN CURRENT.hash]
        
        // Step 2: Find work items for those commits
        FOR wi IN 1..1 INBOUND commit._id
            GRAPH 'workflow_graph'
            RETURN DISTINCT {
                work_item: wi._key,
                title: wi.title,
                commit: commit._key,
                committed_at: commit.committed_at
            }
`

// Parameters
params := map[string]interface{}{
    "repoID":   "codevaldcortex",
    "fileHash": "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3",
}
```

### Agent Contribution Analysis

```go
// What has this agent worked on?
query := `
    FOR wi, e IN 1..1 OUTBOUND
        @agentID
        GRAPH 'workflow_graph'
        
        LET commits = (
            FOR commit IN 1..1 OUTBOUND wi._id
                GRAPH 'workflow_graph'
                RETURN commit
        )
        
        RETURN {
            work_item: wi.title,
            status: wi.status,
            commits: LENGTH(commits),
            completed_at: wi.completed_at
        }
`

// Parameters
params := map[string]interface{}{
    "agentID": "agents/llm-agent-01",
}
```

## Knowledge Graph Queries

### Find Code Expert

```go
// Who knows this file best?
query := `
    FOR agent, edge IN 1..1 INBOUND
        @fileHash
        GRAPH 'knowledge_graph'
        SORT edge.expertise DESC, edge.last_touch DESC
        LIMIT 5
        RETURN {
            agent: agent.name,
            email: agent.email,
            expertise: edge.expertise,
            commits: LENGTH(edge.commits),
            last_contribution: edge.last_touch,
            recency_score: DATE_DIFF(edge.last_touch, DATE_NOW(), 'days') / -365
        }
`

// Parameters
params := map[string]interface{}{
    "fileHash": "git_objects/a94a8fe5ccb19ba61c4c0873d391e987982fbbd3",
}
```

### Agent Expertise Profile

```go
// What code does this agent know?
query := `
    FOR file, edge IN 1..1 OUTBOUND
        @agentID
        GRAPH 'knowledge_graph'
        SORT edge.expertise DESC
        LIMIT 20
        RETURN {
            file: file.path,
            language: file.language,
            expertise: edge.expertise,
            commits: LENGTH(edge.commits),
            last_touch: edge.last_touch,
            contribution_score: edge.expertise * (1 + LENGTH(edge.commits) * 0.1)
        }
`

// Parameters
params := map[string]interface{}{
    "agentID": "agents/llm-agent-01",
}
```

### Build Knowledge Graph

```go
// Automatically build expertise edges from commit history
func (r *Repo) BuildKnowledgeGraph() error {
    query := `
        FOR commit IN git_commits
            FILTER commit.repo_id == @repoID
            
            LET tree = DOCUMENT(CONCAT('git_objects/', commit.tree))
            
            FOR entry IN tree.entries
                FILTER entry.type == "blob"
                
                // Create or update agent expertise edge
                UPSERT {
                    _from: CONCAT('agents/', commit.author.email),
                    _to: CONCAT('git_objects/', entry.hash)
                }
                INSERT {
                    _from: CONCAT('agents/', commit.author.email),
                    _to: CONCAT('git_objects/', entry.hash),
                    commits: [commit._key],
                    expertise: 1.0,
                    last_touch: commit.committed_at
                }
                UPDATE {
                    commits: APPEND(OLD.commits, commit._key, true),
                    expertise: MIN([OLD.expertise + 0.1, 1.0]),
                    last_touch: MAX([OLD.last_touch, commit.committed_at])
                }
                IN agent_code_expertise
    `
    
    _, err := r.db.Query(ctx, query, map[string]interface{}{
        "repoID": r.repoID,
    })
    
    return err
}
```

## Multi-Graph Queries

### Complete Traceability Chain

```go
// From work item → commits → files → experts
query := `
    LET workItem = DOCUMENT(@workItemID)
    
    // 1. Get commits from this work item
    LET commits = (
        FOR commit IN 1..1 OUTBOUND workItem._id
            GRAPH 'workflow_graph'
            RETURN commit
    )
    
    // 2. Get files modified
    LET files = FLATTEN(
        FOR commit IN commits
            LET tree = DOCUMENT(CONCAT('git_objects/', commit.tree))
            FOR entry IN tree.entries
                FILTER entry.type == "blob"
                LET file = DOCUMENT(CONCAT('git_objects/', entry.hash))
                RETURN file
    )
    
    // 3. Find experts for those files
    LET experts = FLATTEN(
        FOR file IN files
            FOR agent, edge IN 1..1 INBOUND file._id
                GRAPH 'knowledge_graph'
                RETURN {
                    file: file.path,
                    expert: agent.name,
                    expertise: edge.expertise
                }
    )
    
    RETURN {
        work_item: workItem.title,
        commits: commits[* RETURN CURRENT._key],
        files: files[* RETURN CURRENT.path],
        experts: experts
    }
`
```

### File Impact + Expert Assignment

```go
// For a file change, find impact and suggest reviewers
query := `
    LET file = DOCUMENT(@fileHash)
    
    // 1. Impact analysis (code graph)
    LET impact = (
        FOR v IN 1..5 INBOUND file._id
            GRAPH 'code_graph'
            OPTIONS {uniqueVertices: "global"}
            RETURN v.path
    )
    
    // 2. Find experts (knowledge graph)
    LET experts = (
        FOR agent, edge IN 1..1 INBOUND file._id
            GRAPH 'knowledge_graph'
            SORT edge.expertise DESC
            LIMIT 3
            RETURN {
                agent: agent.name,
                expertise: edge.expertise
            }
    )
    
    // 3. Calculate impact score
    LET impactScore = LENGTH(impact) * 10 + file.size / 1000
    
    RETURN {
        file: file.path,
        impact_score: impactScore,
        affected_files: LENGTH(impact),
        suggested_reviewers: experts,
        require_reviews: impactScore > 50 ? 2 : 1
    }
`
```

### Agent Productivity Dashboard

```go
// Comprehensive agent analytics
query := `
    FOR agent IN agents
        // Work items
        LET workItems = (
            FOR wi IN 1..1 OUTBOUND agent._id
                GRAPH 'workflow_graph'
                RETURN wi
        )
        
        // Commits (via work items)
        LET commits = FLATTEN(
            FOR wi IN workItems
                FOR commit IN 1..1 OUTBOUND wi._id
                    GRAPH 'workflow_graph'
                    RETURN commit
        )
        
        // Code expertise
        LET expertise = (
            FOR file, edge IN 1..1 OUTBOUND agent._id
                GRAPH 'knowledge_graph'
                RETURN {file: file.path, expertise: edge.expertise}
        )
        
        RETURN {
            agent: agent.name,
            work_items_total: LENGTH(workItems),
            work_items_completed: LENGTH(workItems[* FILTER CURRENT.status == "completed"]),
            commits: LENGTH(commits),
            files_known: LENGTH(expertise),
            avg_expertise: AVG(expertise[* RETURN CURRENT.expertise]),
            top_expertise: expertise[* SORT CURRENT.expertise DESC LIMIT 5]
        }
`
```

## Performance Optimization

### Index Strategy

```go
// Create indexes for common traversal patterns
db.Collection("commit_parents").EnsureIndex(ctx, arangodb.IndexOptions{
    Type:   "persistent",
    Fields: []string{"_from", "_to"},
})

db.Collection("code_dependencies").EnsureIndex(ctx, arangodb.IndexOptions{
    Type:   "persistent", 
    Fields: []string{"_from"},
})

db.Collection("agent_code_expertise").EnsureIndex(ctx, arangodb.IndexOptions{
    Type:   "persistent",
    Fields: []string{"_from", "expertise"},
})
```

### Query Caching

```go
// Cache frequent queries
type GraphCache struct {
    cache *ristretto.Cache
}

func (c *GraphCache) GetOrQuery(key string, queryFunc func() interface{}) interface{} {
    if val, found := c.cache.Get(key); found {
        return val
    }
    
    result := queryFunc()
    c.cache.SetWithTTL(key, result, 1, 5*time.Minute)
    return result
}

// Usage
expertResult := cache.GetOrQuery("expert_for_"+fileHash, func() interface{} {
    return repo.FindCodeExpert(fileHash)
})
```

### Limit Traversal Depth

```go
// Avoid expensive deep traversals
query := `
    FOR v IN 1..5 OUTBOUND @start  // Max depth: 5
        GRAPH 'code_graph'
        OPTIONS {
            uniqueVertices: "global",  // Prevent cycles
            bfs: true                  // Breadth-first for better perf
        }
        LIMIT 1000                     // Cap results
        RETURN v
`
```

---

**See Also**:
- [Git Storage](./git-storage.md) - Git object schema
- [Data Models](./data-models.md) - Complete edge schemas
