package policies

import (
	"fmt"

	"github.com/krzachariassen/ZTDP/internal/common"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

// GraphStoreInterface defines the subset of functionality needed from GraphStore.
type GraphStoreInterface interface {
	GetGraph(env string) (*graph.Graph, error)
	AddNode(env string, node *graph.Node) error
	GetNode(env string, id string) (*graph.Node, error)
	AddEdge(env, fromID, toID, relType string) error
}

// PolicyEvaluator provides graph-based policy evaluation for the control plane.
// It evaluates policies represented as nodes in the graph, with associated checks
// that can satisfy those policies. The policy system is based on a directed graph
// model where policies are attached to transitions (edges) between nodes.
type PolicyEvaluator struct {
	graphStore   GraphStoreInterface
	environment  string
	eventService *events.PolicyEventService
}

// NewPolicyEvaluator creates a new policy evaluator with the given graph store.
func NewPolicyEvaluator(graphStore GraphStoreInterface, environment string) *PolicyEvaluator {
	return &PolicyEvaluator{
		graphStore:  graphStore,
		environment: environment,
		// Event service will be set by SetEventService
	}
}

// SetEventService sets the event service for policy-related events
func (e *PolicyEvaluator) SetEventService(eventService *events.PolicyEventService) {
	e.eventService = eventService
}

// ValidateTransition checks if a transition (adding an edge) is allowed based on attached policies.
// This method uses the graph-based policy model.
func (e *PolicyEvaluator) ValidateTransition(fromID, toID, edgeType, user string) error {
	// Emit transition attempt event if event service is available
	if e.eventService != nil {
		if err := e.eventService.EmitTransitionAttempt(fromID, toID, edgeType, user); err != nil {
			// Log but continue - event failure shouldn't block the transition
			fmt.Printf("Warning: Failed to emit transition attempt event: %v\n", err)
		}
	}

	// Check if the transition is allowed based on the graph-based policy model
	g, err := e.graphStore.GetGraph(e.environment)
	if err != nil {
		return fmt.Errorf("failed to get graph: %w", err)
	}

	err = g.IsTransitionAllowed(fromID, toID, edgeType)

	// Emit result event
	if e.eventService != nil {
		if err != nil {
			e.eventService.EmitTransitionResult(fromID, toID, edgeType, user, false, err.Error())
		} else {
			e.eventService.EmitTransitionResult(fromID, toID, edgeType, user, true, "All policies satisfied")
		}
	}

	return err
}

// CreatePolicyNode creates a new policy node in the graph.
func (e *PolicyEvaluator) CreatePolicyNode(name, description, policyType string, parameters map[string]interface{}) (*graph.Node, error) {
	// Generate a unique ID for the policy
	policyID := fmt.Sprintf("policy-%s", name)

	// Create policy node
	policyNode := &graph.Node{
		ID:   policyID,
		Kind: graph.KindPolicy,
		Metadata: map[string]interface{}{
			"name":        name,
			"description": description,
			"type":        policyType,
			"status":      "active",
		},
		Spec: parameters,
	}

	// Add policy to graph
	if err := e.graphStore.AddNode(e.environment, policyNode); err != nil {
		return nil, err
	}

	return policyNode, nil
}

// CreateCheckNode creates a check node that can potentially satisfy a policy.
func (e *PolicyEvaluator) CreateCheckNode(checkID, name, checkType string, parameters map[string]interface{}) (*graph.Node, error) {
	// Create check node
	checkNode := &graph.Node{
		ID:   checkID,
		Kind: graph.KindCheck,
		Metadata: map[string]interface{}{
			"name":   name,
			"type":   checkType,
			"status": graph.CheckStatusPending,
		},
		Spec: parameters,
	}

	// Add check to graph
	err := e.graphStore.AddNode(e.environment, checkNode)
	if err != nil {
		return nil, err
	}

	return checkNode, nil
}

// UpdateCheckStatus updates the status of a check node.
func (e *PolicyEvaluator) UpdateCheckStatus(checkID, status string, results map[string]interface{}) error {
	// Get the check node
	checkNode, err := e.graphStore.GetNode(e.environment, checkID)
	if err != nil {
		return err
	}

	oldStatus := ""
	if statusVal, ok := checkNode.Metadata["status"]; ok {
		oldStatus = fmt.Sprintf("%v", statusVal)
	}

	// Update the status
	checkNode.Metadata["status"] = status

	// Add results if provided
	if results != nil {
		checkNode.Metadata["results"] = results
	}

	// Find associated policies (what this check is satisfying)
	g, err := e.graphStore.GetGraph(e.environment)
	if err == nil && e.eventService != nil {
		// Look for check-satisfies->policy edges
		policies := g.GetPoliciesSatisfiedByCheck(checkID)

		for _, policyID := range policies {
			// Create details representing the check update
			details := map[string]interface{}{
				"type":       "update_check",
				"check_id":   checkID,
				"kind":       common.KindCheck,
				"metadata":   checkNode.Metadata,
				"spec":       checkNode.Spec,
				"old_status": oldStatus,
				"new_status": status,
				"policy_id":  policyID,
			}

			// Emit event for check update that affects a policy
			e.eventService.EmitPolicyCheckResult(
				policyID,
				status == common.CheckStatusSucceeded,
				fmt.Sprintf("Check %s status changed to %s", checkID, status),
				details,
			)
		}
	}

	// Since we're modifying an existing node, we need to update it in the graph store
	// In a real system, you might need a specific update method depending on your backend
	return e.graphStore.AddNode(e.environment, checkNode)
}

// SatisfyPolicy creates a 'satisfies' relationship from a check to a policy.
func (e *PolicyEvaluator) SatisfyPolicy(checkID, policyID string) error {
	g, err := e.graphStore.GetGraph(e.environment)
	if err != nil {
		return fmt.Errorf("failed to get graph: %w", err)
	}

	// Get the check and policy nodes for event context
	checkNode, err := e.graphStore.GetNode(e.environment, checkID)
	if err != nil {
		return fmt.Errorf("check node not found: %w", err)
	}

	policyNode, err := e.graphStore.GetNode(e.environment, policyID)
	if err != nil {
		return fmt.Errorf("policy node not found: %w", err)
	}

	// Emit event for the satisfaction relationship if event service is available
	if e.eventService != nil {
		details := map[string]interface{}{
			"type":      "satisfy_policy",
			"from":      checkID,
			"to":        policyID,
			"edge_type": common.EdgeTypeSatisfies,
		}
		context := map[string]interface{}{
			"check_type":   checkNode.Metadata["type"],
			"check_status": checkNode.Metadata["status"],
			"policy_type":  policyNode.Metadata["type"],
			"policy_name":  policyNode.Metadata["name"],
			"check_id":     checkID,
			"check_name":   checkNode.Metadata["name"],
		}

		e.eventService.EmitPolicyCheck(
			policyID,
			details,
			context,
		)
	}

	// Use the graph's helper method to create the satisfies relationship
	err = g.MarkPolicySatisfiedByCheck(checkID, policyID)
	if err != nil {
		return err
	}

	return nil
}

// Note: The convertToGraphView function has been removed as part of the policy system simplification.
// The system now only uses the graph-based policy model directly, without conversion.
