package orchestration

import (
	"fmt"
	"sort"
)

// DependencyGraph represents a directed acyclic graph (DAG) for task dependencies
type DependencyGraph struct {
	nodes map[string]*GraphNode
	edges map[string][]string // adjacency list: node -> list of dependent nodes
}

// GraphNode represents a node in the dependency graph
type GraphNode struct {
	ID           string
	Dependencies []string // nodes this node depends on
	Dependents   []string // nodes that depend on this node
	InDegree     int      // number of incoming edges (dependencies)
}

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes: make(map[string]*GraphNode),
		edges: make(map[string][]string),
	}
}

// AddNode adds a node to the graph
func (g *DependencyGraph) AddNode(nodeID string) {
	if _, exists := g.nodes[nodeID]; !exists {
		g.nodes[nodeID] = &GraphNode{
			ID:           nodeID,
			Dependencies: make([]string, 0),
			Dependents:   make([]string, 0),
			InDegree:     0,
		}
		g.edges[nodeID] = make([]string, 0)
	}
}

// AddEdge adds a dependency edge from source to target (source must complete before target)
func (g *DependencyGraph) AddEdge(sourceID, targetID string) error {
	// Ensure both nodes exist
	if _, exists := g.nodes[sourceID]; !exists {
		return fmt.Errorf("source node %s does not exist", sourceID)
	}
	if _, exists := g.nodes[targetID]; !exists {
		return fmt.Errorf("target node %s does not exist", targetID)
	}

	// Avoid duplicate edges
	for _, dep := range g.nodes[targetID].Dependencies {
		if dep == sourceID {
			return nil // Edge already exists
		}
	}

	// Add edge
	g.edges[sourceID] = append(g.edges[sourceID], targetID)
	g.nodes[sourceID].Dependents = append(g.nodes[sourceID].Dependents, targetID)
	g.nodes[targetID].Dependencies = append(g.nodes[targetID].Dependencies, sourceID)
	g.nodes[targetID].InDegree++

	return nil
}

// ValidateAcyclic checks if the graph is acyclic (no circular dependencies)
func (g *DependencyGraph) ValidateAcyclic() error {
	// Use DFS to detect cycles
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	for nodeID := range g.nodes {
		if !visited[nodeID] {
			if g.hasCycleDFS(nodeID, visited, recursionStack) {
				return fmt.Errorf("circular dependency detected involving node %s", nodeID)
			}
		}
	}

	return nil
}

// hasCycleDFS performs depth-first search to detect cycles
func (g *DependencyGraph) hasCycleDFS(nodeID string, visited, recursionStack map[string]bool) bool {
	visited[nodeID] = true
	recursionStack[nodeID] = true

	// Visit all dependents
	for _, dependent := range g.edges[nodeID] {
		if !visited[dependent] {
			if g.hasCycleDFS(dependent, visited, recursionStack) {
				return true
			}
		} else if recursionStack[dependent] {
			return true // Back edge found - cycle detected
		}
	}

	recursionStack[nodeID] = false
	return false
}

// GetExecutionBatches returns groups of tasks that can be executed in parallel
func (g *DependencyGraph) GetExecutionBatches() [][]string {
	batches := make([][]string, 0)

	// Copy in-degrees for processing
	inDegrees := make(map[string]int)
	for nodeID, node := range g.nodes {
		inDegrees[nodeID] = node.InDegree
	}

	processed := make(map[string]bool)

	for len(processed) < len(g.nodes) {
		// Find all nodes with in-degree 0 (ready to execute)
		currentBatch := make([]string, 0)

		for nodeID := range g.nodes {
			if !processed[nodeID] && inDegrees[nodeID] == 0 {
				currentBatch = append(currentBatch, nodeID)
			}
		}

		if len(currentBatch) == 0 {
			// This shouldn't happen if graph is acyclic
			break
		}

		// Sort batch for deterministic ordering
		sort.Strings(currentBatch)
		batches = append(batches, currentBatch)

		// Mark nodes as processed and update in-degrees
		for _, nodeID := range currentBatch {
			processed[nodeID] = true

			// Reduce in-degree for all dependents
			for _, dependent := range g.edges[nodeID] {
				inDegrees[dependent]--
			}
		}
	}

	return batches
}

// GetTopologicalOrder returns a topological ordering of nodes
func (g *DependencyGraph) GetTopologicalOrder() ([]string, error) {
	if err := g.ValidateAcyclic(); err != nil {
		return nil, err
	}

	batches := g.GetExecutionBatches()
	order := make([]string, 0, len(g.nodes))

	for _, batch := range batches {
		order = append(order, batch...)
	}

	return order, nil
}

