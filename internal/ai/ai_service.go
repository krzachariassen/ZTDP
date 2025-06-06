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
		provider: provider,
		graph:    graph,
		logger:   logging.GetLogger().ForComponent("ai-service"),
	}
}

// GenerateDeploymentPlan creates an intelligent deployment plan using AI reasoning
// This is provider-agnostic business logic that works with any AI provider
func (s *AIService) GenerateDeploymentPlan(ctx context.Context, request *PlanningRequest) (*PlanningResponse, error) {
	s.logger.Info("🧠 Generating AI deployment plan for application: %s", request.ApplicationID)

	// Build provider-agnostic prompts
	systemPrompt := s.buildPlanningSystemPrompt()
	userPrompt, err := s.buildPlanningUserPrompt(request)
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
	systemPrompt := s.buildPolicySystemPrompt()
	userPrompt, err := s.buildPolicyUserPrompt(policyContext)
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

	s.logger.Info("✅ Policy evaluation completed (compliant: %t)",
		evaluation.Compliant)

	return evaluation, nil
}

// OptimizeDeploymentPlan refines an existing plan using AI
// Provider-agnostic optimization logic
func (s *AIService) OptimizeDeploymentPlan(ctx context.Context, plan *DeploymentPlan, context *PlanningContext) (*PlanningResponse, error) {
	s.logger.Info("⚡ Optimizing deployment plan using AI")

	// Build provider-agnostic prompts
	systemPrompt := s.buildOptimizationSystemPrompt()
	userPrompt, err := s.buildOptimizationPrompt(plan, context)
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

// *** PROMPT BUILDING METHODS ***

// buildPlanningSystemPrompt creates system prompt for deployment planning
func (s *AIService) buildPlanningSystemPrompt() string {
	return `You are an AI deployment planning assistant. Generate a deployment plan as JSON with the following structure:
{
  "plan": {
    "id": "unique-plan-id",
    "application_id": "app-id",
    "environment": "target-environment",
    "steps": [
      {
        "id": "step-id",
        "type": "deploy|test|validate",
        "description": "step description",
        "command": "command to execute",
        "dependencies": ["previous-step-ids"],
        "metadata": {}
      }
    ]
  },
  "confidence": 0.9,
  "metadata": {}
}`
}

// buildPlanningUserPrompt creates user prompt for deployment planning
func (s *AIService) buildPlanningUserPrompt(request *PlanningRequest) (string, error) {
	if request == nil {
		return "", fmt.Errorf("planning request cannot be nil")
	}

	environment := "unknown"
	if request.Context != nil && request.Context.EnvironmentID != "" {
		environment = request.Context.EnvironmentID
	}

	return fmt.Sprintf(`Create a deployment plan for:
Application: %s
Environment: %s
Intent: %s
Edge Types: %v
Context: %v

Please provide a structured deployment plan with proper dependencies.`,
		request.ApplicationID, environment, request.Intent, request.EdgeTypes, request.Context), nil
}

// buildPolicySystemPrompt creates system prompt for policy evaluation
func (s *AIService) buildPolicySystemPrompt() string {
	return `You are an AI policy evaluation assistant. Evaluate policy compliance and return JSON:
{
  "compliant": true/false,
  "violations": [
    {
      "policy": "policy-name",
      "reason": "violation reason",
      "severity": "high|medium|low",
      "remediation": "how to fix"
    }
  ],
  "warnings": [
    {
      "policy": "policy-name",
      "message": "warning message"
    }
  ],
  "suggestions": ["suggestion1", "suggestion2"],
  "metadata": {}
}`
}

// buildPolicyUserPrompt creates user prompt for policy evaluation
func (s *AIService) buildPolicyUserPrompt(policyContext interface{}) (string, error) {
	return fmt.Sprintf(`Evaluate policy compliance for:
Context: %v

Please check all applicable policies and return compliance status.`, policyContext), nil
}

// buildOptimizationSystemPrompt creates system prompt for plan optimization
func (s *AIService) buildOptimizationSystemPrompt() string {
	return `You are an AI deployment optimization assistant. Optimize the given deployment plan and return the improved version as JSON with the same structure as the input plan.`
}

// buildOptimizationPrompt creates user prompt for plan optimization
func (s *AIService) buildOptimizationPrompt(plan *DeploymentPlan, context *PlanningContext) (string, error) {
	if plan == nil {
		return "", fmt.Errorf("deployment plan cannot be nil")
	}

	planJSON, err := json.Marshal(plan)
	if err != nil {
		return "", fmt.Errorf("failed to marshal plan: %w", err)
	}

	return fmt.Sprintf(`Optimize this deployment plan:
Current Plan: %s
Context: %v

Please provide an optimized version with better efficiency and reliability.`,
		string(planJSON), context), nil
}
