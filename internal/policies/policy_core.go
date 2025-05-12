package policies

// NodeView is a minimal view of a node for policy checks.
type NodeView struct {
	ID       string
	Kind     string
	Metadata map[string]interface{}
	Spec     map[string]interface{}
}

// EdgeView is a minimal view of an edge for policy checks.
type EdgeView struct {
	From     string
	To       string
	Type     string
	Metadata map[string]interface{}
}

// GraphView provides only the data needed for policy checks.
type GraphView struct {
	Nodes map[string]NodeView
	Edges map[string][]EdgeView // key is from node ID
}

// Mutation describes a change to the graph for policy validation.
type Mutation struct {
	Type    string                 // e.g. "add_node", "add_edge"
	Node    *NodeView              // Node being added/updated/deleted (if applicable)
	Edge    *EdgeView              // Edge being added/updated/deleted (if applicable)
	User    string                 // Optionally, user or actor info
	Context map[string]interface{} // Additional context for policy checks
}

// Policy interface for the policy engine, decoupled from graph internals.
type Policy interface {
	Name() string
	Validate(g GraphView, m Mutation) error
}

// PolicyRegistry allows dynamic registration and retrieval of policies.
type PolicyRegistry struct {
	policies map[string]Policy
}

// NewPolicyRegistry returns an empty PolicyRegistry.
func NewPolicyRegistry() *PolicyRegistry {
	return &PolicyRegistry{policies: make(map[string]Policy)}
}

// NewPolicyRegistryWithDefaults returns a PolicyRegistry with all built-in policies registered.
func NewPolicyRegistryWithDefaults() *PolicyRegistry {
	reg := NewPolicyRegistry()
	reg.Register(NewAllowedEnvironmentPolicy())
	reg.Register(NewMustDeployToDevBeforeProdPolicy())
	reg.Register(NewBlockDirectServiceToEnvEdgePolicy())
	return reg
}

func (r *PolicyRegistry) Register(policy Policy) {
	r.policies[policy.Name()] = policy
}

func (r *PolicyRegistry) Get(name string) (Policy, bool) {
	p, ok := r.policies[name]
	return p, ok
}

func (r *PolicyRegistry) All() []Policy {
	result := make([]Policy, 0, len(r.policies))
	for _, p := range r.policies {
		result = append(result, p)
	}
	return result
}
