// Package events provides a simple event-driven architecture for ZTDP.
package events

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EventType defines the type of event
type EventType string

// Application events
const (
	EventTypeApplicationCreated EventType = "application.created"
	EventTypeApplicationUpdated EventType = "application.updated"
	EventTypeApplicationDeleted EventType = "application.deleted"
)

// Event represents a platform event
type Event struct {
	Type      EventType              `json:"type"`
	Source    string                 `json:"source"`
	Subject   string                 `json:"subject"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	Timestamp int64                  `json:"timestamp"`
	ID        string                 `json:"id"`
}

// EventHandler is a function that processes events
type EventHandler func(event Event) error

// EventBus is the simple event system
type EventBus struct {
	handlers     map[EventType][]EventHandler
	mu           sync.RWMutex
	transport    EventTransport
	defaultAsync bool
}

// EventTransport defines the interface for event transport (memory, kafka, etc.)
type EventTransport interface {
	Publish(topic string, data []byte) error
	Subscribe(topic string, handler func([]byte)) error
	Close() error
}

// NewEventBus creates a simple event bus
func NewEventBus(transport EventTransport, defaultAsync bool) *EventBus {
	return &EventBus{
		handlers:     make(map[EventType][]EventHandler),
		transport:    transport,
		defaultAsync: defaultAsync,
	}
}

// Subscribe registers a handler for a specific event type
func (b *EventBus) Subscribe(eventType EventType, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

// Emit publishes an event to the bus (simple interface)
func (b *EventBus) Emit(eventType EventType, source, subject string, payload map[string]interface{}) error {
	event := Event{
		Type:      eventType,
		Source:    source,
		Subject:   subject,
		Payload:   payload,
		Timestamp: time.Now().UnixNano(),
		ID:        uuid.New().String(),
	}

	// Send to transport if available
	if b.transport != nil {
		data, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		if err := b.transport.Publish(string(eventType), data); err != nil {
			return fmt.Errorf("failed to publish event: %w", err)
		}
	}

	// Process local handlers
	b.mu.RLock()
	handlers, exists := b.handlers[eventType]
	b.mu.RUnlock()

	if !exists {
		return nil
	}

	if b.defaultAsync {
		go b.processHandlers(event, handlers)
		return nil
	}

	return b.processHandlers(event, handlers)
}

// processHandlers runs all handlers for an event
func (b *EventBus) processHandlers(event Event, handlers []EventHandler) error {
	for _, handler := range handlers {
		if err := handler(event); err != nil {
			log.Printf("Error handling event %s: %v", event.Type, err)
		}
	}
	return nil
}

// MemoryTransport is a simple in-memory event transport
type MemoryTransport struct {
	subscribers map[string][]func([]byte)
	mu          sync.RWMutex
}

// NewMemoryTransport creates a new memory-based transport
func NewMemoryTransport() *MemoryTransport {
	return &MemoryTransport{
		subscribers: make(map[string][]func([]byte)),
	}
}

// Publish sends data to subscribers of a topic
func (m *MemoryTransport) Publish(topic string, data []byte) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	handlers, exists := m.subscribers[topic]
	if !exists {
		return nil
	}

	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	for _, handler := range handlers {
		go handler(dataCopy)
	}
	return nil
}

// Subscribe registers a handler for a topic
func (m *MemoryTransport) Subscribe(topic string, handler func([]byte)) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscribers[topic] = append(m.subscribers[topic], handler)
	return nil
}

// Close implements the interface but is a no-op for memory transport
func (m *MemoryTransport) Close() error {
	return nil
}

// SetupLogging sets up basic event logging
func SetupLogging(eventBus *EventBus) {
	logger := log.New(log.Writer(), "[EVENT] ", log.LstdFlags)

	eventBus.Subscribe(EventTypeApplicationCreated, func(event Event) error {
		logger.Printf("🎯 Application created: %s", event.Subject)
		return nil
	})

	eventBus.Subscribe(EventTypeApplicationUpdated, func(event Event) error {
		logger.Printf("✏️  Application updated: %s", event.Subject)
		return nil
	})

	eventBus.Subscribe(EventTypeApplicationDeleted, func(event Event) error {
		logger.Printf("🗑️  Application deleted: %s", event.Subject)
		return nil
	})
}

// Global event bus instance
var GlobalEventBus *EventBus

// InitializeEventBus sets up the global event bus
func InitializeEventBus(transport EventTransport) {
	GlobalEventBus = NewEventBus(transport, true)
	SetupLogging(GlobalEventBus)
}
