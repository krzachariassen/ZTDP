package application

import (
	"context"
	"fmt"

	"github.com/ztdp/orchestrator/internal/agent/domain"
)

// AgentService provides application-level agent operations
type AgentService struct {
	repository domain.AgentRepository
}

// NewAgentService creates a new agent service
func NewAgentService(repo domain.AgentRepository) *AgentService {
	return &AgentService{
		repository: repo,
	}
}

// RegisterAgent registers a new agent in the system
func (s *AgentService) RegisterAgent(ctx context.Context, agent *domain.Agent) error {
	if agent == nil {
		return fmt.Errorf("agent cannot be nil")
	}

	// Validate agent before registration
	if err := agent.Validate(); err != nil {
		return fmt.Errorf("agent validation failed: %w", err)
	}

	// Check if agent already exists
	existing, err := s.repository.GetByID(ctx, agent.ID)
	if err == nil && existing != nil {
		return fmt.Errorf("agent with ID %s already exists", agent.ID)
	}

	// Create the agent
	if err := s.repository.Create(ctx, agent); err != nil {
		return fmt.Errorf("failed to register agent: %w", err)
	}

	return nil
}

// DiscoverAgentsByCapability finds agents that have a specific capability
func (s *AgentService) DiscoverAgentsByCapability(ctx context.Context, capabilityName string) ([]*domain.Agent, error) {
	if capabilityName == "" {
		return nil, fmt.Errorf("capability name cannot be empty")
	}

	agents, err := s.repository.GetByCapability(ctx, capabilityName)
	if err != nil {
		return nil, fmt.Errorf("failed to discover agents by capability: %w", err)
	}

	// Filter by online status
	var availableAgents []*domain.Agent
	for _, agent := range agents {
		if agent.Status == domain.AgentStatusOnline {
			availableAgents = append(availableAgents, agent)
		}
	}

	return availableAgents, nil
}

// GetAvailableAgents returns all online agents
func (s *AgentService) GetAvailableAgents(ctx context.Context) ([]*domain.Agent, error) {
	agents, err := s.repository.GetByStatus(ctx, domain.AgentStatusOnline)
	if err != nil {
		return nil, fmt.Errorf("failed to get available agents: %w", err)
	}

	return agents, nil
}

// UpdateAgentStatus updates the status of an agent and its last seen timestamp
func (s *AgentService) UpdateAgentStatus(ctx context.Context, agentID string, status domain.AgentStatus) error {
	if agentID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	if !status.IsValid() {
		return fmt.Errorf("invalid status: %s", status)
	}

	// Update status
	if err := s.repository.UpdateStatus(ctx, agentID, status); err != nil {
		return fmt.Errorf("failed to update agent status: %w", err)
	}

	// Update last seen timestamp
	if err := s.repository.UpdateLastSeen(ctx, agentID); err != nil {
		return fmt.Errorf("failed to update agent last seen: %w", err)
	}

	return nil
}

// DiscoverAgentsByCapabilities discovers agents that match the required capabilities
// This method finds agents that can handle the specified capabilities
func (s *AgentService) DiscoverAgentsByCapabilities(ctx context.Context, requiredCapabilities []string) ([]*domain.Agent, error) {
	if len(requiredCapabilities) == 0 {
		return nil, fmt.Errorf("required capabilities cannot be empty")
	}

	var matchingAgents []*domain.Agent
	agentMap := make(map[string]*domain.Agent) // Prevent duplicates

	// For each required capability, find suitable agents
	for _, capability := range requiredCapabilities {
		agents, err := s.DiscoverAgentsByCapability(ctx, capability)
		if err != nil {
			return nil, fmt.Errorf("failed to discover agents for capability %s: %w", capability, err)
		}

		if len(agents) == 0 {
			return nil, fmt.Errorf("no available agents found for capability: %s", capability)
		}

		// Add the best agent for this capability (first online agent)
		for _, agent := range agents {
			if agent.Status == domain.AgentStatusOnline {
				if _, exists := agentMap[agent.ID]; !exists {
					agentMap[agent.ID] = agent
					matchingAgents = append(matchingAgents, agent)
				}
				break
			}
		}
	}

	if len(matchingAgents) == 0 {
		return nil, fmt.Errorf("no suitable agents found for required capabilities")
	}

	return matchingAgents, nil
}

// GetAgentByID retrieves an agent by its ID
func (s *AgentService) GetAgentByID(ctx context.Context, agentID string) (*domain.Agent, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent ID cannot be empty")
	}

	agent, err := s.repository.GetByID(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	return agent, nil
}

// ListAllAgents returns all agents in the system
func (s *AgentService) ListAllAgents(ctx context.Context) ([]*domain.Agent, error) {
	agents, err := s.repository.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list all agents: %w", err)
	}

	return agents, nil
}

// UnregisterAgent removes an agent from the system
func (s *AgentService) UnregisterAgent(ctx context.Context, agentID string) error {
	if agentID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	// Check if agent exists
	_, err := s.repository.GetByID(ctx, agentID)
	if err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	// Delete the agent
	if err := s.repository.Delete(ctx, agentID); err != nil {
		return fmt.Errorf("failed to unregister agent: %w", err)
	}

	return nil
}
