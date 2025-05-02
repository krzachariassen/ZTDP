package graph

import (
	"errors"
	"fmt"
)

type memoryGraph struct {
	Graphs map[string]*Graph
	Global *Graph
}

func NewMemoryGraph() GraphBackend {
	return &memoryGraph{
		Graphs: make(map[string]*Graph),
		Global: NewGraph(),
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

func (m *memoryGraph) SaveGlobal(g *Graph) error {
	m.Global = g
	return nil
}

func (m *memoryGraph) LoadGlobal() (*Graph, error) {
	if m.Global == nil {
		return nil, fmt.Errorf("no global graph stored")
	}
	return m.Global, nil
}
