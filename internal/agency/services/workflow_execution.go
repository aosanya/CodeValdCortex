package services

import (
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

// Connection represents a from→to connection in the execution graph
type Connection struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// GenerateExecutionFlow creates from→to connections from step structure
// This converts the simple Steps array into an execution graph for orchestration
//
// Execution logic:
// - Sequential steps (1 item): Item executes after all previous items complete
// - Parallel steps (2+ items): All items execute simultaneously (fork), next step waits for all (join)
//
// Example:
//
//	Steps: [step1(item1), step2(item2), step3(item3a, item3b), step4(item4)]
//	Connections: start→item1, item1→item2, item2→item3a, item2→item3b, item3a→item4, item3b→item4, item4→end
func GenerateExecutionFlow(steps models.Steps) []Connection {
	connections := []Connection{}

	// Start with the workflow start marker
	var previousItems []string = []string{"start"}

	for _, step := range steps {
		currentItems := make([]string, len(step.Items))

		// Collect all item IDs in current step
		for i, item := range step.Items {
			currentItems[i] = item.ID
		}

		if len(step.Items) == 1 {
			// Sequential: single item, connects from all previous items
			itemID := currentItems[0]
			for _, prevID := range previousItems {
				connections = append(connections, Connection{
					From: prevID,
					To:   itemID,
				})
			}
			// Only this item is previous for next step
			previousItems = []string{itemID}

		} else {
			// Parallel: multiple items, each connects from all previous (fork)
			for _, itemID := range currentItems {
				for _, prevID := range previousItems {
					connections = append(connections, Connection{
						From: prevID,
						To:   itemID,
					})
				}
			}
			// All parallel items become previous for next step (join)
			previousItems = currentItems
		}
	}

	// Connect all final items to the workflow end marker
	for _, prevID := range previousItems {
		connections = append(connections, Connection{
			From: prevID,
			To:   "end",
		})
	}

	return connections
}

// IsSequential returns true if the step executes sequentially (1 item)
func IsSequential(step models.Step) bool {
	return len(step.Items) == 1
}

// IsParallel returns true if the step executes in parallel (2+ items)
func IsParallel(step models.Step) bool {
	return len(step.Items) > 1
}
