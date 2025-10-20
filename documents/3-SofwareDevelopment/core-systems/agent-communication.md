# Agent Communication System - Design Documentation

## Overview

The Agent Communication System provides database-driven message passing and publish/subscribe capabilities for inter-agent communication using ArangoDB as the central coordination layer. This design prioritizes persistence, auditability, and scalability over in-memory performance.

## 1. Architecture Principles

### 1.1 Database-Driven Communication
- **Persistent Storage**: All messages stored in ArangoDB for durability and audit trails
- **Asynchronous Delivery**: Polling-based message retrieval with configurable intervals
- **Scalable Design**: Database handles message routing and delivery tracking
- **Multi-Pattern Support**: Direct messaging and publish/subscribe patterns

### 1.2 Core Communication Patterns

#### Pattern 1: Direct Messaging (Point-to-Point)
```
Agent A → [ArangoDB] → Agent B

Flow:
1. Agent A writes message to `agent_messages` collection
2. Agent B polls for messages addressed to it
3. Agent B processes message and marks as delivered
4. Optional: Agent B sends response message back to Agent A
```

#### Pattern 2: Publish/Subscribe (Broadcast)
```
Agent A (Publisher) → [ArangoDB] → Multiple Subscribers

Flow:
1. Agent A publishes event to `agent_publications` collection
2. System matches event against `agent_subscriptions`
3. Subscriber agents poll for matching publications
4. Subscribers process events independently
5. System tracks which agents have consumed each publication
```

## 2. Database Schema Design

### 2.1 Collections

#### `agent_messages` Collection (Document)
Direct agent-to-agent messages with delivery tracking.

```javascript
{
  "_key": "msg-550e8400-e29b-41d4-a716-446655440000",
  "_id": "agent_messages/msg-550e8400-e29b-41d4-a716-446655440000",
  "_rev": "_fbGJHte---",
  
  // Message routing
  "from_agent_id": "agent-123",
  "to_agent_id": "agent-456",
  
  // Message content
  "message_type": "task_request",  // task_request, data_share, command, response, notification
  "payload": {
    "task_id": "task-789",
    "action": "process_data",
    "parameters": {
      "dataset": "customers_q4_2025"
    }
  },
  
  // Delivery tracking
  "status": "pending",  // pending, delivered, failed, expired
  "priority": 5,        // 1-10, higher = more important
  "created_at": "2025-10-20T10:00:00.000Z",
  "delivered_at": null,
  "acknowledged_at": null,
  "expires_at": "2025-10-20T11:00:00.000Z",
  
  // Optional metadata
  "correlation_id": "conv-abc123",  // For request/response correlation
  "reply_to": null,                  // For response routing
  "metadata": {
    "source_system": "orchestrator",
    "trace_id": "trace-xyz789"
  }
}
```

**Indexes**:
```javascript
// Fast retrieval for specific agent
db._ensureIndex({
  type: "persistent",
  fields: ["to_agent_id", "status", "created_at"],
  name: "idx_messages_recipient"
});

// Priority-based retrieval
db._ensureIndex({
  type: "persistent",
  fields: ["to_agent_id", "priority", "created_at"],
  name: "idx_messages_priority"
});

// Cleanup expired messages
db._ensureIndex({
  type: "persistent",
  fields: ["expires_at"],
  name: "idx_messages_expiration"
});

// Correlation tracking
db._ensureIndex({
  type: "persistent",
  fields: ["correlation_id"],
  name: "idx_messages_correlation",
  sparse: true
});
```

#### `agent_publications` Collection (Document)
Broadcast events and status updates from publishing agents.

