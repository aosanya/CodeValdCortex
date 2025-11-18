---
title: Work Items Overview
path: /documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items/
---

# Work Items — Overview

CodeValdCortex implements a **GitOps-first** work item execution model where Gitea/GitLab issues trigger LLM-powered goroutines that perform git operations, with all relationships stored in ArangoDB's multi-graph database.

## Architecture

```
GitLab/Gitea Issues (Work Item Triggers)
    ↓ Webhooks
CodeValdCortex Workflow Engine (Goroutines)
    ↓ Git Operations (branch, commit, push, MR)
ArangoDB Multi-Graph Storage
    ├── Git Objects (commits, trees, blobs)
    ├── Work Item Metadata
    ├── Code Dependency Graph
    ├── Knowledge Graph (who knows what)
    └── Workflow Execution State
```

## Documentation Structure

This specification is split into focused documents:

### Core Concepts
- **[work-item-types.md](./work-item-types.md)** - Work item types and execution patterns
- **[gitops-workflow.md](./gitops-workflow.md)** - GitOps execution model and goroutines
- **[llm-integration.md](./llm-integration.md)** - LLM-powered content generation

### Storage & Data
- **[git-storage.md](./git-storage.md)** - Git objects in ArangoDB
- **[graph-queries.md](./graph-queries.md)** - Multi-dimensional graph queries
- **[data-models.md](./data-models.md)** - Complete data models

### Integration
- **[gitea-integration.md](./gitea-integration.md)** - Gitea webhooks and merge automation
- **[api-reference.md](./api-reference.md)** - API endpoints and schemas

### Operations
- **[observability.md](./observability.md)** - Metrics, traces, and analytics
- **[deployment.md](./deployment.md)** - Deployment architecture
- **[examples.md](./examples.md)** - Real-world use cases

## Quick Start

1. **Issue Creation**: Create issue in Gitea with label `work-item`
2. **Classification**: System classifies work type (document, software, proposal, analysis)
3. **Execution**: Goroutine executes git workflow (branch → commit → push → MR)
4. **Graph Storage**: All relationships stored in ArangoDB graphs
5. **Auto-merge**: Based on trust level and CI status

## Key Features

✅ **GitOps-First**: All work results in git commits  
✅ **Graph-Powered**: Multi-dimensional relationship queries  
✅ **LLM-Automated**: AI generates code and documents  
✅ **Searchable**: Full-text search across all code  
✅ **High Throughput**: Parallel goroutine execution  
✅ **Knowledge Tracking**: Automatic expertise graph  

## Goals and Acceptance Criteria

- ✅ GitOps workflow: Issues → Goroutines → Git operations → Merge requests
- ✅ Work Items stored in ArangoDB with full graph relationships
- ✅ Git objects (commits, trees, blobs) stored in ArangoDB for advanced queries
- ✅ LLM agents generate code/documents based on issue requirements
- ✅ Multi-dimensional graph traversal (code deps, commits, work items, agents)
- ✅ Knowledge graph tracks expertise (who knows what code)
- ✅ Searchable code content with full-text and semantic search
- ✅ Impact analysis via graph queries (what breaks when code changes)
- ✅ Concurrency controls prevent duplicate, conflicting external effects
- ✅ Integration with Gitea for issue tracking and code review UI

---

**Last Updated**: November 7, 2025  
**Architecture**: GitOps + ArangoDB Multi-Graph + LLM Agents
