package application

import (
	aiDomain "github.com/ztdp/orchestrator/internal/ai/domain"
	aiInfrastructure "github.com/ztdp/orchestrator/internal/ai/infrastructure"
	"github.com/ztdp/orchestrator/internal/graph"
	"github.com/ztdp/orchestrator/internal/logging"
	"github.com/ztdp/orchestrator/internal/messaging"
	"github.com/ztdp/orchestrator/internal/orchestrator/infrastructure"
)

// ServiceFactory creates properly wired orchestrator service instances
type ServiceFactory struct {
	logger     logging.Logger
	graph      graph.Graph
	messageBus messaging.MessageBus
	aiProvider aiDomain.AIProvider
}

// NewServiceFactory creates a new service factory
func NewServiceFactory(
	logger logging.Logger,
	graph graph.Graph,
	messageBus messaging.MessageBus,
	aiProvider aiDomain.AIProvider,
) *ServiceFactory {
	return &ServiceFactory{
		logger:     logger,
		graph:      graph,
		messageBus: messageBus,
		aiProvider: aiProvider,
	}
}

// CreateOrchestratorService creates a fully wired orchestrator service
func (sf *ServiceFactory) CreateOrchestratorService() *OrchestratorService {
	// Create infrastructure services
	agentService := infrastructure.NewGraphAgentService(sf.graph)
	executionService := infrastructure.NewGraphExecutionService(sf.graph)
	conversationService := infrastructure.NewGraphConversationService(sf.graph)

	// Create all application services with proper dependencies
	aiDecisionEngine := NewAIDecisionEngine(sf.aiProvider)
	graphExplorer := NewGraphExplorer(agentService)
	executionCoordinator := NewExecutionCoordinator(executionService)
	learningService := NewLearningService(conversationService)

	// Wire everything together
	return NewOrchestratorService(
		aiDecisionEngine,
		graphExplorer,
		executionCoordinator,
		learningService,
	)
}

// CreateAIProvider creates an AI provider with the given configuration
func CreateAIProvider(config *aiInfrastructure.OpenAIConfig, logger logging.Logger) aiDomain.AIProvider {
	return aiInfrastructure.NewOpenAIProvider(config, logger)
}
