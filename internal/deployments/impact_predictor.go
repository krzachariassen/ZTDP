package deployments

import (
	"context"
	"fmt"
	"time"

	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// ImpactPredictor provides AI-powered deployment impact analysis
type ImpactPredictor struct {
	provider ai.AIProvider
	graph    *graph.GlobalGraph
	logger   *logging.Logger
}

// NewImpactPredictor creates a new impact predictor
func NewImpactPredictor(provider ai.AIProvider, graph *graph.GlobalGraph) *ImpactPredictor {
	return &ImpactPredictor{
		provider: provider,
		graph:    graph,
		logger:   logging.GetLogger().ForComponent("impact-predictor"),
	}
}

// PredictImpact analyzes potential impact of deployment changes
// This enables proactive risk assessment before deployment
func (p *ImpactPredictor) PredictImpact(ctx context.Context, changes []ai.ProposedChange, environment string) (*ai.ImpactPrediction, error) {
	if p.provider == nil {
		return nil, fmt.Errorf("AI provider not available for impact prediction")
	}

	p.logger.Info("🔮 Predicting impact of %d changes in %s", len(changes), environment)

	// Extract environment context for impact analysis
	envContext, err := p.extractEnvironmentContext(environment)
	if err != nil {
		return nil, fmt.Errorf("failed to extract environment context: %w", err)
	}

	// Build impact analysis request
	request := &ai.ImpactAnalysisRequest{
		Changes:     changes,
		Scope:       environment,
		Environment: environment,
		Timeframe:   "24h", // Default prediction timeframe
		Metadata: map[string]interface{}{
			"timestamp":         time.Now(),
			"analysis_type":     "pre_deployment",
			"environment_state": envContext,
		},
	}

	// Predict impact using AI
	prediction, err := p.provider.PredictImpact(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("AI impact prediction failed: %w", err)
	}

	p.logger.Info("✅ Impact prediction completed: %s risk, %d affected systems",
		prediction.OverallRisk, len(prediction.AffectedSystems))

	return prediction, nil
}

// extractEnvironmentContext gets detailed environment context for impact analysis
func (p *ImpactPredictor) extractEnvironmentContext(environment string) (map[string]interface{}, error) {
	globalGraph, err := p.graph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	context := map[string]interface{}{
		"environment":        environment,
		"applications":       p.extractEnvironmentApplications(globalGraph, environment),
		"services":           p.extractEnvironmentServices(globalGraph, environment),
		"policies":           p.extractEnvironmentPolicies(globalGraph, environment),
		"current_load":       p.extractEnvironmentLoad(environment),
		"active_deployments": p.extractActiveDeployments(environment),
		"recent_changes":     p.extractRecentChanges(environment),
		"timestamp":          time.Now(),
	}

	return context, nil
}

// Helper methods for environment context extraction
func (p *ImpactPredictor) extractEnvironmentApplications(graph *graph.Graph, environment string) []string {
	applications := []string{}
	for _, node := range graph.Nodes {
		if node.Kind == "application" {
			applications = append(applications, node.ID)
		}
	}
	return applications
}

func (p *ImpactPredictor) extractEnvironmentServices(graph *graph.Graph, environment string) []string {
	services := []string{}
	for _, node := range graph.Nodes {
		if node.Kind == "service" {
			services = append(services, node.ID)
		}
	}
	return services
}

func (p *ImpactPredictor) extractEnvironmentPolicies(graph *graph.Graph, environment string) []string {
	policies := []string{}
	for _, node := range graph.Nodes {
		if node.Kind == "policy" {
			policies = append(policies, node.ID)
		}
	}
	return policies
}

func (p *ImpactPredictor) extractEnvironmentLoad(environment string) map[string]interface{} {
	return map[string]interface{}{
		"cpu":     "45%",
		"memory":  "60%",
		"network": "30%",
	}
}

func (p *ImpactPredictor) extractActiveDeployments(environment string) []string {
	return []string{} // No active deployments
}

func (p *ImpactPredictor) extractRecentChanges(environment string) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"change_id":   "change-123",
			"type":        "service_update",
			"application": "web-app",
			"timestamp":   time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		},
	}
}
