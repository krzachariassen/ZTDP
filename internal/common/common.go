package common

// Constants for graph node kinds
const (
	KindApplication      = "application"
	KindService          = "service"
	KindServiceVersion   = "service_version"
	KindEnvironment      = "environment"
	KindResourceRegister = "resource_register"
	KindResourceType     = "resource_type"
	KindResource         = "resource"
	KindPolicy           = "policy"
	KindCheck            = "check"
	KindProcess          = "process"
)

// Constants for graph edge types
const (
	EdgeTypeOwns       = "owns"
	EdgeTypeHasVersion = "has_version"
	EdgeTypeDeploy     = "deploy"
	EdgeTypeUses       = "uses"
	EdgeTypeInstanceOf = "instance_of"
	EdgeTypeRequires   = "requires"
	EdgeTypeSatisfies  = "satisfies"
)

// Constants for policy types
const (
	PolicyTypeCheck    = "check"
	PolicyTypeApproval = "approval"
	PolicyTypeSystem   = "system"
)

// Constants for check status
const (
	CheckStatusPending   = "pending"
	CheckStatusRunning   = "running"
	CheckStatusSucceeded = "succeeded"
	CheckStatusFailed    = "failed"
)

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
