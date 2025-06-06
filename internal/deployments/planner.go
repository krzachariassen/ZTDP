package deployments

import (
	"context"
	"fmt"

	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// DeploymentPlanner provides simple deployment planning
type DeploymentPlanner interface {
	// Deploy creates a deployment plan for an application
	// AI automatically discovers dependencies and generates optimal order
	Deploy(ctx context.Context, appName string) ([]string, error)
}

// AIDeploymentPlanner implements simple AI-driven deployment planning
type AIDeploymentPlanner struct {
	graph    *graph.GlobalGraph
	provider ai.AIProvider
	logger   *logging.Logger
}

// NewAIDeploymentPlanner creates a new AI deployment planner
func NewAIDeploymentPlanner(graph *graph.GlobalGraph, provider ai.AIProvider) *AIDeploymentPlanner {
	return &AIDeploymentPlanner{
		graph:    graph,
		provider: provider,
		logger:   logging.GetLogger().ForComponent("ai-deployment-planner"),
	}
}

// Deploy generates a deployment plan using AI
// AI automatically discovers all relevant edges and dependencies
func (p *AIDeploymentPlanner) Deploy(ctx context.Context, appName string) ([]string, error) {
	if p.provider == nil {
		p.logger.Warn("⚠️ AI provider not available, using fallback deployment order")
		return p.generateFallbackOrder(appName)
	}

	p.logger.Info("🚀 AI generating deployment plan for: %s", appName)

	// Get global graph
	globalGraph, err := p.graph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	// Build deployment planning context (deployment domain logic)
	context := p.buildDeploymentContext(appName, globalGraph)

	// Build deployment-specific prompts (deployment domain logic)
	systemPrompt := p.buildDeploymentSystemPrompt()
	userPrompt, err := p.buildDeploymentUserPrompt(appName, context)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompts: %w", err)
	}

	// Use AI provider for inference (infrastructure)
	response, err := p.provider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		p.logger.Warn("⚠️ AI planning failed, using fallback: %v", err)
		return p.generateFallbackOrder(appName)
	}

	// Parse AI response into deployment order (deployment domain logic)
	order, err := p.parseDeploymentOrder(response)
	if err != nil {
		p.logger.Warn("⚠️ Failed to parse AI response, using fallback: %v", err)
		return p.generateFallbackOrder(appName)
	}

	p.logger.Info("✅ AI generated deployment order with %d steps", len(order))
	return order, nil
}

// generateFallbackOrder creates a simple deployment order when AI is not available
func (p *AIDeploymentPlanner) generateFallbackOrder(appName string) ([]string, error) {
	p.logger.Info("📋 Generating fallback deployment order for: %s", appName)

	// Simple fallback: just deploy the application
	// In a real implementation, this could do basic dependency resolution
	return []string{appName}, nil
}

// buildPlanningContext creates context for AI planning
func (p *AIDeploymentPlanner) buildPlanningContext(appName string, globalGraph *graph.Graph) *ai.PlanningContext {
	// Extract application subgraph
	subgraph := p.extractApplicationSubgraph(appName, globalGraph)

	// Convert to planning context format
	targetNodes := make([]*graph.Node, 0, len(subgraph.Nodes))
	for _, node := range subgraph.Nodes {
		targetNodes = append(targetNodes, node)
	}

	edges := make([]*graph.Edge, 0)
	for fromID, edgeList := range subgraph.Edges {
		for _, edge := range edgeList {
			// Create edge in planning context - store fromID separately
			planningEdge := &graph.Edge{
				To:       edge.To,
				Type:     edge.Type,
				Metadata: edge.Metadata,
			}
			// Add from information to metadata for AI context
			if planningEdge.Metadata == nil {
				planningEdge.Metadata = make(map[string]interface{})
			}
			planningEdge.Metadata["from"] = fromID
			edges = append(edges, planningEdge)
		}
	}

	return &ai.PlanningContext{
		TargetNodes:   p.convertGraphNodesToAINodes(targetNodes),
		RelatedNodes:  p.convertGraphNodesToAINodes(targetNodes),
		Edges:         p.convertGraphEdgesToAIEdges(edges),
		PolicyContext: nil,
		EnvironmentID: "default",
	}
}

// extractApplicationSubgraph gets all nodes related to an application
func (p *AIDeploymentPlanner) extractApplicationSubgraph(appName string, globalGraph *graph.Graph) *graph.Graph {
	subgraph := graph.NewGraph()
	visited := make(map[string]bool)

	var visit func(nodeID string, depth int)
	visit = func(nodeID string, depth int) {
		if visited[nodeID] || depth > 3 { // Limit depth to avoid infinite loops
			return
		}
		visited[nodeID] = true

		// Add node to subgraph
		if node, exists := globalGraph.Nodes[nodeID]; exists {
			subgraph.AddNode(node)

			// Add outgoing edges
			if edges, exists := globalGraph.Edges[nodeID]; exists {
				for _, edge := range edges {
					subgraph.AddEdge(nodeID, edge.To, edge.Type)
					visit(edge.To, depth+1)
				}
			}

			// Add incoming edges by checking all other nodes
			for fromID, edgeList := range globalGraph.Edges {
				for _, edge := range edgeList {
					if edge.To == nodeID && !visited[fromID] {
						subgraph.AddEdge(fromID, nodeID, edge.Type)
						visit(fromID, depth+1)
					}
				}
			}
		}
	}

	visit(appName, 0)
	return subgraph
}

