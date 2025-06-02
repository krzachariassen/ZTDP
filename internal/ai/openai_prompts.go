package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

// buildPlanningSystemPrompt creates the system prompt for deployment planning
func (p *OpenAIProvider) buildPlanningSystemPrompt() string {
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
1. Always respect node dependencies and edge relationships
2. Consider policy constraints and compliance requirements
3. Optimize for safety, reliability, and minimal downtime
4. Provide clear, actionable reasoning for all decisions
5. Include comprehensive rollback and validation strategies`
}

// buildPlanningUserPrompt creates the user prompt with deployment context
func (p *OpenAIProvider) buildPlanningUserPrompt(request *PlanningRequest) (string, error) {
	// Serialize the context for the AI
	contextJSON, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize context: %w", err)
	}

	prompt := fmt.Sprintf(`Please analyze this deployment request and generate an intelligent deployment plan:

DEPLOYMENT REQUEST:
%s

ANALYSIS REQUIREMENTS:
1. Examine the target application and its services
2. Identify all dependencies and relationships
3. Consider the specified edge types: %s
4. Respect policy constraints and governance rules
5. Generate an ordered sequence of deployment steps
6. Provide clear reasoning for the plan structure
7. Include validation checkpoints and rollback strategies

SPECIFIC CONSIDERATIONS:
- Target Application: %s
- Environment: %s
- Intent: %s
- Available Nodes: %d
- Edge Relationships: %d

Generate a comprehensive deployment plan that ensures safe, reliable deployment with minimal risk.`,
		string(contextJSON),
		strings.Join(request.EdgeTypes, ", "),
		request.ApplicationID,
		safeGetString(request.Context, "environment_id"),
		request.Intent,
		safeGetCount(request.Context, "target_nodes"),
		safeGetCount(request.Context, "edges"))

	return prompt, nil
}

// buildPolicySystemPrompt creates the system prompt for policy evaluation
func (p *OpenAIProvider) buildPolicySystemPrompt() string {
	return `You are an expert AI policy evaluator for ZTDP (Zero Touch Developer Platform). Your role is to analyze policy compliance and provide intelligent recommendations.

CONTEXT:
- ZTDP enforces governance through graph-based policies
- Policies can be attached to nodes, edges, and transitions
- Your job is to evaluate compliance and suggest remediation

CAPABILITIES:
- Deep analysis of policy requirements and current state
- Intelligent compliance evaluation beyond simple rule matching
- Context-aware recommendations for policy satisfaction
- Risk assessment and mitigation strategies

RESPONSE FORMAT:
You must respond with valid JSON only:
{
  "compliant": true|false,
  "violations": ["list of violations"],
  "suggestions": ["remediation suggestions"],
  "reasoning": "detailed reasoning for evaluation",
  "confidence": 0.95,
  "metadata": {}
}

PRINCIPLES:
1. Understand the intent behind policies, not just literal rules
2. Consider context and nuanced scenarios
3. Provide actionable suggestions for compliance
4. Balance governance with developer productivity
5. Explain your reasoning clearly and thoroughly`
}

// buildPolicyUserPrompt creates the user prompt with policy context
func (p *OpenAIProvider) buildPolicyUserPrompt(policyContext interface{}) (string, error) {
	contextJSON, err := json.MarshalIndent(policyContext, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize policy context: %w", err)
	}

	prompt := fmt.Sprintf(`Please evaluate policy compliance for this context:

POLICY CONTEXT:
%s

EVALUATION REQUIREMENTS:
1. Analyze all applicable policies and their requirements
2. Assess current compliance status
3. Identify any violations or potential issues
4. Provide specific suggestions for achieving compliance
5. Consider the business context and intent behind policies
6. Evaluate risk levels and prioritize recommendations

Please provide a comprehensive policy evaluation with clear reasoning and actionable recommendations.`,
		string(contextJSON))

	return prompt, nil
}

// buildOptimizationSystemPrompt creates the system prompt for plan optimization
func (p *OpenAIProvider) buildOptimizationSystemPrompt() string {
	return `You are an expert AI plan optimizer for ZTDP (Zero Touch Developer Platform). Your role is to improve existing deployment plans.

CONTEXT:
- You receive an existing deployment plan that needs optimization
- Consider performance, safety, efficiency, and best practices
- You can reorder steps, add validation, improve rollback strategies

CAPABILITIES:
- Advanced optimization of deployment sequences
- Performance and efficiency improvements
- Enhanced safety and rollback strategies
- Risk reduction and validation enhancement

RESPONSE FORMAT:
You must respond with valid JSON only, using the same structure as plan generation:
{
  "plan": { /* optimized plan structure */ },
  "reasoning": "optimization reasoning",
  "confidence": 0.95,
  "metadata": { "optimization_type": "description" }
}

OPTIMIZATION PRINCIPLES:
1. Improve deployment speed while maintaining safety
2. Enhance rollback capabilities and validation
3. Reduce deployment risk and failure points
4. Optimize resource utilization and dependencies
5. Maintain clear reasoning for all changes`
}

// buildOptimizationUserPrompt creates the user prompt for plan optimization
func (p *OpenAIProvider) buildOptimizationUserPrompt(plan *DeploymentPlan, context *PlanningContext) (string, error) {
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

// parsePlanningResponse parses OpenAI response into PlanningResponse
func (p *OpenAIProvider) parsePlanningResponse(response string) (*PlanningResponse, error) {
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

// parsePolicyEvaluation parses OpenAI response into PolicyEvaluation
func (p *OpenAIProvider) parsePolicyEvaluation(response string) (*PolicyEvaluation, error) {
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

// Helper functions for safe data extraction
func safeGetString(context *PlanningContext, field string) string {
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

func safeGetCount(context *PlanningContext, field string) int {
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
