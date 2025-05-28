package graph

import (
	"errors"
	"fmt"
)

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

	// Enforce policy for deploy edges
	if relType == "deploy" {
		if err := g.IsTransitionAllowed(fromID, toID, relType); err != nil {
			return err
		}
	}

	g.Edges[fromID] = append(g.Edges[fromID], Edge{To: toID, Type: relType})
	return nil
}

// UpdateNode updates an existing node in the graph.
// If the node doesn't exist, an error is returned.
func (g *Graph) UpdateNode(node *Node) error {
	if _, exists := g.Nodes[node.ID]; !exists {
		return fmt.Errorf("node with ID %s not found", node.ID)
	}
	g.Nodes[node.ID] = node
	return nil
}
