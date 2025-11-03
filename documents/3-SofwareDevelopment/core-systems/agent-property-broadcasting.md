# Agent Property Auto-Broadcasting System

**Version**: 1.0.0  
**Status**: Proposed  
**Date**: October 23, 2025  
**Related Use Cases**: UC-TRACK-001 (Safiri Salama)

## Overview

The Agent Property Auto-Broadcasting system is a new framework capability that enables agents to automatically publish selected properties (such as GPS location, status, metrics) at configurable intervals to subscribed agents. This introduces intelligent, context-aware property sharing with privacy controls, eliminating the need for manual property updates and enabling real-time data distribution patterns.

**Key Innovation**: Agents autonomously decide **what** to broadcast, **when** to broadcast it, **how frequently**, and **to whom**, based on their current operational context and privacy settings.

## Motivation & Use Cases

### Primary Use Case: Real-Time Location Tracking (UC-TRACK-001)
Vehicle agents (buses, matatus) need to broadcast their GPS location to subscribers (parents, passengers) with varying frequency based on operational context:
- **At stop**: High frequency (5s intervals) - passengers need immediate updates
- **En route**: Moderate frequency (30s intervals) - tracking during transit
- **Highway**: Low frequency (60s intervals) - efficient battery/bandwidth usage
- **Emergency**: Critical frequency (3s intervals) - maximum visibility
- **Privacy mode**: No broadcasting - driver break periods

### Additional Use Cases
This pattern extends beyond location tracking to any scenario requiring automatic property publication:

1. **IoT Sensor Networks**: Equipment status broadcasting (temperature, pressure, vibration)
2. **Inventory Management**: Real-time stock level updates from warehouse agents
3. **Service Availability**: Live capacity broadcasting from service provider agents
4. **Health Monitoring**: Continuous metric reporting from infrastructure agents
5. **Fleet Management**: Vehicle telemetry and diagnostics broadcasting
6. **Smart Building**: Occupancy and environmental condition updates

## Architecture

### Core Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Publishing Agent                          │
│  ┌────────────────────────────────────────────────────────┐ │
│  │         Broadcast Configuration                         │ │
│  │  - Properties to broadcast: [location, speed, status]  │ │
│  │  - Broadcast rules (context-based intervals)           │ │
│  │  - Privacy controls                                     │ │
│  │  - Subscriber filters                                   │ │
│  └────────────────────────────────────────────────────────┘ │
│                           │                                  │
│                           ▼                                  │
│  ┌────────────────────────────────────────────────────────┐ │
│  │         Context Evaluator                               │ │
│  │  - Evaluates current state/context                      │ │
│  │  - Matches broadcast rules                              │ │
│  │  - Determines optimal interval                          │ │
│  └────────────────────────────────────────────────────────┘ │
│                           │                                  │
│                           ▼                                  │
│  ┌────────────────────────────────────────────────────────┐ │
│  │         Property Broadcaster                            │ │
│  │  - Extracts configured properties                       │ │
│  │  - Applies privacy filters                              │ │
│  │  - Publishes to message bus                             │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
        ┌──────────────────────────────────────┐
        │      Message Broker / PubSub          │
        │  - Topic: agent.{id}.property.update  │
        │  - Filtering & routing                │
        │  - Subscriber management               │
        └──────────────────────────────────────┘
                           │
          ┌────────────────┴────────────────┐
          ▼                                  ▼
┌─────────────────────┐          ┌─────────────────────┐
│  Subscriber Agent A  │          │  Subscriber Agent B  │
│  - Receives updates  │          │  - Receives updates  │
│  - Filters relevant  │          │  - Filters relevant  │
│  - Processes data    │          │  - Processes data    │
└─────────────────────┘          └─────────────────────┘
```

### System Components

#### 1. **BroadcastConfiguration** (Per Agent)
Defines what and how an agent broadcasts properties.

```go
type BroadcastConfiguration struct {
    // Enabled indicates if broadcasting is active
    Enabled bool `json:"enabled"`
    
    // PropertiesToBroadcast lists property names to auto-publish
    PropertiesToBroadcast []string `json:"properties_to_broadcast"`
    
    // BroadcastRules define context-based intervals
    BroadcastRules []BroadcastRule `json:"broadcast_rules"`
    
    // DefaultInterval is fallback when no rule matches (seconds)
    DefaultInterval int `json:"default_interval_seconds"`
    
    // PrivacyControls define broadcasting restrictions
    PrivacyControls PrivacyControls `json:"privacy_controls"`
    
    // SubscriberFilters restrict who can receive updates
    SubscriberFilters SubscriberFilters `json:"subscriber_filters"`
    
    // Metadata for tracking and analytics
    Metadata map[string]string `json:"metadata,omitempty"`
}

