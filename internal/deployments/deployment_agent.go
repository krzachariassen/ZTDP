package deployments

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/krzachariassen/ZTDP/internal/agents"
	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// DeploymentAgent implements AgentInterface for intelligent deployment orchestration
type DeploymentAgent struct {
	service       *Service
	agentID       string
	env           string
	eventBus      *events.EventBus
	agentRegistry agents.AgentRegistry
	startTime     time.Time
	logger        *logging.Logger
	currentEvent  *events.Event // Store current event context for correlation
}

// NewDeploymentAgent creates a new DeploymentAgent that auto-registers
func NewDeploymentAgent(graph *graph.GlobalGraph, aiProvider ai.AIProvider, env string, eventBus *events.EventBus, agentRegistry agents.AgentRegistry) (agents.AgentInterface, error) {
	service := NewDeploymentService(graph, aiProvider)
	agent := &DeploymentAgent{
		service:       service,
		agentID:       "deployment-agent",
		env:           env,
		eventBus:      eventBus,
		agentRegistry: agentRegistry,
		startTime:     time.Now(),
		logger:        logging.GetLogger().ForComponent("deployment-agent"),
	}

	// Auto-register with the agent registry
	if agentRegistry != nil {
		ctx := context.Background()
		if err := agentRegistry.RegisterAgent(ctx, agent); err != nil {
			agent.logger.Error("❌ Failed to auto-register deployment agent: %v", err)
			return nil, fmt.Errorf("failed to auto-register deployment agent: %w", err)
		}
		agent.logger.Info("✅ DeploymentAgent auto-registered successfully")
	}

	// Subscribe to events
	if eventBus != nil {
		if err := agent.subscribeToEvents(); err != nil {
			agent.logger.Error("❌ Failed to subscribe to events: %v", err)
			return nil, fmt.Errorf("failed to subscribe to events: %w", err)
		}
		agent.logger.Info("✅ DeploymentAgent subscribed to events")
	}

	return agent, nil
}

// GetID returns the agent's unique identifier
func (a *DeploymentAgent) GetID() string {
	return a.agentID
}

// GetStatus returns the current agent status
func (a *DeploymentAgent) GetStatus() agents.AgentStatus {
	return agents.AgentStatus{
		ID:           a.agentID,
		Type:         "deployment",
		Status:       "running",
		LastActivity: time.Now(),
		LoadFactor:   0.5,
		Version:      "1.0.0",
		Metadata: map[string]interface{}{
			"uptime":      time.Since(a.startTime).String(),
			"ai_provider": a.getAIProviderName(),
			"environment": a.env,
		},
	}
}

// GetCapabilities returns the agent's capabilities
func (a *DeploymentAgent) GetCapabilities() []agents.AgentCapability {
	return []agents.AgentCapability{
		{
			Name:        "deployment_orchestration",
			Description: "AI-native deployment orchestration with intelligent planning and execution",
			Intents:     []string{"deploy application", "execute deployment", "start deployment", "run deployment"},
			InputTypes:  []string{"application_name", "environment", "deployment_plan"},
			OutputTypes: []string{"deployment_result", "deployment_status", "error_details"},
			RoutingKeys: []string{"deployment.request", "deployment.execute", "deployment.orchestrate"},
			Version:     "1.0.0",
		},
		{
			Name:        "deployment_planning",
			Description: "Creates AI-enhanced deployment plans with optimization and risk analysis",
			Intents:     []string{"plan deployment", "create deployment plan", "generate deployment strategy"},
			InputTypes:  []string{"application_name", "environment", "constraints"},
			OutputTypes: []string{"deployment_plan", "optimization_recommendations", "risk_analysis"},
			RoutingKeys: []string{"deployment.plan", "deployment.strategy"},
			Version:     "1.0.0",
		},
		{
			Name:        "deployment_status_updates",
			Description: "Receives status updates and notifications from other agents",
			Intents:     []string{"deployment status update", "release status", "deployment notification"},
			InputTypes:  []string{"status_update", "correlation_id", "release_status"},
			OutputTypes: []string{"acknowledgment", "status_confirmation"},
			RoutingKeys: []string{"deployment.status", "deployment.update", "deployment.notification"},
			Version:     "1.0.0",
		},
	}
}

// Start initializes the agent
func (a *DeploymentAgent) Start(ctx context.Context) error {
	a.logger.Info("🚀 Starting DeploymentAgent...")
	return nil
}

