package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
	"github.com/krzachariassen/ZTDP/internal/resources"
)

// PlatformAgent is the Core Platform Agent - the AI-native interface for ZTDP
// It orchestrates specialized domain services while providing conversational AI capabilities
type PlatformAgent struct {
	// AI Infrastructure
	provider AIProvider
	logger   *logging.Logger

	// Platform Context
	graph *graph.GlobalGraph

	// Domain Service Orchestration (injected dependencies)
	deploymentService  DeploymentService
	policyService      PolicyService
	applicationService ApplicationService
	serviceService     ServiceService
	resourceService    ResourceService
	environmentService EnvironmentService

	// AI-Native Capabilities
	conversationEngine *ConversationEngine
	intentRecognizer   *IntentRecognizer
	responseBuilder    *ResponseBuilder
}

// DeploymentService interface for domain service integration
type DeploymentService interface {
	GenerateDeploymentPlan(ctx context.Context, app string) (*DeploymentPlan, error)
	PredictDeploymentImpact(ctx context.Context, changes []ProposedChange, env string) (*ImpactPrediction, error)
	ExecuteDeployment(ctx context.Context, plan *DeploymentPlan) error
}

// PolicyService interface for domain service integration
type PolicyService interface {
	EvaluatePolicy(ctx context.Context, request *PolicyEvaluationRequest) (*PolicyEvaluation, error)
	ValidateDeployment(ctx context.Context, app, env string) error
}

// ApplicationService interface for domain service integration
type ApplicationService interface {
	CreateApplication(app contracts.ApplicationContract) error
	GetApplication(appName string) (*contracts.ApplicationContract, error)
}

// ServiceService interface for service domain operations
type ServiceService interface {
	CreateService(appName string, svcData map[string]interface{}) (map[string]interface{}, error)
	CreateServiceVersion(serviceName string, versionData map[string]interface{}) (map[string]interface{}, error)
	GetService(appName, serviceName string) (map[string]interface{}, error)
}

// ResourceService interface for resource domain operations
type ResourceService interface {
	CreateResource(req resources.ResourceRequest) (*resources.ResourceResponse, error)
	AddResourceToApplication(appName, resourceName, instanceName string) (*resources.ResourceInstanceResponse, error)
	LinkServiceToResource(appName, serviceName, resourceName string) (*resources.ResourceInstanceResponse, error)
}

// EnvironmentService interface for environment domain operations
type EnvironmentService interface {
	CreateEnvironmentFromData(envData map[string]interface{}) (map[string]interface{}, error)
}

// NewPlatformAgent creates the Core Platform Agent with proper dependency injection
func NewPlatformAgent(
	provider AIProvider,
	globalGraph *graph.GlobalGraph,
	deploymentService DeploymentService,
	policyService PolicyService,
	applicationService ApplicationService,
	serviceService ServiceService,
	resourceService ResourceService,
	environmentService EnvironmentService,
) *PlatformAgent {
	logger := logging.GetLogger().ForComponent("platform-agent")

	agent := &PlatformAgent{
		provider:           provider,
		logger:             logger,
		graph:              globalGraph,
		deploymentService:  deploymentService,
		policyService:      policyService,
		applicationService: applicationService,
		serviceService:     serviceService,
		resourceService:    resourceService,
		environmentService: environmentService,
	}

	// Initialize AI-native capabilities
	agent.conversationEngine = NewConversationEngine(provider, logger)
	agent.intentRecognizer = NewIntentRecognizer(provider, logger)
	agent.responseBuilder = NewResponseBuilder(logger)

	return agent
}

// NewPlatformAgentFromConfig creates Core Platform Agent from environment configuration
func NewPlatformAgentFromConfig(
	globalGraph *graph.GlobalGraph,
	deploymentService DeploymentService,
	policyService PolicyService,
	applicationService ApplicationService,
	serviceService ServiceService,
	resourceService ResourceService,
	environmentService EnvironmentService,
) (*PlatformAgent, error) {
	// Create AI provider (pure infrastructure)
	provider, err := createAIProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create AI provider: %w", err)
	}

	return NewPlatformAgent(provider, globalGraph, deploymentService, policyService, applicationService, serviceService, resourceService, environmentService), nil
}

// createAIProvider creates the appropriate AI provider based on configuration
func createAIProvider() (AIProvider, error) {
	providerName := os.Getenv("AI_PROVIDER")
	if providerName == "" {
		providerName = "openai"
	}

	switch providerName {
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
		}

		config := DefaultOpenAIConfig()
		if model := os.Getenv("OPENAI_MODEL"); model != "" {
			config.Model = model
		}
		if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
			config.BaseURL = baseURL
		}

		return NewOpenAIProvider(config, apiKey)
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", providerName)
	}
}

// *** AI-NATIVE PLATFORM INTERFACE ***

