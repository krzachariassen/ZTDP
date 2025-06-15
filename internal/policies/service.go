package policies

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

// Service provides AI-native policy evaluation capabilities
// This is the main domain service that orchestrates policy evaluation
type Service struct {
	aiProvider  ai.AIProvider
	graphStore  *graph.GraphStore
	globalGraph *graph.GlobalGraph
	policyStore PolicyStore
	env         string
	eventBus    EventBus
}

// NewService creates a new AI-native policy service
func NewService(graphStore *graph.GraphStore, globalGraph *graph.GlobalGraph, env string, eventBus EventBus) *Service {
	return NewServiceWithPolicyStore(graphStore, globalGraph, nil, env, eventBus)
}

// NewServiceWithPolicyStore creates a new AI-native policy service with a policy store
func NewServiceWithPolicyStore(graphStore *graph.GraphStore, globalGraph *graph.GlobalGraph, policyStore PolicyStore, env string, eventBus EventBus) *Service {
	// Initialize AI provider - REQUIRED for AI-native operation
	var aiProvider ai.AIProvider
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey != "" {
		config := ai.DefaultOpenAIConfig()
		if model := os.Getenv("OPENAI_MODEL"); model != "" {
			config.Model = model
		}
		if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
			config.BaseURL = baseURL
		}
		if provider, err := ai.NewOpenAIProvider(config, apiKey); err == nil {
			aiProvider = provider
		}
	}

	return &Service{
		aiProvider:  aiProvider,
		graphStore:  graphStore,
		globalGraph: globalGraph,
		policyStore: policyStore,
		env:         env,
		eventBus:    eventBus,
	}
}

// NewServiceWithAIProvider creates a new service with a custom AI provider (for testing)
func NewServiceWithAIProvider(graphStore *graph.GraphStore, globalGraph *graph.GlobalGraph, aiProvider ai.AIProvider, policyStore PolicyStore, env string, eventBus EventBus) *Service {
	return &Service{
		aiProvider:  aiProvider,
		graphStore:  graphStore,
		globalGraph: globalGraph,
		policyStore: policyStore,
		env:         env,
		eventBus:    eventBus,
	}
}

// =============================================================================
// BUSINESS LOGIC - Node Policy Evaluation
// =============================================================================

// EvaluateNodePolicy evaluates a single node against a single policy - AI NATIVE ONLY
func (s *Service) EvaluateNodePolicy(ctx context.Context, env string, node *graph.Node, policy *Policy) (*PolicyResult, error) {
	if s.aiProvider == nil {
		return nil, fmt.Errorf("AI provider not available - ZTDP is AI-native only")
	}

	// Emit evaluation started event
	if s.eventBus != nil {
		s.eventBus.Emit("policy.node.evaluation.started", map[string]interface{}{
			"node_id":   node.ID,
			"policy_id": policy.ID,
			"env":       env,
		})
	}

	// Check if policy applies to this node
	if policy.Scope != PolicyScopeNode || !s.isPolicyApplicableToNode(policy, node) {
		result := &PolicyResult{
			NodeID:        node.ID,
			NodeKind:      node.Kind,
			Environment:   env,
			OverallStatus: PolicyStatusNotApplicable,
			Status:        PolicyStatusNotApplicable,
			Evaluations:   make(map[string]*PolicyEvaluation),
			EvaluatedAt:   time.Now(),
			EvaluatedBy:   "system",
		}
		return result, nil
	}

	// Use AI evaluation infrastructure
	return s.evaluateNodePolicyWithAI(ctx, node, []*Policy{policy})
}

// EvaluateNode evaluates a node against all applicable policies
func (s *Service) EvaluateNode(ctx context.Context, env string, node *graph.Node) (*PolicyResult, error) {
	// Get applicable policies for this node type
	var policies []*Policy
	var err error

	if s.policyStore != nil {
		policies, err = s.policyStore.GetPoliciesForNodeType(node.Kind)
		if err != nil {
			// Fall back to test policies if store fails
			policies = s.getTestNodePolicies()
		}
	} else {
		// Use test policies if no store available
		policies = s.getTestNodePolicies()
	}

	// Filter to applicable policies
	var applicablePolicies []*Policy
	for _, policy := range policies {
		if policy.Scope == PolicyScopeNode && s.isPolicyApplicableToNode(policy, node) {
			applicablePolicies = append(applicablePolicies, policy)
		}
	}

	if len(applicablePolicies) == 0 {
		return &PolicyResult{
			NodeID:        node.ID,
			NodeKind:      node.Kind,
			Environment:   env,
			OverallStatus: PolicyStatusNotApplicable,
			Evaluations:   make(map[string]*PolicyEvaluation),
			EvaluatedAt:   time.Now(),
		}, nil
	}

	// Use AI evaluation infrastructure
	return s.evaluateNodePolicyWithAI(ctx, node, applicablePolicies)
}

