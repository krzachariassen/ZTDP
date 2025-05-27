package policies

import (
	"github.com/krzachariassen/ZTDP/internal/common"
)

// GraphBasedPolicyValidator adapts the graph-based policy system to implement
// the common.PolicyValidator interface
type GraphBasedPolicyValidator struct{}

// NewGraphBasedPolicyValidator creates a new validator that uses graph-based policies
func NewGraphBasedPolicyValidator() *GraphBasedPolicyValidator {
	return &GraphBasedPolicyValidator{}
}

// ValidateMutation implements common.PolicyValidator
// This is a compatibility layer for the graph-based policy system
// It uses the graph's own policy validation, so it effectively does nothing here
// as the actual validation will happen when edges are added to the graph
func (v *GraphBasedPolicyValidator) ValidateMutation(view common.GraphView, mutation common.Mutation) error {
	// Since graph-based policies are enforced directly when edges are added to the graph,
	// this function is now just a passthrough for backward compatibility
	if mutation.Type == "add_edge" && mutation.Edge != nil {
		// Edge addition is validated by the graph itself via IsTransitionAllowed
		// No need to do anything here
		return nil
	}

	// For other mutation types, we don't have validation yet in the graph-based system
	// So just allow them through
	return nil
}
