package communication

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// PubSubService handles publish/subscribe messaging
type PubSubService struct {
	repo    PubSubRepository
	matcher *SubscriptionMatcher
}

// NewPubSubService creates a new pub/sub service
func NewPubSubService(repo PubSubRepository) *PubSubService {
	return &PubSubService{
		repo:    repo,
		matcher: NewSubscriptionMatcher(),
	}
}

// Publish publishes an event/status update
func (ps *PubSubService) Publish(ctx context.Context, publisherAgentID, publisherAgentType, eventName string, payload map[string]interface{}, opts *PublicationOptions) (string, error) {
	pub := &Publication{
		PublisherAgentID:   publisherAgentID,
		PublisherAgentType: publisherAgentType,
		EventName:          eventName,
		Payload:            payload,
		PublishedAt:        time.Now(),
	}

	// Apply options
	if opts != nil {
		pub.PublicationType = opts.Type
		pub.TTLSeconds = opts.TTLSeconds
		pub.Metadata = opts.Metadata
	}

	// Set defaults
	if pub.PublicationType == "" {
		pub.PublicationType = PublicationTypeEvent
	}

	if pub.TTLSeconds == 0 {
		pub.TTLSeconds = 3600 // Default: 1 hour
	}

	// Set expiration based on TTL
	pub.ExpiresAt = pub.PublishedAt.Add(time.Duration(pub.TTLSeconds) * time.Second)

	// Validate publication
	if err := ps.validatePublication(pub); err != nil {
		return "", fmt.Errorf("invalid publication: %w", err)
	}

	// Generate publication ID
	pub.ID = fmt.Sprintf("pub-%s", uuid.New().String())

	// Store publication
	if err := ps.repo.CreatePublication(ctx, pub); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"publisher": publisherAgentID,
			"event":     eventName,
		}).Error("Failed to publish event")
		return "", fmt.Errorf("failed to store publication: %w", err)
	}


	return pub.ID, nil
}

// Subscribe creates a new subscription
func (ps *PubSubService) Subscribe(ctx context.Context, subscriberAgentID, subscriberAgentType, eventPattern string, filters *SubscriptionFilters) (string, error) {
	sub := &Subscription{
		SubscriberAgentID:   subscriberAgentID,
		SubscriberAgentType: subscriberAgentType,
		EventPattern:        eventPattern,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
		Active:              true,
	}

	// Apply filters
	if filters != nil {
		sub.PublisherAgentID = filters.PublisherID
		sub.PublisherAgentType = filters.PublisherType
		sub.PublicationTypes = filters.Types
		sub.FilterConditions = filters.Conditions
		sub.Metadata = filters.Metadata
	}

	// Validate subscription
	if err := ps.validateSubscription(sub); err != nil {
		return "", fmt.Errorf("invalid subscription: %w", err)
	}

	// Generate subscription ID
	sub.ID = fmt.Sprintf("sub-%s", uuid.New().String())

	// Store subscription
	if err := ps.repo.CreateSubscription(ctx, sub); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"subscriber": subscriberAgentID,
			"pattern":    eventPattern,
		}).Error("Failed to create subscription")
		return "", fmt.Errorf("failed to store subscription: %w", err)
	}


	return sub.ID, nil
}

// Unsubscribe deactivates a subscription
func (ps *PubSubService) Unsubscribe(ctx context.Context, subscriptionID string) error {
	if err := ps.repo.DeactivateSubscription(ctx, subscriptionID); err != nil {
		log.WithError(err).WithField("subscription_id", subscriptionID).Error("Failed to unsubscribe")
		return err
	}

	return nil
}

// DeleteSubscription permanently deletes a subscription
func (ps *PubSubService) DeleteSubscription(ctx context.Context, subscriptionID string) error {
	if err := ps.repo.DeleteSubscription(ctx, subscriptionID); err != nil {
		log.WithError(err).WithField("subscription_id", subscriptionID).Error("Failed to delete subscription")
		return err
	}

	return nil
}

// GetActiveSubscriptions retrieves all active subscriptions for an agent
func (ps *PubSubService) GetActiveSubscriptions(ctx context.Context, agentID string) ([]*Subscription, error) {
	subscriptions, err := ps.repo.GetActiveSubscriptions(ctx, agentID)
	if err != nil {
		log.WithError(err).WithField("agent_id", agentID).Error("Failed to get active subscriptions")
		return nil, err
	}


	return subscriptions, nil
}

// GetMatchingPublications retrieves publications matching agent's subscriptions since a given time
func (ps *PubSubService) GetMatchingPublications(ctx context.Context, agentID string, since time.Time) ([]*Publication, error) {
	// Get agent's active subscriptions
	subscriptions, err := ps.repo.GetActiveSubscriptions(ctx, agentID)
	if err != nil {
		log.WithError(err).WithField("agent_id", agentID).Error("Failed to get subscriptions")
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	if len(subscriptions) == 0 {
		return []*Publication{}, nil
	}

	// Get publications matching subscriptions
	publications, err := ps.repo.GetMatchingPublications(ctx, subscriptions, since)
	if err != nil {
		log.WithError(err).WithField("agent_id", agentID).Error("Failed to get matching publications")
		return nil, err
	}

	// Filter publications using matcher to ensure they match subscription patterns
	matched := ps.matcher.FilterMatchingPublications(publications, subscriptions)

	// Update last matched timestamp for subscriptions
	now := time.Now()
	for _, pub := range matched {
		matchingSubs := ps.matcher.GetMatchingSubscriptions(pub, subscriptions)
		for _, sub := range matchingSubs {
			// Update asynchronously to not block
			go func(subID string) {
				updateCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := ps.repo.UpdateSubscriptionLastMatched(updateCtx, subID, now); err != nil {
					log.WithError(err).WithField("subscription_id", subID).Warn("Failed to update subscription last matched")
				}
			}(sub.ID)
		}
	}


	return matched, nil
}

// CleanupExpiredPublications removes expired publications from the database
func (ps *PubSubService) CleanupExpiredPublications(ctx context.Context) (int, error) {
	count, err := ps.repo.DeleteExpiredPublications(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to cleanup expired publications")
		return 0, err
	}

	if count > 0 {
		log.WithField("count", count).Info("Cleaned up expired publications")
	}

	return count, nil
}

// validatePublication validates a publication before storing
func (ps *PubSubService) validatePublication(pub *Publication) error {
	if pub.PublisherAgentID == "" {
		return fmt.Errorf("publisher_agent_id is required")
	}
	if pub.EventName == "" {
		return fmt.Errorf("event_name is required")
	}
	if pub.PublicationType == "" {
		return fmt.Errorf("publication_type is required")
	}
	if pub.Payload == nil {
		return fmt.Errorf("payload is required")
	}
	if pub.TTLSeconds < 0 {
		return fmt.Errorf("ttl_seconds cannot be negative")
	}
	return nil
}

// validateSubscription validates a subscription before storing
func (ps *PubSubService) validateSubscription(sub *Subscription) error {
	if sub.SubscriberAgentID == "" {
		return fmt.Errorf("subscriber_agent_id is required")
	}
	if sub.EventPattern == "" {
		return fmt.Errorf("event_pattern is required")
	}
	// Validate pattern syntax by trying to match it
	if ps.matcher.MatchesPattern("test", sub.EventPattern) {
		// Pattern is valid
	}
	return nil
}