// ChatWithPlatform is the primary AI-native interface for developer interactions
// This is where developers primarily interact with the platform through natural language
func (agent *PlatformAgent) ChatWithPlatform(ctx context.Context, query string, context string) (*ConversationalResponse, error) {
	agent.logger.Info("🤖 Platform Agent processing conversation: %s", query)

	// 1. Extract platform context for AI reasoning
	platformContext, err := agent.extractPlatformContext(ctx, context)
	if err != nil {
		return nil, fmt.Errorf("failed to extract platform context: %w", err)
	}

	// 2. Recognize intent and determine required actions
	intent, err := agent.intentRecognizer.AnalyzeIntent(ctx, query, platformContext)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze intent: %w", err)
	}

	// 3. Execute actions through domain service orchestration
	actions, err := agent.executeIntentActions(ctx, intent, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute intent actions: %w", err)
	}

	// 4. Generate conversational response using AI
	response, err := agent.conversationEngine.GenerateResponse(ctx, query, intent, actions, platformContext)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	agent.logger.Info("✅ Conversation completed: %s with %d actions", intent.Type, len(actions))
	return response, nil
}

// executeIntentActions orchestrates domain services based on recognized intent
func (agent *PlatformAgent) executeIntentActions(ctx context.Context, intent *Intent, query string) ([]Action, error) {
	// Check if this requires complex orchestration
	if intent.Type == "complex_orchestration" {
		return agent.executeComplexOrchestration(ctx, intent, query)
	}

	// Handle simple single-step operations
	return agent.executeSingleStepOperation(ctx, intent, query)
}

// executeComplexOrchestration handles multi-step scenarios with intelligent planning
func (agent *PlatformAgent) executeComplexOrchestration(ctx context.Context, intent *Intent, query string) ([]Action, error) {
	agent.logger.Info("🎯 Executing complex orchestration for: %s", intent.Parameters["scenario"])

	// 1. Generate execution plan using AI
	plan, err := agent.generateExecutionPlan(ctx, intent, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate execution plan: %w", err)
	}

	// 2. Execute plan steps in dependency order
	actions, err := agent.executePlan(ctx, plan)
	if err != nil {
		return nil, fmt.Errorf("failed to execute plan: %w", err)
	}

	// 3. Return summary of all actions performed
	summaryAction := Action{
		Type: "complex_orchestration_completed",
		Result: map[string]interface{}{
			"scenario":        plan.Scenario,
			"steps_completed": len(plan.Steps),
			"plan_id":         plan.ID,
			"actions":         actions,
		},
		Status: "completed",
	}

	return []Action{summaryAction}, nil
}

// generateExecutionPlan creates a dynamic execution plan using AI
func (agent *PlatformAgent) generateExecutionPlan(ctx context.Context, intent *Intent, query string) (*ExecutionPlan, error) {
	agent.logger.Info("📋 Generating execution plan for complex scenario")

	// Extract plan from intent if already provided by AI
	if steps, ok := intent.Parameters["steps"].([]interface{}); ok {
		return agent.buildPlanFromIntentSteps(intent, steps)
	}

	// Otherwise, generate plan using AI
	systemPrompt := agent.buildPlanningSystemPrompt()
	userPrompt := agent.buildPlanningUserPrompt(query, intent)

	response, err := agent.provider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI planning failed: %w", err)
	}

	return agent.parsePlanFromAI(response, intent)
}

// buildPlanFromIntentSteps converts intent steps to ExecutionPlan
func (agent *PlatformAgent) buildPlanFromIntentSteps(intent *Intent, steps []interface{}) (*ExecutionPlan, error) {
	planID := uuid.New().String()
	scenario, _ := intent.Parameters["scenario"].(string)

	plan := &ExecutionPlan{
		ID:           planID,
		Scenario:     scenario,
		Status:       "planned",
		CreatedAt:    time.Now(),
		Dependencies: make(map[string][]string),
	}

	// Convert interface{} steps to ExecutionStep structs
	for i, stepData := range steps {
		stepMap, ok := stepData.(map[string]interface{})
		if !ok {
			continue
		}

		step := ExecutionStep{
			ID:     fmt.Sprintf("step_%d", i+1),
			Status: "pending",
		}

		// Extract step fields
		if op, ok := stepMap["operation"].(string); ok {
			step.Operation = op
		}
		if resType, ok := stepMap["resource_type"].(string); ok {
			step.ResourceType = resType
		}
		if resName, ok := stepMap["resource_name"].(string); ok {
			step.ResourceName = resName
		}
		if desc, ok := stepMap["description"].(string); ok {
			step.Description = desc
		}
		if params, ok := stepMap["parameters"].(map[string]interface{}); ok {
			step.Parameters = params
		}

		// Extract dependencies
		if deps, ok := stepMap["dependencies"].([]interface{}); ok {
			for _, dep := range deps {
				if depStr, ok := dep.(string); ok {
					step.Dependencies = append(step.Dependencies, depStr)
				}
			}
		}

		plan.Steps = append(plan.Steps, step)
		plan.Dependencies[step.ID] = step.Dependencies
	}

	return plan, nil
}

