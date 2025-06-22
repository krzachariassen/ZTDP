package service

import (
	"context"
	"fmt"

	"github.com/krzachariassen/ZTDP/internal/agentFramework"
	"github.com/krzachariassen/ZTDP/internal/agentRegistry"
	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// ServiceAgent - thin delegation layer, ALL logic in domain service
type ServiceAgent struct {
	domainService *ServiceService
	logger        *logging.Logger
}

// NewServiceAgent creates a new service agent following clean architecture
func NewServiceAgent(
	graph *graph.GlobalGraph,
	aiProvider ai.AIProvider,
	eventBus *events.EventBus,
	registry agentRegistry.AgentRegistry,
) (agentRegistry.AgentInterface, error) {
	// Validate required dependencies
	if graph == nil {
		return nil, fmt.Errorf("graph is required")
	}
	if aiProvider == nil {
		return nil, fmt.Errorf("aiProvider is required for AI-native agent")
	}
	if eventBus == nil {
		return nil, fmt.Errorf("eventBus is required")
	}
	if registry == nil {
		return nil, fmt.Errorf("registry is required")
	}

	// Create domain service (owns ALL business logic)
	domainService := NewAIServiceService(graph, aiProvider, eventBus)

	// Create thin agent wrapper
	wrapper := &ServiceAgent{
		domainService: domainService,
		logger:        logging.GetLogger().ForComponent("service-agent"),
	}

	// Create dependencies for the framework
	deps := agentFramework.AgentDependencies{
		Registry: registry,
		EventBus: eventBus,
	}

	// Build the agent using the framework
	agent, err := agentFramework.NewAgent("service-agent").
		WithType("service").
		WithCapabilities(getServiceCapabilities()).
		WithEventHandler(wrapper.handleEvent).
		Build(deps)

	if err != nil {
		return nil, fmt.Errorf("failed to build service agent: %w", err)
	}

	wrapper.logger.Info("✅ AI-native ServiceAgent created successfully")
	return agent, nil
}

// getServiceCapabilities returns the capabilities for the service agent - ONLY service domain
func getServiceCapabilities() []agentRegistry.AgentCapability {
	return []agentRegistry.AgentCapability{
		{
			Name:        "service_management",
			Description: "AI-native service lifecycle management with intelligent parameter extraction",
			Intents: []string{
				"create service", "list services", "get service", "update service", "version service",
				"delete service", "service management", "service versioning", "manage service versions",
				"deploy service", "service configuration", "service discovery",
			},
			InputTypes:  []string{"user_message"},
			OutputTypes: []string{"service_result", "service_status", "service_list", "version_info", "clarification"},
			RoutingKeys: []string{"service.request", "service.create", "service.list", "service.version", "service.management"},
			Version:     "2.0.0",
		},
	}
}

// handleEvent - thin delegation layer, NO business logic
func (a *ServiceAgent) handleEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	a.logger.Info("🎯 Service Agent delegating to domain service: %s", event.Subject)

	// 1. Validate standardized payload (NO fallbacks - we own all components)
	userMessage, exists := event.Payload["user_message"].(string)
	if !exists {
		return a.createErrorResponse(event, "user_message field is required in event payload"), nil
	}

	// 2. Delegate ALL logic to domain service
	return a.domainService.HandleServiceEvent(ctx, event, userMessage)
}

// Helper method for error responses
func (a *ServiceAgent) createErrorResponse(originalEvent *events.Event, errorMsg string) *events.Event {
	return &events.Event{
		ID:      fmt.Sprintf("service-agent-error-%d", 1),
		Subject: "service.error",
		Payload: map[string]interface{}{
			"status":         "error",
			"message":        errorMsg,
			"correlation_id": originalEvent.Payload["correlation_id"],
			"request_id":     originalEvent.Payload["request_id"],
		},
	}
}
