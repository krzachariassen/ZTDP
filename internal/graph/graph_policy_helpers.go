package graph

// PolicyHelpers provides additional methods specifically for policy operations on the graph.
// This file extends the core Graph model with policy-specific functionality.

// FindPoliciesRequiredForTransition finds all policy nodes that are required for a specific
// transition (edge) between two nodes.
func (g *Graph) FindPoliciesRequiredForTransition(fromID, toID, edgeType string) ([]*Node, error) {
	// Build an identifier for the transition (process node)
	transitionID := fromID + "-" + edgeType + "-" + toID

	// Check if we have a process node for this transition
	processNode, err := g.GetNode(transitionID)
	if err != nil {
		// Process node doesn't exist yet, which means no policies are attached
		return []*Node{}, nil
	}

	// Find all policy nodes that this process requires
	requiredPolicies := []*Node{}
	edges, ok := g.Edges[processNode.ID]
	if !ok {
		return requiredPolicies, nil
	}

	for _, edge := range edges {
		if edge.Type == EdgeTypeRequires {
			policyNode, err := g.GetNode(edge.To)
			if err != nil {
				return nil, err
			}
			if policyNode.Kind == KindPolicy {
				requiredPolicies = append(requiredPolicies, policyNode)
			}
		}
	}

	return requiredPolicies, nil
}

// IsTransitionAllowed checks if a transition (adding an edge) is allowed based on policies.
// Returns nil if allowed, otherwise returns an error explaining why it's not allowed.
func (g *Graph) IsTransitionAllowed(fromID, toID, edgeType string) error {
	// Find required policies
	requiredPolicies, err := g.FindPoliciesRequiredForTransition(fromID, toID, edgeType)
	if err != nil {
		return err
	}

	// If no policies are required, transition is allowed
	if len(requiredPolicies) == 0 {
		return nil
	}

	// Check each policy for satisfaction
	for _, policy := range requiredPolicies {
		satisfied, err := g.IsPolicySatisfied(policy.ID)
		if err != nil {
			return err
		}
		if !satisfied {
			return &PolicyNotSatisfiedError{PolicyID: policy.ID, PolicyName: policy.Metadata["name"].(string)}
		}
	}

	return nil
}

// IsPolicySatisfied checks if a policy has been satisfied by a check.
func (g *Graph) IsPolicySatisfied(policyID string) (bool, error) {
	// Get all edges pointing to this policy
	for from, edges := range g.Edges {
		for _, edge := range edges {
			if edge.To == policyID && edge.Type == EdgeTypeSatisfies {
				// Found a satisfies edge, check if it's from a valid check node
				checkNode, err := g.GetNode(from)
				if err != nil {
					return false, err
				}

				// Check if the check node has succeeded
				if checkNode.Kind == KindCheck {
					if status, ok := checkNode.Metadata["status"]; ok {
						return status == CheckStatusSucceeded, nil
					}
				}
			}
		}
	}

	// No satisfying check found
	return false, nil
}

// AttachPolicyToTransition creates a policy requirement for a transition.
// If the process node doesn't exist, it creates one.
func (g *Graph) AttachPolicyToTransition(fromID, toID, edgeType, policyID string) error {
	// Build an identifier for the transition (process node)
	transitionID := fromID + "-" + edgeType + "-" + toID

	// Check if process node exists, create if not
	processNode, err := g.GetNode(transitionID)
	if err != nil {
		// Create process node
		processNode = &Node{
			ID:   transitionID,
			Kind: "process",
			Metadata: map[string]interface{}{
				"name":        "Process " + transitionID,
				"description": "Process node for transition from " + fromID + " to " + toID + " via " + edgeType,
				"fromID":      fromID,
				"toID":        toID,
				"edgeType":    edgeType,
			},
			Spec: map[string]interface{}{},
		}

		err = g.AddNode(processNode)
		if err != nil {
			return err
		}
	}

	// Check if edge already exists
	if edges, ok := g.Edges[processNode.ID]; ok {
		for _, edge := range edges {
			if edge.To == policyID && edge.Type == EdgeTypeRequires {
				return nil // Already exists
			}
		}
	}

	// Add new edge
	g.AddEdge(processNode.ID, policyID, EdgeTypeRequires)

	return nil
}

// MarkPolicySatisfiedByCheck creates a satisfies edge from a check to a policy.
func (g *Graph) MarkPolicySatisfiedByCheck(checkID, policyID string) error {
	// Verify both nodes exist
	_, err := g.GetNode(checkID)
	if err != nil {
		return err
	}

	_, err = g.GetNode(policyID)
	if err != nil {
		return err
	}

	// Check if edge already exists
	if edges, ok := g.Edges[checkID]; ok {
		for _, edge := range edges {
			if edge.To == policyID && edge.Type == EdgeTypeSatisfies {
				return nil // Already exists
			}
		}
	}

	// Add new edge
	g.AddEdge(checkID, policyID, EdgeTypeSatisfies)

	return nil
}

// GetPoliciesSatisfiedByCheck returns a list of policy IDs that are satisfied by a specific check
func (g *Graph) GetPoliciesSatisfiedByCheck(checkID string) []string {
	policies := []string{}

	// Look through all edges from the check node
	if edges, ok := g.Edges[checkID]; ok {
		for _, edge := range edges {
			if edge.Type == EdgeTypeSatisfies {
				// Found a satisfies edge to a policy
				policyNode, err := g.GetNode(edge.To)
				if err == nil && policyNode.Kind == KindPolicy {
					policies = append(policies, policyNode.ID)
				}
			}
		}
	}

	return policies
}

// PolicyNotSatisfiedError is returned when a required policy is not satisfied.
type PolicyNotSatisfiedError struct {
	PolicyID   string
	PolicyName string
}

func (e *PolicyNotSatisfiedError) Error() string {
	return "Policy not satisfied: " + e.PolicyName + " (ID: " + e.PolicyID + ")"
}
