package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/ztdp/orchestrator/internal/agent/domain"
	"github.com/ztdp/orchestrator/internal/graph"
)

// GraphAgentRepository implements the AgentRepository interface using the graph backend
type GraphAgentRepository struct {
	graph graph.Graph
}

// NewGraphAgentRepository creates a new graph-based agent repository
func NewGraphAgentRepository(g graph.Graph) *GraphAgentRepository {
	return &GraphAgentRepository{
		graph: g,
	}
}

// Create persists a new agent to the graph
func (r *GraphAgentRepository) Create(ctx context.Context, agent *domain.Agent) error {
	if err := agent.Validate(); err != nil {
		return fmt.Errorf("invalid agent: %w", err)
	}

	// Convert domain model to graph data
	data := agent.ToMap()

	// Store in graph with proper node type
	nodeID := fmt.Sprintf("agent:%s", agent.ID)

	// Create the agent node
	if err := r.graph.AddNode(ctx, "agent", nodeID, data); err != nil {
		return fmt.Errorf("failed to create agent node: %w", err)
	}

	// Add capability relationships
	for _, capability := range agent.Capabilities {
		capabilityNodeID := fmt.Sprintf("capability:%s:%s", agent.ID, capability.Name)
		capabilityData := map[string]interface{}{
			"name":        capability.Name,
			"description": capability.Description,
			"parameters":  capability.Parameters,
		}

		// Create capability node
		if err := r.graph.AddNode(ctx, "capability", capabilityNodeID, capabilityData); err != nil {
			return fmt.Errorf("failed to create capability node: %w", err)
		}

		// Create relationship
		if err := r.graph.AddEdge(ctx, "agent", nodeID, "capability", capabilityNodeID, "HAS_CAPABILITY", nil); err != nil {
			return fmt.Errorf("failed to create capability relationship: %w", err)
		}
	}

	return nil
}

// GetByID retrieves an agent by its ID from the graph
func (r *GraphAgentRepository) GetByID(ctx context.Context, id string) (*domain.Agent, error) {
	if id == "" {
		return nil, fmt.Errorf("agent ID cannot be empty")
	}

	nodeID := fmt.Sprintf("agent:%s", id)

	// Get agent node data
	node, err := r.graph.GetNode(ctx, "agent", nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent node: %w", err)
	}

	if node == nil {
		return nil, fmt.Errorf("agent not found: %s", id)
	}

	// Get capabilities using graph traversal
	capabilities, err := r.getAgentCapabilities(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent capabilities: %w", err)
	}

	// Create a map with node properties and capabilities
	nodeData := make(map[string]interface{})
	for k, v := range node {
		nodeData[k] = v
	}
	nodeData["capabilities"] = capabilities

	// Convert from graph data to domain model
	agent, err := domain.AgentFromMap(nodeData)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize agent: %w", err)
	}

	return agent, nil
}

// GetAll retrieves all agents from the graph
func (r *GraphAgentRepository) GetAll(ctx context.Context) ([]*domain.Agent, error) {
	// Get all agent nodes
	nodes, err := r.graph.QueryNodes(ctx, "agent", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent nodes: %w", err)
	}

	agents := make([]*domain.Agent, 0, len(nodes))
	for _, node := range nodes {
		// Extract agent ID from node data
		agentID, ok := node["id"].(string)
		if !ok {
			continue // Skip nodes without valid ID
		}

		// Get capabilities for this agent
		nodeID := fmt.Sprintf("agent:%s", agentID)
		capabilities, err := r.getAgentCapabilities(ctx, nodeID)
		if err != nil {
			continue // Skip agents with capability errors
		}

		// Create map with node properties and capabilities
		nodeData := make(map[string]interface{})
		for k, v := range node {
			nodeData[k] = v
		}
		nodeData["capabilities"] = capabilities

		// Convert to domain model
		agent, err := domain.AgentFromMap(nodeData)
		if err != nil {
			continue // Skip invalid agents
		}

		agents = append(agents, agent)
	}

	return agents, nil
}

// GetByStatus retrieves agents by their status
func (r *GraphAgentRepository) GetByStatus(ctx context.Context, status domain.AgentStatus) ([]*domain.Agent, error) {
	// Get all agents and filter by status
	allAgents, err := r.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var filteredAgents []*domain.Agent
	for _, agent := range allAgents {
		if agent.Status == status {
			filteredAgents = append(filteredAgents, agent)
		}
	}

	return filteredAgents, nil
}

