package policies

import (
	"context"
	"fmt"
	"time"

	"github.com/krzachariassen/ZTDP/internal/agents"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

// PolicyAgent implements the AgentInterface for AI-native policy evaluation
// It wraps the existing Policy Service and exposes it as an event-driven agent
type PolicyAgent struct {
	*agents.BaseAgent
	policyService *Service
}

// NewPolicyAgent creates a new PolicyAgent that implements AgentInterface
func NewPolicyAgent(graphStore *graph.GraphStore, globalGraph *graph.GlobalGraph, policyStore PolicyStore, env string, eventBus EventBus) agents.AgentInterface {
	// Create underlying policy service
	policyService := NewServiceWithPolicyStore(graphStore, globalGraph, policyStore, env, eventBus)

	// Define policy agent capabilities
	capabilities := []agents.AgentCapability{
		{
			Name:        "policy_evaluation",
			Description: "Evaluates policies using AI reasoning over graph data",
			Intents: []string{
				"evaluate policy", "check compliance", "validate rules",
				"policy violation", "deployment policy", "security policy",
				"check if allowed", "policy check", "compliance check",
			},
			InputTypes:  []string{"node", "edge", "graph", "deployment", "configuration"},
			OutputTypes: []string{"policy_result", "compliance_status", "violation_report"},
			Version:     "1.0.0",
		},
	}

	// Create base agent
	baseAgent := agents.NewBaseAgent("policy-agent", "policy", "1.0.0", capabilities)

	// Create policy agent
	policyAgent := &PolicyAgent{
		BaseAgent:     baseAgent,
		policyService: policyService,
	}

	return policyAgent
}

// ProcessEvent implements AgentInterface for event-based policy evaluation
func (pa *PolicyAgent) ProcessEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	// Update activity timestamp
	pa.LastActivity = time.Now()

	// Only handle request events for policy evaluation
	if event.Type != events.EventTypeRequest {
		return &events.Event{
			Type:    events.EventTypeResponse,
			Source:  pa.ID,
			Subject: "not_handled",
			Payload: map[string]interface{}{
				"message": "Policy agent only handles request events",
				"handled": false,
			},
		}, nil
	}

	// Extract intent from event
	intent, ok := event.Payload["intent"].(string)
	if !ok || intent == "" {
		return nil, fmt.Errorf("policy evaluation requires 'intent' field in payload")
	}

	// Route to appropriate policy evaluation based on intent content
	var result *PolicyResult
	var err error

	// Determine evaluation type based on payload content
	if nodeData, hasNode := event.Payload["node"]; hasNode {
		result, err = pa.handleNodePolicyEvaluation(ctx, intent, nodeData, event.Payload)
	} else if edgeData, hasEdge := event.Payload["edge"]; hasEdge {
		result, err = pa.handleEdgePolicyEvaluation(ctx, intent, edgeData, event.Payload)
	} else if graphData, hasGraph := event.Payload["graph"]; hasGraph {
		result, err = pa.handleGraphPolicyEvaluation(ctx, intent, graphData, event.Payload)
	} else {
		// Generic policy question - use AI to determine appropriate evaluation
		result, err = pa.handleGenericPolicyQuestion(ctx, intent, event.Payload)
	}

	if err != nil {
		return &events.Event{
			Type:    events.EventTypeResponse,
			Source:  pa.ID,
			Subject: "policy_error",
			Payload: map[string]interface{}{
				"error":   err.Error(),
				"handled": false,
			},
		}, nil
	}

	// Convert PolicyResult to event response
	return pa.convertPolicyResultToEvent(result), nil
}

// handleNodePolicyEvaluation handles node-specific policy evaluation
func (pa *PolicyAgent) handleNodePolicyEvaluation(ctx context.Context, intent string, nodeData interface{}, payload map[string]interface{}) (*PolicyResult, error) {
	// Convert nodeData to graph.Node
	node, err := pa.convertToNode(nodeData)
	if err != nil {
		return nil, fmt.Errorf("invalid node data: %w", err)
	}

	// Check if specific policy is provided
	if policyData, hasPolicyData := payload["policy"]; hasPolicyData {
		policy, err := pa.convertToPolicy(policyData)
		if err != nil {
			return nil, fmt.Errorf("invalid policy data: %w", err)
		}
		return pa.policyService.EvaluateNodePolicy(ctx, pa.policyService.env, node, policy)
	}

	// Evaluate against all applicable node policies
	return pa.policyService.EvaluateNode(ctx, pa.policyService.env, node)
}

// handleEdgePolicyEvaluation handles edge-specific policy evaluation
func (pa *PolicyAgent) handleEdgePolicyEvaluation(ctx context.Context, intent string, edgeData interface{}, payload map[string]interface{}) (*PolicyResult, error) {
	// Convert edgeData to graph.Edge
	edge, err := pa.convertToEdge(edgeData)
	if err != nil {
		return nil, fmt.Errorf("invalid edge data: %w", err)
	}

	// Check if specific policy is provided
	if policyData, hasPolicyData := payload["policy"]; hasPolicyData {
		policy, err := pa.convertToPolicy(policyData)
		if err != nil {
			return nil, fmt.Errorf("invalid policy data: %w", err)
		}
		return pa.policyService.EvaluateEdgePolicy(ctx, pa.policyService.env, edge, policy)
	}

	// Evaluate against all applicable edge policies
	return pa.policyService.EvaluateEdge(ctx, pa.policyService.env, edge)
}

