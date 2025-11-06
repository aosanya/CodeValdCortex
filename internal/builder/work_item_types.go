package builder

import (
	"github.com/aosanya/CodeValdCortex/internal/agency"
)

// RefineWorkItemRequest contains the context for refining a work item
type RefineWorkItemRequest struct {
	AgencyID          string             `json:"agency_id"`
	CurrentWorkItem   *agency.WorkItem   `json:"current_work_item"`
	Title             string             `json:"title"`
	Description       string             `json:"description"`
	Deliverables      []string           `json:"deliverables"`
	ExistingWorkItems []*agency.WorkItem `json:"existing_work_items"`
	Goals             []*agency.Goal     `json:"goals"`
	AgencyContext     *agency.Agency     `json:"agency_context"`
}

// RefineWorkItemResponse contains the AI-refined work item
type RefineWorkItemResponse struct {
	RefinedTitle        string   `json:"refined_title"`
	RefinedDescription  string   `json:"refined_description"`
	RefinedDeliverables []string `json:"refined_deliverables"`
	SuggestedType       string   `json:"suggested_type"`
	SuggestedPriority   string   `json:"suggested_priority"`
	SuggestedEffort     int      `json:"suggested_effort"`
	SuggestedTags       []string `json:"suggested_tags"`
	WasChanged          bool     `json:"was_changed"`
	Explanation         string   `json:"explanation"`
}

// GenerateWorkItemRequest contains the context for generating a new work item
type GenerateWorkItemRequest struct {
	AgencyID          string             `json:"agency_id"`
	AgencyContext     *agency.Agency     `json:"agency_context"`
	ExistingWorkItems []*agency.WorkItem `json:"existing_work_items"`
	Goals             []*agency.Goal     `json:"goals"`
	UserInput         string             `json:"user_input"`
}

// GenerateWorkItemResponse contains the AI-generated work item
type GenerateWorkItemResponse struct {
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	Deliverables      []string `json:"deliverables"`
	SuggestedCode     string   `json:"suggested_code"`
	SuggestedType     string   `json:"suggested_type"`
	SuggestedPriority string   `json:"suggested_priority"`
	SuggestedEffort   int      `json:"suggested_effort"`
	SuggestedTags     []string `json:"suggested_tags"`
	Explanation       string   `json:"explanation"`
}

// GenerateWorkItemsResponse contains multiple AI-generated work items
type GenerateWorkItemsResponse struct {
	WorkItems   []GenerateWorkItemResponse `json:"work_items"`
	Explanation string                     `json:"explanation"`
}

// ConsolidateWorkItemsRequest contains the context for consolidating work items
type ConsolidateWorkItemsRequest struct {
	AgencyID         string             `json:"agency_id"`
	AgencyContext    *agency.Agency     `json:"agency_context"`
	CurrentWorkItems []*agency.WorkItem `json:"current_work_items"`
	Goals            []*agency.Goal     `json:"goals"`
}

// ConsolidateWorkItemsResponse contains the consolidated work items
type ConsolidateWorkItemsResponse struct {
	ConsolidatedWorkItems []ConsolidatedWorkItem `json:"consolidated_work_items"`
	RemovedWorkItems      []string               `json:"removed_work_items"` // Keys of work items that were consolidated/removed
	Summary               string                 `json:"summary"`
	Explanation           string                 `json:"explanation"`
}

// ConsolidatedWorkItem represents a work item after consolidation
type ConsolidatedWorkItem struct {
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	Deliverables      []string `json:"deliverables"`
	SuggestedCode     string   `json:"suggested_code"`
	SuggestedType     string   `json:"suggested_type"`
	SuggestedPriority string   `json:"suggested_priority"`
	SuggestedEffort   int      `json:"suggested_effort"`
	SuggestedTags     []string `json:"suggested_tags"`
	ConsolidatedFrom  []string `json:"consolidated_from"` // Keys of original work items
	Rationale         string   `json:"rationale"`
}

// RefineWorkItemsRequest contains the context for dynamic work item processing
type RefineWorkItemsRequest struct {
	AgencyID          string             `json:"agency_id"`
	UserMessage       string             `json:"user_message"`
	TargetWorkItems   []*agency.WorkItem `json:"target_work_items"`   // Specific work items to operate on (nil means all)
	ExistingWorkItems []*agency.WorkItem `json:"existing_work_items"` // All current work items for context
	Goals             []*agency.Goal     `json:"goals"`               // Agency goals for context
	AgencyContext     *agency.Agency     `json:"agency_context"`
}

// RefineWorkItemsResponse contains the dynamic work item processing results
type RefineWorkItemsResponse struct {
	Action             string                        `json:"action"`               // What action was determined (refine, generate, consolidate, enhance_all, etc.)
	RefinedWorkItems   []RefinedWorkItemResult       `json:"refined_work_items"`   // Work items that were refined
	GeneratedWorkItems []GenerateWorkItemResponse    `json:"generated_work_items"` // Newly generated work items
	ConsolidatedData   *ConsolidateWorkItemsResponse `json:"consolidated_data"`    // Consolidation results if applicable
	Explanation        string                        `json:"explanation"`          // What was done and why
	NoActionNeeded     bool                          `json:"no_action_needed"`     // True if work items are already optimal
}

// RefinedWorkItemResult represents a single refined work item
type RefinedWorkItemResult struct {
	OriginalKey         string   `json:"original_key"`
	RefinedTitle        string   `json:"refined_title"`
	RefinedDescription  string   `json:"refined_description"`
	RefinedDeliverables []string `json:"refined_deliverables"`
	SuggestedCode       string   `json:"suggested_code"` // Updated work item code
	SuggestedType       string   `json:"suggested_type"`
	SuggestedPriority   string   `json:"suggested_priority"`
	SuggestedEffort     int      `json:"suggested_effort"`
	SuggestedTags       []string `json:"suggested_tags"`
	WasChanged          bool     `json:"was_changed"`
	Explanation         string   `json:"explanation"`
}
