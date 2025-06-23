package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/ztdp/orchestrator/internal/graph"
	"github.com/ztdp/orchestrator/internal/types"
)

// Service handles agent registry operations
type Service struct {
	graph  graph.Graph
	logger graph.Logger
}

// NewService creates a new registry service
func NewService(g graph.Graph, logger graph.Logger) *Service {
	return &Service{
		graph:  g,
		logger: logger,
	}
}

// RegisterAgent registers a new agent
func (s *Service) RegisterAgent(ctx context.Context, agent *types.Agent) error {
	if agent == nil {
		return fmt.Errorf("agent cannot be nil")
	}

	if agent.ID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	if agent.Name == "" {
		return fmt.Errorf("agent name cannot be empty")
	}

	// Set defaults
	if agent.Status == "" {
		agent.Status = types.AgentStatusActive
	}

	properties := map[string]interface{}{
		"name":         agent.Name,
		"type":         agent.Type,
		"status":       agent.Status,
		"capabilities": agent.Capabilities,
		"endpoint":     agent.Endpoint,
		"last_seen":    agent.LastSeen,
		"metadata":     agent.Metadata,
		"created_at":   time.Now(),
		"updated_at":   time.Now(),
	}

	err := s.graph.AddNode(ctx, "agent", agent.ID, properties)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("Failed to register agent", err, "agent_id", agent.ID)
		}
		return fmt.Errorf("failed to register agent: %w", err)
	}

	if s.logger != nil {
		s.logger.Info("Agent registered successfully", "agent_id", agent.ID, "name", agent.Name)
	}

	return nil
}

// GetAgentsByCapability finds agents with a specific capability
func (s *Service) GetAgentsByCapability(ctx context.Context, capability string) ([]*types.Agent, error) {
	if capability == "" {
		return nil, fmt.Errorf("capability cannot be empty")
	}

	// Get all agents and filter by capability
	nodes, err := s.graph.QueryNodes(ctx, "agent", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query agents: %w", err)
	}

	var agents []*types.Agent
	for _, nodeData := range nodes {
		agentID, ok := nodeData["id"].(string)
		if !ok {
			continue
		}

		agent, err := s.nodeToAgent(agentID, nodeData)
		if err != nil {
			if s.logger != nil {
				s.logger.Error("Failed to convert node to agent", err, "agent_id", agentID)
			}
			continue
		}

		// Check if agent has the required capability
		if s.hasCapability(agent, capability) {
			agents = append(agents, agent)
		}
	}

	if s.logger != nil {
		s.logger.Debug("Found agents by capability", "capability", capability, "count", len(agents))
	}

	return agents, nil
}

// UpdateAgentStatus updates an agent's status
func (s *Service) UpdateAgentStatus(ctx context.Context, agentID, status string) error {
	if agentID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	if status == "" {
		return fmt.Errorf("status cannot be empty")
	}

	// Validate status
	switch status {
	case types.AgentStatusActive, types.AgentStatusInactive, types.AgentStatusError:
		// Valid status
	default:
		return fmt.Errorf("invalid status: %s", status)
	}

	properties := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == types.AgentStatusActive {
		properties["last_seen"] = time.Now()
	}

	err := s.graph.UpdateNode(ctx, "agent", agentID, properties)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("Failed to update agent status", err, "agent_id", agentID, "status", status)
		}
		return fmt.Errorf("failed to update agent status: %w", err)
	}

	if s.logger != nil {
		s.logger.Info("Agent status updated", "agent_id", agentID, "status", status)
	}

	return nil
}

// GetActiveAgents returns all active agents
func (s *Service) GetActiveAgents(ctx context.Context) ([]*types.Agent, error) {
	filters := map[string]interface{}{
		"status": types.AgentStatusActive,
	}

	nodes, err := s.graph.QueryNodes(ctx, "agent", filters)
	if err != nil {
		return nil, fmt.Errorf("failed to query active agents: %w", err)
	}

	var agents []*types.Agent
	for _, nodeData := range nodes {
		agentID, ok := nodeData["id"].(string)
		if !ok {
			continue
		}

		agent, err := s.nodeToAgent(agentID, nodeData)
		if err != nil {
			if s.logger != nil {
				s.logger.Error("Failed to convert node to agent", err, "agent_id", agentID)
			}
			continue
		}

		agents = append(agents, agent)
	}

	if s.logger != nil {
		s.logger.Debug("Found active agents", "count", len(agents))
	}

	return agents, nil
}

// Helper methods

func (s *Service) nodeToAgent(agentID string, nodeData map[string]interface{}) (*types.Agent, error) {
	agent := &types.Agent{
		ID: agentID,
	}

	if name, ok := nodeData["name"].(string); ok {
		agent.Name = name
	}

	if agentType, ok := nodeData["type"].(string); ok {
		agent.Type = agentType
	}

	if status, ok := nodeData["status"].(string); ok {
		agent.Status = status
	}

	if endpoint, ok := nodeData["endpoint"].(string); ok {
		agent.Endpoint = endpoint
	}

	// Handle capabilities
	if capData, ok := nodeData["capabilities"]; ok {
		if caps, ok := capData.([]string); ok {
			agent.Capabilities = caps
		} else if capsInterface, ok := capData.([]interface{}); ok {
			agent.Capabilities = make([]string, len(capsInterface))
			for i, cap := range capsInterface {
				if capStr, ok := cap.(string); ok {
					agent.Capabilities[i] = capStr
				}
			}
		}
	}

	// Handle metadata
	if metadataStr, ok := nodeData["metadata"].(map[string]string); ok {
		agent.Metadata = metadataStr
	} else if metadataInterface, ok := nodeData["metadata"].(map[string]interface{}); ok {
		agent.Metadata = make(map[string]string)
		for key, value := range metadataInterface {
			if strValue, ok := value.(string); ok {
				agent.Metadata[key] = strValue
			}
		}
	}

	// Handle time fields
	if lastSeen, ok := nodeData["last_seen"].(time.Time); ok {
		agent.LastSeen = lastSeen
	}

	return agent, nil
}

func (s *Service) hasCapability(agent *types.Agent, capability string) bool {
	for _, cap := range agent.Capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}
