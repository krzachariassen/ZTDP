package messaging

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ztdp/orchestrator/internal/graph"
	"github.com/ztdp/orchestrator/internal/logging"
)

// AIMessageBus provides natural language communication between AI, agents, and users
// It builds on top of the base MessageBus to provide AI-native conversational patterns
type AIMessageBus interface {
	// AI-to-Agent communication
	SendToAgent(ctx context.Context, msg *AIToAgentMessage) error

	// Agent-to-AI communication (clarifications, questions)
	SendToAI(ctx context.Context, msg *AgentToAIMessage) error

	// Agent-to-Agent communication (AI-mediated)
	SendBetweenAgents(ctx context.Context, msg *AgentToAgentMessage) error

	// User-to-AI communication
	SendUserToAI(ctx context.Context, msg *UserToAIMessage) error

	// Subscribe to conversations by participant
	Subscribe(ctx context.Context, participantID string) (<-chan *Message, error)

	// Get conversation history from graph
	GetConversationHistory(ctx context.Context, correlationID string) ([]*Message, error)
}

// AIToAgentMessage represents AI instructions to an agent
type AIToAgentMessage struct {
	AgentID       string                 `json:"agent_id"`
	Content       string                 `json:"content"`
	Intent        string                 `json:"intent"`
	CorrelationID string                 `json:"correlation_id"`
	Context       map[string]interface{} `json:"context"`
	Timeout       time.Duration          `json:"timeout,omitempty"`
}

// AgentToAIMessage represents agent communication to AI
type AgentToAIMessage struct {
	AgentID       string                 `json:"agent_id"`
	Content       string                 `json:"content"`
	MessageType   MessageType            `json:"message_type"`
	CorrelationID string                 `json:"correlation_id"`
	Context       map[string]interface{} `json:"context"`
	NeedsHelp     bool                   `json:"needs_help"`
}

// AgentToAgentMessage represents agent-to-agent communication (AI mediated)
type AgentToAgentMessage struct {
	FromAgentID   string                 `json:"from_agent_id"`
	ToAgentID     string                 `json:"to_agent_id"`
	Content       string                 `json:"content"`
	CorrelationID string                 `json:"correlation_id"`
	Context       map[string]interface{} `json:"context"`
	Purpose       string                 `json:"purpose"`
}

// UserToAIMessage represents user requests to AI
type UserToAIMessage struct {
	UserID        string                 `json:"user_id"`
	Content       string                 `json:"content"`
	CorrelationID string                 `json:"correlation_id"`
	Context       map[string]interface{} `json:"context"`
	Intent        string                 `json:"intent,omitempty"`
	SessionID     string                 `json:"session_id,omitempty"`
}

// AIMessageBusImpl implements the AI message bus
type AIMessageBusImpl struct {
	messageBus MessageBus
	graph      graph.Graph
	logger     logging.Logger
}

// NewAIMessageBus creates a new AI message bus
func NewAIMessageBus(messageBus MessageBus, graph graph.Graph, logger logging.Logger) AIMessageBus {
	return &AIMessageBusImpl{
		messageBus: messageBus,
		graph:      graph,
		logger:     logger,
	}
}

// SendToAgent sends AI instructions to an agent
func (bus *AIMessageBusImpl) SendToAgent(ctx context.Context, msg *AIToAgentMessage) error {
	bus.logger.Info("🤖➡️🤖 AI emitting instruction to agent",
		"event_type", "ai_to_agent",
		"agent_id", msg.AgentID,
		"correlation_id", msg.CorrelationID,
		"intent", msg.Intent,
		"content_length", len(msg.Content),
		"has_context", len(msg.Context) > 0)

	// Convert to generic message
	message := &Message{
		ID:            uuid.New().String(),
		CorrelationID: msg.CorrelationID,
		FromID:        "ai-orchestrator",
		ToID:          msg.AgentID,
		Content:       msg.Content,
		MessageType:   MessageTypeAIToAgent,
		Metadata:      msg.Context,
		Timestamp:     time.Now(),
	}

	bus.logger.Debug("📦 Message details",
		"message_id", message.ID,
		"from", message.FromID,
		"to", message.ToID,
		"content", message.Content)

	// Store in graph for conversation history
	if err := bus.storeMessageInGraph(ctx, message); err != nil {
		bus.logger.Error("❌ Failed to store AI-to-agent message in graph", err)
	}

	// Send via message bus
	if err := bus.messageBus.SendMessage(ctx, message); err != nil {
		bus.logger.Error("❌ Failed to emit AI-to-agent event", err,
			"agent_id", msg.AgentID,
			"correlation_id", msg.CorrelationID)
		return fmt.Errorf("failed to send AI message to agent %s: %w", msg.AgentID, err)
	}

	bus.logger.Info("✅ AI-to-agent event successfully emitted",
		"agent_id", msg.AgentID,
		"correlation_id", msg.CorrelationID,
		"message_id", message.ID,
		"intent", msg.Intent)

	return nil
}

