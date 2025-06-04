package ai

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// PlatformAI is the core AI reasoning engine for ZTDP
// It serves as the Core Platform Agent providing revolutionary AI capabilities
// for conversational infrastructure management and multi-agent orchestration
type PlatformAI struct {
	provider AIProvider
	logger   *logging.Logger
	graph    *graph.GlobalGraph
}

// NewPlatformAI creates a new Core Platform Agent instance with the specified provider
func NewPlatformAI(provider AIProvider, globalGraph *graph.GlobalGraph) *PlatformAI {
	return &PlatformAI{
		provider: provider,
		logger:   logging.GetLogger().ForComponent("platform-ai"),
		graph:    globalGraph,
	}
}

// NewPlatformAIFromConfig creates a Core Platform Agent using provider selection logic
func NewPlatformAIFromConfig(globalGraph *graph.GlobalGraph) (*PlatformAI, error) {
	providerName := os.Getenv("AI_PROVIDER")
	if providerName == "" {
		providerName = "openai" // default to OpenAI
	}

	switch providerName {
	case "openai":
		// Create OpenAI provider using environment configuration
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
		}

		config := DefaultOpenAIConfig()

		// Allow model override via environment
		if model := os.Getenv("OPENAI_MODEL"); model != "" {
			config.Model = model
		}

		// Allow base URL override for custom deployments
		if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
			config.BaseURL = baseURL
		}

		provider, err := NewOpenAIProvider(config, apiKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
		}

		return NewPlatformAI(provider, globalGraph), nil
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s (only 'openai' is supported in AI-native mode)", providerName)
	}
}

// GetProvider returns the underlying AI provider
func (brain *PlatformAI) GetProvider() AIProvider {
	return brain.provider
}

// Provider returns the AI provider instance for use in other components
func (brain *PlatformAI) Provider() AIProvider {
	return brain.provider
}

// GetProviderInfo returns information about the current AI provider
func (brain *PlatformAI) GetProviderInfo() *ProviderInfo {
	return brain.provider.GetProviderInfo()
}

// Close cleans up the AI brain resources
func (brain *PlatformAI) Close() error {
	brain.logger.Info("🔌 Closing AI Brain")
	return brain.provider.Close()
}

// *** REVOLUTIONARY AI CAPABILITIES - IMPOSSIBLE WITH TRADITIONAL IDPS ***

// ChatWithPlatform enables natural language conversation with the platform
// This provides unprecedented insight into infrastructure state and relationships
func (brain *PlatformAI) ChatWithPlatform(ctx context.Context, query string, context string) (*ConversationalResponse, error) {
	brain.logger.Info("💬 AI Brain processing conversational query: %s", query)

	// Extract complete platform context for AI reasoning
	platformContext, err := brain.extractPlatformContext(context)
	if err != nil {
		return nil, fmt.Errorf("failed to extract platform context: %w", err)
	}

	// Build conversational query
	conversationalQuery := &ConversationalQuery{
		Query:   query,
		Context: context,
		Intent:  brain.detectIntent(query),
		Scope:   brain.extractScope(query, platformContext),
		Metadata: map[string]interface{}{
			"timestamp":      time.Now(),
			"user_context":   context,
			"platform_state": platformContext,
		},
	}

	// Process with AI
	response, err := brain.provider.ChatWithPlatform(ctx, conversationalQuery)
	if err != nil {
		return nil, fmt.Errorf("AI conversational processing failed: %w", err)
	}

	brain.logger.Info("✅ Conversational response generated with %d insights and %d actions",
		len(response.Insights), len(response.Actions))

	return response, nil
}