// parsePlanFromAI parses AI-generated execution plan
func (agent *PlatformAgent) parsePlanFromAI(response string, intent *Intent) (*ExecutionPlan, error) {
	var planData struct {
		Scenario string `json:"scenario"`
		Steps    []struct {
			ID           string                 `json:"id"`
			Operation    string                 `json:"operation"`
			ResourceType string                 `json:"resource_type"`
			ResourceName string                 `json:"resource_name"`
			Description  string                 `json:"description"`
			Dependencies []string               `json:"dependencies"`
			Parameters   map[string]interface{} `json:"parameters"`
		} `json:"steps"`
	}

	if err := json.Unmarshal([]byte(response), &planData); err != nil {
		return nil, fmt.Errorf("failed to parse AI execution plan: %w", err)
	}

	planID := uuid.New().String()
	plan := &ExecutionPlan{
		ID:           planID,
		Scenario:     planData.Scenario,
		Status:       "planned",
		CreatedAt:    time.Now(),
		Dependencies: make(map[string][]string),
	}

	// Convert parsed steps to ExecutionStep structs
	for _, stepData := range planData.Steps {
		step := ExecutionStep{
			ID:           stepData.ID,
			Operation:    stepData.Operation,
			ResourceType: stepData.ResourceType,
			ResourceName: stepData.ResourceName,
			Description:  stepData.Description,
			Dependencies: stepData.Dependencies,
			Parameters:   stepData.Parameters,
			Status:       "pending",
		}

		plan.Steps = append(plan.Steps, step)
		plan.Dependencies[step.ID] = step.Dependencies
	}

	return plan, nil
}

// executePlan executes plan steps in correct dependency order
func (agent *PlatformAgent) executePlan(ctx context.Context, plan *ExecutionPlan) ([]Action, error) {
	agent.logger.Info("🚀 Executing plan with %d steps", len(plan.Steps))

	var allActions []Action

	// Track completed steps
	completed := make(map[string]bool)

	// Execute steps in dependency order
	for len(completed) < len(plan.Steps) {
		progress := false

		for i := range plan.Steps {
			step := &plan.Steps[i]

			// Skip if already completed
			if completed[step.ID] {
				continue
			}

			// Check if all dependencies are completed
			canExecute := true
			for _, dep := range step.Dependencies {
				if !completed[dep] {
					canExecute = false
					break
				}
			}

			if !canExecute {
				continue
			}

			// Execute the step
			agent.logger.Info("⚡ Executing step: %s - %s", step.ID, step.Description)
			step.StartedAt = &time.Time{}
			*step.StartedAt = time.Now()
			step.Status = "executing"

			action, err := agent.executeStep(ctx, step)
			if err != nil {
				step.Status = "failed"
				step.Error = err.Error()
				return allActions, fmt.Errorf("step %s failed: %w", step.ID, err)
			}

			step.Status = "completed"
			step.CompletedAt = &time.Time{}
			*step.CompletedAt = time.Now()
			step.Result = action.Result

			completed[step.ID] = true
			allActions = append(allActions, action)
			progress = true

			agent.logger.Info("✅ Step completed: %s", step.ID)
		}

		// If no progress was made, we have circular dependencies
		if !progress {
			return allActions, fmt.Errorf("circular dependencies detected in execution plan")
		}
	}

	plan.Status = "completed"
	plan.CompletedAt = &time.Time{}
	*plan.CompletedAt = time.Now()

	return allActions, nil
}

// executeStep executes a single step in the execution plan
func (agent *PlatformAgent) executeStep(ctx context.Context, step *ExecutionStep) (Action, error) {
	switch step.Operation {
	case "application_creation":
		return agent.executeApplicationCreation(ctx, step)
	case "service_creation":
		return agent.executeServiceCreation(ctx, step)
	case "resource_creation":
		return agent.executeResourceCreation(ctx, step)
	case "resource_linking":
		return agent.executeResourceLinking(ctx, step)
	case "environment_setup":
		return agent.executeEnvironmentSetup(ctx, step)
	case "policy_creation":
		return agent.executePolicyCreation(ctx, step)
	default:
		return Action{}, fmt.Errorf("unsupported operation: %s", step.Operation)
	}
}

// executeApplicationCreation executes application creation step
func (agent *PlatformAgent) executeApplicationCreation(ctx context.Context, step *ExecutionStep) (Action, error) {
	appName := step.ResourceName
	description, _ := step.Parameters["description"].(string)
	tags, _ := step.Parameters["tags"].([]string)

	// Create application contract
	app := contracts.ApplicationContract{
		Metadata: contracts.Metadata{
			Name:  appName,
			Owner: "platform-agent",
		},
		Spec: contracts.ApplicationSpec{
			Description: description,
			Tags:        tags,
		},
	}

	// Create through domain service
	err := agent.applicationService.CreateApplication(app)
	if err != nil {
		return Action{}, fmt.Errorf("application creation failed: %w", err)
	}

	return Action{
		Type: "application_created",
		Result: map[string]interface{}{
			"application": appName,
			"status":      "created",
		},
		Status: "completed",
	}, nil
}

// executeServiceCreation executes service creation step
func (agent *PlatformAgent) executeServiceCreation(ctx context.Context, step *ExecutionStep) (Action, error) {
	serviceName := step.ResourceName
	appName, _ := step.Parameters["application"].(string)
	port, _ := step.Parameters["port"].(float64)
	description, _ := step.Parameters["description"].(string)

	serviceData := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":  serviceName,
			"owner": "platform-agent",
		},
		"spec": map[string]interface{}{
			"application": appName,
			"port":        int(port),
			"public":      false,
			"description": description,
			"tags":        []string{"ai-generated"},
		},
	}

	result, err := agent.serviceService.CreateService(appName, serviceData)
	if err != nil {
		return Action{}, fmt.Errorf("service creation failed: %w", err)
	}

	return Action{
		Type:   "service_created",
		Result: result,
		Status: "completed",
	}, nil
}