// =============================================================================
// BUSINESS LOGIC - Edge Policy Evaluation
// =============================================================================

// EvaluateEdgePolicy evaluates a single edge against a single policy
func (s *Service) EvaluateEdgePolicy(ctx context.Context, env string, edge *graph.Edge, policy *Policy) (*PolicyResult, error) {
	if s.aiProvider == nil {
		return nil, fmt.Errorf("AI provider not available - ZTDP is AI-native only")
	}

	// Check if policy applies to this edge
	if policy.Scope != PolicyScopeEdge || !s.isPolicyApplicableToEdge(policy, edge) {
		return &PolicyResult{
			EdgeTo:        edge.To,
			Relationship:  edge.Type,
			Environment:   env,
			OverallStatus: PolicyStatusNotApplicable,
			Status:        PolicyStatusNotApplicable,
			Evaluations:   make(map[string]*PolicyEvaluation),
			EvaluatedAt:   time.Now(),
			EvaluatedBy:   "system",
		}, nil
	}

	// Use AI evaluation infrastructure
	return s.evaluateEdgePolicyWithAI(ctx, edge, []*Policy{policy})
}

// EvaluateEdge evaluates an edge against all applicable policies
func (s *Service) EvaluateEdge(ctx context.Context, env string, edge *graph.Edge) (*PolicyResult, error) {
	// Get applicable policies for this edge type
	var policies []*Policy
	var err error

	if s.policyStore != nil {
		policies, err = s.policyStore.GetPoliciesForEdgeType(edge.Type)
		if err != nil {
			// Fall back to test policies if store fails
			policies = s.getTestEdgePolicies()
		}
	} else {
		// Use test policies if no store available
		policies = s.getTestEdgePolicies()
	}

	// Filter to applicable policies
	var applicablePolicies []*Policy
	for _, policy := range policies {
		if policy.Scope == PolicyScopeEdge && s.isPolicyApplicableToEdge(policy, edge) {
			applicablePolicies = append(applicablePolicies, policy)
		}
	}

	if len(applicablePolicies) == 0 {
		return &PolicyResult{
			EdgeTo:        edge.To,
			Relationship:  edge.Type,
			Environment:   env,
			OverallStatus: PolicyStatusNotApplicable,
			Evaluations:   make(map[string]*PolicyEvaluation),
			EvaluatedAt:   time.Now(),
		}, nil
	}

	// Use AI evaluation infrastructure
	return s.evaluateEdgePolicyWithAI(ctx, edge, applicablePolicies)
}

// =============================================================================
// BUSINESS LOGIC - Graph Policy Evaluation
// =============================================================================

// EvaluateGraphPolicy evaluates a graph against a single policy
func (s *Service) EvaluateGraphPolicy(ctx context.Context, env string, g *graph.Graph, policy *Policy) (*PolicyResult, error) {
	if s.aiProvider == nil {
		return nil, fmt.Errorf("AI provider not available - ZTDP is AI-native only")
	}

	// Check if policy applies to graphs
	if policy.Scope != PolicyScopeGraph {
		return &PolicyResult{
			GraphScope:    true,
			Environment:   env,
			OverallStatus: PolicyStatusNotApplicable,
			Status:        PolicyStatusNotApplicable,
			Evaluations:   make(map[string]*PolicyEvaluation),
			EvaluatedAt:   time.Now(),
			EvaluatedBy:   "system",
		}, nil
	}

	// Use AI evaluation infrastructure
	return s.evaluateGraphPolicyWithAI(ctx, g, []*Policy{policy})
}

// EvaluateGraph evaluates a graph against all applicable policies
func (s *Service) EvaluateGraph(ctx context.Context, env string, g *graph.Graph) (*PolicyResult, error) {
	// Get applicable policies for graphs
	var policies []*Policy
	var err error

	if s.policyStore != nil {
		policies, err = s.policyStore.GetGraphPolicies()
		if err != nil {
			// Fall back to test policies if store fails
			policies = s.getTestGraphPolicies()
		}
	} else {
		// Use test policies if no store available
		policies = s.getTestGraphPolicies()
	}

	// Filter to graph-scope policies
	var applicablePolicies []*Policy
	for _, policy := range policies {
		if policy.Scope == PolicyScopeGraph {
			applicablePolicies = append(applicablePolicies, policy)
		}
	}

	if len(applicablePolicies) == 0 {
		return &PolicyResult{
			GraphScope:    true,
			Environment:   env,
			OverallStatus: PolicyStatusNotApplicable,
			Evaluations:   make(map[string]*PolicyEvaluation),
			EvaluatedAt:   time.Now(),
		}, nil
	}

	// Use AI evaluation infrastructure
	return s.evaluateGraphPolicyWithAI(ctx, g, applicablePolicies)
}