// Stop shuts down the agent
func (a *DeploymentAgent) Stop(ctx context.Context) error {
	a.logger.Info("🛑 Stopping DeploymentAgent...")
	return nil
}

// Health returns the agent's health status
func (a *DeploymentAgent) Health() agents.HealthStatus {
	return agents.HealthStatus{
		Healthy: true,
		Status:  "healthy",
		Message: "DeploymentAgent is operational",
	}
}

// ProcessEvent handles incoming events for deployment operations
func (a *DeploymentAgent) ProcessEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	// Store current event for correlation context
	a.currentEvent = event

	a.logger.Info("🎯 Processing event: %s", event.Subject)

	// Extract intent from event payload
	intent, ok := event.Payload["intent"].(string)
	if !ok || intent == "" {
		return a.createErrorResponse(event, "intent field required in payload"), nil
	}

	// Route based on intent
	switch {
	case strings.Contains(intent, "deploy"):
		return a.handleDeployApplication(ctx, event)
	case strings.Contains(intent, "plan"):
		return a.handleCreateDeploymentPlan(ctx, event)
	case strings.Contains(intent, "status"):
		return a.handleGetDeploymentStatus(ctx, event)
	default:
		return a.handleGenericQuestion(ctx, event, intent)
	}
}

// handleDeployApplication processes application deployment requests using AI-native parsing
func (a *DeploymentAgent) handleDeployApplication(ctx context.Context, event *events.Event) (*events.Event, error) {
	a.logger.Info("🔍 Payload keys: %v", getKeys(event.Payload))

	// Extract user message from context or top-level payload
	var userMessage string
	if msg, ok := event.Payload["user_message"].(string); ok {
		userMessage = msg
		a.logger.Info("🔍 Found user_message at top level: %s", userMessage)
	} else if contextData, ok := event.Payload["context"].(map[string]interface{}); ok {
		a.logger.Info("🔍 Context keys: %v", getKeys(contextData))
		if msg, ok := contextData["user_message"].(string); ok {
			userMessage = msg
			a.logger.Info("🔍 Found user_message in context: %s", userMessage)
		}
	}

	if userMessage == "" {
		return a.createErrorResponse(event, "user_message required for AI-native deployment processing"), nil
	}

	a.logger.Info("🤖 AI-parsing deployment request: %s", userMessage)

	// Step 1: Use AI to extract application and environment from natural language
	appName, environment, err := a.parseDeploymentRequest(ctx, userMessage)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("failed to parse deployment request: %v", err)), nil
	}

	a.logger.Info("🎯 Resolved deployment: %s -> %s", appName, environment)

	// Step 2: Validate application and environment existence
	if err := a.validateApplicationExists(ctx, appName); err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("application validation failed: %v", err)), nil
	}
	if err := a.validateEnvironmentExists(ctx, environment); err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("environment validation failed: %v", err)), nil
	}

	// Step 3: Build rich AI-native deployment plan with Release node
	deploymentPlan, err := a.buildDeploymentPlan(ctx, appName, environment)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("failed to build deployment plan: %v", err)), nil
	}

	// Step 4: Create conversational response explaining the deployment plan
	planExplanation, err := a.generateDeploymentPlanExplanation(ctx, deploymentPlan)
	if err != nil {
		planExplanation = "Basic deployment plan generated successfully"
	}

	// Step 5: Execute deployment using the service
	result, err := a.service.DeployApplication(ctx, appName, environment)
	if err != nil {
		// Rich error response with context
		errorMsg := fmt.Sprintf("%s\n\n❌ Deployment execution failed: %v", planExplanation, err)
		return a.createErrorResponse(event, errorMsg), nil
	}

	// Step 6: Emit success event
	if a.eventBus != nil {
		a.eventBus.Emit(events.EventTypeNotify, "deployment-agent", "deployment.completed", map[string]interface{}{
			"application_name":  appName,
			"environment":       environment,
			"status":            result.Status,
			"services_deployed": result.Summary.Deployed,
			"deployment_id":     result.DeploymentID,
			"release_id":        deploymentPlan["release_id"],
		})
	}

	// Step 7: Rich success response
	releaseID, _ := deploymentPlan["release_id"].(string)
	successMsg := fmt.Sprintf("%s\n\n✅ Deployment completed successfully! Release %s is now live in %s.",
		planExplanation, releaseID, environment)

	return a.createResponse(successMsg, map[string]interface{}{
		"status":            "success",
		"operation":         "deploy",
		"application_name":  appName,
		"environment":       environment,
		"release_id":        releaseID,
		"deployment_result": result,
		"explanation":       planExplanation,
	}, event), nil
}