// executeResourceCreation executes resource creation step
func (agent *PlatformAgent) executeResourceCreation(ctx context.Context, step *ExecutionStep) (Action, error) {
	resourceName := step.ResourceName
	resourceType := step.ResourceType

	// Create resource in catalog
	req := resources.ResourceRequest{
		Kind: "resource",
		Metadata: map[string]interface{}{
			"name":  resourceName,
			"owner": "platform-agent",
		},
		Spec: map[string]interface{}{
			"type":       resourceType,
			"config_ref": fmt.Sprintf("config/%s/%s", "platform-agent", resourceName),
		},
	}

	// Merge additional parameters
	for key, value := range step.Parameters {
		req.Spec[key] = value
	}

	result, err := agent.resourceService.CreateResource(req)
	if err != nil {
		return Action{}, fmt.Errorf("resource creation failed: %w", err)
	}

	return Action{
		Type:   "resource_created",
		Result: result,
		Status: "completed",
	}, nil
}

// executeResourceLinking executes resource linking step
func (agent *PlatformAgent) executeResourceLinking(ctx context.Context, step *ExecutionStep) (Action, error) {
	appName, _ := step.Parameters["application"].(string)
	serviceName, _ := step.Parameters["service"].(string)
	resourceName := step.ResourceName

	if appName != "" && serviceName == "" {
		// Link resource to application
		result, err := agent.resourceService.AddResourceToApplication(appName, resourceName, "")
		if err != nil {
			return Action{}, fmt.Errorf("resource to application linking failed: %w", err)
		}

		return Action{
			Type:   "resource_linked_to_application",
			Result: result,
			Status: "completed",
		}, nil
	} else if appName != "" && serviceName != "" {
		// Link service to resource
		result, err := agent.resourceService.LinkServiceToResource(appName, serviceName, resourceName)
		if err != nil {
			return Action{}, fmt.Errorf("service to resource linking failed: %w", err)
		}

		return Action{
			Type:   "service_linked_to_resource",
			Result: result,
			Status: "completed",
		}, nil
	}

	return Action{}, fmt.Errorf("invalid resource linking parameters")
}

// executeEnvironmentSetup executes environment setup step
func (agent *PlatformAgent) executeEnvironmentSetup(ctx context.Context, step *ExecutionStep) (Action, error) {
	envName := step.ResourceName
	description, _ := step.Parameters["description"].(string)

	envData := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":  envName,
			"owner": "platform-agent",
		},
		"spec": map[string]interface{}{
			"description": description,
		},
	}

	result, err := agent.environmentService.CreateEnvironmentFromData(envData)
	if err != nil {
		return Action{}, fmt.Errorf("environment setup failed: %w", err)
	}

	return Action{
		Type:   "environment_created",
		Result: result,
		Status: "completed",
	}, nil
}

// executePolicyCreation executes policy creation step
func (agent *PlatformAgent) executePolicyCreation(ctx context.Context, step *ExecutionStep) (Action, error) {
	// For now, return a placeholder action since policy creation interface isn't fully defined
	return Action{
		Type: "policy_creation_placeholder",
		Result: map[string]interface{}{
			"policy": step.ResourceName,
			"status": "placeholder",
		},
		Status: "completed",
	}, nil
}

// buildPlanningSystemPrompt creates system prompt for AI planning
func (agent *PlatformAgent) buildPlanningSystemPrompt() string {
	return `You are an expert infrastructure orchestration planner for ZTDP platform.

Your job is to create detailed execution plans for complex infrastructure scenarios.

AVAILABLE OPERATIONS:
- application_creation: Create new applications
- service_creation: Create services (APIs, microservices)
- resource_creation: Create infrastructure (databases, storage, queues)
- resource_linking: Link services to resources or applications
- environment_setup: Set up environments
- policy_creation: Create governance policies

RESOURCE TYPES:
- application: Top-level application containers
- service: API services, web services, microservices
- database: SQL/NoSQL databases
- storage: File storage, object storage
- queue: Message queues, event systems
- cache: Redis, Memcached

RESPONSE FORMAT (JSON):
{
  "scenario": "clear description of what we're building",
  "steps": [
    {
      "id": "step_1",
      "operation": "application_creation",
      "resource_type": "application",
      "resource_name": "my-app",
      "description": "Create the main application container",
      "dependencies": [],
      "parameters": {
        "description": "Application description",
        "tags": ["tag1", "tag2"]
      }
    },
    {
      "id": "step_2",
      "operation": "resource_creation",
      "resource_type": "database",
      "resource_name": "my-app-db",
      "description": "Create PostgreSQL database",
      "dependencies": ["step_1"],
      "parameters": {
        "type": "postgresql",
        "size": "small"
      }
    }
  ]
}

Create logical steps with proper dependencies. Each step should depend on previous steps it needs.`
}

// buildPlanningUserPrompt creates user prompt for AI planning
func (agent *PlatformAgent) buildPlanningUserPrompt(query string, intent *Intent) string {
	prompt := fmt.Sprintf("Create an execution plan for: %s\n\n", query)

	if scenario, ok := intent.Parameters["scenario"].(string); ok {
		prompt += fmt.Sprintf("Scenario: %s\n\n", scenario)
	}

	prompt += "Generate a detailed step-by-step execution plan in JSON format."
	return prompt
}

