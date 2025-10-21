package communication

import (
	"context"
	"testing"
	"time"

	driver "github.com/arangodb/go-driver"
)

// mockMessageRepo is a mock implementation for testing MessageService
type mockMessageRepo struct {
	messages  map[string]*Message
	createErr error
	getErr    error
	updateErr error
	queryErr  error
}

func newMockMessageRepo() *mockMessageRepo {
	return &mockMessageRepo{
		messages: make(map[string]*Message),
	}
}

func (m *mockMessageRepo) CreateMessage(ctx context.Context, msg *Message) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.messages[msg.ID] = msg
	return nil
}

func (m *mockMessageRepo) GetMessage(ctx context.Context, id string) (*Message, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	msg, exists := m.messages[id]
	if !exists {
		return nil, driver.ArangoError{Code: 404}
	}
	return msg, nil
}

func (m *mockMessageRepo) GetPendingMessages(ctx context.Context, agentID string, limit int) ([]*Message, error) {
	if m.queryErr != nil {
		return nil, m.queryErr
	}
	var pending []*Message
	now := time.Now()
	for _, msg := range m.messages {
		if msg.ToAgentID == agentID && msg.Status == MessageStatusPending {
			if msg.ExpiresAt == nil || msg.ExpiresAt.After(now) {
				pending = append(pending, msg)
			}
		}
	}
	return pending, nil
}

func (m *mockMessageRepo) UpdateMessageStatus(ctx context.Context, id string, status MessageStatus, deliveredAt *time.Time) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	msg, exists := m.messages[id]
	if !exists {
		return driver.ArangoError{Code: 404}
	}
	msg.Status = status
	msg.DeliveredAt = deliveredAt
	return nil
}

func (m *mockMessageRepo) UpdateMessageAcknowledgment(ctx context.Context, id string, acknowledgedAt *time.Time) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	msg, exists := m.messages[id]
	if !exists {
		return driver.ArangoError{Code: 404}
	}
	msg.AcknowledgedAt = acknowledgedAt
	return nil
}

func (m *mockMessageRepo) GetMessagesByCorrelation(ctx context.Context, correlationID string) ([]*Message, error) {
	if m.queryErr != nil {
		return nil, m.queryErr
	}
	var messages []*Message
	for _, msg := range m.messages {
		if msg.CorrelationID == correlationID {
			messages = append(messages, msg)
		}
	}
	return messages, nil
}

func (m *mockMessageRepo) DeleteExpiredMessages(ctx context.Context) (int, error) {
	count := 0
	now := time.Now()
	for id, msg := range m.messages {
		if msg.ExpiresAt != nil && msg.ExpiresAt.Before(now) {
			delete(m.messages, id)
			count++
		}
	}
	return count, nil
}

