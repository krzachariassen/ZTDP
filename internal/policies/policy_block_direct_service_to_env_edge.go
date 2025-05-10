package policies

import "fmt"

// BlockDirectServiceToEnvEdgePolicy blocks direct service-to-environment edges.
type BlockDirectServiceToEnvEdgePolicy struct{}

func NewBlockDirectServiceToEnvEdgePolicy() *BlockDirectServiceToEnvEdgePolicy {
	return &BlockDirectServiceToEnvEdgePolicy{}
}

func (p *BlockDirectServiceToEnvEdgePolicy) Name() string {
	return "BlockDirectServiceToEnvEdgePolicy"
}

func (p *BlockDirectServiceToEnvEdgePolicy) Validate(g GraphView, m Mutation) error {
	if m.Type != "add_edge" || m.Edge == nil {
		return nil
	}
	if m.Edge.Type == "deployed_in" {
		fromNode, ok := g.Nodes[m.Edge.From]
		toNode, ok2 := g.Nodes[m.Edge.To]
		if ok && ok2 && fromNode.Kind == "service" && toNode.Kind == "environment" {
			return fmt.Errorf("direct service-to-environment 'deployed_in' edges are not allowed")
		}
	}
	return nil
}