type BroadcastRule struct {
    // ID is unique identifier for this rule
    ID string `json:"id"`
    
    // Condition is expression evaluated against agent state
    // Examples: "status == 'at_stop'", "speed > 80", "passenger_count < capacity"
    Condition string `json:"condition"`
    
    // IntervalSeconds is broadcast frequency when condition matches
    IntervalSeconds int `json:"interval_seconds"`
    
    // Priority affects rule evaluation order (higher = evaluated first)
    Priority string `json:"priority"` // "critical", "high", "normal", "low"
    
    // NotifyFavorites triggers notification to favorited subscribers
    NotifyFavorites bool `json:"notify_favorites,omitempty"`
    
    // BroadcastMessage is optional custom message included in update
    BroadcastMessage string `json:"broadcast_message,omitempty"`
    
    // AdditionalProperties to include when this rule fires
    AdditionalProperties []string `json:"additional_properties,omitempty"`
}

type PrivacyControls struct {
    // AllowDriverPause permits manual pause of broadcasting
    AllowDriverPause bool `json:"allow_driver_pause"`
    
    // MaxSilenceMinutes is longest allowed pause duration
    MaxSilenceMinutes int `json:"max_silence_minutes"`
    
    // GeofenceRestrictions are areas where broadcasting is disabled
    GeofenceRestrictions []string `json:"geofence_restrictions"`
    
    // TimeRestrictions define when broadcasting is allowed
    TimeRestrictions []TimeWindow `json:"time_restrictions,omitempty"`
    
    // PropertyMasking defines which properties to redact in certain contexts
    PropertyMasking map[string]MaskingRule `json:"property_masking,omitempty"`
}

type SubscriberFilters struct {
    // AllowedSubscriberTypes restricts by role
    AllowedSubscriberTypes []string `json:"allowed_subscriber_types"`
    
    // EnableFavoriteNotifications allows special alerts to favorited subscribers
    EnableFavoriteNotifications bool `json:"enable_favorite_notifications"`
    
    // EnablePublicTracking allows non-subscribed viewing
    EnablePublicTracking bool `json:"enable_public_tracking"`
    
    // MaxSubscribers limits total subscriber count
    MaxSubscribers int `json:"max_subscribers,omitempty"`
    
    // RequireApproval needs publisher to approve subscriptions
    RequireApproval bool `json:"require_approval"`
}
```

#### 2. **PropertyBroadcaster** (Core Service)
Manages the broadcasting lifecycle for agents.

```go
type PropertyBroadcaster interface {
    // Configure sets up broadcasting for an agent
    Configure(ctx context.Context, agentID string, config BroadcastConfiguration) error
    
    // Start begins automatic broadcasting for an agent
    Start(ctx context.Context, agentID string) error
    
    // Stop halts broadcasting for an agent
    Stop(ctx context.Context, agentID string) error
    
    // Pause temporarily disables broadcasting (privacy mode)
    Pause(ctx context.Context, agentID string, duration time.Duration) error
    
    // Resume re-enables broadcasting after pause
    Resume(ctx context.Context, agentID string) error
    
    // UpdateInterval dynamically adjusts broadcast frequency
    UpdateInterval(ctx context.Context, agentID string, interval int, reason string) error
    
    // BroadcastNow forces immediate property publication (on-demand)
    BroadcastNow(ctx context.Context, agentID string) error
    
    // GetConfiguration retrieves current broadcast config
    GetConfiguration(ctx context.Context, agentID string) (*BroadcastConfiguration, error)
    
    // GetMetrics returns broadcasting statistics
    GetMetrics(ctx context.Context, agentID string) (*BroadcastMetrics, error)
}

