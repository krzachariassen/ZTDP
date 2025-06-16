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
// This agent handles complex deployment workflows, AI-enhanced planning, and failure analysis
type DeploymentAgent struct {
	service       *Service
	agentID       string
	env           string
	eventBus      *events.EventBus // Use actual EventBus instead of interface
	agentRegistry agents.AgentRegistry
	startTime     time.Time
	logger        *logging.Logger
}

// NewDeploymentAgent creates a new DeploymentAgent that auto-registers with the agent registry
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
				agent.logger.Info("✅ DeploymentAgent subscribed to routing key: %s", routingKey)
			}
		}
	} else {
		agent.logger.Warn("⚠️ No event bus provided - agent will not receive events")
	}

	return agent, nil
}

// GetID returns the agent's unique identifier
func (a *DeploymentAgent) GetID() string {
	return a.agentID
}

// GetStatus returns current agent status information
func (a *DeploymentAgent) GetStatus() agents.AgentStatus {
	return agents.AgentStatus{
		ID:           a.agentID,
		Type:         "deployment",
		Status:       "running",
		LastActivity: time.Now(),
		LoadFactor:   0.3, // Deployments can be resource intensive
		Version:      "1.0.0",
		Metadata: map[string]interface{}{
			"environment":     a.env,
			"ai_capabilities": a.service.HasAICapabilities(),
			"ai_provider":     a.getAIProviderName(),
			"operations":      []string{"deploy", "plan", "optimize", "troubleshoot", "predict", "rollback"},
		},
	}
}

// GetCapabilities returns the agent's capabilities
func (a *DeploymentAgent) GetCapabilities() []agents.AgentCapability {
	return []agents.AgentCapability{
		{
			Name:        "deployment_orchestration",
			Description: "Orchestrates complex deployment workflows with AI-enhanced decision making",
			Intents:     []string{"deploy application", "execute deployment", "orchestrate deployment"},
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
			Name:        "deployment_optimization",
			Description: "Optimizes deployment plans using AI analysis for better performance and reliability",
			Intents:     []string{"optimize deployment", "improve deployment plan", "enhance deployment strategy"},
			InputTypes:  []string{"deployment_plan", "performance_metrics", "constraints"},
			OutputTypes: []string{"optimization_recommendations", "performance_improvements"},
			RoutingKeys: []string{"deployment.optimize", "deployment.enhance"},
			Version:     "1.0.0",
		},
		{
			Name:        "deployment_troubleshooting",
			Description: "AI-powered troubleshooting and failure analysis for deployment issues",
			Intents:     []string{"troubleshoot deployment", "analyze deployment failure", "diagnose deployment issues"},
			InputTypes:  []string{"incident_id", "error_description", "symptoms", "logs"},
			OutputTypes: []string{"troubleshooting_response", "root_cause_analysis", "resolution_steps"},
			RoutingKeys: []string{"deployment.troubleshoot", "deployment.diagnose"},
			Version:     "1.0.0",
		},
		{
			Name:        "impact_prediction",
			Description: "Predicts deployment impact and potential risks using AI analysis",
			Intents:     []string{"predict impact", "analyze deployment risks", "assess deployment effects"},
			InputTypes:  []string{"proposed_changes", "environment", "current_state"},
			OutputTypes: []string{"impact_prediction", "risk_assessment", "mitigation_strategies"},
			Version:     "1.0.0",
		},
	}
}

// Start initializes the agent
func (a *DeploymentAgent) Start(ctx context.Context) error {
	a.logger.Info("🤖 DeploymentAgent starting up")
	return nil
}

// Stop gracefully shuts down the agent
func (a *DeploymentAgent) Stop(ctx context.Context) error {
	a.logger.Info("🤖 DeploymentAgent shutting down")
	return nil
}

// Health returns the agent's health status
func (a *DeploymentAgent) Health() agents.HealthStatus {
	aiHealthy := a.service.HasAICapabilities()
	status := "healthy"
	if !aiHealthy {
		status = "degraded"
	}

	return agents.HealthStatus{
		Healthy: true, // Agent can work without AI, just degraded
		Status:  status,
		Message: "Deployment agent is operational",
		Checks: map[string]interface{}{
			"graph_connection":  "connected",
			"event_bus":         "connected",
			"ai_provider":       aiHealthy,
			"deployment_engine": "ready",
		},
		CheckedAt: time.Now(),
	}
}

