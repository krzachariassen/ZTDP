package deployments

import (
	"encoding/json"
	"fmt"

	"github.com/krzachariassen/ZTDP/internal/ai"
)

// buildPlanningSystemPrompt creates the system prompt for deployment planning
func buildPlanningSystemPrompt() string {
	return `You are an expert AI deployment planner for ZTDP (Zero Touch Developer Platform). Your role is to analyze deployment intents and generate intelligent, ordered deployment plans.

CONTEXT:
- ZTDP uses a graph-based architecture where applications, services, environments, and resources are nodes
- Edges define relationships: deploy, owns, has_version, uses, create, etc.
- You must consider dependencies, policies, and best practices when planning deployments

CAPABILITIES:
- Analyze complex deployment scenarios with deep reasoning
- Generate ordered deployment steps that respect dependencies
- Consider policy constraints and governance requirements
- Provide clear reasoning for each decision
- Suggest rollback strategies and validation steps

RESPONSE FORMAT:
You must respond with valid JSON only, following this exact structure:
{
  "plan": {
    "steps": [
      {
        "id": "step-1",
        "action": "deploy|create|configure|validate",
        "target": "node-id",
        "dependencies": ["step-ids"],
        "metadata": {},
        "reasoning": "why this step is needed"
      }
    ],
    "strategy": "rolling|blue-green|canary|all-at-once",
    "validation": ["validation steps"],
    "rollback": {
      "steps": [],
      "triggers": ["failure conditions"],
      "metadata": {}
    },
    "metadata": {}
  },
  "reasoning": "overall reasoning for the plan",
  "confidence": 0.95,
  "metadata": {}
}

PRINCIPLES:
1. Always ensure dependencies are deployed first
2. Consider rollback strategies for each step
3. Include validation checkpoints
4. Optimize for safety over speed
5. Explain your reasoning clearly
6. Handle edge cases gracefully`
}

// buildPlanningUserPrompt creates the user prompt with deployment context
func buildPlanningUserPrompt(context *ai.PlanningContext) (string, error) {
	contextJSON, err := json.MarshalIndent(context, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize planning context: %w", err)
	}

	prompt := fmt.Sprintf(`Please create a deployment plan for this context:

DEPLOYMENT CONTEXT:
%s

PLANNING REQUIREMENTS:
1. Analyze the deployment requirements and dependencies
2. Generate an ordered sequence of deployment steps
3. Consider policy constraints and governance requirements
4. Include validation and rollback strategies
5. Optimize for safety and reliability
6. Provide clear reasoning for the plan structure

Focus on creating a robust, executable deployment plan that minimizes risk while ensuring all dependencies are properly handled.`,
		string(contextJSON))

	return prompt, nil
}

// buildOptimizationPrompt creates the prompt for plan optimization
func buildOptimizationPrompt(plan interface{}, context *ai.PlanningContext) (string, error) {
	planJSON, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize plan: %w", err)
	}

	contextJSON, err := json.MarshalIndent(context, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize context: %w", err)
	}

	prompt := fmt.Sprintf(`Please optimize this deployment plan:

CURRENT PLAN:
%s

DEPLOYMENT CONTEXT:
%s

OPTIMIZATION REQUIREMENTS:
1. Analyze the current plan for inefficiencies
2. Identify opportunities for parallelization
3. Improve rollback and validation strategies
4. Enhance safety and risk management
5. Optimize for performance and reliability
6. Maintain dependency relationships and constraints

Focus on practical improvements that reduce deployment time and risk while maintaining safety and compliance.`,
		string(planJSON),
		string(contextJSON))

	return prompt, nil
}

// parsePlanningResponse parses AI response into PlanningResponse
func parsePlanningResponse(response string) (*ai.PlanningResponse, error) {
	var planResponse ai.PlanningResponse

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

// Helper functions for safe data extraction
func safeGetString(context *ai.PlanningContext, field string) string {
	if context == nil {
		return "unknown"
	}

	switch field {
	case "environment_id":
		return context.EnvironmentID
	default:
		return "unknown"
	}
}

func safeGetCount(context *ai.PlanningContext, field string) int {
	if context == nil {
		return 0
	}

	switch field {
	case "target_nodes":
		return len(context.TargetNodes)
	case "edges":
		return len(context.Edges)
	default:
		return 0
	}
}
