package graph

import (
	"sync"
)

type GlobalGraph struct {
	Backend GraphBackend
	mu      sync.Mutex
}

func NewGlobalGraph(backend GraphBackend) *GlobalGraph {
	return &GlobalGraph{
		Backend: backend,
	}
}

// Graph returns always-fresh graph data from backend
// This enables both GlobalGraph.Graph().Nodes and currentGraph := GlobalGraph.Graph() patterns
func (gg *GlobalGraph) Graph() (*Graph, error) {
	return gg.Backend.LoadGlobal()
}

func (gg *GlobalGraph) AddNode(node *Node) {
	gg.mu.Lock()
	defer gg.mu.Unlock()

	// Get current global graph or create new one
	currentGraph, err := gg.Backend.LoadGlobal()
	if err != nil {
		currentGraph = NewGraph()
	}

	// Add node to current graph
	currentGraph.AddNode(node)

	// Save back to backend
	gg.Backend.SaveGlobal(currentGraph)
}

func (gg *GlobalGraph) AddEdge(fromID, toID, relType string) error {
	gg.mu.Lock()
	defer gg.mu.Unlock()

	// Get current graph state for policy checking
	currentGraph, err := gg.Backend.LoadGlobal()
	if err != nil {
		// If no graph exists yet, create empty one for policy check
		currentGraph = NewGraph()
	}

	// Check policies before adding the edge
	if err := currentGraph.IsTransitionAllowed(fromID, toID, relType); err != nil {
		return err
	}

	// Add edge to current graph
	if err := currentGraph.AddEdge(fromID, toID, relType); err != nil {
		return err
	}

	// Save back to backend
	return gg.Backend.SaveGlobal(currentGraph)
}

func (gg *GlobalGraph) Apply(env string) (*Graph, error) {
	// Always get fresh data from backend
	return gg.Backend.LoadGlobal()
}

func (gg *GlobalGraph) Save() error {
	// Get current graph and save it (for compatibility with tests that expect explicit save)
	currentGraph, err := gg.Backend.LoadGlobal()
	if err != nil {
		return err
	}
	return gg.Backend.SaveGlobal(currentGraph)
}

func (gg *GlobalGraph) Load() error {
	// For the new architecture, this is a no-op since we always read from backend
	// But we keep it for compatibility
	return nil
}

// HasEdge checks if an edge exists by querying the backend
func (gg *GlobalGraph) HasEdge(fromID, toID, relType string) (bool, error) {
	currentGraph, err := gg.Backend.LoadGlobal()
	if err != nil {
		return false, err
	}

	if edges, ok := currentGraph.Edges[fromID]; ok {
		for _, edge := range edges {
			if edge.To == toID && edge.Type == relType {
				return true, nil
			}
		}
	}
	return false, nil
}

// HasDeploymentEdge checks if a deployment edge exists
func (gg *GlobalGraph) HasDeploymentEdge(serviceID, environment string) (bool, error) {
	return gg.HasEdge(serviceID, environment, "deploy")
}

// Convenience methods for direct access to fresh data
// These enable GlobalGraph.Nodes() and GlobalGraph.Edges() syntax

// Nodes returns fresh nodes from backend
// For AI-native platform: gracefully handle backend errors and return empty data
func (gg *GlobalGraph) Nodes() (map[string]*Node, error) {
	g, err := gg.Backend.LoadGlobal()
	if err != nil {
		// Log error but return empty map for graceful degradation
		// This ensures AI agents get empty results instead of errors when Redis is down/empty
		return make(map[string]*Node), nil
	}
	return g.Nodes, nil
}

// Edges returns fresh edges from backend
// For AI-native platform: gracefully handle backend errors and return empty data
func (gg *GlobalGraph) Edges() (map[string][]Edge, error) {
	g, err := gg.Backend.LoadGlobal()
	if err != nil {
		// Return empty map for graceful degradation
		return make(map[string][]Edge), nil
	}
	return g.Edges, nil
}

// GetNode returns a fresh node from backend
// For AI-native platform: gracefully handle backend errors
func (gg *GlobalGraph) GetNode(id string) (*Node, error) {
	g, err := gg.Backend.LoadGlobal()
	if err != nil {
		// Return nil (not found) when backend is unavailable
		return nil, nil
	}
	return g.GetNode(id)
}

// Policy convenience methods

// AttachPolicyToTransition attaches a policy to a specific transition
func (gg *GlobalGraph) AttachPolicyToTransition(fromID, toID, edgeType, policyID string) error {
	gg.mu.Lock()
	defer gg.mu.Unlock()

	// Get current graph state
	currentGraph, err := gg.Backend.LoadGlobal()
	if err != nil {
		// If no graph exists yet, create empty one
		currentGraph = NewGraph()
	}

	// Use the graph's AttachPolicyToTransition method
	if err := currentGraph.AttachPolicyToTransition(fromID, toID, edgeType, policyID); err != nil {
		return err
	}

	// Save back to backend
	return gg.Backend.SaveGlobal(currentGraph)
}

// GetEdge retrieves an edge from the global graph
func (gg *GlobalGraph) GetEdge(edgeID string) (*Edge, bool) {
	gg.mu.Lock()
	defer gg.mu.Unlock()

	// Get current graph state
	currentGraph, err := gg.Backend.LoadGlobal()
	if err != nil {
		return nil, false
	}

	return currentGraph.GetEdge(edgeID)
}

// UpdateEdge updates an edge in the global graph
func (gg *GlobalGraph) UpdateEdge(edge *Edge) error {
	gg.mu.Lock()
	defer gg.mu.Unlock()

	// Get current graph state
	currentGraph, err := gg.Backend.LoadGlobal()
	if err != nil {
		return err
	}

	// Update the edge
	if err := currentGraph.UpdateEdge(edge); err != nil {
		return err
	}

	// Save back to backend
	return gg.Backend.SaveGlobal(currentGraph)
}

// GetEdgeByFromToType retrieves an edge by explicit from, to, and type parameters
func (gg *GlobalGraph) GetEdgeByFromToType(fromID, toID, edgeType string) (*Edge, bool) {
	gg.mu.Lock()
	defer gg.mu.Unlock()

	// Get current graph state
	currentGraph, err := gg.Backend.LoadGlobal()
	if err != nil {
		return nil, false
	}

	return currentGraph.GetEdgeByFromToType(fromID, toID, edgeType)
}
