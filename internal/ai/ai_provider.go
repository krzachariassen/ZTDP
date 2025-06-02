package ai

import (
	"context"

	"github.com/krzachariassen/ZTDP/internal/graph"
)

// PlanningRequest contains all context needed for AI planning
type PlanningRequest struct {
	Intent        string                 `json:"intent"`         // Human-readable deployment intent
	ApplicationID string                 `json:"application_id"` // Target application
	EdgeTypes     []string               `json:"edge_types"`     // Edge types to consider (deploy, create, owns, etc.)
	Context       *PlanningContext       `json:"context"`        // Complete graph context
	Metadata      map[string]interface{} `json:"metadata"`       // Additional metadata
}

// PlanningContext provides complete graph state for AI reasoning
type PlanningContext struct {
	TargetNodes   []*graph.Node `json:"target_nodes"`   // Nodes in deployment scope
	RelatedNodes  []*graph.Node `json:"related_nodes"`  // Dependencies and related nodes
	Edges         []*graph.Edge `json:"edges"`          // All relevant edges
	PolicyContext interface{}   `json:"policy_context"` // Policy constraints (flexible type)
	EnvironmentID string        `json:"environment_id"` // Target environment
}

// PlanningResponse contains AI-generated deployment plan with reasoning
type PlanningResponse struct {
	Plan       *DeploymentPlan        `json:"plan"`       // Generated deployment plan
	Reasoning  string                 `json:"reasoning"`  // AI reasoning explanation
	Confidence float64                `json:"confidence"` // AI confidence score (0-1)
	Metadata   map[string]interface{} `json:"metadata"`   // Additional response metadata
}

// DeploymentPlan represents an AI-generated deployment plan
type DeploymentPlan struct {
	Steps      []*DeploymentStep      `json:"steps"`      // Ordered deployment steps
	Strategy   string                 `json:"strategy"`   // Deployment strategy (rolling, blue-green, etc.)
	Validation []string               `json:"validation"` // Validation checks to perform
	Rollback   *RollbackPlan          `json:"rollback"`   // Rollback plan if needed
	Metadata   map[string]interface{} `json:"metadata"`   // Additional plan metadata
}

// DeploymentStep represents a single step in the deployment plan
type DeploymentStep struct {
	ID           string                 `json:"id"`           // Unique step identifier
	Action       string                 `json:"action"`       // Action to perform (deploy, create, configure, etc.)
	Target       string                 `json:"target"`       // Target node/resource ID
	Dependencies []string               `json:"dependencies"` // Step dependencies
	Metadata     map[string]interface{} `json:"metadata"`     // Step-specific metadata
	Reasoning    string                 `json:"reasoning"`    // Why this step is needed
}

// RollbackPlan contains instructions for rolling back a deployment
type RollbackPlan struct {
	Steps    []*DeploymentStep      `json:"steps"`    // Rollback steps
	Triggers []string               `json:"triggers"` // Conditions that trigger rollback
	Metadata map[string]interface{} `json:"metadata"` // Rollback metadata
}

// AIProvider defines the interface for AI reasoning providers
// This follows the same pattern as GraphBackend for clean abstraction
type AIProvider interface {
	// GeneratePlan creates an intelligent deployment plan using AI reasoning
	GeneratePlan(ctx context.Context, request *PlanningRequest) (*PlanningResponse, error)

	// EvaluatePolicy uses AI to evaluate policy compliance and suggest actions
	EvaluatePolicy(ctx context.Context, policyContext interface{}) (*PolicyEvaluation, error)

	// OptimizePlan refines an existing plan based on additional context
	OptimizePlan(ctx context.Context, plan *DeploymentPlan, context *PlanningContext) (*PlanningResponse, error)

	// GetProviderInfo returns information about the AI provider
	GetProviderInfo() *ProviderInfo

	// Close cleans up provider resources
	Close() error
}

// PolicyEvaluation contains AI-driven policy evaluation results
type PolicyEvaluation struct {
	Compliant   bool                   `json:"compliant"`   // Whether policies are satisfied
	Violations  []string               `json:"violations"`  // Policy violations found
	Suggestions []string               `json:"suggestions"` // AI suggestions for compliance
	Reasoning   string                 `json:"reasoning"`   // AI reasoning for evaluation
	Confidence  float64                `json:"confidence"`  // Confidence in evaluation
	Metadata    map[string]interface{} `json:"metadata"`    // Additional evaluation metadata
}

// ProviderInfo contains metadata about an AI provider
type ProviderInfo struct {
	Name         string                 `json:"name"`         // Provider name (e.g., "openai-gpt4")
	Version      string                 `json:"version"`      // Provider version
	Capabilities []string               `json:"capabilities"` // Supported capabilities
	Metadata     map[string]interface{} `json:"metadata"`     // Provider-specific metadata
}
