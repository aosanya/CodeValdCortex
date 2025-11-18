# MVP-052: Workflow Visual Designer - Coding Session

**Date**: November 7, 2025  
**Task**: Integrate workflow visual designer with GitOps work items architecture  
**Branch**: `feature/MVP-052_workflow_visual_designer`  
**Status**: Architecture Review - Realigning with Existing Implementation

## Discovery (November 7, 2025)

### Existing Implementation Found

**The visual workflow designer is already implemented** with jsPlumb (not xyflow as MVP spec suggested):

**Existing Files**:
- ✅ `/internal/web/pages/agency_designer/workflow_designer.templ` (457 lines) - jsPlumb-based template
- ✅ `/static/js/agency-designer/workflow-designer.js` (609 lines) - Complete Alpine.js + jsPlumb implementation
- ✅ `/internal/web/handlers/agency_designer_handler.go` - Handler with `ShowWorkflowDesigner` method
- ✅ Route: `GET /agencies/:id/designer/workflows/:workflowId` - Already registered

### Discrepancy Analysis

**MVP-052 Specification** said: "Build drag-and-drop workflow designer using **xyflow** (vanilla JS)"

**Actual Codebase** uses: **jsPlumb Community Edition** with Alpine.js

**Root Cause**: MVP spec was written before implementation, or spec wasn't updated when jsPlumb was chosen.

### Files Created (Now Redundant - Using xyflow)

These files were created based on the MVP spec but don't match the existing architecture:

1. ❌ `/workspaces/CodeValdCortex/static/js/workflow-designer.js` (485 lines) - xyflow implementation
2. ❌ `/workspaces/CodeValdCortex/static/css/workflow-designer.css` (210+ lines) - xyflow styling
3. ❌ `/workspaces/CodeValdCortex/internal/web/templates/pages/workflows/designer.templ` - Duplicate template
4. ⚠️ `/workspaces/CodeValdCortex/internal/models/workflow.go` (122 lines) - May need alignment
5. ⚠️ `/workspaces/CodeValdCortex/internal/web/handlers/workflows/handler.go` (362 lines) - Duplicate handlers

### Correct Architecture (jsPlumb-based)

The existing workflow designer already supports:
- ✅ Drag-and-drop node creation
- ✅ Visual connections between nodes  
- ✅ Node types: start, end, decision, parallel, work-item
- ✅ Properties panel for node configuration
- ✅ Pan/zoom canvas navigation
- ✅ Save/load workflows
- ✅ Undo/redo support
- ✅ Keyboard shortcuts

## Next Steps - GitOps Integration

**Objective**: Integrate existing jsPlumb workflow designer with GitOps work items architecture

### Phase 1: Model Alignment (CURRENT TASK)

**Goal**: Ensure workflow models support GitOps requirements from [work-items documentation](../../2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items/)

**Tasks**:
1. ✅ Review existing `internal/workflow/` models
2. ✅ Add GitOps-specific fields to NodeData:
   - Added `GitConfig` struct (flexible for any Git backend: Gitea, GitLab, GitHub)
   - Fields: backend, repo, branch_pattern, merge_strategy, auto_merge, require_reviews, require_ci
   - Added `Labels` field for issue/MR labels (works with any Git platform)
   - Updated `WorkItemType` comment to document types: document, software, proposal, analysis
3. 🔄 Update workflow service to support Git issue creation
4. 🔄 Add validation for GitOps workflow constraints

**Changes Made**:
- File: `/internal/workflow/models.go`
  - Added `GitConfig` struct with `backend` field for flexibility
  - Added `git_config` and `labels` fields to `NodeData`
  - Documented work item types in comments
  - **Design Decision**: Used `GitConfig` instead of `GiteaConfig` to support multiple Git backends (Gitea, GitLab, GitHub, etc.)

### Phase 2: Gitea Integration (NEXT)

**Goal**: Connect workflow execution to Gitea issue creation

