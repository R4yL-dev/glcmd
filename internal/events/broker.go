package events

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// EventType defines the types of events supported
type EventType string

const (
	EventTypeGlucose   EventType = "glucose"
	EventTypeSensor    EventType = "sensor"
	EventTypeKeepalive EventType = "keepalive"
)

// Event represents a generic event
type Event struct {
	Type EventType
	Data interface{} // *domain.GlucoseMeasurement or *domain.SensorConfig
}

// Subscriber represents a subscriber with optional type filtering
type Subscriber struct {
	ID      string
	Channel chan Event
	Types   []EventType // Types to receive (empty = all)
}

// wantsEvent returns true if the subscriber wants events of the given type
func (s *Subscriber) wantsEvent(eventType EventType) bool {
	if len(s.Types) == 0 {
		return true // No filter = all events
	}
	for _, t := range s.Types {
		if t == eventType {
			return true
		}
	}
	return false
}

// Broker manages subscriptions and event distribution
type Broker struct {
	subscribers map[string]*Subscriber
	mu          sync.RWMutex
	bufferSize  int
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	logger      *slog.Logger
}

// NewBroker creates a new event broker with the specified channel buffer size
func NewBroker(bufferSize int, logger *slog.Logger) *Broker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Broker{
		subscribers: make(map[string]*Subscriber),
		bufferSize:  bufferSize,
		ctx:         ctx,
		cancel:      cancel,
		logger:      logger,
	}
}

// Subscribe registers a new subscriber and returns the event channel.
// types specifies which event types to receive (empty = all types).
func (b *Broker) Subscribe(id string, types []EventType) <-chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan Event, b.bufferSize)
	b.subscribers[id] = &Subscriber{
		ID:      id,
		Channel: ch,
		Types:   types,
	}

	b.logger.Debug("subscriber added",
		"clientID", id,
		"types", types,
		"subscribers", len(b.subscribers),
	)

	return ch
}

// Unsubscribe removes a subscriber and closes its channel
func (b *Broker) Unsubscribe(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if sub, ok := b.subscribers[id]; ok {
		close(sub.Channel)
		delete(b.subscribers, id)

		b.logger.Debug("subscriber removed",
			"clientID", id,
			"subscribers", len(b.subscribers),
		)
	}
}

// Publish sends an event to all matching subscribers.
// Uses non-blocking sends to prevent slow subscribers from blocking.
func (b *Broker) Publish(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, sub := range b.subscribers {
		if !sub.wantsEvent(event.Type) {
			continue
		}

		select {
		case sub.Channel <- event:
			// Event sent successfully
		default:
			// Channel full, subscriber too slow
			b.logger.Warn("SSE subscriber slow, event dropped",
				"clientID", sub.ID,
				"eventType", event.Type,
			)
		}
	}

	if event.Type != EventTypeKeepalive {
		b.logger.Debug("event published",
			"type", event.Type,
			"subscribers", len(b.subscribers),
		)
	}
}

// Start begins the heartbeat goroutine
func (b *Broker) Start() {
	b.wg.Add(1)
	go b.heartbeatLoop()
}

// Stop gracefully stops the broker and closes all subscriber channels
func (b *Broker) Stop() {
	b.cancel()
	b.wg.Wait()

	// Close all subscriber channels
	b.mu.Lock()
	defer b.mu.Unlock()
	for id, sub := range b.subscribers {
		close(sub.Channel)
		delete(b.subscribers, id)
	}
}

// SubscriberCount returns the current number of subscribers
func (b *Broker) SubscriberCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribers)
}

// heartbeatLoop sends keepalive events every 30 seconds
func (b *Broker) heartbeatLoop() {
	defer b.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.Publish(Event{Type: EventTypeKeepalive, Data: nil})
		case <-b.ctx.Done():
			return
		}
	}
}