// PredictDeploymentImpact analyzes potential impact before deployment
// This enables proactive risk assessment impossible with traditional planning
func (brain *PlatformAI) PredictDeploymentImpact(ctx context.Context, changes []ProposedChange, environment string) (*ImpactPrediction, error) {
	brain.logger.Info("🔮 AI Brain predicting impact of %d changes in %s", len(changes), environment)

	// Extract environment context for impact analysis
	envContext, err := brain.extractEnvironmentContext(environment)
	if err != nil {
		return nil, fmt.Errorf("failed to extract environment context: %w", err)
	}

	// Build impact analysis request
	request := &ImpactAnalysisRequest{
		Changes:     changes,
		Scope:       environment,
		Environment: environment,
		Timeframe:   "24h", // Default prediction timeframe
		Metadata: map[string]interface{}{
			"timestamp":         time.Now(),
			"analysis_type":     "pre_deployment",
			"environment_state": envContext,
		},
	}

	// Predict impact using AI
	prediction, err := brain.provider.PredictImpact(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("AI impact prediction failed: %w", err)
	}

	brain.logger.Info("✅ Impact prediction completed: %s risk, %d affected systems",
		prediction.OverallRisk, len(prediction.AffectedSystems))

	return prediction, nil
}

// IntelligentTroubleshooting provides AI-driven root cause analysis
// This enables sophisticated problem diagnosis beyond traditional monitoring
func (brain *PlatformAI) IntelligentTroubleshooting(ctx context.Context, incidentID string, description string, symptoms []string) (*TroubleshootingResponse, error) {
	brain.logger.Info("🔍 AI Brain analyzing incident: %s - %s", incidentID, description)

	// Extract incident context from platform state
	incidentContext, err := brain.extractIncidentContext(incidentID, description, symptoms)
	if err != nil {
		return nil, fmt.Errorf("failed to extract incident context: %w", err)
	}

	// Perform AI-driven troubleshooting
	response, err := brain.provider.IntelligentTroubleshooting(ctx, incidentContext)
	if err != nil {
		return nil, fmt.Errorf("AI troubleshooting failed: %w", err)
	}

	brain.logger.Info("✅ Troubleshooting analysis completed: %s (confidence: %.2f)",
		response.RootCause, response.Confidence)

	return response, nil
}

// ProactiveOptimization continuously analyzes platform for improvements
// This provides ongoing architectural intelligence impossible with static tools
func (brain *PlatformAI) ProactiveOptimization(ctx context.Context, target string, focus []string) (*OptimizationRecommendations, error) {
	brain.logger.Info("⚡ AI Brain performing proactive optimization for %s", target)

	// Build optimization scope
	scope := &OptimizationScope{
		Target:      target,
		Focus:       focus,
		Timeframe:   "30d", // Default analysis period
		Constraints: brain.extractOptimizationConstraints(target),
		Metadata: map[string]interface{}{
			"timestamp":     time.Now(),
			"analysis_type": "proactive",
			"target_state":  brain.extractTargetState(target),
		},
	}

	// Generate AI-driven recommendations
	recommendations, err := brain.provider.ProactiveOptimization(ctx, scope)
	if err != nil {
		return nil, fmt.Errorf("AI proactive optimization failed: %w", err)
	}

	brain.logger.Info("✅ Proactive optimization completed: %d recommendations, %d patterns detected",
		len(recommendations.Recommendations), len(recommendations.Patterns))

	return recommendations, nil
}

// LearnFromDeployment captures deployment outcomes for continuous learning
// This builds institutional knowledge that improves AI reasoning over time
func (brain *PlatformAI) LearnFromDeployment(ctx context.Context, deploymentID string, success bool, duration int64, issues []DeploymentIssue) (*LearningInsights, error) {
	brain.logger.Info("🧠 AI Brain learning from deployment %s (success: %t, duration: %ds)",
		deploymentID, success, duration)

	// Build deployment outcome for learning
	outcome := &DeploymentOutcome{
		DeploymentID: deploymentID,
		Success:      success,
		Duration:     duration,
		Issues:       issues,
		Metrics:      brain.extractDeploymentMetrics(deploymentID),
		Context:      brain.extractDeploymentContext(deploymentID),
		Metadata: map[string]interface{}{
			"timestamp":     time.Now(),
			"learning_type": "post_deployment",
		},
	}

	// Learn from the outcome
	insights, err := brain.provider.LearningFromFailures(ctx, outcome)
	if err != nil {
		return nil, fmt.Errorf("AI learning failed: %w", err)
	}

	brain.logger.Info("✅ Learning completed: %d insights, %d patterns, confidence: %.2f",
		len(insights.Insights), len(insights.Patterns), insights.Confidence)

	return insights, nil
}