**Tasks**:
1. 🔄 Create Gitea client service
2. 🔄 Implement workflow-to-issue converter
3. 🔄 Add webhook handler for issue lifecycle events
4. 🔄 Update workflow service with `ExecuteWorkflow()` method that:
   - Creates Gitea issues for each work item node
   - Sets up dependencies between issues
   - Applies labels based on work item type
   - Configures auto-merge settings

### Phase 3: UI Enhancement (FUTURE)

**Goal**: Update jsPlumb designer UI to support GitOps fields

**Tasks**:
1. 🔄 Add GitOps configuration panel to properties sidebar
2. 🔄 Update work item node creation to include Gitea config
3. 🔄 Add visual indicators for auto-merge vs manual review
4. 🔄 Implement work item type selector (document/software/proposal/analysis)
5. 🔄 Add label editor UI component

## Objective

Build a visual workflow designer using xyflow (vanilla JS) that enables users to:
1. Visually design work item workflows that trigger Gitea issues
2. Define dependencies, conditional routing, and parallel execution paths
3. Integrate with the GitOps + ArangoDB architecture
4. Save/load workflows as JSON with execution visualization

## Architecture Context

**Reference**: [Work Items Documentation](../../2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items/)

### GitOps Workflow Integration

The visual designer creates workflows that:
1. **Generate Gitea Issues**: Each workflow node represents a work item that creates a Gitea issue
2. **Trigger Goroutines**: Gitea webhooks spawn goroutines for parallel execution
3. **Git Operations**: Each work item performs git operations (branch, commit, push, merge)
4. **Graph Storage**: Relationships stored in ArangoDB's workflow_graph
5. **LLM Execution**: Nodes trigger LLM agents to generate content based on work type

### Work Item Types (from architecture)

```typescript
// From work-items/work-item-types.md
type WorkItemType = "document" | "software" | "proposal" | "analysis";

interface WorkItemNode {
  id: string;
  type: WorkItemType;
  title: string;
  description: string;
  labels: string[];
  dependencies: string[];  // IDs of prerequisite nodes
  gitea_config: {
    repo: string;
    branch_pattern: string;
    merge_strategy: "squash" | "merge" | "rebase";
    auto_merge: boolean;
    require_reviews: number;
  };
}
```

## Implementation Plan

### Phase 1: Visual Designer Foundation (Current)

**Goal**: Basic xyflow canvas with work item nodes

#### 1.1 HTML/Templ Structure