```javascript
{
  "_key": "pub-660e8400-e29b-41d4-a716-446655440001",
  "_id": "agent_publications/pub-660e8400-e29b-41d4-a716-446655440001",
  "_rev": "_fbGJHte---",
  
  // Publisher information
  "publisher_agent_id": "agent-123",
  "publisher_agent_type": "data-processor",
  
  // Publication details
  "publication_type": "status_change",  // status_change, event, metric, alert, broadcast
  "event_name": "state.changed",        // Hierarchical event naming: category.action
  
  // Event payload
  "payload": {
    "old_state": "running",
    "new_state": "paused",
    "reason": "manual_intervention",
    "timestamp": "2025-10-20T10:00:00.000Z"
  },
  
  // Publication metadata
  "published_at": "2025-10-20T10:00:00.000Z",
  "ttl_seconds": 3600,                  // Time to live for this publication
  "expires_at": "2025-10-20T11:00:00.000Z",
  
  // Optional metadata
  "metadata": {
    "severity": "info",
    "source": "lifecycle_manager",
    "tags": ["manual", "operational"]
  }
}
```

**Indexes**:
```javascript
// Publisher-based retrieval
db._ensureIndex({
  type: "persistent",
  fields: ["publisher_agent_id", "published_at"],
  name: "idx_publications_publisher"
});

// Event name pattern matching
db._ensureIndex({
  type: "persistent",
  fields: ["event_name", "published_at"],
  name: "idx_publications_event"
});

// Type-based filtering
db._ensureIndex({
  type: "persistent",
  fields: ["publication_type", "published_at"],
  name: "idx_publications_type"
});

// Expiration cleanup
db._ensureIndex({
  type: "persistent",
  fields: ["expires_at"],
  name: "idx_publications_expiration"
});
```

#### `agent_subscriptions` Collection (Document)
Agent subscription registrations with event filtering.

```javascript
{
  "_key": "sub-770e8400-e29b-41d4-a716-446655440002",
  "_id": "agent_subscriptions/sub-770e8400-e29b-41d4-a716-446655440002",
  "_rev": "_fbGJHte---",
  
  // Subscriber information
  "subscriber_agent_id": "agent-456",
  "subscriber_agent_type": "coordinator",
  
  // Subscription filters
  "publisher_agent_id": "agent-123",    // null or "*" for all agents
  "publisher_agent_type": null,         // null for all types
  "event_pattern": "state.*",           // Glob pattern: state.*, task.completed, *, etc.
  "publication_types": ["status_change", "event"],  // null for all types
  
  // Optional filtering conditions (AQL-compatible)
  "filter_conditions": {
    "severity": ["warning", "error"],
    "tags": {"$contains": "critical"}
  },
  
  // Subscription metadata
  "created_at": "2025-10-20T09:00:00.000Z",
  "updated_at": "2025-10-20T09:00:00.000Z",
  "active": true,
  "last_matched_at": null,
  
  // Optional metadata
  "metadata": {
    "purpose": "monitor_upstream_agent_health",
    "created_by": "orchestrator"
  }
}
```

**Indexes**:
```javascript
// Fast subscriber lookup
db._ensureIndex({
  type: "persistent",
  fields: ["subscriber_agent_id", "active"],
  name: "idx_subscriptions_subscriber"
});

// Publisher-based matching
db._ensureIndex({
  type: "persistent",
  fields: ["publisher_agent_id", "active"],
  name: "idx_subscriptions_publisher",
  sparse: true
});

// Pattern matching optimization
db._ensureIndex({
  type: "persistent",
  fields: ["event_pattern", "active"],
  name: "idx_subscriptions_pattern"
});
```

#### `agent_publication_deliveries` Collection (Edge - Optional)
Tracks which subscribers have consumed which publications.

```javascript
{
  "_key": "del-880e8400-e29b-41d4-a716-446655440003",
  "_from": "agent_publications/pub-660e8400-e29b-41d4-a716-446655440001",
  "_to": "agents/agent-456",
  "_rev": "_fbGJHte---",
  
  "subscription_id": "sub-770e8400-e29b-41d4-a716-446655440002",
  "delivered_at": "2025-10-20T10:00:05.000Z",
  "acknowledged": true,
  "processed": true,
  "processing_result": "success",  // success, failed, skipped
  
  "metadata": {
    "processing_time_ms": 150,
    "error_message": null
  }
}
```

