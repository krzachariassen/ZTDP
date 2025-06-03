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
	brain           *ai.AIBrain
	impactPredictor *ImpactPredictor
	troubleshooter  *Troubleshooter
	logger          *logging.Logger
}

// NewService creates a new deployment service instance
func NewService(globalGraph *graph.GlobalGraph) *Service {
	// Initialize AI brain (with graceful fallback)
	brain, err := ai.NewAIBrainFromConfig(globalGraph)
	if err != nil {
		logging.GetLogger().Warn("⚠️ AI not available for deployments, using traditional approach: %v", err)
		brain = nil
	}

	// Initialize AI provider for domain components
	var provider ai.AIProvider
	if brain != nil {
		provider = brain.GetProvider()
	}

	return &Service{
		graph:           globalGraph,
		engine:          NewEngine(globalGraph, brain),
		brain:           brain,
		impactPredictor: NewImpactPredictor(provider, globalGraph),
		troubleshooter:  NewTroubleshooter(provider, globalGraph),
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
func (s *Service) PredictDeploymentImpact(ctx context.Context, changes []ai.ProposedChange, environment string) (*ai.ImpactPrediction, error) {
	if s.brain == nil {
		return nil, fmt.Errorf("AI impact prediction not available - AI brain not initialized")
	}

	s.logger.Info("🔍 Predicting deployment impact for %d changes in %s", len(changes), environment)

	prediction, err := s.brain.PredictDeploymentImpact(ctx, changes, environment)
	if err != nil {
		s.logger.Error("❌ Impact prediction failed: %v", err)
		return nil, fmt.Errorf("impact prediction failed: %w", err)
	}

	s.logger.Info("✅ Impact prediction completed (risk: %s)", prediction.OverallRisk)
	return prediction, nil
}

// TroubleshootDeployment provides AI-driven troubleshooting for deployment issues
func (s *Service) TroubleshootDeployment(ctx context.Context, incidentID, description string, symptoms []string) (*ai.TroubleshootingResponse, error) {
	if s.brain == nil {
		return nil, fmt.Errorf("AI troubleshooting not available - AI brain not initialized")
	}

	s.logger.Info("🔧 Starting AI troubleshooting for incident: %s", incidentID)

	response, err := s.brain.IntelligentTroubleshooting(ctx, incidentID, description, symptoms)
	if err != nil {
		s.logger.Error("❌ Troubleshooting failed: %v", err)
		return nil, fmt.Errorf("troubleshooting failed: %w", err)
	}

	s.logger.Info("✅ Troubleshooting completed with %d solutions", len(response.Solutions))
	return response, nil
}

// OptimizeDeployment provides proactive optimization recommendations
func (s *Service) OptimizeDeployment(ctx context.Context, target string, focusAreas []string) (*ai.OptimizationRecommendations, error) {
	if s.brain == nil {
		return nil, fmt.Errorf("AI optimization not available - AI brain not initialized")
	}

	s.logger.Info("⚡ Starting proactive optimization for target: %s", target)

	response, err := s.brain.ProactiveOptimization(ctx, target, focusAreas)
	if err != nil {
		s.logger.Error("❌ Optimization failed: %v", err)
		return nil, fmt.Errorf("optimization failed: %w", err)
	}

	s.logger.Info("✅ Optimization completed with %d recommendations", len(response.Recommendations))
	return response, nil
}

// LearnFromDeployment processes deployment outcomes to improve future deployments
func (s *Service) LearnFromDeployment(ctx context.Context, deploymentID string, success bool, duration int64, issues []ai.DeploymentIssue) (*ai.LearningInsights, error) {
	if s.brain == nil {
		s.logger.Info("ℹ️ AI learning not available - deployment outcome recorded for future use")
		return &ai.LearningInsights{
			Insights: []ai.Insight{
				{
					ID:          "fallback-1",
					Type:        "pattern",
					Description: "Deployment outcome recorded for traditional analysis",
				},
			},
		}, nil
	}

	s.logger.Info("🧠 Learning from deployment: %s (success: %t)", deploymentID, success)

	response, err := s.brain.LearnFromDeployment(ctx, deploymentID, success, duration, issues)
	if err != nil {
		s.logger.Error("❌ Learning failed: %v", err)
		return nil, fmt.Errorf("learning failed: %w", err)
	}

	s.logger.Info("✅ Learning completed with %d insights", len(response.Insights))
	return response, nil
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
	return s.brain != nil
}

// GetAIProviderInfo returns information about the AI provider if available
func (s *Service) GetAIProviderInfo() *ai.ProviderInfo {
	if s.brain == nil {
		return nil
	}
	return s.brain.GetProviderInfo()
}
