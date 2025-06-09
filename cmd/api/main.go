package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/krzachariassen/ZTDP/api/handlers"
	"github.com/krzachariassen/ZTDP/api/server"
	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/application"
	"github.com/krzachariassen/ZTDP/internal/deployments"
	"github.com/krzachariassen/ZTDP/internal/environment"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
	"github.com/krzachariassen/ZTDP/internal/policies"
	"github.com/krzachariassen/ZTDP/internal/resources"
	"github.com/krzachariassen/ZTDP/internal/service"
)

func main() {
	// Initialize centralized logging system
	logLevel := logging.LevelInfo
	if os.Getenv("ZTDP_LOG_LEVEL") == "debug" {
		logLevel = logging.LevelDebug
	}
	logging.InitializeLogger("ztdp-api", logLevel)

	// Create real-time log sink for WebSocket broadcasting
	realtimeSink := logging.NewRealtimeLogSink()
	logging.GetLogger().AddSink(realtimeSink)

	logger := logging.GetLogger()
	logger.Info("🚀 Starting ZTDP API Server")

	// Configure event system
	var eventTransport events.EventTransport

	// Check if NATS is configured
	natsURL := os.Getenv("ZTDP_NATS_URL")
	if natsURL != "" {
		logger.Info("🔔 Using NATS event transport: %s", natsURL)
		natsConfig := events.DefaultNATSConfig()
		natsConfig.URL = natsURL

		var err error
		eventTransport, err = events.NewNATSTransport(natsConfig)
		if err != nil {
			logger.Warn("⚠️ Failed to connect to NATS, falling back to memory transport: %v", err)
			eventTransport = events.NewMemoryTransport()
		}
	} else {
		logger.Info("🔔 Using in-memory event transport")
		eventTransport = events.NewMemoryTransport()
	}

	// Initialize simple event system
	events.InitializeEventBus(eventTransport)
	logger.Info("🔔 Event system initialized")

	// Initialize log manager for real-time WebSocket streaming
	handlers.InitLogManager()
	logger.Info("📊 Log manager initialized")

	var backend graph.GraphBackend
	switch os.Getenv("ZTDP_GRAPH_BACKEND") {
	case "redis":
		logger.Info("⚙️  Using backend: Redis")
		backend = graph.NewRedisGraph(graph.RedisGraphConfig{})
	default:
		logger.Info("⚙️  Using backend: Memory")
		backend = graph.NewMemoryGraph()
	}
	handlers.GlobalGraph = graph.NewGlobalGraph(backend)

	// Load persisted graph from backend (Redis)
	if err := handlers.GlobalGraph.Load(); err != nil {
		logger.Info("No existing global graph found, starting fresh")
	}

	// Initialize AI infrastructure with domain services
	logger.Info("🤖 Initializing AI platform with domain services...")

	// Create AI provider for services
	aiProvider, err := createAIProvider()
	if err != nil {
		logger.Warn("⚠️  Failed to create AI provider: %v. AI features will be limited.", err)
		aiProvider = nil
	}

	// Initialize domain services with proper dependencies
	deploymentService := createDeploymentService(handlers.GlobalGraph, aiProvider)
	policyService := createPolicyService(handlers.GlobalGraph, aiProvider)
	applicationService := createApplicationService(handlers.GlobalGraph)
	serviceService := createServiceService(handlers.GlobalGraph)
	resourceService := createResourceService(handlers.GlobalGraph)
	environmentService := createEnvironmentService(handlers.GlobalGraph)

	// Initialize global V3 AI agent with all service dependencies
	if err := handlers.InitializeGlobalV3Agent(
		deploymentService,
		policyService,
		applicationService,
		serviceService,
		resourceService,
		environmentService,
	); err != nil {
		logger.Warn("⚠️  Failed to initialize V3 AI agent: %v. AI features will be limited.", err)
	} else {
		logger.Info("🤖 AI platform agent initialized successfully")
	}

	r := server.NewRouter()

	// Add logging middleware to router
	loggedRouter := logging.CreateHTTPLoggingMiddleware("api-server")(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("🌐 Starting API server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, loggedRouter))
}

// createAIProvider creates the AI provider if configured
func createAIProvider() (ai.AIProvider, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	config := ai.DefaultOpenAIConfig()
	if model := os.Getenv("OPENAI_MODEL"); model != "" {
		config.Model = model
	}
	if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
		config.BaseURL = baseURL
	}

	return ai.NewOpenAIProvider(config, apiKey)
}

// createDeploymentService initializes the deployment service
func createDeploymentService(globalGraph *graph.GlobalGraph, aiProvider ai.AIProvider) ai.DeploymentService {
	return deployments.NewDeploymentService(globalGraph, aiProvider)
}

// createPolicyService initializes the policy service
func createPolicyService(globalGraph *graph.GlobalGraph, aiProvider ai.AIProvider) ai.PolicyService {
	graphStore := graph.NewGraphStore(globalGraph.Backend)
	return policies.NewService(graphStore, globalGraph, os.Getenv("ZTDP_ENV"))
}

// createApplicationService initializes the application service
func createApplicationService(globalGraph *graph.GlobalGraph) ai.ApplicationService {
	return application.NewService(globalGraph)
}

// createServiceService initializes the service service
func createServiceService(globalGraph *graph.GlobalGraph) ai.ServiceService {
	return service.NewServiceService(globalGraph)
}

// createResourceService initializes the resource service
func createResourceService(globalGraph *graph.GlobalGraph) ai.ResourceService {
	return resources.NewService(globalGraph)
}

// createEnvironmentService initializes the environment service
func createEnvironmentService(globalGraph *graph.GlobalGraph) ai.EnvironmentService {
	return environment.NewService(globalGraph)
}