// SendToAI sends agent questions/updates to AI
func (bus *AIMessageBusImpl) SendToAI(ctx context.Context, msg *AgentToAIMessage) error {
	// Convert to generic message
	message := &Message{
		ID:            uuid.New().String(),
		CorrelationID: msg.CorrelationID,
		FromID:        msg.AgentID,
		ToID:          "ai-orchestrator",
		Content:       msg.Content,
		MessageType:   msg.MessageType,
		Metadata:      msg.Context,
		Timestamp:     time.Now(),
	}

	// Store in graph for AI context and learning
	if err := bus.storeMessageInGraph(ctx, message); err != nil {
		bus.logger.Error("Failed to store agent-to-AI message in graph", err)
	}

	// Send to AI orchestrator
	if err := bus.messageBus.SendMessage(ctx, message); err != nil {
		return fmt.Errorf("failed to send agent message to AI: %w", err)
	}

	bus.logger.Info("Agent message sent to AI",
		"agent_id", msg.AgentID,
		"correlation_id", msg.CorrelationID,
		"needs_help", msg.NeedsHelp)

	return nil
}

// SendBetweenAgents handles agent-to-agent communication (AI mediated)
func (bus *AIMessageBusImpl) SendBetweenAgents(ctx context.Context, msg *AgentToAgentMessage) error {
	// Convert to generic message
	message := &Message{
		ID:            uuid.New().String(),
		CorrelationID: msg.CorrelationID,
		FromID:        msg.FromAgentID,
		ToID:          msg.ToAgentID,
		Content:       msg.Content,
		MessageType:   MessageTypeAgentToAgent,
		Metadata:      msg.Context,
		Timestamp:     time.Now(),
	}

	// Store in graph for conversation history
	if err := bus.storeMessageInGraph(ctx, message); err != nil {
		bus.logger.Error("Failed to store agent-to-agent message in graph", err)
	}

	// Send via message bus (AI mediates routing)
	if err := bus.messageBus.SendMessage(ctx, message); err != nil {
		return fmt.Errorf("failed to send message from agent %s to %s: %w",
			msg.FromAgentID, msg.ToAgentID, err)
	}

	bus.logger.Info("Agent-to-agent message sent",
		"from_agent", msg.FromAgentID,
		"to_agent", msg.ToAgentID,
		"correlation_id", msg.CorrelationID,
		"purpose", msg.Purpose)

	return nil
}

// SendUserToAI sends user requests to AI
func (bus *AIMessageBusImpl) SendUserToAI(ctx context.Context, msg *UserToAIMessage) error {
	// Convert to generic message
	message := &Message{
		ID:            uuid.New().String(),
		CorrelationID: msg.CorrelationID,
		FromID:        msg.UserID,
		ToID:          "ai-orchestrator",
		Content:       msg.Content,
		MessageType:   MessageTypeRequest,
		Metadata:      msg.Context,
		Timestamp:     time.Now(),
	}

	// Store in graph for conversation history
	if err := bus.storeMessageInGraph(ctx, message); err != nil {
		bus.logger.Error("Failed to store user-to-AI message in graph", err)
	}

	// Send to AI orchestrator
	if err := bus.messageBus.SendMessage(ctx, message); err != nil {
		return fmt.Errorf("failed to send user message to AI: %w", err)
	}

	bus.logger.Info("User message sent to AI",
		"user_id", msg.UserID,
		"correlation_id", msg.CorrelationID,
		"intent", msg.Intent)

	return nil
}

// Subscribe to conversations by participant
func (bus *AIMessageBusImpl) Subscribe(ctx context.Context, participantID string) (<-chan *Message, error) {
	return bus.messageBus.Subscribe(ctx, participantID)
}

// GetConversationHistory retrieves conversation history from graph
func (bus *AIMessageBusImpl) GetConversationHistory(ctx context.Context, correlationID string) ([]*Message, error) {
	// Use graph to retrieve conversation history
	// For now, return the message bus history
	return bus.messageBus.GetConversationHistory(ctx, correlationID)
}

// storeMessageInGraph stores a message in the graph for persistence and AI learning
func (bus *AIMessageBusImpl) storeMessageInGraph(ctx context.Context, message *Message) error {
	// For now, we'll just log that we're storing in graph
	// In full implementation, this would use graph operations to persist conversation context
	bus.logger.Debug("Storing message in graph",
		"message_id", message.ID,
		"correlation_id", message.CorrelationID,
		"from", message.FromID,
		"to", message.ToID)

	// TODO: Implement actual graph storage when graph operations are defined
	return nil
}
