package policies

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/krzachariassen/ZTDP/internal/graph"
)

// AI-specific evaluation logic - Infrastructure layer for AI operations

// evaluateNodePolicyWithAI uses AI to evaluate node policies
func (s *Service) evaluateNodePolicyWithAI(ctx context.Context, node *graph.Node, policies []*Policy) (*PolicyResult, error) {
	if s.aiProvider == nil {
		return nil, fmt.Errorf("AI provider not available - ZTDP is AI-native only")
	}

	result := &PolicyResult{
		NodeID:      node.ID,
		NodeKind:    node.Kind,
		Environment: s.env,
		Evaluations: make(map[string]*PolicyEvaluation),
		EvaluatedAt: time.Now(),
		EvaluatedBy: "ai-system",
	}

	overallStatus := PolicyStatusAllowed
	for _, policy := range policies {
		prompt, err := s.BuildNodePolicyPrompt(ctx, node, policy)
		if err != nil {
			continue // Skip policies that can't generate prompts
		}

		response, err := s.aiProvider.CallAI(ctx, prompt.System, prompt.User)
		if err != nil {
			continue // Skip policies with AI failures
		}

		evaluation, err := s.ParseAIResponse(response)
		if err != nil {
			continue // Skip unparseable responses
		}

		result.Evaluations[policy.ID] = evaluation

		// Determine overall status priority: blocked > warning > allowed
		if evaluation.Status == PolicyStatusBlocked {
			overallStatus = PolicyStatusBlocked
		} else if evaluation.Status == PolicyStatusWarning && overallStatus != PolicyStatusBlocked {
			overallStatus = PolicyStatusWarning
		}

		// For single policy evaluations, populate direct result fields for test compatibility
		if len(policies) == 1 {
			result.Status = evaluation.Status
			result.Confidence = evaluation.Confidence
			result.AIReasoning = evaluation.AIReasoning
			result.Reason = evaluation.Reason
		}
	}

	result.OverallStatus = overallStatus
	return result, nil
}

// evaluateEdgePolicyWithAI uses AI to evaluate edge policies
func (s *Service) evaluateEdgePolicyWithAI(ctx context.Context, edge *graph.Edge, policies []*Policy) (*PolicyResult, error) {
	if s.aiProvider == nil {
		return nil, fmt.Errorf("AI provider not available - ZTDP is AI-native only")
	}

	result := &PolicyResult{
		EdgeTo:       edge.To,
		Relationship: edge.Type,
		Environment:  s.env,
		Evaluations:  make(map[string]*PolicyEvaluation),
		EvaluatedAt:  time.Now(),
		EvaluatedBy:  "ai-system",
	}

	overallStatus := PolicyStatusAllowed
	for _, policy := range policies {
		prompt, err := s.BuildEdgePolicyPrompt(ctx, edge, policy)
		if err != nil {
			continue // Skip policies that can't generate prompts
		}

		response, err := s.aiProvider.CallAI(ctx, prompt.System, prompt.User)
		if err != nil {
			continue // Skip policies with AI failures
		}

		evaluation, err := s.ParseAIResponse(response)
		if err != nil {
			continue // Skip unparseable responses
		}

		result.Evaluations[policy.ID] = evaluation
		if evaluation.Status == PolicyStatusBlocked {
			overallStatus = PolicyStatusBlocked
		} else if evaluation.Status == PolicyStatusWarning && overallStatus != PolicyStatusBlocked {
			overallStatus = PolicyStatusWarning
		}

		// For single policy evaluations, populate direct result fields for test compatibility
		if len(policies) == 1 {
			result.Status = evaluation.Status
			result.Confidence = evaluation.Confidence
			result.AIReasoning = evaluation.AIReasoning
			result.Reason = evaluation.Reason
		}
	}

	result.OverallStatus = overallStatus
	return result, nil
}

// evaluateGraphPolicyWithAI uses AI to evaluate graph-level policies
func (s *Service) evaluateGraphPolicyWithAI(ctx context.Context, g *graph.Graph, policies []*Policy) (*PolicyResult, error) {
	if s.aiProvider == nil {
		return nil, fmt.Errorf("AI provider not available - ZTDP is AI-native only")
	}

	result := &PolicyResult{
		GraphScope:  true,
		Environment: s.env,
		Evaluations: make(map[string]*PolicyEvaluation),
		EvaluatedAt: time.Now(),
		EvaluatedBy: "ai-system",
	}

	overallStatus := PolicyStatusAllowed
	for _, policy := range policies {
		prompt, err := s.BuildGraphPolicyPrompt(ctx, g, policy)
		if err != nil {
			continue // Skip policies that can't generate prompts
		}

		response, err := s.aiProvider.CallAI(ctx, prompt.System, prompt.User)
		if err != nil {
			continue // Skip policies with AI failures
		}

		evaluation, err := s.ParseAIResponse(response)
		if err != nil {
			continue // Skip unparseable responses
		}

		result.Evaluations[policy.ID] = evaluation
		if evaluation.Status == PolicyStatusBlocked {
			overallStatus = PolicyStatusBlocked
		} else if evaluation.Status == PolicyStatusWarning && overallStatus != PolicyStatusBlocked {
			overallStatus = PolicyStatusWarning
		}

		// For single policy evaluations, populate direct result fields for test compatibility
		if len(policies) == 1 {
			result.Status = evaluation.Status
			result.Confidence = evaluation.Confidence
			result.AIReasoning = evaluation.AIReasoning
			result.Reason = evaluation.Reason
		}
	}

	result.OverallStatus = overallStatus
	return result, nil
}

// ParseAIResponse parses AI response into PolicyEvaluation
func (s *Service) ParseAIResponse(response string) (*PolicyEvaluation, error) {
	if response == "" {
		return nil, fmt.Errorf("AI returned empty response")
	}

	// Clean response - remove markdown code blocks if present
	cleanResponse := strings.TrimSpace(response)
	if strings.HasPrefix(cleanResponse, "```json") {
		cleanResponse = strings.TrimPrefix(cleanResponse, "```json")
		cleanResponse = strings.TrimSuffix(cleanResponse, "```")
		cleanResponse = strings.TrimSpace(cleanResponse)
	} else if strings.HasPrefix(cleanResponse, "```") {
		cleanResponse = strings.TrimPrefix(cleanResponse, "```")
		cleanResponse = strings.TrimSuffix(cleanResponse, "```")
		cleanResponse = strings.TrimSpace(cleanResponse)
	}

	// Parse as a flexible response that can handle different field names
	var rawResponse map[string]interface{}
	if err := json.Unmarshal([]byte(cleanResponse), &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Check for required status field
	statusVal, hasStatus := rawResponse["status"]
	if !hasStatus {
		return nil, fmt.Errorf("AI response missing required status field")
	}

	status, ok := statusVal.(string)
	if !ok || status == "" {
		return nil, fmt.Errorf("AI response has invalid status field")
	}

	// Extract other fields with flexible field names
	reason, _ := rawResponse["reason"].(string)
	confidence, _ := rawResponse["confidence"].(float64)

	// Handle both "reasoning" and "ai_reasoning" field names
	var aiReasoning string
	if val, exists := rawResponse["ai_reasoning"]; exists {
		aiReasoning, _ = val.(string)
	} else if val, exists := rawResponse["reasoning"]; exists {
		aiReasoning, _ = val.(string)
	}

	return &PolicyEvaluation{
		Status:      PolicyStatus(status),
		Reason:      reason,
		Confidence:  confidence,
		AIReasoning: aiReasoning,
		EvaluatedAt: time.Now(),
	}, nil
}
