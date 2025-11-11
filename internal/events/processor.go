package events

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// ProcessorConfig configures the event processor
type ProcessorConfig struct {
	// QueueSize is the maximum number of events in the queue
	QueueSize int

	// WorkerCount is the number of worker goroutines for processing events
	WorkerCount int

	// MaxRetries is the maximum number of retry attempts for failed events
	MaxRetries int

	// RetryDelay is the delay between retry attempts
	RetryDelay time.Duration

	// ProcessingTimeout is the maximum time to spend processing a single event
	ProcessingTimeout time.Duration
}

// DefaultProcessorConfig returns a default processor configuration
func DefaultProcessorConfig() ProcessorConfig {
	return ProcessorConfig{
		QueueSize:         1000,
		WorkerCount:       5,
		MaxRetries:        3,
		RetryDelay:        time.Second,
		ProcessingTimeout: 30 * time.Second,
	}
}

// Processor manages event processing with internal event loops
type Processor struct {
	config    ProcessorConfig
	eventChan chan *Event
	registry  *HandlerRegistry
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
	running   bool
	metrics   *ProcessorMetrics
}

// ProcessorMetrics tracks event processing statistics
type ProcessorMetrics struct {
	TotalEvents     int64
	ProcessedEvents int64
	FailedEvents    int64
	RetryAttempts   int64
	AverageLatency  time.Duration
	mu              sync.RWMutex
}

// NewProcessor creates a new event processor
func NewProcessor(config ProcessorConfig) *Processor {
	ctx, cancel := context.WithCancel(context.Background())

	return &Processor{
		config:    config,
		eventChan: make(chan *Event, config.QueueSize),
		registry:  NewHandlerRegistry(),
		ctx:       ctx,
		cancel:    cancel,
		metrics:   &ProcessorMetrics{},
	}
}

// Start begins the event processing loops
func (p *Processor) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return fmt.Errorf("processor is already running")
	}

	p.running = true

	// Start worker goroutines
	for i := 0; i < p.config.WorkerCount; i++ {
		p.wg.Add(1)
		go p.eventLoop(i)
	}

	log.WithField("workers", p.config.WorkerCount).Info("Event processor started")
	return nil
}

// Stop gracefully shuts down the event processor
func (p *Processor) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return fmt.Errorf("processor is not running")
	}

	p.running = false
	p.cancel()

	// Close event channel to signal workers to stop
	close(p.eventChan)

	// Wait for all workers to finish
	p.wg.Wait()

	log.Info("Event processor stopped")
	return nil
}

// PublishEvent sends an event for processing
func (p *Processor) PublishEvent(event *Event) error {
	p.mu.RLock()
	running := p.running
	p.mu.RUnlock()

	if !running {
		return fmt.Errorf("processor is not running")
	}

	// Generate ID if not provided
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Set context if not provided
	if event.Context == nil {
		event.Context = context.Background()
	}

	// Try to send event to channel (non-blocking)
	select {
	case p.eventChan <- event:
		p.updateMetrics(func(m *ProcessorMetrics) {
			m.TotalEvents++
		})
		return nil
	default:
		return fmt.Errorf("event queue is full")
	}
}

// RegisterHandler registers an event handler
func (p *Processor) RegisterHandler(handler EventHandler, eventTypes ...EventType) error {
	return p.registry.RegisterHandler(handler, eventTypes...)
}

// UnregisterHandler removes an event handler
func (p *Processor) UnregisterHandler(handler EventHandler) error {
	return p.registry.UnregisterHandler(handler)
}

// GetMetrics returns current processor metrics
func (p *Processor) GetMetrics() ProcessorMetrics {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()

	// Return a copy to avoid exposing the mutex
	return ProcessorMetrics{
		TotalEvents:     p.metrics.TotalEvents,
		ProcessedEvents: p.metrics.ProcessedEvents,
		FailedEvents:    p.metrics.FailedEvents,
		RetryAttempts:   p.metrics.RetryAttempts,
		AverageLatency:  p.metrics.AverageLatency,
	}
}

// eventLoop is the main processing loop for a worker
func (p *Processor) eventLoop(workerID int) {
	defer p.wg.Done()

	logger := log.WithField("worker_id", workerID)

	for event := range p.eventChan {
		startTime := time.Now()

		// Process event with timeout
		err := p.processEventWithRetry(event)

		duration := time.Since(startTime)

		// Update metrics
		p.updateMetrics(func(m *ProcessorMetrics) {
			if err == nil {
				m.ProcessedEvents++
			} else {
				m.FailedEvents++
			}

			// Update average latency (simple moving average)
			if m.ProcessedEvents > 0 {
				m.AverageLatency = time.Duration(
					(int64(m.AverageLatency)*m.ProcessedEvents + int64(duration)) / (m.ProcessedEvents + 1),
				)
			} else {
				m.AverageLatency = duration
			}
		})

		// Log processing result
		if err != nil {
			logger.WithFields(log.Fields{
				"event_id":   event.ID,
				"event_type": event.Type,
				"error":      err,
				"duration":   duration,
			}).Error("Event processing failed")
		}
	}
}

// processEventWithRetry processes an event with retry logic
func (p *Processor) processEventWithRetry(event *Event) error {
	var lastError error

	for attempt := 0; attempt <= p.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-time.After(p.config.RetryDelay):
			case <-p.ctx.Done():
				return p.ctx.Err()
			}

			p.updateMetrics(func(m *ProcessorMetrics) {
				m.RetryAttempts++
			})
		}

		// Create timeout context for this processing attempt
		ctx, cancel := context.WithTimeout(event.Context, p.config.ProcessingTimeout)

		err := p.processEvent(ctx, event)
		cancel()

		if err == nil {
			return nil // Success
		}

		lastError = err

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			break
		}
	}

	return fmt.Errorf("event processing failed after %d attempts: %w", p.config.MaxRetries+1, lastError)
}

// processEvent processes a single event
func (p *Processor) processEvent(ctx context.Context, event *Event) error {
	// Get handlers for this event type
	handlers := p.registry.GetHandlers(event.Type)
	if len(handlers) == 0 {
		return nil
	}

	// Sort handlers by priority (higher priority first)
	sort.Slice(handlers, func(i, j int) bool {
		return handlers[i].Priority() > handlers[j].Priority()
	})

	// Process with each handler
	var errors []error
	for _, handler := range handlers {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := handler.Handle(ctx, event)
		if err != nil {
			errors = append(errors, fmt.Errorf("handler %s failed: %w", handler.Name(), err))

			log.WithFields(log.Fields{
				"event_id":     event.ID,
				"event_type":   event.Type,
				"handler_name": handler.Name(),
				"error":        err,
			}).Error("Event handler failed")
		}
	}

	// Return combined errors if any
	if len(errors) > 0 {
		return fmt.Errorf("some handlers failed: %v", errors)
	}

	return nil
}

// updateMetrics safely updates processor metrics
func (p *Processor) updateMetrics(fn func(*ProcessorMetrics)) {
	p.metrics.mu.Lock()
	fn(p.metrics)
	p.metrics.mu.Unlock()
}