// ProcessEvent handles incoming events for the deployment agent
func (a *DeploymentAgent) ProcessEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	// Validate event has required intent field
	intent, ok := event.Payload["intent"].(string)
	if !ok {
		return nil, fmt.Errorf("deployment agent requires 'intent' field in payload")
	}

	a.logger.Info("🤖 Processing deployment event with intent: %s", intent)

	// Route based on intent
	switch intent {
	case "deployment_orchestration", "deploy_application", "deploy application":
		return a.handleDeployApplication(ctx, event)
	case "deployment_planning", "create_deployment_plan":
		return a.handleCreateDeploymentPlan(ctx, event)
	case "deployment_optimization", "optimize_deployment_plan":
		return a.handleOptimizeDeploymentPlan(ctx, event)
	case "deployment_troubleshooting", "troubleshoot_deployment":
		return a.handleTroubleshootDeployment(ctx, event)
	case "impact_prediction", "predict_deployment_impact":
		return a.handlePredictDeploymentImpact(ctx, event)
	case "deployment_status", "get_deployment_status":
		return a.handleGetDeploymentStatus(ctx, event)
	default:
		return a.handleGenericQuestion(ctx, event, intent)
	}
}

// handleDeployApplication processes application deployment requests using AI-native parsing
func (a *DeploymentAgent) handleDeployApplication(ctx context.Context, event *events.Event) (*events.Event, error) {
	a.logger.Info("🔍 Payload keys: %v", getKeys(event.Payload))
	
	// AI-NATIVE APPROACH: Extract user message from context or top-level payload
	var userMessage string
	
	// Try to get user_message from top level first
	if msg, ok := event.Payload["user_message"].(string); ok {
		userMessage = msg
		a.logger.Info("🔍 Found user_message at top level: %s", userMessage)
	} else if contextData, ok := event.Payload["context"].(map[string]interface{}); ok {
		// Try to get user_message from nested context
		a.logger.Info("🔍 Context keys: %v", getKeys(contextData))
		if msg, ok := contextData["user_message"].(string); ok {
			userMessage = msg
			a.logger.Info("🔍 Found user_message in context: %s", userMessage)
		}
	}
	
	if userMessage == "" {
		return a.createErrorResponse(event, "user_message required for AI-native deployment processing")
	}

	a.logger.Info("🤖 AI-parsing deployment request: %s", userMessage)

	// Step 1: Use AI to extract application and environment from natural language
	appName, environment, err := a.parseDeploymentRequest(ctx, userMessage)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("failed to parse deployment request: %v", err))
	}

	a.logger.Info("🎯 Resolved deployment: %s -> %s", appName, environment)

	// TODO: Consult PolicyAgent for deployment policies before executing
	// This will be added in a future iteration to ensure compliance

	// Step 2: Execute deployment using the service
	result, err := a.service.DeployApplication(ctx, appName, environment)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("deployment failed: %v", err))
	}

	// Step 3: Emit success event
	if a.eventBus != nil {
		a.eventBus.Emit(events.EventTypeNotify, "deployment-agent", "deployment.completed", map[string]interface{}{
			"application_name":  appName,
			"environment":       environment,
			"status":            result.Status,
			"services_deployed": result.Summary.Deployed,
			"deployment_id":     result.DeploymentID,
		})
	}

	return a.createResponse("Application deployed successfully", map[string]interface{}{
		"status":            "success",
		"operation":         "deploy",
		"application_name":  appName,
		"environment":       environment,
		"deployment_result": result,
	}, event), nil
}

// handleCreateDeploymentPlan processes deployment planning requests
func (a *DeploymentAgent) handleCreateDeploymentPlan(ctx context.Context, event *events.Event) (*events.Event, error) {
	appName, ok := event.Payload["application_name"].(string)
	if !ok {
		return a.createErrorResponse(event, "application_name required for deployment planning")
	}

	a.logger.Info("📋 Creating deployment plan for %s", appName)

	plan, err := a.service.GenerateDeploymentPlan(ctx, appName)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("deployment planning failed: %v", err))
	}

	return a.createResponse("Deployment plan created", map[string]interface{}{
		"status":           "success",
		"operation":        "plan",
		"application_name": appName,
		"deployment_plan":  plan,
	}, event), nil
}

