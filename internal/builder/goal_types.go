package builder

import (
	"github.com/aosanya/CodeValdCortex/internal/agency"
)

// RefineGoalRequest contains the context for refining a goal
type RefineGoalRequest struct {
	AgencyID       string             `json:"agency_id"`
	CurrentGoal    *agency.Goal       `json:"current_goal"`
	Description    string             `json:"description"`
	Scope          string             `json:"scope"`
	SuccessMetrics []string           `json:"success_metrics"`
	ExistingGoals  []*agency.Goal     `json:"existing_goals"`
	WorkItems      []*agency.WorkItem `json:"work_items"`
	AgencyContext  *agency.Agency     `json:"agency_context"`
}

// RefineGoalResponse contains the AI-refined goal
type RefineGoalResponse struct {
	RefinedDescription string   `json:"refined_description"`
	RefinedScope       string   `json:"refined_scope"`
	RefinedMetrics     []string `json:"refined_metrics"`
	SuggestedPriority  string   `json:"suggested_priority"`
	SuggestedCategory  string   `json:"suggested_category"`
	SuggestedTags      []string `json:"suggested_tags"`
	WasChanged         bool     `json:"was_changed"`
	Explanation        string   `json:"explanation"`
}

// GenerateGoalRequest contains the context for generating a new goal
type GenerateGoalRequest struct {
	AgencyID      string             `json:"agency_id"`
	UserInput     string             `json:"user_input"`
	ExistingGoals []*agency.Goal     `json:"existing_goals"`
	WorkItems     []*agency.WorkItem `json:"work_items"`
	AgencyContext *agency.Agency     `json:"agency_context"`
}

// GenerateGoalResponse contains the AI-generated goal
type GenerateGoalResponse struct {
	Description       string   `json:"description"`
	Scope             string   `json:"scope"`
	SuccessMetrics    []string `json:"success_metrics"`
	SuggestedCode     string   `json:"suggested_code"`
	SuggestedPriority string   `json:"suggested_priority"`
	SuggestedCategory string   `json:"suggested_category"`
	SuggestedTags     []string `json:"suggested_tags"`
	Explanation       string   `json:"explanation"`
}

// GenerateGoalsResponse contains multiple AI-generated goals
type GenerateGoalsResponse struct {
	Goals       []GenerateGoalResponse `json:"goals"`
	Explanation string                 `json:"explanation"`
}

// ConsolidateGoalsRequest contains the context for consolidating goals
type ConsolidateGoalsRequest struct {
	AgencyID      string             `json:"agency_id"`
	AgencyContext *agency.Agency     `json:"agency_context"`
	CurrentGoals  []*agency.Goal     `json:"current_goals"`
	WorkItems     []*agency.WorkItem `json:"work_items"`
}

// ConsolidateGoalsResponse contains the consolidated goals
type ConsolidateGoalsResponse struct {
	ConsolidatedGoals []ConsolidatedGoal `json:"consolidated_goals"`
	RemovedGoals      []string           `json:"removed_goals"` // Keys of goals that were consolidated/removed
	Summary           string             `json:"summary"`
	Explanation       string             `json:"explanation"`
}

// ConsolidatedGoal represents a goal after consolidation
type ConsolidatedGoal struct {
	Description       string   `json:"description"`
	Scope             string   `json:"scope"`
	SuccessMetrics    []string `json:"success_metrics"`
	SuggestedCode     string   `json:"suggested_code"`
	SuggestedPriority string   `json:"suggested_priority"`
	SuggestedCategory string   `json:"suggested_category"`
	SuggestedTags     []string `json:"suggested_tags"`
	ConsolidatedFrom  []string `json:"consolidated_from"` // Keys of original goals
	Rationale         string   `json:"rationale"`
}

// RefineGoalsRequest contains the context for dynamically processing goals based on user message
type RefineGoalsRequest struct {
	AgencyID      string             `json:"agency_id"`
	UserMessage   string             `json:"user_message"`   // Natural language instruction from user
	TargetGoals   []*agency.Goal     `json:"target_goals"`   // Optional: specific goals to operate on
	ExistingGoals []*agency.Goal     `json:"existing_goals"` // All existing goals for context
	WorkItems     []*agency.WorkItem `json:"work_items"`     // Work items for context
	AgencyContext *agency.Agency     `json:"agency_context"`
}

// RefineGoalsResponse contains the results of dynamic goal processing
type RefineGoalsResponse struct {
	Action           string                    `json:"action"`            // What action was determined (refine, generate, consolidate, enhance_all, etc.)
	RefinedGoals     []RefinedGoalResult       `json:"refined_goals"`     // Goals that were refined
	GeneratedGoals   []GenerateGoalResponse    `json:"generated_goals"`   // Newly generated goals
	ConsolidatedData *ConsolidateGoalsResponse `json:"consolidated_data"` // Consolidation results if applicable
	Explanation      string                    `json:"explanation"`       // What was done and why
	NoActionNeeded   bool                      `json:"no_action_needed"`  // True if goals are already optimal
}

// RefinedGoalResult represents a single refined goal
type RefinedGoalResult struct {
	OriginalKey        string   `json:"original_key"`
	RefinedDescription string   `json:"refined_description"`
	RefinedScope       string   `json:"refined_scope"`
	RefinedMetrics     []string `json:"refined_metrics"`
	SuggestedPriority  string   `json:"suggested_priority"`
	SuggestedCategory  string   `json:"suggested_category"`
	SuggestedTags      []string `json:"suggested_tags"`
	WasChanged         bool     `json:"was_changed"`
	Explanation        string   `json:"explanation"`
}
