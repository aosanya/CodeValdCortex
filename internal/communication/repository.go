package communication

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/database"
	driver "github.com/arangodb/go-driver"
	log "github.com/sirupsen/logrus"
)

const (
	// CollectionMessages is the messages collection name
	CollectionMessages = "agent_messages"
	// CollectionPublications is the publications collection name
	CollectionPublications = "agent_publications"
	// CollectionSubscriptions is the subscriptions collection name
	CollectionSubscriptions = "agent_subscriptions"
	// CollectionDeliveries is the deliveries collection name (edge)
	CollectionDeliveries = "agent_publication_deliveries"
)

// Repository handles communication persistence in ArangoDB
type Repository struct {
	db               *database.ArangoClient
	messagesCol      driver.Collection
	publicationsCol  driver.Collection
	subscriptionsCol driver.Collection
	deliveriesCol    driver.Collection
}

// NewRepository creates a new communication repository
func NewRepository(dbClient *database.ArangoClient) (*Repository, error) {
	ctx := dbClient.Context()
	db := dbClient.Database()

	// Ensure collections exist
	messagesCol, err := ensureCollection(ctx, db, CollectionMessages, false)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure messages collection: %w", err)
	}

	publicationsCol, err := ensureCollection(ctx, db, CollectionPublications, false)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure publications collection: %w", err)
	}

	subscriptionsCol, err := ensureCollection(ctx, db, CollectionSubscriptions, false)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure subscriptions collection: %w", err)
	}

	deliveriesCol, err := ensureCollection(ctx, db, CollectionDeliveries, true)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure deliveries collection: %w", err)
	}

	// Create indexes
	if err := createIndexes(ctx, messagesCol, publicationsCol, subscriptionsCol); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	log.Info("Communication repository initialized successfully")

	return &Repository{
		db:               dbClient,
		messagesCol:      messagesCol,
		publicationsCol:  publicationsCol,
		subscriptionsCol: subscriptionsCol,
		deliveriesCol:    deliveriesCol,
	}, nil
}