// handleOptimizeDeploymentPlan processes deployment optimization requests
func (a *DeploymentAgent) handleOptimizeDeploymentPlan(ctx context.Context, event *events.Event) (*events.Event, error) {
	appID, ok := event.Payload["application_id"].(string)
	if !ok {
		return a.createErrorResponse(event, "application_id required for optimization")
	}

	// Extract current plan from event payload
	planData, ok := event.Payload["current_plan"]
	if !ok {
		return a.createErrorResponse(event, "current_plan required for optimization")
	}

	// Convert to deployment steps
	var currentPlan []ai.DeploymentStep
	if err := a.convertToDeploymentSteps(planData, &currentPlan); err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("invalid deployment plan format: %v", err))
	}

	a.logger.Info("🔧 Optimizing deployment plan for %s", appID)

	recommendations, err := a.service.OptimizeDeploymentPlan(ctx, appID, currentPlan)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("deployment optimization failed: %v", err))
	}

	return &events.Event{
		Type:    events.EventTypeResponse,
		Source:  a.agentID,
		Subject: "Deployment plan optimized",
		Payload: map[string]interface{}{
			"status":          "success",
			"operation":       "optimize",
			"application_id":  appID,
			"recommendations": recommendations,
		},
	}, nil
}

// handleTroubleshootDeployment processes deployment troubleshooting requests
func (a *DeploymentAgent) handleTroubleshootDeployment(ctx context.Context, event *events.Event) (*events.Event, error) {
	incidentID, ok := event.Payload["incident_id"].(string)
	if !ok {
		return a.createErrorResponse(event, "incident_id required for troubleshooting")
	}

	description, ok := event.Payload["description"].(string)
	if !ok {
		return a.createErrorResponse(event, "description required for troubleshooting")
	}

	symptoms, _ := event.Payload["symptoms"].([]string)
	if symptoms == nil {
		symptoms = []string{}
	}

	a.logger.Info("🔍 Troubleshooting deployment incident %s", incidentID)

	response, err := a.service.TroubleshootDeployment(ctx, incidentID, description, symptoms)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("deployment troubleshooting failed: %v", err))
	}

	return &events.Event{
		Type:    events.EventTypeResponse,
		Source:  a.agentID,
		Subject: "Deployment troubleshooting completed",
		Payload: map[string]interface{}{
			"status":                   "success",
			"operation":                "troubleshoot",
			"incident_id":              incidentID,
			"troubleshooting_response": response,
		},
	}, nil
}

// handlePredictDeploymentImpact processes impact prediction requests
func (a *DeploymentAgent) handlePredictDeploymentImpact(ctx context.Context, event *events.Event) (*events.Event, error) {
	environment, ok := event.Payload["environment"].(string)
	if !ok {
		return a.createErrorResponse(event, "environment required for impact prediction")
	}

	changesData, ok := event.Payload["proposed_changes"]
	if !ok {
		return a.createErrorResponse(event, "proposed_changes required for impact prediction")
	}

	// Convert to proposed changes
	var changes []ai.ProposedChange
	if err := a.convertToProposedChanges(changesData, &changes); err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("invalid proposed changes format: %v", err))
	}

	a.logger.Info("📊 Predicting deployment impact for %d changes in %s", len(changes), environment)

	prediction, err := a.service.PredictDeploymentImpact(ctx, changes, environment)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("impact prediction failed: %v", err))
	}

	return &events.Event{
		Type:    events.EventTypeResponse,
		Source:  a.agentID,
		Subject: "Deployment impact predicted",
		Payload: map[string]interface{}{
			"status":            "success",
			"operation":         "predict_impact",
			"environment":       environment,
			"impact_prediction": prediction,
		},
	}, nil
}

// handleGetDeploymentStatus processes deployment status requests
func (a *DeploymentAgent) handleGetDeploymentStatus(ctx context.Context, event *events.Event) (*events.Event, error) {
	appName, ok := event.Payload["application_name"].(string)
	if !ok {
		return a.createErrorResponse(event, "application_name required for status query")
	}

	environment, ok := event.Payload["environment"].(string)
	if !ok {
		return a.createErrorResponse(event, "environment required for status query")
	}

	status, err := a.service.GetDeploymentStatus(appName, environment)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("failed to get deployment status: %v", err))
	}

	return &events.Event{
		Type:    events.EventTypeResponse,
		Source:  a.agentID,
		Subject: "Deployment status retrieved",
		Payload: map[string]interface{}{
			"status":            "success",
			"operation":         "get_status",
			"application_name":  appName,
			"environment":       environment,
			"deployment_status": status,
		},
	}, nil
}

// handleGenericQuestion processes general deployment-related questions
func (a *DeploymentAgent) handleGenericQuestion(ctx context.Context, event *events.Event, intent string) (*events.Event, error) {
	return &events.Event{
		Type:    events.EventTypeResponse,
		Source:  a.agentID,
		Subject: "Deployment agent response",
		Payload: map[string]interface{}{
			"status":  "processed",
			"agent":   "deployment",
			"intent":  intent,
			"message": fmt.Sprintf("Deployment agent received intent '%s'. Available operations: deploy, plan, optimize, troubleshoot, predict_impact, get_status", intent),
		},
	}, nil
}