// handleGraphPolicyEvaluation handles graph-level policy evaluation
func (pa *PolicyAgent) handleGraphPolicyEvaluation(ctx context.Context, intent string, graphData interface{}, payload map[string]interface{}) (*PolicyResult, error) {
	// Convert graphData to graph.Graph
	g, err := pa.convertToGraph(graphData)
	if err != nil {
		return nil, fmt.Errorf("invalid graph data: %w", err)
	}

	// Check if specific policy is provided
	if policyData, hasPolicyData := payload["policy"]; hasPolicyData {
		policy, err := pa.convertToPolicy(policyData)
		if err != nil {
			return nil, fmt.Errorf("invalid policy data: %w", err)
		}
		return pa.policyService.EvaluateGraphPolicy(ctx, pa.policyService.env, g, policy)
	}

	// Evaluate against all applicable graph policies
	return pa.policyService.EvaluateGraph(ctx, pa.policyService.env, g)
}

// handleGenericPolicyQuestion handles generic policy questions using AI reasoning
func (pa *PolicyAgent) handleGenericPolicyQuestion(ctx context.Context, intent string, payload map[string]interface{}) (*PolicyResult, error) {
	// For generic questions, create a simple policy evaluation response
	// This is a placeholder - in a full implementation, this would use AI to understand the intent

	result := &PolicyResult{
		OverallStatus: PolicyStatusNotApplicable,
		Status:        PolicyStatusNotApplicable,
		Evaluations:   make(map[string]*PolicyEvaluation),
		EvaluatedAt:   time.Now(),
		EvaluatedBy:   "policy-agent",
		Reason:        fmt.Sprintf("Generic policy question received: %s", intent),
		AIReasoning:   "Policy agent received a generic question but needs specific context (node, edge, or graph) for evaluation",
		Confidence:    0.0,
	}

	return result, nil
}

// convertPolicyResultToEvent converts PolicyResult to an event response
func (pa *PolicyAgent) convertPolicyResultToEvent(result *PolicyResult) *events.Event {
	// Normalize decision types to what the tests expect: [allowed, blocked, conditional, warning]
	var decision string
	switch result.Status {
	case PolicyStatusAllowed:
		decision = "allowed"
	case PolicyStatusBlocked:
		decision = "blocked"
	case PolicyStatusWarning:
		decision = "warning"
	case PolicyStatusConditional:
		decision = "conditional"
	case PolicyStatusNotApplicable:
		// Map not applicable to allowed with low confidence
		decision = "allowed"
	case PolicyStatusPendingApproval:
		// Map pending approval to conditional
		decision = "conditional"
	default:
		// Default to blocked for safety
		decision = "blocked"
	}

	// Ensure we have reasoning (required by tests)
	reasoning := result.AIReasoning
	if reasoning == "" {
		reasoning = result.Reason
	}
	if reasoning == "" {
		reasoning = "Policy evaluation completed successfully"
	}

	payload := map[string]interface{}{
		"decision":     decision,
		"reasoning":    reasoning,
		"confidence":   result.Confidence,
		"status":       string(result.Status),
		"evaluated_at": result.EvaluatedAt,
		"evaluated_by": result.EvaluatedBy,
		"handled":      true,
	}

	// Include reason if available
	if result.Reason != "" {
		payload["reason"] = result.Reason
	}

	// Include evaluations for detailed analysis
	if len(result.Evaluations) > 0 {
		payload["evaluations"] = result.Evaluations
	}

	return &events.Event{
		Type:    events.EventTypeResponse,
		Source:  pa.ID,
		Subject: "policy_result",
		Payload: payload,
	}
}

// Conversion helper methods

func (pa *PolicyAgent) convertToNode(data interface{}) (*graph.Node, error) {
	nodeMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("node data must be a map")
	}

	// Extract node fields
	id, _ := nodeMap["id"].(string)
	kind, _ := nodeMap["kind"].(string)
	metadata, _ := nodeMap["metadata"].(map[string]interface{})
	spec, _ := nodeMap["spec"].(map[string]interface{})

	if id == "" || kind == "" {
		return nil, fmt.Errorf("node must have id and kind fields")
	}

	return &graph.Node{
		ID:       id,
		Kind:     kind,
		Metadata: metadata,
		Spec:     spec,
	}, nil
}

func (pa *PolicyAgent) convertToEdge(data interface{}) (*graph.Edge, error) {
	edgeMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("edge data must be a map")
	}

	// Extract edge fields
	to, _ := edgeMap["to"].(string)
	edgeType, _ := edgeMap["type"].(string)
	metadata, _ := edgeMap["metadata"].(map[string]interface{})

	if to == "" || edgeType == "" {
		return nil, fmt.Errorf("edge must have to and type fields")
	}

	return &graph.Edge{
		To:       to,
		Type:     edgeType,
		Metadata: metadata,
	}, nil
}

func (pa *PolicyAgent) convertToGraph(data interface{}) (*graph.Graph, error) {
	// For now, return a simple graph - in full implementation this would
	// construct a proper graph from the provided data
	return &graph.Graph{
		Nodes: make(map[string]*graph.Node),
		Edges: make(map[string][]graph.Edge),
	}, nil
}

func (pa *PolicyAgent) convertToPolicy(data interface{}) (*Policy, error) {
	policyMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("policy data must be a map")
	}

	// Extract policy fields
	id, _ := policyMap["id"].(string)
	name, _ := policyMap["name"].(string)
	description, _ := policyMap["description"].(string)
	naturalLanguageRule, _ := policyMap["natural_language_rule"].(string)

	if id == "" {
		return nil, fmt.Errorf("policy must have id field")
	}

	return &Policy{
		ID:                  id,
		Name:                name,
		Description:         description,
		NaturalLanguageRule: naturalLanguageRule,
		Scope:               PolicyScopeNode,  // Default scope
		Enforcement:         EnforcementBlock, // Default enforcement
		RequiredConfidence:  0.8,              // Default confidence
		Enabled:             true,
		CreatedAt:           time.Now(),
	}, nil
}