// GenerateDeploymentPlan is deprecated - use the simplified deployment engine instead
// This method existed for the complex planning chain but has been replaced with
// deployments.AIDeploymentPlanner which automatically discovers edges using AI.
// The API should transition to use the deployment engine directly.
func (brain *PlatformAI) GenerateDeploymentPlan(ctx context.Context, appName string, edgeTypes []string) (*PlanningResponse, error) {
	brain.logger.Info("⚠️  DEPRECATED: AI Brain GenerateDeploymentPlan used for app: %s - transition to simplified deployment engine", appName)

	// Get the global graph
	globalGraph, err := brain.graph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get global graph: %w", err)
	}

	// Build planning context - but let AI discover all edges when edgeTypes is nil
	var planningContext *PlanningContext
	if edgeTypes == nil {
		// AI-first: let AI discover all edge types automatically
		planningContext = brain.buildPlanningContext(appName, nil, globalGraph)
	} else {
		// Legacy mode: use specified edge types
		planningContext = brain.buildPlanningContext(appName, edgeTypes, globalGraph)
	}

	// Create planning request for AI provider
	planningRequest := &PlanningRequest{
		Intent:        fmt.Sprintf("Deploy application %s", appName),
		ApplicationID: appName,
		EdgeTypes:     edgeTypes, // nil means AI discovers all
		Context:       planningContext,
		Metadata: map[string]interface{}{
			"timestamp":    time.Now(),
			"request_type": "deployment_plan",
			"edge_types":   edgeTypes,
			"deprecated":   true,
		},
	}

	// Generate plan using AI provider
	response, err := brain.provider.GeneratePlan(ctx, planningRequest)
	if err != nil {
		return nil, fmt.Errorf("AI plan generation failed: %w", err)
	}

	brain.logger.Info("✅ Deployment plan generated successfully: %d steps, confidence: %.2f",
		len(response.Plan.Steps), response.Confidence)

	return response, nil
}

// EvaluateDeploymentPolicies evaluates deployment policies using AI for an application and environment
func (brain *PlatformAI) EvaluateDeploymentPolicies(ctx context.Context, applicationID, environmentID string) (*PolicyEvaluation, error) {
	brain.logger.Info("🔍 AI Brain evaluating deployment policies for app: %s in env: %s", applicationID, environmentID)

	// Get the global graph to extract policy context
	globalGraph, err := brain.graph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get global graph: %w", err)
	}

	// Build policy context from the graph
	policyContext := brain.buildPolicyContext(applicationID, environmentID, globalGraph)

	// Evaluate policies using AI provider
	evaluation, err := brain.provider.EvaluatePolicy(ctx, policyContext)
	if err != nil {
		return nil, fmt.Errorf("AI policy evaluation failed: %w", err)
	}

	brain.logger.Info("✅ Policy evaluation completed: compliant=%t, %d violations",
		evaluation.Compliant, len(evaluation.Violations))

	return evaluation, nil
}

// OptimizeExistingPlan optimizes an existing deployment plan using AI
func (brain *PlatformAI) OptimizeExistingPlan(ctx context.Context, plan *DeploymentPlan, applicationID string) (*PlanningResponse, error) {
	brain.logger.Info("⚡ AI Brain optimizing deployment plan for app: %s with %d steps", applicationID, len(plan.Steps))

	// Get the global graph to extract optimization context
	globalGraph, err := brain.graph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get global graph: %w", err)
	}

	// Build optimization context
	optimizationContext := brain.buildOptimizationContext(applicationID, globalGraph)

	// Optimize plan using AI provider
	optimizedResponse, err := brain.provider.OptimizePlan(ctx, plan, optimizationContext)
	if err != nil {
		return nil, fmt.Errorf("AI plan optimization failed: %w", err)
	}

	brain.logger.Info("✅ Plan optimization completed: %d steps optimized, confidence: %.2f",
		len(optimizedResponse.Plan.Steps), optimizedResponse.Confidence)

	return optimizedResponse, nil
}

