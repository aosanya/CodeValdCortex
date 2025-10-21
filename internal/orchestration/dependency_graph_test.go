package orchestration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDependencyGraph_BasicOperations(t *testing.T) {
	graph := NewDependencyGraph()

	// Test adding nodes
	graph.AddNode("task1")
	graph.AddNode("task2")
	graph.AddNode("task3")

	// Test checking node existence (via nodes map)
	assert.Contains(t, graph.nodes, "task1")
	assert.Contains(t, graph.nodes, "task2")
	assert.Contains(t, graph.nodes, "task3")
	assert.NotContains(t, graph.nodes, "task4")

	// Test adding edges
	err := graph.AddEdge("task1", "task2") // task2 depends on task1
	require.NoError(t, err)

	err = graph.AddEdge("task2", "task3") // task3 depends on task2
	require.NoError(t, err)

	// Test dependencies
	deps, err := graph.GetNodeDependencies("task2")
	require.NoError(t, err)
	assert.Contains(t, deps, "task1")
	assert.Len(t, deps, 1)

	deps, err = graph.GetNodeDependencies("task3")
	require.NoError(t, err)
	assert.Contains(t, deps, "task2")
	assert.Len(t, deps, 1)

	// Test dependents
	dependents, err := graph.GetNodeDependents("task1")
	require.NoError(t, err)
	assert.Contains(t, dependents, "task2")
	assert.Len(t, dependents, 1)

	dependents, err = graph.GetNodeDependents("task2")
	require.NoError(t, err)
	assert.Contains(t, dependents, "task3")
	assert.Len(t, dependents, 1)
}

func TestDependencyGraph_CycleDetection(t *testing.T) {
	graph := NewDependencyGraph()

	// Add nodes
	graph.AddNode("task1")
	graph.AddNode("task2")
	graph.AddNode("task3")

	// Add valid edges
	require.NoError(t, graph.AddEdge("task1", "task2"))
	require.NoError(t, graph.AddEdge("task2", "task3"))

	// Validate acyclic graph
	err := graph.ValidateAcyclic()
	assert.NoError(t, err)

	// Add cycle
	err = graph.AddEdge("task3", "task1") // Creates cycle: task1 -> task2 -> task3 -> task1
	require.NoError(t, err)

	// Validate should detect cycle
	err = graph.ValidateAcyclic()
	assert.Error(t, err)
}

func TestDependencyGraph_TopologicalSort(t *testing.T) {
	graph := NewDependencyGraph()

	// Create a simple DAG
	// task1 -> task2 -> task4
	//       -> task3 -> task4
	graph.AddNode("task1")
	graph.AddNode("task2")
	graph.AddNode("task3")
	graph.AddNode("task4")

	require.NoError(t, graph.AddEdge("task1", "task2"))
	require.NoError(t, graph.AddEdge("task1", "task3"))
	require.NoError(t, graph.AddEdge("task2", "task4"))
	require.NoError(t, graph.AddEdge("task3", "task4"))

	// Get topological order
	order, err := graph.GetTopologicalOrder()
	require.NoError(t, err)
	require.Len(t, order, 4)

	// Verify task1 comes before task2 and task3
	task1Pos := findPosition(order, "task1")
	task2Pos := findPosition(order, "task2")
	task3Pos := findPosition(order, "task3")
	task4Pos := findPosition(order, "task4")

	assert.True(t, task1Pos < task2Pos)
	assert.True(t, task1Pos < task3Pos)
	assert.True(t, task2Pos < task4Pos)
	assert.True(t, task3Pos < task4Pos)
}

func TestDependencyGraph_ExecutionBatches(t *testing.T) {
	graph := NewDependencyGraph()

	// Create a DAG with clear levels
	// Level 0: task1
	// Level 1: task2, task3 (both depend on task1)
	// Level 2: task4 (depends on task2 and task3)
	graph.AddNode("task1")
	graph.AddNode("task2")
	graph.AddNode("task3")
	graph.AddNode("task4")

	require.NoError(t, graph.AddEdge("task1", "task2"))
	require.NoError(t, graph.AddEdge("task1", "task3"))
	require.NoError(t, graph.AddEdge("task2", "task4"))
	require.NoError(t, graph.AddEdge("task3", "task4"))

	// Get execution batches
	batches := graph.GetExecutionBatches()
	require.Len(t, batches, 3)

	// Level 0: task1
	assert.Contains(t, batches[0], "task1")
	assert.Len(t, batches[0], 1)

	// Level 1: task2, task3
	assert.Contains(t, batches[1], "task2")
	assert.Contains(t, batches[1], "task3")
	assert.Len(t, batches[1], 2)

	// Level 2: task4
	assert.Contains(t, batches[2], "task4")
	assert.Len(t, batches[2], 1)
}

func TestDependencyGraph_EmptyGraph(t *testing.T) {
	graph := NewDependencyGraph()

	err := graph.ValidateAcyclic()
	assert.NoError(t, err)

	order, err := graph.GetTopologicalOrder()
	assert.NoError(t, err)
	assert.Empty(t, order)

	assert.Empty(t, graph.GetExecutionBatches())
}

func TestDependencyGraph_SingleNode(t *testing.T) {
	graph := NewDependencyGraph()
	graph.AddNode("task1")

	err := graph.ValidateAcyclic()
	assert.NoError(t, err)

	order, err := graph.GetTopologicalOrder()
	require.NoError(t, err)
	assert.Len(t, order, 1)
	assert.Equal(t, "task1", order[0])

	batches := graph.GetExecutionBatches()
	assert.Len(t, batches, 1)
	assert.Contains(t, batches[0], "task1")
}

