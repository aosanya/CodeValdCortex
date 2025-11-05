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

// aiWorkItemRefinementResponse represents the JSON structure returned by the AI
type aiWorkItemRefinementResponse struct {
	RefinedTitle        string   `json:"refined_title"`
	RefinedDescription  string   `json:"refined_description"`
	RefinedDeliverables []string `json:"refined_deliverables"`
	SuggestedType       string   `json:"suggested_type"`
	SuggestedPriority   string   `json:"suggested_priority"`
	SuggestedEffort     int      `json:"suggested_effort"`
	SuggestedTags       []string `json:"suggested_tags"`
	Explanation         string   `json:"explanation"`
	Changed             bool     `json:"changed"`
}
