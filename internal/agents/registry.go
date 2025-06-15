package agents

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/krzachariassen/ZTDP/internal/events"
)

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

	status := agent.GetStatus()
	agentID := status.ID

	// Store agent
	r.agents[agentID] = agent

	// Index capabilities
	for _, capability := range agent.GetCapabilities() {
		if r.capabilities[capability.Name] == nil {
			r.capabilities[capability.Name] = []string{}
		}
		r.capabilities[capability.Name] = append(r.capabilities[capability.Name], agentID)
	}

	return nil
}

// UnregisterAgent removes an agent from the registry
func (r *InMemoryAgentRegistry) UnregisterAgent(ctx context.Context, agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	// Remove from capability index
	for _, capability := range agent.GetCapabilities() {
		agentIDs := r.capabilities[capability.Name]
		for i, id := range agentIDs {
			if id == agentID {
				r.capabilities[capability.Name] = append(agentIDs[:i], agentIDs[i+1:]...)
				break
			}
		}
	}

	// Remove agent
	delete(r.agents, agentID)
	return nil
}

// FindAgentsByCapability finds agents that have a specific capability
func (r *InMemoryAgentRegistry) FindAgentsByCapability(ctx context.Context, capability string) ([]AgentStatus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agentIDs, exists := r.capabilities[capability]
	if !exists {
		return []AgentStatus{}, nil
	}

	var statuses []AgentStatus
	for _, agentID := range agentIDs {
		if agent, exists := r.agents[agentID]; exists {
			statuses = append(statuses, agent.GetStatus())
		}
	}

	return statuses, nil
}

// FindAgentByID finds a specific agent by ID
func (r *InMemoryAgentRegistry) FindAgentByID(ctx context.Context, agentID string) (AgentInterface, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}

	return agent, nil
}

// ListAllAgents returns all registered agents
func (r *InMemoryAgentRegistry) ListAllAgents(ctx context.Context) ([]AgentStatus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var statuses []AgentStatus
	for _, agent := range r.agents {
		statuses = append(statuses, agent.GetStatus())
	}

	return statuses, nil
}

// GetAvailableCapabilities returns all available capabilities
func (r *InMemoryAgentRegistry) GetAvailableCapabilities(ctx context.Context) ([]AgentCapability, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var capabilities []AgentCapability
	seen := make(map[string]bool)

	for _, agent := range r.agents {
		for _, cap := range agent.GetCapabilities() {
			if !seen[cap.Name] {
				capabilities = append(capabilities, cap)
				seen[cap.Name] = true
			}
		}
	}

	return capabilities, nil
}

// GetAgentHealth returns health status for a specific agent
func (r *InMemoryAgentRegistry) GetAgentHealth(ctx context.Context, agentID string) (HealthStatus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return HealthStatus{}, fmt.Errorf("agent %s not found", agentID)
	}

	return agent.Health(), nil
}

// SimpleAgentCoordinator implements AgentCoordinator
type SimpleAgentCoordinator struct {
	registry AgentRegistry
	eventBus *events.EventBus
}

// NewAgentCoordinator creates a new agent coordinator
func NewAgentCoordinator(registry AgentRegistry, eventBus *events.EventBus) AgentCoordinator {
	return &SimpleAgentCoordinator{
		registry: registry,
		eventBus: eventBus,
	}
}

// ResolveIntent analyzes intent and finds matching agents and capabilities
func (c *SimpleAgentCoordinator) ResolveIntent(ctx context.Context, intent string) ([]AgentStatus, string, error) {
	// Get all available capabilities
	allCapabilities, err := c.registry.GetAvailableCapabilities(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get capabilities: %w", err)
	}

	// Simple intent matching - in real implementation this would use AI
	intentLower := strings.ToLower(intent)
	var matchedCapability string
	var bestScore int

	for _, capability := range allCapabilities {
		score := 0
		for _, intentPattern := range capability.Intents {
			for _, word := range strings.Fields(strings.ToLower(intentPattern)) {
				if strings.Contains(intentLower, word) {
					score++
				}
			}
		}

		if score > bestScore {
			bestScore = score
			matchedCapability = capability.Name
		}
	}

	if matchedCapability == "" {
		return nil, "", fmt.Errorf("no matching capability found for intent: %s", intent)
	}

	// Find agents with this capability
	agents, err := c.registry.FindAgentsByCapability(ctx, matchedCapability)
	if err != nil {
		return nil, "", fmt.Errorf("failed to find agents for capability %s: %w", matchedCapability, err)
	}

	return agents, matchedCapability, nil
}

// RouteIntent routes an intent to appropriate agents
func (c *SimpleAgentCoordinator) RouteIntent(ctx context.Context, intent string, payload map[string]interface{}) (*events.Event, error) {
	agents, capability, err := c.ResolveIntent(ctx, intent)
	if err != nil {
		return nil, err
	}

	if len(agents) == 0 {
		return nil, fmt.Errorf("no agents available for capability: %s", capability)
	}

	// Route to first available agent (simple strategy)
	targetAgent := agents[0]

	event := &events.Event{
		Type:    events.EventTypeRequest,
		Source:  "coordinator",
		Subject: intent,
		Payload: payload,
	}

	return c.SendToAgent(ctx, targetAgent.ID, event)
}

// SendToAgent sends an event to a specific agent
func (c *SimpleAgentCoordinator) SendToAgent(ctx context.Context, targetAgentID string, event *events.Event) (*events.Event, error) {
	agent, err := c.registry.FindAgentByID(ctx, targetAgentID)
	if err != nil {
		return nil, fmt.Errorf("target agent not found: %w", err)
	}

	return agent.ProcessEvent(ctx, event)
}

// BroadcastToCapability broadcasts an event to all agents with a capability
func (c *SimpleAgentCoordinator) BroadcastToCapability(ctx context.Context, capability string, event *events.Event) ([]*events.Event, error) {
	agents, err := c.registry.FindAgentsByCapability(ctx, capability)
	if err != nil {
		return nil, fmt.Errorf("failed to find agents for capability %s: %w", capability, err)
	}

	var responses []*events.Event
	for _, agentStatus := range agents {
		agent, err := c.registry.FindAgentByID(ctx, agentStatus.ID)
		if err != nil {
			continue // Skip unavailable agents
		}

		response, err := agent.ProcessEvent(ctx, event)
		if err != nil {
			continue // Skip failed agents
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// RequestFromAgent sends a request to an agent (placeholder for timeout support)
func (c *SimpleAgentCoordinator) RequestFromAgent(ctx context.Context, targetAgentID string, event *events.Event, timeout time.Duration) (*events.Event, error) {
	// For now, just delegate to SendToAgent (timeout would be implemented with context)
	return c.SendToAgent(ctx, targetAgentID, event)
}
