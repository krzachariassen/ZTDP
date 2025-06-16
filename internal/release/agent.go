package release

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/krzachariassen/ZTDP/internal/agents"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// ReleaseAgent implements AgentInterface for intelligent release management
// This agent handles release creation, tracking, and management workflows
type ReleaseAgent struct {
	service       *Service
	agentID       string
	eventBus      *events.EventBus
	agentRegistry agents.AgentRegistry
	startTime     time.Time
	logger        *logging.Logger
}

// NewReleaseAgent creates a new ReleaseAgent that auto-registers with the agent registry
func NewReleaseAgent(graph *graph.GlobalGraph, eventBus *events.EventBus, agentRegistry agents.AgentRegistry) (agents.AgentInterface, error) {
	service := NewService(graph)
	agent := &ReleaseAgent{
		service:       service,
		agentID:       "release-agent",
		eventBus:      eventBus,
		agentRegistry: agentRegistry,
		startTime:     time.Now(),
		logger:        logging.GetLogger().ForComponent("release-agent"),
	}

	// Auto-register with the agent registry
	if agentRegistry != nil {
		ctx := context.Background()
		if err := agentRegistry.RegisterAgent(ctx, agent); err != nil {
			agent.logger.Error("❌ Failed to auto-register release agent: %v", err)
			return nil, fmt.Errorf("failed to auto-register release agent: %w", err)
		}
		agent.logger.Info("✅ ReleaseAgent auto-registered successfully")
	} else {
		agent.logger.Warn("⚠️ No agent registry provided - agent will not be discoverable")
	}

	// Subscribe to specific routing keys so the agent only receives relevant events
	if eventBus != nil {
		// Subscribe to all routing keys this agent handles
		capabilities := agent.GetCapabilities()
		for _, capability := range capabilities {
			for _, routingKey := range capability.RoutingKeys {
				eventBus.SubscribeToRoutingKey(routingKey, agent.handleIncomingEvent)
				agent.logger.Info("✅ ReleaseAgent subscribed to routing key: %s", routingKey)
			}
		}
	} else {
		agent.logger.Warn("⚠️ No event bus provided - agent will not receive events")
	}

	return agent, nil
}

// GetID returns the agent's unique identifier
func (a *ReleaseAgent) GetID() string {
	return a.agentID
}

// GetStatus returns current agent status information
func (a *ReleaseAgent) GetStatus() agents.AgentStatus {
	return agents.AgentStatus{
		ID:           a.agentID,
		Type:         "release",
		Status:       "running",
		LastActivity: time.Now(),
		LoadFactor:   0.2, // Release operations are generally lightweight
		Version:      "1.0.0",
		Metadata: map[string]interface{}{
			"operations": []string{"create_release", "get_release", "list_releases"},
		},
	}
}

// GetCapabilities returns the agent's capabilities
func (a *ReleaseAgent) GetCapabilities() []agents.AgentCapability {
	return []agents.AgentCapability{
		{
			Name:        "release_management",
			Description: "Manages application releases with AI-enhanced tracking and coordination",
			Intents:     []string{"create release", "new release", "release creation", "manage release"},
			InputTypes:  []string{"application", "service_versions", "notes"},
			OutputTypes: []string{"release_contract", "release_status", "release_list"},
			RoutingKeys: []string{"release.create", "release.get", "release.list"},
			Version:     "1.0.0",
		},
	}
}

// ProcessEvent handles incoming events for the release agent
func (a *ReleaseAgent) ProcessEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	// Validate event has required intent field
	intent, ok := event.Payload["intent"].(string)
	if !ok {
		return nil, fmt.Errorf("release agent requires 'intent' field in payload")
	}

	a.logger.Info("🤖 Processing release event with intent: %s", intent)

	// Route based on intent
	switch intent {
	case "create_release", "create release", "new release", "release creation":
		return a.handleCreateRelease(ctx, event)
	case "get_release", "get release", "show release", "release details":
		return a.handleGetRelease(ctx, event)
	case "list_releases", "list releases", "show releases", "releases":
		return a.handleListReleases(ctx, event)
	default:
		return nil, fmt.Errorf("unsupported intent: %s", intent)
	}
}

