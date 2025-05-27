// Package events provides event-driven architecture capabilities for ZTDP.
package events

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

// EventType defines the type of event
type EventType string

// Define standard event types
const (
	EventTypePolicyCheck        EventType = "policy.check"
	EventTypePolicyCheckResult  EventType = "policy.check.result"
	EventTypeTransitionAttempt  EventType = "transition.attempt"
	EventTypeTransitionApproved EventType = "transition.approved"
	EventTypeTransitionRejected EventType = "transition.rejected"

	// Graph events
	EventTypeGraphNodeAdded   EventType = "graph.node.added"
	EventTypeGraphNodeUpdated EventType = "graph.node.updated"
	EventTypeGraphNodeRemoved EventType = "graph.node.removed"
	EventTypeGraphEdgeAdded   EventType = "graph.edge.added"
	EventTypeGraphEdgeRemoved EventType = "graph.edge.removed"
)

// Event represents a platform event
type Event struct {
	Type      EventType              `json:"type"`
	Source    string                 `json:"source"`
	Subject   string                 `json:"subject"`
	Action    string                 `json:"action,omitempty"`
	Status    string                 `json:"status,omitempty"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	Timestamp int64                  `json:"timestamp"`
	ID        string                 `json:"id"`
}

// EventHandler is a function that processes events
type EventHandler func(event Event) error

// EventBus is the central event dispatch mechanism
type EventBus struct {
	handlers     map[EventType][]EventHandler
	mu           sync.RWMutex
	transport    EventTransport
	isAsync      bool
	defaultAsync bool
}

// EventTransport defines the interface for event transport mechanisms
type EventTransport interface {
	Publish(topic string, data []byte) error
	Subscribe(topic string, handler func([]byte)) error
	Close() error
}

// NewEventBus creates a new event bus
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

// Publish sends an event to all subscribers
func (b *EventBus) Publish(event Event, async bool) error {
	// If transport is available, serialize and send through transport
	if b.transport != nil {
		data, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		topic := string(event.Type)
		if err := b.transport.Publish(topic, data); err != nil {
			return fmt.Errorf("failed to publish event to transport: %w", err)
		}
	}

	b.mu.RLock()
	handlers, exists := b.handlers[event.Type]
	b.mu.RUnlock()

	if !exists {
		return nil
	}

	// Process with local handlers
	if async {
		go b.processHandlers(event, handlers)
		return nil
	}

	return b.processHandlers(event, handlers)
}

// PublishDefault sends an event using the default async setting
func (b *EventBus) PublishDefault(event Event) error {
	return b.Publish(event, b.defaultAsync)
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
// for testing and single-process deployments
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

	// Make a copy to avoid holding the lock during callback execution
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
