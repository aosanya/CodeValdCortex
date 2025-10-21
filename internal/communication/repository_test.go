package communication

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/config"
	"github.com/aosanya/CodeValdCortex/internal/database"
)

// skipIfNoDatabase skips the test if ArangoDB is not available
func skipIfNoDatabase(t *testing.T) *database.ArangoClient {
	host := os.Getenv("ARANGO_HOST")
	if host == "" {
		host = "localhost"
	}

	portStr := os.Getenv("ARANGO_PORT")
	port := 8529
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	dbName := os.Getenv("ARANGO_TEST_DB")
	if dbName == "" {
		dbName = "codeval_cortex_test"
	}

	user := os.Getenv("ARANGO_USER")
	if user == "" {
		user = "root"
	}

	password := os.Getenv("ARANGO_PASSWORD")
	if password == "" {
		password = ""
	}

	cfg := &config.DatabaseConfig{
		Host:     host,
		Port:     port,
		Database: dbName,
		Username: user,
		Password: password,
	}

	client, err := database.NewArangoClient(cfg)
	if err != nil {
		t.Skipf("Skipping test: ArangoDB not available: %v", err)
		return nil
	}

	return client
}

// cleanupTestData removes all test data from collections
func cleanupTestData(t *testing.T, repo *Repository) {
	ctx := context.Background()

	// Delete all documents from test collections
	if repo.messagesCol != nil {
		repo.messagesCol.Truncate(ctx)
	}
	if repo.publicationsCol != nil {
		repo.publicationsCol.Truncate(ctx)
	}
	if repo.subscriptionsCol != nil {
		repo.subscriptionsCol.Truncate(ctx)
	}
	if repo.deliveriesCol != nil {
		repo.deliveriesCol.Truncate(ctx)
	}
}

// TestRepository_CreateAndGetMessage tests message creation and retrieval
func TestRepository_CreateAndGetMessage(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()

	msg := &Message{
		FromAgentID: "agent-1",
		ToAgentID:   "agent-2",
		MessageType: MessageTypeTaskRequest,
		Payload:     map[string]interface{}{"task": "test"},
		Status:      MessageStatusPending,
		Priority:    5,
		CreatedAt:   now,
	}

	// Create message
	err = repo.CreateMessage(ctx, msg)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	if msg.ID == "" {
		t.Error("Message ID should be set after creation")
	}

	// Get message
	retrieved, err := repo.GetMessage(ctx, msg.ID)
	if err != nil {
		t.Fatalf("Failed to get message: %v", err)
	}

	if retrieved.ID != msg.ID {
		t.Errorf("ID = %v, want %v", retrieved.ID, msg.ID)
	}
	if retrieved.FromAgentID != msg.FromAgentID {
		t.Errorf("FromAgentID = %v, want %v", retrieved.FromAgentID, msg.FromAgentID)
	}
	if retrieved.ToAgentID != msg.ToAgentID {
		t.Errorf("ToAgentID = %v, want %v", retrieved.ToAgentID, msg.ToAgentID)
	}
	if retrieved.MessageType != msg.MessageType {
		t.Errorf("MessageType = %v, want %v", retrieved.MessageType, msg.MessageType)
	}
	if retrieved.Status != msg.Status {
		t.Errorf("Status = %v, want %v", retrieved.Status, msg.Status)
	}
}

