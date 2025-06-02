package ai

import (
	"context"
	"fmt"

	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// AIPlanner replaces the hard-coded planner with AI-driven planning
// It maintains compatibility with the existing planner interface while using AI reasoning
type AIPlanner struct {
	brain     *AIBrain
	subgraph  *graph.Graph
	logger    *logging.Logger
	appID     string
	edgeTypes []string
}

// NewAIPlanner creates a new AI-driven planner instance
// This replaces the traditional NewPlanner() function
func NewAIPlanner(brain *AIBrain, subgraph *graph.Graph, appID string) *AIPlanner {
	return &AIPlanner{
		brain:     brain,
		subgraph:  subgraph,
		logger:    logging.GetLogger().ForComponent("ai-planner"),
		appID:     appID,
		edgeTypes: []string{"deploy", "create", "owns"}, // Default edge types
	}
}

// PlanWithEdgeTypes generates an AI-driven deployment plan
// This replaces the hard-coded topological sort in the original planner
func (p *AIPlanner) PlanWithEdgeTypes(edgeTypes []string) ([]string, error) {
	p.logger.Info("🧠 AI Planner generating deployment plan for %s with edge types: %v", p.appID, edgeTypes)

	p.edgeTypes = edgeTypes

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*30) // 30 second timeout
	defer cancel()

	// Generate AI plan
	planResponse, err := p.brain.GenerateDeploymentPlan(ctx, p.appID, edgeTypes)
	if err != nil {
		p.logger.Error("❌ AI plan generation failed: %v", err)
		// Fallback to basic ordering if AI fails
		return p.fallbackPlan()
	}

	// Convert AI plan to deployment order
	order, err := p.convertPlanToOrder(planResponse.Plan)
	if err != nil {
		p.logger.Error("❌ Failed to convert AI plan to order: %v", err)
		return p.fallbackPlan()
	}

	p.logger.Info("✅ AI Planner generated deployment order with %d steps (confidence: %.2f)",
		len(order), planResponse.Confidence)

	// Log the reasoning for transparency
	if planResponse.Reasoning != "" {
		p.logger.Info("🎯 AI Reasoning: %s", planResponse.Reasoning)
	}

	return order, nil
}

// GetSubgraph returns the application subgraph
func (p *AIPlanner) GetSubgraph() *graph.Graph {
	return p.subgraph
}

// GetApplicationID returns the application ID
func (p *AIPlanner) GetApplicationID() string {
	return p.appID
}

// convertPlanToOrder converts an AI-generated DeploymentPlan to a simple deployment order
// This maintains compatibility with the existing deployment engine expectations
func (p *AIPlanner) convertPlanToOrder(plan *DeploymentPlan) ([]string, error) {
	if plan == nil || len(plan.Steps) == 0 {
		return nil, fmt.Errorf("invalid deployment plan")
	}

	order := make([]string, 0, len(plan.Steps))
	stepMap := make(map[string]*DeploymentStep)

	// Build step map for dependency resolution
	for _, step := range plan.Steps {
		stepMap[step.ID] = step
	}

	// Track processed steps to detect cycles
	processed := make(map[string]bool)
	processing := make(map[string]bool)

	// Process steps in dependency order
	for _, step := range plan.Steps {
		if !processed[step.ID] {
			if err := p.processStepDependencies(step, stepMap, processed, processing, &order); err != nil {
				return nil, fmt.Errorf("dependency resolution failed: %w", err)
			}
		}
	}

	return order, nil
}

// processStepDependencies recursively processes step dependencies
func (p *AIPlanner) processStepDependencies(step *DeploymentStep, stepMap map[string]*DeploymentStep,
	processed, processing map[string]bool, order *[]string) error {

	if processing[step.ID] {
		return fmt.Errorf("circular dependency detected involving step: %s", step.ID)
	}

	if processed[step.ID] {
		return nil
	}

	processing[step.ID] = true

	// Process dependencies first
	for _, depID := range step.Dependencies {
		if depStep, exists := stepMap[depID]; exists {
			if err := p.processStepDependencies(depStep, stepMap, processed, processing, order); err != nil {
				return err
			}
		}
	}

	// Add this step to the order
	*order = append(*order, step.Target)
	processed[step.ID] = true
	processing[step.ID] = false

	return nil
}

// fallbackPlan provides a simple fallback when AI planning fails
func (p *AIPlanner) fallbackPlan() ([]string, error) {
	p.logger.Warn("🔄 Using fallback planning due to AI failure")

	// Simple topological sort as fallback
	order := make([]string, 0)

	// Add nodes in a basic order: services first, then other components
	for _, node := range p.subgraph.Nodes {
		if node.Kind == "service" {
			order = append(order, node.ID)
		}
	}

	// Add non-service nodes
	for _, node := range p.subgraph.Nodes {
		if node.Kind != "service" {
			order = append(order, node.ID)
		}
	}

	return order, nil
}

// ExtractApplicationSubgraph extracts the application subgraph for AI planning
// This replaces the original planner.ExtractApplicationSubgraph function
func ExtractApplicationSubgraph(globalGraph *graph.GlobalGraph, appID string) (*graph.Graph, error) {
	logger := logging.GetLogger().ForComponent("ai-planner-extract")
	logger.Info("📊 Extracting application subgraph for: %s", appID)

	// Get the global graph
	sourceGraph, err := globalGraph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get global graph: %w", err)
	}

	// Create a new subgraph
	subgraph := graph.NewGraph()

	// Add the application node
	appNode, err := sourceGraph.GetNode(appID)
	if err != nil {
		return nil, fmt.Errorf("application %s not found: %w", appID, err)
	}
	subgraph.AddNode(appNode)

	// Add all nodes connected to the application
	visited := make(map[string]bool)
	addConnectedNodes(sourceGraph, subgraph, appID, visited, 0, 3) // Max depth of 3

	logger.Info("✅ Extracted subgraph with %d nodes and %d edges",
		len(subgraph.Nodes), len(subgraph.Edges))

	return subgraph, nil
}

// addConnectedNodes recursively adds connected nodes to the subgraph
func addConnectedNodes(source *graph.Graph, target *graph.Graph, nodeID string, visited map[string]bool, depth, maxDepth int) {
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
				addConnectedNodes(source, target, edge.To, visited, depth+1, maxDepth)
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
					addConnectedNodes(source, target, fromID, visited, depth+1, maxDepth)
				}
			}
		}
	}
}
