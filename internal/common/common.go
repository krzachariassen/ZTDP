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