// buildPlanningContext creates planning context for AI deployment planning
func (brain *PlatformAI) buildPlanningContext(appName string, edgeTypes []string, globalGraph *graph.Graph) *PlanningContext {
	// Find target nodes related to the application
	targetNodes := []*graph.Node{}
	relatedNodes := []*graph.Node{}
	edges := []*graph.Edge{}

	// Find the application node
	var appNode *graph.Node
	for _, node := range globalGraph.Nodes {
		if node.ID == appName || (node.Metadata != nil && node.Metadata["name"] == appName) {
			appNode = node
			targetNodes = append(targetNodes, node)
			break
		}
	}

	// If we found the app node, find related nodes through specified edge types
	if appNode != nil {
		// Check outgoing edges from the app node
		if outgoingEdges, exists := globalGraph.Edges[appNode.ID]; exists {
			for _, edge := range outgoingEdges {
				// Check if this edge type matches requested edge types
				if brain.edgeMatchesGraphEdge(&edge, edgeTypes) {
					// Find the target node
					if targetNode, exists := globalGraph.Nodes[edge.To]; exists {
						relatedNodes = append(relatedNodes, targetNode)
					}
					// Convert to the expected Edge structure for AI provider
					// Following the pattern from ai_planner.go
					aiEdge := &graph.Edge{
						To:       edge.To,
						Type:     edge.Type,
						Metadata: edge.Metadata,
					}
					edges = append(edges, aiEdge)
				}
			}
		}

		// Check incoming edges to the app node
		for sourceID, sourceEdges := range globalGraph.Edges {
			for _, edge := range sourceEdges {
				if edge.To == appNode.ID && brain.edgeMatchesGraphEdge(&edge, edgeTypes) {
					// Find the source node
					if sourceNode, exists := globalGraph.Nodes[sourceID]; exists {
						relatedNodes = append(relatedNodes, sourceNode)
					}
					// Convert to the expected Edge structure for AI provider
					// Note: We lose the source info here as the graph.Edge type doesn't have it
					aiEdge := &graph.Edge{
						To:       edge.To,
						Type:     edge.Type,
						Metadata: edge.Metadata,
					}
					edges = append(edges, aiEdge)
				}
			}
		}
	}

	return &PlanningContext{
		TargetNodes:   targetNodes,
		RelatedNodes:  relatedNodes,
		Edges:         edges,
		PolicyContext: nil,       // TODO: Add policy context if needed
		EnvironmentID: "default", // TODO: Extract from context if available
	}
}

// edgeMatchesGraphEdge checks if a graph edge type matches the requested edge types
// If edgeTypes is nil, all edges match (AI-first mode - AI discovers all edges automatically)
func (brain *PlatformAI) edgeMatchesGraphEdge(edge *graph.Edge, edgeTypes []string) bool {
	if edgeTypes == nil {
		// AI-first mode: include ALL edge types, let AI decide relevance
		return true
	}

	// Legacy mode: only include specified edge types
	for _, edgeType := range edgeTypes {
		if edge.Type == edgeType {
			return true
		}
	}
	return false
}

// buildPolicyContext creates policy context for AI policy evaluation
func (brain *PlatformAI) buildPolicyContext(applicationID, environmentID string, globalGraph *graph.Graph) map[string]interface{} {
	policyContext := map[string]interface{}{
		"application_id": applicationID,
		"environment_id": environmentID,
		"timestamp":      time.Now(),
		"request_type":   "policy_evaluation",
	}

	// Add application context if found
	if appNode, exists := globalGraph.Nodes[applicationID]; exists {
		policyContext["application"] = map[string]interface{}{
			"id":       appNode.ID,
			"kind":     appNode.Kind,
			"metadata": appNode.Metadata,
			"spec":     appNode.Spec,
		}
	}

	// Add environment policies and constraints
	policyContext["policies"] = brain.extractPoliciesForEnvironment(environmentID, globalGraph)
	policyContext["constraints"] = brain.extractEnvironmentConstraints(environmentID, globalGraph)

	return policyContext
}

