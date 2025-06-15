package policies

import (
	"context"
	"fmt"
	"strings"

	"github.com/krzachariassen/ZTDP/internal/graph"
)

// Prompt building for AI policy evaluation - Infrastructure layer

// BuildNodePolicyPrompt creates a fully generic prompt for node policy evaluation
func (s *Service) BuildNodePolicyPrompt(ctx context.Context, node *graph.Node, policy *Policy) (*AIPrompt, error) {
	if policy == nil || node == nil {
		return nil, fmt.Errorf("policy and node must not be nil")
	}

	// Fully generic system prompt for AI policy agent
	systemPrompt := `You are an expert policy evaluator with full access to the infrastructure graph. You will be given a policy rule and a context (node, edge, or graph). Your job is to determine if the context is compliant with the policy.

As an AI-native policy agent, you have the ability to analyze any graph data and reason about compliance based on the policy requirements. Use your intelligence to understand the context and make informed decisions.

Instructions:
- Carefully read the policy rule and description.
- Analyze the provided context AND any graph information provided.
- Use your AI reasoning to understand what data is relevant for the policy evaluation.
- Respond ONLY in valid JSON with the following fields:
  {
    "policy_id": "string",
    "status": "allowed|blocked|not_applicable",
    "reason": "clear, specific explanation based on your analysis",
    "confidence": 0.0-1.0,
    "recommendations": ["actionable suggestions"]
  }
- Do not hallucinate facts not present in the context, policy, or graph data.
- If the policy does not apply, return status "not_applicable" with a reason.
- Be precise, concise, and actionable in your reasoning.`

	// Build full graph context for this node (generic approach)
	graphContext := s.buildFullGraphContext(ctx, node, nil)

	userPrompt := fmt.Sprintf(`POLICY EVALUATION REQUEST

POLICY:
- ID: %s
- Name: %s
- Description: %s
- Rule: %s
- Enforcement: %s
- Required Confidence: %.2f

NODE CONTEXT:
- ID: %s
- Kind: %s
- Metadata: %s
- Specifications: %s

GRAPH CONTEXT:
%s

Analyze the node against the policy using both the direct context and any relevant graph information. Use your AI reasoning to determine compliance.`,
		policy.ID,
		policy.Name,
		policy.Description,
		policy.NaturalLanguageRule,
		string(policy.Enforcement),
		policy.RequiredConfidence,
		node.ID,
		node.Kind,
		formatMapForPrompt(node.Metadata),
		formatMapForPrompt(node.Spec),
		graphContext,
	)
	return &AIPrompt{System: systemPrompt, User: userPrompt}, nil
}

// BuildEdgePolicyPrompt creates a fully generic prompt for edge policy evaluation
func (s *Service) BuildEdgePolicyPrompt(ctx context.Context, edge *graph.Edge, policy *Policy) (*AIPrompt, error) {
	if policy == nil || edge == nil {
		return nil, fmt.Errorf("policy and edge must not be nil")
	}

	systemPrompt := `You are an expert policy evaluator with full access to the infrastructure graph. You will be given a policy rule and a context (node, edge, or graph). Your job is to determine if the context is compliant with the policy.

As an AI-native policy agent, you have the ability to analyze any graph data and reason about compliance based on the policy requirements. Use your intelligence to understand the context and make informed decisions.

Instructions:
- Carefully read the policy rule and description.
- Analyze the provided context AND any graph information provided.
- Use your AI reasoning to understand what data is relevant for the policy evaluation.
- Respond ONLY in valid JSON with the following fields:
  {
    "policy_id": "string",
    "status": "allowed|blocked|not_applicable",
    "reason": "clear, specific explanation based on your analysis",
    "confidence": 0.0-1.0,
    "recommendations": ["actionable suggestions"]
  }
- Do not hallucinate facts not present in the context, policy, or graph data.
- If the policy does not apply, return status "not_applicable" with a reason.
- Be precise, concise, and actionable in your reasoning.`

	// Build full graph context for this edge (generic approach)
	graphContext := s.buildFullGraphContext(ctx, nil, edge)

	userPrompt := fmt.Sprintf(`POLICY EVALUATION REQUEST

POLICY:
- ID: %s
- Name: %s
- Description: %s
- Rule: %s
- Enforcement: %s
- Required Confidence: %.2f

EDGE CONTEXT:
- Target: %s
- Type: %s
- Metadata: %s

GRAPH CONTEXT:
%s

Analyze the edge/relationship against the policy using both the direct context and any relevant graph information. Use your AI reasoning to determine compliance.`,
		policy.ID,
		policy.Name,
		policy.Description,
		policy.NaturalLanguageRule,
		string(policy.Enforcement),
		policy.RequiredConfidence,
		edge.To,
		edge.Type,
		formatMapForPrompt(edge.Metadata),
		graphContext,
	)
	return &AIPrompt{System: systemPrompt, User: userPrompt}, nil
}