// GetNodeDependencies returns the direct dependencies of a node
func (g *DependencyGraph) GetNodeDependencies(nodeID string) ([]string, error) {
	node, exists := g.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node %s does not exist", nodeID)
	}

	dependencies := make([]string, len(node.Dependencies))
	copy(dependencies, node.Dependencies)
	return dependencies, nil
}

// GetNodeDependents returns the direct dependents of a node
func (g *DependencyGraph) GetNodeDependents(nodeID string) ([]string, error) {
	node, exists := g.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node %s does not exist", nodeID)
	}

	dependents := make([]string, len(node.Dependents))
	copy(dependents, node.Dependents)
	return dependents, nil
}

// GetAllDependencies returns all transitive dependencies of a node
func (g *DependencyGraph) GetAllDependencies(nodeID string) ([]string, error) {
	if _, exists := g.nodes[nodeID]; !exists {
		return nil, fmt.Errorf("node %s does not exist", nodeID)
	}

	visited := make(map[string]bool)
	dependencies := make([]string, 0)

	g.collectDependenciesDFS(nodeID, visited, &dependencies)

	// Remove the node itself from dependencies
	filteredDeps := make([]string, 0, len(dependencies))
	for _, dep := range dependencies {
		if dep != nodeID {
			filteredDeps = append(filteredDeps, dep)
		}
	}

	sort.Strings(filteredDeps)
	return filteredDeps, nil
}

// collectDependenciesDFS recursively collects all dependencies
func (g *DependencyGraph) collectDependenciesDFS(nodeID string, visited map[string]bool, dependencies *[]string) {
	if visited[nodeID] {
		return
	}

	visited[nodeID] = true
	*dependencies = append(*dependencies, nodeID)

	// Visit all dependencies
	node := g.nodes[nodeID]
	for _, dep := range node.Dependencies {
		g.collectDependenciesDFS(dep, visited, dependencies)
	}
}

// IsReady checks if a node is ready for execution (all dependencies completed)
func (g *DependencyGraph) IsReady(nodeID string, completedNodes map[string]bool) bool {
	node, exists := g.nodes[nodeID]
	if !exists {
		return false
	}

	// Check if all dependencies are completed
	for _, dep := range node.Dependencies {
		if !completedNodes[dep] {
			return false
		}
	}

	return true
}

// GetReadyNodes returns all nodes that are ready for execution
func (g *DependencyGraph) GetReadyNodes(completedNodes map[string]bool) []string {
	readyNodes := make([]string, 0)

	for nodeID := range g.nodes {
		if !completedNodes[nodeID] && g.IsReady(nodeID, completedNodes) {
			readyNodes = append(readyNodes, nodeID)
		}
	}

	sort.Strings(readyNodes)
	return readyNodes
}

// GetGraphInfo returns information about the graph structure
func (g *DependencyGraph) GetGraphInfo() map[string]interface{} {
	return map[string]interface{}{
		"total_nodes": len(g.nodes),
		"total_edges": g.countEdges(),
		"max_depth":   g.calculateMaxDepth(),
		"parallelism": g.calculateMaxParallelism(),
		"is_acyclic":  g.ValidateAcyclic() == nil,
	}
}

// countEdges counts the total number of edges in the graph
func (g *DependencyGraph) countEdges() int {
	count := 0
	for _, dependents := range g.edges {
		count += len(dependents)
	}
	return count
}

// calculateMaxDepth calculates the maximum depth of the dependency graph
func (g *DependencyGraph) calculateMaxDepth() int {
	batches := g.GetExecutionBatches()
	return len(batches)
}

// calculateMaxParallelism calculates the maximum number of tasks that can run in parallel
func (g *DependencyGraph) calculateMaxParallelism() int {
	batches := g.GetExecutionBatches()
	maxParallelism := 0

	for _, batch := range batches {
		if len(batch) > maxParallelism {
			maxParallelism = len(batch)
		}
	}

	return maxParallelism
}

// Clone creates a deep copy of the dependency graph
func (g *DependencyGraph) Clone() *DependencyGraph {
	clone := NewDependencyGraph()

	// Copy nodes
	for nodeID := range g.nodes {
		clone.AddNode(nodeID)
	}

	// Copy edges
	for sourceID, dependents := range g.edges {
		for _, targetID := range dependents {
			clone.AddEdge(sourceID, targetID)
		}
	}

	return clone
}

// String returns a string representation of the graph
func (g *DependencyGraph) String() string {
	result := "DependencyGraph {\n"

	for nodeID, node := range g.nodes {
		result += fmt.Sprintf("  %s: deps=[%v], dependents=[%v], in_degree=%d\n",
			nodeID, node.Dependencies, node.Dependents, node.InDegree)
	}

	result += "}"
	return result
}