// ensureCollection creates a collection if it doesn't exist
func ensureCollection(ctx context.Context, db driver.Database, name string, isEdge bool) (driver.Collection, error) {
	exists, err := db.CollectionExists(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}

	if exists {
		col, err := db.Collection(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("failed to open collection: %w", err)
		}
		log.WithField("collection", name).Info("Using existing collection")
		return col, nil
	}

	// Create collection
	options := &driver.CreateCollectionOptions{}
	if isEdge {
		options.Type = driver.CollectionTypeEdge
	}

	col, err := db.CreateCollection(ctx, name, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	log.WithField("collection", name).Info("Created new collection")
	return col, nil
}

// createIndexes creates all required indexes for communication collections
func createIndexes(ctx context.Context, messages, publications, subscriptions driver.Collection) error {
	// Messages indexes
	messageIndexes := []struct {
		name   string
		fields []string
		unique bool
	}{
		{"idx_messages_recipient", []string{"to_agent_id", "status", "created_at"}, false},
		{"idx_messages_priority", []string{"to_agent_id", "priority", "created_at"}, false},
		{"idx_messages_expiration", []string{"expires_at"}, false},
		{"idx_messages_correlation", []string{"correlation_id"}, false},
	}

	for _, idx := range messageIndexes {
		_, _, err := messages.EnsurePersistentIndex(ctx, idx.fields, &driver.EnsurePersistentIndexOptions{
			Name:   idx.name,
			Unique: idx.unique,
			Sparse: idx.name == "idx_messages_correlation", // Sparse for optional fields
		})
		if err != nil {
			return fmt.Errorf("failed to create index %s: %w", idx.name, err)
		}
	}

	// Publications indexes
	publicationIndexes := []struct {
		name   string
		fields []string
	}{
		{"idx_publications_publisher", []string{"publisher_agent_id", "published_at"}},
		{"idx_publications_event", []string{"event_name", "published_at"}},
		{"idx_publications_type", []string{"publication_type", "published_at"}},
		{"idx_publications_expiration", []string{"expires_at"}},
	}

	for _, idx := range publicationIndexes {
		_, _, err := publications.EnsurePersistentIndex(ctx, idx.fields, &driver.EnsurePersistentIndexOptions{
			Name: idx.name,
		})
		if err != nil {
			return fmt.Errorf("failed to create index %s: %w", idx.name, err)
		}
	}

	// Subscriptions indexes
	subscriptionIndexes := []struct {
		name   string
		fields []string
		sparse bool
	}{
		{"idx_subscriptions_subscriber", []string{"subscriber_agent_id", "active"}, false},
		{"idx_subscriptions_publisher", []string{"publisher_agent_id", "active"}, true},
		{"idx_subscriptions_pattern", []string{"event_pattern", "active"}, false},
	}

	for _, idx := range subscriptionIndexes {
		_, _, err := subscriptions.EnsurePersistentIndex(ctx, idx.fields, &driver.EnsurePersistentIndexOptions{
			Name:   idx.name,
			Sparse: idx.sparse,
		})
		if err != nil {
			return fmt.Errorf("failed to create index %s: %w", idx.name, err)
		}
	}

	log.Info("Created all communication indexes successfully")
	return nil
}

// Message operations

// CreateMessage creates a new message in the database
func (r *Repository) CreateMessage(ctx context.Context, msg *Message) error {
	meta, err := r.messagesCol.CreateDocument(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	msg.ID = meta.Key
	msg.Rev = meta.Rev
	return nil
}

// GetMessage retrieves a message by ID
func (r *Repository) GetMessage(ctx context.Context, id string) (*Message, error) {
	var msg Message
	meta, err := r.messagesCol.ReadDocument(ctx, id, &msg)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, fmt.Errorf("message not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	msg.ID = meta.Key
	msg.Rev = meta.Rev
	return &msg, nil
}

// GetPendingMessages retrieves pending messages for an agent
func (r *Repository) GetPendingMessages(ctx context.Context, agentID string, limit int) ([]*Message, error) {
	query := `
		FOR msg IN @@collection
		FILTER msg.to_agent_id == @agentID
		FILTER msg.status == @status
		FILTER msg.expires_at == null OR msg.expires_at > @now
		SORT msg.priority DESC, msg.created_at ASC
		LIMIT @limit
		RETURN msg
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionMessages,
		"agentID":     agentID,
		"status":      MessageStatusPending,
		"now":         time.Now(),
		"limit":       limit,
	}

	cursor, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending messages: %w", err)
	}
	defer cursor.Close()

	var messages []*Message
	for cursor.HasMore() {
		var msg Message
		_, err := cursor.ReadDocument(ctx, &msg)
		if err != nil {
			return nil, fmt.Errorf("failed to read message from cursor: %w", err)
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

// UpdateMessageStatus updates the status of a message
func (r *Repository) UpdateMessageStatus(ctx context.Context, id string, status MessageStatus, deliveredAt *time.Time) error {
	update := map[string]interface{}{
		"status": status,
	}

	if deliveredAt != nil {
		update["delivered_at"] = deliveredAt
	}

	_, err := r.messagesCol.UpdateDocument(ctx, id, update)
	if err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}

	return nil
}

// UpdateMessageAcknowledgment marks a message as acknowledged
func (r *Repository) UpdateMessageAcknowledgment(ctx context.Context, id string, acknowledgedAt *time.Time) error {
	update := map[string]interface{}{
		"acknowledged_at": acknowledgedAt,
	}

	_, err := r.messagesCol.UpdateDocument(ctx, id, update)
	if err != nil {
		return fmt.Errorf("failed to update message acknowledgment: %w", err)
	}

	return nil
}

// GetMessagesByCorrelation retrieves messages by correlation ID
func (r *Repository) GetMessagesByCorrelation(ctx context.Context, correlationID string) ([]*Message, error) {
	query := `
		FOR msg IN @@collection
		FILTER msg.correlation_id == @correlationID
		SORT msg.created_at ASC
		RETURN msg
	`

	bindVars := map[string]interface{}{
		"@collection":   CollectionMessages,
		"correlationID": correlationID,
	}

	cursor, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages by correlation: %w", err)
	}
	defer cursor.Close()

	var messages []*Message
	for cursor.HasMore() {
		var msg Message
		_, err := cursor.ReadDocument(ctx, &msg)
		if err != nil {
			return nil, fmt.Errorf("failed to read message from cursor: %w", err)
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

// DeleteExpiredMessages deletes messages that have expired
func (r *Repository) DeleteExpiredMessages(ctx context.Context) (int, error) {
	query := `
		FOR msg IN @@collection
		FILTER msg.expires_at != null AND msg.expires_at < @now
		REMOVE msg IN @@collection
		RETURN OLD
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionMessages,
		"now":         time.Now(),
	}

	cursor, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired messages: %w", err)
	}
	defer cursor.Close()

	count := 0
	for cursor.HasMore() {
		cursor.ReadDocument(ctx, &Message{})
		count++
	}

	return count, nil
}

// Publication operations

// CreatePublication creates a new publication
func (r *Repository) CreatePublication(ctx context.Context, pub *Publication) error {
	meta, err := r.publicationsCol.CreateDocument(ctx, pub)
	if err != nil {
		return fmt.Errorf("failed to create publication: %w", err)
	}

	pub.ID = meta.Key
	pub.Rev = meta.Rev
	return nil
}

// GetPublication retrieves a publication by ID
func (r *Repository) GetPublication(ctx context.Context, id string) (*Publication, error) {
	var pub Publication
	meta, err := r.publicationsCol.ReadDocument(ctx, id, &pub)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, fmt.Errorf("publication not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read publication: %w", err)
	}

	pub.ID = meta.Key
	pub.Rev = meta.Rev
	return &pub, nil
}

