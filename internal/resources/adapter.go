package resources

import (
	"github.com/krzachariassen/ZTDP/internal/contracts"
)

// LoadNode hydrates a contract from kind and spec
// This is the main entry point for instantiating contracts from nodes
func LoadNode(kind string, spec map[string]interface{}, metadata contracts.Metadata) (contracts.Contract, error) {
	return LoadNodeFromSpec(kind, spec, metadata)
}