type BroadcastMetrics struct {
    AgentID              string    `json:"agent_id"`
    TotalBroadcasts      int64     `json:"total_broadcasts"`
    LastBroadcastTime    time.Time `json:"last_broadcast_time"`
    CurrentInterval      int       `json:"current_interval_seconds"`
    ActiveSubscribers    int       `json:"active_subscribers"`
    AverageBroadcastSize int       `json:"average_broadcast_size_bytes"`
    ErrorCount           int64     `json:"error_count"`
    PauseCount           int64     `json:"pause_count"`
}
```

#### 3. **ContextEvaluator**
Evaluates agent state to determine appropriate broadcast interval.

```go
type ContextEvaluator interface {
    // EvaluateRules matches agent state against broadcast rules
    EvaluateRules(ctx context.Context, agentID string, agentState map[string]interface{}) (*BroadcastRule, error)
    
    // ShouldBroadcast determines if broadcasting should occur now
    ShouldBroadcast(ctx context.Context, agentID string) (bool, error)
    
    // IsPrivacyRestricted checks if current location/time triggers privacy controls
    IsPrivacyRestricted(ctx context.Context, agentID string, location *GeoLocation) (bool, string, error)
}
```

#### 4. **SubscriptionManager**
Manages subscriber relationships and permissions.

```go
type SubscriptionManager interface {
    // Subscribe registers a subscriber to receive property updates
    Subscribe(ctx context.Context, publisherID, subscriberID string, filters SubscriptionFilters) error
    
    // Unsubscribe removes a subscriber
    Unsubscribe(ctx context.Context, publisherID, subscriberID string) error
    
    // AddFavorite marks publisher as favorite for prioritized notifications
    AddFavorite(ctx context.Context, subscriberID, publisherID string) error
    
    // RemoveFavorite removes favorite status
    RemoveFavorite(ctx context.Context, subscriberID, publisherID string) error
    
    // GetSubscribers returns all subscribers for a publisher
    GetSubscribers(ctx context.Context, publisherID string) ([]Subscriber, error)
    
    // GetSubscriptions returns all publishers a subscriber follows
    GetSubscriptions(ctx context.Context, subscriberID string) ([]Publisher, error)
    
    // ApproveSubscription approves pending subscription request
    ApproveSubscription(ctx context.Context, publisherID, subscriberID string) error
}

type Subscriber struct {
    ID              string                `json:"id"`
    Type            string                `json:"type"`
    SubscribedAt    time.Time             `json:"subscribed_at"`
    IsFavorite      bool                  `json:"is_favorite"`
    Filters         SubscriptionFilters   `json:"filters"`
    NotificationPreferences NotificationPreferences `json:"notification_preferences"`
}

