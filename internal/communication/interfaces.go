package communication

import (
	"context"
	"time"
)

// MessageRepository defines the interface for message persistence operations
type MessageRepository interface {
	CreateMessage(ctx context.Context, msg *Message) error
	GetMessage(ctx context.Context, id string) (*Message, error)
	GetPendingMessages(ctx context.Context, agentID string, limit int) ([]*Message, error)
	UpdateMessageStatus(ctx context.Context, id string, status MessageStatus, deliveredAt *time.Time) error
	UpdateMessageAcknowledgment(ctx context.Context, id string, acknowledgedAt *time.Time) error
	GetMessagesByCorrelation(ctx context.Context, correlationID string) ([]*Message, error)
	DeleteExpiredMessages(ctx context.Context) (int, error)
}

// PubSubRepository defines the interface for pub/sub persistence operations
type PubSubRepository interface {
	CreatePublication(ctx context.Context, pub *Publication) error
	GetPublication(ctx context.Context, id string) (*Publication, error)
	GetMatchingPublications(ctx context.Context, subscriptions []*Subscription, since time.Time) ([]*Publication, error)
	DeleteExpiredPublications(ctx context.Context) (int, error)
	CreateSubscription(ctx context.Context, sub *Subscription) error
	GetSubscription(ctx context.Context, id string) (*Subscription, error)
	GetActiveSubscriptions(ctx context.Context, agentID string) ([]*Subscription, error)
	DeactivateSubscription(ctx context.Context, id string) error
	DeleteSubscription(ctx context.Context, id string) error
	CreateDelivery(ctx context.Context, delivery *PublicationDelivery) error
	UpdateSubscriptionLastMatched(ctx context.Context, id string, matchedAt time.Time) error
}

// Ensure Repository implements both interfaces
var _ MessageRepository = (*Repository)(nil)
var _ PubSubRepository = (*Repository)(nil)