```html
<!-- internal/web/templates/pages/workflows/designer.templ -->
package workflows

templ WorkflowDesigner(workflow *models.Workflow) {
  <div class="workflow-designer" x-data="workflowDesigner()">
    <!-- Toolbar -->
    <div class="toolbar">
      <div class="node-palette">
        <button @click="addNode('document')" class="button is-small">
          <span class="icon"><i class="fas fa-file-alt"></i></span>
          <span>Document</span>
        </button>
        <button @click="addNode('software')" class="button is-small">
          <span class="icon"><i class="fas fa-code"></i></span>
          <span>Software</span>
        </button>
        <button @click="addNode('proposal')" class="button is-small">
          <span class="icon"><i class="fas fa-file-contract"></i></span>
          <span>Proposal</span>
        </button>
        <button @click="addNode('analysis')" class="button is-small">
          <span class="icon"><i class="fas fa-chart-line"></i></span>
          <span>Analysis</span>
        </button>
      </div>
      
      <div class="actions">
        <button @click="saveWorkflow()" class="button is-primary is-small">
          <span class="icon"><i class="fas fa-save"></i></span>
          <span>Save</span>
        </button>
        <button @click="validateWorkflow()" class="button is-info is-small">
          <span class="icon"><i class="fas fa-check-circle"></i></span>
          <span>Validate</span>
        </button>
        <button @click="exportWorkflow()" class="button is-light is-small">
          <span class="icon"><i class="fas fa-download"></i></span>
          <span>Export</span>
        </button>
      </div>
    </div>

    <!-- Canvas -->
    <div id="workflow-canvas" class="workflow-canvas"></div>

    <!-- Properties Panel -->
    <div class="properties-panel" x-show="selectedNode">
      <h3 class="title is-5">Node Properties</h3>
      
      <div class="field">
        <label class="label">Work Item Type</label>
        <div class="control">
          <span class="tag is-info" x-text="selectedNode?.data?.type"></span>
        </div>
      </div>

      <div class="field">
        <label class="label">Title</label>
        <div class="control">
          <input 
            class="input" 
            type="text" 
            x-model="selectedNode.data.title"
            @input="updateNode()"
          />
        </div>
      </div>

      <div class="field">
        <label class="label">Description</label>
        <div class="control">
          <textarea 
            class="textarea" 
            x-model="selectedNode.data.description"
            @input="updateNode()"
            rows="4"
          ></textarea>
        </div>
      </div>

      <div class="field">
        <label class="label">Labels</label>
        <div class="control">
          <input 
            class="input" 
            type="text" 
            x-model="selectedNode.data.labelsText"
            @input="updateNode()"
            placeholder="work-item, documentation, P1"
          />
          <p class="help">Comma-separated labels for Gitea issue</p>
        </div>
      </div>

      <!-- GitOps Configuration -->
      <div class="box">
        <h4 class="title is-6">GitOps Settings</h4>
        
        <div class="field">
          <label class="label">Repository</label>
          <div class="control">
            <input 
              class="input" 
              type="text" 
              x-model="selectedNode.data.gitea_config.repo"
              @input="updateNode()"
            />
          </div>
        </div>

        <div class="field">
          <label class="label">Merge Strategy</label>
          <div class="control">
            <div class="select">
              <select x-model="selectedNode.data.gitea_config.merge_strategy" @change="updateNode()">
                <option value="squash">Squash</option>
                <option value="merge">Merge</option>
                <option value="rebase">Rebase</option>
              </select>
            </div>
          </div>
        </div>

        <div class="field">
          <label class="checkbox">
            <input 
              type="checkbox" 
              x-model="selectedNode.data.gitea_config.auto_merge"
              @change="updateNode()"
            />
            Auto-merge on CI success
          </label>
        </div>

        <div class="field">
          <label class="label">Required Reviews</label>
          <div class="control">
            <input 
              class="input" 
              type="number" 
              min="0" 
              max="5"
              x-model.number="selectedNode.data.gitea_config.require_reviews"
              @input="updateNode()"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
}
```

#### 1.2 Alpine.js Component

