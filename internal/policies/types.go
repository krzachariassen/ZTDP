package policies

import (
	"context"
	"errors"
	"time"

	"github.com/krzachariassen/ZTDP/internal/graph"
)

// Core errors
var (
	ErrPolicyNotFound     = errors.New("policy not found")
	ErrAIEvaluationFailed = errors.New("AI evaluation failed")
	ErrInvalidPolicy      = errors.New("invalid policy")
)

// PolicyScope defines where a policy applies
type PolicyScope string

const (
	PolicyScopeNode          PolicyScope = "node"
	PolicyScopeEdge          PolicyScope = "edge"
	PolicyScopeGraph         PolicyScope = "graph"
	PolicyScopeOperation     PolicyScope = "operation"
	PolicyScopeConfiguration PolicyScope = "configuration"
)

// PolicyStatus represents the outcome of policy evaluation
type PolicyStatus string

const (
	PolicyStatusAllowed         PolicyStatus = "allowed"
	PolicyStatusBlocked         PolicyStatus = "blocked"
	PolicyStatusPendingApproval PolicyStatus = "pending_approval"
	PolicyStatusConditional     PolicyStatus = "conditional"
	PolicyStatusWarning         PolicyStatus = "warning"
	PolicyStatusNotApplicable   PolicyStatus = "not_applicable"
)

// PolicyEnforcement defines how policies are enforced
type PolicyEnforcement string

const (
	EnforcementBlock   PolicyEnforcement = "block"
	EnforcementWarn    PolicyEnforcement = "warn"
	EnforcementApprove PolicyEnforcement = "approve"
	EnforcementAudit   PolicyEnforcement = "audit"
	EnforcementMonitor PolicyEnforcement = "monitor"
)

// Policy represents an AI-native policy definition
type Policy struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	// Scope and applicability
	Scope     PolicyScope `json:"scope"`
	NodeTypes []string    `json:"node_types,omitempty"`
	EdgeTypes []string    `json:"edge_types,omitempty"`

	// AI-native rule definition
	NaturalLanguageRule string `json:"natural_language_rule"`
	AIPromptTemplate    string `json:"ai_prompt_template,omitempty"`

	// Enforcement configuration
	Enforcement PolicyEnforcement `json:"enforcement"`
	Priority    int               `json:"priority"`

	// AI configuration
	RequiredConfidence float64 `json:"required_confidence"`

	// Context and conditions
	Conditions map[string]interface{} `json:"conditions,omitempty"`
	Exceptions []string               `json:"exceptions,omitempty"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by"`
	Enabled   bool      `json:"enabled"`
}

// PolicyResult contains the outcome of policy evaluation
type PolicyResult struct {
	// What was evaluated
	NodeID       string `json:"node_id,omitempty"`
	NodeKind     string `json:"node_kind,omitempty"`
	EdgeFrom     string `json:"edge_from,omitempty"`
	EdgeTo       string `json:"edge_to,omitempty"`
	Relationship string `json:"relationship,omitempty"`
	GraphScope   bool   `json:"graph_scope,omitempty"`
	Environment  string `json:"environment"`

	// Evaluation results
	OverallStatus PolicyStatus                 `json:"overall_status"`
	Status        PolicyStatus                 `json:"status"` // For single policy evaluations
	Evaluations   map[string]*PolicyEvaluation `json:"evaluations"`

	// AI evaluation fields (for direct access in tests)
	Confidence  float64 `json:"confidence,omitempty"`
	AIReasoning string  `json:"ai_reasoning,omitempty"`
	Reason      string  `json:"reason,omitempty"`

	// Metadata
	EvaluatedAt time.Time `json:"evaluated_at"`
	EvaluatedBy string    `json:"evaluated_by"`
}

// PolicyEvaluation contains the result of a single policy check
type PolicyEvaluation struct {
	PolicyID    string       `json:"policy_id"`
	Status      PolicyStatus `json:"status"`
	Reason      string       `json:"reason"`
	Confidence  float64      `json:"confidence"`
	AIReasoning string       `json:"ai_reasoning,omitempty"`

	// Actions and recommendations
	RequiredActions []PolicyAction `json:"required_actions,omitempty"`
	Recommendations []string       `json:"recommendations,omitempty"`

	// Approval workflow
	RequiresApproval bool   `json:"requires_approval"`
	ApprovalWorkflow string `json:"approval_workflow,omitempty"`

	EvaluatedAt time.Time `json:"evaluated_at"`
}

// PolicyAction represents an action required by a policy
type PolicyAction struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Required    bool                   `json:"required"`
}

// AIPrompt represents a structured prompt for AI evaluation
type AIPrompt struct {
	System string `json:"system"`
	User   string `json:"user"`
}

// PolicyStore interface for policy persistence
type PolicyStore interface {
	Store(policy *Policy) error
	Get(id string) (*Policy, error)
	GetPoliciesForNodeType(nodeType string) ([]*Policy, error)
	GetPoliciesForEdgeType(edgeType string) ([]*Policy, error)
	GetGraphPolicies() ([]*Policy, error)
}

// TestableService interface for testing - to be implemented by existing Service
type TestableService interface {
	// Node policy evaluation
	EvaluateNodePolicy(ctx context.Context, env string, node *graph.Node, policy *Policy) (*PolicyResult, error)
	EvaluateNode(ctx context.Context, env string, node *graph.Node) (*PolicyResult, error)

	// Edge policy evaluation
	EvaluateEdgePolicy(ctx context.Context, env string, edge *graph.Edge, policy *Policy) (*PolicyResult, error)
	EvaluateEdge(ctx context.Context, env string, edge *graph.Edge) (*PolicyResult, error)

	// Graph policy evaluation
	EvaluateGraphPolicy(ctx context.Context, env string, graph *graph.Graph, policy *Policy) (*PolicyResult, error)
	EvaluateGraph(ctx context.Context, env string, graph *graph.Graph) (*PolicyResult, error)

	// AI integration
	BuildNodePolicyPrompt(ctx context.Context, node *graph.Node, policy *Policy) (*AIPrompt, error)
	BuildEdgePolicyPrompt(ctx context.Context, edge *graph.Edge, policy *Policy) (*AIPrompt, error)
	BuildGraphPolicyPrompt(ctx context.Context, graph *graph.Graph, policy *Policy) (*AIPrompt, error)
	ParseAIResponse(response string) (*PolicyEvaluation, error)
}

// EventBus interface for emitting policy events
type EventBus interface {
	Emit(eventType string, data map[string]interface{}) error
}
