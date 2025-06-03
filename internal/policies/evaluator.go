package policies

import (
	"context"
	"fmt"
	"time"

	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// Evaluator provides AI-native policy evaluation
// This integrates AI capabilities into the existing policy evaluation system
type Evaluator struct {
	provider    ai.AIProvider
	graph       *graph.GlobalGraph
	environment string
	logger      *logging.Logger
}

// NewEvaluator creates a new AI-first policy evaluator
func NewEvaluator(provider ai.AIProvider, graph *graph.GlobalGraph, environment string) *Evaluator {
	return &Evaluator{
		provider:    provider,
		graph:       graph,
		environment: environment,
		logger:      logging.GetLogger().ForComponent("policy-evaluator"),
	}
}

// EvaluateDeploymentPolicies provides intelligent policy evaluation for deployments
// Uses AI to understand complex policy interactions and compliance requirements
func (e *Evaluator) EvaluateDeploymentPolicies(ctx context.Context, applicationID string, environmentID string) (*ai.PolicyEvaluation, error) {
	if e.provider == nil {
		return nil, fmt.Errorf("AI provider is required for policy evaluation - this is an AI-native system")
	}

	e.logger.Info("🔍 Evaluating deployment policies with AI: %s -> %s", applicationID, environmentID)

	// Extract comprehensive policy context
	policyContext, err := e.extractPolicyContext(applicationID, environmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to extract policy context: %w", err)
	}

	// Use AI to evaluate policies intelligently
	evaluation, err := e.provider.EvaluatePolicy(ctx, policyContext)
	if err != nil {
		return nil, fmt.Errorf("AI policy evaluation failed: %w", err)
	}

	e.logger.Info("✅ AI policy evaluation completed (compliant: %t, violations: %d)",
		evaluation.Compliant, len(evaluation.Violations))

	return evaluation, nil
}

// ValidateTransition uses AI to validate complex policy transitions
// This extends the existing graph-based validation with intelligent reasoning
func (e *Evaluator) ValidateTransition(ctx context.Context, fromID, toID, edgeType, user string) (*ai.PolicyEvaluation, error) {
	if e.provider == nil {
		return nil, fmt.Errorf("AI provider is required for transition validation - this is an AI-native system")
	}

	e.logger.Info("🔍 Validating transition with AI: %s -> %s (%s)", fromID, toID, edgeType)

	// Build transition context for AI evaluation
	transitionContext := e.buildTransitionContext(fromID, toID, edgeType, user)

	// Use AI to evaluate the transition
	evaluation, err := e.provider.EvaluatePolicy(ctx, transitionContext)
	if err != nil {
		return nil, fmt.Errorf("AI transition validation failed: %w", err)
	}

	e.logger.Info("✅ AI transition validation completed (allowed: %t)", evaluation.Compliant)
	return evaluation, nil
}

// generateSimplePolicyEvaluation provides deterministic fallback evaluation
func (e *Evaluator) generateSimplePolicyEvaluation(applicationID, environmentID string) *ai.PolicyEvaluation {
	e.logger.Info("🔄 Generating simple policy evaluation for %s -> %s", applicationID, environmentID)

	// Basic heuristic: production environments have stricter policies
	compliant := true
	violations := []ai.PolicyViolation{}

	if environmentID == "production" {
		// In production, require additional checks
		violations = append(violations, ai.PolicyViolation{
			PolicyID:    "prod-approval-required",
			Severity:    "medium",
			Description: "Production deployments require approval",
			Suggestion:  "Obtain approval from operations team",
		})
		compliant = false
	}

	return &ai.PolicyEvaluation{
		Compliant:       compliant,
		Violations:      violations,
		Recommendations: []string{"Follow standard deployment procedures", "Monitor deployment progress"},
		Confidence:      0.7, // Lower confidence for simple evaluation
		Reasoning:       "Basic heuristic-based policy evaluation",
		Metadata: map[string]interface{}{
			"evaluation_type": "deterministic",
			"timestamp":       time.Now(),
			"environment":     environmentID,
		},
	}
}

// generateSimpleTransitionEvaluation provides deterministic transition validation
func (e *Evaluator) generateSimpleTransitionEvaluation(fromID, toID, edgeType, user string) *ai.PolicyEvaluation {
	e.logger.Info("🔄 Generating simple transition evaluation: %s -> %s", fromID, toID)

	// Basic validation rules
	compliant := true
	violations := []ai.PolicyViolation{}

	// Simple rule: deployments require proper permissions
	if edgeType == "deploy" && user != "admin" {
		violations = append(violations, ai.PolicyViolation{
			PolicyID:    "deploy-permission-required",
			Severity:    "high",
			Description: "Deployment requires administrative permissions",
			Suggestion:  "Request deployment permissions or use administrative account",
		})
		compliant = false
	}

	return &ai.PolicyEvaluation{
		Compliant:       compliant,
		Violations:      violations,
		Recommendations: []string{"Verify user permissions", "Follow deployment protocols"},
		Confidence:      0.6,
		Reasoning:       "Simple rule-based transition validation",
		Metadata: map[string]interface{}{
			"evaluation_type": "deterministic",
			"timestamp":       time.Now(),
			"transition":      fmt.Sprintf("%s->%s", fromID, toID),
		},
	}
}

// extractPolicyContext builds comprehensive context for AI policy evaluation
func (e *Evaluator) extractPolicyContext(applicationID, environmentID string) (map[string]interface{}, error) {
	graph, err := e.graph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	context := map[string]interface{}{
		"application_id":   applicationID,
		"environment_id":   environmentID,
		"graph_state":      e.extractGraphState(graph, applicationID),
		"policies":         e.extractRelevantPolicies(graph, environmentID),
		"constraints":      e.extractEnvironmentConstraints(environmentID),
		"current_time":     time.Now(),
		"evaluation_scope": "deployment",
	}

	return context, nil
}

// buildTransitionContext creates context for transition validation
func (e *Evaluator) buildTransitionContext(fromID, toID, edgeType, user string) map[string]interface{} {
	return map[string]interface{}{
		"from_node":       fromID,
		"to_node":         toID,
		"edge_type":       edgeType,
		"user":            user,
		"timestamp":       time.Now(),
		"environment":     e.environment,
		"transition_type": "graph_edge_creation",
	}
}

// Helper methods for context extraction
func (e *Evaluator) extractGraphState(graph *graph.Graph, applicationID string) map[string]interface{} {
	// Extract relevant graph state for the application
	nodeCount := len(graph.Nodes)
	edgeCount := 0
	for _, edges := range graph.Edges {
		edgeCount += len(edges)
	}

	return map[string]interface{}{
		"total_nodes":       nodeCount,
		"total_edges":       edgeCount,
		"application_nodes": e.getApplicationNodes(graph, applicationID),
		"dependencies":      e.getApplicationDependencies(graph, applicationID),
	}
}

func (e *Evaluator) extractRelevantPolicies(graph *graph.Graph, environmentID string) []map[string]interface{} {
	policies := []map[string]interface{}{}

	for _, node := range graph.Nodes {
		if node.Kind == "policy" {
			if env, exists := node.Metadata["environment"]; exists && env == environmentID {
				policies = append(policies, map[string]interface{}{
					"policy_id":   node.ID,
					"policy_type": node.Metadata["type"],
					"severity":    node.Metadata["severity"],
					"description": node.Metadata["description"],
				})
			}
		}
	}

	return policies
}

func (e *Evaluator) extractEnvironmentConstraints(environmentID string) map[string]interface{} {
	constraints := map[string]interface{}{
		"environment": environmentID,
	}

	// Add environment-specific constraints
	switch environmentID {
	case "production":
		constraints["approval_required"] = true
		constraints["rollback_required"] = true
		constraints["monitoring_required"] = true
	case "staging":
		constraints["testing_required"] = true
		constraints["approval_required"] = false
	default:
		constraints["approval_required"] = false
	}

	return constraints
}

func (e *Evaluator) getApplicationNodes(graph *graph.Graph, applicationID string) []string {
	nodes := []string{}

	for _, node := range graph.Nodes {
		if appID, exists := node.Metadata["application_id"]; exists && appID == applicationID {
			nodes = append(nodes, node.ID)
		}
	}

	return nodes
}

func (e *Evaluator) getApplicationDependencies(graph *graph.Graph, applicationID string) []string {
	dependencies := []string{}

	// Find dependencies by looking at edges
	if edges, exists := graph.Edges[applicationID]; exists {
		for _, edge := range edges {
			if edge.Type == "depends_on" {
				dependencies = append(dependencies, edge.To)
			}
		}
	}

	return dependencies
}