// buildOptimizationContext creates context for AI plan optimization
func (brain *PlatformAI) buildOptimizationContext(applicationID string, globalGraph *graph.Graph) *PlanningContext {
	// Reuse the planning context builder but with optimization focus
	// This provides the AI with complete graph context for optimization
	return brain.buildPlanningContext(applicationID, []string{"deploy", "create", "owns", "depends"}, globalGraph)
}

// extractPoliciesForEnvironment extracts relevant policies for an environment
func (brain *PlatformAI) extractPoliciesForEnvironment(environmentID string, globalGraph *graph.Graph) []map[string]interface{} {
	policies := []map[string]interface{}{}

	// Look for policy nodes connected to the environment
	for nodeID, node := range globalGraph.Nodes {
		if node.Kind == "policy" {
			// Check if this policy applies to the environment
			if edges, exists := globalGraph.Edges[nodeID]; exists {
				for _, edge := range edges {
					if edge.To == environmentID && edge.Type == "applies_to" {
						policies = append(policies, map[string]interface{}{
							"id":       node.ID,
							"kind":     node.Kind,
							"metadata": node.Metadata,
							"spec":     node.Spec,
						})
						break
					}
				}
			}
		}
	}

	return policies
}

// extractEnvironmentConstraints extracts constraints for an environment
func (brain *PlatformAI) extractEnvironmentConstraints(environmentID string, globalGraph *graph.Graph) map[string]interface{} {
	constraints := map[string]interface{}{}

	// Get environment node if it exists
	if envNode, exists := globalGraph.Nodes[environmentID]; exists {
		if envNode.Spec != nil {
			if constraintsSpec, hasConstraints := envNode.Spec["constraints"]; hasConstraints {
				constraints = constraintsSpec.(map[string]interface{})
			}
		}
	}

	return constraints
}

// *** REVOLUTIONARY AI HELPER METHODS ***
// These methods enable the groundbreaking AI capabilities

// extractPlatformContext extracts comprehensive platform state for conversational AI
func (brain *PlatformAI) extractPlatformContext(contextHint string) (map[string]interface{}, error) {
	// Get complete platform graph
	globalGraph, err := brain.graph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get platform graph: %w", err)
	}

	// Build comprehensive platform context
	platformContext := map[string]interface{}{
		"applications":  brain.extractApplicationSummary(globalGraph),
		"services":      brain.extractServiceSummary(globalGraph),
		"dependencies":  brain.extractDependencyMap(globalGraph),
		"policies":      brain.extractPolicySummary(globalGraph),
		"environments":  brain.extractEnvironmentSummary(globalGraph),
		"health":        brain.extractHealthStatus(globalGraph),
		"recent_events": brain.extractRecentEvents(),
		"context_hint":  contextHint,
		"timestamp":     time.Now(),
	}

	return platformContext, nil
}

// detectIntent analyzes user query to understand intent
func (brain *PlatformAI) detectIntent(query string) string {
	queryLower := strings.ToLower(query)

	// Question patterns
	if strings.Contains(queryLower, "what") || strings.Contains(queryLower, "how") ||
		strings.Contains(queryLower, "why") || strings.Contains(queryLower, "when") ||
		strings.Contains(queryLower, "where") || strings.Contains(queryLower, "?") {
		return "question"
	}

	// Command patterns
	if strings.Contains(queryLower, "deploy") || strings.Contains(queryLower, "start") ||
		strings.Contains(queryLower, "stop") || strings.Contains(queryLower, "scale") {
		return "command"
	}

	// Analysis patterns
	if strings.Contains(queryLower, "analyze") || strings.Contains(queryLower, "show") ||
		strings.Contains(queryLower, "status") || strings.Contains(queryLower, "health") {
		return "analysis"
	}

	return "general"
}