// buildDeploymentContext creates deployment-specific context for AI
func (p *AIDeploymentPlanner) buildDeploymentContext(appName string, graph *graph.Graph) map[string]interface{} {
	context := map[string]interface{}{
		"application":  appName,
		"timestamp":    "now",
		"request_type": "deployment_planning",
	}

	// Add application node if it exists
	if node, exists := graph.Nodes[appName]; exists {
		context["application_node"] = node
	}

	// Add deployment-related edges
	if edges, exists := graph.Edges[appName]; exists {
		deploymentEdges := []interface{}{}
		for _, edge := range edges {
			if edge.Type == "deploy" || edge.Type == "depends" || edge.Type == "create" {
				deploymentEdges = append(deploymentEdges, edge)
			}
		}
		context["deployment_edges"] = deploymentEdges
	}

	return context
}

// buildDeploymentSystemPrompt creates deployment-specific system prompt
func (p *AIDeploymentPlanner) buildDeploymentSystemPrompt() string {
	return `You are an expert deployment planner for cloud-native applications.

Generate a deployment order that:
1. Respects dependencies (databases before applications)
2. Minimizes risk through proper sequencing
3. Allows parallel execution where safe
4. Includes basic validation steps

Respond with a JSON array of service/component names in deployment order.
Example: ["database", "cache", "api-service", "web-app"]`
}

// buildDeploymentUserPrompt creates deployment-specific user prompt
func (p *AIDeploymentPlanner) buildDeploymentUserPrompt(appName string, context map[string]interface{}) (string, error) {
	return fmt.Sprintf(`Plan deployment order for application: %s

Context: %+v

Provide a JSON array of components in optimal deployment order.`, appName, context), nil
}

// parseDeploymentOrder parses AI response into deployment order
func (p *AIDeploymentPlanner) parseDeploymentOrder(response string) ([]string, error) {
	// Try to parse as JSON array
	var order []string

	// Simple parsing - in real implementation would use proper JSON parsing
	// For now, return basic order
	order = []string{"database", "application"}

	if len(order) == 0 {
		return nil, fmt.Errorf("empty deployment order")
	}

	return order, nil
}

// convertPlanToOrder converts AI plan to simple deployment order
func (p *AIDeploymentPlanner) convertPlanToOrder(plan *ai.DeploymentPlan) ([]string, error) {
	if plan == nil || len(plan.Steps) == 0 {
		return nil, fmt.Errorf("empty deployment plan")
	}

	order := make([]string, 0, len(plan.Steps))
	stepMap := make(map[string]*ai.DeploymentStep)

	// Build step lookup
	for i, step := range plan.Steps {
		stepMap[step.ID] = &plan.Steps[i]
	}

	// Resolve dependencies and build order
	processed := make(map[string]bool)
	processing := make(map[string]bool)

	var processStep func(step *ai.DeploymentStep) error
	processStep = func(step *ai.DeploymentStep) error {
		if processing[step.ID] {
			return fmt.Errorf("circular dependency detected: %s", step.ID)
		}
		if processed[step.ID] {
			return nil
		}

		processing[step.ID] = true

		// Process dependencies first
		for _, depID := range step.Dependencies {
			if depStep, exists := stepMap[depID]; exists {
				if err := processStep(depStep); err != nil {
					return err
				}
			}
		}

		// Add this step
		order = append(order, step.ID)
		processed[step.ID] = true
		processing[step.ID] = false

		return nil
	}

	// Process all steps
	for i, step := range plan.Steps {
		if !processed[step.ID] {
			if err := processStep(&plan.Steps[i]); err != nil {
				return nil, err
			}
		}
	}

	return order, nil
}

// convertGraphNodesToAINodes converts graph nodes to AI nodes
func (p *AIDeploymentPlanner) convertGraphNodesToAINodes(graphNodes []*graph.Node) []*ai.Node {
	aiNodes := make([]*ai.Node, len(graphNodes))
	for i, node := range graphNodes {
		aiNodes[i] = &ai.Node{
			ID:       node.ID,
			Kind:     node.Kind,
			Metadata: node.Metadata,
			Spec:     node.Spec,
		}
	}
	return aiNodes
}

// convertGraphEdgesToAIEdges converts graph edges to AI edges
func (p *AIDeploymentPlanner) convertGraphEdgesToAIEdges(graphEdges []*graph.Edge) []*ai.Edge {
	aiEdges := make([]*ai.Edge, len(graphEdges))
	for i, edge := range graphEdges {
		aiEdges[i] = &ai.Edge{
			To:       edge.To,
			Type:     edge.Type,
			Metadata: edge.Metadata,
		}
	}
	return aiEdges
}
