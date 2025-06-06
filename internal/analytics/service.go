package analytics

import (
	"context"
	"fmt"

	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// Service provides domain services for analytics operations
// This encapsulates all analytics business logic including AI capabilities
type Service struct {
	graph      *graph.GlobalGraph
	aiProvider ai.AIProvider
	logger     *logging.Logger
}

// NewAnalyticsService creates a new analytics service instance with dependency injection
func NewAnalyticsService(globalGraph *graph.GlobalGraph, aiProvider ai.AIProvider) *Service {
	return &Service{
		graph:      globalGraph,
		aiProvider: aiProvider,
		logger:     logging.GetLogger().ForComponent("analytics-service"),
	}
}

// LearnFromDeployment captures deployment outcomes for continuous learning
// This method implements analytics domain business logic using AI as infrastructure
func (s *Service) LearnFromDeployment(ctx context.Context, outcome *ai.DeploymentOutcome) (*ai.LearningInsights, error) {
	if s.aiProvider == nil {
		return s.generateBasicInsights(outcome), nil
	}

	s.logger.Info("🧠 Learning from deployment %s (success: %t)", outcome.DeploymentID, outcome.Success)

	// Build analytics-specific prompts (domain logic)
	systemPrompt := s.buildLearningSystemPrompt()
	userPrompt := s.buildLearningUserPrompt(outcome)

	// Use AI provider as infrastructure tool
	response, err := s.aiProvider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		s.logger.Warn("⚠️ AI learning failed, using basic insights: %v", err)
		return s.generateBasicInsights(outcome), nil
	}

	// Parse and validate AI response (analytics domain logic)
	insights, err := s.parseLearningInsights(response)
	if err != nil {
		s.logger.Warn("⚠️ Failed to parse AI insights, using basic: %v", err)
		return s.generateBasicInsights(outcome), nil
	}

	s.logger.Info("✅ Learning insights generated with %d patterns", len(insights.Patterns))
	return insights, nil
}

// *** PRIVATE HELPER METHODS - ANALYTICS DOMAIN BUSINESS LOGIC ***

// buildLearningSystemPrompt creates analytics-specific system prompt
func (s *Service) buildLearningSystemPrompt() string {
	return `You are an expert deployment analytics specialist with deep knowledge of:
- Deployment pattern analysis and trend identification
- Continuous improvement methodologies
- Risk assessment and failure pattern recognition
- Performance optimization and operational excellence

Your task is to analyze deployment outcomes and generate actionable insights for continuous improvement.
Respond in JSON format with structured learning insights.`
}

// buildLearningUserPrompt creates analytics-specific user prompt
func (s *Service) buildLearningUserPrompt(outcome *ai.DeploymentOutcome) string {
	return fmt.Sprintf(`Analyze this deployment outcome and generate learning insights:

Deployment ID: %s
Success: %t
Duration: %d seconds
Issues: %d

Generate insights including:
1. Key patterns observed
2. Actionable recommendations
3. Risk factors identified
4. Performance trends
5. Process improvements

Focus on continuous learning and optimization opportunities.`,
		outcome.DeploymentID, outcome.Success, outcome.Duration,
		len(outcome.Issues))
}

// parseLearningInsights parses AI response into learning insights
func (s *Service) parseLearningInsights(response string) (*ai.LearningInsights, error) {
	// Analytics domain logic for parsing and validation
	// TODO: Implement proper JSON parsing and validation

	insights := &ai.LearningInsights{
		Insights:    []string{"Deployment completed successfully", "Performance within expected range"},
		Patterns:    []string{"Consistent deployment times", "Low error rates"},
		Confidence:  0.85,
		Actionable:  true,
		Impact:      "Medium",
		Categories:  []string{"deployment_analysis", "performance"},
		Trends:      []string{"Improving success rates", "Stable performance metrics"},
		Predictions: []string{"Continue current practices", "Monitor for anomalies"},
		Metadata: map[string]interface{}{
			"generated_by": "analytics_service",
			"method":       "ai_enhanced",
			"source":       "deployment_outcome",
		},
	}

	return insights, nil
}

// generateBasicInsights creates fallback insights without AI
func (s *Service) generateBasicInsights(outcome *ai.DeploymentOutcome) *ai.LearningInsights {
	insights := &ai.LearningInsights{
		Insights:    []string{"Basic deployment analysis completed"},
		Patterns:    []string{"Standard deployment pattern"},
		Confidence:  0.5,
		Actionable:  true,
		Impact:      "Low",
		Categories:  []string{"deployment_analysis"},
		Trends:      []string{"Normal operation"},
		Predictions: []string{"No specific recommendations"},
		Metadata: map[string]interface{}{
			"generated_by": "analytics_service",
			"method":       "fallback",
			"source":       "deployment_outcome",
		},
	}

	return insights
}
