package policies

import (
	"github.com/krzachariassen/ZTDP/internal/common"
)

// Import types from common package
type (
	NodeView  = common.NodeView
	EdgeView  = common.EdgeView
	GraphView = common.GraphView
	Mutation  = common.Mutation
)

// PolicyDeclaration represents a policy node in the graph
type PolicyDeclaration struct {
	ID          string
	Name        string
	Description string
	Type        string // check, approval, system
	Parameters  map[string]interface{}
}

// Policy interface for the policy engine, decoupled from graph internals.
//
// Deprecated: This interface is part of the legacy policy system.
// The system now uses graph-based policies directly.
type Policy interface {
	Name() string
	Validate(g GraphView, m Mutation) error
}

// PolicyRegistry allows dynamic registration and retrieval of policies.
// It also implements the common.PolicyValidator interface
//
// Deprecated: This type is part of the legacy policy system and will be removed in a future version.
// The system now uses graph-based policies directly.
type PolicyRegistry struct {
	policies map[string]Policy
}

// NewPolicyRegistry returns an empty PolicyRegistry.
//
// Deprecated: This function is part of the legacy policy system.
// The system now uses graph-based policies directly.
func NewPolicyRegistry() *PolicyRegistry {
	return &PolicyRegistry{policies: make(map[string]Policy)}
}

// NewPolicyRegistryWithDefaults returns a PolicyRegistry with all built-in policies registered.
//
// Deprecated: This function is part of the legacy policy system.
// The system now uses graph-based policies directly.
func NewPolicyRegistryWithDefaults() *PolicyRegistry {
	reg := NewPolicyRegistry()
	reg.Register(NewAllowedEnvironmentPolicy())
	reg.Register(NewMustDeployToDevBeforeProdPolicy())
	reg.Register(NewBlockDirectServiceToEnvEdgePolicy())
	return reg
}

// Deprecated: This method is part of the legacy policy system.
func (r *PolicyRegistry) Register(policy Policy) {
	r.policies[policy.Name()] = policy
}

// Deprecated: This method is part of the legacy policy system.
func (r *PolicyRegistry) Get(name string) (Policy, bool) {
	p, ok := r.policies[name]
	return p, ok
}

// Deprecated: This method is part of the legacy policy system.
func (r *PolicyRegistry) All() []Policy {
	result := make([]Policy, 0, len(r.policies))
	for _, p := range r.policies {
		result = append(result, p)
	}
	return result
}

// ValidateMutation implements the common.PolicyValidator interface
// It validates the mutation against all registered policies
//
// Deprecated: This method is part of the legacy policy system.
// The system now uses graph-based policies directly.
func (r *PolicyRegistry) ValidateMutation(view GraphView, mutation Mutation) error {
	// Run through all policies
	for _, policy := range r.All() {
		if err := policy.Validate(view, mutation); err != nil {
			return err
		}
	}
	return nil
}
