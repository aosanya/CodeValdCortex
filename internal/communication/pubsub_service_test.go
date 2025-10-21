package communication

import (
	"context"
	"testing"
	"time"

	driver "github.com/arangodb/go-driver"
)

// mockPubSubRepo is a mock implementation for testing PubSubService
type mockPubSubRepo struct {
	publications  map[string]*Publication
	subscriptions map[string]*Subscription
	deliveries    map[string]map[string]bool // pubID -> subID -> delivered
	createErr     error
	getErr        error
	updateErr     error
	queryErr      error
	deleteErr     error
}

func newMockPubSubRepo() *mockPubSubRepo {
	return &mockPubSubRepo{
		publications:  make(map[string]*Publication),
		subscriptions: make(map[string]*Subscription),
		deliveries:    make(map[string]map[string]bool),
	}
}

func (m *mockPubSubRepo) CreatePublication(ctx context.Context, pub *Publication) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.publications[pub.ID] = pub
	return nil
}

func (m *mockPubSubRepo) GetPublication(ctx context.Context, id string) (*Publication, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	pub, exists := m.publications[id]
	if !exists {
		return nil, driver.ArangoError{Code: 404}
	}
	return pub, nil
}

func (m *mockPubSubRepo) GetMatchingPublications(ctx context.Context, subscriptions []*Subscription, since time.Time) ([]*Publication, error) {
	if m.queryErr != nil {
		return nil, m.queryErr
	}
	var matching []*Publication
	matcher := &SubscriptionMatcher{}

	for _, pub := range m.publications {
		if pub.PublishedAt.Before(since) {
			continue
		}
		for _, sub := range subscriptions {
			if matcher.MatchesSubscription(pub, sub) {
				// Check if not already delivered
				if !m.hasBeenDelivered(pub.ID, sub.ID) {
					matching = append(matching, pub)
					break
				}
			}
		}
	}
	return matching, nil
}

func (m *mockPubSubRepo) DeleteExpiredPublications(ctx context.Context) (int, error) {
	count := 0
	now := time.Now()
	for id, pub := range m.publications {
		if pub.ExpiresAt.Before(now) {
			delete(m.publications, id)
			count++
		}
	}
	return count, nil
}

func (m *mockPubSubRepo) CreateSubscription(ctx context.Context, sub *Subscription) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.subscriptions[sub.ID] = sub
	return nil
}

func (m *mockPubSubRepo) GetSubscription(ctx context.Context, id string) (*Subscription, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	sub, exists := m.subscriptions[id]
	if !exists {
		return nil, driver.ArangoError{Code: 404}
	}
	return sub, nil
}

func (m *mockPubSubRepo) GetActiveSubscriptions(ctx context.Context, agentID string) ([]*Subscription, error) {
	if m.queryErr != nil {
		return nil, m.queryErr
	}
	var active []*Subscription
	for _, sub := range m.subscriptions {
		if sub.SubscriberAgentID == agentID && sub.Active {
			active = append(active, sub)
		}
	}
	return active, nil
}

func (m *mockPubSubRepo) DeactivateSubscription(ctx context.Context, id string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	sub, exists := m.subscriptions[id]
	if !exists {
		return driver.ArangoError{Code: 404}
	}
	sub.Active = false
	return nil
}

func (m *mockPubSubRepo) DeleteSubscription(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, exists := m.subscriptions[id]; !exists {
		return driver.ArangoError{Code: 404}
	}
	delete(m.subscriptions, id)
	return nil
}

func (m *mockPubSubRepo) CreateDelivery(ctx context.Context, delivery *PublicationDelivery) error {
	if m.createErr != nil {
		return m.createErr
	}
	// Extract pubID from delivery.From (format: "collection/id")
	// For testing, we'll assume From contains the publication ID
	pubID := delivery.From
	if m.deliveries[pubID] == nil {
		m.deliveries[pubID] = make(map[string]bool)
	}
	m.deliveries[pubID][delivery.SubscriptionID] = true
	return nil
}

func (m *mockPubSubRepo) UpdateSubscriptionLastMatched(ctx context.Context, id string, matchedAt time.Time) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	sub, exists := m.subscriptions[id]
	if !exists {
		return driver.ArangoError{Code: 404}
	}
	sub.LastMatchedAt = &matchedAt
	return nil
}

func (m *mockPubSubRepo) hasBeenDelivered(pubID, subID string) bool {
	if subs, exists := m.deliveries[pubID]; exists {
		return subs[subID]
	}
	return false
}