## 3. Go Implementation Structure

### 3.1 Package Organization

```
internal/
├── communication/
│   ├── types.go              # Message, Publication, Subscription types
│   ├── repository.go         # Database operations
│   ├── message_service.go    # Direct messaging logic
│   ├── pubsub_service.go     # Publish/subscribe logic
│   ├── poller.go             # Message/publication polling
│   ├── matcher.go            # Subscription pattern matching
│   └── cleaner.go            # Expired message cleanup
└── agent/
    └── communicator.go       # Agent-level communication interface
```

### 3.2 Core Types

```go
package communication

import (
    "time"
)

// MessageType defines the type of message
type MessageType string

const (
    MessageTypeTaskRequest  MessageType = "task_request"
    MessageTypeDataShare    MessageType = "data_share"
    MessageTypeCommand      MessageType = "command"
    MessageTypeResponse     MessageType = "response"
    MessageTypeNotification MessageType = "notification"
)

// MessageStatus represents the delivery status
type MessageStatus string

const (
    MessageStatusPending   MessageStatus = "pending"
    MessageStatusDelivered MessageStatus = "delivered"
    MessageStatusFailed    MessageStatus = "failed"
    MessageStatusExpired   MessageStatus = "expired"
)

// Message represents a direct agent-to-agent message
type Message struct {
    ID            string                 `json:"_key,omitempty"`
    FromAgentID   string                 `json:"from_agent_id"`
    ToAgentID     string                 `json:"to_agent_id"`
    MessageType   MessageType            `json:"message_type"`
    Payload       map[string]interface{} `json:"payload"`
    Status        MessageStatus          `json:"status"`
    Priority      int                    `json:"priority"`
    CreatedAt     time.Time              `json:"created_at"`
    DeliveredAt   *time.Time             `json:"delivered_at"`
    AcknowledgedAt *time.Time            `json:"acknowledged_at"`
    ExpiresAt     *time.Time             `json:"expires_at"`
    CorrelationID string                 `json:"correlation_id,omitempty"`
    ReplyTo       string                 `json:"reply_to,omitempty"`
    Metadata      map[string]string      `json:"metadata,omitempty"`
}

// PublicationType defines the type of publication
type PublicationType string

const (
    PublicationTypeStatusChange PublicationType = "status_change"
    PublicationTypeEvent        PublicationType = "event"
    PublicationTypeMetric       PublicationType = "metric"
    PublicationTypeAlert        PublicationType = "alert"
    PublicationTypeBroadcast    PublicationType = "broadcast"
)

// Publication represents a broadcast event or status update
type Publication struct {
    ID                string                 `json:"_key,omitempty"`
    PublisherAgentID  string                 `json:"publisher_agent_id"`
    PublisherAgentType string                `json:"publisher_agent_type"`
    PublicationType   PublicationType        `json:"publication_type"`
    EventName         string                 `json:"event_name"`
    Payload           map[string]interface{} `json:"payload"`
    PublishedAt       time.Time              `json:"published_at"`
    TTLSeconds        int                    `json:"ttl_seconds"`
    ExpiresAt         time.Time              `json:"expires_at"`
    Metadata          map[string]string      `json:"metadata,omitempty"`
}

// Subscription represents an agent's subscription to publications
type Subscription struct {
    ID                 string              `json:"_key,omitempty"`
    SubscriberAgentID  string              `json:"subscriber_agent_id"`
    SubscriberAgentType string             `json:"subscriber_agent_type"`
    PublisherAgentID   *string             `json:"publisher_agent_id"` // nil for all agents
    PublisherAgentType *string             `json:"publisher_agent_type"`
    EventPattern       string              `json:"event_pattern"` // Glob pattern
    PublicationTypes   []PublicationType   `json:"publication_types,omitempty"`
    FilterConditions   map[string]interface{} `json:"filter_conditions,omitempty"`
    CreatedAt          time.Time           `json:"created_at"`
    UpdatedAt          time.Time           `json:"updated_at"`
    Active             bool                `json:"active"`
    LastMatchedAt      *time.Time          `json:"last_matched_at"`
    Metadata           map[string]string   `json:"metadata,omitempty"`
}
```

