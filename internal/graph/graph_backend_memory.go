package graph

import (
	"errors"
	"fmt"
)

type memoryGraph struct {
	Graphs map[string]*Graph
}

func NewMemoryGraph() GraphBackend {
	return &memoryGraph{
		Graphs: make(map[string]*Graph),
	}
}

func (m *memoryGraph) getOrCreate(env string) *Graph {
	if g, ok := m.Graphs[env]; ok {
		return g
	}
	g := NewGraph()
	m.Graphs[env] = g
	return g
}

func (m *memoryGraph) AddNode(env string, node *Node) error {
	if node == nil {
		return errors.New("cannot add nil node")
	}
	return m.getOrCreate(env).AddNode(node)
}

func (m *memoryGraph) AddEdge(env, fromID, toID string) error {
	return m.getOrCreate(env).AddEdge(fromID, toID)
}

func (m *memoryGraph) GetNode(env, id string) (*Node, error) {
	return m.getOrCreate(env).GetNode(id)
}

func (m *memoryGraph) GetAll(env string) (*Graph, error) {
	g, ok := m.Graphs[env]
	if !ok {
		return nil, fmt.Errorf("graph for env %s not found", env)
	}
	return g, nil
}
