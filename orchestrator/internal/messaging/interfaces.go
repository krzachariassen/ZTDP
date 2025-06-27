package messaging

import (
	"context"
)

// MessageBus handles natural language communication between AI, agents, and users
// This is the event-driven messaging system for conversational orchestration
type MessageBus interface {
	// Send message to specific recipient
	SendMessage(ctx context.Context, message *Message) error

	// Subscribe to messages for a specific participant
	Subscribe(ctx context.Context, participantID string) (<-chan *Message, error)

	// Unsubscribe from messages
	Unsubscribe(ctx context.Context, participantID string) error

	// Publish message to multiple recipients (broadcast)
	PublishMessage(ctx context.Context, message *Message, recipients []string) error

	// Get conversation history
	GetConversationHistory(ctx context.Context, conversationID string) ([]*Message, error)

	// Create new conversation context
	CreateConversation(ctx context.Context, participants []string, context map[string]interface{}) (*ConversationContext, error)
}

// MessageHandler handles incoming messages
type MessageHandler interface {
	// Handle incoming message
	HandleMessage(ctx context.Context, message *Message) error

	// Get handler ID for routing
	GetHandlerID() string
}

// ClarificationRequest represents a request for clarification
type ClarificationRequest struct {
	RequestID     string                 `json:"request_id"`
	AgentID       string                 `json:"agent_id"`
	Question      string                 `json:"question"`
	Context       map[string]interface{} `json:"context"`
	CorrelationID string                 `json:"correlation_id"`
}

// ClarificationResponse represents the response to a clarification
type ClarificationResponse struct {
	RequestID     string                 `json:"request_id"`
	Answer        string                 `json:"answer"`
	Context       map[string]interface{} `json:"context"`
	CorrelationID string                 `json:"correlation_id"`
}