// TestMessageService_SendMessage tests sending messages
func TestMessageService_SendMessage(t *testing.T) {
	repo := newMockMessageRepo()
	svc := NewMessageService(repo)
	ctx := context.Background()

	tests := []struct {
		name        string
		fromAgentID string
		toAgentID   string
		msgType     MessageType
		payload     map[string]interface{}
		opts        *MessageOptions
		expectError bool
		validateMsg func(t *testing.T, msg *Message)
	}{
		{
			name:        "send basic message",
			fromAgentID: "agent-1",
			toAgentID:   "agent-2",
			msgType:     MessageTypeTaskRequest,
			payload:     map[string]interface{}{"task": "test"},
			opts:        nil,
			expectError: false,
			validateMsg: func(t *testing.T, msg *Message) {
				if msg.FromAgentID != "agent-1" {
					t.Errorf("FromAgentID = %v, want agent-1", msg.FromAgentID)
				}
				if msg.ToAgentID != "agent-2" {
					t.Errorf("ToAgentID = %v, want agent-2", msg.ToAgentID)
				}
				if msg.Priority != 5 {
					t.Errorf("Priority = %v, want 5 (default)", msg.Priority)
				}
				if msg.Status != MessageStatusPending {
					t.Errorf("Status = %v, want pending", msg.Status)
				}
			},
		},
		{
			name:        "send with options",
			fromAgentID: "agent-1",
			toAgentID:   "agent-2",
			msgType:     MessageTypeCommand,
			payload:     map[string]interface{}{"action": "restart"},
			opts: &MessageOptions{
				Priority:      8,
				TTL:           60,
				CorrelationID: "conv-123",
				ReplyTo:       "agent-1",
				Metadata:      map[string]string{"source": "test"},
			},
			expectError: false,
			validateMsg: func(t *testing.T, msg *Message) {
				if msg.Priority != 8 {
					t.Errorf("Priority = %v, want 8", msg.Priority)
				}
				if msg.CorrelationID != "conv-123" {
					t.Errorf("CorrelationID = %v, want conv-123", msg.CorrelationID)
				}
				if msg.ReplyTo != "agent-1" {
					t.Errorf("ReplyTo = %v, want agent-1", msg.ReplyTo)
				}
			},
		},
		{
			name:        "invalid priority",
			fromAgentID: "agent-1",
			toAgentID:   "agent-2",
			msgType:     MessageTypeCommand,
			payload:     map[string]interface{}{"action": "test"},
			opts: &MessageOptions{
				Priority: 15, // Invalid
			},
			expectError: true,
		},
		{
			name:        "missing from agent",
			fromAgentID: "",
			toAgentID:   "agent-2",
			msgType:     MessageTypeCommand,
			payload:     map[string]interface{}{"action": "test"},
			expectError: true,
		},
		{
			name:        "missing to agent",
			fromAgentID: "agent-1",
			toAgentID:   "",
			msgType:     MessageTypeCommand,
			payload:     map[string]interface{}{"action": "test"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgID, err := svc.SendMessage(ctx, tt.fromAgentID, tt.toAgentID, tt.msgType, tt.payload, tt.opts)

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

			if msgID == "" {
				t.Error("Expected non-empty message ID")
			}

			// Validate stored message
			msg := repo.messages[msgID]
			if msg == nil {
				t.Fatal("Message not stored in repository")
			}

			if tt.validateMsg != nil {
				tt.validateMsg(t, msg)
			}
		})
	}
}

// TestMessageService_GetPendingMessages tests retrieving pending messages
func TestMessageService_GetPendingMessages(t *testing.T) {
	repo := newMockMessageRepo()
	svc := NewMessageService(repo)
	ctx := context.Background()

	// Create test messages
	now := time.Now()
	future := now.Add(1 * time.Hour)
	past := now.Add(-1 * time.Hour)

	repo.messages["msg-1"] = &Message{
		ID:        "msg-1",
		ToAgentID: "agent-1",
		Status:    MessageStatusPending,
		ExpiresAt: &future,
	}
	repo.messages["msg-2"] = &Message{
		ID:        "msg-2",
		ToAgentID: "agent-1",
		Status:    MessageStatusPending,
		ExpiresAt: &future,
	}
	repo.messages["msg-3"] = &Message{
		ID:        "msg-3",
		ToAgentID: "agent-1",
		Status:    MessageStatusDelivered,
		ExpiresAt: &future,
	}
	repo.messages["msg-4"] = &Message{
		ID:        "msg-4",
		ToAgentID: "agent-2",
		Status:    MessageStatusPending,
		ExpiresAt: &future,
	}
	repo.messages["msg-5"] = &Message{
		ID:        "msg-5",
		ToAgentID: "agent-1",
		Status:    MessageStatusPending,
		ExpiresAt: &past, // Expired
	}

	messages, err := svc.GetPendingMessages(ctx, "agent-1", 100)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should return msg-1 and msg-2 (both pending, not expired, for agent-1)
	if len(messages) != 2 {
		t.Errorf("Expected 2 pending messages, got %d", len(messages))
	}

	msgIDs := make(map[string]bool)
	for _, msg := range messages {
		msgIDs[msg.ID] = true
	}

	if !msgIDs["msg-1"] {
		t.Error("Expected msg-1 in results")
	}
	if !msgIDs["msg-2"] {
		t.Error("Expected msg-2 in results")
	}
	if msgIDs["msg-3"] {
		t.Error("msg-3 should not be in results (already delivered)")
	}
	if msgIDs["msg-4"] {
		t.Error("msg-4 should not be in results (wrong agent)")
	}
	if msgIDs["msg-5"] {
		t.Error("msg-5 should not be in results (expired)")
	}
}