// TestRepository_GetPendingMessages tests retrieving pending messages
func TestRepository_GetPendingMessages(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()
	future := now.Add(1 * time.Hour)
	past := now.Add(-1 * time.Hour)

	// Create test messages
	messages := []*Message{
		{
			FromAgentID: "agent-1",
			ToAgentID:   "agent-2",
			MessageType: MessageTypeTaskRequest,
			Payload:     map[string]interface{}{"task": "test-1"},
			Status:      MessageStatusPending,
			Priority:    8,
			CreatedAt:   now,
			ExpiresAt:   &future,
		},
		{
			FromAgentID: "agent-1",
			ToAgentID:   "agent-2",
			MessageType: MessageTypeTaskRequest,
			Payload:     map[string]interface{}{"task": "test-2"},
			Status:      MessageStatusPending,
			Priority:    5,
			CreatedAt:   now,
			ExpiresAt:   &future,
		},
		{
			FromAgentID: "agent-1",
			ToAgentID:   "agent-2",
			MessageType: MessageTypeTaskRequest,
			Payload:     map[string]interface{}{"task": "test-3"},
			Status:      MessageStatusDelivered,
			Priority:    5,
			CreatedAt:   now,
			ExpiresAt:   &future,
		},
		{
			FromAgentID: "agent-1",
			ToAgentID:   "agent-3",
			MessageType: MessageTypeTaskRequest,
			Payload:     map[string]interface{}{"task": "test-4"},
			Status:      MessageStatusPending,
			Priority:    5,
			CreatedAt:   now,
			ExpiresAt:   &future,
		},
		{
			FromAgentID: "agent-1",
			ToAgentID:   "agent-2",
			MessageType: MessageTypeTaskRequest,
			Payload:     map[string]interface{}{"task": "test-5"},
			Status:      MessageStatusPending,
			Priority:    5,
			CreatedAt:   now,
			ExpiresAt:   &past, // Expired
		},
	}

	for _, msg := range messages {
		if err := repo.CreateMessage(ctx, msg); err != nil {
			t.Fatalf("Failed to create message: %v", err)
		}
	}

	// Get pending messages for agent-2
	pending, err := repo.GetPendingMessages(ctx, "agent-2", 100)
	if err != nil {
		t.Fatalf("Failed to get pending messages: %v", err)
	}

	// Should return 2 messages (msg-0 and msg-1), not msg-2 (delivered), msg-3 (wrong agent), msg-4 (expired)
	if len(pending) != 2 {
		t.Errorf("Expected 2 pending messages, got %d", len(pending))
	}

	// Check that messages are ordered by priority (highest first)
	if len(pending) >= 2 {
		if pending[0].Priority != 8 {
			t.Errorf("First message priority = %d, want 8", pending[0].Priority)
		}
		if pending[1].Priority != 5 {
			t.Errorf("Second message priority = %d, want 5", pending[1].Priority)
		}
	}
}

// TestRepository_UpdateMessageStatus tests updating message status
func TestRepository_UpdateMessageStatus(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()

	msg := &Message{
		FromAgentID: "agent-1",
		ToAgentID:   "agent-2",
		MessageType: MessageTypeTaskRequest,
		Payload:     map[string]interface{}{"task": "test"},
		Status:      MessageStatusPending,
		Priority:    5,
		CreatedAt:   now,
	}

	if err := repo.CreateMessage(ctx, msg); err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	// Update status to delivered
	deliveredAt := time.Now()
	err = repo.UpdateMessageStatus(ctx, msg.ID, MessageStatusDelivered, &deliveredAt)
	if err != nil {
		t.Fatalf("Failed to update message status: %v", err)
	}

	// Verify update
	retrieved, err := repo.GetMessage(ctx, msg.ID)
	if err != nil {
		t.Fatalf("Failed to get message: %v", err)
	}

	if retrieved.Status != MessageStatusDelivered {
		t.Errorf("Status = %v, want delivered", retrieved.Status)
	}
	if retrieved.DeliveredAt == nil {
		t.Error("DeliveredAt should be set")
	}
}

// TestRepository_UpdateMessageAcknowledgment tests updating message acknowledgment
func TestRepository_UpdateMessageAcknowledgment(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()

	msg := &Message{
		FromAgentID: "agent-1",
		ToAgentID:   "agent-2",
		MessageType: MessageTypeTaskRequest,
		Payload:     map[string]interface{}{"task": "test"},
		Status:      MessageStatusDelivered,
		Priority:    5,
		CreatedAt:   now,
	}

	if err := repo.CreateMessage(ctx, msg); err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	// Update acknowledgment
	ackAt := time.Now()
	err = repo.UpdateMessageAcknowledgment(ctx, msg.ID, &ackAt)
	if err != nil {
		t.Fatalf("Failed to update message acknowledgment: %v", err)
	}

	// Verify update
	retrieved, err := repo.GetMessage(ctx, msg.ID)
	if err != nil {
		t.Fatalf("Failed to get message: %v", err)
	}

	if retrieved.AcknowledgedAt == nil {
		t.Error("AcknowledgedAt should be set")
	}
}

