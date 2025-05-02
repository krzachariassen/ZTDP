package graph

type GraphBackend interface {
	AddNode(env string, n *Node) error
	AddEdge(env string, fromID, toID string) error
	GetNode(env, id string) (*Node, error)
	GetAll(env string) (*Graph, error)

	SaveGlobal(g *Graph) error
	LoadGlobal() (*Graph, error)
}