// executeSingleStepOperation handles simple single-step operations
func (agent *PlatformAgent) executeSingleStepOperation(ctx context.Context, intent *Intent, query string) ([]Action, error) {
	var actions []Action

	switch intent.Type {
	case "application_creation":
		action, err := agent.handleApplicationCreation(ctx, intent)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)

	case "service_creation":
		action, err := agent.handleServiceCreation(ctx, intent)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)

	case "resource_creation":
		action, err := agent.handleResourceCreation(ctx, intent)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)

	case "deployment":
		action, err := agent.handleDeployment(ctx, intent)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)

	case "policy_check":
		action, err := agent.handlePolicyCheck(ctx, intent)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)

	case "analysis":
		analysis, err := agent.performPlatformAnalysis(ctx, intent, query)
		if err != nil {
			return nil, fmt.Errorf("platform analysis failed: %w", err)
		}
		actions = append(actions, Action{
			Type:   "platform_analysis",
			Result: analysis,
			Status: "completed",
		})

	case "troubleshooting":
		troubleshooting, err := agent.performIntelligentTroubleshooting(ctx, intent, query)
		if err != nil {
			return nil, fmt.Errorf("intelligent troubleshooting failed: %w", err)
		}
		actions = append(actions, Action{
			Type:   "troubleshooting_analysis",
			Result: troubleshooting,
			Status: "completed",
		})

	default:
		actions = append(actions, Action{
			Type:   "question_response",
			Result: map[string]interface{}{"message": "I understand your request, but I'm not sure how to help with that specific operation."},
			Status: "unsupported",
		})
	}

	return actions, nil
}

// handleApplicationCreation handles intelligent application creation using conversation flow
func (agent *PlatformAgent) handleApplicationCreation(ctx context.Context, intent *Intent) (Action, error) {
	agent.logger.Info("🎯 Starting intelligent application creation process")

	// Step 1: Check if we have a conversation ID for this creation process
	conversationID, hasConversation := intent.Parameters["conversation_id"].(string)
	if !hasConversation {
		// Start new conversation
		conversationID = fmt.Sprintf("app-creation-%d", time.Now().Unix())
		intent.Parameters["conversation_id"] = conversationID
	}

	// Step 2: Load Application Contract schema to understand required vs optional fields
	contractSchema := agent.getApplicationContractSchema()

	// Step 3: Collect and validate current parameters against contract requirements
	collectedData := agent.extractApplicationDataFromIntent(intent)
	missingRequired := agent.validateRequiredApplicationFields(collectedData, contractSchema)

	// Step 4: If we have missing required fields, use conversation engine to ask for them
	if len(missingRequired) > 0 {
		return agent.requestMissingApplicationFields(ctx, conversationID, missingRequired, collectedData)
	}

	// Step 5: All required data collected - create the application contract
	app := contracts.ApplicationContract{
		Metadata: contracts.Metadata{
			Name:  collectedData["name"].(string),
			Owner: agent.getOwnerFromContext(collectedData),
		},
		Spec: contracts.ApplicationSpec{
			Description: agent.getStringValue(collectedData, "description"),
			Tags:        agent.getStringSliceValue(collectedData, "tags"),
			Lifecycle:   agent.getLifecycleValue(collectedData),
		},
	}

	// Step 6: Create through domain service (maintains all business logic and validation)
	err := agent.applicationService.CreateApplication(app)
	if err != nil {
		// If creation fails, provide helpful guidance
		return Action{
			Type: "application_creation_failed",
			Result: map[string]interface{}{
				"error":           err.Error(),
				"conversation_id": conversationID,
				"guidance":        "Please check the application details and try again.",
			},
			Status: "failed",
		}, fmt.Errorf("application creation failed: %w", err)
	}

	agent.logger.Info("✅ Successfully created application: %s", app.Metadata.Name)

	return Action{
		Type: "application_created",
		Result: map[string]interface{}{
			"application":     app.Metadata.Name,
			"status":          "created",
			"conversation_id": conversationID,
			"contract":        app,
		},
		Status: "completed",
	}, nil
}

// handleServiceCreation handles single service creation
func (agent *PlatformAgent) handleServiceCreation(ctx context.Context, intent *Intent) (Action, error) {
	serviceName, ok := intent.Parameters["service"].(string)
	if !ok || serviceName == "" {
		return Action{}, fmt.Errorf("service name is required")
	}

	appName, _ := intent.Parameters["app"].(string)
	port, _ := intent.Parameters["port"].(float64)
	description, _ := intent.Parameters["description"].(string)

	serviceData := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":  serviceName,
			"owner": "platform-agent",
		},
		"spec": map[string]interface{}{
			"application": appName,
			"port":        int(port),
			"public":      false,
			"description": description,
			"tags":        []string{"ai-generated"},
		},
	}

	result, err := agent.serviceService.CreateService(appName, serviceData)
	if err != nil {
		return Action{}, fmt.Errorf("service creation failed: %w", err)
	}

	return Action{
		Type:   "service_created",
		Result: result,
		Status: "completed",
	}, nil
}