// TestMessageService_MarkDelivered tests marking messages as delivered
func TestMessageService_MarkDelivered(t *testing.T) {
	repo := newMockMessageRepo()
	svc := NewMessageService(repo)
	ctx := context.Background()

	// Create test message
	repo.messages["msg-1"] = &Message{
		ID:     "msg-1",
		Status: MessageStatusPending,
	}

	err := svc.MarkDelivered(ctx, "msg-1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	msg := repo.messages["msg-1"]
	if msg.Status != MessageStatusDelivered {
		t.Errorf("Status = %v, want delivered", msg.Status)
	}
	if msg.DeliveredAt == nil {
		t.Error("DeliveredAt should be set")
	}
}

// TestMessageService_AcknowledgeMessage tests message acknowledgment
func TestMessageService_AcknowledgeMessage(t *testing.T) {
	repo := newMockMessageRepo()
	svc := NewMessageService(repo)
	ctx := context.Background()

	// Create test message
	repo.messages["msg-1"] = &Message{
		ID:     "msg-1",
		Status: MessageStatusDelivered,
	}

	err := svc.AcknowledgeMessage(ctx, "msg-1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	msg := repo.messages["msg-1"]
	if msg.AcknowledgedAt == nil {
		t.Error("AcknowledgedAt should be set")
	}
}

// TestMessageService_GetConversationHistory tests retrieving conversation history
func TestMessageService_GetConversationHistory(t *testing.T) {
	repo := newMockMessageRepo()
	svc := NewMessageService(repo)
	ctx := context.Background()

	// Create test messages with correlation ID
	repo.messages["msg-1"] = &Message{
		ID:            "msg-1",
		CorrelationID: "conv-123",
		MessageType:   MessageTypeTaskRequest,
	}
	repo.messages["msg-2"] = &Message{
		ID:            "msg-2",
		CorrelationID: "conv-123",
		MessageType:   MessageTypeResponse,
	}
	repo.messages["msg-3"] = &Message{
		ID:            "msg-3",
		CorrelationID: "conv-456",
		MessageType:   MessageTypeTaskRequest,
	}

	messages, err := svc.GetConversationHistory(ctx, "conv-123")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	msgIDs := make(map[string]bool)
	for _, msg := range messages {
		msgIDs[msg.ID] = true
	}

	if !msgIDs["msg-1"] {
		t.Error("Expected msg-1 in conversation")
	}
	if !msgIDs["msg-2"] {
		t.Error("Expected msg-2 in conversation")
	}
	if msgIDs["msg-3"] {
		t.Error("msg-3 should not be in conversation")
	}
}

// TestMessageService_CleanupExpiredMessages tests cleanup of expired messages
func TestMessageService_CleanupExpiredMessages(t *testing.T) {
	repo := newMockMessageRepo()
	svc := NewMessageService(repo)
	ctx := context.Background()

	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	// Create test messages
	repo.messages["msg-1"] = &Message{
		ID:        "msg-1",
		ExpiresAt: &past,
	}
	repo.messages["msg-2"] = &Message{
		ID:        "msg-2",
		ExpiresAt: &past,
	}
	repo.messages["msg-3"] = &Message{
		ID:        "msg-3",
		ExpiresAt: &future,
	}

	count, err := svc.CleanupExpiredMessages(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 messages cleaned up, got %d", count)
	}

	if _, exists := repo.messages["msg-1"]; exists {
		t.Error("msg-1 should be deleted")
	}
	if _, exists := repo.messages["msg-2"]; exists {
		t.Error("msg-2 should be deleted")
	}
	if _, exists := repo.messages["msg-3"]; !exists {
		t.Error("msg-3 should not be deleted")
	}
}
