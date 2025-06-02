package ai

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// AIBrain is the core AI reasoning engine for ZTDP
// It replaces hard-coded planner logic with intelligent AI-driven planning
type AIBrain struct {
	provider AIProvider
	graph    *graph.GlobalGraph
	logger   *logging.Logger
}

// NewAIBrain creates a new AI brain instance with the specified provider
func NewAIBrain(provider AIProvider, globalGraph *graph.GlobalGraph) *AIBrain {
	return &AIBrain{
		provider: provider,
		graph:    globalGraph,
		logger:   logging.GetLogger().ForComponent("ai-brain"),
	}
}

// NewAIBrainWithOpenAI creates an AI brain with OpenAI provider using environment configuration
func NewAIBrainWithOpenAI(globalGraph *graph.GlobalGraph) (*AIBrain, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	config := DefaultOpenAIConfig()

	// Allow model override via environment
	if model := os.Getenv("OPENAI_MODEL"); model != "" {
		config.Model = model
	}

	// Allow base URL override for custom deployments
	if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
		config.BaseURL = baseURL
	}

	provider, err := NewOpenAIProvider(config, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
	}

	return NewAIBrain(provider, globalGraph), nil
}

// NewAIBrainFromConfig creates an AI brain using provider selection logic (OpenAI or fallback)
func NewAIBrainFromConfig(globalGraph *graph.GlobalGraph) (*AIBrain, error) {
	providerName := os.Getenv("AI_PROVIDER")
	if providerName == "" {
		providerName = "openai" // default to OpenAI
	}

	switch providerName {
	case "openai":
		return NewAIBrainWithOpenAI(globalGraph)
	case "none", "fallback":
		// TODO: Implement a fallback provider (e.g., deterministic planner)
		return nil, fmt.Errorf("fallback AI provider not implemented")
	default:
		return nil, fmt.Errorf("unknown AI provider: %s", providerName)
	}
}

// GenerateDeploymentPlan replaces the hard-coded topological sort with AI reasoning
// This is the main method that replaces planner.PlanWithEdgeTypes()
func (brain *AIBrain) GenerateDeploymentPlan(ctx context.Context, applicationID string, edgeTypes []string) (*PlanningResponse, error) {
	brain.logger.Info("🧠 AI Brain generating deployment plan for application: %s", applicationID)

	startTime := time.Now()

	// Extract complete application context from the graph
	context, err := brain.extractPlanningContext(applicationID, edgeTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to extract planning context: %w", err)
	}

	// Build the planning request with rich context
	request := &PlanningRequest{
		Intent:        fmt.Sprintf("Deploy application %s with all its services and dependencies", applicationID),
		ApplicationID: applicationID,
		EdgeTypes:     edgeTypes,
		Context:       context,
		Metadata: map[string]interface{}{
			"timestamp":    time.Now(),
			"requested_by": "ztdp-deployment-engine",
			"edge_types":   edgeTypes,
		},
	}

	// Generate the plan using AI reasoning
	response, err := brain.provider.GeneratePlan(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("AI plan generation failed: %w", err)
	}

	// Add timing and metadata
	if response.Metadata == nil {
		response.Metadata = make(map[string]interface{})
	}
	response.Metadata["generation_time_ms"] = time.Since(startTime).Milliseconds()
	response.Metadata["provider"] = brain.provider.GetProviderInfo().Name
	response.Metadata["model"] = brain.provider.GetProviderInfo().Version

	brain.logger.Info("✅ AI deployment plan generated in %dms with %d steps (confidence: %.2f)",
		time.Since(startTime).Milliseconds(),
		len(response.Plan.Steps),
		response.Confidence)

	return response, nil
}

// EvaluateDeploymentPolicies uses AI to evaluate policy compliance for a deployment
func (brain *AIBrain) EvaluateDeploymentPolicies(ctx context.Context, applicationID string, environmentID string) (*PolicyEvaluation, error) {
	brain.logger.Info("🔍 AI Brain evaluating policies for deployment: %s -> %s", applicationID, environmentID)

	// Extract policy context
	policyContext, err := brain.extractPolicyContext(applicationID, environmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to extract policy context: %w", err)
	}

	// Use AI to evaluate policies
	evaluation, err := brain.provider.EvaluatePolicy(ctx, policyContext)
	if err != nil {
		return nil, fmt.Errorf("AI policy evaluation failed: %w", err)
	}

	brain.logger.Info("✅ Policy evaluation completed (compliant: %t, violations: %d)",
		evaluation.Compliant, len(evaluation.Violations))

	return evaluation, nil
}

// OptimizeExistingPlan takes an existing plan and optimizes it using AI
func (brain *AIBrain) OptimizeExistingPlan(ctx context.Context, plan *DeploymentPlan, applicationID string) (*PlanningResponse, error) {
	brain.logger.Info("⚡ AI Brain optimizing existing deployment plan for: %s", applicationID)

	// Extract current context
	context, err := brain.extractPlanningContext(applicationID, []string{"deploy", "create", "owns"})
	if err != nil {
		return nil, fmt.Errorf("failed to extract context for optimization: %w", err)
	}

	// Optimize the plan
	response, err := brain.provider.OptimizePlan(ctx, plan, context)
	if err != nil {
		return nil, fmt.Errorf("AI plan optimization failed: %w", err)
	}

	brain.logger.Info("✅ Plan optimization completed with %d steps", len(response.Plan.Steps))

	return response, nil
}

// GetProviderInfo returns information about the current AI provider
func (brain *AIBrain) GetProviderInfo() *ProviderInfo {
	return brain.provider.GetProviderInfo()
}

