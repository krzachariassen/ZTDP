package types

import (
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

// DeploymentStep represents a single step in deployment plan
type DeploymentStep struct {
	ID           string                 `json:"id"`           // Unique step identifier
	Action       string                 `json:"action"`       // Action to perform (deploy, create, configure, etc.)
	Target       string                 `json:"target"`       // Target node/resource identifier
	Dependencies []string               `json:"dependencies"` // Step IDs that must complete first
	Metadata     map[string]interface{} `json:"metadata"`     // Step-specific metadata
	Reasoning    string                 `json:"reasoning"`    // AI reasoning for this step
}

// RollbackPlan defines rollback strategy for deployment plan
type RollbackPlan struct {
	Steps    []*DeploymentStep      `json:"steps"`    // Ordered rollback steps
	Triggers []string               `json:"triggers"` // Conditions that trigger rollback
	Metadata map[string]interface{} `json:"metadata"` // Rollback metadata
}

// PolicyEvaluation represents AI policy compliance evaluation
type PolicyEvaluation struct {
	Compliant   bool                   `json:"compliant"`   // Whether deployment is policy compliant
	Violations  []PolicyViolation      `json:"violations"`  // List of policy violations
	Suggestions []string               `json:"suggestions"` // AI suggestions for compliance
	Confidence  float64                `json:"confidence"`  // AI confidence in evaluation (0-1)
	Reasoning   string                 `json:"reasoning"`   // AI reasoning for evaluation
	Metadata    map[string]interface{} `json:"metadata"`    // Additional evaluation metadata
}

// PolicyViolation represents a specific policy violation
type PolicyViolation struct {
	PolicyID    string                 `json:"policy_id"`   // Violated policy identifier
	Severity    string                 `json:"severity"`    // Violation severity (low, medium, high, critical)
	Description string                 `json:"description"` // Human-readable violation description
	Suggestion  string                 `json:"suggestion"`  // AI suggestion to fix violation
	Metadata    map[string]interface{} `json:"metadata"`    // Additional violation metadata
}
