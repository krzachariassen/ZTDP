package environment

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

// EnvironmentAgent - thin delegation layer, ALL logic in domain service
type EnvironmentAgent struct {
	domainService *EnvironmentService
	logger        *logging.Logger
}

// NewEnvironmentAgent creates a new environment agent following clean architecture
func NewEnvironmentAgent(
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
	domainService := NewAIEnvironmentService(graph, aiProvider, eventBus)

	// Create thin agent wrapper
	wrapper := &EnvironmentAgent{
		domainService: domainService,
		logger:        logging.GetLogger().ForComponent("environment-agent"),
	}

	// Create dependencies for the framework
	deps := agentFramework.AgentDependencies{
		Registry: registry,
		EventBus: eventBus,
	}

	// Build the agent using the framework
	agent, err := agentFramework.NewAgent("environment-agent").
		WithType("environment").
		WithCapabilities(getEnvironmentCapabilities()).
		WithEventHandler(wrapper.handleEvent).
		Build(deps)

	if err != nil {
		return nil, fmt.Errorf("failed to build environment agent: %w", err)
	}

	wrapper.logger.Info("✅ AI-native EnvironmentAgent created successfully")
	return agent, nil
}

// getEnvironmentCapabilities returns the capabilities for the environment agent - ONLY environment domain
func getEnvironmentCapabilities() []agentRegistry.AgentCapability {
	return []agentRegistry.AgentCapability{
		{
			Name:        "environment_management",
			Description: "AI-native environment lifecycle management with intelligent parameter extraction",
			Intents: []string{
				"create environment", "list environments", "get environment", "update environment",
				"delete environment", "environment management", "environment configuration",
				"manage environments", "setup environment", "environment permissions",
			},
			InputTypes:  []string{"user_message"},
			OutputTypes: []string{"environment_result", "environment_status", "environment_list", "clarification"},
			RoutingKeys: []string{"environment.request", "environment.create", "environment.list", "environment.management"},
			Version:     "2.0.0",
		},
	}
}

// handleEvent - thin delegation layer, NO business logic
func (a *EnvironmentAgent) handleEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	a.logger.Info("🎯 Environment Agent delegating to domain service: %s", event.Subject)

	// 1. Validate standardized payload (NO fallbacks - we own all components)
	userMessage, exists := event.Payload["user_message"].(string)
	if !exists {
		return a.createErrorResponse(event, "user_message field is required in event payload"), nil
	}

	// 2. Delegate ALL logic to domain service
	return a.domainService.HandleEnvironmentEvent(ctx, event, userMessage)
}

// Helper method for error responses
func (a *EnvironmentAgent) createErrorResponse(originalEvent *events.Event, errorMsg string) *events.Event {
	return &events.Event{
		ID:      fmt.Sprintf("environment-agent-error-%d", 1),
		Subject: "environment.error",
		Payload: map[string]interface{}{
			"status":         "error",
			"message":        errorMsg,
			"correlation_id": originalEvent.Payload["correlation_id"],
			"request_id":     originalEvent.Payload["request_id"],
		},
	}
}