// GetByCapability retrieves agents that have a specific capability
func (r *GraphAgentRepository) GetByCapability(ctx context.Context, capabilityName string) ([]*domain.Agent, error) {
	// Get all agents and filter by capability
	allAgents, err := r.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var filteredAgents []*domain.Agent
	for _, agent := range allAgents {
		for _, capability := range agent.Capabilities {
			if capability.Name == capabilityName {
				filteredAgents = append(filteredAgents, agent)
				break
			}
		}
	}

	return filteredAgents, nil
}

// Update updates an existing agent in the graph
func (r *GraphAgentRepository) Update(ctx context.Context, agent *domain.Agent) error {
	if err := agent.Validate(); err != nil {
		return fmt.Errorf("invalid agent: %w", err)
	}

	nodeID := fmt.Sprintf("agent:%s", agent.ID)

	// Check if agent exists
	existing, err := r.graph.GetNode(ctx, "agent", nodeID)
	if err != nil {
		return fmt.Errorf("failed to check existing agent: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("agent not found: %s", agent.ID)
	}

	// Update agent node
	data := agent.ToMap()
	if err := r.graph.UpdateNode(ctx, "agent", nodeID, data); err != nil {
		return fmt.Errorf("failed to update agent node: %w", err)
	}

	// TODO: Handle capability updates (remove old, add new)
	// For now, we'll just update the main agent properties

	return nil
}

// Delete removes an agent from the graph
func (r *GraphAgentRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	nodeID := fmt.Sprintf("agent:%s", id)

	// Get and remove capability nodes and edges
	edges, err := r.graph.GetEdges(ctx, "agent", nodeID)
	if err == nil { // Don't fail if no edges found
		for _, edge := range edges {
			// Check if this is a HAS_CAPABILITY edge
			if edgeType, ok := edge["type"].(string); ok && edgeType == "HAS_CAPABILITY" {
				// Extract target capability node ID
				if targetID, ok := edge["target_id"].(string); ok {
					// Delete capability node
					if err := r.graph.DeleteNode(ctx, "capability", targetID); err != nil {
						// Log but don't fail
					}
				}
			}
		}
	}

	// Delete the agent node (this will also remove associated edges)
	if err := r.graph.DeleteNode(ctx, "agent", nodeID); err != nil {
		return fmt.Errorf("failed to delete agent node: %w", err)
	}

	return nil
}

// UpdateStatus updates the status of an agent
func (r *GraphAgentRepository) UpdateStatus(ctx context.Context, id string, status domain.AgentStatus) error {
	if id == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	if !status.IsValid() {
		return fmt.Errorf("invalid status: %s", status)
	}

	nodeID := fmt.Sprintf("agent:%s", id)

	// Update just the status property
	properties := map[string]interface{}{
		"status": string(status),
	}

	if err := r.graph.UpdateNode(ctx, "agent", nodeID, properties); err != nil {
		return fmt.Errorf("failed to update agent status: %w", err)
	}

	return nil
}

// UpdateLastSeen updates the last seen timestamp of an agent
func (r *GraphAgentRepository) UpdateLastSeen(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	nodeID := fmt.Sprintf("agent:%s", id)

	// Update just the last_seen property
	properties := map[string]interface{}{
		"last_seen": time.Now().Unix(),
	}

	if err := r.graph.UpdateNode(ctx, "agent", nodeID, properties); err != nil {
		return fmt.Errorf("failed to update agent last seen: %w", err)
	}

	return nil
}

// getAgentCapabilities retrieves capabilities for an agent
func (r *GraphAgentRepository) getAgentCapabilities(ctx context.Context, agentNodeID string) ([]interface{}, error) {
	// Get edges from the agent node
	edges, err := r.graph.GetEdges(ctx, "agent", agentNodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get capability edges: %w", err)
	}

	var capabilities []interface{}
	for _, edge := range edges {
		// Check if this is a HAS_CAPABILITY edge
		edgeType, ok := edge["type"].(string)
		if !ok || edgeType != "HAS_CAPABILITY" {
			continue
		}

		// Get target capability node ID
		targetID, ok := edge["target_id"].(string)
		if !ok {
			continue
		}

		// Get capability node data
		capabilityNode, err := r.graph.GetNode(ctx, "capability", targetID)
		if err != nil {
			continue // Skip if capability node is not found
		}

		// Convert to capability data
		capabilityData := map[string]interface{}{
			"name":        capabilityNode["name"],
			"description": capabilityNode["description"],
			"parameters":  capabilityNode["parameters"],
		}
		capabilities = append(capabilities, capabilityData)
	}

	return capabilities, nil
}
