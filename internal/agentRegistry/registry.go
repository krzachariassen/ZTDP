package agentRegistry

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AgentInterface defines the contract that all agents must implement
// This is the registry's view of what an agent should be - pure registration/discovery
type AgentInterface interface {
	GetID() string
	GetStatus() AgentStatus
	GetCapabilities() []AgentCapability
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Health() HealthStatus
}

// AgentRegistry defines the interface for agent registration and discovery
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

// AgentStatus represents the current status of an agent
type AgentStatus struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Status       string                 `json:"status"`
	LastActivity time.Time              `json:"last_activity"`
	LoadFactor   float64                `json:"load_factor"`
	Version      string                 `json:"version"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// AgentCapability defines what an agent can do
type AgentCapability struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Intents     []string `json:"intents"`
	InputTypes  []string `json:"input_types"`
	OutputTypes []string `json:"output_types"`
	RoutingKeys []string `json:"routing_keys"`
	Version     string   `json:"version"`
}

// HealthStatus represents the health status of an agent
type HealthStatus struct {
	Healthy bool   `json:"healthy"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// InMemoryAgentRegistry implements AgentRegistry for development and testing
type InMemoryAgentRegistry struct {
	agents       map[string]AgentInterface
	capabilities map[string][]string // capability -> agent IDs
	mu           sync.RWMutex
}

// NewInMemoryAgentRegistry creates a new in-memory agent registry
func NewInMemoryAgentRegistry() AgentRegistry {
	return &InMemoryAgentRegistry{
		agents:       make(map[string]AgentInterface),
		capabilities: make(map[string][]string),
	}
}

// RegisterAgent registers an agent with the registry
func (r *InMemoryAgentRegistry) RegisterAgent(ctx context.Context, agent AgentInterface) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agentID := agent.GetID()
	if agentID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	// Check if agent already exists
	if _, exists := r.agents[agentID]; exists {
		return fmt.Errorf("agent with ID %s already registered", agentID)
	}

	// Register the agent
	r.agents[agentID] = agent

	// Register capabilities
	capabilities := agent.GetCapabilities()
	for _, cap := range capabilities {
		for _, intent := range cap.Intents {
			r.capabilities[intent] = append(r.capabilities[intent], agentID)
		}
	}

	return nil
}

// UnregisterAgent removes an agent from the registry
func (r *InMemoryAgentRegistry) UnregisterAgent(ctx context.Context, agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent with ID %s not found", agentID)
	}

	// Remove capabilities
	capabilities := agent.GetCapabilities()
	for _, cap := range capabilities {
		for _, intent := range cap.Intents {
			agents := r.capabilities[intent]
			for i, id := range agents {
				if id == agentID {
					r.capabilities[intent] = append(agents[:i], agents[i+1:]...)
					break
				}
			}
		}
	}

	// Remove agent
	delete(r.agents, agentID)
	return nil
}

// GetAgent returns a specific agent by ID
func (r *InMemoryAgentRegistry) GetAgent(agentID string) (AgentInterface, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent with ID %s not found", agentID)
	}

	return agent, nil
}

// ListAgents returns all registered agents
func (r *InMemoryAgentRegistry) ListAgents() []AgentInterface {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]AgentInterface, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}

	return agents
}

// DiscoverByCapability finds agents that have a specific capability
func (r *InMemoryAgentRegistry) DiscoverByCapability(capability string) []AgentInterface {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var agents []AgentInterface
	for _, agent := range r.agents {
		capabilities := agent.GetCapabilities()
		for _, cap := range capabilities {
			if cap.Name == capability {
				agents = append(agents, agent)
				break
			}
		}
	}

	return agents
}

// DiscoverByIntent finds agents that can handle a specific intent
func (r *InMemoryAgentRegistry) DiscoverByIntent(intent string) []AgentInterface {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agentIDs, exists := r.capabilities[intent]
	if !exists {
		return nil
	}

	agents := make([]AgentInterface, 0, len(agentIDs))
	for _, agentID := range agentIDs {
		if agent, exists := r.agents[agentID]; exists {
			agents = append(agents, agent)
		}
	}

	return agents
}

// FindAgentsByCapability finds all agents that have a specific capability  
func (r *InMemoryAgentRegistry) FindAgentsByCapability(ctx context.Context, capability string) ([]AgentStatus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var statuses []AgentStatus
	for _, agent := range r.agents {
		capabilities := agent.GetCapabilities()
		for _, cap := range capabilities {
			if cap.Name == capability {
				statuses = append(statuses, agent.GetStatus())
				break
			}
		}
	}

	return statuses, nil
}

// FindAgentByID finds an agent by its ID
func (r *InMemoryAgentRegistry) FindAgentByID(ctx context.Context, agentID string) (AgentInterface, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent with ID %s not found", agentID)
	}

	return agent, nil
}

// ListAllAgents returns the status of all registered agents
func (r *InMemoryAgentRegistry) ListAllAgents(ctx context.Context) ([]AgentStatus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	statuses := make([]AgentStatus, 0, len(r.agents))
	for _, agent := range r.agents {
		statuses = append(statuses, agent.GetStatus())
	}

	return statuses, nil
}

// GetAvailableCapabilities returns all available capabilities across all agents
func (r *InMemoryAgentRegistry) GetAvailableCapabilities(ctx context.Context) ([]AgentCapability, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	capabilityMap := make(map[string]AgentCapability)
	for _, agent := range r.agents {
		capabilities := agent.GetCapabilities()
		for _, capability := range capabilities {
			capabilityMap[capability.Name] = capability
		}
	}

	var allCapabilities []AgentCapability
	for _, capability := range capabilityMap {
		allCapabilities = append(allCapabilities, capability)
	}

	return allCapabilities, nil
}

// GetAgentHealth returns the health status of a specific agent
func (r *InMemoryAgentRegistry) GetAgentHealth(ctx context.Context, agentID string) (HealthStatus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return HealthStatus{}, fmt.Errorf("agent with ID %s not found", agentID)
	}

	return agent.Health(), nil
}