// handleCreateDeploymentPlan processes deployment planning requests
func (a *DeploymentAgent) handleCreateDeploymentPlan(ctx context.Context, event *events.Event) (*events.Event, error) {
	appName, ok := event.Payload["application_name"].(string)
	if !ok {
		return a.createErrorResponse(event, "application_name required for deployment planning"), nil
	}

	a.logger.Info("📋 Creating deployment plan for %s", appName)

	// Validate application existence before planning
	if err := a.validateApplicationExists(ctx, appName); err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("application validation failed: %v", err)), nil
	}

	plan, err := a.service.GenerateDeploymentPlan(ctx, appName)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("deployment planning failed: %v", err)), nil
	}

	return a.createResponse("Deployment plan created", map[string]interface{}{
		"status":           "success",
		"operation":        "plan",
		"application_name": appName,
		"deployment_plan":  plan,
	}, event), nil
}

// handleGetDeploymentStatus processes deployment status requests
func (a *DeploymentAgent) handleGetDeploymentStatus(ctx context.Context, event *events.Event) (*events.Event, error) {
	appName, ok := event.Payload["application_name"].(string)
	if !ok {
		return a.createErrorResponse(event, "application_name required for status query"), nil
	}

	environment, ok := event.Payload["environment"].(string)
	if !ok {
		return a.createErrorResponse(event, "environment required for status query"), nil
	}

	// Validate application and environment existence before status query
	if err := a.validateApplicationExists(ctx, appName); err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("application validation failed: %v", err)), nil
	}
	if err := a.validateEnvironmentExists(ctx, environment); err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("environment validation failed: %v", err)), nil
	}

	status, err := a.service.GetDeploymentStatus(appName, environment)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("failed to get deployment status: %v", err)), nil
	}

	return a.createResponse("Deployment status retrieved", map[string]interface{}{
		"status":            "success",
		"operation":         "status",
		"application_name":  appName,
		"environment":       environment,
		"deployment_status": status,
	}, event), nil
}

// handleGenericQuestion processes generic deployment-related questions
func (a *DeploymentAgent) handleGenericQuestion(ctx context.Context, event *events.Event, intent string) (*events.Event, error) {
	a.logger.Info("🤔 Handling generic deployment question: %s", intent)

	response := fmt.Sprintf("I'm the DeploymentAgent. I can help with deployment operations like:\n- Deploying applications\n- Creating deployment plans\n- Checking deployment status\n\nYour question: %s", intent)

	return a.createResponse("DeploymentAgent capabilities", map[string]interface{}{
		"status":       "info",
		"capabilities": a.GetCapabilities(),
		"response":     response,
	}, event), nil
}

// Event subscription setup
func (a *DeploymentAgent) subscribeToEvents() error {
	// Subscribe to specific routing keys for deployment events
	routingKeys := []string{
		"deployment.request",
		"deployment.execute",
		"deployment.orchestrate",
		"deployment.plan",
		"deployment.strategy",
		"deployment.status",
		"deployment.update",
		"deployment.notification",
	}

	for _, key := range routingKeys {
		a.eventBus.SubscribeToRoutingKey(key, a.handleIncomingEvent)
	}

	return nil
}

// handleIncomingEvent processes events from the event bus
func (a *DeploymentAgent) handleIncomingEvent(event events.Event) error {
	a.logger.Info("📨 Received event: %s from %s", event.Subject, event.Source)

	// Process the event synchronously to handle errors properly
	ctx := context.Background()
	response, err := a.ProcessEvent(ctx, &event)
	if err != nil {
		a.logger.Error("❌ Failed to process event: %v", err)
		// Still send error response for correlation
		if a.eventBus != nil {
			errorResponse := a.createErrorResponse(&event, fmt.Sprintf("Processing failed: %v", err))
			a.eventBus.EmitEvent(*errorResponse)
		}
		return err
	}

	if response != nil {
		// Check if this is an error response for better logging
		if status, ok := response.Payload["status"].(string); ok && status == "error" {
			a.logger.Info("❌ Processed event with error, sending error response")
		} else {
			a.logger.Info("✅ Processed event successfully, sending response")
		}
		// Emit complete response event to preserve correlation_id
		if a.eventBus != nil {
			a.eventBus.EmitEvent(*response)
		}
	}

	return nil
}

