package graph

type GraphBackend interface {
	// Global graph operations (the only storage mechanism)
	SaveGlobal(g *Graph) error
	LoadGlobal() (*Graph, error)

	// Clear removes all global data (useful for testing)
	Clear() error
}
