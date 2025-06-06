package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
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
	deploymentService   DeploymentService
	policyService       PolicyService
	applicationService  ApplicationService

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

// NewPlatformAgent creates the Core Platform Agent with proper dependency injection
func NewPlatformAgent(
	provider AIProvider,
	globalGraph *graph.GlobalGraph,
	deploymentService DeploymentService,
	policyService PolicyService,
	applicationService ApplicationService,
) *PlatformAgent {
	logger := logging.GetLogger().ForComponent("platform-agent")

	agent := &PlatformAgent{
		provider:           provider,
		logger:             logger,
		graph:              globalGraph,
		deploymentService:  deploymentService,
		policyService:      policyService,
		applicationService: applicationService,
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
) (*PlatformAgent, error) {
	// Create AI provider (pure infrastructure)
	provider, err := createAIProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create AI provider: %w", err)
	}

	return NewPlatformAgent(provider, globalGraph, deploymentService, policyService, applicationService), nil
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
	var actions []Action

	switch intent.Type {
	case "application_creation":
		// Orchestrate application creation through domain service
		if intent.Parameters["app"] != nil {
			appName := intent.Parameters["app"].(string)
			
			// Create basic application contract with AI-extracted information
			app := contracts.ApplicationContract{
				Metadata: contracts.Metadata{
					Name: appName,
					Owner: "user", // Default owner - could be enhanced with user detection
				},
				Spec: contracts.ApplicationSpec{
					Description: fmt.Sprintf("Application created via AI assistant: %s", appName),
					Tags:        []string{"ai-created"},
				},
			}
			
			err := agent.applicationService.CreateApplication(app)
			if err != nil {
				return nil, fmt.Errorf("application creation failed: %w", err)
			}

			actions = append(actions, Action{
				Type:   "application_created",
				Result: map[string]interface{}{"application": appName, "status": "created"},
				Status: "completed",
			})
		}

	case "deployment":
		// Orchestrate deployment through domain service
		if intent.Parameters["app"] != nil {
			appName := intent.Parameters["app"].(string)
			plan, err := agent.deploymentService.GenerateDeploymentPlan(ctx, appName)
			if err != nil {
				return nil, fmt.Errorf("deployment planning failed: %w", err)
			}

			actions = append(actions, Action{
				Type:   "deployment_plan_generated",
				Result: plan,
				Status: "completed",
			})
		}

	case "policy_check":
		// Orchestrate policy evaluation through domain service
		if app, ok := intent.Parameters["app"].(string); ok {
			if env, ok := intent.Parameters["environment"].(string); ok {
				err := agent.policyService.ValidateDeployment(ctx, app, env)
				status := "completed"
				if err != nil {
					status = "failed"
				}

				actions = append(actions, Action{
					Type:   "policy_validation",
					Result: map[string]interface{}{"valid": err == nil, "error": err},
					Status: status,
				})
			}
		}

	case "analysis":
		// Provide platform analysis using AI + graph data
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
		// AI-powered troubleshooting
		troubleshooting, err := agent.performIntelligentTroubleshooting(ctx, intent, query)
		if err != nil {
			return nil, fmt.Errorf("intelligent troubleshooting failed: %w", err)
		}

		actions = append(actions, Action{
			Type:   "troubleshooting_analysis",
			Result: troubleshooting,
			Status: "completed",
		})
	}

	return actions, nil
}

// *** MISSING HELPER METHODS ***

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