// handleResourceCreation handles single resource creation
func (agent *PlatformAgent) handleResourceCreation(ctx context.Context, intent *Intent) (Action, error) {
	resourceName, ok := intent.Parameters["resource"].(string)
	if !ok || resourceName == "" {
		return Action{}, fmt.Errorf("resource name is required")
	}

	resourceType, _ := intent.Parameters["type"].(string)

	req := resources.ResourceRequest{
		Kind: "resource",
		Metadata: map[string]interface{}{
			"name":  resourceName,
			"owner": "platform-agent",
		},
		Spec: map[string]interface{}{
			"type":       resourceType,
			"config_ref": fmt.Sprintf("config/%s/%s", "platform-agent", resourceName),
		},
	}

	result, err := agent.resourceService.CreateResource(req)
	if err != nil {
		return Action{}, fmt.Errorf("resource creation failed: %w", err)
	}

	return Action{
		Type:   "resource_created",
		Result: result,
		Status: "completed",
	}, nil
}

// handleDeployment handles deployment operations
func (agent *PlatformAgent) handleDeployment(ctx context.Context, intent *Intent) (Action, error) {
	appName, ok := intent.Parameters["app"].(string)
	if !ok || appName == "" {
		return Action{}, fmt.Errorf("application name is required for deployment")
	}

	plan, err := agent.deploymentService.GenerateDeploymentPlan(ctx, appName)
	if err != nil {
		return Action{}, fmt.Errorf("deployment planning failed: %w", err)
	}

	return Action{
		Type:   "deployment_plan_generated",
		Result: plan,
		Status: "completed",
	}, nil
}

// handlePolicyCheck handles policy validation
func (agent *PlatformAgent) handlePolicyCheck(ctx context.Context, intent *Intent) (Action, error) {
	appName, ok := intent.Parameters["app"].(string)
	if !ok || appName == "" {
		return Action{}, fmt.Errorf("application name is required for policy check")
	}

	env, _ := intent.Parameters["environment"].(string)
	if env == "" {
		env = "dev" // default environment
	}

	err := agent.policyService.ValidateDeployment(ctx, appName, env)
	status := "completed"
	if err != nil {
		status = "failed"
	}

	return Action{
		Type: "policy_validation",
		Result: map[string]interface{}{
			"valid": err == nil,
			"error": err,
		},
		Status: status,
	}, nil
}

// performPlatformAnalysis performs AI-driven platform analysis
func (agent *PlatformAgent) performPlatformAnalysis(ctx context.Context, intent *Intent, query string) (map[string]interface{}, error) {
	agent.logger.Info("🔍 Performing platform analysis for query: %s", query)

	// Build analysis prompt
	systemPrompt := `You are a platform analysis assistant. Analyze the platform state and provide insights as JSON:
{
  "summary": "analysis summary",
  "health_status": "healthy|warning|critical",
  "key_insights": ["insight1", "insight2"],
  "recommendations": ["rec1", "rec2"],
  "metrics": {},
  "risks": []
}`

	userPrompt := fmt.Sprintf(`Analyze the platform for: %s
Please provide comprehensive analysis and recommendations.`, query)

	// Call AI provider
	response, err := agent.provider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI analysis call failed: %w", err)
	}

	// Parse response
	var analysis map[string]interface{}
	if err := json.Unmarshal([]byte(response), &analysis); err != nil {
		// Fallback to simple analysis if parsing fails
		analysis = map[string]interface{}{
			"summary":         "Platform analysis completed",
			"health_status":   "healthy",
			"key_insights":    []string{"Analysis available via AI provider"},
			"recommendations": []string{"Continue monitoring platform health"},
			"raw_response":    response,
		}
	}

	return analysis, nil
}

// performIntelligentTroubleshooting performs AI-driven troubleshooting
func (agent *PlatformAgent) performIntelligentTroubleshooting(ctx context.Context, intent *Intent, query string) (map[string]interface{}, error) {
	agent.logger.Info("🚨 Performing intelligent troubleshooting for: %s", query)

	// Build troubleshooting prompt
	systemPrompt := `You are a troubleshooting assistant. Analyze the issue and provide solutions as JSON:
{
  "issue_type": "deployment|network|policy|resource",
  "severity": "low|medium|high|critical",
  "probable_causes": ["cause1", "cause2"],
  "solutions": [
    {
      "title": "solution title",
      "steps": ["step1", "step2"],
      "priority": "high|medium|low"
    }
  ],
  "prevention": ["prevention1", "prevention2"],
  "related_docs": ["doc1", "doc2"]
}`

	userPrompt := fmt.Sprintf(`Troubleshoot this issue: %s
Please provide detailed analysis and solutions.`, query)

	// Call AI provider
	response, err := agent.provider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI troubleshooting call failed: %w", err)
	}

	// Parse response
	var troubleshooting map[string]interface{}
	if err := json.Unmarshal([]byte(response), &troubleshooting); err != nil {
		// Fallback to simple troubleshooting if parsing fails
		troubleshooting = map[string]interface{}{
			"issue_type":      "general",
			"severity":        "medium",
			"probable_causes": []string{"Investigation in progress"},
			"solutions": []map[string]interface{}{
				{
					"title":    "Contact Support",
					"steps":    []string{"Gather logs", "Create support ticket"},
					"priority": "medium",
				},
			},
			"raw_response": response,
		}
	}

	return troubleshooting, nil
}

// *** LEGACY COMPATIBILITY METHODS ***
// These maintain compatibility while the codebase transitions to AI-native patterns

