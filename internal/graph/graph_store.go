package graph

type GraphStore struct {
	backend GraphBackend
}

func NewGraphStore(backend GraphBackend) *GraphStore {
	return &GraphStore{backend: backend}
}

func (gs *GraphStore) AddNode(env string, node *Node) error {
	return gs.backend.AddNode(env, node)
}

func (gs *GraphStore) AddEdge(env, fromID, toID string) error {
	return gs.backend.AddEdge(env, fromID, toID)
}

func (gs *GraphStore) GetNode(env, id string) (*Node, error) {
	return gs.backend.GetNode(env, id)
}

func (gs *GraphStore) GetGraph(env string) (*Graph, error) {
	return gs.backend.GetAll(env)
}
