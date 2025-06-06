package deployments

import (
	"context"
	"fmt"

	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// Service provides domain services for deployment operations
// This encapsulates all deployment business logic including AI capabilities
type Service struct {
	graph           *graph.GlobalGraph
	engine          *Engine
	aiProvider      ai.AIProvider
	impactPredictor *ImpactPredictor
	troubleshooter  *Troubleshooter
	logger          *logging.Logger
}

// NewDeploymentService creates a new deployment service instance with dependency injection
func NewDeploymentService(globalGraph *graph.GlobalGraph, aiProvider ai.AIProvider) *Service {
	return &Service{
		graph:           globalGraph,
		engine:          NewEngineWithProvider(globalGraph, aiProvider),
		aiProvider:      aiProvider,
		impactPredictor: NewImpactPredictor(aiProvider, globalGraph),
		troubleshooter:  NewTroubleshooter(aiProvider, globalGraph),
		logger:          logging.GetLogger().ForComponent("deployment-service"),
	}
}

// DeployApplication orchestrates the deployment of an entire application
func (s *Service) DeployApplication(ctx context.Context, appName, environment string) (*DeploymentResult, error) {
	s.logger.Info("🚀 Starting application deployment: %s -> %s", appName, environment)

	// Delegate to engine for actual deployment execution
	// The engine handles AI vs traditional planning internally
	result, err := s.engine.ExecuteApplicationDeployment(appName, environment)
	if err != nil {
		s.logger.Error("❌ Application deployment failed: %v", err)
		return nil, fmt.Errorf("deployment failed: %w", err)
	}

	s.logger.Info("✅ Application deployment completed: %s -> %s", appName, environment)
	return result, nil
}

// PredictDeploymentImpact analyzes potential impact of deployment changes
// This method implements deployment domain business logic for impact prediction
func (s *Service) PredictDeploymentImpact(ctx context.Context, changes []ai.ProposedChange, environment string) (*ai.ImpactPrediction, error) {
	if s.impactPredictor == nil {
		return nil, fmt.Errorf("AI impact prediction not available - AI provider not initialized")
	}

	s.logger.Info("🔍 Predicting deployment impact for %d changes in %s", len(changes), environment)

	prediction, err := s.impactPredictor.PredictImpact(ctx, changes, environment)
	if err != nil {
		s.logger.Error("❌ Impact prediction failed: %v", err)
		return nil, fmt.Errorf("impact prediction failed: %w", err)
	}

	s.logger.Info("✅ Impact prediction completed (risk: %s)", prediction.OverallRisk)
	return prediction, nil
}

// TroubleshootDeployment provides AI-driven troubleshooting for deployment issues
// This method implements deployment domain business logic for troubleshooting
func (s *Service) TroubleshootDeployment(ctx context.Context, incidentID, description string, symptoms []string) (*ai.TroubleshootingResponse, error) {
	if s.troubleshooter == nil {
		return nil, fmt.Errorf("AI troubleshooting not available - AI provider not initialized")
	}

	s.logger.Info("🔧 Starting AI troubleshooting for incident: %s", incidentID)

	response, err := s.troubleshooter.Troubleshoot(ctx, incidentID, description, symptoms)
	if err != nil {
		s.logger.Error("❌ Troubleshooting failed: %v", err)
		return nil, fmt.Errorf("troubleshooting failed: %w", err)
	}

	s.logger.Info("✅ Troubleshooting completed with %d recommendations", len(response.Recommendations))
	return response, nil
}

// GenerateDeploymentPlan creates a deployment plan for an application
// This method implements deployment domain business logic for plan generation
func (s *Service) GenerateDeploymentPlan(ctx context.Context, appName string) (*ai.DeploymentPlan, error) {
	if s.aiProvider == nil {
		return nil, fmt.Errorf("AI plan generation not available - AI provider not initialized")
	}

	s.logger.Info("🧠 Generating deployment plan for application: %s", appName)

	// Extract application context from graph
	graph, err := s.graph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	// Build deployment planning context (deployment domain logic)
	context := s.buildDeploymentPlanningContext(appName, graph)

	// Build deployment-specific prompts (deployment domain logic)
	systemPrompt := s.buildDeploymentSystemPrompt()
	userPrompt, err := s.buildDeploymentUserPrompt(appName, context)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompts: %w", err)
	}

	// Use AI provider for inference (infrastructure)
	response, err := s.aiProvider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI inference failed: %w", err)
	}

	// Parse and validate response (deployment domain logic)
	plan, err := s.parseDeploymentPlan(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse deployment plan: %w", err)
	}

	s.logger.Info("✅ Deployment plan generated with %d steps", len(plan.Steps))
	return plan, nil
}

