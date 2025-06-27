package messaging

import (
	"time"
)

// Message represents a conversational message in the system
type Message struct {
	ID            string                 `json:"id"`
	CorrelationID string                 `json:"correlation_id"`
	FromID        string                 `json:"from_id"`
	ToID          string                 `json:"to_id"`
	Content       string                 `json:"content"`
	MessageType   MessageType            `json:"message_type"`
	Metadata      map[string]interface{} `json:"metadata"`
	Timestamp     time.Time              `json:"timestamp"`
}

// MessageType defines the type of message
type MessageType string

const (
	MessageTypeRequest       MessageType = "request"
	MessageTypeResponse      MessageType = "response"
	MessageTypeClarification MessageType = "clarification"
	MessageTypeNotification  MessageType = "notification"
	MessageTypeAgentToAgent  MessageType = "agent_to_agent"
	MessageTypeAIToAgent     MessageType = "ai_to_agent"
	MessageTypeAgentToAI     MessageType = "agent_to_ai"
)

// ConversationContext represents the context of a conversation
type ConversationContext struct {
	ConversationID string                 `json:"conversation_id"`
	Participants   []string               `json:"participants"`
	Context        map[string]interface{} `json:"context"`
	StartTime      time.Time              `json:"start_time"`
	LastActivity   time.Time              `json:"last_activity"`
}
