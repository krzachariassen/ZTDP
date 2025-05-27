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