// GetProvider returns the underlying AI provider for legacy compatibility
func (agent *PlatformAgent) GetProvider() AIProvider {
	return agent.provider
}

// Provider returns the AI provider instance for legacy compatibility
func (agent *PlatformAgent) Provider() AIProvider {
	return agent.provider
}

// GetProviderInfo returns AI provider information
func (agent *PlatformAgent) GetProviderInfo() *ProviderInfo {
	return agent.provider.GetProviderInfo()
}

// Close cleans up the Platform Agent resources
func (agent *PlatformAgent) Close() error {
	agent.logger.Info("🔌 Closing Platform Agent")
	return agent.provider.Close()
}

// *** PLATFORM CONTEXT EXTRACTION ***

// extractPlatformContext builds comprehensive platform state for AI reasoning
func (agent *PlatformAgent) extractPlatformContext(ctx context.Context, contextHint string) (*PlatformContext, error) {
	globalGraph, err := agent.graph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get platform graph: %w", err)
	}

	context := &PlatformContext{
		Applications: agent.extractApplicationSummary(globalGraph),
		Services:     agent.extractServiceSummary(globalGraph),
		Dependencies: agent.extractDependencyMap(globalGraph),
		Policies:     agent.extractPolicySummary(globalGraph),
		Environments: agent.extractEnvironmentSummary(globalGraph),
		Health:       agent.extractHealthStatus(globalGraph),
		RecentEvents: agent.extractRecentEvents(),
		ContextHint:  contextHint,
		Timestamp:    time.Now(),
	}

	return context, nil
}

// extractApplicationSummary extracts application information from the graph
func (agent *PlatformAgent) extractApplicationSummary(graph *graph.Graph) map[string]interface{} {
	applications := make(map[string]interface{})

	for nodeID, node := range graph.Nodes {
		if node.Kind == "application" {
			applications[nodeID] = map[string]interface{}{
				"id":       node.ID,
				"kind":     node.Kind,
				"metadata": node.Metadata,
				"spec":     node.Spec,
			}
		}
	}

	return applications
}

// extractServiceSummary extracts service information from the graph
func (agent *PlatformAgent) extractServiceSummary(graph *graph.Graph) map[string]interface{} {
	services := make(map[string]interface{})

	for nodeID, node := range graph.Nodes {
		if node.Kind == "service" {
			services[nodeID] = map[string]interface{}{
				"id":       node.ID,
				"kind":     node.Kind,
				"metadata": node.Metadata,
				"spec":     node.Spec,
			}
		}
	}

	return services
}

// extractDependencyMap extracts dependency relationships from the graph
func (agent *PlatformAgent) extractDependencyMap(graph *graph.Graph) map[string]interface{} {
	dependencies := make(map[string]interface{})

	for sourceID, edges := range graph.Edges {
		var deps []string
		for _, edge := range edges {
			if edge.Type == "depends" {
				deps = append(deps, edge.To)
			}
		}
		if len(deps) > 0 {
			dependencies[sourceID] = deps
		}
	}

	return dependencies
}

// extractPolicySummary extracts policy information from the graph
func (agent *PlatformAgent) extractPolicySummary(graph *graph.Graph) map[string]interface{} {
	policies := make(map[string]interface{})

	for nodeID, node := range graph.Nodes {
		if node.Kind == "policy" {
			policies[nodeID] = map[string]interface{}{
				"id":       node.ID,
				"kind":     node.Kind,
				"metadata": node.Metadata,
				"spec":     node.Spec,
			}
		}
	}

	return policies
}

// extractEnvironmentSummary extracts environment information from the graph
func (agent *PlatformAgent) extractEnvironmentSummary(graph *graph.Graph) map[string]interface{} {
	environments := make(map[string]interface{})

	for nodeID, node := range graph.Nodes {
		if node.Kind == "environment" {
			environments[nodeID] = map[string]interface{}{
				"id":       node.ID,
				"kind":     node.Kind,
				"metadata": node.Metadata,
				"spec":     node.Spec,
			}
		}
	}

	return environments
}

// extractHealthStatus extracts health status information
func (agent *PlatformAgent) extractHealthStatus(graph *graph.Graph) map[string]interface{} {
	health := map[string]interface{}{
		"total_applications": len(agent.extractApplicationSummary(graph)),
		"total_services":     len(agent.extractServiceSummary(graph)),
		"total_policies":     len(agent.extractPolicySummary(graph)),
		"status":             "operational", // Would integrate with health monitoring
		"last_checked":       time.Now(),
	}

	return health
}

// extractRecentEvents extracts recent platform events
func (agent *PlatformAgent) extractRecentEvents() []map[string]interface{} {
	// Would integrate with event system to get recent events
	events := []map[string]interface{}{
		{
			"type":      "platform.agent.started",
			"timestamp": time.Now(),
			"source":    "platform-agent",
		},
	}

	return events
}

// *** INTELLIGENT APPLICATION CREATION HELPERS ***