// handleIncomingEvent is the EventBus handler that routes events to ProcessEvent
func (a *ReleaseAgent) handleIncomingEvent(event events.Event) error {
	a.logger.Info("📨 ReleaseAgent received event: %s from %s", event.Subject, event.Source)

	// Debug: Check what's in the event payload
	if correlationID, ok := event.Payload["correlation_id"]; ok {
		a.logger.Info("🔍 Event contains correlation_id: %v", correlationID)
	} else {
		a.logger.Warn("⚠️ Event missing correlation_id in payload")
	}

	// Check if this event is intended for this agent (or broadcast)
	targetAgent, hasTarget := event.Payload["target_agent"].(string)
	if hasTarget && targetAgent != a.agentID && targetAgent != "*" {
		// Event is for a different agent, ignore it
		return nil
	}

	// Process the event using the agent's main processing logic
	ctx := context.Background()
	responseEvent, err := a.ProcessEvent(ctx, &event)
	if err != nil {
		a.logger.Error("❌ Failed to process event: %v", err)
		return err
	}

	// If we have a response event, emit it back to the event bus
	if responseEvent != nil {
		if a.eventBus != nil {
			// Preserve correlation information for agent-to-agent communication
			if correlationID, ok := event.Payload["correlation_id"]; ok {
				if responseEvent.Payload == nil {
					responseEvent.Payload = make(map[string]interface{})
				}
				responseEvent.Payload["correlation_id"] = correlationID
			}
			if requestID, ok := event.Payload["request_id"]; ok {
				if responseEvent.Payload == nil {
					responseEvent.Payload = make(map[string]interface{})
				}
				responseEvent.Payload["request_id"] = requestID
			}

			a.eventBus.EmitEvent(*responseEvent)
			a.logger.Info("✅ ReleaseAgent sent response: %s", responseEvent.Subject)
		} else {
			a.logger.Warn("⚠️ No event bus available to send response")
		}
	} else {
		a.logger.Warn("⚠️ No response event generated from ProcessEvent")
	}

	return nil
}

// handleCreateRelease processes release creation requests
func (a *ReleaseAgent) handleCreateRelease(ctx context.Context, event *events.Event) (*events.Event, error) {
	application, ok := event.Payload["application"].(string)
	if !ok {
		return a.createErrorResponse(event, "application required for release creation"), nil
	}

	// Extract service versions
	var serviceVersions []string
	if svs, ok := event.Payload["service_versions"].([]string); ok {
		serviceVersions = svs
	} else if svs, ok := event.Payload["service_versions"].([]interface{}); ok {
		for _, sv := range svs {
			if str, ok := sv.(string); ok {
				serviceVersions = append(serviceVersions, str)
			}
		}
	}

	// If no service versions provided, create a default one
	if len(serviceVersions) == 0 {
		serviceVersions = []string{fmt.Sprintf("%s-latest", application)}
	}

	notes := ""
	if n, ok := event.Payload["notes"].(string); ok {
		notes = n
	}

	a.logger.Info("🔨 Creating release for application %s with %d service versions", application, len(serviceVersions))

	// Call release service
	release, err := a.service.CreateReleaseFromRequest(ctx, application, "", serviceVersions, notes)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("release creation failed: %v", err)), nil
	}

	a.logger.Info("✅ Successfully created release: %s", release.ID())

	// Return success response
	return a.createSuccessResponse(event, map[string]interface{}{
		"release_id": release.ID(),
		"version":    release.Spec.Version,
		"status":     release.Spec.Status,
		"message":    fmt.Sprintf("Successfully created release %s for application %s", release.ID(), application),
	}), nil
}

// handleGetRelease processes release retrieval requests
func (a *ReleaseAgent) handleGetRelease(ctx context.Context, event *events.Event) (*events.Event, error) {
	releaseID, ok := event.Payload["release_id"].(string)
	if !ok {
		return a.createErrorResponse(event, "release_id required for release retrieval"), nil
	}

	release, err := a.service.GetRelease(releaseID)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("failed to get release: %v", err)), nil
	}

	return a.createSuccessResponse(event, map[string]interface{}{
		"release": release,
	}), nil
}

// handleListReleases processes release listing requests
func (a *ReleaseAgent) handleListReleases(ctx context.Context, event *events.Event) (*events.Event, error) {
	application := ""
	if app, ok := event.Payload["application"].(string); ok {
		application = app
	}

	releases, err := a.service.ListReleases(application)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("failed to list releases: %v", err)), nil
	}

	return a.createSuccessResponse(event, map[string]interface{}{
		"releases": releases,
		"count":    len(releases),
	}), nil
}

// Agent Discovery and Communication Methods for ReleaseAgent

// discoverAgentsByIntent finds agents that can handle a specific intent
func (a *ReleaseAgent) discoverAgentsByIntent(ctx context.Context, intent string) ([]agents.AgentStatus, error) {
	if a.agentRegistry == nil {
		return nil, fmt.Errorf("no agent registry available for discovery")
	}

	// Get all available capabilities
	capabilities, err := a.agentRegistry.GetAvailableCapabilities(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get available capabilities: %w", err)
	}

	var matchingAgents []agents.AgentStatus

	// Find agents whose capabilities match the intent
	for _, capability := range capabilities {
		for _, supportedIntent := range capability.Intents {
			if a.intentMatches(intent, supportedIntent) {
				// Find agents with this capability
				agentsWithCapability, err := a.agentRegistry.FindAgentsByCapability(ctx, capability.Name)
				if err != nil {
					a.logger.Warn("⚠️ Failed to find agents for capability %s: %v", capability.Name, err)
					continue
				}
				matchingAgents = append(matchingAgents, agentsWithCapability...)
			}
		}
	}

	return matchingAgents, nil
}

