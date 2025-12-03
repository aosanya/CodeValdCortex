package prompts

import (
	_ "embed"
	"log"
)

// Work Items System Prompts
var (
	//go:embed workitems/dynamic_system.txt
	WorkItemsDynamicSystem string

	//go:embed workitems/refinement_system.txt
	WorkItemsRefinementSystem string

	//go:embed workitems/generation_single_system.txt
	WorkItemsGenerationSingleSystem string

	//go:embed workitems/generation_multiple_system.txt
	WorkItemsGenerationMultipleSystem string

	//go:embed workitems/consolidation_system.txt
	WorkItemsConsolidationSystem string
)

// Deliverables Prompts
var (
	//go:embed deliverables/node_enhancement_system.txt
	DeliverablesNodeEnhancementSystem string
)

// Goals Prompts
var (
// Placeholder for future goal-specific prompts
)

// Workflows Prompts
var (
// Placeholder for future workflow-specific prompts
)

// TODO: Remove debug prints for MVP-054 after issue is resolved
func init() {
	log.Printf("[MVP-054] Prompts loaded - DeliverablesNodeEnhancementSystem: %d bytes", len(DeliverablesNodeEnhancementSystem))
	log.Printf("[MVP-054] Prompts loaded - WorkItemsDynamicSystem: %d bytes", len(WorkItemsDynamicSystem))
	log.Printf("[MVP-054] Prompts loaded - WorkItemsRefinementSystem: %d bytes", len(WorkItemsRefinementSystem))
}
