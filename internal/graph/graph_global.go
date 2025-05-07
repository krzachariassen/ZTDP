package graph

import (
	"sync"
)

type GlobalGraph struct {
	Graph   *Graph
	Backend GraphBackend
	mu      sync.Mutex
}

func NewGlobalGraph(backend GraphBackend) *GlobalGraph {
	return &GlobalGraph{
		Graph:   NewGraph(),
		Backend: backend,
	}
}

func (gg *GlobalGraph) AddNode(node *Node) {
	gg.mu.Lock()
	defer gg.mu.Unlock()
	gg.Graph.Nodes[node.ID] = node
}

func (gg *GlobalGraph) AddEdge(fromID, toID, relType string) error {
	return gg.Graph.AddEdge(fromID, toID, relType)
}

func (gg *GlobalGraph) Apply(env string) (*Graph, error) {
	applied := NewGraph()

	for id, node := range gg.Graph.Nodes {
		copy := *node
		applied.Nodes[id] = &copy
	}

	for from, toList := range gg.Graph.Edges {
		applied.Edges[from] = append([]Edge{}, toList...)
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