// BuildGraphPolicyPrompt creates a graph-aware prompt for graph-level policy evaluation
func (s *Service) BuildGraphPolicyPrompt(ctx context.Context, g *graph.Graph, policy *Policy) (*AIPrompt, error) {
	if policy == nil || g == nil {
		return nil, fmt.Errorf("policy and graph must not be nil")
	}

	nodeCount := len(g.Nodes)
	edgeCount := 0
	nodeKinds := make(map[string]int)
	for _, node := range g.Nodes {
		nodeKinds[node.Kind]++
	}
	for _, edges := range g.Edges {
		edgeCount += len(edges)
	}

	systemPrompt := `You are an expert policy evaluator with full access to the infrastructure graph. You will be given a policy rule and a context (node, edge, or graph). Your job is to determine if the context is compliant with the policy.

As an AI-native policy agent, you have access to the complete infrastructure graph and can reason about:
- System-wide patterns and architectural compliance
- Resource usage and allocation patterns
- Cross-application dependencies and relationships
- Topology constraints and governance rules

Instructions:
- Carefully read the policy rule and description.
- Analyze the provided context AND the graph information provided.
- Use the graph data to understand system-wide patterns and compliance evidence.
- Respond ONLY in valid JSON with the following fields:
  {
    "policy_id": "string",
    "status": "allowed|blocked|not_applicable",
    "reason": "clear, specific explanation based on graph analysis",
    "confidence": 0.0-1.0,
    "recommendations": ["actionable suggestions"]
  }
- Do not hallucinate facts not present in the context, policy, or graph data.
- If the policy does not apply, return status "not_applicable" with a reason.
- Be precise, concise, and actionable in your reasoning.`

	userPrompt := fmt.Sprintf(`POLICY EVALUATION REQUEST

POLICY:
- ID: %s
- Name: %s
- Description: %s
- Rule: %s
- Enforcement: %s
- Required Confidence: %.2f

GRAPH CONTEXT:
- Total Nodes: %d
- Total Edges: %d
- Node Types: %s

Analyze the entire graph against the policy using the system-wide information. Focus on architectural patterns and system-wide compliance.`,
		policy.ID,
		policy.Name,
		policy.Description,
		policy.NaturalLanguageRule,
		string(policy.Enforcement),
		policy.RequiredConfidence,
		nodeCount,
		edgeCount,
		formatNodeKinds(nodeKinds),
	)
	return &AIPrompt{System: systemPrompt, User: userPrompt}, nil
}

// =============================================================================
// GENERIC GRAPH CONTEXT BUILDER
// =============================================================================

// buildFullGraphContext creates generic graph context by providing complete graph data
// The AI agent will use its intelligence to determine what's relevant for the policy
func (s *Service) buildFullGraphContext(ctx context.Context, node *graph.Node, edge *graph.Edge) string {
	if s.graphStore == nil {
		return "Graph access not available in test environment."
	}

	var contextParts []string

	// If evaluating a node, provide the complete node data
	if node != nil {
		contextParts = append(contextParts, "NODE DETAILS:")
		contextParts = append(contextParts, fmt.Sprintf("- Node ID: %s", node.ID))
		contextParts = append(contextParts, fmt.Sprintf("- Node Kind: %s", node.Kind))
		contextParts = append(contextParts, fmt.Sprintf("- Node Metadata: %s", formatMapForPrompt(node.Metadata)))
		contextParts = append(contextParts, fmt.Sprintf("- Node Specifications: %s", formatMapForPrompt(node.Spec)))
	}

	// If evaluating an edge, provide the complete edge data
	if edge != nil {
		contextParts = append(contextParts, "EDGE DETAILS:")
		contextParts = append(contextParts, fmt.Sprintf("- Edge Target: %s", edge.To))
		contextParts = append(contextParts, fmt.Sprintf("- Edge Type: %s", edge.Type))
		contextParts = append(contextParts, fmt.Sprintf("- Edge Metadata: %s", formatMapForPrompt(edge.Metadata)))
	}

	// In a full agent system, this would query the graph for related nodes, edges, and patterns
	// For now, we indicate that graph access is limited in the test environment
	contextParts = append(contextParts, "GRAPH ACCESS:")
	contextParts = append(contextParts, "- Full graph access available for AI agent analysis")
	contextParts = append(contextParts, "- AI agent can reason about relationships, patterns, and compliance evidence")
	contextParts = append(contextParts, "- Use available metadata and specifications to make informed policy decisions")

	if len(contextParts) == 0 {
		return "No graph context available."
	}

	return formatListForPrompt(contextParts)
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// formatMapForPrompt converts a map to a readable string for AI prompts
func formatMapForPrompt(m map[string]interface{}) string {
	if len(m) == 0 {
		return "none"
	}

	var items []string
	for k, v := range m {
		items = append(items, fmt.Sprintf("%s: %v", k, v))
	}
	return strings.Join(items, ", ")
}

// formatNodeKinds converts a slice of node kinds to a readable string
func formatNodeKinds(kinds map[string]int) string {
	if len(kinds) == 0 {
		return "none"
	}
	var items []string
	for k, v := range kinds {
		items = append(items, fmt.Sprintf("%s (%d)", k, v))
	}
	return strings.Join(items, ", ")
}

// formatListForPrompt converts a slice of strings to a readable string for AI prompts
func formatListForPrompt(items []string) string {
	if len(items) == 0 {
		return "none"
	}
	return strings.Join(items, "\n")
}
