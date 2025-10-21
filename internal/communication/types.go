package communication

import (
	"time"
)

// MessageType defines the type of message
type MessageType string

const (
	// MessageTypeTaskRequest represents a task assignment request
	MessageTypeTaskRequest MessageType = "task_request"
	// MessageTypeDataShare represents data sharing between agents
	MessageTypeDataShare MessageType = "data_share"
	// MessageTypeCommand represents a command to be executed
	MessageTypeCommand MessageType = "command"
	// MessageTypeResponse represents a response to a previous message
	MessageTypeResponse MessageType = "response"
	// MessageTypeNotification represents a notification message
	MessageTypeNotification MessageType = "notification"
)

// MessageStatus represents the delivery status of a message
type MessageStatus string

const (
	// MessageStatusPending indicates message is awaiting delivery
	MessageStatusPending MessageStatus = "pending"
	// MessageStatusDelivered indicates message has been delivered
	MessageStatusDelivered MessageStatus = "delivered"
	// MessageStatusFailed indicates message delivery failed
	MessageStatusFailed MessageStatus = "failed"
	// MessageStatusExpired indicates message has expired
	MessageStatusExpired MessageStatus = "expired"
)

// Message represents a direct agent-to-agent message
type Message struct {
	// ID is the unique message identifier (ArangoDB _key)
	ID string `json:"_key,omitempty"`

	// Rev is the ArangoDB revision
	Rev string `json:"_rev,omitempty"`

	// FromAgentID is the sender agent ID
	FromAgentID string `json:"from_agent_id"`

	// ToAgentID is the recipient agent ID
	ToAgentID string `json:"to_agent_id"`

	// MessageType categorizes the message
	MessageType MessageType `json:"message_type"`

	// Payload contains the message data
	Payload map[string]interface{} `json:"payload"`

	// Status tracks delivery status
	Status MessageStatus `json:"status"`

	// Priority determines delivery order (1-10, higher = more important)
	Priority int `json:"priority"`

	// CreatedAt is when the message was created
	CreatedAt time.Time `json:"created_at"`

	// DeliveredAt is when the message was delivered (nil if not delivered)
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`

	// AcknowledgedAt is when the message was acknowledged (nil if not acknowledged)
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`

	// ExpiresAt is when the message expires (nil for no expiration)
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// CorrelationID links related messages (e.g., request/response pairs)
	CorrelationID string `json:"correlation_id,omitempty"`

	// ReplyTo specifies where responses should be sent
	ReplyTo string `json:"reply_to,omitempty"`

	// Metadata contains additional message metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// PublicationType defines the type of publication
type PublicationType string

const (
	// PublicationTypeStatusChange represents agent status changes
	PublicationTypeStatusChange PublicationType = "status_change"
	// PublicationTypeEvent represents general events
	PublicationTypeEvent PublicationType = "event"
	// PublicationTypeMetric represents metric updates
	PublicationTypeMetric PublicationType = "metric"
	// PublicationTypeAlert represents alerts or warnings
	PublicationTypeAlert PublicationType = "alert"
	// PublicationTypeBroadcast represents general broadcasts
	PublicationTypeBroadcast PublicationType = "broadcast"
)

// Publication represents a broadcast event or status update
type Publication struct {
	// ID is the unique publication identifier (ArangoDB _key)
	ID string `json:"_key,omitempty"`

	// Rev is the ArangoDB revision
	Rev string `json:"_rev,omitempty"`

	// PublisherAgentID is the agent that published this event
	PublisherAgentID string `json:"publisher_agent_id"`

	// PublisherAgentType is the type of the publishing agent
	PublisherAgentType string `json:"publisher_agent_type"`

	// PublicationType categorizes the publication
	PublicationType PublicationType `json:"publication_type"`

	// EventName is a hierarchical event identifier (e.g., "state.changed", "task.completed")
	EventName string `json:"event_name"`

	// Payload contains the event data
	Payload map[string]interface{} `json:"payload"`

	// PublishedAt is when the event was published
	PublishedAt time.Time `json:"published_at"`

	// TTLSeconds defines how long this publication remains valid
	TTLSeconds int `json:"ttl_seconds"`

	// ExpiresAt is when this publication expires
	ExpiresAt time.Time `json:"expires_at"`

	// Metadata contains additional publication metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Subscription represents an agent's subscription to publications
type Subscription struct {
	// ID is the unique subscription identifier (ArangoDB _key)
	ID string `json:"_key,omitempty"`

	// Rev is the ArangoDB revision
	Rev string `json:"_rev,omitempty"`

	// SubscriberAgentID is the agent that created this subscription
	SubscriberAgentID string `json:"subscriber_agent_id"`

	// SubscriberAgentType is the type of the subscribing agent
	SubscriberAgentType string `json:"subscriber_agent_type"`

	// PublisherAgentID filters publications by publisher (nil/empty for all)
	PublisherAgentID *string `json:"publisher_agent_id,omitempty"`

	// PublisherAgentType filters publications by publisher type
	PublisherAgentType *string `json:"publisher_agent_type,omitempty"`

	// EventPattern is a glob pattern for matching event names (e.g., "state.*", "task.completed")
	EventPattern string `json:"event_pattern"`

	// PublicationTypes filters by publication types (nil/empty for all)
	PublicationTypes []PublicationType `json:"publication_types,omitempty"`

	// FilterConditions contains additional filtering rules
	FilterConditions map[string]interface{} `json:"filter_conditions,omitempty"`

	// CreatedAt is when the subscription was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the subscription was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// Active indicates if the subscription is currently active
	Active bool `json:"active"`

	// LastMatchedAt is when a publication last matched this subscription
	LastMatchedAt *time.Time `json:"last_matched_at,omitempty"`

	// Metadata contains additional subscription metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// PublicationDelivery tracks which subscribers have consumed which publications
type PublicationDelivery struct {
	// ID is the unique delivery identifier (ArangoDB _key)
	ID string `json:"_key,omitempty"`

	// From is the publication document ID (ArangoDB _from for edge)
	From string `json:"_from"`

	// To is the agent document ID (ArangoDB _to for edge)
	To string `json:"_to"`

	// Rev is the ArangoDB revision
	Rev string `json:"_rev,omitempty"`

	// SubscriptionID is the subscription that matched
	SubscriptionID string `json:"subscription_id"`

	// DeliveredAt is when the publication was delivered
	DeliveredAt time.Time `json:"delivered_at"`

	// Acknowledged indicates if the subscriber acknowledged receipt
	Acknowledged bool `json:"acknowledged"`

	// Processed indicates if the subscriber processed the publication
	Processed bool `json:"processed"`

	// ProcessingResult indicates the result of processing
	ProcessingResult string `json:"processing_result"` // success, failed, skipped

	// Metadata contains additional delivery metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// MessageOptions contains options for sending messages
type MessageOptions struct {
	// Priority for message delivery (1-10)
	Priority int

	// TTL is the time-to-live in seconds (0 for default)
	TTL int

	// CorrelationID for linking related messages
	CorrelationID string

	// ReplyTo for response routing
	ReplyTo string

	// Metadata for additional context
	Metadata map[string]string
}

// PublicationOptions contains options for publishing events
type PublicationOptions struct {
	// Type of publication
	Type PublicationType

	// TTLSeconds is how long the publication remains valid
	TTLSeconds int

	// Metadata for additional context
	Metadata map[string]string
}

// SubscriptionFilters contains filters for creating subscriptions
type SubscriptionFilters struct {
	// PublisherID filters by specific publisher agent
	PublisherID *string

	// PublisherType filters by publisher agent type
	PublisherType *string

	// Types filters by publication types
	Types []PublicationType

	// Conditions contains custom filter conditions
	Conditions map[string]interface{}

	// Metadata for additional context
	Metadata map[string]string
}