type SubscriptionFilters struct {
    // PropertyFilter limits which properties subscriber receives
    PropertyFilter []string `json:"property_filter,omitempty"`
    
    // MinimumPriority only receives updates at or above this priority
    MinimumPriority string `json:"minimum_priority,omitempty"`
    
    // GeofenceFilter only receives updates within specified areas
    GeofenceFilter []string `json:"geofence_filter,omitempty"`
    
    // TimeFilter only receives updates during specified windows
    TimeFilter []TimeWindow `json:"time_filter,omitempty"`
}
```

### Data Structures

#### Property Update Message

```go
type PropertyUpdateMessage struct {
    // MessageID is unique identifier for this update
    MessageID string `json:"message_id"`
    
    // MessageType is always "property_update"
    MessageType string `json:"message_type"`
    
    // Timestamp when update was generated
    Timestamp time.Time `json:"timestamp"`
    
    // PublisherID is agent broadcasting the update
    PublisherID string `json:"publisher_id"`
    
    // PublisherType is type of publishing agent
    PublisherType string `json:"publisher_type"`
    
    // Properties contains the actual property values
    Properties map[string]interface{} `json:"properties"`
    
    // BroadcastRule indicates which rule triggered this update
    BroadcastRule string `json:"broadcast_rule,omitempty"`
    
    // Priority of this update
    Priority string `json:"priority"`
    
    // CustomMessage optional message for subscribers
    CustomMessage string `json:"custom_message,omitempty"`
    
    // Metadata additional context
    Metadata map[string]string `json:"metadata,omitempty"`
}
```

#### Example Message

```json
{
  "message_id": "msg-abc123-xyz789",
  "message_type": "property_update",
  "timestamp": "2025-10-23T14:30:15Z",
  "publisher_id": "MATATU-KCA-123X",
  "publisher_type": "vehicle",
  "properties": {
    "current_location": {
      "latitude": -1.2921,
      "longitude": 36.8219,
      "heading": 45
    },
    "current_speed": 35,
    "status": "approaching_stop",
    "current_passenger_count": 12,
    "capacity": 14,
    "eta_to_next_stop": 3,
    "next_stop_id": "STOP-NGONG-CBD-05",
    "vehicle_name": "Ngong Flyer"
  },
  "broadcast_rule": "approaching_stop",
  "priority": "high",
  "custom_message": "Arriving at CBD in 3 minutes - 2 seats available",
  "metadata": {
    "sacco": "Ngong Route SACCO",
    "route": "Ngong-CBD",
    "driver_id": "DRV-12345"
  }
}
```

## Implementation Guidelines

### Agent Base Class Extensions

```go
// Add to Agent struct
type Agent struct {
    // ... existing fields ...
    
    // BroadcastConfig defines property broadcasting behavior
    BroadcastConfig *BroadcastConfiguration `json:"broadcast_config,omitempty"`
    
    // broadcaster is internal service for managing broadcasts
    broadcaster PropertyBroadcaster
    
    // broadcastTicker manages periodic broadcasts
    broadcastTicker *time.Ticker
    
    // broadcastCtx controls broadcast lifecycle
    broadcastCtx context.Context
    broadcastCancel context.CancelFunc
}

// New methods to add to Agent
func (a *Agent) EnableBroadcasting(config BroadcastConfiguration) error {
    a.BroadcastConfig = &config
    return a.broadcaster.Configure(context.Background(), a.ID, config)
}

func (a *Agent) StartBroadcasting() error {
    if a.BroadcastConfig == nil || !a.BroadcastConfig.Enabled {
        return errors.New("broadcasting not configured")
    }
    
    a.broadcastCtx, a.broadcastCancel = context.WithCancel(context.Background())
    return a.broadcaster.Start(a.broadcastCtx, a.ID)
}

func (a *Agent) StopBroadcasting() error {
    if a.broadcastCancel != nil {
        a.broadcastCancel()
    }
    return a.broadcaster.Stop(context.Background(), a.ID)
}

func (a *Agent) UpdateBroadcastInterval(interval int, reason string) error {
    return a.broadcaster.UpdateInterval(context.Background(), a.ID, interval, reason)
}

func (a *Agent) PauseBroadcasting(duration time.Duration) error {
    return a.broadcaster.Pause(context.Background(), a.ID, duration)
}

func (a *Agent) ResumeBroadcasting() error {
    return a.broadcaster.Resume(context.Background(), a.ID)
}

func (a *Agent) BroadcastNow() error {
    return a.broadcaster.BroadcastNow(context.Background(), a.ID)
}

