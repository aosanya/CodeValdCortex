package communication

import (
	"testing"
)

// TestMatchesPattern tests the pattern matching functionality
func TestMatchesPattern(t *testing.T) {
	matcher := NewSubscriptionMatcher()

	tests := []struct {
		name      string
		eventName string
		pattern   string
		want      bool
	}{
		{
			name:      "exact match",
			eventName: "state.changed",
			pattern:   "state.changed",
			want:      true,
		},
		{
			name:      "wildcard all",
			eventName: "state.changed",
			pattern:   "*",
			want:      true,
		},
		{
			name:      "prefix wildcard",
			eventName: "state.changed",
			pattern:   "state.*",
			want:      true,
		},
		{
			name:      "suffix wildcard",
			eventName: "task.processing.completed",
			pattern:   "*.completed",
			want:      true,
		},
		{
			name:      "middle wildcard",
			eventName: "task.processing.completed",
			pattern:   "task.*.completed",
			want:      true,
		},
		{
			name:      "no match",
			eventName: "state.changed",
			pattern:   "task.*",
			want:      false,
		},
		{
			name:      "partial no match",
			eventName: "state.changed.extra",
			pattern:   "state.changed",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matcher.MatchesPattern(tt.eventName, tt.pattern)
			if got != tt.want {
				t.Errorf("MatchesPattern(%q, %q) = %v, want %v",
					tt.eventName, tt.pattern, got, tt.want)
			}
		})
	}
}

// TestMatchesSubscription tests subscription matching logic
func TestMatchesSubscription(t *testing.T) {
	matcher := NewSubscriptionMatcher()

	// Helper function to create string pointers
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name string
		pub  *Publication
		sub  *Subscription
		want bool
	}{
		{
			name: "basic pattern match",
			pub: &Publication{
				PublisherAgentID:   "agent-1",
				PublisherAgentType: "worker",
				PublicationType:    PublicationTypeStatusChange,
				EventName:          "state.changed",
			},
			sub: &Subscription{
				SubscriberAgentID:   "agent-2",
				SubscriberAgentType: "coordinator",
				EventPattern:        "state.*",
				Active:              true,
			},
			want: true,
		},
		{
			name: "inactive subscription",
			pub: &Publication{
				EventName: "state.changed",
			},
			sub: &Subscription{
				EventPattern: "state.*",
				Active:       false,
			},
			want: false,
		},
		{
			name: "publisher ID filter match",
			pub: &Publication{
				PublisherAgentID: "agent-1",
				EventName:        "state.changed",
			},
			sub: &Subscription{
				PublisherAgentID: strPtr("agent-1"),
				EventPattern:     "state.*",
				Active:           true,
			},
			want: true,
		},
		{
			name: "publisher ID filter no match",
			pub: &Publication{
				PublisherAgentID: "agent-1",
				EventName:        "state.changed",
			},
			sub: &Subscription{
				PublisherAgentID: strPtr("agent-2"),
				EventPattern:     "state.*",
				Active:           true,
			},
			want: false,
		},
		{
			name: "publisher type filter match",
			pub: &Publication{
				PublisherAgentType: "worker",
				EventName:          "state.changed",
			},
			sub: &Subscription{
				PublisherAgentType: strPtr("worker"),
				EventPattern:       "state.*",
				Active:             true,
			},
			want: true,
		},
		{
			name: "publication type filter match",
			pub: &Publication{
				PublicationType: PublicationTypeStatusChange,
				EventName:       "state.changed",
			},
			sub: &Subscription{
				PublicationTypes: []PublicationType{PublicationTypeStatusChange},
				EventPattern:     "state.*",
				Active:           true,
			},
			want: true,
		},
		{
			name: "publication type filter no match",
			pub: &Publication{
				PublicationType: PublicationTypeEvent,
				EventName:       "state.changed",
			},
			sub: &Subscription{
				PublicationTypes: []PublicationType{PublicationTypeStatusChange},
				EventPattern:     "state.*",
				Active:           true,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matcher.MatchesSubscription(tt.pub, tt.sub)
			if got != tt.want {
				t.Errorf("MatchesSubscription() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFilterMatchingPublications tests filtering publications by subscriptions
func TestFilterMatchingPublications(t *testing.T) {
	matcher := NewSubscriptionMatcher()

	publications := []*Publication{
		{
			ID:        "pub-1",
			EventName: "state.changed",
		},
		{
			ID:        "pub-2",
			EventName: "task.completed",
		},
		{
			ID:        "pub-3",
			EventName: "state.updated",
		},
	}

	subscriptions := []*Subscription{
		{
			ID:           "sub-1",
			EventPattern: "state.*",
			Active:       true,
		},
		{
			ID:           "sub-2",
			EventPattern: "task.*",
			Active:       true,
		},
	}

	matched := matcher.FilterMatchingPublications(publications, subscriptions)

	// Should match all 3 publications
	if len(matched) != 3 {
		t.Errorf("FilterMatchingPublications() returned %d matches, want 3", len(matched))
	}

	// Verify specific matches
	matchedIDs := make(map[string]bool)
	for _, pub := range matched {
		matchedIDs[pub.ID] = true
	}

	expectedIDs := []string{"pub-1", "pub-2", "pub-3"}
	for _, id := range expectedIDs {
		if !matchedIDs[id] {
			t.Errorf("Expected publication %s to be matched", id)
		}
	}
}

// TestGetMatchingSubscriptions tests finding subscriptions that match a publication
func TestGetMatchingSubscriptions(t *testing.T) {
	matcher := NewSubscriptionMatcher()

	pub := &Publication{
		EventName: "state.changed",
	}

	subscriptions := []*Subscription{
		{
			ID:           "sub-1",
			EventPattern: "state.*",
			Active:       true,
		},
		{
			ID:           "sub-2",
			EventPattern: "task.*",
			Active:       true,
		},
		{
			ID:           "sub-3",
			EventPattern: "*",
			Active:       true,
		},
	}

	matched := matcher.GetMatchingSubscriptions(pub, subscriptions)

	// Should match sub-1 (state.*) and sub-3 (*), but not sub-2 (task.*)
	if len(matched) != 2 {
		t.Errorf("GetMatchingSubscriptions() returned %d matches, want 2", len(matched))
	}

	// Verify specific matches
	matchedIDs := make(map[string]bool)
	for _, sub := range matched {
		matchedIDs[sub.ID] = true
	}

	if !matchedIDs["sub-1"] {
		t.Error("Expected sub-1 to match")
	}
	if matchedIDs["sub-2"] {
		t.Error("Expected sub-2 to not match")
	}
	if !matchedIDs["sub-3"] {
		t.Error("Expected sub-3 to match")
	}
}
