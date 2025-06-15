package agents

import (
	"context"
	"time"

	"github.com/krzachariassen/ZTDP/internal/events"
)

// AgentInterface defines the contract for all AI agents in the platform
// This is domain logic, not infrastructure
type AgentInterface interface {
	// Core Operations
	ProcessEvent(ctx context.Context, event *events.Event) (*events.Event, error)
	GetCapabilities() []AgentCapability
	GetStatus() AgentStatus

	// Lifecycle
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Health() HealthStatus
}

// AgentCapability describes what an agent can do
type AgentCapability struct {
	Name        string   `json:"name"`         // e.g., "policy_evaluation"
	Description string   `json:"description"`  // Human-readable description
	Intents     []string `json:"intents"`      // Natural language patterns it handles
	InputTypes  []string `json:"input_types"`  // Expected input data types
	OutputTypes []string `json:"output_types"` // Response data types
	Version     string   `json:"version"`      // Capability version
}

// AgentStatus represents current agent state
type AgentStatus struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`   // "platform", "policy", "deployment"
	Status       string                 `json:"status"` // "running", "idle", "busy", "error"
	LastActivity time.Time              `json:"last_activity"`
	LoadFactor   float64                `json:"load_factor"` // 0.0 to 1.0
	Version      string                 `json:"version"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// HealthStatus represents agent health
type HealthStatus struct {
	Healthy   bool                   `json:"healthy"`
	Status    string                 `json:"status"` // "healthy", "degraded", "unhealthy"
	Message   string                 `json:"message"`
	Checks    map[string]interface{} `json:"checks"`
	CheckedAt time.Time              `json:"checked_at"`
}

// AgentRegistry manages agent discovery and coordination
type AgentRegistry interface {
	// Registration
	RegisterAgent(ctx context.Context, agent AgentInterface) error
	UnregisterAgent(ctx context.Context, agentID string) error

	// Discovery
	FindAgentsByCapability(ctx context.Context, capability string) ([]AgentStatus, error)
	FindAgentByID(ctx context.Context, agentID string) (AgentInterface, error)
	ListAllAgents(ctx context.Context) ([]AgentStatus, error)

	// Capabilities
	GetAvailableCapabilities(ctx context.Context) ([]AgentCapability, error)

	// Health
	GetAgentHealth(ctx context.Context, agentID string) (HealthStatus, error)
}

// AgentEventHandler defines how agents handle events
type AgentEventHandler func(ctx context.Context, event *events.Event) (*events.Event, error)

// AgentCoordinator handles agent-to-agent communication
type AgentCoordinator interface {
	// Intent-based routing
	RouteIntent(ctx context.Context, intent string, payload map[string]interface{}) (*events.Event, error)
	ResolveIntent(ctx context.Context, intent string) ([]AgentStatus, string, error)

	// Direct agent communication
	SendToAgent(ctx context.Context, targetAgentID string, event *events.Event) (*events.Event, error)

	// Broadcast to multiple agents
	BroadcastToCapability(ctx context.Context, capability string, event *events.Event) ([]*events.Event, error)

	// Request-response patterns
	RequestFromAgent(ctx context.Context, targetAgentID string, event *events.Event, timeout time.Duration) (*events.Event, error)
}