// Internal method called by broadcaster
func (a *Agent) collectBroadcastProperties() map[string]interface{} {
    if a.BroadcastConfig == nil {
        return nil
    }
    
    properties := make(map[string]interface{})
    
    // Extract configured properties from agent state
    for _, propName := range a.BroadcastConfig.PropertiesToBroadcast {
        if value, exists := a.Properties[propName]; exists {
            properties[propName] = value
        }
    }
    
    return properties
}
```

### PubSub Service Integration

Extend the existing PubSub service to support property broadcasting:

```go
// Add to PubSubService
type PubSubService interface {
    // ... existing methods ...
    
    // PublishPropertyUpdate broadcasts property changes to subscribers
    PublishPropertyUpdate(ctx context.Context, publisherID string, properties map[string]interface{}, priority string) error
    
    // SubscribeToProperties registers for property updates from publisher
    SubscribeToProperties(ctx context.Context, publisherID, subscriberID string, filters SubscriptionFilters) error
    
    // UnsubscribeFromProperties removes property update subscription
    UnsubscribeFromProperties(ctx context.Context, publisherID, subscriberID string) error
}
```

### Configuration Repository

Store broadcast configurations persistently:

```go
type BroadcastConfigRepository interface {
    // Save stores broadcast configuration
    Save(ctx context.Context, agentID string, config *BroadcastConfiguration) error
    
    // Get retrieves broadcast configuration
    Get(ctx context.Context, agentID string) (*BroadcastConfiguration, error)
    
    // Delete removes broadcast configuration
    Delete(ctx context.Context, agentID string) error
    
    // List returns all configurations (for admin/monitoring)
    List(ctx context.Context, filters map[string]interface{}) ([]*BroadcastConfiguration, error)
}
```

## API Endpoints

### REST API for Broadcasting Management

```
POST   /api/v1/agents/{agentId}/broadcasting/configure
GET    /api/v1/agents/{agentId}/broadcasting/config
PUT    /api/v1/agents/{agentId}/broadcasting/config
DELETE /api/v1/agents/{agentId}/broadcasting/config

POST   /api/v1/agents/{agentId}/broadcasting/start
POST   /api/v1/agents/{agentId}/broadcasting/stop
POST   /api/v1/agents/{agentId}/broadcasting/pause
POST   /api/v1/agents/{agentId}/broadcasting/resume

PUT    /api/v1/agents/{agentId}/broadcasting/interval
POST   /api/v1/agents/{agentId}/broadcasting/now

GET    /api/v1/agents/{agentId}/broadcasting/metrics
GET    /api/v1/agents/{agentId}/broadcasting/subscribers

POST   /api/v1/agents/{agentId}/broadcasting/subscribe
DELETE /api/v1/agents/{agentId}/broadcasting/unsubscribe
POST   /api/v1/agents/{agentId}/broadcasting/favorite
DELETE /api/v1/agents/{agentId}/broadcasting/favorite
```

### Example API Calls

**Configure Broadcasting**:
```bash
POST /api/v1/agents/MATATU-KCA-123X/broadcasting/configure
Content-Type: application/json

{
  "enabled": true,
  "properties_to_broadcast": [
    "current_location",
    "current_speed",
    "status",
    "current_passenger_count",
    "eta_to_next_stop"
  ],
  "broadcast_rules": [
    {
      "id": "at_stop",
      "condition": "status == 'at_stop'",
      "interval_seconds": 5,
      "priority": "high",
      "notify_favorites": true
    },
    {
      "id": "approaching_stop",
      "condition": "status == 'approaching_stop'",
      "interval_seconds": 10,
      "priority": "high",
      "notify_favorites": true
    },
    {
      "id": "en_route",
      "condition": "status == 'en_route'",
      "interval_seconds": 30,
      "priority": "normal"
    }
  ],
  "default_interval_seconds": 60,
  "privacy_controls": {
    "allow_driver_pause": true,
    "max_silence_minutes": 30,
    "geofence_restrictions": ["driver_home", "maintenance_yard"]
  },
  "subscriber_filters": {
    "allowed_subscriber_types": ["passenger", "sacco_manager"],
    "enable_favorite_notifications": true,
    "enable_public_tracking": true
  }
}
```

**Subscribe to Property Updates**:
```bash
POST /api/v1/agents/MATATU-KCA-123X/broadcasting/subscribe
Content-Type: application/json