// getApplicationContractSchema returns the schema for Application Contract validation
func (agent *PlatformAgent) getApplicationContractSchema() map[string]interface{} {
	return map[string]interface{}{
		"required": []string{"name"}, // Only name is required per ApplicationContract.Validate()
		"optional": []string{"description", "tags", "lifecycle", "owner"},
		"fields": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Application name (must be unique)",
				"required":    true,
			},
			"owner": map[string]interface{}{
				"type":        "string",
				"description": "Application owner (team or individual)",
				"required":    false,
				"default":     "platform-agent",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Application description",
				"required":    false,
			},
			"tags": map[string]interface{}{
				"type":        "array",
				"description": "Application tags for categorization",
				"required":    false,
			},
			"lifecycle": map[string]interface{}{
				"type":        "object",
				"description": "Lifecycle configuration",
				"required":    false,
			},
		},
	}
}

// extractApplicationDataFromIntent extracts application data from intent parameters
func (agent *PlatformAgent) extractApplicationDataFromIntent(intent *Intent) map[string]interface{} {
	data := make(map[string]interface{})

	// Try various parameter names that might contain the application name
	if name, ok := intent.Parameters["app"].(string); ok && name != "" {
		data["name"] = name
	} else if name, ok := intent.Parameters["application"].(string); ok && name != "" {
		data["name"] = name
	} else if name, ok := intent.Parameters["name"].(string); ok && name != "" {
		data["name"] = name
	}

	// Extract other optional fields
	if desc, ok := intent.Parameters["description"].(string); ok && desc != "" {
		data["description"] = desc
	}

	if owner, ok := intent.Parameters["owner"].(string); ok && owner != "" {
		data["owner"] = owner
	}

	if tags, ok := intent.Parameters["tags"].([]interface{}); ok {
		stringTags := make([]string, 0, len(tags))
		for _, tag := range tags {
			if str, ok := tag.(string); ok {
				stringTags = append(stringTags, str)
			}
		}
		if len(stringTags) > 0 {
			data["tags"] = stringTags
		}
	} else if tags, ok := intent.Parameters["tags"].([]string); ok {
		data["tags"] = tags
	}

	if lifecycle, ok := intent.Parameters["lifecycle"].(map[string]interface{}); ok {
		data["lifecycle"] = lifecycle
	}

	return data
}

// validateRequiredApplicationFields checks which required fields are missing
func (agent *PlatformAgent) validateRequiredApplicationFields(data map[string]interface{}, schema map[string]interface{}) []string {
	required, ok := schema["required"].([]string)
	if !ok {
		return []string{}
	}

	var missing []string
	for _, field := range required {
		if value, exists := data[field]; !exists || value == "" {
			missing = append(missing, field)
		}
	}

	return missing
}

// requestMissingApplicationFields creates a conversation response asking for missing fields
func (agent *PlatformAgent) requestMissingApplicationFields(ctx context.Context, conversationID string, missingFields []string, currentData map[string]interface{}) (Action, error) {
	agent.logger.Info("🤔 Missing required fields for application creation: %v", missingFields)

	// Build a helpful prompt using AI to ask for missing information
	systemPrompt := `You are helping a developer create an application in ZTDP platform. 
The developer has provided some information but is missing required fields.
Be friendly, helpful, and specific about what information is needed.
Keep your response concise and actionable.`

	userPrompt := fmt.Sprintf(`The developer wants to create an application but is missing these required fields: %v

Current information provided: %v

Please ask for the missing information in a friendly, helpful way. Explain what each field is for.`,
		missingFields, currentData)

	response, err := agent.provider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		// Fallback to simple prompt if AI is unavailable
		response = fmt.Sprintf("I need more information to create your application. Please provide: %s", strings.Join(missingFields, ", "))
	}

	return Action{
		Type: "conversation_response",
		Result: map[string]interface{}{
			"message":         response,
			"conversation_id": conversationID,
			"missing_fields":  missingFields,
			"current_data":    currentData,
			"next_action":     "provide_missing_fields",
		},
		Status: "waiting_for_input",
	}, nil
}

// getOwnerFromContext determines the owner for the application
func (agent *PlatformAgent) getOwnerFromContext(data map[string]interface{}) string {
	if owner, ok := data["owner"].(string); ok && owner != "" {
		return owner
	}
	return "platform-agent" // Default owner
}

// getStringValue safely extracts a string value with default
func (agent *PlatformAgent) getStringValue(data map[string]interface{}, key string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return ""
}

// getStringSliceValue safely extracts a string slice with default
func (agent *PlatformAgent) getStringSliceValue(data map[string]interface{}, key string) []string {
	if value, ok := data[key].([]string); ok {
		return value
	}
	return []string{}
}

// getLifecycleValue safely extracts lifecycle configuration
func (agent *PlatformAgent) getLifecycleValue(data map[string]interface{}) map[string]contracts.LifecycleDefinition {
	if value, ok := data["lifecycle"].(map[string]interface{}); ok {
		// Convert to proper lifecycle format
		lifecycle := make(map[string]contracts.LifecycleDefinition)
		for k, v := range value {
			if def, ok := v.(map[string]interface{}); ok {
				lifecycle[k] = contracts.LifecycleDefinition{
					Gates: agent.getStringSliceFromInterface(def["gates"]),
				}
			}
		}
		return lifecycle
	}
	return make(map[string]contracts.LifecycleDefinition) // Default empty lifecycle
}

// getStringSliceFromInterface safely converts interface{} to []string
func (agent *PlatformAgent) getStringSliceFromInterface(value interface{}) []string {
	switch v := value.(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	default:
		return []string{}
	}
}
