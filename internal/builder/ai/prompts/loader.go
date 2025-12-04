package prompts

import (
	_ "embed"
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
