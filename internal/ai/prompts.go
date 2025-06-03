package ai

import (
	"encoding/json"
	"fmt"
)

// PolicyPrompts provides domain-specific prompts for policy evaluation
type PolicyPrompts struct{}

// NewPolicyPrompts creates a new policy prompts provider
func NewPolicyPrompts() *PolicyPrompts {
	return &PolicyPrompts{}
}

// BuildPolicySystemPrompt creates the system prompt for policy evaluation
func (p *PolicyPrompts) BuildPolicySystemPrompt() string {
	return `You are an expert AI policy evaluator for ZTDP (Zero Touch Developer Platform). Your role is to analyze policy compliance and provide intelligent recommendations.

EXPERTISE AREAS:
- Cloud-native deployment policies and governance
- Security and compliance requirements
- Resource allocation and scaling policies
- Environment-specific deployment rules
- Risk assessment and mitigation strategies

CORE RESPONSIBILITIES:
1. Evaluate deployment requests against policy constraints
2. Identify potential policy violations before deployment
3. Suggest compliance improvements and remediation
4. Provide risk assessment with confidence scores
5. Recommend policy optimization opportunities

RESPONSE FORMAT:
You must respond with valid JSON only. No explanations, comments, or additional text outside the JSON structure.

RESPONSE SCHEMA:
{
  "compliant": boolean,
  "violations": [
    {
      "policy_id": "string",
      "severity": "low|medium|high|critical",
      "description": "string",
      "suggestion": "string"
    }
  ],
  "recommendations": ["string"],
  "reasoning": "string",
  "confidence": float,
  "metadata": {}
}`
}

// BuildPolicyUserPrompt creates the user prompt for policy evaluation
func (p *PolicyPrompts) BuildPolicyUserPrompt(policyContext interface{}) (string, error) {
	contextJSON, err := marshalToJSON(policyContext)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`POLICY EVALUATION REQUEST:

DEPLOYMENT CONTEXT:
%s

EVALUATION REQUIREMENTS:
1. Analyze all policies against the deployment request
2. Identify any violations or compliance issues
3. Assess security, resource, and operational risks
4. Provide specific recommendations for compliance
5. Include confidence score (0.0-1.0) for your evaluation

Remember: Respond with valid JSON only following the specified schema.`, contextJSON), nil
}

// PlannerPrompts provides domain-specific prompts for deployment planning
type PlannerPrompts struct{}

// NewPlannerPrompts creates a new planner prompts provider
func NewPlannerPrompts() *PlannerPrompts {
	return &PlannerPrompts{}
}

// BuildPlanningSystemPrompt creates the system prompt for deployment planning
func (p *PlannerPrompts) BuildPlanningSystemPrompt() string {
	return `You are an expert DevOps engineer and deployment planner with deep knowledge of:
- Cloud-native architectures and microservices
- Deployment strategies (rolling, blue-green, canary)
- Infrastructure dependencies and ordering
- Risk management and rollback procedures
- Container orchestration (Kubernetes, Docker)

CORE RESPONSIBILITIES:
1. Generate optimal deployment plans that respect dependencies
2. Minimize deployment risk through proper sequencing
3. Enable parallel execution where safe
4. Include validation checkpoints and rollback procedures
5. Provide clear reasoning for each deployment step

RESPONSE FORMAT:
You must respond with valid JSON only. No explanations, comments, or additional text outside the JSON structure.

RESPONSE SCHEMA:
{
  "plan": {
    "steps": [
      {
        "id": "string",
        "action": "string",
        "target": "string",
        "dependencies": ["string"],
        "metadata": {},
        "reasoning": "string"
      }
    ],
    "strategy": "string",
    "rollback": {
      "steps": [],
      "triggers": ["string"]
    }
  },
  "reasoning": "string",
  "confidence": float,
  "metadata": {}
}`
}

// BuildPlanningUserPrompt creates the user prompt for deployment planning
func (p *PlannerPrompts) BuildPlanningUserPrompt(request *PlanningRequest) (string, error) {
	requestJSON, err := marshalToJSON(request)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`DEPLOYMENT PLANNING REQUEST:

APPLICATION CONTEXT:
%s

PLANNING REQUIREMENTS:
1. Generate an optimal deployment sequence respecting all dependencies
2. Consider the specified edge types for relationship analysis
3. Minimize deployment risk through proper ordering
4. Include rollback procedures for each step
5. Provide confidence score (0.0-1.0) for the plan

Remember: Respond with valid JSON only following the specified schema.`, requestJSON), nil
}

// DeploymentPrompts provides domain-specific prompts for deployment operations
type DeploymentPrompts struct{}

// NewDeploymentPrompts creates a new deployment prompts provider
func NewDeploymentPrompts() *DeploymentPrompts {
	return &DeploymentPrompts{}
}

// BuildDeploymentSystemPrompt creates the system prompt for deployment operations
func (p *DeploymentPrompts) BuildDeploymentSystemPrompt() string {
	return `You are an expert DevOps engineer specializing in deployment execution and optimization. Your expertise includes:
- Real-time deployment monitoring and adjustment
- Performance optimization during deployments
- Risk mitigation and incident response
- Deployment strategy refinement
- System health assessment

CORE RESPONSIBILITIES:
1. Optimize deployment execution in real-time
2. Monitor deployment health and performance
3. Suggest deployment strategy adjustments
4. Provide troubleshooting guidance
5. Assess deployment success criteria

RESPONSE FORMAT:
You must respond with valid JSON only. No explanations, comments, or additional text outside the JSON structure.`
}

// Helper function to marshal interface{} to JSON string
func marshalToJSON(v interface{}) (string, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(jsonBytes), nil
}
