package deployments

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// Service provides ALL deployment domain logic (Clean Architecture - business logic only here)
type Service struct {
	globalGraph *graph.GlobalGraph
	aiProvider  ai.AIProvider
	logger      *logging.Logger
}

// NewDeploymentService creates a new deployment service with AI capabilities
func NewDeploymentService(globalGraph *graph.GlobalGraph, aiProvider ai.AIProvider) *Service {
	return &Service{
		globalGraph: globalGraph,
		aiProvider:  aiProvider,
		logger:      logging.GetLogger().ForComponent("deployment-service"),
	}
}

// DeployApplication deploys an application to an environment (Core Business Logic)
func (s *Service) DeployApplication(ctx context.Context, appName, environment string) (*DeploymentResult, error) {
	s.logger.Info("🚀 Deploying application %s to %s", appName, environment)

	if s.aiProvider == nil {
		return nil, fmt.Errorf("AI provider required - this is an AI-native platform")
	}

	// 1. Validate application exists
	if err := s.ValidateNodeExists("application", appName); err != nil {
		return nil, fmt.Errorf("application validation failed: %w", err)
	}

	// 2. Validate environment exists
	if err := s.ValidateNodeExists("environment", environment); err != nil {
		return nil, fmt.Errorf("environment validation failed: %w", err)
	}

	// 3. Generate deployment plan using AI
	plan, err := s.generateDeploymentPlan(ctx, appName, environment)
	if err != nil {
		return nil, fmt.Errorf("deployment planning failed: %w", err)
	}

	// 4. Execute deployment plan
	result, err := s.executeDeploymentPlan(ctx, appName, environment, plan)
	if err != nil {
		return nil, fmt.Errorf("deployment execution failed: %w", err)
	}

	s.logger.Info("✅ Deployment completed: %s", result.Status)
	return result, nil
}

// GenerateDeploymentPlan creates a deployment plan for an application (Core Business Logic)
func (s *Service) GenerateDeploymentPlan(ctx context.Context, appName string) ([]string, error) {
	s.logger.Info("📋 Generating deployment plan for %s", appName)

	if s.aiProvider == nil {
		return nil, fmt.Errorf("AI provider required - this is an AI-native platform")
	}

	// Validate application exists
	if err := s.ValidateNodeExists("application", appName); err != nil {
		return nil, fmt.Errorf("application validation failed: %w", err)
	}

	// Generate plan using AI
	plan, err := s.generateDeploymentPlan(ctx, appName, "")
	if err != nil {
		return nil, fmt.Errorf("plan generation failed: %w", err)
	}

	s.logger.Info("✅ Plan generated with %d steps", len(plan))
	return plan, nil
}

// GetDeploymentStatus returns the deployment status for an application (Core Business Logic)
func (s *Service) GetDeploymentStatus(appName, environment string) (map[string]interface{}, error) {
	s.logger.Info("📊 Getting deployment status for %s in %s", appName, environment)

	// Get current graph state
	currentGraph, err := s.globalGraph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	// Check if application exists
	if _, ok := currentGraph.Nodes[appName]; !ok {
		return map[string]interface{}{
			"status":      "not_found",
			"application": appName,
			"environment": environment,
			"message":     "Application not found",
		}, nil
	}

	// Check deployment status by looking for deploy edges to the environment
	if edges, exists := currentGraph.Edges[appName]; exists {
		for _, edge := range edges {
			if edge.To == environment && edge.Type == "deploy" {
				return map[string]interface{}{
					"status":        "deployed",
					"application":   appName,
					"environment":   environment,
					"deployment_id": edge.Metadata["deployment_id"],
				}, nil
			}
		}
	}

	return map[string]interface{}{
		"status":      "not_deployed",
		"application": appName,
		"environment": environment,
		"message":     "Application not deployed to environment",
	}, nil
}

// generateDeploymentPlan uses AI to create a deployment plan (AI-native only)
func (s *Service) generateDeploymentPlan(ctx context.Context, appName, environment string) ([]string, error) {
	// Get graph context for AI
	currentGraph, err := s.globalGraph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	// Build system prompt for deployment planning
	systemPrompt := `You are a deployment planning expert. Generate an ordered list of deployment steps.
Return ONLY a JSON array of strings representing the deployment order.
Example: ["database", "api", "frontend"]`

	// Build user prompt with context
	userPrompt := fmt.Sprintf(`Plan deployment for application: %s
Graph context: %v
Environment: %s

Return deployment order as JSON array.`, appName, currentGraph.Nodes, environment)

	// Call AI
	response, err := s.aiProvider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI deployment planning failed: %w", err)
	}

	// Parse deployment order from AI response
	return s.parseDeploymentOrder(response)
}

// executeDeploymentPlan executes the deployment plan
func (s *Service) executeDeploymentPlan(ctx context.Context, appName, environment string, plan []string) (*DeploymentResult, error) {
	result := &DeploymentResult{
		Application:  appName,
		Environment:  environment,
		DeploymentID: fmt.Sprintf("deploy-%s-%s", appName, environment),
		Deployments:  []string{},
		Skipped:      []string{},
		Failed:       []map[string]interface{}{},
		Status:       "completed",
		Summary: DeploymentSummary{
			TotalServices: len(plan),
			Deployed:      len(plan),
			Success:       true,
			Message:       "Deployment completed successfully",
		},
	}

	// For now, mark all as deployed (simplified implementation)
	result.Deployments = plan

	return result, nil
}

// parseDeploymentOrder parses AI response into deployment order
func (s *Service) parseDeploymentOrder(response string) ([]string, error) {
	// Simple implementation - look for JSON array in response
	start := strings.Index(response, "[")
	end := strings.LastIndex(response, "]")

	if start == -1 || end == -1 {
		return []string{}, fmt.Errorf("no valid JSON array found in response")
	}

	jsonStr := response[start : end+1]
	var order []string

	if err := json.Unmarshal([]byte(jsonStr), &order); err != nil {
		return []string{}, fmt.Errorf("failed to parse deployment order: %w", err)
	}

	return order, nil
}

// Graph returns the global graph (for agent access)
func (s *Service) Graph() *graph.GlobalGraph {
	return s.globalGraph
}

// ValidateNodeExists validates that a node exists in the graph
func (s *Service) ValidateNodeExists(kind, name string) error {
	currentGraph, err := s.globalGraph.Graph()
	if err != nil {
		return fmt.Errorf("failed to get graph: %w", err)
	}

	if _, exists := currentGraph.Nodes[name]; !exists {
		return fmt.Errorf("%s '%s' not found in graph", kind, name)
	}

	return nil
}
