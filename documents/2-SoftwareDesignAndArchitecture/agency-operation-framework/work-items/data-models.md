# Data Models

Complete schema reference for all collections and edges in the work item system.

## Document Collections

### work_items

```typescript
interface WorkItem {
  _key: string;                    // Auto-generated
  _id: string;                     // "work_items/{key}"
  
  // Identity
  work_item_id: string;            // WI-{agency}-{type}-{timestamp}-{hash}
  agency_id: string;               // Reference to agency
  gitea_issue_id: number;          // Gitea issue index
  
  // Classification
  work_type: "document" | "software" | "proposal" | "analysis";
  priority: "P0" | "P1" | "P2" | "P3";
  
  // Status
  status: "pending" | "executing" | "review" | "merged" | "failed" | "completed";
  
  // Metadata
  title: string;
  description: string;
  labels: string[];
  
  // Execution
  assigned_agent_id?: string;
  started_at?: ISODateTime;
  completed_at?: ISODateTime;
  
  // Git integration
  branch_name?: string;
  merge_request_url?: string;
  commit_hashes?: string[];
  
  // Tracking
  idempotence_key: string;
  retry_count: number;
  error_message?: string;
  
  // Timestamps
  created_at: ISODateTime;
  updated_at: ISODateTime;
}
```

**Indexes**:
```javascript
[
  {type: "hash", fields: ["agency_id"]},
  {type: "hash", fields: ["gitea_issue_id"]},
  {type: "hash", fields: ["idempotence_key"], unique: true},
  {type: "persistent", fields: ["status", "created_at"]},
  {type: "persistent", fields: ["work_type", "status"]},
]
```

### git_objects (Blobs & Trees)

```typescript
interface GitBlob {
  _key: string;                    // SHA-1 hash
  _id: string;                     // "git_objects/{hash}"
  
  type: "blob";
  repo_id: string;
  
  // Content
  size: number;
  content_raw: Buffer;             // Binary content
  content_text?: string;           // Text content (if applicable)
  
  // Metadata
  path: string;                    // File path
  language?: string;               // Programming language
  encoding: string;                // utf-8, binary, etc.
  
  // Timestamps
  stored_at: ISODateTime;
}

interface GitTree {
  _key: string;                    // SHA-1 hash
  _id: string;                     // "git_objects/{hash}"
  
  type: "tree";
  repo_id: string;
  
  // Tree entries
  entries: TreeEntry[];
  
  // Timestamps
  stored_at: ISODateTime;
}

interface TreeEntry {
  mode: string;                    // 100644, 100755, 040000
  type: "blob" | "tree";
  name: string;
  hash: string;                    // SHA-1
}
```

**Indexes**:
```javascript
[
  {type: "hash", fields: ["repo_id"]},
  {type: "persistent", fields: ["repo_id", "path"]},
  {type: "fulltext", fields: ["content_text"], analyzer: "text_en_code"},
]
```

### git_commits

```typescript
interface GitCommit {
  _key: string;                    // Commit SHA-1
  _id: string;                     // "git_commits/{hash}"
  
  type: "commit";
  repo_id: string;
  
  // Tree and parents
  tree: string;                    // Tree hash
  parents: string[];               // Parent commit hashes
  
  // Authorship
  author: GitSignature;
  committer: GitSignature;
  
  // Message
  message: string;
  
  // Work item link
  work_item_id?: string;
  
  // Timestamps
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
```javascript
[
  {type: "hash", fields: ["repo_id"]},
  {type: "persistent", fields: ["work_item_id"]},
  {type: "persistent", fields: ["author.email"]},
  {type: "persistent", fields: ["committed_at"]},
  {type: "fulltext", fields: ["message"]},
]
```

### git_refs

```typescript
interface GitRef {
  _key: string;                    // Sanitized ref name
  _id: string;                     // "git_refs/{key}"
  
  type: "ref";
  repo_id: string;
  
  // Ref details
  name: string;                    // Full ref name (refs/heads/main)
  target: string;                  // Commit hash
  ref_type: "branch" | "tag" | "HEAD";
  
  // Timestamps
  updated_at: ISODateTime;
}
```

**Indexes**:
```javascript
[
  {type: "hash", fields: ["repo_id"]},
  {type: "persistent", fields: ["repo_id", "name"], unique: true},
]
```

### agencies

```typescript
interface Agency {
  _key: string;
  _id: string;                     // "agencies/{key}"
  
  // Identity
  name: string;
  code: string;                    // Short code (FINSERV, HEALTHTECH)
  
  // Configuration
  repo_id: string;
  gitea_project_id: number;
  
  // Status
  status: "active" | "inactive";
  
  // Timestamps
  created_at: ISODateTime;
  updated_at: ISODateTime;
}
```

### agents

```typescript
interface Agent {
  _key: string;
  _id: string;                     // "agents/{key}"
  
  // Identity
  name: string;
  email: string;
  type: "llm" | "human";
  
  // LLM configuration
  model?: string;                  // gpt-4, gpt-3.5-turbo
  temperature?: number;
  max_tokens?: number;
  
  // Budget
  budget?: AgentBudget;
  
  // Status
  status: "active" | "inactive";
  
  // Timestamps
  created_at: ISODateTime;
}

interface AgentBudget {
  max_tokens_per_day: number;
  max_cost_per_day: number;
  used_tokens: number;
  used_cost: number;
  last_reset: ISODateTime;
}
```

### workflow_executions

```typescript
interface WorkflowExecution {
  _key: string;
  _id: string;                     // "workflow_executions/{key}"
  
  // Reference
  work_item_key: string;
  
  // Status
  status: "executing" | "completed" | "failed";
  
  // Execution steps
  steps: ExecutionStep[];
  
  // LLM tracking
  llm_calls: LLMCall[];
  
