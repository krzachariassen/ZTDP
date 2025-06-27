package domain

import "context"

// AgentRepository defines the interface for agent persistence
type AgentRepository interface {
	// Create a new agent
	Create(ctx context.Context, agent *Agent) error

	// Get agent by ID
	GetByID(ctx context.Context, id string) (*Agent, error)

	// Get all agents
	GetAll(ctx context.Context) ([]*Agent, error)

	// Get agents by status
	GetByStatus(ctx context.Context, status AgentStatus) ([]*Agent, error)

	// Get agents with specific capability
	GetByCapability(ctx context.Context, capabilityName string) ([]*Agent, error)

	// Update an existing agent
	Update(ctx context.Context, agent *Agent) error

	// Delete an agent
	Delete(ctx context.Context, id string) error

	// Update agent status
	UpdateStatus(ctx context.Context, id string, status AgentStatus) error

	// Update last seen timestamp
	UpdateLastSeen(ctx context.Context, id string) error
}
