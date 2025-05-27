package common

// PolicyValidator defines an interface for validating graph mutations.
type PolicyValidator interface {
	// ValidateMutation checks if a mutation is allowed according to policies.
	ValidateMutation(view GraphView, mutation Mutation) error
}

// PolicyDeclaration represents a policy node in the graph
type PolicyDeclaration struct {
	ID          string
	Name        string
	Description string
	Type        string // check, approval, system
	Parameters  map[string]interface{}
}
