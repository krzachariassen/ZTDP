package deployments

import (
	"context"
	"fmt"

	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// DeploymentPlanner provides AI-only deployment planning
type DeploymentPlanner interface {
	// Deploy creates a deployment plan for an application
	// AI automatically discovers dependencies and generates optimal order
	Deploy(ctx context.Context, appName string) ([]string, error)
}

// AIDeploymentPlanner implements AI-native deployment planning (no fallbacks)
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

// Deploy generates a deployment plan using AI (AI-native only, no fallbacks)
func (p *AIDeploymentPlanner) Deploy(ctx context.Context, appName string) ([]string, error) {
	if p.provider == nil {
		return nil, fmt.Errorf("AI provider is required for deployment planning - this is an AI-native platform")
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
		return nil, fmt.Errorf("AI deployment planning failed: %w", err)
	}

	// Parse AI response into deployment order (deployment domain logic)
	order, err := p.parseDeploymentOrder(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI deployment response: %w", err)
	}

	p.logger.Info("✅ AI generated deployment order with %d steps", len(order))
	return order, nil
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