  // Git operations
  git_operations: GitOperation[];
  
  // Metrics
  metrics: ExecutionMetrics;
  
  // Error
  error?: string;
  
  // Timestamps
  started_at: ISODateTime;
  completed_at?: ISODateTime;
}

interface ExecutionStep {
  step: string;                    // classify, generate, commit, merge
  status: "pending" | "executing" | "completed" | "failed";
  started_at: ISODateTime;
  completed_at?: ISODateTime;
  output?: string;
}

interface LLMCall {
  model: string;
  prompt: string;
  response: string;
  tokens: number;
  cost: number;
  duration_ms: number;
}

interface GitOperation {
  type: "branch" | "commit" | "push" | "merge";
  hash?: string;
  branch?: string;
  mr_url?: string;
  timestamp: ISODateTime;
}

interface ExecutionMetrics {
  total_duration_ms: number;
  llm_duration_ms: number;
  git_duration_ms: number;
  total_tokens: number;
  total_cost: number;
}
```

### llm_usage

```typescript
interface LLMUsage {
  _key: string;
  _id: string;                     // "llm_usage/{key}"
  
  // Reference
  agent_id: string;
  work_item_id?: string;
  
  // Request
  model: string;
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
  
  // Cost
  cost: number;                    // USD
  
  // Performance
  duration_ms: number;
  
  // Timestamp
  timestamp: ISODateTime;
}
```

**Indexes**:
```javascript
[
  {type: "hash", fields: ["agent_id"]},
  {type: "hash", fields: ["work_item_id"]},
  {type: "persistent", fields: ["timestamp"]},
]
```

### mutex_locks

```typescript
interface MutexLock {
  _key: string;                    // Resource identifier
  _id: string;                     // "mutex_locks/{key}"
  
  // Lock details
  owner: string;                   // Work item or agent ID
  acquired_at: ISODateTime;
  ttl: number;                     // Seconds
  
  // Auto-expire with TTL index
}
```

**TTL Index**:
```javascript
{
  type: "ttl",
  fields: ["acquired_at"],
  expireAfter: 300  // 5 minutes
}
```

## Edge Collections

### commit_parents

Links commits to their parent commits (git history).

```typescript
interface CommitParentEdge {
  _key: string;
  _id: string;                     // "commit_parents/{key}"
  _from: string;                   // "git_commits/{child_hash}"
  _to: string;                     // "git_commits/{parent_hash}"
  
  // Metadata
  parent_index: number;            // 0 for first parent, 1 for second (merges)
  created_at: ISODateTime;
}
```

**Indexes**:
```javascript
[
  {type: "persistent", fields: ["_from"]},
  {type: "persistent", fields: ["_to"]},
]
```

### code_dependencies

Links files to their dependencies (imports, includes).

```typescript
interface CodeDependencyEdge {
  _key: string;
  _id: string;                     // "code_dependencies/{key}"
  _from: string;                   // "git_objects/{file_hash}" (dependent)
  _to: string;                     // "git_objects/{dependency_hash}"
  
  // Metadata
  type: "import" | "include" | "require";
  language: string;
  import_statement: string;
  
  // Timestamps
  created_at: ISODateTime;
}
```

**Indexes**:
```javascript
[
  {type: "persistent", fields: ["_from"]},
  {type: "persistent", fields: ["_to"]},
  {type: "hash", fields: ["language"]},
]
```

### work_item_commits

Links work items to commits they created.

```typescript
interface WorkItemCommitEdge {
  _key: string;
  _id: string;                     // "work_item_commits/{key}"
  _from: string;                   // "work_items/{work_item_key}"
  _to: string;                     // "git_commits/{commit_hash}"
  
  // Metadata
  commit_index: number;            // Order of commits in work item
  created_at: ISODateTime;
}
```

**Indexes**:
```javascript
[
  {type: "persistent", fields: ["_from"]},
  {type: "persistent", fields: ["_to"]},
]
```

### agent_work_items

Links agents to work items they executed.

```typescript
interface AgentWorkItemEdge {
  _key: string;
  _id: string;                     // "agent_work_items/{key}"
  _from: string;                   // "agents/{agent_key}"
  _to: string;                     // "work_items/{work_item_key}"
  
  // Execution metadata
  started_at: ISODateTime;
  completed_at?: ISODateTime;
  status: "executing" | "completed" | "failed";
}
```

**Indexes**:
```javascript
[
  {type: "persistent", fields: ["_from", "status"]},
]
```

### agent_code_expertise

Links agents to code files they have expertise in (knowledge graph).

```typescript
interface AgentCodeExpertiseEdge {
  _key: string;
  _id: string;                     // "agent_code_expertise/{key}"
  _from: string;                   // "agents/{agent_key}"
  _to: string;                     // "git_objects/{file_hash}"
  
  // Expertise metrics
  expertise: number;               // 0.0-1.0 score
  commits: string[];               // Commit hashes where agent touched this file
  last_touch: ISODateTime;         // Most recent commit
  
  // Computed scores
  recency_score?: number;          // Time-based decay
  contribution_score?: number;     // Weighted by expertise + commits
}
```

**Indexes**:
```javascript
[
  {type: "persistent", fields: ["_from", "expertise"]},
  {type: "persistent", fields: ["_to", "expertise"]},
]
```

## Named Graphs

### commit_graph

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

### code_graph

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

### workflow_graph

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
      "collection": "agent_work_items",
      "from": ["agents"],
      "to": ["work_items"]
    }
  ]
}
```

### knowledge_graph

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

## View Definitions

### git_content_search

Full-text search view for code content.

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

### Analyzer: text_en_code

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

---

**See Also**:
- [Git Storage](./git-storage.md) - Git object storage details
- [Graph Queries](./graph-queries.md) - Query patterns
