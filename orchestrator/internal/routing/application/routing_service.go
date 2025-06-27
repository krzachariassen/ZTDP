package application

import (
	"context"

	"github.com/ztdp/orchestrator/internal/planning/domain"
)

// RoutingService defines the application service interface for AI-native agent routing
// This service routes messages between the AI orchestrator and agents using a messaging system
// The AI has already decided which agents to use - routing just delivers messages
type RoutingService interface {
	// RouteExecutionStep routes an execution step to the agent specified in the plan
	RouteExecutionStep(ctx context.Context, agentID string, step *ExecutionStep) error

	// RouteStatusUpdate routes status updates from agents back to the orchestrator
	RouteStatusUpdate(ctx context.Context, update *StatusUpdate) error

	// GetAvailableAgents retrieves the catalog of available agents for AI planning
	GetAvailableAgents(ctx context.Context) ([]*AgentInfo, error)
}

// MessageBus defines the interface for the messaging system
type MessageBus interface {
	// SendMessage sends a message to a specific agent or service
	SendMessage(ctx context.Context, targetID string, message *AgentMessage) error

	// GetMessagesForAgent gets messages sent to a specific agent (for testing)
	GetMessagesForAgent(agentID string) []*AgentMessage
}

// ExecutionStep represents a step to be executed by an agent
type ExecutionStep struct {
	ID         string                     `json:"id"`
	PlanID     string                     `json:"plan_id"`
	AgentID    string                     `json:"agent_id"`
	Action     string                     `json:"action"`
	Parameters map[string]interface{}     `json:"parameters"`
	Status     domain.ExecutionPlanStatus `json:"status"`
}

// StatusUpdate represents a status update from an agent
type StatusUpdate struct {
	StepID    string                     `json:"step_id"`
	AgentID   string                     `json:"agent_id"`
	Status    domain.ExecutionPlanStatus `json:"status"`
	Result    map[string]interface{}     `json:"result,omitempty"`
	Error     string                     `json:"error,omitempty"`
	Timestamp string                     `json:"timestamp"`
}

// AgentMessage represents a message sent through the messaging system
type AgentMessage struct {
	ID      string                 `json:"id"`
	Type    MessageType            `json:"type"`
	From    string                 `json:"from"`
	To      string                 `json:"to"`
	Payload map[string]interface{} `json:"payload"`
	SentAt  string                 `json:"sent_at"`
}

// MessageType represents the type of message being sent
type MessageType string

const (
	MessageTypeExecutionStep MessageType = "execution_step"
	MessageTypeStatusUpdate  MessageType = "status_update"
)

// AgentInfo represents agent information for AI planning
type AgentInfo struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Capabilities []string `json:"capabilities"`
	Status       string   `json:"status"`
}
