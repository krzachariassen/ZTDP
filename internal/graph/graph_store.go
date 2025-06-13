package graph

type GraphStore struct {
	backend GraphBackend
}

func NewGraphStore(backend GraphBackend) *GraphStore {
	return &GraphStore{backend: backend}
}

func (gs *GraphStore) AddNode(env string, node *Node) error {
	// Load global graph
	graph, err := gs.backend.LoadGlobal()
	if err != nil {
		// If graph doesn't exist yet, create a new one
		graph = &Graph{
			Nodes: make(map[string]*Node),
			Edges: make(map[string][]Edge),
		}
	}

	// Add environment metadata to node
	if node.Metadata == nil {
		node.Metadata = make(map[string]interface{})
	}
	node.Metadata["environment"] = env

	// Add node to graph
	graph.Nodes[node.ID] = node

	// Save back to backend
	return gs.backend.SaveGlobal(graph)
}

func (gs *GraphStore) AddEdge(env, fromID, toID, relType string) error {
	// Load global graph
	graph, err := gs.backend.LoadGlobal()
	if err != nil {
		// If graph doesn't exist yet, create a new one
		graph = &Graph{
			Nodes: make(map[string]*Node),
			Edges: make(map[string][]Edge),
		}
	}

	// Create edge with environment metadata
	edge := Edge{
		To:   toID,
		Type: relType,
		Metadata: map[string]interface{}{
			"environment": env,
		},
	}

	// Add edge to graph (edges are stored as slice per fromID)
	graph.Edges[fromID] = append(graph.Edges[fromID], edge)

	// Save back to backend
	return gs.backend.SaveGlobal(graph)
}

func (gs *GraphStore) GetNode(env, id string) (*Node, error) {
	// Load global graph
	graph, err := gs.backend.LoadGlobal()
	if err != nil {
		return nil, err
	}

	// Find node and check environment
	node, exists := graph.Nodes[id]
	if !exists {
		return nil, nil
	}

	// Check if node belongs to the specified environment
	if nodeEnv, ok := node.Metadata["environment"].(string); !ok || nodeEnv != env {
		return nil, nil
	}

	return node, nil
}

func (gs *GraphStore) GetGraph(env string) (*Graph, error) {
	// Load global graph
	globalGraph, err := gs.backend.LoadGlobal()
	if err != nil {
		return nil, err
	}

	// Create filtered graph for the environment
	filteredGraph := &Graph{
		Nodes: make(map[string]*Node),
		Edges: make(map[string][]Edge),
	}

	// Filter nodes by environment
	for id, node := range globalGraph.Nodes {
		if nodeEnv, ok := node.Metadata["environment"].(string); ok && nodeEnv == env {
			filteredGraph.Nodes[id] = node
		}
	}

	// Filter edges by environment
	for fromID, edges := range globalGraph.Edges {
		var filteredEdges []Edge
		for _, edge := range edges {
			if edgeEnv, ok := edge.Metadata["environment"].(string); ok && edgeEnv == env {
				filteredEdges = append(filteredEdges, edge)
			}
		}
		if len(filteredEdges) > 0 {
			filteredGraph.Edges[fromID] = filteredEdges
		}
	}

	return filteredGraph, nil
}
