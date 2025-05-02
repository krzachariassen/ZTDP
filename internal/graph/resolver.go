package graph

import "github.com/krzachariassen/ZTDP/internal/contracts"

type Node struct {
	ID       string
	Kind     string
	Metadata contracts.Metadata
	Spec     interface{}
}

func ResolveContract(c contracts.Contract) (*Node, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	md := c.GetMetadata()

	return &Node{
		ID:       c.ID(),
		Kind:     c.Kind(),
		Metadata: md,
		Spec:     c,
	}, nil
}