// GetDeploymentStatus retrieves the current status of a deployment
func (s *Service) GetDeploymentStatus(appName, environment string) (*DeploymentStatus, error) {
	s.logger.Info("📊 Retrieving deployment status: %s -> %s", appName, environment)

	// Get deployment status from the graph
	graph, err := s.graph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	// Find deployment edges for this app
	if edges, exists := graph.Edges[appName]; exists {
		for _, edge := range edges {
			if edge.Type == "deploy" {
				// Check if this is for the target environment
				if envID, exists := edge.Metadata["environment_id"]; exists && envID == environment {
					status, message, found := GetDeploymentStatus(edge.Metadata)
					if found {
						s.logger.Info("✅ Found deployment status: %s - %s", status, message)
						return &status, nil
					}
				}
			}
		}
	}

	// No deployment found - return default status
	defaultStatus := StatusPending
	return &defaultStatus, nil
}

// HasAICapabilities returns whether AI features are available
func (s *Service) HasAICapabilities() bool {
	return s.aiProvider != nil
}

// GetAIProviderInfo returns information about the AI provider if available
func (s *Service) GetAIProviderInfo() *ai.ProviderInfo {
	if s.aiProvider == nil {
		return nil
	}
	return s.aiProvider.GetProviderInfo()
}

// *** PRIVATE HELPER METHODS - DEPLOYMENT DOMAIN BUSINESS LOGIC ***

// buildDeploymentPlanningContext creates deployment-specific planning context
func (s *Service) buildDeploymentPlanningContext(appName string, graph *graph.Graph) map[string]interface{} {
	// This is deployment domain logic - understanding deployment context
	context := map[string]interface{}{
		"application":  appName,
		"timestamp":    "now",
		"request_type": "deployment_planning",
	}

	// Add application nodes and dependencies (deployment domain logic)
	if nodes, exists := graph.Nodes[appName]; exists {
		context["application_node"] = nodes
	}

	// Add deployment-related edges (deployment domain logic)
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
func (s *Service) buildDeploymentSystemPrompt() string {
	// Deployment domain knowledge encoded in prompts
	return `You are an expert deployment planner specializing in cloud-native applications.

Your expertise includes:
- Container orchestration and microservices
- Deployment strategies (rolling, blue-green, canary)
- Dependency management and ordering
- Risk assessment and rollback procedures
- Infrastructure provisioning and configuration

Generate deployment plans that:
1. Respect all dependencies and ordering constraints
2. Minimize deployment risk through proper sequencing
3. Allow for parallel execution where safe
4. Include validation checkpoints
5. Provide clear rollback procedures

Respond in JSON format with a deployment plan.`
}

// buildDeploymentUserPrompt creates deployment-specific user prompt
func (s *Service) buildDeploymentUserPrompt(appName string, context map[string]interface{}) (string, error) {
	// Deployment domain logic for prompt construction
	return fmt.Sprintf(`Plan deployment for application: %s

Context: %+v

Generate an optimal deployment plan considering:
- Application dependencies
- Infrastructure requirements  
- Risk mitigation strategies
- Parallel execution opportunities

Provide the plan in JSON format.`, appName, context), nil
}

// parseDeploymentPlan parses AI response into deployment plan
func (s *Service) parseDeploymentPlan(response string) (*ai.DeploymentPlan, error) {
	// Deployment domain logic for parsing and validation
	var plan ai.DeploymentPlan
	// TODO: Implement proper JSON parsing and validation
	// This is deployment domain business logic

	// For now, return a basic plan
	plan.Steps = []ai.DeploymentStep{
		{
			ID:          "step-1",
			Type:        "deploy",
			Description: "Deploy application",
		},
	}

	return &plan, nil
}
