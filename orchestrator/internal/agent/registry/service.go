package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ztdp/orchestrator/internal/agent/domain"
	"github.com/ztdp/orchestrator/internal/graph"
	"github.com/ztdp/orchestrator/internal/logging"
)

// Ensure Service implements AgentRegistry interface
var _ domain.AgentRegistry = (*Service)(nil)

// Service handles agent registry operations using graph storage
type Service struct {
	graph  graph.Graph
	logger logging.Logger
}

// NewService creates a new registry service
func NewService(g graph.Graph, logger logging.Logger) *Service {
	return &Service{
		graph:  g,
		logger: logger,
	}
}

// RegisterAgent registers a new agent
func (s *Service) RegisterAgent(ctx context.Context, agent *domain.Agent) error {
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
		agent.Status = domain.AgentStatusOnline
	}

	// Serialize metadata to JSON string for Neo4j storage
	var metadataJSON string
	if len(agent.Metadata) > 0 {
		if metadataBytes, err := json.Marshal(agent.Metadata); err == nil {
			metadataJSON = string(metadataBytes)
		}
	}

	// Serialize capabilities
	var capabilitiesJSON string
	if len(agent.Capabilities) > 0 {
		if capBytes, err := json.Marshal(agent.Capabilities); err == nil {
			capabilitiesJSON = string(capBytes)
		}
	}

	properties := map[string]interface{}{
		"name":         agent.Name,
		"description":  agent.Description,
		"status":       string(agent.Status),
		"capabilities": capabilitiesJSON,
		"last_seen":    agent.LastSeen,
		"metadata":     metadataJSON,
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

// UnregisterAgent removes an agent from the registry
func (s *Service) UnregisterAgent(ctx context.Context, agentID string) error {
	if agentID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	// Delete the agent node from the graph
	err := s.graph.DeleteNode(ctx, "agent", agentID)
	if err != nil {
		return fmt.Errorf("failed to delete agent from graph: %w", err)
	}

	if s.logger != nil {
		s.logger.Info("Agent unregistered successfully", "agent_id", agentID)
	}

	return nil
}

// GetAgent retrieves an agent by ID
func (s *Service) GetAgent(ctx context.Context, agentID string) (*domain.Agent, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent ID cannot be empty")
	}

	nodeData, err := s.graph.GetNode(ctx, "agent", agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	if nodeData == nil {
		return nil, fmt.Errorf("agent not found")
	}

	return s.nodeToAgent(agentID, nodeData)
}

// GetAllAgents retrieves all registered agents
func (s *Service) GetAllAgents(ctx context.Context) ([]*domain.Agent, error) {
	nodes, err := s.graph.QueryNodes(ctx, "agent", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query agents: %w", err)
	}

	var agents []*domain.Agent
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
		s.logger.Debug("Retrieved all agents", "count", len(agents))
	}

	return agents, nil
}

// GetAgentsByStatus retrieves agents with a specific status
func (s *Service) GetAgentsByStatus(ctx context.Context, status domain.AgentStatus) ([]*domain.Agent, error) {
	filters := map[string]interface{}{
		"status": string(status),
	}

	nodes, err := s.graph.QueryNodes(ctx, "agent", filters)
	if err != nil {
		return nil, fmt.Errorf("failed to query agents by status: %w", err)
	}

	var agents []*domain.Agent
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
		s.logger.Debug("Found agents by status", "status", status, "count", len(agents))
	}

	return agents, nil
}