### 3.3 Message Service

```go
package communication

import (
    "context"
    "fmt"
    "time"
    
    "github.com/google/uuid"
)

// MessageService handles direct agent-to-agent messaging
type MessageService struct {
    repo *Repository
}

// NewMessageService creates a new message service
func NewMessageService(repo *Repository) *MessageService {
    return &MessageService{repo: repo}
}

// SendMessage sends a direct message from one agent to another
func (ms *MessageService) SendMessage(ctx context.Context, msg *Message) (string, error) {
    // Generate message ID if not provided
    if msg.ID == "" {
        msg.ID = fmt.Sprintf("msg-%s", uuid.New().String())
    }
    
    // Set timestamps
    msg.CreatedAt = time.Now()
    msg.Status = MessageStatusPending
    
    // Set expiration if not provided (default: 1 hour)
    if msg.ExpiresAt == nil {
        expiresAt := msg.CreatedAt.Add(1 * time.Hour)
        msg.ExpiresAt = &expiresAt
    }
    
    // Validate message
    if err := ms.validateMessage(msg); err != nil {
        return "", fmt.Errorf("invalid message: %w", err)
    }
    
    // Store message in database
    if err := ms.repo.CreateMessage(ctx, msg); err != nil {
        return "", fmt.Errorf("failed to store message: %w", err)
    }
    
    return msg.ID, nil
}

// GetPendingMessages retrieves pending messages for an agent
func (ms *MessageService) GetPendingMessages(ctx context.Context, agentID string, limit int) ([]*Message, error) {
    return ms.repo.GetPendingMessages(ctx, agentID, limit)
}

// MarkDelivered marks a message as delivered
func (ms *MessageService) MarkDelivered(ctx context.Context, messageID string) error {
    now := time.Now()
    return ms.repo.UpdateMessageStatus(ctx, messageID, MessageStatusDelivered, &now)
}

// MarkAcknowledged marks a message as acknowledged
func (ms *MessageService) MarkAcknowledged(ctx context.Context, messageID string) error {
    now := time.Now()
    return ms.repo.UpdateMessageAcknowledgment(ctx, messageID, &now)
}

// GetMessagesByCorrelation retrieves messages by correlation ID
func (ms *MessageService) GetMessagesByCorrelation(ctx context.Context, correlationID string) ([]*Message, error) {
    return ms.repo.GetMessagesByCorrelation(ctx, correlationID)
}

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
    return nil
}
```

### 3.4 Pub/Sub Service