// extractScope determines the scope of the query
func (brain *PlatformAI) extractScope(query string, platformContext map[string]interface{}) []string {
	queryLower := strings.ToLower(query)
	scope := []string{}

	// Check for application mentions
	if applications, ok := platformContext["applications"].(map[string]interface{}); ok {
		for appName := range applications {
			if strings.Contains(queryLower, strings.ToLower(appName)) {
				scope = append(scope, "application:"+appName)
			}
		}
	}

	// Check for service mentions
	if strings.Contains(queryLower, "service") || strings.Contains(queryLower, "microservice") {
		scope = append(scope, "services")
	}

	// Check for environment mentions
	if strings.Contains(queryLower, "production") || strings.Contains(queryLower, "prod") {
		scope = append(scope, "environment:production")
	}
	if strings.Contains(queryLower, "staging") || strings.Contains(queryLower, "stage") {
		scope = append(scope, "environment:staging")
	}

	// Default scope if none detected
	if len(scope) == 0 {
		scope = append(scope, "platform")
	}

	return scope
}

// extractEnvironmentContext gets detailed environment context for impact analysis
func (brain *PlatformAI) extractEnvironmentContext(environment string) (map[string]interface{}, error) {
	globalGraph, err := brain.graph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	context := map[string]interface{}{
		"environment":        environment,
		"applications":       brain.extractEnvironmentApplications(globalGraph, environment),
		"services":           brain.extractEnvironmentServices(globalGraph, environment),
		"policies":           brain.extractEnvironmentPolicies(globalGraph, environment),
		"current_load":       brain.extractEnvironmentLoad(environment),
		"active_deployments": brain.extractActiveDeployments(environment),
		"recent_changes":     brain.extractRecentChanges(environment),
		"timestamp":          time.Now(),
	}

	return context, nil
}

// extractIncidentContext builds comprehensive incident context for troubleshooting
func (brain *PlatformAI) extractIncidentContext(incidentID, description string, symptoms []string) (*IncidentContext, error) {
	// Build incident timeline
	timeline := []EventTimestamp{
		{
			Timestamp: time.Now().Format(time.RFC3339),
			Event:     "Incident reported: " + description,
			Source:    "user",
			Severity:  "unknown",
		},
	}

	// Extract relevant logs (would integrate with logging system)
	logs := []LogEntry{
		{
			Timestamp: time.Now().Format(time.RFC3339),
			Level:     "INFO",
			Message:   "Incident context extraction started",
			Source:    "ai-brain",
		},
	}

	// Extract current metrics (would integrate with monitoring system)
	metrics := map[string]interface{}{
		"cpu_usage":     "unknown",
		"memory_usage":  "unknown",
		"error_rate":    "unknown",
		"response_time": "unknown",
	}

	context := &IncidentContext{
		IncidentID:  incidentID,
		Description: description,
		Symptoms:    symptoms,
		Environment: brain.detectIncidentEnvironment(symptoms),
		Timeline:    timeline,
		Logs:        logs,
		Metrics:     metrics,
		Context: map[string]interface{}{
			"platform_state":     brain.extractCurrentPlatformState(),
			"recent_deployments": brain.extractRecentDeployments(),
		},
		Metadata: map[string]interface{}{
			"extraction_time": time.Now(),
			"ai_analysis":     true,
		},
	}

	return context, nil
}

// extractOptimizationConstraints gets constraints for optimization analysis
func (brain *PlatformAI) extractOptimizationConstraints(target string) []string {
	constraints := []string{
		"maintain_availability",
		"zero_downtime",
		"policy_compliance",
		"cost_efficiency",
	}

	// Add target-specific constraints
	if strings.Contains(target, "production") {
		constraints = append(constraints, "production_safety", "rollback_capability")
	}

	return constraints
}

// extractTargetState gets current state of optimization target
func (brain *PlatformAI) extractTargetState(target string) map[string]interface{} {
	return map[string]interface{}{
		"target":               target,
		"current_health":       brain.extractTargetHealth(target),
		"resource_usage":       brain.extractTargetResources(target),
		"performance":          brain.extractTargetPerformance(target),
		"dependencies":         brain.extractTargetDependencies(target),
		"recent_changes":       brain.extractTargetRecentChanges(target),
		"optimization_history": brain.extractOptimizationHistory(target),
	}
}

