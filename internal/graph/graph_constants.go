package graph

import (
	"github.com/krzachariassen/ZTDP/internal/common"
)

// Import and re-export constants from common package
const (
	// Node kinds
	KindApplication      = common.KindApplication
	KindService          = common.KindService
	KindServiceVersion   = common.KindServiceVersion
	KindEnvironment      = common.KindEnvironment
	KindResourceRegister = common.KindResourceRegister
	KindResourceType     = common.KindResourceType
	KindResource         = common.KindResource
	KindPolicy           = common.KindPolicy
	KindCheck            = common.KindCheck
	KindProcess          = common.KindProcess

	// Edge types
	EdgeTypeOwns       = common.EdgeTypeOwns
	EdgeTypeHasVersion = common.EdgeTypeHasVersion
	EdgeTypeDeploy     = common.EdgeTypeDeploy
	EdgeTypeCreate     = "create"
	EdgeTypeUses       = common.EdgeTypeUses
	EdgeTypeInstanceOf = common.EdgeTypeInstanceOf
	EdgeTypeRequires   = common.EdgeTypeRequires
	EdgeTypeSatisfies  = common.EdgeTypeSatisfies

	// Policy types
	PolicyTypeCheck    = common.PolicyTypeCheck
	PolicyTypeApproval = common.PolicyTypeApproval
	PolicyTypeSystem   = common.PolicyTypeSystem

	// Check statuses
	CheckStatusPending   = common.CheckStatusPending
	CheckStatusRunning   = common.CheckStatusRunning
	CheckStatusSucceeded = common.CheckStatusSucceeded
	CheckStatusFailed    = common.CheckStatusFailed
)

// Allowed edge types for the platform
var AllowedEdgeTypes = map[string]struct{}{
	EdgeTypeOwns:       {},
	EdgeTypeHasVersion: {},
	EdgeTypeDeploy:     {},
	EdgeTypeCreate:     {},
	EdgeTypeUses:       {},
	EdgeTypeInstanceOf: {},
	EdgeTypeRequires:   {},
	EdgeTypeSatisfies:  {},
	"allowed_in":       {}, // Policy edge type for environment access
	// Add more as needed
}

// IsValidEdgeType returns true if the edge type is allowed
func IsValidEdgeType(edgeType string) bool {
	_, ok := AllowedEdgeTypes[edgeType]
	return ok
}