{
  "subscriber_id": "PASSENGER-USER-789",
  "filters": {
    "property_filter": ["current_location", "eta_to_next_stop", "status"],
    "minimum_priority": "normal"
  },
  "notification_preferences": {
    "channels": ["push", "sms"],
    "notify_on_favorite_approaching": true
  }
}
```

## Performance Considerations

### Scalability Targets

- **Concurrent Broadcasting Agents**: Support 10,000+ agents broadcasting simultaneously
- **Subscribers per Agent**: Handle 1,000+ subscribers per popular agent
- **Broadcast Frequency**: Maintain sub-second latency even at 3-second intervals
- **Message Throughput**: Process 100,000+ property updates per minute
- **Subscriber Notifications**: Deliver within 500ms of broadcast

### Optimization Strategies

1. **Batching**: Group multiple property updates in single message when possible
2. **Caching**: Cache subscriber lists and filters in memory with TTL
3. **Async Publishing**: Use async message broker (NATS, RabbitMQ, Kafka)
4. **Sharding**: Distribute agents across multiple broadcaster instances
5. **Filtering at Source**: Apply subscriber filters before broadcasting to reduce network traffic
6. **Connection Pooling**: Maintain persistent connections to message broker
7. **Compression**: Compress large property payloads
8. **Rate Limiting**: Throttle excessive broadcasts from misbehaving agents

### Resource Management

```go
type BroadcastResourceLimits struct {
    // MaxBroadcastsPerMinute limits broadcast frequency per agent
    MaxBroadcastsPerMinute int `json:"max_broadcasts_per_minute"`
    
    // MaxSubscribers limits total subscribers per agent
    MaxSubscribers int `json:"max_subscribers"`
    
    // MaxPropertySize limits individual property value size (bytes)
    MaxPropertySize int `json:"max_property_size_bytes"`
    
    // MaxPropertiesPerBroadcast limits properties per update
    MaxPropertiesPerBroadcast int `json:"max_properties_per_broadcast"`
    
    // MinIntervalSeconds enforces minimum broadcast interval
    MinIntervalSeconds int `json:"min_interval_seconds"`
}
```

## Security & Privacy

### Privacy Controls Implementation

1. **Geofence Restrictions**: Disable broadcasting in sensitive areas
   ```go
   func (p *PropertyBroadcaster) isInRestrictedZone(location *GeoLocation, restrictions []string) bool {
       for _, zone := range restrictions {
           if p.geofenceService.IsInside(location, zone) {
               return true
           }
       }
       return false
   }
   ```

2. **Property Masking**: Redact sensitive properties in certain contexts
   ```go
   func (p *PropertyBroadcaster) applyMasking(properties map[string]interface{}, rules map[string]MaskingRule) {
       for key, rule := range rules {
           if value, exists := properties[key]; exists {
               properties[key] = rule.Mask(value)
           }
       }
   }
   ```

3. **Time-Based Restrictions**: Only broadcast during specified hours
4. **Subscriber Authentication**: Verify subscriber identity and permissions
5. **Encryption**: Encrypt sensitive properties in transit
6. **Audit Logging**: Track all subscription requests and broadcasts

### Permission Model

```go
type BroadcastPermissions struct {
    // CanSubscribe checks if agent can subscribe to publisher
    CanSubscribe(subscriberID, publisherID string) bool
    
    // CanViewProperty checks if subscriber can access specific property
    CanViewProperty(subscriberID, publisherID, propertyName string) bool
    
    // RequiresApproval checks if subscription needs publisher approval
    RequiresApproval(subscriberID, publisherID string) bool
}
```

## Monitoring & Observability

### Metrics to Track

1. **Broadcast Metrics**:
   - Total broadcasts per agent
   - Average broadcast size
   - Broadcast failure rate
   - Interval distribution (how often each rule fires)

2. **Subscription Metrics**:
   - Active subscribers count
   - Subscription growth rate
   - Favorite count per agent
   - Churn rate (unsubscribes)

3. **Performance Metrics**:
   - Broadcast-to-delivery latency
   - Message broker queue depth
   - Memory usage per broadcaster
   - CPU usage during peak broadcasts

4. **Business Metrics** (UC-TRACK-001 specific):
   - Trackable vehicles count
   - Average subscribers per vehicle
   - Favorite rate (% of subscribers who favorite)
   - Notification open rate

### Logging

```go
type BroadcastEvent struct {
    Timestamp   time.Time              `json:"timestamp"`
    EventType   string                 `json:"event_type"` // "broadcast", "subscribe", "pause", etc.
    AgentID     string                 `json:"agent_id"`
    SubscriberID string                `json:"subscriber_id,omitempty"`
    Properties  []string               `json:"properties,omitempty"`
    Rule        string                 `json:"rule,omitempty"`
    Interval    int                    `json:"interval,omitempty"`
    Error       string                 `json:"error,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}
```

## Testing Strategy

### Unit Tests

1. **BroadcastConfiguration Validation**: Test rule parsing and validation
2. **Context Evaluation**: Test rule matching against various agent states
3. **Privacy Controls**: Test geofence, time, and masking logic
4. **Subscriber Filtering**: Test filter application and permissions

### Integration Tests

1. **End-to-End Broadcasting**: Agent publishes → Subscriber receives
2. **Dynamic Interval Adjustment**: Context changes → Interval updates
3. **Subscription Management**: Subscribe → Receive updates → Unsubscribe
4. **Favorite Notifications**: Favorite → Receive priority alerts

### Performance Tests

1. **Load Testing**: 10,000 agents broadcasting simultaneously
2. **Scalability Testing**: Add subscribers incrementally, measure latency
3. **Stress Testing**: Maximum broadcast frequency (3s intervals)
4. **Endurance Testing**: 24-hour continuous broadcasting

### Test Scenarios (UC-TRACK-001)

```go
func TestVehicleLocationBroadcasting(t *testing.T) {
    // Setup
    vehicle := createVehicleAgent("MATATU-TEST-001")
    passenger := createPassengerAgent("PASSENGER-001")
    
    // Configure broadcasting
    config := BroadcastConfiguration{
        Enabled: true,
        PropertiesToBroadcast: []string{"current_location", "status", "eta_to_next_stop"},
        BroadcastRules: []BroadcastRule{
            {Condition: "status == 'approaching_stop'", IntervalSeconds: 10, Priority: "high"},
        },
        DefaultInterval: 60,
    }
    vehicle.EnableBroadcasting(config)
    
    // Subscribe passenger
    passenger.SubscribeToVehicle(vehicle.ID)
    
    // Start broadcasting
    vehicle.StartBroadcasting()
    
    // Simulate vehicle approaching stop
    vehicle.UpdateProperty("status", "approaching_stop")
    
    // Assert: Passenger should receive update within 10 seconds
    update := passenger.WaitForUpdate(15 * time.Second)
    assert.NotNil(t, update)
    assert.Equal(t, "approaching_stop", update.Properties["status"])
}
```

## Migration Path

### Phase 1: Core Infrastructure (MVP-016)
- Implement BroadcastConfiguration data structures
- Build PropertyBroadcaster service
- Create ContextEvaluator for rule matching
- Integrate with existing PubSub service

### Phase 2: Subscription Management (MVP-017)
- Implement SubscriptionManager
- Build subscriber filtering logic
- Add favorite functionality
- Create subscription API endpoints

### Phase 3: Privacy & Security (MVP-018)
- Implement geofencing service integration
- Add property masking
- Build permission model
- Add audit logging

### Phase 4: Optimization & Scale (MVP-019)
- Performance tuning and caching
- Load balancing for broadcasters
- Message broker optimization
- Monitoring and alerting

### Phase 5: UC-TRACK-001 Integration (MVP-020)
- Vehicle agent implementation
- Passenger subscription features
- SACCO management portal
- Mobile app integration

## Configuration Examples

### Example 1: School Bus (High Safety Priority)

```yaml
agent_id: "BUS-SCHOOL-001"
broadcast_config:
  enabled: true
  properties_to_broadcast:
    - current_location
    - status
    - current_passenger_count
    - next_stop_id
    - eta_to_next_stop
  broadcast_rules:
    - id: "emergency"
      condition: "status == 'emergency'"
      interval_seconds: 3
      priority: "critical"
      notify_favorites: true
      broadcast_message: "EMERGENCY - All parents notified"
    - id: "at_school_stop"
      condition: "status == 'at_stop' && stop_type == 'school'"
      interval_seconds: 5
      priority: "high"
    - id: "pickup_dropoff"
      condition: "status == 'at_stop'"
      interval_seconds: 10
      priority: "high"
    - id: "en_route"
      condition: "status == 'en_route'"
      interval_seconds: 30
      priority: "normal"
  default_interval_seconds: 60
  privacy_controls:
    allow_driver_pause: false  # Never pause for school buses
    max_silence_minutes: 5
    geofence_restrictions: []  # No restrictions for school buses
  subscriber_filters:
    allowed_subscriber_types: ["parent", "school_admin", "route_manager"]
    enable_public_tracking: false  # Privacy - only enrolled parents
    require_approval: true  # Parents must be pre-enrolled
```

### Example 2: Matatu (Customer Loyalty Focus)

```yaml
agent_id: "MATATU-KCA-123X"
broadcast_config:
  enabled: true
  properties_to_broadcast:
    - current_location
    - status
    - current_passenger_count
    - capacity
    - eta_to_next_stop
    - vehicle_name
    - rating_average
  broadcast_rules:
    - id: "approaching_favorite_stop"
      condition: "status == 'approaching_stop' && has_favorited_passengers_at_stop"
      interval_seconds: 5
      priority: "high"
      notify_favorites: true
      broadcast_message: "Your favorite matatu arriving soon!"
    - id: "at_stop_with_space"
      condition: "status == 'at_stop' && current_passenger_count < capacity"
      interval_seconds: 5
      priority: "high"
      notify_favorites: true
      additional_properties: ["seats_available"]
    - id: "at_stop_full"
      condition: "status == 'at_stop' && current_passenger_count >= capacity"
      interval_seconds: 30
      priority: "low"
      broadcast_message: "Vehicle at capacity"
    - id: "en_route"
      condition: "status == 'en_route'"
      interval_seconds: 30
      priority: "normal"
  default_interval_seconds: 60
  privacy_controls:
    allow_driver_pause: true
    max_silence_minutes: 30
    geofence_restrictions: ["driver_home", "maintenance_yard"]
    time_restrictions:
      - start_time: "05:00"
        end_time: "22:00"
        days: ["monday", "tuesday", "wednesday", "thursday", "friday"]
  subscriber_filters:
    allowed_subscriber_types: ["passenger", "sacco_manager", "route_manager"]
    enable_favorite_notifications: true
    enable_public_tracking: true  # Anyone can track
    max_subscribers: 1000
    require_approval: false  # Open subscription
```

## Documentation Requirements

1. **API Documentation**: OpenAPI/Swagger specs for all endpoints
2. **Developer Guide**: How to enable broadcasting in custom agents
3. **Configuration Guide**: Detailed explanation of all config options
4. **Best Practices**: Recommendations for broadcast intervals, privacy, etc.
5. **Troubleshooting**: Common issues and solutions
6. **Performance Tuning**: Optimization tips for high-scale deployments

## Related Work & References

- **PubSub System**: `internal/communication/pubsub_service.go`
- **Agent Lifecycle**: `documents/3-SofwareDevelopment/core-systems/agent-lifecycle.md`
- **Message Service**: `internal/communication/message_service.go`
- **UC-TRACK-001**: `documents/1-SoftwareRequirements/requirements/use-cases/UC-TRACK-001-safiri-salama.md`

## Future Enhancements

1. **Conditional Property Inclusion**: Only include properties when conditions met
2. **Broadcast Aggregation**: Combine multiple properties into batches
3. **Predictive Intervals**: ML-based interval adjustment based on patterns
4. **Cross-Agent Coordination**: Multiple agents broadcasting in sync
5. **Broadcast Analytics Dashboard**: Visual monitoring of broadcast patterns
6. **A/B Testing Framework**: Test different interval strategies
7. **Cost Optimization**: Smart bandwidth usage based on subscriber activity

---

**Status**: Ready for MVP Planning  
**Complexity**: High  
**Timeline Estimate**: 4-6 sprints across 5 MVP tasks  
**Dependencies**: PubSub service, Agent lifecycle management  
**Risk Level**: Medium - New pattern, requires performance validation