// Agent discovery and communication methods

// discoverAgentsByIntent finds agents that can handle a specific intent
func (a *DeploymentAgent) discoverAgentsByIntent(ctx context.Context, intent string) ([]agents.AgentStatus, error) {
	if a.agentRegistry == nil {
		return nil, fmt.Errorf("no agent registry available")
	}

	// Get all registered agents
	allAgents, err := a.agentRegistry.ListAllAgents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	var matchingAgents []agents.AgentStatus
	for _, agent := range allAgents {
		// Check if any of the agent's capabilities match our intent
		capabilities := a.getAgentCapabilities(ctx, agent.ID)
		for _, capability := range capabilities {
			for _, supportedIntent := range capability.Intents {
				if a.intentMatches(intent, supportedIntent) {
					matchingAgents = append(matchingAgents, agent)
					break
				}
			}
		}
	}

	return matchingAgents, nil
}

// getAgentCapabilities gets capabilities for an agent
func (a *DeploymentAgent) getAgentCapabilities(ctx context.Context, agentID string) []agents.AgentCapability {
	if a.agentRegistry == nil {
		return nil
	}

	capabilities, err := a.agentRegistry.GetAvailableCapabilities(ctx)
	if err != nil {
		return nil
	}

	var agentCapabilities []agents.AgentCapability
	for _, capability := range capabilities {
		// For now, assume all capabilities belong to all agents
		// In a real implementation, we'd filter by agent ID
		agentCapabilities = append(agentCapabilities, capability)
	}

	return agentCapabilities
}