// TestPubSubService_Publish tests publishing events
func TestPubSubService_Publish(t *testing.T) {
	repo := newMockPubSubRepo()
	svc := NewPubSubService(repo)
	ctx := context.Background()

	tests := []struct {
		name             string
		publisherAgentID string
		publisherType    string
		eventName        string
		data             map[string]interface{}
		opts             *PublicationOptions
		expectError      bool
		validatePub      func(t *testing.T, pub *Publication)
	}{
		{
			name:             "publish basic event",
			publisherAgentID: "agent-1",
			publisherType:    "worker",
			eventName:        "state.changed",
			data:             map[string]interface{}{"old": "idle", "new": "active"},
			opts:             nil,
			expectError:      false,
			validatePub: func(t *testing.T, pub *Publication) {
				if pub.EventName != "state.changed" {
					t.Errorf("EventName = %v, want state.changed", pub.EventName)
				}
				if pub.PublisherAgentID != "agent-1" {
					t.Errorf("PublisherAgentID = %v, want agent-1", pub.PublisherAgentID)
				}
				if pub.PublicationType != PublicationTypeEvent {
					t.Errorf("PublicationType = %v, want event", pub.PublicationType)
				}
			},
		},
		{
			name:             "publish with options",
			publisherAgentID: "agent-1",
			publisherType:    "worker",
			eventName:        "task.completed",
			data:             map[string]interface{}{"task_id": "task-123"},
			opts: &PublicationOptions{
				Type:       PublicationTypeAlert,
				TTLSeconds: 300,
				Metadata:   map[string]string{"priority": "high"},
			},
			expectError: false,
			validatePub: func(t *testing.T, pub *Publication) {
				if pub.PublicationType != PublicationTypeAlert {
					t.Errorf("PublicationType = %v, want alert", pub.PublicationType)
				}
				if pub.TTLSeconds != 300 {
					t.Errorf("TTLSeconds = %v, want 300", pub.TTLSeconds)
				}
				if pub.Metadata["priority"] != "high" {
					t.Errorf("Metadata priority = %v, want high", pub.Metadata["priority"])
				}
			},
		},
		{
			name:             "missing publisher agent ID",
			publisherAgentID: "",
			publisherType:    "worker",
			eventName:        "test.event",
			data:             map[string]interface{}{},
			expectError:      true,
		},
		{
			name:             "missing event name",
			publisherAgentID: "agent-1",
			publisherType:    "worker",
			eventName:        "",
			data:             map[string]interface{}{},
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pubID, err := svc.Publish(ctx, tt.publisherAgentID, tt.publisherType, tt.eventName, tt.data, tt.opts)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if pubID == "" {
				t.Error("Expected non-empty publication ID")
			}

			// Validate stored publication
			pub := repo.publications[pubID]
			if pub == nil {
				t.Fatal("Publication not stored in repository")
			}

			if tt.validatePub != nil {
				tt.validatePub(t, pub)
			}
		})
	}
}

// TestPubSubService_Subscribe tests creating subscriptions
func TestPubSubService_Subscribe(t *testing.T) {
	repo := newMockPubSubRepo()
	svc := NewPubSubService(repo)
	ctx := context.Background()

	tests := []struct {
		name           string
		subscriberID   string
		subscriberType string
		pattern        string
		filters        *SubscriptionFilters
		expectError    bool
		validateSub    func(t *testing.T, sub *Subscription)
	}{
		{
			name:           "subscribe basic pattern",
			subscriberID:   "agent-1",
			subscriberType: "worker",
			pattern:        "state.*",
			filters:        nil,
			expectError:    false,
			validateSub: func(t *testing.T, sub *Subscription) {
				if sub.EventPattern != "state.*" {
					t.Errorf("EventPattern = %v, want state.*", sub.EventPattern)
				}
				if sub.SubscriberAgentID != "agent-1" {
					t.Errorf("SubscriberAgentID = %v, want agent-1", sub.SubscriberAgentID)
				}
				if !sub.Active {
					t.Error("Subscription should be active")
				}
			},
		},
		{
			name:           "subscribe with filters",
			subscriberID:   "agent-2",
			subscriberType: "worker",
			pattern:        "task.*",
			filters: &SubscriptionFilters{
				Types: []PublicationType{PublicationTypeEvent, PublicationTypeAlert},
			},
			expectError: false,
			validateSub: func(t *testing.T, sub *Subscription) {
				if len(sub.PublicationTypes) != 2 {
					t.Errorf("PublicationTypes length = %v, want 2", len(sub.PublicationTypes))
				}
			},
		},
		{
			name:           "missing subscriber ID",
			subscriberID:   "",
			subscriberType: "worker",
			pattern:        "test.*",
			expectError:    true,
		},
		{
			name:           "missing pattern",
			subscriberID:   "agent-1",
			subscriberType: "worker",
			pattern:        "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subID, err := svc.Subscribe(ctx, tt.subscriberID, tt.subscriberType, tt.pattern, tt.filters)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if subID == "" {
				t.Error("Expected non-empty subscription ID")
			}

			// Validate stored subscription
			sub := repo.subscriptions[subID]
			if sub == nil {
				t.Fatal("Subscription not stored in repository")
			}

			if tt.validateSub != nil {
				tt.validateSub(t, sub)
			}
		})
	}
}

