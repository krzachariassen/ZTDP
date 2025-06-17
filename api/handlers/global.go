package handlers

import (
	"log"
	"os"
	"sync"

	"github.com/krzachariassen/ZTDP/internal/graph"
)

var (
	GlobalGraph *graph.GlobalGraph
	graphStore  *graph.GraphStore
	initOnce    sync.Once
	logger      *log.Logger
)

func init() {
	// Initialize logger
	logger = log.New(os.Stdout, "[ZTDP] ", log.LstdFlags)
}

// getGraphStore returns the global graph store instance, initializing it if needed
func getGraphStore() *graph.GraphStore {
	initOnce.Do(func() {
		// In a production environment, this would use a persistent backend
		// like Redis or a database. For now, we use an in-memory backend.
		backend := graph.NewMemoryGraph()
		graphStore = graph.NewGraphStore(backend)
	})
	return graphStore
}