```javascript
// static/js/workflow-designer.js

function workflowDesigner() {
  return {
    flowInstance: null,
    selectedNode: null,
    nodes: [],
    edges: [],
    workflowId: null,

    init() {
      this.initializeFlow();
      this.loadWorkflow();
    },

    initializeFlow() {
      const container = document.getElementById('workflow-canvas');
      
      // Initialize xyflow
      this.flowInstance = new XYFlow.default(container, {
        nodes: this.nodes,
        edges: this.edges,
        nodeTypes: {
          workItem: this.createWorkItemNode()
        },
        onNodeClick: (event, node) => {
          this.selectedNode = node;
        },
        onConnect: (params) => {
          this.addEdge(params);
        }
      });
    },

    createWorkItemNode() {
      return {
        render: (node) => {
          const typeIcons = {
            document: 'fa-file-alt',
            software: 'fa-code',
            proposal: 'fa-file-contract',
            analysis: 'fa-chart-line'
          };

          const typeColors = {
            document: 'is-info',
            software: 'is-success',
            proposal: 'is-warning',
            analysis: 'is-danger'
          };

          return `
            <div class="work-item-node box ${typeColors[node.data.type]}">
              <div class="node-header">
                <span class="icon">
                  <i class="fas ${typeIcons[node.data.type]}"></i>
                </span>
                <strong>${node.data.title || 'Untitled'}</strong>
              </div>
              <div class="node-body">
                <p class="is-size-7">${node.data.description?.substring(0, 50) || ''}...</p>
                <div class="tags">
                  ${node.data.labels?.map(l => `<span class="tag is-small">${l}</span>`).join('') || ''}
                </div>
              </div>
              <div class="node-footer">
                <span class="icon is-small" title="${node.data.gitea_config?.auto_merge ? 'Auto-merge enabled' : 'Manual merge'}">
                  <i class="fas ${node.data.gitea_config?.auto_merge ? 'fa-bolt' : 'fa-hand-paper'}"></i>
                </span>
              </div>
            </div>
          `;
        }
      };
    },

    addNode(type) {
      const nodeId = `node-${Date.now()}`;
      const newNode = {
        id: nodeId,
        type: 'workItem',
        position: { 
          x: Math.random() * 400 + 100, 
          y: Math.random() * 300 + 100 
        },
        data: {
          type: type,
          title: `New ${type} work item`,
          description: '',
          labels: ['work-item', type],
          labelsText: `work-item, ${type}`,
          gitea_config: {
            repo: 'main-repo',
            branch_pattern: `issue-{issue_id}-{slug}`,
            merge_strategy: type === 'document' ? 'squash' : 'merge',
            auto_merge: type === 'document',
            require_reviews: type === 'document' ? 0 : 1
          }
        }
      };

      this.nodes.push(newNode);
      this.flowInstance.setNodes([...this.nodes]);
      this.selectedNode = newNode;
    },

    addEdge(params) {
      const newEdge = {
        id: `edge-${params.source}-${params.target}`,
        source: params.source,
        target: params.target,
        type: 'smoothstep',
        animated: true,
        label: 'depends on'
      };

      this.edges.push(newEdge);
      this.flowInstance.setEdges([...this.edges]);
    },

    updateNode() {
      if (this.selectedNode) {
        // Parse labels text
        this.selectedNode.data.labels = this.selectedNode.data.labelsText
          .split(',')
          .map(l => l.trim())
          .filter(l => l);

        // Update flow
        this.flowInstance.setNodes([...this.nodes]);
      }
    },

    async saveWorkflow() {
      const workflow = {
        id: this.workflowId,
        name: 'Agency Workflow',
        nodes: this.nodes.map(n => ({
          id: n.id,
          type: n.data.type,
          title: n.data.title,
          description: n.data.description,
          labels: n.data.labels,
          position: n.position,
          gitea_config: n.data.gitea_config
        })),
        edges: this.edges.map(e => ({
          id: e.id,
          source: e.source,
          target: e.target
        }))
      };

      try {
        const response = await fetch('/api/v1/workflows', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(workflow)
        });

        if (response.ok) {
          const data = await response.json();
          this.workflowId = data.id;
          this.showNotification('Workflow saved successfully', 'success');
        } else {
          throw new Error('Save failed');
        }
      } catch (error) {
        this.showNotification('Failed to save workflow', 'danger');
      }
    },

    async validateWorkflow() {
      // Check for cycles
      const hasCycle = this.detectCycle();
      if (hasCycle) {
        this.showNotification('Workflow has circular dependencies', 'warning');
        return false;
      }

      // Check for orphaned nodes
      const orphans = this.findOrphanedNodes();
      if (orphans.length > 0) {
        this.showNotification(`Found ${orphans.length} orphaned nodes`, 'warning');
      }

      // Check required fields
      const invalidNodes = this.nodes.filter(n => 
        !n.data.title || !n.data.description || !n.data.gitea_config.repo
      );

      if (invalidNodes.length > 0) {
        this.showNotification('Some nodes are missing required fields', 'danger');
        return false;
      }

      this.showNotification('Workflow is valid!', 'success');
      return true;
    },

    detectCycle() {
      // Simple DFS cycle detection
      const visited = new Set();
      const recStack = new Set();

      const hasCycleUtil = (nodeId) => {
        visited.add(nodeId);
        recStack.add(nodeId);

        const outgoingEdges = this.edges.filter(e => e.source === nodeId);
        for (const edge of outgoingEdges) {
          if (!visited.has(edge.target)) {
            if (hasCycleUtil(edge.target)) return true;
          } else if (recStack.has(edge.target)) {
            return true;
          }
        }

        recStack.delete(nodeId);
        return false;
      };

      for (const node of this.nodes) {
        if (!visited.has(node.id)) {
          if (hasCycleUtil(node.id)) return true;
        }
      }

      return false;
    },

    findOrphanedNodes() {
      const connectedNodes = new Set();
      this.edges.forEach(e => {
        connectedNodes.add(e.source);
        connectedNodes.add(e.target);
      });

      return this.nodes.filter(n => !connectedNodes.has(n.id) && this.nodes.length > 1);
    },

    async exportWorkflow() {
      const workflow = {
        version: '1.0',
        nodes: this.nodes,
        edges: this.edges,
        metadata: {
          created: new Date().toISOString(),
          generator: 'CodeValdCortex Workflow Designer'
        }
      };

      const blob = new Blob([JSON.stringify(workflow, null, 2)], { 
        type: 'application/json' 
      });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'workflow.json';
      a.click();
    },

    async loadWorkflow() {
      const urlParams = new URLSearchParams(window.location.search);
      const workflowId = urlParams.get('id');

      if (workflowId) {
        try {
          const response = await fetch(`/api/v1/workflows/${workflowId}`);
          const workflow = await response.json();

          this.workflowId = workflow.id;
          this.nodes = workflow.nodes.map(n => ({
            ...n,
            type: 'workItem',
            data: {
              ...n,
              labelsText: n.labels?.join(', ') || ''
            }
          }));
          this.edges = workflow.edges;

          this.flowInstance.setNodes(this.nodes);
          this.flowInstance.setEdges(this.edges);
        } catch (error) {
          console.error('Failed to load workflow:', error);
        }
      }
    },

    showNotification(message, type) {
      // Use Bulma notification
      const notification = document.createElement('div');
      notification.className = `notification is-${type}`;
      notification.innerHTML = `
        <button class="delete"></button>
        ${message}
      `;
      document.body.appendChild(notification);

      setTimeout(() => notification.remove(), 3000);
    }
  };
}
```