// handleIncomingEvent is the EventBus handler that routes events to ProcessEvent
func (a *DeploymentAgent) handleIncomingEvent(event events.Event) error {
	a.logger.Info("📨 DeploymentAgent received event: %s from %s", event.Subject, event.Source)

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

	// If we got a response, emit it back
	if responseEvent != nil && a.eventBus != nil {
		a.logger.Info("📤 DeploymentAgent emitting response: %s with correlation_id: %v", responseEvent.Subject, responseEvent.Payload["correlation_id"])
		err = a.eventBus.Emit(responseEvent.Type, responseEvent.Source, responseEvent.Subject, responseEvent.Payload)
		if err != nil {
			a.logger.Error("❌ Failed to emit response event: %v", err)
		} else {
			a.logger.Info("✅ Successfully emitted response event")
		}
	} else {
		if responseEvent == nil {
			a.logger.Warn("⚠️ No response event generated from ProcessEvent")
		}
		if a.eventBus == nil {
			a.logger.Warn("⚠️ No event bus available to emit response")
		}
	}

	return nil
}

// Helper functions

func (a *DeploymentAgent) createErrorResponse(originalEvent *events.Event, errorMsg string) (*events.Event, error) {
	a.logger.Error("❌ Deployment operation failed: %s", errorMsg)
	return a.createResponse("Deployment operation failed", map[string]interface{}{
		"status":  "error",
		"agent":   "deployment",
		"message": errorMsg,
	}, originalEvent), nil
}

// createResponse creates a standardized response event with correlation preservation
func (a *DeploymentAgent) createResponse(subject string, payload map[string]interface{}, originalEvent *events.Event) *events.Event {
	a.logger.Info("🔧 createResponse called with originalEvent payload keys: %v", getKeys(originalEvent.Payload))
	
	// Preserve correlation_id from original request
	if correlationID, ok := originalEvent.Payload["correlation_id"]; ok {
		a.logger.Info("✅ Found correlation_id in original event: %v", correlationID)
		payload["correlation_id"] = correlationID
	} else {
		a.logger.Warn("❌ No correlation_id found in original event payload")
	}

	// Preserve request_id from original request
	if requestID, ok := originalEvent.Payload["request_id"]; ok {
		a.logger.Info("✅ Found request_id in original event: %v", requestID)
		payload["request_id"] = requestID
	} else {
		a.logger.Warn("❌ No request_id found in original event payload")
	}

	a.logger.Info("🔧 Final response payload keys: %v", getKeys(payload))

	return &events.Event{
		Type:    events.EventTypeResponse,
		Source:  a.agentID,
		Subject: subject,
		Payload: payload,
	}
}

// Helper function to get map keys for debugging
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (a *DeploymentAgent) getAIProviderName() string {
	if a.service.HasAICapabilities() {
		info := a.service.GetAIProviderInfo()
		if info != nil {
			return info.Name
		}
	}
	return "none"
}

func (a *DeploymentAgent) convertToDeploymentSteps(data interface{}, steps *[]ai.DeploymentStep) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal deployment steps: %v", err)
	}

	err = json.Unmarshal(jsonData, steps)
	if err != nil {
		return fmt.Errorf("failed to unmarshal to DeploymentStep array: %v", err)
	}

	return nil
}

func (a *DeploymentAgent) convertToProposedChanges(data interface{}, changes *[]ai.ProposedChange) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal proposed changes: %v", err)
	}

	err = json.Unmarshal(jsonData, changes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal to ProposedChange array: %v", err)
	}

	return nil
}

// parseDeploymentRequest uses AI to extract application and environment from natural language
func (a *DeploymentAgent) parseDeploymentRequest(ctx context.Context, userMessage string) (string, string, error) {
	// Step 1: Query graph to get available applications and environments
	applications, err := a.getAvailableApplications(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to query applications: %w", err)
	}

	environments, err := a.getAvailableEnvironments(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to query environments: %w", err)
	}

	// Step 2: Use AI to parse the user request and match against available options
	if a.service.HasAICapabilities() {
		return a.parseWithAI(ctx, userMessage, applications, environments)
	}

	// Fallback: Simple text matching for cases without AI
	return a.parseWithFallback(userMessage, applications, environments)
}