func TestDependencyGraph_ComplexDAG(t *testing.T) {
	graph := NewDependencyGraph()

	// Create a more complex DAG
	//     task1
	//    /  |  \
	// task2 task3 task4
	//   |    |    |
	// task5 task6 task7
	//   \    |    /
	//     task8

	nodes := []string{"task1", "task2", "task3", "task4", "task5", "task6", "task7", "task8"}
	for _, node := range nodes {
		graph.AddNode(node)
	}

	// Level 1 dependencies
	require.NoError(t, graph.AddEdge("task1", "task2"))
	require.NoError(t, graph.AddEdge("task1", "task3"))
	require.NoError(t, graph.AddEdge("task1", "task4"))

	// Level 2 dependencies
	require.NoError(t, graph.AddEdge("task2", "task5"))
	require.NoError(t, graph.AddEdge("task3", "task6"))
	require.NoError(t, graph.AddEdge("task4", "task7"))

	// Level 3 dependencies
	require.NoError(t, graph.AddEdge("task5", "task8"))
	require.NoError(t, graph.AddEdge("task6", "task8"))
	require.NoError(t, graph.AddEdge("task7", "task8"))

	// Validate
	err := graph.ValidateAcyclic()
	assert.NoError(t, err)

	// Check execution batches
	batches := graph.GetExecutionBatches()
	require.Len(t, batches, 4)

	// Level 0: task1
	assert.Len(t, batches[0], 1)
	assert.Contains(t, batches[0], "task1")

	// Level 1: task2, task3, task4
	assert.Len(t, batches[1], 3)
	assert.Contains(t, batches[1], "task2")
	assert.Contains(t, batches[1], "task3")
	assert.Contains(t, batches[1], "task4")

	// Level 2: task5, task6, task7
	assert.Len(t, batches[2], 3)
	assert.Contains(t, batches[2], "task5")
	assert.Contains(t, batches[2], "task6")
	assert.Contains(t, batches[2], "task7")

	// Level 3: task8
	assert.Len(t, batches[3], 1)
	assert.Contains(t, batches[3], "task8")
}

func TestDependencyGraph_SelfDependency(t *testing.T) {
	graph := NewDependencyGraph()
	graph.AddNode("task1")

	// Try to add self-dependency
	err := graph.AddEdge("task1", "task1")
	require.NoError(t, err) // AddEdge doesn't prevent this

	// But validation should catch it
	err = graph.ValidateAcyclic()
	assert.Error(t, err)
}

func TestDependencyGraph_NonExistentNodes(t *testing.T) {
	graph := NewDependencyGraph()
	graph.AddNode("task1")

	// Try to add edge with non-existent target node
	err := graph.AddEdge("task1", "nonexistent")
	assert.Error(t, err) // Should fail because target doesn't exist

	// Add the target node and try again
	graph.AddNode("nonexistent")
	err = graph.AddEdge("task1", "nonexistent")
	assert.NoError(t, err)

	// Verify the edge was created
	dependents, err := graph.GetNodeDependents("task1")
	require.NoError(t, err)
	assert.Contains(t, dependents, "nonexistent")
}

func TestDependencyGraph_ReadyNodes(t *testing.T) {
	graph := NewDependencyGraph()

	// Create a simple chain: task1 -> task2 -> task3
	graph.AddNode("task1")
	graph.AddNode("task2")
	graph.AddNode("task3")

	require.NoError(t, graph.AddEdge("task1", "task2"))
	require.NoError(t, graph.AddEdge("task2", "task3"))

	// Initially, only task1 should be ready
	completed := make(map[string]bool)
	ready := graph.GetReadyNodes(completed)
	assert.Len(t, ready, 1)
	assert.Contains(t, ready, "task1")

	// After completing task1, task2 should be ready
	completed["task1"] = true
	ready = graph.GetReadyNodes(completed)
	assert.Len(t, ready, 1)
	assert.Contains(t, ready, "task2")

	// After completing task2, task3 should be ready
	completed["task2"] = true
	ready = graph.GetReadyNodes(completed)
	assert.Len(t, ready, 1)
	assert.Contains(t, ready, "task3")

	// After completing task3, no tasks should be ready
	completed["task3"] = true
	ready = graph.GetReadyNodes(completed)
	assert.Empty(t, ready)
}

// Helper function to find position of element in slice
func findPosition(slice []string, element string) int {
	for i, v := range slice {
		if v == element {
			return i
		}
	}
	return -1
}

// Benchmark tests
func BenchmarkDependencyGraph_AddNode(b *testing.B) {
	graph := NewDependencyGraph()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		graph.AddNode(string(rune(i)))
	}
}

func BenchmarkDependencyGraph_AddEdge(b *testing.B) {
	graph := NewDependencyGraph()

	// Pre-populate with nodes
	for i := 0; i < 1000; i++ {
		graph.AddNode(string(rune(i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		from := string(rune(i % 1000))
		to := string(rune((i + 1) % 1000))
		graph.AddEdge(from, to)
	}
}

func BenchmarkDependencyGraph_TopologicalSort(b *testing.B) {
	graph := NewDependencyGraph()

	// Create a larger graph
	for i := 0; i < 100; i++ {
		graph.AddNode(string(rune(i)))
		if i > 0 {
			graph.AddEdge(string(rune(i-1)), string(rune(i)))
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		graph.GetTopologicalOrder()
	}
}
