package communication

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// MessageHandler processes received messages
type MessageHandler func(msg *Message) error

// PublicationHandler processes received publications
type PublicationHandler func(pub *Publication) error

// MessagePoller polls for new messages at regular intervals
type MessagePoller struct {
	agentID        string
	messageService *MessageService
	interval       time.Duration
	batchSize      int
	handler        MessageHandler
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	running        bool
	mu             sync.RWMutex
}

// MessagePollerConfig configures a message poller
type MessagePollerConfig struct {
	AgentID   string
	Interval  time.Duration
	BatchSize int // Max messages to retrieve per poll (default: 100)
}

// NewMessagePoller creates a new message poller
func NewMessagePoller(config MessagePollerConfig, messageService *MessageService, handler MessageHandler) *MessagePoller {
	ctx, cancel := context.WithCancel(context.Background())

	// Set defaults
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}
	if config.Interval == 0 {
		config.Interval = 5 * time.Second
	}

	return &MessagePoller{
		agentID:        config.AgentID,
		messageService: messageService,
		interval:       config.Interval,
		batchSize:      config.BatchSize,
		handler:        handler,
		ctx:            ctx,
		cancel:         cancel,
		running:        false,
	}
}

// Start begins polling for messages
func (mp *MessagePoller) Start() {
	mp.mu.Lock()
	if mp.running {
		mp.mu.Unlock()
		log.WithField("agent_id", mp.agentID).Warn("Message poller already running")
		return
	}
	mp.running = true
	mp.mu.Unlock()

	mp.wg.Add(1)
	go mp.run()

	log.WithFields(log.Fields{
		"agent_id":   mp.agentID,
		"interval":   mp.interval,
		"batch_size": mp.batchSize,
	}).Info("Message poller started")
}

// Stop stops the poller
func (mp *MessagePoller) Stop() {
	mp.mu.Lock()
	if !mp.running {
		mp.mu.Unlock()
		return
	}
	mp.running = false
	mp.mu.Unlock()

	mp.cancel()
	mp.wg.Wait()

	log.WithField("agent_id", mp.agentID).Info("Message poller stopped")
}

// IsRunning returns whether the poller is currently running
func (mp *MessagePoller) IsRunning() bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.running
}

func (mp *MessagePoller) run() {
	defer mp.wg.Done()

	ticker := time.NewTicker(mp.interval)
	defer ticker.Stop()

	// Poll immediately on start
	mp.poll()

	for {
		select {
		case <-ticker.C:
			mp.poll()
		case <-mp.ctx.Done():
			return
		}
	}
}

