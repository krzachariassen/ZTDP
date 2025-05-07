package graph

import (
	"encoding/json"

	"github.com/krzachariassen/ZTDP/internal/contracts"
)

type Node struct {
	ID       string                 `json:"id"`
	Kind     string                 `json:"kind"`
	Metadata contracts.Metadata     `json:"metadata"`
	Spec     map[string]interface{} `json:"spec"`
}

func ResolveContract(c contracts.Contract) (*Node, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	md := c.GetMetadata()
	// Marshal contract, then extract the spec as a map
	data, _ := json.Marshal(c)
	var contractMap map[string]interface{}
	_ = json.Unmarshal(data, &contractMap)
	spec, _ := contractMap["spec"].(map[string]interface{})
	return &Node{
		ID:       c.ID(),
		Kind:     c.Kind(),
		Metadata: md,
		Spec:     spec,
	}, nil
}
