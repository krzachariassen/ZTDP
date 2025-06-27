package domain

import "context"

// AgentRegistry defines the interface for agent registration and discovery
// This is different from AgentRepository - Registry is for service discovery, Repository is for persistence
type AgentRegistry interface {
	// RegisterAgent registers a new agent in the registry
	RegisterAgent(ctx context.Context, agent *Agent) error

	// UnregisterAgent removes an agent from the registry
	UnregisterAgent(ctx context.Context, agentID string) error

	// GetAgent retrieves an agent by ID
	GetAgent(ctx context.Context, agentID string) (*Agent, error)

	// GetAllAgents retrieves all registered agents
	GetAllAgents(ctx context.Context) ([]*Agent, error)

	// GetAgentsByStatus retrieves agents with a specific status
	GetAgentsByStatus(ctx context.Context, status AgentStatus) ([]*Agent, error)

	// GetAgentsByCapability finds agents with a specific capability
	GetAgentsByCapability(ctx context.Context, capability string) ([]*Agent, error)

	// UpdateAgentStatus updates an agent's status
	UpdateAgentStatus(ctx context.Context, agentID string, status AgentStatus) error

	// UpdateAgentLastSeen updates the last seen timestamp for an agent
	UpdateAgentLastSeen(ctx context.Context, agentID string) error

	// IsAgentHealthy checks if an agent is healthy and responsive
	IsAgentHealthy(ctx context.Context, agentID string) (bool, error)
}
