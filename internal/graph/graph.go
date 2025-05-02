package graph

import (
	"errors"
	"fmt"
)

type Graph struct {
	Nodes map[string]*Node
	Edges map[string][]string // from -> [to...]
}

func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[string]*Node),
		Edges: make(map[string][]string),
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
	n, exists := g.Nodes[id]
	if !exists {
		return nil, fmt.Errorf("node with ID %s not found", id)
	}
	return n, nil
}

func (g *Graph) AddEdge(fromID, toID string) error {
	if _, ok := g.Nodes[fromID]; !ok {
		return fmt.Errorf("source node %s does not exist", fromID)
	}
	if _, ok := g.Nodes[toID]; !ok {
		return fmt.Errorf("target node %s does not exist", toID)
	}

	// Prevent duplicate edges
	for _, existing := range g.Edges[fromID] {
		if existing == toID {
			return errors.New("edge already exists")
		}
	}

	// Add directed edge
	g.Edges[fromID] = append(g.Edges[fromID], toID)
	return nil
}
