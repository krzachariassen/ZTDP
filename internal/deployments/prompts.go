package deployments

import (
	"encoding/json"
	"fmt"
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
- Generate ordered step-by-step deployment plans
- Consider rollback strategies and failure scenarios
- Optimize for safety, performance, and reliability

OUTPUT FORMAT:
Your response must be valid JSON with this structure:
{
  "steps": [
    {
      "action": "create|deploy|update|verify|rollback",
      "resource": "resource_name",
      "environment": "target_environment",
      "description": "human readable description",
      "dependencies": ["step1", "step2"],
      "rollback_plan": "rollback instructions",
      "verification": "how to verify success"
    }
  ],
  "estimated_duration": "15m",
  "risk_level": "low|medium|high",
  "rollback_strategy": "detailed rollback approach"
}`
}

// buildPlanningUserPrompt creates the user prompt for deployment planning
func buildPlanningUserPrompt(context interface{}) string {
	return fmt.Sprintf(`Please analyze this deployment request and generate an intelligent deployment plan:

REQUEST CONTEXT:
%s

Please provide a step-by-step deployment plan that:
1. Respects all dependencies and ordering requirements
2. Includes verification steps for each stage
3. Provides clear rollback instructions
4. Estimates timing and risk levels
5. Follows ZTDP best practices

Return your response as valid JSON only.`, formatContext(context))
}

// buildOptimizationPrompt creates prompts for plan optimization
func buildOptimizationPrompt(response interface{}) string {
	return fmt.Sprintf(`Analyze this deployment plan and suggest optimizations:

CURRENT PLAN:
%s

Please suggest improvements for:
1. Execution time reduction
2. Risk mitigation
3. Resource efficiency
4. Parallel execution opportunities
5. Better failure recovery

Provide specific, actionable recommendations.`, formatPlanningResponse(response))
}

// buildAnalysisPrompt creates prompts for deployment analysis
func buildAnalysisPrompt(context interface{}) string {
	return fmt.Sprintf(`Analyze this deployment scenario for potential issues:

DEPLOYMENT CONTEXT:
%s

Please identify:
1. Potential risks and failure points
2. Missing dependencies or prerequisites
3. Resource conflicts or constraints
4. Best practice violations
5. Optimization opportunities

Provide detailed analysis with recommendations.`, formatContext(context))
}

// Helper functions for formatting
func formatContext(context interface{}) string {
	if context == nil {
		return "No context provided"
	}
	
	data, err := json.MarshalIndent(context, "", "  ")
	if err != nil {
		return fmt.Sprintf("Context: %+v", context)
	}
	return string(data)
}

func formatPlanningResponse(response interface{}) string {
	if response == nil {
		return "No response provided"
	}
	
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Sprintf("Response: %+v", response)
	}
	return string(data)
}
