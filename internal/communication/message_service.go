package communication

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// MessageService handles direct agent-to-agent messaging
type MessageService struct {
	repo MessageRepository
}

// NewMessageService creates a new message service
func NewMessageService(repo MessageRepository) *MessageService {
	return &MessageService{repo: repo}
}

// SendMessage sends a direct message from one agent to another
func (ms *MessageService) SendMessage(ctx context.Context, fromAgentID, toAgentID string, msgType MessageType, payload map[string]interface{}, opts *MessageOptions) (string, error) {
	msg := &Message{
		FromAgentID: fromAgentID,
		ToAgentID:   toAgentID,
		MessageType: msgType,
		Payload:     payload,
		Status:      MessageStatusPending,
		CreatedAt:   time.Now(),
	}

	// Apply options
	if opts != nil {
		msg.Priority = opts.Priority
		msg.CorrelationID = opts.CorrelationID
		msg.ReplyTo = opts.ReplyTo
		msg.Metadata = opts.Metadata

		if opts.TTL > 0 {
			expiresAt := msg.CreatedAt.Add(time.Duration(opts.TTL) * time.Second)
			msg.ExpiresAt = &expiresAt
		}
	}

	// Set default priority if not provided
	if msg.Priority == 0 {
		msg.Priority = 5
	}

	// Set default expiration if not provided (1 hour)
	if msg.ExpiresAt == nil {
		expiresAt := msg.CreatedAt.Add(1 * time.Hour)
		msg.ExpiresAt = &expiresAt
	}

	// Validate message
	if err := ms.validateMessage(msg); err != nil {
		return "", fmt.Errorf("invalid message: %w", err)
	}

	// Generate message ID
	msg.ID = fmt.Sprintf("msg-%s", uuid.New().String())

	// Store message in database
	if err := ms.repo.CreateMessage(ctx, msg); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"from": fromAgentID,
			"to":   toAgentID,
			"type": msgType,
		}).Error("Failed to send message")
		return "", fmt.Errorf("failed to store message: %w", err)
	}

	log.WithFields(log.Fields{
		"message_id": msg.ID,
		"from":       fromAgentID,
		"to":         toAgentID,
		"type":       msgType,
		"priority":   msg.Priority,
	}).Debug("Message sent successfully")

	return msg.ID, nil
}

// GetMessage retrieves a specific message by ID
func (ms *MessageService) GetMessage(ctx context.Context, messageID string) (*Message, error) {
	return ms.repo.GetMessage(ctx, messageID)
}

// GetPendingMessages retrieves pending messages for an agent
func (ms *MessageService) GetPendingMessages(ctx context.Context, agentID string, limit int) ([]*Message, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}

	messages, err := ms.repo.GetPendingMessages(ctx, agentID, limit)
	if err != nil {
		log.WithError(err).WithField("agent_id", agentID).Error("Failed to get pending messages")
		return nil, err
	}

	log.WithFields(log.Fields{
		"agent_id": agentID,
		"count":    len(messages),
	}).Debug("Retrieved pending messages")

	return messages, nil
}

// MarkDelivered marks a message as delivered
func (ms *MessageService) MarkDelivered(ctx context.Context, messageID string) error {
	now := time.Now()
	if err := ms.repo.UpdateMessageStatus(ctx, messageID, MessageStatusDelivered, &now); err != nil {
		log.WithError(err).WithField("message_id", messageID).Error("Failed to mark message as delivered")
		return err
	}

	log.WithField("message_id", messageID).Debug("Message marked as delivered")
	return nil
}

// MarkFailed marks a message as failed
func (ms *MessageService) MarkFailed(ctx context.Context, messageID string) error {
	if err := ms.repo.UpdateMessageStatus(ctx, messageID, MessageStatusFailed, nil); err != nil {
		log.WithError(err).WithField("message_id", messageID).Error("Failed to mark message as failed")
		return err
	}

	log.WithField("message_id", messageID).Warn("Message marked as failed")
	return nil
}

// AcknowledgeMessage marks a message as acknowledged
func (ms *MessageService) AcknowledgeMessage(ctx context.Context, messageID string) error {
	now := time.Now()
	if err := ms.repo.UpdateMessageAcknowledgment(ctx, messageID, &now); err != nil {
		log.WithError(err).WithField("message_id", messageID).Error("Failed to acknowledge message")
		return err
	}

	log.WithField("message_id", messageID).Debug("Message acknowledged")
	return nil
}

// GetConversationHistory retrieves messages by correlation ID (conversation thread)
func (ms *MessageService) GetConversationHistory(ctx context.Context, correlationID string) ([]*Message, error) {
	messages, err := ms.repo.GetMessagesByCorrelation(ctx, correlationID)
	if err != nil {
		log.WithError(err).WithField("correlation_id", correlationID).Error("Failed to get conversation history")
		return nil, err
	}

	log.WithFields(log.Fields{
		"correlation_id": correlationID,
		"count":          len(messages),
	}).Debug("Retrieved conversation history")

	return messages, nil
}

// CleanupExpiredMessages removes expired messages from the database
func (ms *MessageService) CleanupExpiredMessages(ctx context.Context) (int, error) {
	count, err := ms.repo.DeleteExpiredMessages(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to cleanup expired messages")
		return 0, err
	}

	if count > 0 {
		log.WithField("count", count).Info("Cleaned up expired messages")
	}

	return count, nil
}

// validateMessage validates a message before sending
func (ms *MessageService) validateMessage(msg *Message) error {
	if msg.FromAgentID == "" {
		return fmt.Errorf("from_agent_id is required")
	}
	if msg.ToAgentID == "" {
		return fmt.Errorf("to_agent_id is required")
	}
	if msg.MessageType == "" {
		return fmt.Errorf("message_type is required")
	}
	if msg.Priority < 1 || msg.Priority > 10 {
		return fmt.Errorf("priority must be between 1 and 10")
	}
	if msg.Payload == nil {
		return fmt.Errorf("payload is required")
	}
	return nil
}
