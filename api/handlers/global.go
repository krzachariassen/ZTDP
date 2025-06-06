package handlers

import (
	"log"
	"os"
	"sync"

	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

var (
	GlobalGraph   *graph.GlobalGraph
	GlobalV3Agent *ai.V3Agent // Using V3 Agent - the ultra simple one!
	graphStore    *graph.GraphStore
	initOnce      sync.Once
	logger        *log.Logger
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

// getGlobalGraph returns the initialized global graph instance
func getGlobalGraph() *graph.GlobalGraph {
	return GlobalGraph
}

// InitializeGlobalV3Agent initializes the global V3 AI agent with proper service dependencies
// This should be called once during application startup in main.go
func InitializeGlobalV3Agent(
	deploymentService ai.DeploymentService,
	policyService ai.PolicyService,
	applicationService ai.ApplicationService,
	serviceService ai.ServiceService,
	resourceService ai.ResourceService,
	environmentService ai.EnvironmentService,
) error {
	// Create AI provider the simple way!
	config := ai.DefaultOpenAIConfig()
	apiKey := os.Getenv("OPENAI_API_KEY")

	provider, err := ai.NewOpenAIProvider(config, apiKey)
	if err != nil {
		return err
	}

	// Create the V3 Agent - ultra simple ChatGPT-style agent!
	GlobalV3Agent = ai.NewV3Agent(
		provider,
		GlobalGraph,
		applicationService,
		serviceService,
		resourceService,
		environmentService,
		deploymentService,
		policyService,
	)

	return nil
}

// GetGlobalV3Agent returns the initialized global V3 AI agent
// Returns nil if the agent hasn't been initialized
func GetGlobalV3Agent() *ai.V3Agent {
	return GlobalV3Agent
}