func (mp *MessagePoller) poll() {
	messages, err := mp.messageService.GetPendingMessages(mp.ctx, mp.agentID, mp.batchSize)
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
		// Handle message
		if err := mp.handler(msg); err != nil {
			log.WithFields(log.Fields{
				"agent_id":   mp.agentID,
				"message_id": msg.ID,
				"error":      err,
			}).Error("Failed to handle message")
			// Mark as failed
			if err := mp.messageService.MarkFailed(mp.ctx, msg.ID); err != nil {
				log.WithFields(log.Fields{
					"message_id": msg.ID,
					"error":      err,
				}).Error("Failed to mark message as failed")
			}
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

// PublicationPoller polls for new publications at regular intervals
type PublicationPoller struct {
	agentID       string
	pubSubService *PubSubService
	interval      time.Duration
	handler       PublicationHandler
	lastPoll      time.Time
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	running       bool
	mu            sync.RWMutex
}

// PublicationPollerConfig configures a publication poller
type PublicationPollerConfig struct {
	AgentID  string
	Interval time.Duration
}

// NewPublicationPoller creates a new publication poller
func NewPublicationPoller(config PublicationPollerConfig, pubSubService *PubSubService, handler PublicationHandler) *PublicationPoller {
	ctx, cancel := context.WithCancel(context.Background())

	// Set defaults
	if config.Interval == 0 {
		config.Interval = 5 * time.Second
	}

	return &PublicationPoller{
		agentID:       config.AgentID,
		pubSubService: pubSubService,
		interval:      config.Interval,
		handler:       handler,
		lastPoll:      time.Now(),
		ctx:           ctx,
		cancel:        cancel,
		running:       false,
	}
}

// Start begins polling for publications
func (pp *PublicationPoller) Start() {
	pp.mu.Lock()
	if pp.running {
		pp.mu.Unlock()
		log.WithField("agent_id", pp.agentID).Warn("Publication poller already running")
		return
	}
	pp.running = true
	pp.mu.Unlock()

	pp.wg.Add(1)
	go pp.run()

	log.WithFields(log.Fields{
		"agent_id": pp.agentID,
		"interval": pp.interval,
	}).Info("Publication poller started")
}

// Stop stops the poller
func (pp *PublicationPoller) Stop() {
	pp.mu.Lock()
	if !pp.running {
		pp.mu.Unlock()
		return
	}
	pp.running = false
	pp.mu.Unlock()

	pp.cancel()
	pp.wg.Wait()

	log.WithField("agent_id", pp.agentID).Info("Publication poller stopped")
}

// IsRunning returns whether the poller is currently running
func (pp *PublicationPoller) IsRunning() bool {
	pp.mu.RLock()
	defer pp.mu.RUnlock()
	return pp.running
}

func (pp *PublicationPoller) run() {
	defer pp.wg.Done()

	ticker := time.NewTicker(pp.interval)
	defer ticker.Stop()

	// Poll immediately on start
	pp.poll()

	for {
		select {
		case <-ticker.C:
			pp.poll()
		case <-pp.ctx.Done():
			return
		}
	}
}

func (pp *PublicationPoller) poll() {
	// Get publications since last poll
	since := pp.lastPoll
	pp.lastPoll = time.Now()

	publications, err := pp.pubSubService.GetMatchingPublications(pp.ctx, pp.agentID, since)
	if err != nil {
		log.WithFields(log.Fields{
			"agent_id": pp.agentID,
			"error":    err,
		}).Error("Failed to poll publications")
		return
	}

	if len(publications) == 0 {
		return
	}

	log.WithFields(log.Fields{
		"agent_id": pp.agentID,
		"count":    len(publications),
	}).Debug("Received publications")

	for _, pub := range publications {
		// Handle publication
		if err := pp.handler(pub); err != nil {
			log.WithFields(log.Fields{
				"agent_id":       pp.agentID,
				"publication_id": pub.ID,
				"event":          pub.EventName,
				"error":          err,
			}).Error("Failed to handle publication")
			continue
		}
	}
}

// CommunicationPoller combines message and publication polling
type CommunicationPoller struct {
	messagePoller     *MessagePoller
	publicationPoller *PublicationPoller
}

// NewCommunicationPoller creates a combined poller
func NewCommunicationPoller(
	messageConfig MessagePollerConfig,
	publicationConfig PublicationPollerConfig,
	messageService *MessageService,
	pubSubService *PubSubService,
	messageHandler MessageHandler,
	publicationHandler PublicationHandler,
) *CommunicationPoller {
	return &CommunicationPoller{
		messagePoller:     NewMessagePoller(messageConfig, messageService, messageHandler),
		publicationPoller: NewPublicationPoller(publicationConfig, pubSubService, publicationHandler),
	}
}

// Start starts both pollers
func (cp *CommunicationPoller) Start() {
	cp.messagePoller.Start()
	cp.publicationPoller.Start()
}

// Stop stops both pollers
func (cp *CommunicationPoller) Stop() {
	cp.messagePoller.Stop()
	cp.publicationPoller.Stop()
}

// IsRunning returns whether both pollers are running
func (cp *CommunicationPoller) IsRunning() bool {
	return cp.messagePoller.IsRunning() && cp.publicationPoller.IsRunning()
}