// GetMatchingPublications retrieves publications matching subscriptions
func (r *Repository) GetMatchingPublications(ctx context.Context, subscriptions []*Subscription, since time.Time) ([]*Publication, error) {
	if len(subscriptions) == 0 {
		return []*Publication{}, nil
	}

	// Build filter for event patterns
	patterns := make([]string, 0, len(subscriptions))
	for _, sub := range subscriptions {
		patterns = append(patterns, sub.EventPattern)
	}

	query := `
		FOR pub IN @@collection
		FILTER pub.published_at > @since
		FILTER pub.expires_at > @now
		FILTER REGEX_TEST(pub.event_name, @pattern)
		SORT pub.published_at DESC
		RETURN pub
	`

	// For simplicity, we'll query with a combined pattern
	// In production, this should be more sophisticated with proper pattern matching
	combinedPattern := ".*" // Accept all for now, filtering done in matcher

	bindVars := map[string]interface{}{
		"@collection": CollectionPublications,
		"since":       since,
		"now":         time.Now(),
		"pattern":     combinedPattern,
	}

	cursor, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query matching publications: %w", err)
	}
	defer cursor.Close()

	var publications []*Publication
	for cursor.HasMore() {
		var pub Publication
		_, err := cursor.ReadDocument(ctx, &pub)
		if err != nil {
			return nil, fmt.Errorf("failed to read publication from cursor: %w", err)
		}
		publications = append(publications, &pub)
	}

	return publications, nil
}

// DeleteExpiredPublications deletes publications that have expired
func (r *Repository) DeleteExpiredPublications(ctx context.Context) (int, error) {
	query := `
		FOR pub IN @@collection
		FILTER pub.expires_at < @now
		REMOVE pub IN @@collection
		RETURN OLD
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionPublications,
		"now":         time.Now(),
	}

	cursor, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired publications: %w", err)
	}
	defer cursor.Close()

	count := 0
	for cursor.HasMore() {
		cursor.ReadDocument(ctx, &Publication{})
		count++
	}

	return count, nil
}

// Subscription operations

// CreateSubscription creates a new subscription
func (r *Repository) CreateSubscription(ctx context.Context, sub *Subscription) error {
	meta, err := r.subscriptionsCol.CreateDocument(ctx, sub)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	sub.ID = meta.Key
	sub.Rev = meta.Rev
	return nil
}

// GetSubscription retrieves a subscription by ID
func (r *Repository) GetSubscription(ctx context.Context, id string) (*Subscription, error) {
	var sub Subscription
	meta, err := r.subscriptionsCol.ReadDocument(ctx, id, &sub)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, fmt.Errorf("subscription not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read subscription: %w", err)
	}

	sub.ID = meta.Key
	sub.Rev = meta.Rev
	return &sub, nil
}

// GetActiveSubscriptions retrieves all active subscriptions for an agent
func (r *Repository) GetActiveSubscriptions(ctx context.Context, agentID string) ([]*Subscription, error) {
	query := `
		FOR sub IN @@collection
		FILTER sub.subscriber_agent_id == @agentID
		FILTER sub.active == true
		RETURN sub
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionSubscriptions,
		"agentID":     agentID,
	}

	cursor, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query active subscriptions: %w", err)
	}
	defer cursor.Close()

	var subscriptions []*Subscription
	for cursor.HasMore() {
		var sub Subscription
		_, err := cursor.ReadDocument(ctx, &sub)
		if err != nil {
			return nil, fmt.Errorf("failed to read subscription from cursor: %w", err)
		}
		subscriptions = append(subscriptions, &sub)
	}

	return subscriptions, nil
}

// DeactivateSubscription deactivates a subscription
func (r *Repository) DeactivateSubscription(ctx context.Context, id string) error {
	update := map[string]interface{}{
		"active":     false,
		"updated_at": time.Now(),
	}

	_, err := r.subscriptionsCol.UpdateDocument(ctx, id, update)
	if err != nil {
		return fmt.Errorf("failed to deactivate subscription: %w", err)
	}

	return nil
}

// UpdateSubscriptionLastMatched updates the last matched timestamp
func (r *Repository) UpdateSubscriptionLastMatched(ctx context.Context, id string, matchedAt time.Time) error {
	update := map[string]interface{}{
		"last_matched_at": matchedAt,
		"updated_at":      time.Now(),
	}

	_, err := r.subscriptionsCol.UpdateDocument(ctx, id, update)
	if err != nil {
		return fmt.Errorf("failed to update subscription last matched: %w", err)
	}

	return nil
}

// DeleteSubscription deletes a subscription
func (r *Repository) DeleteSubscription(ctx context.Context, id string) error {
	_, err := r.subscriptionsCol.RemoveDocument(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	return nil
}

// Delivery tracking operations

// CreateDelivery creates a delivery record
func (r *Repository) CreateDelivery(ctx context.Context, delivery *PublicationDelivery) error {
	meta, err := r.deliveriesCol.CreateDocument(ctx, delivery)
	if err != nil {
		return fmt.Errorf("failed to create delivery: %w", err)
	}

	delivery.ID = meta.Key
	delivery.Rev = meta.Rev
	return nil
}