// intentMatches checks if a user intent matches a supported capability intent
func (a *ReleaseAgent) intentMatches(userIntent, supportedIntent string) bool {
	// Direct match
	if strings.EqualFold(userIntent, supportedIntent) {
		return true
	}

	// Normalize and check for substring matches
	userNormalized := strings.ToLower(strings.ReplaceAll(userIntent, "_", " "))
	supportedNormalized := strings.ToLower(strings.ReplaceAll(supportedIntent, "_", " "))

	return strings.Contains(userNormalized, supportedNormalized) ||
		strings.Contains(supportedNormalized, userNormalized)
}

// requestDeploymentUpdate sends deployment status updates back to DeploymentAgent
func (a *ReleaseAgent) requestDeploymentUpdate(ctx context.Context, correlationID string, deploymentStatus map[string]interface{}) error {
	a.logger.Info("🔄 Sending deployment update with correlation_id: %s", correlationID)

	// Discover DeploymentAgent
	matchingAgents, err := a.discoverAgentsByIntent(ctx, "deployment update")
	if err != nil || len(matchingAgents) == 0 {
		// Fallback: try to find any deployment agent
		matchingAgents, err = a.discoverAgentsByIntent(ctx, "deployment")
		if err != nil || len(matchingAgents) == 0 {
			a.logger.Warn("⚠️ No deployment agents found for status update")
			return fmt.Errorf("no deployment agents available")
		}
	}

	targetAgent := matchingAgents[0]
	a.logger.Info("🎯 ReleaseAgent updating DeploymentAgent: %s", targetAgent.ID)

	// Prepare the event payload
	payload := map[string]interface{}{
		"intent":         "deployment_status_update",
		"target_agent":   targetAgent.ID,
		"correlation_id": correlationID,
		"source_agent":   a.agentID,
		"status":         deploymentStatus,
	}

	// Emit the update event
	if a.eventBus != nil {
		a.eventBus.Emit(events.EventTypeResponse, a.agentID, "deployment.status", payload)
		a.logger.Info("📡 ReleaseAgent sent deployment update")
		return nil
	}

	return fmt.Errorf("no event bus available")
}

// notifyDeploymentCompletion notifies deployment agent about release completion
func (a *ReleaseAgent) notifyDeploymentCompletion(ctx context.Context, releaseID string, success bool, message string) error {
	deploymentStatus := map[string]interface{}{
		"release_id": releaseID,
		"success":    success,
		"message":    message,
		"timestamp":  time.Now().Unix(),
	}

	return a.requestDeploymentUpdate(ctx, releaseID, deploymentStatus)
}

// Helper methods
func (a *ReleaseAgent) createSuccessResponse(originalEvent *events.Event, data map[string]interface{}) *events.Event {
	return &events.Event{
		Type:      events.EventTypeResponse,
		Source:    a.GetID(),
		Subject:   "release_response",
		Payload:   map[string]interface{}{"status": "success", "data": data},
		Timestamp: time.Now().Unix(),
		ID:        fmt.Sprintf("resp-%d", time.Now().UnixNano()),
	}
}

func (a *ReleaseAgent) createErrorResponse(originalEvent *events.Event, errorMsg string) *events.Event {
	a.logger.Error("❌ ReleaseAgent error: %s", errorMsg)
	return &events.Event{
		Type:      events.EventTypeResponse,
		Source:    a.GetID(),
		Subject:   "release_error",
		Payload:   map[string]interface{}{"status": "error", "error": errorMsg},
		Timestamp: time.Now().Unix(),
		ID:        fmt.Sprintf("err-%d", time.Now().UnixNano()),
	}
}

// Implement the remaining AgentInterface methods
func (a *ReleaseAgent) Start(ctx context.Context) error {
	a.logger.Info("🚀 Starting ReleaseAgent...")
	return nil
}

func (a *ReleaseAgent) Stop(ctx context.Context) error {
	a.logger.Info("🛑 Stopping ReleaseAgent...")
	return nil
}

func (a *ReleaseAgent) Health() agents.HealthStatus {
	return agents.HealthStatus{
		Healthy: true,
		Status:  "healthy",
		Message: "ReleaseAgent is running normally",
	}
}
