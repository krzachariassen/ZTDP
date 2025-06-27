package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/ztdp/orchestrator/internal/agent/domain"
)

// AgentService defines the interface for agent operations
type AgentService interface {
	GetAvailableAgents(ctx context.Context) ([]*domain.Agent, error)
	RegisterAgent(ctx context.Context, agent *domain.Agent) error
	DiscoverAgentsByCapability(ctx context.Context, capability string) ([]*domain.Agent, error)
	UpdateAgentStatus(ctx context.Context, agentID string, status domain.AgentStatus) error
}

// GraphExplorer handles agent discovery and context formatting for AI consumption
type GraphExplorer struct {
	agentService AgentService
}

// NewGraphExplorer creates a new GraphExplorer instance
func NewGraphExplorer(agentService AgentService) *GraphExplorer {
	return &GraphExplorer{
		agentService: agentService,
	}
}

// GetAgentContext retrieves all available agents and formats them for AI consumption
// Replaces the getAllAgents() functionality from the old orchestrator
func (g *GraphExplorer) GetAgentContext(ctx context.Context) (string, error) {
	agents, err := g.agentService.GetAvailableAgents(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get available agents: %w", err)
	}

	if len(agents) == 0 {
		return "No agents currently registered", nil
	}

	var context strings.Builder
	context.WriteString("Available agents:\n")

	for _, agent := range agents {
		context.WriteString(fmt.Sprintf("- %s (ID: %s, Status: %s)\n",
			agent.Name, agent.ID, string(agent.Status)))

		if len(agent.Capabilities) > 0 {
			capabilityNames := make([]string, len(agent.Capabilities))
			for i, cap := range agent.Capabilities {
				capabilityNames[i] = cap.Name
			}
			context.WriteString(fmt.Sprintf("  Capabilities: %s\n",
				strings.Join(capabilityNames, ", ")))
		}
	}

	return context.String(), nil
}

// FindCapableAgents finds agents with specific capabilities
func (g *GraphExplorer) FindCapableAgents(ctx context.Context, capabilities []string) ([]*domain.Agent, error) {
	var allAgents []*domain.Agent
	agentMap := make(map[string]*domain.Agent)

	// Find agents for each capability and deduplicate
	for _, capability := range capabilities {
		agents, err := g.agentService.DiscoverAgentsByCapability(ctx, capability)
		if err != nil {
			return nil, fmt.Errorf("failed to discover agents for capability %s: %w", capability, err)
		}

		for _, agent := range agents {
			if _, exists := agentMap[agent.ID]; !exists {
				agentMap[agent.ID] = agent
				allAgents = append(allAgents, agent)
			}
		}
	}

	return allAgents, nil
}
