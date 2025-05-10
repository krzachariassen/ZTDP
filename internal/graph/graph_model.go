package graph

import (
	"errors"
	"fmt"

	"github.com/krzachariassen/ZTDP/internal/policies"
)

var policyRegistry *policies.PolicyRegistry

func SetPolicyRegistry(reg *policies.PolicyRegistry) {
	policyRegistry = reg
}

type Edge struct {
	To   string
	Type string
}

type Graph struct {
	Nodes map[string]*Node
	Edges map[string][]Edge
}

type Node struct {
	ID       string
	Kind     string
	Metadata map[string]interface{}
	Spec     map[string]interface{}
}

func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[string]*Node),
		Edges: make(map[string][]Edge),
	}
}

func (g *Graph) AddNode(n *Node) error {
	if _, exists := g.Nodes[n.ID]; exists {
		return fmt.Errorf("node with ID %s already exists", n.ID)
	}
	g.Nodes[n.ID] = n
	return nil
}

func (g *Graph) GetNode(id string) (*Node, error) {
	n, ok := g.Nodes[id]
	if !ok {
		return nil, fmt.Errorf("node with ID %s not found", id)
	}
	return n, nil
}

func (g *Graph) AddEdge(fromID, toID, relType string) error {
	if _, ok := g.Nodes[fromID]; !ok {
		return fmt.Errorf("source node %s does not exist", fromID)
	}
	if _, ok := g.Nodes[toID]; !ok {
		return fmt.Errorf("target node %s does not exist", toID)
	}
	for _, existing := range g.Edges[fromID] {
		if existing.To == toID && existing.Type == relType {
			return errors.New("edge already exists")
		}
	}
	// Policy enforcement (restored)
	if policyRegistry != nil {
		for _, p := range policyRegistry.All() {
			mutation := policies.Mutation{
				Type: "add_edge",
				Edge: &policies.EdgeView{
					From: fromID,
					To:   toID,
					Type: relType,
				},
			}
			gv := policies.GraphView{
				Nodes: make(map[string]policies.NodeView),
				Edges: make(map[string][]policies.EdgeView),
			}
			for id, node := range g.Nodes {
				gv.Nodes[id] = policies.NodeView{
					ID:       node.ID,
					Kind:     node.Kind,
					Metadata: node.Metadata,
					Spec:     node.Spec,
				}
			}
			for from, edges := range g.Edges {
				for _, edge := range edges {
					gv.Edges[from] = append(gv.Edges[from], policies.EdgeView{
						From: from,
						To:   edge.To,
						Type: edge.Type,
					})
				}
			}
			if err := p.Validate(gv, mutation); err != nil {
				return err
			}
		}
	}
	g.Edges[fromID] = append(g.Edges[fromID], Edge{To: toID, Type: relType})
	return nil
}
