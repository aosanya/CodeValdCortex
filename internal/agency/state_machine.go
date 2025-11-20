package agency

import (
	"fmt"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

// StateTransition represents a valid state transition
type StateTransition struct {
	From    models.AgencyState
	To      models.AgencyState
	Event   string
	Guards  []Guard
	Actions []Action
}

// Guard is a function that checks if transition is allowed
type Guard func(*models.Agency) error

// Action is a function executed during transition
type Action func(*models.Agency) error

// AgencyStateMachine manages state transitions
type AgencyStateMachine struct {
	transitions map[string][]StateTransition
}

// NewAgencyStateMachine creates a new state machine
func NewAgencyStateMachine() *AgencyStateMachine {
	sm := &AgencyStateMachine{
		transitions: make(map[string][]StateTransition),
	}
	sm.defineTransitions()
	return sm
}

// defineTransitions sets up all valid state transitions
func (sm *AgencyStateMachine) defineTransitions() {
	transitions := []StateTransition{
		{
			From:  models.AgencyStateDraft,
			To:    models.AgencyStateValidated,
			Event: "validate",
			Guards: []Guard{
				guardHasIntroduction,
				guardHasGoals,
				guardHasRoles,
				guardHasWorkItems,
				guardHasWorkflows,
				guardHasRACIMatrix,
			},
			Actions: []Action{
				actionMarkAsValidated,
			},
		},
		{
			From:  models.AgencyStateValidated,
			To:    models.AgencyStatePublished,
			Event: "publish",
			Guards: []Guard{
				guardIsValidated,
				guardNoDuplicatePublication,
			},
			Actions: []Action{
				actionCreatePublication,
				actionUpdatePublishMetadata,
			},
		},
		{
			From:  models.AgencyStatePublished,
			To:    models.AgencyStateActive,
			Event: "activate",
			Guards: []Guard{
				guardIsPublished,
				guardResourcesAvailable,
			},
			Actions: []Action{
				actionSpawnAgents,
				actionInitializeWorkflows,
				actionStartMonitoring,
				actionUpdateActivationMetadata,
			},
		},
		{
			From:  models.AgencyStateActive,
			To:    models.AgencyStatePaused,
			Event: "pause",
			Actions: []Action{
				actionStopAcceptingWork,
				actionPauseAgents,
			},
		},
		{
			From:  models.AgencyStatePaused,
			To:    models.AgencyStateActive,
			Event: "resume",
			Actions: []Action{
				actionResumeAgents,
				actionResumeAcceptingWork,
			},
		},
		{
			From:  models.AgencyStateActive,
			To:    models.AgencyStateDraining,
			Event: "drain",
			Actions: []Action{
				actionStopAcceptingNewWork,
			},
		},
		{
			From:  models.AgencyStateDraining,
			To:    models.AgencyStateStopped,
			Event: "drain_complete",
			Actions: []Action{
				actionStopAllAgents,
				actionCleanupResources,
			},
		},
		{
			From:  models.AgencyStateActive,
			To:    models.AgencyStateStopped,
			Event: "force_stop",
			Actions: []Action{
				actionStopAllAgents,
				actionCleanupResources,
			},
		},
		{
			From:  models.AgencyStatePaused,
			To:    models.AgencyStateStopped,
			Event: "force_stop",
			Actions: []Action{
				actionStopAllAgents,
				actionCleanupResources,
			},
		},
		{
			From:  models.AgencyStateDraining,
			To:    models.AgencyStateStopped,
			Event: "force_stop",
			Actions: []Action{
				actionStopAllAgents,
				actionCleanupResources,
			},
		},
	}

	for _, t := range transitions {
		key := string(t.From)
		sm.transitions[key] = append(sm.transitions[key], t)
	}
}

// CanTransition checks if transition is allowed
func (sm *AgencyStateMachine) CanTransition(agency *models.Agency, event string) error {
	transitions, ok := sm.transitions[string(agency.State)]
	if !ok {
		return fmt.Errorf("no transitions defined for state: %s", agency.State)
	}

	for _, t := range transitions {
		if t.Event == event {
			// Check guards
			for _, guard := range t.Guards {
				if err := guard(agency); err != nil {
					return fmt.Errorf("guard failed: %w", err)
				}
			}
			return nil
		}
	}

	return fmt.Errorf("event '%s' not valid for state '%s'", event, agency.State)
}

// Transition executes a state transition
func (sm *AgencyStateMachine) Transition(agency *models.Agency, event string) error {
	if err := sm.CanTransition(agency, event); err != nil {
		return err
	}

	transitions := sm.transitions[string(agency.State)]
	for _, t := range transitions {
		if t.Event == event {
			// Execute actions
			for _, action := range t.Actions {
				if err := action(agency); err != nil {
					return fmt.Errorf("action failed: %w", err)
				}
			}

			// Update state
			agency.State = t.To
			return nil
		}
	}

	return fmt.Errorf("transition not found")
}

// Guard implementations (stubs for now - will be implemented in later tasks)

func guardHasIntroduction(a *models.Agency) error {
	// TODO: Check agency has introduction (MVP-PUB-003)
	return nil
}

func guardHasGoals(a *models.Agency) error {
	// TODO: Check agency has goals (MVP-PUB-003)
	return nil
}

func guardHasRoles(a *models.Agency) error {
	// TODO: Check agency has roles (MVP-PUB-003)
	return nil
}

func guardHasWorkItems(a *models.Agency) error {
	// TODO: Check agency has work items (MVP-PUB-003)
	return nil
}

func guardHasWorkflows(a *models.Agency) error {
	// TODO: Check agency has workflows (MVP-PUB-003)
	return nil
}

func guardHasRACIMatrix(a *models.Agency) error {
	// TODO: Check agency has RACI matrix (MVP-PUB-003)
	return nil
}

func guardIsValidated(a *models.Agency) error {
	if a.State != models.AgencyStateValidated {
		return fmt.Errorf("agency must be validated")
	}
	return nil
}

func guardNoDuplicatePublication(a *models.Agency) error {
	// TODO: Check no active publication exists (MVP-PUB-003)
	return nil
}

func guardIsPublished(a *models.Agency) error {
	if a.State != models.AgencyStatePublished {
		return fmt.Errorf("agency must be published")
	}
	return nil
}

func guardResourcesAvailable(a *models.Agency) error {
	// TODO: Check resource quotas available (MVP-PUB-004)
	return nil
}

// Action implementations (stubs for now)

func actionMarkAsValidated(a *models.Agency) error {
	// Will be implemented in MVP-PUB-003
	return nil
}

func actionCreatePublication(a *models.Agency) error {
	// Will be implemented in MVP-PUB-003
	return nil
}

func actionUpdatePublishMetadata(a *models.Agency) error {
	// Will be implemented in MVP-PUB-003
	return nil
}

func actionSpawnAgents(a *models.Agency) error {
	// Will be implemented in MVP-PUB-004
	return nil
}

func actionInitializeWorkflows(a *models.Agency) error {
	// Will be implemented in MVP-PUB-004
	return nil
}

func actionStartMonitoring(a *models.Agency) error {
	// Will be implemented in MVP-PUB-004
	return nil
}

func actionUpdateActivationMetadata(a *models.Agency) error {
	// Will be implemented in MVP-PUB-004
	return nil
}

func actionStopAcceptingWork(a *models.Agency) error {
	// Will be implemented in MVP-PUB-004
	return nil
}

func actionPauseAgents(a *models.Agency) error {
	// Will be implemented in MVP-PUB-004
	return nil
}

func actionResumeAgents(a *models.Agency) error {
	// Will be implemented in MVP-PUB-004
	return nil
}

func actionResumeAcceptingWork(a *models.Agency) error {
	// Will be implemented in MVP-PUB-004
	return nil
}

func actionStopAcceptingNewWork(a *models.Agency) error {
	// Will be implemented in MVP-PUB-004
	return nil
}

func actionStopAllAgents(a *models.Agency) error {
	// Will be implemented in MVP-PUB-004
	return nil
}

func actionCleanupResources(a *models.Agency) error {
	// Will be implemented in MVP-PUB-004
	return nil
}
