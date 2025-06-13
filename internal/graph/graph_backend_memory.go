package graph

import (
	"fmt"
)

type memoryGraph struct {
	Global *Graph
}

func NewMemoryGraph() GraphBackend {
	return &memoryGraph{
		Global: NewGraph(),
	}
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

// Clear removes all global data (useful for testing)
func (m *memoryGraph) Clear() error {
	m.Global = NewGraph()
	return nil
}