```go
package communication

import (
    "context"
    "fmt"
    "time"
    
    "github.com/google/uuid"
)

// PubSubService handles publish/subscribe messaging
type PubSubService struct {
    repo    *Repository
    matcher *SubscriptionMatcher
}

// NewPubSubService creates a new pub/sub service
func NewPubSubService(repo *Repository) *PubSubService {
    return &PubSubService{
        repo:    repo,
        matcher: NewSubscriptionMatcher(),
    }
}

// Publish publishes an event/status update
func (ps *PubSubService) Publish(ctx context.Context, pub *Publication) (string, error) {
    // Generate publication ID
    if pub.ID == "" {
        pub.ID = fmt.Sprintf("pub-%s", uuid.New().String())
    }
    
    // Set timestamps
    pub.PublishedAt = time.Now()
    
    // Set expiration based on TTL (default: 1 hour)
    if pub.TTLSeconds == 0 {
        pub.TTLSeconds = 3600
    }
    pub.ExpiresAt = pub.PublishedAt.Add(time.Duration(pub.TTLSeconds) * time.Second)
    
    // Validate publication
    if err := ps.validatePublication(pub); err != nil {
        return "", fmt.Errorf("invalid publication: %w", err)
    }
    
    // Store publication
    if err := ps.repo.CreatePublication(ctx, pub); err != nil {
        return "", fmt.Errorf("failed to store publication: %w", err)
    }
    
    return pub.ID, nil
}

// Subscribe creates a new subscription
func (ps *PubSubService) Subscribe(ctx context.Context, sub *Subscription) (string, error) {
    // Generate subscription ID
    if sub.ID == "" {
        sub.ID = fmt.Sprintf("sub-%s", uuid.New().String())
    }
    
    // Set timestamps
    now := time.Now()
    sub.CreatedAt = now
    sub.UpdatedAt = now
    sub.Active = true
    
    // Validate subscription
    if err := ps.validateSubscription(sub); err != nil {
        return "", fmt.Errorf("invalid subscription: %w", err)
    }
    
    // Store subscription
    if err := ps.repo.CreateSubscription(ctx, sub); err != nil {
        return "", fmt.Errorf("failed to store subscription: %w", err)
    }
    
    return sub.ID, nil
}

// Unsubscribe deactivates a subscription
func (ps *PubSubService) Unsubscribe(ctx context.Context, subscriptionID string) error {
    return ps.repo.DeactivateSubscription(ctx, subscriptionID)
}

// GetMatchingPublications retrieves publications matching agent's subscriptions
func (ps *PubSubService) GetMatchingPublications(ctx context.Context, agentID string, since time.Time) ([]*Publication, error) {
    // Get agent's active subscriptions
    subscriptions, err := ps.repo.GetActiveSubscriptions(ctx, agentID)
    if err != nil {
        return nil, fmt.Errorf("failed to get subscriptions: %w", err)
    }
    
    if len(subscriptions) == 0 {
        return []*Publication{}, nil
    }
    
    // Get publications matching subscriptions
    return ps.repo.GetMatchingPublications(ctx, subscriptions, since)
}

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
    return nil
}

func (ps *PubSubService) validateSubscription(sub *Subscription) error {
    if sub.SubscriberAgentID == "" {
        return fmt.Errorf("subscriber_agent_id is required")
    }
    if sub.EventPattern == "" {
        return fmt.Errorf("event_pattern is required")
    }
    return nil
}
```

## 4. Polling Mechanism

### 4.1 Message Poller

```go
package communication

import (
    "context"
    "time"
    
    log "github.com/sirupsen/logrus"
)

// MessagePoller polls for new messages at regular intervals
type MessagePoller struct {
    agentID       string
    messageService *MessageService
    interval      time.Duration
    handler       MessageHandler
    ctx           context.Context
    cancel        context.CancelFunc
}

// MessageHandler processes received messages
type MessageHandler func(msg *Message) error

// NewMessagePoller creates a new message poller
func NewMessagePoller(agentID string, svc *MessageService, interval time.Duration, handler MessageHandler) *MessagePoller {
    ctx, cancel := context.WithCancel(context.Background())
    return &MessagePoller{
        agentID:       agentID,
        messageService: svc,
        interval:      interval,
        handler:       handler,
        ctx:           ctx,
        cancel:        cancel,
    }
}

// Start begins polling for messages
func (mp *MessagePoller) Start() {
    ticker := time.NewTicker(mp.interval)
    defer ticker.Stop()
    
    log.WithFields(log.Fields{
        "agent_id": mp.agentID,
        "interval": mp.interval,
    }).Info("Starting message poller")
    
    for {
        select {
        case <-ticker.C:
            mp.poll()
        case <-mp.ctx.Done():
            log.WithField("agent_id", mp.agentID).Info("Message poller stopped")
            return
        }
    }
}

// Stop stops the poller
func (mp *MessagePoller) Stop() {
    mp.cancel()
}

func (mp *MessagePoller) poll() {
    messages, err := mp.messageService.GetPendingMessages(mp.ctx, mp.agentID, 100)
    if err != nil {
        log.WithFields(log.Fields{
            "agent_id": mp.agentID,
            "error":    err,
        }).Error("Failed to poll messages")
        return
    }
    
    if len(messages) == 0 {
        return
    }
    
    log.WithFields(log.Fields{
        "agent_id": mp.agentID,
        "count":    len(messages),
    }).Debug("Received messages")
    
    for _, msg := range messages {
        if err := mp.handler(msg); err != nil {
            log.WithFields(log.Fields{
                "agent_id":   mp.agentID,
                "message_id": msg.ID,
                "error":      err,
            }).Error("Failed to handle message")
            continue
        }
        
        // Mark as delivered
        if err := mp.messageService.MarkDelivered(mp.ctx, msg.ID); err != nil {
            log.WithFields(log.Fields{
                "message_id": msg.ID,
                "error":      err,
            }).Error("Failed to mark message as delivered")
        }
    }
}
```

