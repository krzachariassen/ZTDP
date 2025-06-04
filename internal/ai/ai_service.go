package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// AIService provides domain-agnostic AI business logic
// This separates business rules from provider-specific infrastructure
type AIService struct {
	provider AIProvider         // Infrastructure provider (OpenAI, Anthropic, etc.)
	graph    *graph.GlobalGraph // Platform graph access
	logger   *logging.Logger    // Logging component
}

// NewAIService creates a new AI service with the specified provider
func NewAIService(provider AIProvider, graph *graph.GlobalGraph) *AIService {
	return &AIService{
		provider:          provider,
		graph:             graph,
		logger:            logging.GetLogger().ForComponent("ai-service"),
		planningPrompts:   NewPlannerPrompts(),
		policyPrompts:     NewPolicyPrompts(),
		deploymentPrompts: prompts.NewDeploymentPrompts(),
	}
}

// GenerateDeploymentPlan creates an intelligent deployment plan using AI reasoning
// This is provider-agnostic business logic that works with any AI provider
func (s *AIService) GenerateDeploymentPlan(ctx context.Context, request *PlanningRequest) (*PlanningResponse, error) {
	s.logger.Info("🧠 Generating AI deployment plan for application: %s", request.ApplicationID)

	// Build provider-agnostic prompts
	systemPrompt := s.deploymentPrompts.BuildPlanningSystemPrompt()
	userPrompt, err := s.deploymentPrompts.BuildPlanningUserPrompt(request)
	if err != nil {
		return nil, fmt.Errorf("failed to build user prompt: %w", err)
	}

	// Use provider for AI inference (provider handles communication)
	response, err := s.provider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI inference failed: %w", err)
	}

	// Parse response using domain logic
	planResponse, err := s.parsePlanningResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Validate business rules
	if err := s.validateDeploymentPlan(planResponse.Plan); err != nil {
		return nil, fmt.Errorf("invalid deployment plan: %w", err)
	}

	s.logger.Info("✅ AI deployment plan generated with %d steps (confidence: %.2f)",
		len(planResponse.Plan.Steps), planResponse.Confidence)

	return planResponse, nil
}

// EvaluateDeploymentPolicy uses AI to evaluate policy compliance
// Generic business logic that works with any AI provider
func (s *AIService) EvaluateDeploymentPolicy(ctx context.Context, policyContext interface{}) (*PolicyEvaluation, error) {
	s.logger.Info("🔍 Evaluating policy compliance using AI")

	// Build provider-agnostic prompts
	systemPrompt := s.policyPrompts.BuildPolicySystemPrompt()
	userPrompt, err := s.policyPrompts.BuildPolicyUserPrompt(policyContext)
	if err != nil {
		return nil, fmt.Errorf("failed to build policy prompt: %w", err)
	}

	// Use provider for AI inference
	response, err := s.provider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI policy evaluation failed: %w", err)
	}

	// Parse response using domain logic
	evaluation, err := s.parsePolicyEvaluation(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse policy evaluation: %w", err)
	}

	// Apply business rules
	if evaluation.Confidence == 0 {
		evaluation.Confidence = 0.8 // Default confidence
	}

	s.logger.Info("✅ Policy evaluation completed (compliant: %t, confidence: %.2f)",
		evaluation.Compliant, evaluation.Confidence)

	return evaluation, nil
}

// OptimizeDeploymentPlan refines an existing plan using AI
// Provider-agnostic optimization logic
func (s *AIService) OptimizeDeploymentPlan(ctx context.Context, plan *DeploymentPlan, context *PlanningContext) (*PlanningResponse, error) {
	s.logger.Info("⚡ Optimizing deployment plan using AI")

	// Build provider-agnostic prompts
	systemPrompt := s.deploymentPrompts.BuildOptimizationSystemPrompt()
	userPrompt, err := s.deploymentPrompts.BuildOptimizationPrompt(plan, context)
	if err != nil {
		return nil, fmt.Errorf("failed to build optimization prompt: %w", err)
	}

	// Use provider for AI inference
	response, err := s.provider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI optimization failed: %w", err)
	}

	// Parse and validate response
	planResponse, err := s.parsePlanningResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse optimization response: %w", err)
	}

	if err := s.validateDeploymentPlan(planResponse.Plan); err != nil {
		return nil, fmt.Errorf("invalid optimized plan: %w", err)
	}

	s.logger.Info("✅ Plan optimization completed with %d steps", len(planResponse.Plan.Steps))
	return planResponse, nil
}

// GetProviderInfo returns information about the current AI provider
func (s *AIService) GetProviderInfo() *ProviderInfo {
	return s.provider.GetProviderInfo()
}

// Close cleans up AI service resources
func (s *AIService) Close() error {
	s.logger.Info("🔌 Closing AI service")
	return s.provider.Close()
}

// *** PRIVATE HELPER METHODS - DOMAIN BUSINESS LOGIC ***

// parsePlanningResponse parses AI response into PlanningResponse
func (s *AIService) parsePlanningResponse(response string) (*PlanningResponse, error) {
	var planResponse PlanningResponse
	if err := json.Unmarshal([]byte(response), &planResponse); err != nil {
		return nil, fmt.Errorf("failed to parse planning response: %w", err)
	}

	// Validate the response structure
	if planResponse.Plan == nil {
		return nil, fmt.Errorf("response missing deployment plan")
	}

	if len(planResponse.Plan.Steps) == 0 {
		return nil, fmt.Errorf("deployment plan has no steps")
	}

	// Set default confidence if not provided
	if planResponse.Confidence == 0 {
		planResponse.Confidence = 0.8
	}

	return &planResponse, nil
}

// parsePolicyEvaluation parses AI response into PolicyEvaluation
func (s *AIService) parsePolicyEvaluation(response string) (*PolicyEvaluation, error) {
	var evaluation PolicyEvaluation
	if err := json.Unmarshal([]byte(response), &evaluation); err != nil {
		return nil, fmt.Errorf("failed to parse policy evaluation: %w", err)
	}

	// Set default confidence if not provided
	if evaluation.Confidence == 0 {
		evaluation.Confidence = 0.8
	}

	return &evaluation, nil
}

// validateDeploymentPlan validates deployment plan business rules
func (s *AIService) validateDeploymentPlan(plan *DeploymentPlan) error {
	if plan == nil {
		return fmt.Errorf("plan cannot be nil")
	}

	if len(plan.Steps) == 0 {
		return fmt.Errorf("plan must have at least one step")
	}

	// Additional domain-specific validation rules
	// ...

	return nil
}
