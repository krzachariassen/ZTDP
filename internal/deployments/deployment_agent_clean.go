package deployments

import (
	"context"
	"fmt"
	"strings"

	"github.com/krzachariassen/ZTDP/internal/agentFramework"
	"github.com/krzachariassen/ZTDP/internal/agentRegistry"
	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// FrameworkDeploymentAgent is a MINIMAL agent that only handles deployment intents
// ALL business logic is in the Service (Clean Architecture)
type FrameworkDeploymentAgent struct {
	service *Service
	logger  *logging.Logger
}

// NewDeploymentAgent creates a minimal DeploymentAgent using the agent framework
func NewDeploymentAgent(
	graph *graph.GlobalGraph,
	aiProvider ai.AIProvider,
	env string,
	eventBus *events.EventBus,
	registry agentRegistry.AgentRegistry,
) (agentRegistry.AgentInterface, error) {
	// Create the deployment service for ALL business logic
	service := NewDeploymentService(graph, aiProvider)

	// Create the minimal wrapper
	wrapper := &FrameworkDeploymentAgent{
		service: service,
		logger:  logging.GetLogger().ForComponent("deployment-agent"),
	}

	// Create dependencies for the framework
	deps := agentFramework.AgentDependencies{
		Registry: registry,
		EventBus: eventBus,
	}

	// Build the agent using the framework
	agent, err := agentFramework.NewAgent(
		"deployment",
		wrapper.processEvent, // Only process deployment events
		deps,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment agent: %w", err)
	}

	wrapper.logger.Info("✅ Deployment agent created and ready")
	return agent, nil
}

// processEvent handles ONLY deployment-specific events (thin layer)
func (a *FrameworkDeploymentAgent) processEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	a.logger.Debug("🔄 Processing deployment event: %s", event.Type)

	// Extract user message for intent detection
	userMessage, ok := event.Data["message"].(string)
	if !ok {
		return nil, fmt.Errorf("no message found in event data")
	}

	// Handle only deployment-specific intents
	switch {
	case a.isDeployIntent(userMessage):
		return a.handleDeploy(ctx, event, userMessage)
	case a.isPlanIntent(userMessage):
		return a.handlePlan(ctx, event, userMessage)
	case a.isStatusIntent(userMessage):
		return a.handleStatus(ctx, event, userMessage)
	default:
		// Not a deployment intent - let orchestrator handle it
		return nil, fmt.Errorf("not a deployment intent")
	}
}

// isDeployIntent checks if the message is asking to deploy something
func (a *FrameworkDeploymentAgent) isDeployIntent(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "deploy") &&
		(strings.Contains(lower, "to") || strings.Contains(lower, "application"))
}

// isPlanIntent checks if the message is asking for a deployment plan
func (a *FrameworkDeploymentAgent) isPlanIntent(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "plan") && strings.Contains(lower, "deploy")
}

// isStatusIntent checks if the message is asking for deployment status
func (a *FrameworkDeploymentAgent) isStatusIntent(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "status") &&
		(strings.Contains(lower, "deploy") || strings.Contains(lower, "application"))
}

// handleDeploy handles deployment requests (delegates to service)
func (a *FrameworkDeploymentAgent) handleDeploy(ctx context.Context, event *events.Event, userMessage string) (*events.Event, error) {
	a.logger.Info("🚀 Handling deployment request")

	// Parse app and environment from message (simple parsing)
	app, env, err := a.parseDeployRequest(userMessage)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("Failed to parse deployment request: %v", err))
	}

	// Delegate ALL business logic to service
	result, err := a.service.DeployApplication(ctx, app, env)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("Deployment failed: %v", err))
	}

	// Return success response
	return a.createSuccessResponse(event, fmt.Sprintf("✅ Deployment successful: %s to %s", app, env), result)
}

// handlePlan handles plan generation requests (delegates to service)
func (a *FrameworkDeploymentAgent) handlePlan(ctx context.Context, event *events.Event, userMessage string) (*events.Event, error) {
	a.logger.Info("📋 Handling plan generation request")

	// Parse app from message
	app, err := a.parseAppFromMessage(userMessage)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("Failed to parse application: %v", err))
	}

	// Delegate ALL business logic to service
	plan, err := a.service.GenerateDeploymentPlan(ctx, app)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("Plan generation failed: %v", err))
	}

	// Return success response
	return a.createSuccessResponse(event, fmt.Sprintf("📋 Deployment plan for %s", app), plan)
}

// handleStatus handles status requests (delegates to service)
func (a *FrameworkDeploymentAgent) handleStatus(ctx context.Context, event *events.Event, userMessage string) (*events.Event, error) {
	a.logger.Info("📊 Handling status request")

	// Parse app and environment from message
	app, env, err := a.parseStatusRequest(userMessage)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("Failed to parse status request: %v", err))
	}

	// Delegate ALL business logic to service
	status, err := a.service.GetDeploymentStatus(app, env)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("Status check failed: %v", err))
	}

	// Return success response
	return a.createSuccessResponse(event, fmt.Sprintf("📊 Status for %s in %s", app, env), status)
}

// parseDeployRequest extracts app and environment from deploy message
func (a *FrameworkDeploymentAgent) parseDeployRequest(message string) (string, string, error) {
	// Simple parsing - look for "deploy [app] to [env]"
	words := strings.Fields(strings.ToLower(message))

	for i, word := range words {
		if word == "deploy" && i+3 < len(words) && words[i+2] == "to" {
			return words[i+1], words[i+3], nil
		}
	}

	return "", "", fmt.Errorf("could not parse deployment request - expected format: 'deploy [app] to [env]'")
}

// parseAppFromMessage extracts application name from message
func (a *FrameworkDeploymentAgent) parseAppFromMessage(message string) (string, error) {
	// Simple parsing - look for "plan for [app]" or "plan [app]"
	words := strings.Fields(strings.ToLower(message))

	for i, word := range words {
		if word == "plan" && i+1 < len(words) {
			if words[i+1] == "for" && i+2 < len(words) {
				return words[i+2], nil
			}
			return words[i+1], nil
		}
	}

	return "", fmt.Errorf("could not parse application name from message")
}

// parseStatusRequest extracts app and environment from status message
func (a *FrameworkDeploymentAgent) parseStatusRequest(message string) (string, string, error) {
	// Simple parsing - look for "status of [app] in [env]"
	words := strings.Fields(strings.ToLower(message))

	for i, word := range words {
		if word == "status" && i+4 < len(words) && words[i+1] == "of" && words[i+3] == "in" {
			return words[i+2], words[i+4], nil
		}
	}

	return "", "", fmt.Errorf("could not parse status request - expected format: 'status of [app] in [env]'")
}

// createSuccessResponse creates a success event response
func (a *FrameworkDeploymentAgent) createSuccessResponse(originalEvent *events.Event, message string, data interface{}) (*events.Event, error) {
	return &events.Event{
		ID:     originalEvent.ID + "-response",
		Type:   "deployment.response",
		Source: "deployment-agent",
		Data: map[string]interface{}{
			"success": true,
			"message": message,
			"data":    data,
		},
	}, nil
}

// createErrorResponse creates an error event response
func (a *FrameworkDeploymentAgent) createErrorResponse(originalEvent *events.Event, message string) (*events.Event, error) {
	return &events.Event{
		ID:     originalEvent.ID + "-error",
		Type:   "deployment.error",
		Source: "deployment-agent",
		Data: map[string]interface{}{
			"success": false,
			"error":   message,
		},
	}, nil
}