// intentMatches checks if a user intent matches a supported intent pattern
func (a *DeploymentAgent) intentMatches(userIntent, supportedIntent string) bool {
	// Simple keyword matching - could be enhanced with AI/NLP
	userWords := strings.Fields(strings.ToLower(userIntent))
	supportedWords := strings.Fields(strings.ToLower(supportedIntent))

	// Check if all supported words are found in user intent
	for _, supportedWord := range supportedWords {
		found := false
		for _, userWord := range userWords {
			if strings.Contains(userWord, supportedWord) || strings.Contains(supportedWord, userWord) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// requestAgentAction sends a request to another agent and waits for response
func (a *DeploymentAgent) requestAgentAction(ctx context.Context, intent string, payload map[string]interface{}) (*events.Event, error) {
	// Find agents that can handle this intent
	agents, err := a.discoverAgentsByIntent(ctx, intent)
	if err != nil {
		return nil, fmt.Errorf("failed to discover agents for intent '%s': %w", intent, err)
	}

	if len(agents) == 0 {
		return nil, fmt.Errorf("no agents found for intent: %s", intent)
	}

	// Use the first available agent
	targetAgent := agents[0]
	a.logger.Info("🎯 Sending request to agent: %s for intent: %s", targetAgent.ID, intent)

	// Find appropriate routing key for this intent
	routingKey := a.findRoutingKeyForIntent(targetAgent, intent)
	if routingKey == "" {
		routingKey = "default.request" // fallback
	}

	// Create request event
	requestEvent := &events.Event{
		Type:    events.EventTypeRequest,
		Source:  a.agentID,
		Subject: fmt.Sprintf("Request to %s: %s", targetAgent.ID, intent),
		Payload: payload,
	}

	// Add intent to payload
	requestEvent.Payload["intent"] = intent

	// Send the request via event bus
	if a.eventBus != nil {
		a.eventBus.Emit(requestEvent.Type, requestEvent.Source, routingKey, requestEvent.Payload)
	}

	// For now, return a simple response - in a full implementation,
	// we'd wait for a correlated response
	return &events.Event{
		Type:    events.EventTypeResponse,
		Source:  targetAgent.ID,
		Subject: "Agent communication response",
		Payload: map[string]interface{}{
			"status":       "request_sent",
			"target_agent": targetAgent.ID,
			"intent":       intent,
		},
	}, nil
}

// findRoutingKeyForIntent finds the appropriate routing key for an intent
func (a *DeploymentAgent) findRoutingKeyForIntent(agentStatus agents.AgentStatus, intent string) string {
	// Get agent capabilities to find routing keys
	capabilities := a.getAgentCapabilities(context.Background(), agentStatus.ID)
	for _, capability := range capabilities {
		// Check if this capability matches the intent
		for _, supportedIntent := range capability.Intents {
			if a.intentMatches(intent, supportedIntent) {
				// Return the first routing key for this capability
				if len(capability.RoutingKeys) > 0 {
					return capability.RoutingKeys[0]
				}
			}
		}
	}
	return ""
}

// Helper methods

// getKeys returns the keys from a map[string]interface{}
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// getAIProviderName returns the name of the AI provider
func (a *DeploymentAgent) getAIProviderName() string {
	if a.service != nil && a.service.aiProvider != nil {
		// Try to get provider info
		if providerInfo := a.service.aiProvider.GetProviderInfo(); providerInfo != nil {
			return providerInfo.Name
		}
	}
	return "unknown"
}

// createErrorResponse creates a standardized error response event
func (a *DeploymentAgent) createErrorResponse(originalEvent *events.Event, errorMessage string) *events.Event {
	response := &events.Event{
		Type:    events.EventTypeResponse,
		Source:  a.agentID,
		Subject: "Deployment operation failed",
		Payload: map[string]interface{}{
			"status":  "error",
			"error":   errorMessage,
			"context": "deployment-agent",
		},
		Timestamp: time.Now().Unix(),
		ID:        fmt.Sprintf("dep-error-%d", time.Now().UnixNano()),
	}

	// Copy correlation fields from original event
	if originalEvent != nil && originalEvent.Payload != nil {
		if correlationID, ok := originalEvent.Payload["correlation_id"]; ok {
			response.Payload["correlation_id"] = correlationID
		}
		if requestID, ok := originalEvent.Payload["request_id"]; ok {
			response.Payload["request_id"] = requestID
		}
	}

	return response
}

// createResponse creates a standardized success response event
func (a *DeploymentAgent) createResponse(subject string, payload map[string]interface{}, originalEvent *events.Event) *events.Event {
	response := &events.Event{
		Type:      events.EventTypeResponse,
		Source:    a.agentID,
		Subject:   subject,
		Payload:   payload,
		Timestamp: time.Now().Unix(),
		ID:        fmt.Sprintf("dep-resp-%d", time.Now().UnixNano()),
	}

	// Copy correlation fields from original event
	if originalEvent != nil && originalEvent.Payload != nil {
		if correlationID, ok := originalEvent.Payload["correlation_id"]; ok {
			response.Payload["correlation_id"] = correlationID
		}
		if requestID, ok := originalEvent.Payload["request_id"]; ok {
			response.Payload["request_id"] = requestID
		}
	}

	return response
}

// parseDeploymentRequest uses AI to extract application and environment from natural language
func (a *DeploymentAgent) parseDeploymentRequest(ctx context.Context, userMessage string) (string, string, error) {
	if a.service == nil || a.service.aiProvider == nil {
		// Fallback parsing - try to extract from simple patterns
		return a.parseDeploymentRequestFallback(userMessage)
	}

	// Use AI to parse the deployment request
	systemPrompt := `You are a deployment request parser. Extract the application name and environment from the user's message.
Return ONLY a JSON object with "application" and "environment" fields.
If environment is not specified, use "development" as default.

Examples:
Input: "deploy myapp to production"
Output: {"application": "myapp", "environment": "production"}

Input: "I want to deploy the user-service"
Output: {"application": "user-service", "environment": "development"}
`

	userPrompt := fmt.Sprintf("Parse this deployment request: %s", userMessage)

	response, err := a.service.aiProvider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		a.logger.Warn("AI parsing failed, using fallback: %v", err)
		return a.parseDeploymentRequestFallback(userMessage)
	}

	// Parse the AI response
	var parsed struct {
		Application string `json:"application"`
		Environment string `json:"environment"`
	}

	if err := parseJSONResponse(response, &parsed); err != nil {
		a.logger.Warn("AI response parsing failed, using fallback: %v", err)
		return a.parseDeploymentRequestFallback(userMessage)
	}

	if parsed.Application == "" {
		return "", "", fmt.Errorf("could not extract application name from message")
	}

	if parsed.Environment == "" {
		parsed.Environment = "development"
	}

	return parsed.Application, parsed.Environment, nil
}

// parseDeploymentRequestFallback provides simple pattern-based parsing when AI is unavailable
func (a *DeploymentAgent) parseDeploymentRequestFallback(userMessage string) (string, string, error) {
	lower := strings.ToLower(userMessage)

	// Try to extract application name after "deploy"
	var appName string
	if strings.Contains(lower, "deploy ") {
		parts := strings.Split(lower, "deploy ")
		if len(parts) > 1 {
			remaining := strings.TrimSpace(parts[1])
			words := strings.Fields(remaining)
			if len(words) > 0 {
				appName = words[0]
			}
		}
	}

	// Try to extract environment
	environment := "development" // default
	if strings.Contains(lower, "production") || strings.Contains(lower, "prod") {
		environment = "production"
	} else if strings.Contains(lower, "staging") {
		environment = "staging"
	} else if strings.Contains(lower, "test") {
		environment = "test"
	}

	if appName == "" {
		return "", "", fmt.Errorf("could not extract application name from message: %s", userMessage)
	}

	return appName, environment, nil
}

// buildDeploymentPlan creates a comprehensive deployment plan and coordinates with ReleaseAgent
func (a *DeploymentAgent) buildDeploymentPlan(ctx context.Context, appName, environment string) (map[string]interface{}, error) {
	a.logger.Info("🏗️ Building deployment plan for %s in %s", appName, environment)

	// Step 1: Create a Release via ReleaseAgent (agent-to-agent communication)
	releaseID, err := a.createReleaseViaAgent(ctx, appName, environment, a.currentEvent)
	if err != nil {
		a.logger.Error("Failed to create release via ReleaseAgent: %v", err)
		return nil, fmt.Errorf("failed to create release: %w", err)
	}

	a.logger.Info("✅ Created release %s via ReleaseAgent", releaseID)

	// Step 2: Generate deployment plan using AI
	deploymentPlan, err := a.service.GenerateDeploymentPlan(ctx, appName)
	if err != nil {
		a.logger.Error("Failed to generate deployment plan: %v", err)
		return nil, fmt.Errorf("failed to generate deployment plan: %w", err)
	}

	// Step 3: Create comprehensive plan response
	plan := map[string]interface{}{
		"application":      appName,
		"environment":      environment,
		"release_id":       releaseID,
		"deployment_steps": deploymentPlan,
		"created_at":       time.Now().UTC(),
		"agent":            a.agentID,
		"status":           "planned",
	}

	return plan, nil
}

// createReleaseViaAgent communicates with ReleaseAgent to create a release
func (a *DeploymentAgent) createReleaseViaAgent(ctx context.Context, appName, environment string, originalEvent *events.Event) (string, error) {
	a.logger.Info("📞 Requesting ReleaseAgent to create release for %s", appName)

	// Discover ReleaseAgent by intent
	intent := "create release"
	agents, err := a.discoverAgentsByIntent(ctx, intent)
	if err != nil {
		return "", fmt.Errorf("failed to discover release agent: %w", err)
	}

	if len(agents) == 0 {
		return "", fmt.Errorf("no release agent available to handle intent: %s", intent)
	}

	// Use the first available ReleaseAgent
	releaseAgent := agents[0]
	a.logger.Info("🎯 Found ReleaseAgent: %s", releaseAgent.ID)

	// Create request payload with separate correlation context for internal agent communication
	payload := map[string]interface{}{
		"application":  appName,
		"environment":  environment,
		"requested_by": a.agentID,
		"user_message": fmt.Sprintf("create release for %s in %s", appName, environment),
	}

	// Use a different correlation_id for ReleaseAgent communication to avoid interfering with V3Agent response
	internalCorrelationID := fmt.Sprintf("dep-to-rel-%d", time.Now().UnixNano())
	payload["correlation_id"] = internalCorrelationID

	// Also preserve original request info for tracking
	if originalEvent != nil && originalEvent.Payload != nil {
		if requestID, ok := originalEvent.Payload["request_id"]; ok {
			payload["request_id"] = requestID
		}
		if sourceAgent, ok := originalEvent.Payload["source_agent"]; ok {
			payload["source_agent"] = sourceAgent
		}
	}

	// Send request to ReleaseAgent
	response, err := a.requestAgentAction(ctx, intent, payload)
	if err != nil {
		return "", fmt.Errorf("failed to request release creation: %w", err)
	}

	// Extract release ID from response
	if response.Payload == nil {
		return "", fmt.Errorf("empty response from release agent")
	}

	releaseID, ok := response.Payload["release_id"].(string)
	if !ok {
		// For now, create a mock release ID since this is just for demonstration
		releaseID = fmt.Sprintf("rel-%s-%d", appName, time.Now().Unix())
		a.logger.Info("Using mock release ID: %s", releaseID)
	}

	return releaseID, nil
}

// generateDeploymentPlanExplanation creates a human-readable explanation of the deployment plan
func (a *DeploymentAgent) generateDeploymentPlanExplanation(ctx context.Context, plan map[string]interface{}) (string, error) {
	if a.service == nil || a.service.aiProvider == nil {
		return a.generateBasicExplanation(plan), nil
	}

	// Use AI to generate a comprehensive explanation
	systemPrompt := `You are a deployment expert. Create a clear, human-readable explanation of the deployment plan.
Focus on what will happen, in what order, and any important considerations.
Keep it concise but informative.`

	planJSON, _ := json.Marshal(plan)
	userPrompt := fmt.Sprintf("Explain this deployment plan: %s", string(planJSON))

	response, err := a.service.aiProvider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		a.logger.Warn("AI explanation generation failed, using basic explanation: %v", err)
		return a.generateBasicExplanation(plan), nil
	}

	return response, nil
}

// generateBasicExplanation creates a simple explanation when AI is not available
func (a *DeploymentAgent) generateBasicExplanation(plan map[string]interface{}) string {
	appName, _ := plan["application"].(string)
	environment, _ := plan["environment"].(string)
	releaseID, _ := plan["release_id"].(string)

	explanation := fmt.Sprintf("Deployment plan for %s to %s environment", appName, environment)
	if releaseID != "" {
		explanation += fmt.Sprintf(" (Release: %s)", releaseID)
	}

	if steps, ok := plan["deployment_steps"].([]interface{}); ok {
		explanation += fmt.Sprintf("\nPlanned steps: %d deployment actions", len(steps))
	}

	return explanation
}

// parseJSONResponse parses a JSON response string into the given target
func parseJSONResponse(response string, target interface{}) error {
	// Clean up the response - remove any markdown code blocks
	cleaned := strings.TrimSpace(response)
	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
	}
	if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
	}
	if strings.HasSuffix(cleaned, "```") {
		cleaned = strings.TrimSuffix(cleaned, "```")
	}
	cleaned = strings.TrimSpace(cleaned)

	return json.Unmarshal([]byte(cleaned), target)
}