// Close cleans up the AI brain resources
func (brain *AIBrain) Close() error {
	brain.logger.Info("🔌 Closing AI Brain")
	return brain.provider.Close()
}

// extractPlanningContext extracts complete context from the graph for AI planning
func (brain *AIBrain) extractPlanningContext(applicationID string, edgeTypes []string) (*PlanningContext, error) {
	// Get application subgraph (this uses the existing planner extraction logic)
	subgraph, err := brain.extractApplicationSubgraph(applicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to extract application subgraph: %w", err)
	}

	// Extract target nodes (nodes that need to be deployed)
	targetNodes := make([]*graph.Node, 0)
	relatedNodes := make([]*graph.Node, 0)

	for _, node := range subgraph.Nodes {
		// Nodes directly related to the application
		if node.Kind == "service" || node.Kind == "application" {
			targetNodes = append(targetNodes, node)
		} else {
			relatedNodes = append(relatedNodes, node)
		}
	}

	// Extract relevant edges
	edges := make([]*graph.Edge, 0)
	for fromID, edgeList := range subgraph.Edges {
		for _, edge := range edgeList {
			// Include edges of the specified types
			for _, edgeType := range edgeTypes {
				if edge.Type == edgeType {
					// Create a new edge structure for AI context
					aiEdge := &graph.Edge{
						To:       edge.To,
						Type:     edge.Type,
						Metadata: edge.Metadata,
					}
					// Add from information to metadata for AI context
					if aiEdge.Metadata == nil {
						aiEdge.Metadata = make(map[string]interface{})
					}
					aiEdge.Metadata["from"] = fromID
					edges = append(edges, aiEdge)
					break
				}
			}
		}
	}

	return &PlanningContext{
		TargetNodes:   targetNodes,
		RelatedNodes:  relatedNodes,
		Edges:         edges,
		PolicyContext: brain.extractPolicyContextForNodes(targetNodes),
		EnvironmentID: brain.extractEnvironmentID(subgraph),
	}, nil
}

// extractApplicationSubgraph extracts the application subgraph (similar to planner logic)
func (brain *AIBrain) extractApplicationSubgraph(applicationID string) (*graph.Graph, error) {
	// Get the global graph
	globalGraph, err := brain.graph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get global graph: %w", err)
	}

	// Create a new subgraph
	subgraph := graph.NewGraph()

	// Add the application node
	appNode, err := globalGraph.GetNode(applicationID)
	if err != nil {
		return nil, fmt.Errorf("application %s not found: %w", applicationID, err)
	}
	subgraph.AddNode(appNode)

	// Recursively add connected nodes and edges
	visited := make(map[string]bool)
	brain.addConnectedNodes(globalGraph, subgraph, applicationID, visited, 0, 3) // Max depth of 3

	return subgraph, nil
}

// addConnectedNodes recursively adds connected nodes to the subgraph
func (brain *AIBrain) addConnectedNodes(source *graph.Graph, target *graph.Graph, nodeID string, visited map[string]bool, depth, maxDepth int) {
	if depth >= maxDepth || visited[nodeID] {
		return
	}

	visited[nodeID] = true

	// Get all edges from this node (outgoing edges)
	if edges, exists := source.Edges[nodeID]; exists {
		for _, edge := range edges {
			// Add the target node
			if targetNode, err := source.GetNode(edge.To); err == nil && !visited[edge.To] {
				target.AddNode(targetNode)
				target.AddEdge(nodeID, edge.To, edge.Type)

				// Recursively add connected nodes
				brain.addConnectedNodes(source, target, edge.To, visited, depth+1, maxDepth)
			}
		}
	}

	// Get all edges to this node (incoming edges) by iterating through all edges
	for fromID, edges := range source.Edges {
		for _, edge := range edges {
			if edge.To == nodeID && !visited[fromID] {
				// Add the source node
				if sourceNode, err := source.GetNode(fromID); err == nil {
					target.AddNode(sourceNode)
					target.AddEdge(fromID, edge.To, edge.Type)

					// Recursively add connected nodes
					brain.addConnectedNodes(source, target, fromID, visited, depth+1, maxDepth)
				}
			}
		}
	}
}

// extractPolicyContext extracts policy context for AI evaluation
func (brain *AIBrain) extractPolicyContext(applicationID, environmentID string) (interface{}, error) {
	// Extract relevant policy information
	policyContext := map[string]interface{}{
		"application_id": applicationID,
		"environment_id": environmentID,
		"timestamp":      time.Now(),
		"policies":       []interface{}{}, // TODO: Extract actual policies
		"constraints":    []interface{}{}, // TODO: Extract constraints
	}

	return policyContext, nil
}

// extractPolicyContextForNodes extracts policy context for a set of nodes
func (brain *AIBrain) extractPolicyContextForNodes(nodes []*graph.Node) interface{} {
	policies := make([]interface{}, 0)

	for _, node := range nodes {
		// Extract policies attached to each node
		if node.Metadata != nil {
			if nodePolicies, exists := node.Metadata["policies"]; exists {
				policies = append(policies, nodePolicies)
			}
		}
	}

	return map[string]interface{}{
		"node_policies": policies,
		"node_count":    len(nodes),
	}
}

// extractEnvironmentID extracts the environment ID from the subgraph
func (brain *AIBrain) extractEnvironmentID(subgraph *graph.Graph) string {
	// Look for environment nodes in the subgraph
	for _, node := range subgraph.Nodes {
		if node.Kind == "environment" {
			return node.ID
		}
	}

	// Default to production if no environment found
	return "production"
}