// GetAgentsByCapability finds agents with a specific capability
func (s *Service) GetAgentsByCapability(ctx context.Context, capability string) ([]*domain.Agent, error) {
	if capability == "" {
		return nil, fmt.Errorf("capability cannot be empty")
	}

	// Get all agents and filter by capability
	nodes, err := s.graph.QueryNodes(ctx, "agent", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query agents: %w", err)
	}

	var agents []*domain.Agent
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
func (s *Service) UpdateAgentStatus(ctx context.Context, agentID string, status domain.AgentStatus) error {
	if agentID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	properties := map[string]interface{}{
		"status":     string(status),
		"updated_at": time.Now(),
	}

	err := s.graph.UpdateNode(ctx, "agent", agentID, properties)
	if err != nil {
		return fmt.Errorf("failed to update agent status: %w", err)
	}

	if s.logger != nil {
		s.logger.Info("Agent status updated", "agent_id", agentID, "status", status)
	}

	return nil
}

// UpdateAgentLastSeen updates the last seen timestamp for an agent
func (s *Service) UpdateAgentLastSeen(ctx context.Context, agentID string) error {
	if agentID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	properties := map[string]interface{}{
		"last_seen":  time.Now(),
		"updated_at": time.Now(),
	}

	err := s.graph.UpdateNode(ctx, "agent", agentID, properties)
	if err != nil {
		return fmt.Errorf("failed to update agent last seen: %w", err)
	}

	if s.logger != nil {
		s.logger.Debug("Agent last seen updated", "agent_id", agentID)
	}

	return nil
}

// IsAgentHealthy checks if an agent is healthy and responsive
func (s *Service) IsAgentHealthy(ctx context.Context, agentID string) (bool, error) {
	agent, err := s.GetAgent(ctx, agentID)
	if err != nil {
		return false, err
	}

	// Consider agent healthy if it's online and was seen recently (within 5 minutes)
	if agent.Status != domain.AgentStatusOnline {
		return false, nil
	}

	if time.Since(agent.LastSeen) > 5*time.Minute {
		return false, nil
	}

	return true, nil
}

// Helper methods

// nodeToAgent converts a graph node to an Agent domain object
func (s *Service) nodeToAgent(agentID string, nodeData map[string]interface{}) (*domain.Agent, error) {
	agent := &domain.Agent{
		ID: agentID,
	}

	if name, ok := nodeData["name"].(string); ok {
		agent.Name = name
	}

	if description, ok := nodeData["description"].(string); ok {
		agent.Description = description
	}

	if status, ok := nodeData["status"].(string); ok {
		agent.Status = domain.AgentStatus(status)
	}

	// Handle time fields
	if lastSeenTime, ok := nodeData["last_seen"].(time.Time); ok {
		agent.LastSeen = lastSeenTime
	} else if lastSeenStr, ok := nodeData["last_seen"].(string); ok {
		if lastSeen, err := time.Parse(time.RFC3339, lastSeenStr); err == nil {
			agent.LastSeen = lastSeen
		}
	}

	if createdAtTime, ok := nodeData["created_at"].(time.Time); ok {
		agent.CreatedAt = createdAtTime
	}

	if updatedAtTime, ok := nodeData["updated_at"].(time.Time); ok {
		agent.UpdatedAt = updatedAtTime
	}

	// Parse capabilities JSON
	if capabilitiesJSON, ok := nodeData["capabilities"].(string); ok && capabilitiesJSON != "" {
		var capabilities []domain.AgentCapability
		if err := json.Unmarshal([]byte(capabilitiesJSON), &capabilities); err == nil {
			agent.Capabilities = capabilities
		}
	}

	// Parse metadata JSON
	if metadataJSON, ok := nodeData["metadata"].(string); ok && metadataJSON != "" {
		var metadataInterface map[string]interface{}
		if err := json.Unmarshal([]byte(metadataJSON), &metadataInterface); err == nil {
			// Convert to map[string]string
			agent.Metadata = make(map[string]string)
			for k, v := range metadataInterface {
				if strValue, ok := v.(string); ok {
					agent.Metadata[k] = strValue
				}
			}
		}
	}

	return agent, nil
}

// hasCapability checks if an agent has a specific capability
func (s *Service) hasCapability(agent *domain.Agent, capability string) bool {
	for _, cap := range agent.Capabilities {
		if cap.Name == capability {
			return true
		}
	}
	return false
}
