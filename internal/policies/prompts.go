package policies

import (
	"encoding/json"
	"fmt"

	"github.com/krzachariassen/ZTDP/internal/ai"
)

// buildPolicySystemPrompt creates the system prompt for policy evaluation
func buildPolicySystemPrompt() string {
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
func buildPolicyUserPrompt(policyContext interface{}) (string, error) {
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
4. Provide actionable recommendations for achieving compliance
5. Consider the business context and practical constraints

Focus on practical, actionable guidance that helps achieve compliance while maintaining developer productivity.`,
		string(contextJSON))

	return prompt, nil
}

// parsePolicyEvaluation parses AI response into PolicyEvaluation
func parsePolicyEvaluation(response string) (*ai.PolicyEvaluation, error) {
	var evaluation ai.PolicyEvaluation

	if err := json.Unmarshal([]byte(response), &evaluation); err != nil {
		return nil, fmt.Errorf("failed to parse policy evaluation: %w", err)
	}

	return &evaluation, nil
}