// TestRepository_GetMessagesByCorrelation tests retrieving messages by correlation ID
func TestRepository_GetMessagesByCorrelation(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()
	correlationID := "conv-123"

	// Create messages with same correlation ID
	messages := []*Message{
		{
			FromAgentID:   "agent-1",
			ToAgentID:     "agent-2",
			MessageType:   MessageTypeTaskRequest,
			Payload:       map[string]interface{}{"task": "test-1"},
			Status:        MessageStatusPending,
			Priority:      5,
			CreatedAt:     now,
			CorrelationID: correlationID,
		},
		{
			FromAgentID:   "agent-2",
			ToAgentID:     "agent-1",
			MessageType:   MessageTypeResponse,
			Payload:       map[string]interface{}{"result": "ok"},
			Status:        MessageStatusDelivered,
			Priority:      5,
			CreatedAt:     now.Add(1 * time.Second),
			CorrelationID: correlationID,
		},
		{
			FromAgentID:   "agent-1",
			ToAgentID:     "agent-3",
			MessageType:   MessageTypeTaskRequest,
			Payload:       map[string]interface{}{"task": "test-3"},
			Status:        MessageStatusPending,
			Priority:      5,
			CreatedAt:     now,
			CorrelationID: "conv-456",
		},
	}

	for _, msg := range messages {
		if err := repo.CreateMessage(ctx, msg); err != nil {
			t.Fatalf("Failed to create message: %v", err)
		}
	}

	// Get messages by correlation ID
	related, err := repo.GetMessagesByCorrelation(ctx, correlationID)
	if err != nil {
		t.Fatalf("Failed to get messages by correlation: %v", err)
	}

	if len(related) != 2 {
		t.Errorf("Expected 2 messages with correlation ID %s, got %d", correlationID, len(related))
	}

	// Verify they're ordered by creation time
	if len(related) >= 2 {
		if related[0].CreatedAt.After(related[1].CreatedAt) {
			t.Error("Messages should be ordered by creation time")
		}
	}
}

// TestRepository_DeleteExpiredMessages tests deleting expired messages
func TestRepository_DeleteExpiredMessages(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	// Create messages
	messages := []*Message{
		{
			FromAgentID: "agent-1",
			ToAgentID:   "agent-2",
			MessageType: MessageTypeTaskRequest,
			Payload:     map[string]interface{}{"task": "test-1"},
			Status:      MessageStatusPending,
			Priority:    5,
			CreatedAt:   now,
			ExpiresAt:   &past, // Expired
		},
		{
			FromAgentID: "agent-1",
			ToAgentID:   "agent-2",
			MessageType: MessageTypeTaskRequest,
			Payload:     map[string]interface{}{"task": "test-2"},
			Status:      MessageStatusPending,
			Priority:    5,
			CreatedAt:   now,
			ExpiresAt:   &past, // Expired
		},
		{
			FromAgentID: "agent-1",
			ToAgentID:   "agent-2",
			MessageType: MessageTypeTaskRequest,
			Payload:     map[string]interface{}{"task": "test-3"},
			Status:      MessageStatusPending,
			Priority:    5,
			CreatedAt:   now,
			ExpiresAt:   &future, // Not expired
		},
	}

	for _, msg := range messages {
		if err := repo.CreateMessage(ctx, msg); err != nil {
			t.Fatalf("Failed to create message: %v", err)
		}
	}

	// Delete expired messages
	count, err := repo.DeleteExpiredMessages(ctx)
	if err != nil {
		t.Fatalf("Failed to delete expired messages: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 expired messages deleted, got %d", count)
	}

	// Verify only non-expired message remains
	pending, err := repo.GetPendingMessages(ctx, "agent-2", 100)
	if err != nil {
		t.Fatalf("Failed to get pending messages: %v", err)
	}

	if len(pending) != 1 {
		t.Errorf("Expected 1 pending message remaining, got %d", len(pending))
	}
}

