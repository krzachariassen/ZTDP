package messaging

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ztdp/orchestrator/internal/logging"
)

// MemoryMessageBus implements MessageBus using in-memory channels
// This is perfect for development and testing
type MemoryMessageBus struct {
	subscribers   map[string]chan *Message
	conversations map[string]*ConversationContext
	history       map[string][]*Message
	mutex         sync.RWMutex
	logger        logging.Logger
}

// NewMemoryMessageBus creates a new in-memory message bus
func NewMemoryMessageBus(logger logging.Logger) *MemoryMessageBus {
	return &MemoryMessageBus{
		subscribers:   make(map[string]chan *Message),
		conversations: make(map[string]*ConversationContext),
		history:       make(map[string][]*Message),
		logger:        logger,
	}
}

// SendMessage sends a message to a specific recipient
func (mb *MemoryMessageBus) SendMessage(ctx context.Context, message *Message) error {
	mb.mutex.RLock()
	subscriber, exists := mb.subscribers[message.ToID]
	mb.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("no subscriber found for participant %s", message.ToID)
	}

	// Store message in history
	mb.mutex.Lock()
	if message.CorrelationID != "" {
		mb.history[message.CorrelationID] = append(mb.history[message.CorrelationID], message)
	}
	mb.mutex.Unlock()

	// Send message (non-blocking)
	select {
	case subscriber <- message:
		// Log successful message delivery to subscriber
		if mb.logger != nil {
			mb.logger.Debug("📨 Message delivered to subscriber",
				"message_id", message.ID,
				"correlation_id", message.CorrelationID,
				"from", message.FromID,
				"to", message.ToID,
				"message_type", message.MessageType,
				"content_length", len(message.Content))
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("subscriber channel full for participant %s", message.ToID)
	}
}

// Subscribe subscribes to messages for a specific participant
func (mb *MemoryMessageBus) Subscribe(ctx context.Context, participantID string) (<-chan *Message, error) {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()

	if _, exists := mb.subscribers[participantID]; exists {
		return nil, fmt.Errorf("participant %s already subscribed", participantID)
	}

	ch := make(chan *Message, 100) // Buffered channel
	mb.subscribers[participantID] = ch
	return ch, nil
}

// Unsubscribe unsubscribes from messages
func (mb *MemoryMessageBus) Unsubscribe(ctx context.Context, participantID string) error {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()

	if ch, exists := mb.subscribers[participantID]; exists {
		close(ch)
		delete(mb.subscribers, participantID)
		return nil
	}

	return fmt.Errorf("participant %s not subscribed", participantID)
}

// PublishMessage publishes a message to multiple recipients
func (mb *MemoryMessageBus) PublishMessage(ctx context.Context, message *Message, recipients []string) error {
	for _, recipient := range recipients {
		msg := *message // Copy message
		msg.ToID = recipient
		if err := mb.SendMessage(ctx, &msg); err != nil {
			return fmt.Errorf("failed to send to %s: %w", recipient, err)
		}
	}
	return nil
}

// GetConversationHistory returns the conversation history
func (mb *MemoryMessageBus) GetConversationHistory(ctx context.Context, conversationID string) ([]*Message, error) {
	mb.mutex.RLock()
	defer mb.mutex.RUnlock()

	history, exists := mb.history[conversationID]
	if !exists {
		return []*Message{}, nil
	}

	// Return copy of history
	result := make([]*Message, len(history))
	copy(result, history)
	return result, nil
}

// CreateConversation creates a new conversation context
func (mb *MemoryMessageBus) CreateConversation(ctx context.Context, participants []string, context map[string]interface{}) (*ConversationContext, error) {
	conversationID := uuid.New().String()

	conversation := &ConversationContext{
		ConversationID: conversationID,
		Participants:   participants,
		Context:        context,
		StartTime:      time.Now(),
		LastActivity:   time.Now(),
	}

	mb.mutex.Lock()
	mb.conversations[conversationID] = conversation
	mb.mutex.Unlock()

	return conversation, nil
}
