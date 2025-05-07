package graph

import (
	"encoding/json"
	"fmt"

	"github.com/krzachariassen/ZTDP/internal/contracts"
)

// ResourceRegistry maps kind to a factory for contract types
var ResourceRegistry = map[string]func() contracts.Contract{
	"application": func() contracts.Contract { return &contracts.ApplicationContract{} },
	"service":     func() contracts.Contract { return &contracts.ServiceContract{} },
}

// LoadNode hydrates a contract from kind and spec
func LoadNode(kind string, spec map[string]interface{}, metadata contracts.Metadata) (contracts.Contract, error) {
	factory, ok := ResourceRegistry[kind]
	if !ok {
		return nil, fmt.Errorf("unknown resource kind: %s", kind)
	}
	resource := factory()
	// Build the full contract map
	contractMap := map[string]interface{}{
		"metadata": metadata,
		"spec":     spec,
	}
	data, _ := json.Marshal(contractMap)
	if err := json.Unmarshal(data, resource); err != nil {
		return nil, err
	}
	return resource, nil
}