// TestRepository_CreateAndGetPublication tests publication creation and retrieval
func TestRepository_CreateAndGetPublication(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()

	pub := &Publication{
		PublisherAgentID:   "agent-1",
		PublisherAgentType: "worker",
		PublicationType:    PublicationTypeEvent,
		EventName:          "state.changed",
		Payload:            map[string]interface{}{"old": "idle", "new": "active"},
		PublishedAt:        now,
		TTLSeconds:         300,
		ExpiresAt:          now.Add(300 * time.Second),
	}

	// Create publication
	err = repo.CreatePublication(ctx, pub)
	if err != nil {
		t.Fatalf("Failed to create publication: %v", err)
	}

	if pub.ID == "" {
		t.Error("Publication ID should be set after creation")
	}

	// Get publication
	retrieved, err := repo.GetPublication(ctx, pub.ID)
	if err != nil {
		t.Fatalf("Failed to get publication: %v", err)
	}

	if retrieved.ID != pub.ID {
		t.Errorf("ID = %v, want %v", retrieved.ID, pub.ID)
	}
	if retrieved.PublisherAgentID != pub.PublisherAgentID {
		t.Errorf("PublisherAgentID = %v, want %v", retrieved.PublisherAgentID, pub.PublisherAgentID)
	}
	if retrieved.EventName != pub.EventName {
		t.Errorf("EventName = %v, want %v", retrieved.EventName, pub.EventName)
	}
}

// TestRepository_CreateAndGetSubscription tests subscription creation and retrieval
func TestRepository_CreateAndGetSubscription(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()

	sub := &Subscription{
		SubscriberAgentID:   "agent-1",
		SubscriberAgentType: "worker",
		EventPattern:        "state.*",
		PublicationTypes:    []PublicationType{PublicationTypeEvent},
		CreatedAt:           now,
		UpdatedAt:           now,
		Active:              true,
	}

	// Create subscription
	err = repo.CreateSubscription(ctx, sub)
	if err != nil {
		t.Fatalf("Failed to create subscription: %v", err)
	}

	if sub.ID == "" {
		t.Error("Subscription ID should be set after creation")
	}

	// Get subscription
	retrieved, err := repo.GetSubscription(ctx, sub.ID)
	if err != nil {
		t.Fatalf("Failed to get subscription: %v", err)
	}

	if retrieved.ID != sub.ID {
		t.Errorf("ID = %v, want %v", retrieved.ID, sub.ID)
	}
	if retrieved.SubscriberAgentID != sub.SubscriberAgentID {
		t.Errorf("SubscriberAgentID = %v, want %v", retrieved.SubscriberAgentID, sub.SubscriberAgentID)
	}
	if retrieved.EventPattern != sub.EventPattern {
		t.Errorf("EventPattern = %v, want %v", retrieved.EventPattern, sub.EventPattern)
	}
	if !retrieved.Active {
		t.Error("Subscription should be active")
	}
}