// =============================================================================
// BUSINESS LOGIC HELPERS
// =============================================================================

// isPolicyApplicableToNode checks if a policy applies to a specific node
func (s *Service) isPolicyApplicableToNode(policy *Policy, node *graph.Node) bool {
	// If no specific node types specified, applies to all
	if len(policy.NodeTypes) == 0 {
		return true
	}

	// Check if node kind matches any specified types
	for _, nodeType := range policy.NodeTypes {
		if nodeType == node.Kind {
			return true
		}
	}
	return false
}

// isPolicyApplicableToEdge checks if a policy applies to a specific edge
func (s *Service) isPolicyApplicableToEdge(policy *Policy, edge *graph.Edge) bool {
	// If no specific edge types specified, applies to all
	if len(policy.EdgeTypes) == 0 {
		return true
	}

	// Check if edge relationship matches any specified types
	for _, edgeType := range policy.EdgeTypes {
		if edgeType == edge.Type {
			return true
		}
	}
	return false
}

// =============================================================================
// TEMPORARY TEST HELPERS - TODO: Replace with PolicyStore
// =============================================================================

// TODO: These should be replaced with a proper PolicyStore implementation
func (s *Service) getTestNodePolicies() []*Policy {
	return []*Policy{
		{
			ID:                  "node-security-check",
			Name:                "Node Security Check",
			Description:         "Ensures all nodes meet security requirements",
			Scope:               PolicyScopeNode,
			NaturalLanguageRule: "All nodes must have security configurations and valid authentication",
			Enforcement:         EnforcementBlock,
			RequiredConfidence:  0.8,
			CreatedAt:           time.Now(),
			Enabled:             true,
		},
		{
			ID:                  "app-service-limit",
			Name:                "Application Service Limit",
			Description:         "Applications must have fewer than 10 services",
			Scope:               PolicyScopeNode,
			NodeTypes:           []string{"application"},
			NaturalLanguageRule: "Applications must have fewer than 10 services to maintain manageable complexity",
			Enforcement:         EnforcementBlock,
			RequiredConfidence:  0.8,
			CreatedAt:           time.Now(),
			Enabled:             true,
		},
	}
}

func (s *Service) getTestEdgePolicies() []*Policy {
	return []*Policy{
		{
			ID:                  "edge-connection-policy",
			Name:                "Edge Connection Policy",
			Description:         "Validates edge connections follow architectural patterns",
			Scope:               PolicyScopeEdge,
			NaturalLanguageRule: "Edges must connect compatible node types and maintain architectural integrity",
			Enforcement:         EnforcementBlock,
			RequiredConfidence:  0.8,
			CreatedAt:           time.Now(),
			Enabled:             true,
		},
		{
			ID:                  "no-direct-prod-deployment",
			Name:                "No Direct Production Deployment",
			Description:         "Prevents direct deployment to production without staging",
			Scope:               PolicyScopeEdge,
			EdgeTypes:           []string{"deploys_to"},
			NaturalLanguageRule: "Direct deployment to production is not allowed - must go through staging first",
			Enforcement:         EnforcementBlock,
			RequiredConfidence:  0.8,
			CreatedAt:           time.Now(),
			Enabled:             true,
		},
	}
}

func (s *Service) getTestGraphPolicies() []*Policy {
	return []*Policy{
		{
			ID:                  "graph-compliance-policy",
			Name:                "Graph Compliance Policy",
			Description:         "Ensures overall graph structure meets compliance requirements",
			Scope:               PolicyScopeGraph,
			NaturalLanguageRule: "Graph must maintain security boundaries and compliance with regulatory requirements",
			Enforcement:         EnforcementBlock,
			RequiredConfidence:  0.8,
			CreatedAt:           time.Now(),
			Enabled:             true,
		},
		{
			ID:                  "max-apps-per-customer",
			Name:                "Maximum Applications Per Customer",
			Description:         "Limits the number of applications per customer",
			Scope:               PolicyScopeGraph,
			NaturalLanguageRule: "Each customer should not have more than 5 applications to maintain manageable complexity",
			Enforcement:         EnforcementBlock,
			RequiredConfidence:  0.8,
			CreatedAt:           time.Now(),
			Enabled:             true,
		},
	}
}
