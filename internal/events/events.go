// Package events provides pure event infrastructure for ZTDP platform.
// This package contains NO business logic or domain-specific knowledge.
package events

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Generic event types for infrastructure - NO domain-specific types
type EventType string

const (
	EventTypeRequest   EventType = "request"   // Generic request event
	EventTypeResponse  EventType = "response"  // Generic response event
	EventTypeBroadcast EventType = "broadcast" // Generic broadcast event
	EventTypeNotify    EventType = "notify"    // Generic notification event
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

// SubscribeToRoutingKey registers a handler for events with specific routing keys
// This is more efficient than subscribing to all events and filtering
func (b *EventBus) SubscribeToRoutingKey(routingKey string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Create a wrapper handler that filters by routing key
	routingHandler := func(event Event) error {
		if event.Subject == routingKey {
			return handler(event)
		}
		return nil
	}

	// Add directly to handlers without calling Subscribe (avoid deadlock)
	b.handlers[EventTypeRequest] = append(b.handlers[EventTypeRequest], routingHandler)
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

// EmitEvent publishes a complete event to the bus (preserves all event fields)
func (b *EventBus) EmitEvent(event Event) error {
	// Send to transport if available
	if b.transport != nil {
		data, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		if err := b.transport.Publish(string(event.Type), data); err != nil {
			return fmt.Errorf("failed to publish event: %w", err)
		}
	}

	// Process local handlers
	b.mu.RLock()
	handlers, exists := b.handlers[event.Type]
	b.mu.RUnlock()

	if !exists {
		return nil
	}

	if b.defaultAsync {
		go b.processHandlers(event, handlers)
	} else {
		b.processHandlers(event, handlers)
	}

	return nil
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

// SetupLogging sets up basic infrastructure event logging
func SetupLogging(eventBus *EventBus) {
	logger := log.New(log.Writer(), "[EVENT] ", log.LstdFlags)

	// Generic infrastructure event logging - no domain knowledge
	eventBus.Subscribe(EventTypeRequest, func(event Event) error {
		logger.Printf("📨 Request: %s -> %s", event.Source, event.Subject)
		return nil
	})

	eventBus.Subscribe(EventTypeResponse, func(event Event) error {
		logger.Printf("📬 Response: %s -> %s", event.Source, event.Subject)
		return nil
	})

	eventBus.Subscribe(EventTypeBroadcast, func(event Event) error {
		logger.Printf("� Broadcast: %s", event.Subject)
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