#### 1.3 Backend - Workflow Storage

```go
// internal/web/handlers/workflow_handler.go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/aosanya/CodeValdCortex/internal/database"
)

type WorkflowHandler struct {
	db *database.Database
}

type Workflow struct {
	Key       string         `json:"_key,omitempty"`
	ID        string         `json:"_id,omitempty"`
	AgencyID  string         `json:"agency_id"`
	Name      string         `json:"name"`
	Nodes     []WorkflowNode `json:"nodes"`
	Edges     []WorkflowEdge `json:"edges"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type WorkflowNode struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // document, software, proposal, analysis
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Labels      []string               `json:"labels"`
	Position    Position               `json:"position"`
	GiteaConfig GiteaConfig            `json:"gitea_config"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type GiteaConfig struct {
	Repo           string `json:"repo"`
	BranchPattern  string `json:"branch_pattern"`
	MergeStrategy  string `json:"merge_strategy"`  // squash, merge, rebase
	AutoMerge      bool   `json:"auto_merge"`
	RequireReviews int    `json:"require_reviews"`
}

type WorkflowEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

func (h *WorkflowHandler) SaveWorkflow(c *gin.Context) {
	var workflow Workflow
	if err := c.ShouldBindJSON(&workflow); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get agency ID from context
	agencyID := c.GetString("agency_id")
	workflow.AgencyID = agencyID
	workflow.UpdatedAt = time.Now()

	if workflow.Key == "" {
		// Create new
		workflow.CreatedAt = time.Now()
		meta, err := h.db.Collection("workflows").CreateDocument(context.Background(), workflow)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		workflow.Key = meta.Key
	} else {
		// Update existing
		_, err := h.db.Collection("workflows").UpdateDocument(
			context.Background(),
			workflow.Key,
			workflow,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, workflow)
}

func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	workflowID := c.Param("id")

	var workflow Workflow
	_, err := h.db.Collection("workflows").ReadDocument(
		context.Background(),
		workflowID,
		&workflow,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	c.JSON(http.StatusOK, workflow)
}

func (h *WorkflowHandler) ExecuteWorkflow(c *gin.Context) {
	workflowID := c.Param("id")

	var workflow Workflow
	_, err := h.db.Collection("workflows").ReadDocument(
		context.Background(),
		workflowID,
		&workflow,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	// Execute workflow by creating Gitea issues for each node
	go h.executeWorkflowAsync(workflow)

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Workflow execution started",
		"workflow_id": workflowID,
	})
}

func (h *WorkflowHandler) executeWorkflowAsync(workflow Workflow) {
	// Build dependency graph
	dependencies := make(map[string][]string)
	for _, edge := range workflow.Edges {
		dependencies[edge.Target] = append(dependencies[edge.Target], edge.Source)
	}

	// Topological sort to get execution order
	executed := make(map[string]bool)
	
	var executeNode func(nodeID string) error
	executeNode = func(nodeID string) error {
		if executed[nodeID] {
			return nil
		}

		// Execute dependencies first
		for _, depID := range dependencies[nodeID] {
			if err := executeNode(depID); err != nil {
				return err
			}
		}

		// Find node
		var node *WorkflowNode
		for _, n := range workflow.Nodes {
			if n.ID == nodeID {
				node = &n
				break
			}
		}

		if node == nil {
			return nil
		}

		// Create Gitea issue
		if err := h.createGiteaIssue(node); err != nil {
			return err
		}

		executed[nodeID] = true
		return nil
	}

	// Execute all nodes
	for _, node := range workflow.Nodes {
		if err := executeNode(node.ID); err != nil {
			// Log error
			continue
		}
	}
}

func (h *WorkflowHandler) createGiteaIssue(node *WorkflowNode) error {
	// Create Gitea issue that will trigger work item execution
	issueBody := node.Description + "\n\n" +
		"Repository: " + node.GiteaConfig.Repo + "\n" +
		"Branch Pattern: " + node.GiteaConfig.BranchPattern + "\n" +
		"Merge Strategy: " + node.GiteaConfig.MergeStrategy + "\n" +
		"Auto-merge: " + fmt.Sprintf("%v", node.GiteaConfig.AutoMerge)

	// TODO: Call Gitea API to create issue
	// This will trigger the webhook → goroutine → git operations flow

	return nil
}
```

### Phase 2: GitOps Integration (Next)

- [ ] Gitea API integration for issue creation
- [ ] Webhook setup for workflow execution
- [ ] Real-time execution visualization
- [ ] ArangoDB workflow_graph integration

### Phase 3: Advanced Features (Future)

- [ ] Conditional routing (if/else branches)
- [ ] Parallel execution paths
- [ ] Retry and error handling configuration
- [ ] Workflow templates library

## Testing Plan

1. **Unit Tests**: Node validation, cycle detection, dependency resolution
2. **Integration Tests**: Gitea issue creation, workflow execution
3. **E2E Tests**: Complete workflow from design to merge

## Current Status

✅ Documentation split completed  
🔄 Visual designer foundation (Phase 1)  
⏳ GitOps integration (Phase 2)  
⏳ Advanced features (Phase 3)

## Next Steps

1. Complete Phase 1: Basic visual designer
2. Test node creation and connections
3. Implement save/load functionality
4. Begin Gitea API integration

## References

- [Work Items Documentation](../../2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items/)
- [xyflow Documentation](https://xyflow.com/)
- [Gitea API](https://docs.gitea.io/en-us/api-usage/)
