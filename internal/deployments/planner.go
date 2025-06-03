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
	p.logger.Info("🚀 AI generating deployment plan for: %s", appName)

	// Get global graph
	globalGraph, err := p.graph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	// Build planning context - AI will discover all relevant edges automatically
	planningContext := p.buildPlanningContext(appName, globalGraph)

	// Create planning request - no edge type limitations
	request := &ai.PlanningRequest{
		Intent:        fmt.Sprintf("Deploy application %s", appName),
		ApplicationID: appName,
		EdgeTypes:     nil, // AI discovers edges dynamically
		Context:       planningContext,
	}

	// Generate AI plan
	response, err := p.provider.GeneratePlan(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("AI planning failed: %w", err)
	}

	// Convert to simple deployment order
	order, err := p.convertPlanToOrder(response.Plan)
	if err != nil {
		return nil, fmt.Errorf("failed to convert plan: %w", err)
	}

	p.logger.Info("✅ AI generated deployment order with %d steps (confidence: %.2f)",
		len(order), response.Confidence)

	if response.Reasoning != "" {
		p.logger.Info("🎯 AI Reasoning: %s", response.Reasoning)
	}

	return order, nil
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
		TargetNodes:   targetNodes,
		RelatedNodes:  targetNodes,
		Edges:         edges,
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

// convertPlanToOrder converts AI plan to simple deployment order
func (p *AIDeploymentPlanner) convertPlanToOrder(plan *ai.DeploymentPlan) ([]string, error) {
	if plan == nil || len(plan.Steps) == 0 {
		return nil, fmt.Errorf("empty deployment plan")
	}

	order := make([]string, 0, len(plan.Steps))
	stepMap := make(map[string]*ai.DeploymentStep)

	// Build step lookup
	for _, step := range plan.Steps {
		stepMap[step.ID] = step
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
		order = append(order, step.Target)
		processed[step.ID] = true
		processing[step.ID] = false

		return nil
	}

	// Process all steps
	for _, step := range plan.Steps {
		if !processed[step.ID] {
			if err := processStep(step); err != nil {
				return nil, err
			}
		}
	}

	return order, nil
}
