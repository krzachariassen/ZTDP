package graph

import (
	"encoding/json"

	"github.com/krzachariassen/ZTDP/internal/contracts"
)

func StructToMap(v interface{}) map[string]interface{} {
	b, _ := json.Marshal(v)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

func ResolveContract(c contracts.Contract) (*Node, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	md := c.GetMetadata()
	mdMap := StructToMap(md)
	// Marshal contract, then unmarshal only the spec field into a map
	data, _ := json.Marshal(c)
	var contractMap map[string]interface{}
	_ = json.Unmarshal(data, &contractMap)
	spec, _ := contractMap["spec"].(map[string]interface{})
	return &Node{
		ID:       c.ID(),
		Kind:     c.Kind(),
		Metadata: mdMap,
		Spec:     spec,
	}, nil
}

// This function was moved to resource_registry.go
// See resource_registry.LoadNode