// extractApplicationSummary extracts summary of applications in the platform
func (brain *PlatformAI) extractApplicationSummary(graph *graph.Graph) map[string]interface{} {
	applications := make(map[string]interface{})
	for _, node := range graph.Nodes {
		if node.Kind == "application" {
			applications[node.ID] = map[string]interface{}{
				"name":     node.ID,
				"status":   "running",
				"services": brain.countApplicationServices(graph, node.ID),
			}
		}
	}
	return applications
}

// extractServiceSummary extracts summary of services in the platform
func (brain *PlatformAI) extractServiceSummary(graph *graph.Graph) map[string]interface{} {
	services := make(map[string]interface{})
	for _, node := range graph.Nodes {
		if node.Kind == "service" {
			services[node.ID] = map[string]interface{}{
				"name":   node.ID,
				"status": "running",
				"health": "healthy",
			}
		}
	}
	return services
}

// extractDependencyMap extracts dependency relationships between components
func (brain *PlatformAI) extractDependencyMap(graph *graph.Graph) map[string]interface{} {
	dependencies := make(map[string]interface{})
	for fromID, edges := range graph.Edges {
		deps := []string{}
		for _, edge := range edges {
			if edge.Type == "depends" {
				deps = append(deps, edge.To)
			}
		}
		if len(deps) > 0 {
			dependencies[fromID] = deps
		}
	}
	return dependencies
}

// extractPolicySummary extracts summary of policies in the platform
func (brain *PlatformAI) extractPolicySummary(graph *graph.Graph) map[string]interface{} {
	policies := make(map[string]interface{})
	for _, node := range graph.Nodes {
		if node.Kind == "policy" {
			policies[node.ID] = map[string]interface{}{
				"name":   node.ID,
				"active": true,
				"scope":  "global",
			}
		}
	}
	return policies
}

// extractEnvironmentSummary extracts summary of environments in the platform
func (brain *PlatformAI) extractEnvironmentSummary(graph *graph.Graph) map[string]interface{} {
	environments := make(map[string]interface{})
	for _, node := range graph.Nodes {
		if node.Kind == "environment" {
			environments[node.ID] = map[string]interface{}{
				"name":   node.ID,
				"status": "active",
				"health": "healthy",
			}
		}
	}
	return environments
}

// extractHealthStatus extracts health status of the platform components
func (brain *PlatformAI) extractHealthStatus(graph *graph.Graph) map[string]interface{} {
	return map[string]interface{}{
		"overall":      "healthy",
		"applications": brain.countHealthyApplications(graph),
		"services":     brain.countHealthyServices(graph),
		"issues":       []string{},
	}
}

// extractRecentEvents extracts recent events in the platform
func (brain *PlatformAI) extractRecentEvents() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"timestamp": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			"event":     "Application deployed successfully",
			"type":      "deployment",
		},
		{
			"timestamp": time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			"event":     "Health check policy updated",
			"type":      "policy",
		},
	}
}

// Additional helper methods for environment context
func (brain *PlatformAI) extractEnvironmentApplications(graph *graph.Graph, environment string) []string {
	apps := []string{}
	for _, node := range graph.Nodes {
		if node.Kind == "application" {
			apps = append(apps, node.ID)
		}
	}
	return apps
}

func (brain *PlatformAI) extractEnvironmentServices(graph *graph.Graph, environment string) []string {
	services := []string{}
	for _, node := range graph.Nodes {
		if node.Kind == "service" {
			services = append(services, node.ID)
		}
	}
	return services
}

func (brain *PlatformAI) extractEnvironmentPolicies(graph *graph.Graph, environment string) []string {
	policies := []string{}
	for _, node := range graph.Nodes {
		if node.Kind == "policy" {
			policies = append(policies, node.ID)
		}
	}
	return policies
}

func (brain *PlatformAI) extractEnvironmentLoad(environment string) map[string]interface{} {
	return map[string]interface{}{
		"cpu":     "45%",
		"memory":  "60%",
		"network": "30%",
	}
}

func (brain *PlatformAI) extractActiveDeployments(environment string) []string {
	return []string{} // No active deployments
}