// getAvailableApplications queries the graph for all applications
func (a *DeploymentAgent) getAvailableApplications(ctx context.Context) ([]string, error) {
	// Access the graph through the service
	globalGraph := a.service.graph
	if globalGraph == nil {
		return nil, fmt.Errorf("graph not available")
	}

	// Get current graph state
	currentGraph, err := globalGraph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to load graph: %w", err)
	}

	// Filter nodes by kind "application"
	var applications []string
	for _, node := range currentGraph.Nodes {
		if node.Kind == "application" {
			if name, ok := node.Metadata["name"].(string); ok && name != "" {
				applications = append(applications, name)
			} else {
				// Fallback to node ID if no name metadata
				applications = append(applications, node.ID)
			}
		}
	}

	a.logger.Info("🔍 Found %d applications in graph: %v", len(applications), applications)
	return applications, nil
}

// getAvailableEnvironments queries the graph for all environments
func (a *DeploymentAgent) getAvailableEnvironments(ctx context.Context) ([]string, error) {
	// Access the graph through the service
	globalGraph := a.service.graph
	if globalGraph == nil {
		return nil, fmt.Errorf("graph not available")
	}

	// Get current graph state
	currentGraph, err := globalGraph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to load graph: %w", err)
	}

	// Filter nodes by kind "environment"
	var environments []string
	for _, node := range currentGraph.Nodes {
		if node.Kind == "environment" {
			if name, ok := node.Metadata["name"].(string); ok && name != "" {
				environments = append(environments, name)
			} else {
				// Fallback to node ID if no name metadata
				environments = append(environments, node.ID)
			}
		}
	}

	a.logger.Info("🔍 Found %d environments in graph: %v", len(environments), environments)
	return environments, nil
}

// parseWithAI uses AI to intelligently parse the deployment request
func (a *DeploymentAgent) parseWithAI(ctx context.Context, userMessage string, applications, environments []string) (string, string, error) {
	systemPrompt := `You are an intelligent deployment parser. Given a user's deployment request and lists of available applications and environments, extract the specific application name and environment name.

IMPORTANT: You must respond with ONLY a JSON object in this exact format:
{
  "application": "exact_application_name",
  "environment": "exact_environment_name"
}

Rules:
- Match user terms to the closest available application and environment names
- For environments: "prod"/"production" usually maps to "prod", "dev"/"development" to "dev", "staging"/"stage" to "staging"
- For applications: Use fuzzy matching - "checkout" might match "checkout-service" or "checkout-app"
- If no good match exists, use your best judgment based on common patterns
- Always return valid JSON with exactly these two fields`

	userPrompt := fmt.Sprintf(`User request: "%s"

Available applications: %v
Available environments: %v

Extract the application and environment:`, userMessage, applications, environments)

	response, err := a.service.aiProvider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", "", fmt.Errorf("AI parsing failed: %w", err)
	}

	// Parse AI response
	var result struct {
		Application string `json:"application"`
		Environment string `json:"environment"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		a.logger.Warn("⚠️ AI response not valid JSON, trying fallback parsing")
		return a.parseWithFallback(userMessage, applications, environments)
	}

	if result.Application == "" || result.Environment == "" {
		return "", "", fmt.Errorf("AI could not extract application or environment from request")
	}

	a.logger.Info("🎯 AI extracted: app=%s, env=%s", result.Application, result.Environment)
	return result.Application, result.Environment, nil
}

// parseWithFallback provides simple text-based parsing as a fallback
func (a *DeploymentAgent) parseWithFallback(userMessage string, applications, environments []string) (string, string, error) {
	lower := strings.ToLower(userMessage)
	
	// Find application by substring matching
	var matchedApp string
	for _, app := range applications {
		if strings.Contains(lower, strings.ToLower(app)) {
			matchedApp = app
			break
		}
	}

	// Find environment by common patterns
	var matchedEnv string
	for _, env := range environments {
		envLower := strings.ToLower(env)
		if strings.Contains(lower, envLower) {
			matchedEnv = env
			break
		}
	}

	// Handle common environment aliases
	if matchedEnv == "" {
		if strings.Contains(lower, "prod") || strings.Contains(lower, "production") {
			for _, env := range environments {
				if strings.Contains(strings.ToLower(env), "prod") {
					matchedEnv = env
					break
				}
			}
		} else if strings.Contains(lower, "dev") || strings.Contains(lower, "development") {
			for _, env := range environments {
				if strings.Contains(strings.ToLower(env), "dev") {
					matchedEnv = env
					break
				}
			}
		}
	}

	if matchedApp == "" {
		return "", "", fmt.Errorf("could not identify application from message: %s", userMessage)
	}
	if matchedEnv == "" {
		return "", "", fmt.Errorf("could not identify environment from message: %s", userMessage)
	}

	a.logger.Info("🎯 Fallback extracted: app=%s, env=%s", matchedApp, matchedEnv)
	return matchedApp, matchedEnv, nil
}