## 5. Performance Considerations

### 5.1 Polling Intervals
- **Default**: 5 seconds for normal priority agents
- **High Priority**: 1-2 seconds for time-critical agents
- **Low Priority**: 10-30 seconds for background agents
- **Adaptive**: Adjust based on message volume

### 5.2 Message Batching
- Retrieve up to 100 messages per poll
- Process in order of priority, then creation time
- Use database indexes for fast retrieval

### 5.3 Cleanup Strategy
- Automatically delete delivered messages after 7 days
- Delete expired messages after TTL
- Archive important messages for audit trail

### 5.4 Optimization Opportunities (Future)
- Implement ArangoDB change streams for push-based delivery
- Add message queue for high-throughput scenarios
- Implement message compression for large payloads
- Add distributed caching layer

## 6. Message Flow Examples

### Example 1: Direct Task Request
```
Agent A (Worker) → Agent B (Coordinator)

1. Agent A sends task completion notification:
   - MessageType: "notification"
   - Payload: {task_id: "task-123", status: "completed"}

2. Message stored in agent_messages collection
   - Status: "pending"
   - Priority: 7

3. Agent B polls every 2 seconds
4. Agent B retrieves pending message
5. Agent B processes notification
6. Message marked as "delivered"
```

### Example 2: Status Change Broadcast
```
Agent A (Data Processor) → All Monitoring Agents

1. Agent A publishes state change:
   - PublicationType: "status_change"
   - EventName: "state.changed"
   - Payload: {old_state: "running", new_state: "paused"}

2. Publication stored in agent_publications collection

3. Monitoring agents have subscriptions:
   - EventPattern: "state.*"
   - PublisherAgentID: "agent-A"

4. Monitoring agents poll for matching publications
5. Each agent processes the state change
6. Optional: Record delivery in agent_publication_deliveries
```

## 7. Error Handling

### 7.1 Message Delivery Failures
- Messages expire after TTL
- Status changed to "expired"
- Can optionally trigger dead letter queue
- Retry mechanism for failed deliveries

### 7.2 Subscription Matching Errors
- Log pattern matching failures
- Continue processing other subscriptions
- Alert on persistent matching errors

## 8. Security Considerations

### 8.1 Message Authentication
- Verify sender agent exists and is active
- Validate agent has permission to send to recipient
- Check message size limits

### 8.2 Access Control
- Agents can only read their own messages
- Subscription filtering prevents unauthorized access
- Audit trail for all message operations

## 9. Testing Strategy

### 9.1 Unit Tests
- Message validation
- Subscription pattern matching
- Repository operations
- Service logic

### 9.2 Integration Tests
- End-to-end message delivery
- Pub/sub subscription matching
- Database operations
- Polling mechanism

### 9.3 Performance Tests
- Message throughput
- Subscription matching scalability
- Database query performance
- Polling overhead

## 10. Metrics and Monitoring

### 10.1 Key Metrics
- Messages sent/received per second
- Message delivery latency
- Publication matching time
- Active subscriptions count
- Expired message count

### 10.2 Health Indicators
- Poller health status
- Database connection status
- Message backlog size
- Subscription processing lag

---

**Last Updated**: 2025-10-20  
**Status**: Design Complete - Ready for Implementation  
**Related Tasks**: MVP-005