// validateApplicationExists checks if an application exists in the graph
// This is deployment domain business logic - the DeploymentAgent is responsible 
// for validating that applications exist before attempting deployment operations
func (a *DeploymentAgent) validateApplicationExists(ctx context.Context, appName string) error {
	if appName == "" {
		return fmt.Errorf("application name cannot be empty")
	}

	// Get all nodes from the graph
	nodes, err := a.service.graph.Nodes()
	if err != nil {
		a.logger.Error("❌ Failed to query nodes from graph: %v", err)
		return fmt.Errorf("failed to validate application existence: %w", err)
	}

	// Check if any application has the matching name
	for _, node := range nodes {
		if node.Kind == "application" {
			if nodeName, ok := node.Metadata["name"].(string); ok && nodeName == appName {
				a.logger.Info("✅ Application '%s' validated - exists in graph", appName)
				return nil
			}
		}
	}

	// Application not found - this is a domain-specific validation error
	a.logger.Warn("⚠️ Application '%s' not found in graph", appName)
	return fmt.Errorf("application '%s' does not exist. Please create the application first before attempting deployment", appName)
}

// validateEnvironmentExists checks if an environment exists in the graph
// This is deployment domain business logic for environment validation
func (a *DeploymentAgent) validateEnvironmentExists(ctx context.Context, envName string) error {
	if envName == "" {
		return fmt.Errorf("environment name cannot be empty")
	}

	// Get all nodes from the graph
	nodes, err := a.service.graph.Nodes()
	if err != nil {
		a.logger.Error("❌ Failed to query nodes from graph: %v", err)
		return fmt.Errorf("failed to validate environment existence: %w", err)
	}

	// Check if any environment has the matching name
	for _, node := range nodes {
		if node.Kind == "environment" {
			if nodeName, ok := node.Metadata["name"].(string); ok && nodeName == envName {
				a.logger.Info("✅ Environment '%s' validated - exists in graph", envName)
				return nil
			}
		}
	}

	// Environment not found - this is a domain-specific validation error
	a.logger.Warn("⚠️ Environment '%s' not found in graph", envName)
	return fmt.Errorf("environment '%s' does not exist. Please create the environment first before attempting deployment", envName)
}