// TestPubSubService_Unsubscribe tests deactivating subscriptions
func TestPubSubService_Unsubscribe(t *testing.T) {
	repo := newMockPubSubRepo()
	svc := NewPubSubService(repo)
	ctx := context.Background()

	// Create test subscription
	repo.subscriptions["sub-1"] = &Subscription{
		ID:     "sub-1",
		Active: true,
	}

	err := svc.Unsubscribe(ctx, "sub-1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	sub := repo.subscriptions["sub-1"]
	if sub.Active {
		t.Error("Subscription should be deactivated")
	}
}

// TestPubSubService_GetMatchingPublications tests retrieving matching publications
func TestPubSubService_GetMatchingPublications(t *testing.T) {
	repo := newMockPubSubRepo()
	svc := NewPubSubService(repo)
	ctx := context.Background()

	// Create test subscriptions
	sub1 := &Subscription{
		ID:                "sub-1",
		SubscriberAgentID: "agent-1",
		EventPattern:      "state.*",
		Active:            true,
	}
	sub2 := &Subscription{
		ID:                "sub-2",
		SubscriberAgentID: "agent-1",
		EventPattern:      "task.completed",
		Active:            true,
	}
	repo.subscriptions["sub-1"] = sub1
	repo.subscriptions["sub-2"] = sub2

	// Create test publications
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	older := now.Add(-2 * time.Hour)

	repo.publications["pub-1"] = &Publication{
		ID:          "pub-1",
		EventName:   "state.changed",
		PublishedAt: now,
	}
	repo.publications["pub-2"] = &Publication{
		ID:          "pub-2",
		EventName:   "task.completed",
		PublishedAt: now,
	}
	repo.publications["pub-3"] = &Publication{
		ID:          "pub-3",
		EventName:   "other.event",
		PublishedAt: now,
	}
	repo.publications["pub-4"] = &Publication{
		ID:          "pub-4",
		EventName:   "state.changed",
		PublishedAt: older, // Too old
	}

	pubs, err := svc.GetMatchingPublications(ctx, "agent-1", past)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should match pub-1 (state.*) and pub-2 (task.completed)
	if len(pubs) != 2 {
		t.Errorf("Expected 2 matching publications, got %d", len(pubs))
	}

	pubIDs := make(map[string]bool)
	for _, pub := range pubs {
		pubIDs[pub.ID] = true
	}

	if !pubIDs["pub-1"] {
		t.Error("Expected pub-1 in results")
	}
	if !pubIDs["pub-2"] {
		t.Error("Expected pub-2 in results")
	}
	if pubIDs["pub-3"] {
		t.Error("pub-3 should not match any subscription")
	}
}

// TestPubSubService_GetActiveSubscriptions tests retrieving active subscriptions
func TestPubSubService_GetActiveSubscriptions(t *testing.T) {
	repo := newMockPubSubRepo()
	svc := NewPubSubService(repo)
	ctx := context.Background()

	// Create test subscriptions
	repo.subscriptions["sub-1"] = &Subscription{
		ID:                "sub-1",
		SubscriberAgentID: "agent-1",
		Active:            true,
	}
	repo.subscriptions["sub-2"] = &Subscription{
		ID:                "sub-2",
		SubscriberAgentID: "agent-1",
		Active:            true,
	}
	repo.subscriptions["sub-3"] = &Subscription{
		ID:                "sub-3",
		SubscriberAgentID: "agent-1",
		Active:            false,
	}
	repo.subscriptions["sub-4"] = &Subscription{
		ID:                "sub-4",
		SubscriberAgentID: "agent-2",
		Active:            true,
	}

	subs, err := svc.GetActiveSubscriptions(ctx, "agent-1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should return sub-1 and sub-2 (active, for agent-1)
	if len(subs) != 2 {
		t.Errorf("Expected 2 active subscriptions, got %d", len(subs))
	}

	subIDs := make(map[string]bool)
	for _, sub := range subs {
		subIDs[sub.ID] = true
	}

	if !subIDs["sub-1"] {
		t.Error("Expected sub-1 in results")
	}
	if !subIDs["sub-2"] {
		t.Error("Expected sub-2 in results")
	}
	if subIDs["sub-3"] {
		t.Error("sub-3 should not be in results (inactive)")
	}
	if subIDs["sub-4"] {
		t.Error("sub-4 should not be in results (wrong agent)")
	}
}