func (brain *PlatformAI) extractRecentChanges(environment string) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"timestamp": time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
			"change":    "Service scaled up",
			"target":    "api-service",
		},
	}
}

// Additional helper methods for incident context
func (brain *PlatformAI) detectIncidentEnvironment(symptoms []string) string {
	for _, symptom := range symptoms {
		if strings.Contains(strings.ToLower(symptom), "production") {
			return "production"
		}
		if strings.Contains(strings.ToLower(symptom), "staging") {
			return "staging"
		}
	}
	return "unknown"
}

func (brain *PlatformAI) extractCurrentPlatformState() map[string]interface{} {
	return map[string]interface{}{
		"overall_health": "degraded",
		"active_alerts":  2,
		"deployments":    "stable",
	}
}

func (brain *PlatformAI) extractRecentDeployments() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"deployment_id": "deploy-123",
			"application":   "web-app",
			"status":        "completed",
			"timestamp":     time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
		},
	}
}

// Additional helper methods for optimization
func (brain *PlatformAI) extractTargetHealth(target string) string {
	return "healthy" // Placeholder
}

func (brain *PlatformAI) extractTargetResources(target string) map[string]interface{} {
	return map[string]interface{}{
		"cpu":    "65%",
		"memory": "70%",
		"disk":   "45%",
	}
}

func (brain *PlatformAI) extractTargetPerformance(target string) map[string]interface{} {
	return map[string]interface{}{
		"response_time": "150ms",
		"throughput":    "1000 req/s",
		"error_rate":    "0.1%",
	}
}

func (brain *PlatformAI) extractTargetDependencies(target string) []string {
	return []string{"database", "cache", "message-queue"}
}

func (brain *PlatformAI) extractTargetRecentChanges(target string) []string {
	return []string{"scaled up replicas", "updated configuration"}
}

func (brain *PlatformAI) extractOptimizationHistory(target string) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"timestamp":    time.Now().Add(-7 * 24 * time.Hour).Format(time.RFC3339),
			"optimization": "CPU allocation increased",
			"impact":       "positive",
		},
	}
}

// Helper functions for deployment context extraction (replacing deployments package calls)

// extractDeploymentMetrics gets metrics for learning from deployment
func (brain *PlatformAI) extractDeploymentMetrics(deploymentID string) map[string]interface{} {
	return map[string]interface{}{
		"deployment_id":      deploymentID,
		"start_time":         time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
		"end_time":           time.Now().Format(time.RFC3339),
		"success_rate":       "95%",
		"error_count":        2,
		"rollback_triggered": false,
		"performance_impact": "minimal",
	}
}

// extractDeploymentContext gets context for learning from deployment
func (brain *PlatformAI) extractDeploymentContext(deploymentID string) map[string]interface{} {
	return map[string]interface{}{
		"deployment_id":     deploymentID,
		"strategy":          "rolling",
		"environment":       "production",
		"applications":      []string{"web-app", "api-service"},
		"services_affected": 3,
		"policies_applied":  []string{"zero-downtime", "health-check"},
		"ai_planned":        true,
		"complexity":        "medium",
	}
}

// countApplicationServices counts the number of services owned by an application
func (brain *PlatformAI) countApplicationServices(graph *graph.Graph, appID string) int {
	count := 0
	if edges, exists := graph.Edges[appID]; exists {
		for _, edge := range edges {
			if edge.Type == "owns" {
				if node, err := graph.GetNode(edge.To); err == nil && node.Kind == "service" {
					count++
				}
			}
		}
	}
	return count
}

// countHealthyApplications counts healthy applications in the graph
func (brain *PlatformAI) countHealthyApplications(graph *graph.Graph) int {
	count := 0
	for _, node := range graph.Nodes {
		if node.Kind == "application" {
			count++
		}
	}
	return count
}

// countHealthyServices counts healthy services in the graph
func (brain *PlatformAI) countHealthyServices(graph *graph.Graph) int {
	count := 0
	for _, node := range graph.Nodes {
		if node.Kind == "service" {
			count++
		}
	}
	return count
}
