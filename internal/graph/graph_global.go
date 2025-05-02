package graph

import (
	"fmt"
)

type GlobalGraph struct {
	Graph   *Graph
	Backend GraphBackend
}

func NewGlobalGraph(backend GraphBackend) *GlobalGraph {
	return &GlobalGraph{
		Graph:   NewGraph(),
		Backend: backend,
	}
}

func (gg *GlobalGraph) AddNode(node *Node) error {
	if node == nil {
		return fmt.Errorf("nil node")
	}
	return gg.Graph.AddNode(node)
}

func (gg *GlobalGraph) AddEdge(fromID, toID string) error {
	return gg.Graph.AddEdge(fromID, toID)
}

func (gg *GlobalGraph) Apply(env string) (*Graph, error) {
	applied := NewGraph()

	for id, node := range gg.Graph.Nodes {
		copy := *node
		applied.Nodes[id] = &copy
	}

	for from, toList := range gg.Graph.Edges {
		applied.Edges[from] = append([]string{}, toList...)
	}

	return applied, nil
}

func (gg *GlobalGraph) Save() error {
	return gg.Backend.SaveGlobal(gg.Graph)
}

func (gg *GlobalGraph) Load() error {
	g, err := gg.Backend.LoadGlobal()
	if err != nil {
		return err
	}
	gg.Graph = g
	return nil
}