// TestRepository_GetActiveSubscriptions tests retrieving active subscriptions
func TestRepository_GetActiveSubscriptions(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()

	// Create subscriptions
	subscriptions := []*Subscription{
		{
			SubscriberAgentID:   "agent-1",
			SubscriberAgentType: "worker",
			EventPattern:        "state.*",
			CreatedAt:           now,
			UpdatedAt:           now,
			Active:              true,
		},
		{
			SubscriberAgentID:   "agent-1",
			SubscriberAgentType: "worker",
			EventPattern:        "task.*",
			CreatedAt:           now,
			UpdatedAt:           now,
			Active:              true,
		},
		{
			SubscriberAgentID:   "agent-1",
			SubscriberAgentType: "worker",
			EventPattern:        "metric.*",
			CreatedAt:           now,
			UpdatedAt:           now,
			Active:              false,
		},
		{
			SubscriberAgentID:   "agent-2",
			SubscriberAgentType: "worker",
			EventPattern:        "state.*",
			CreatedAt:           now,
			UpdatedAt:           now,
			Active:              true,
		},
	}

	for _, sub := range subscriptions {
		if err := repo.CreateSubscription(ctx, sub); err != nil {
			t.Fatalf("Failed to create subscription: %v", err)
		}
	}

	// Get active subscriptions for agent-1
	active, err := repo.GetActiveSubscriptions(ctx, "agent-1")
	if err != nil {
		t.Fatalf("Failed to get active subscriptions: %v", err)
	}

	// Should return 2 subscriptions (first two only)
	if len(active) != 2 {
		t.Errorf("Expected 2 active subscriptions, got %d", len(active))
	}

	for _, sub := range active {
		if sub.SubscriberAgentID != "agent-1" {
			t.Errorf("Subscription has wrong agent ID: %s", sub.SubscriberAgentID)
		}
		if !sub.Active {
			t.Error("Subscription should be active")
		}
	}
}

// TestRepository_DeactivateSubscription tests deactivating a subscription
func TestRepository_DeactivateSubscription(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()

	sub := &Subscription{
		SubscriberAgentID:   "agent-1",
		SubscriberAgentType: "worker",
		EventPattern:        "state.*",
		CreatedAt:           now,
		UpdatedAt:           now,
		Active:              true,
	}

	if err := repo.CreateSubscription(ctx, sub); err != nil {
		t.Fatalf("Failed to create subscription: %v", err)
	}

	// Deactivate subscription
	err = repo.DeactivateSubscription(ctx, sub.ID)
	if err != nil {
		t.Fatalf("Failed to deactivate subscription: %v", err)
	}

	// Verify deactivation
	retrieved, err := repo.GetSubscription(ctx, sub.ID)
	if err != nil {
		t.Fatalf("Failed to get subscription: %v", err)
	}

	if retrieved.Active {
		t.Error("Subscription should be deactivated")
	}
}

// TestRepository_CreateDelivery tests creating a delivery record
func TestRepository_CreateDelivery(t *testing.T) {
	client := skipIfNoDatabase(t)
	if client == nil {
		return
	}

	repo, err := NewRepository(client)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer cleanupTestData(t, repo)

	ctx := context.Background()
	now := time.Now()

	// Create a publication and subscription first
	pub := &Publication{
		PublisherAgentID:   "agent-1",
		PublisherAgentType: "worker",
		PublicationType:    PublicationTypeEvent,
		EventName:          "state.changed",
		Payload:            map[string]interface{}{"test": "data"},
		PublishedAt:        now,
		TTLSeconds:         300,
		ExpiresAt:          now.Add(300 * time.Second),
	}

	if err := repo.CreatePublication(ctx, pub); err != nil {
		t.Fatalf("Failed to create publication: %v", err)
	}

	sub := &Subscription{
		SubscriberAgentID:   "agent-2",
		SubscriberAgentType: "worker",
		EventPattern:        "state.*",
		CreatedAt:           now,
		UpdatedAt:           now,
		Active:              true,
	}

	if err := repo.CreateSubscription(ctx, sub); err != nil {
		t.Fatalf("Failed to create subscription: %v", err)
	}

	// Create delivery
	delivery := &PublicationDelivery{
		From:             CollectionPublications + "/" + pub.ID,
		To:               "agents/" + sub.SubscriberAgentID, // Assuming agents collection
		SubscriptionID:   sub.ID,
		DeliveredAt:      now,
		Acknowledged:     false,
		Processed:        false,
		ProcessingResult: "pending",
	}

	err = repo.CreateDelivery(ctx, delivery)
	if err != nil {
		t.Fatalf("Failed to create delivery: %v", err)
	}

	if delivery.ID == "" {
		t.Error("Delivery ID should be set after creation")
	}
}
