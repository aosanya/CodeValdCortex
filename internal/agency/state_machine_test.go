package agency

import (
	"testing"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

func TestNewAgencyStateMachine(t *testing.T) {
	sm := NewAgencyStateMachine()
	if sm == nil {
		t.Fatal("Expected state machine, got nil")
	}
	if sm.transitions == nil {
		t.Fatal("Expected transitions map, got nil")
	}
}

func TestValidTransitions(t *testing.T) {
	sm := NewAgencyStateMachine()

	tests := []struct {
		name          string
		initialState  models.AgencyState
		event         string
		expectedState models.AgencyState
		shouldFail    bool
	}{
		{
			name:          "Draft to Validated",
			initialState:  models.AgencyStateDraft,
			event:         "validate",
			expectedState: models.AgencyStateValidated,
			shouldFail:    false,
		},
		{
			name:          "Validated to Published",
			initialState:  models.AgencyStateValidated,
			event:         "publish",
			expectedState: models.AgencyStatePublished,
			shouldFail:    false,
		},
		{
			name:          "Published to Active",
			initialState:  models.AgencyStatePublished,
			event:         "activate",
			expectedState: models.AgencyStateActive,
			shouldFail:    false,
		},
		{
			name:          "Active to Paused",
			initialState:  models.AgencyStateActive,
			event:         "pause",
			expectedState: models.AgencyStatePaused,
			shouldFail:    false,
		},
		{
			name:          "Paused to Active",
			initialState:  models.AgencyStatePaused,
			event:         "resume",
			expectedState: models.AgencyStateActive,
			shouldFail:    false,
		},
		{
			name:          "Active to Draining",
			initialState:  models.AgencyStateActive,
			event:         "drain",
			expectedState: models.AgencyStateDraining,
			shouldFail:    false,
		},
		{
			name:          "Draining to Stopped",
			initialState:  models.AgencyStateDraining,
			event:         "drain_complete",
			expectedState: models.AgencyStateStopped,
			shouldFail:    false,
		},
		{
			name:          "Active to Stopped (force)",
			initialState:  models.AgencyStateActive,
			event:         "force_stop",
			expectedState: models.AgencyStateStopped,
			shouldFail:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agency := &models.Agency{
				ID:    "test-agency",
				State: tt.initialState,
			}

			err := sm.Transition(agency, tt.event)

			if tt.shouldFail && err == nil {
				t.Errorf("Expected transition to fail, but it succeeded")
			}

			if !tt.shouldFail && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.shouldFail && agency.State != tt.expectedState {
				t.Errorf("Expected state %s, got %s", tt.expectedState, agency.State)
			}
		})
	}
}

func TestInvalidTransitions(t *testing.T) {
	sm := NewAgencyStateMachine()

	tests := []struct {
		name  string
		state models.AgencyState
		event string
	}{
		{
			name:  "Draft to Active (invalid)",
			state: models.AgencyStateDraft,
			event: "activate",
		},
		{
			name:  "Draft to Published (invalid)",
			state: models.AgencyStateDraft,
			event: "publish",
		},
		{
			name:  "Validated to Active (invalid)",
			state: models.AgencyStateValidated,
			event: "activate",
		},
		{
			name:  "Published to Paused (invalid)",
			state: models.AgencyStatePublished,
			event: "pause",
		},
		{
			name:  "Stopped to Active (invalid)",
			state: models.AgencyStateStopped,
			event: "activate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agency := &models.Agency{
				ID:    "test-agency",
				State: tt.state,
			}

			err := sm.Transition(agency, tt.event)
			if err == nil {
				t.Errorf("Expected error for invalid transition, got nil")
			}
		})
	}
}

func TestCanTransition(t *testing.T) {
	sm := NewAgencyStateMachine()

	tests := []struct {
		name       string
		state      models.AgencyState
		event      string
		shouldPass bool
	}{
		{
			name:       "Valid: Draft to Validated",
			state:      models.AgencyStateDraft,
			event:      "validate",
			shouldPass: true,
		},
		{
			name:       "Invalid: Draft to Activate",
			state:      models.AgencyStateDraft,
			event:      "activate",
			shouldPass: false,
		},
		{
			name:       "Valid: Published to Active",
			state:      models.AgencyStatePublished,
			event:      "activate",
			shouldPass: true,
		},
		{
			name:       "Valid: Active to Paused",
			state:      models.AgencyStateActive,
			event:      "pause",
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agency := &models.Agency{
				ID:    "test-agency",
				State: tt.state,
			}

			err := sm.CanTransition(agency, tt.event)

			if tt.shouldPass && err != nil {
				t.Errorf("Expected transition to be allowed, got error: %v", err)
			}

			if !tt.shouldPass && err == nil {
				t.Errorf("Expected transition to be denied, but it was allowed")
			}
		})
	}
}

func TestGuardEvaluation(t *testing.T) {
	sm := NewAgencyStateMachine()

	// Test that guards are evaluated for Draft -> Validated transition
	agency := &models.Agency{
		ID:    "test-agency",
		State: models.AgencyStateDraft,
	}

	// This should pass because guard stubs return nil
	err := sm.CanTransition(agency, "validate")
	if err != nil {
		t.Errorf("Guard evaluation failed: %v", err)
	}
}

func TestStateNotChangedOnFailure(t *testing.T) {
	sm := NewAgencyStateMachine()

	agency := &models.Agency{
		ID:    "test-agency",
		State: models.AgencyStateDraft,
	}

	originalState := agency.State

	// Try invalid transition
	_ = sm.Transition(agency, "invalid_event")

	if agency.State != originalState {
		t.Errorf("State should not change on failed transition. Expected %s, got %s",
			originalState, agency.State)
	}
}

func TestMultipleTransitions(t *testing.T) {
	sm := NewAgencyStateMachine()

	agency := &models.Agency{
		ID:    "test-agency",
		State: models.AgencyStateDraft,
	}

	// Simulate full lifecycle
	transitions := []struct {
		event         string
		expectedState models.AgencyState
	}{
		{"validate", models.AgencyStateValidated},
		{"publish", models.AgencyStatePublished},
		{"activate", models.AgencyStateActive},
		{"pause", models.AgencyStatePaused},
		{"resume", models.AgencyStateActive},
		{"drain", models.AgencyStateDraining},
		{"drain_complete", models.AgencyStateStopped},
	}

	for i, trans := range transitions {
		err := sm.Transition(agency, trans.event)
		if err != nil {
			t.Fatalf("Transition %d (%s) failed: %v", i, trans.event, err)
		}

		if agency.State != trans.expectedState {
			t.Errorf("After transition %d: expected state %s, got %s",
				i, trans.expectedState, agency.State)
		}
	}
}
