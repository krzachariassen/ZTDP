package domain

import (
	"time"

	"github.com/ztdp/orchestrator/internal/planning/domain"
)

// AgentMessage represents a message sent to an agent
type AgentMessage struct {
	ID         string                 `json:"id"`
	Type       MessageType            `json:"type"`
	AgentID    string                 `json:"agent_id"`
	StepID     string                 `json:"step_id,omitempty"`
	PlanID     string                 `json:"plan_id,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	ReplyTo    string                 `json:"reply_to,omitempty"`
}

// MessageType represents the type of message being sent
type MessageType string

const (
	MessageTypeExecutionStep MessageType = "execution_step"
	MessageTypeStatusUpdate  MessageType = "status_update"
	MessageTypeHealthCheck   MessageType = "health_check"
	MessageTypeCancel        MessageType = "cancel"
)

// StatusUpdate represents a status update from an agent back to the orchestrator
type StatusUpdate struct {
	StepID    string                     `json:"step_id"`
	PlanID    string                     `json:"plan_id"`
	AgentID   string                     `json:"agent_id"`
	Status    domain.ExecutionPlanStatus `json:"status"`
	Result    map[string]interface{}     `json:"result,omitempty"`
	Error     string                     `json:"error,omitempty"`
	Timestamp time.Time                  `json:"timestamp"`
}

// AgentInfo represents agent information for routing
type AgentInfo struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Status       string   `json:"status"`
	Capabilities []string `json:"capabilities"`
	Endpoint     string   `json:"endpoint"`
}

// RoutingResult represents the result of a routing operation
type RoutingResult struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id"`
	Error     string `json:"error,omitempty"`
}
