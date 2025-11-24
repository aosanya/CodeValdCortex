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
			name:          "Draft to Published",
			initialState:  models.AgencyStateDraft,
			event:         "publish",
			expectedState: models.AgencyStatePublished,
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
			name:  "Draft to Activate (invalid)",
			state: models.AgencyStateDraft,
			event: "activate",
		},
		{
			name:  "Published to Pause (invalid)",
			state: models.AgencyStatePublished,
			event: "pause",
		},
		{
			name:  "Archived to Publish (invalid)",
			state: models.AgencyStateArchived,
			event: "publish",
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
			name:       "Valid: Draft to Publish",
			state:      models.AgencyStateDraft,
			event:      "publish",
			shouldPass: true,
		},
		{
			name:       "Invalid: Draft to Activate",
			state:      models.AgencyStateDraft,
			event:      "activate",
			shouldPass: false,
		},
		{
			name:       "Invalid: Published to Pause",
			state:      models.AgencyStatePublished,
			event:      "pause",
			shouldPass: false,
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

	// Test that guards are evaluated for Draft -> Published transition
	agency := &models.Agency{
		ID:    "test-agency",
		State: models.AgencyStateDraft,
	}

	// This should pass because guard stubs return nil
	err := sm.CanTransition(agency, "publish")
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

	// Simulate simplified lifecycle (only draft -> published now)
	transitions := []struct {
		event         string
		expectedState models.AgencyState
	}{
		{"publish", models.AgencyStatePublished},
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
