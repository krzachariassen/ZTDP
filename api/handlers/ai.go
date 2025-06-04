package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/deployments"
)

// AIGeneratePlanRequest represents the request for AI plan generation
type AIGeneratePlanRequest struct {
	AppName   string   `json:"app_name"`
	EdgeTypes []string `json:"edge_types,omitempty"`
	Timeout   int      `json:"timeout,omitempty"` // Timeout in seconds
	// Default timeout
	timeout := 60 * time.Second
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	// Create AI brain
	brain, err := ai.NewPlatformAIFromConfig(GlobalGraph)
	if err != nil {
		WriteJSONError(w, "AI service unavailable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}uatePolicyRequest represents the request for AI policy evaluation
type AIEvaluatePolicyRequest struct {
	ApplicationID string `json:"application_id"`
	EnvironmentID string `json:"environment_id"`
}

// AIOptimizePlanRequest represents the request for AI plan optimization
type AIOptimizePlanRequest struct {
	CurrentPlan   []ai.DeploymentStep `json:"current_plan"`
	ApplicationID string              `json:"application_id"`
}

// AIProviderInfo represents AI provider information
type AIProviderInfo struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Capabilities []string          `json:"capabilities"`
	Available    bool              `json:"available"`
	Model        string            `json:"model,omitempty"`
	Config       map[string]string `json:"config,omitempty"`
}

// AIGeneratePlan godoc
// @Summary      Generate AI deployment plan
// @Description  Uses AI to generate an intelligent deployment plan for an application
// @Tags         ai
// @Accept       json
// @Produce      json
// @Param        request  body  AIGeneratePlanRequest  true  "AI plan generation request"
// @Success      200  {object}  ai.PlanResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /v1/ai/plans/generate [post]
func AIGeneratePlan(w http.ResponseWriter, r *http.Request) {
	var req AIGeneratePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.AppName == "" {
		WriteJSONError(w, "app_name is required", http.StatusBadRequest)
		return
	}

	// Default timeout
	timeout := 30 * time.Second
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Use deployment service instead of AI Brain
	deploymentService, err := deployments.NewService(GlobalGraph, nil) // AI provider will be initialized internally
	if err != nil {
		WriteJSONError(w, "Deployment service unavailable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Generate deployment plan using domain service
	plan, err := deploymentService.GenerateDeploymentPlan(ctx, req.AppName, req.EdgeTypes)
	if err != nil {
		WriteJSONError(w, "Deployment plan generation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to legacy response format for backward compatibility
	planResponse := &ai.PlanningResponse{
		Plan: &ai.DeploymentPlan{
			ApplicationName: req.AppName,
			Steps:           convertPlanToDeploymentSteps(plan.Services),
		},
		Reasoning:  plan.Reasoning,
		Confidence: plan.Confidence,
		Metadata: map[string]interface{}{
			"total_services": len(plan.Services),
			"strategy":       plan.Strategy,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(planResponse)
}

// Helper function to convert deployment results to legacy format
func convertResultToDeploymentSteps(deployments []string) []*ai.DeploymentStep {
	steps := make([]*ai.DeploymentStep, len(deployments))
	for i, serviceName := range deployments {
		steps[i] = &ai.DeploymentStep{
			ServiceName: serviceName,
			Action:      "deploy",
			Metadata: map[string]interface{}{
				"position": i + 1,
				"total":    len(deployments),
			},
		}
	}
	return steps
}

// AIEvaluatePolicy godoc
// @Summary      Evaluate deployment policies using AI
// @Description  Uses AI to evaluate deployment policies for an application and environment
// @Tags         ai
// @Accept       json
// @Produce      json
// @Param        request  body  AIEvaluatePolicyRequest  true  "AI policy evaluation request"
// @Success      200  {object}  ai.PolicyEvaluation
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /v1/ai/policies/evaluate [post]
func AIEvaluatePolicy(w http.ResponseWriter, r *http.Request) {
	var req AIEvaluatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.ApplicationID == "" {
		WriteJSONError(w, "application_id is required", http.StatusBadRequest)
		return
	}

	if req.EnvironmentID == "" {
		WriteJSONError(w, "environment_id is required", http.StatusBadRequest)
		return
	}

	// Create AI brain
	brain, err := ai.NewPlatformAIFromConfig(GlobalGraph)
	if err != nil {
		WriteJSONError(w, "AI service unavailable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Evaluate policies with AI
	evaluation, err := brain.EvaluateDeploymentPolicies(ctx, req.ApplicationID, req.EnvironmentID)
	if err != nil {
		WriteJSONError(w, "AI policy evaluation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(evaluation)
}

// AIOptimizePlan godoc
// @Summary      Optimize deployment plan using AI
// @Description  Uses AI to optimize an existing deployment plan for better performance
// @Tags         ai
// @Accept       json
// @Produce      json
// @Param        request  body  AIOptimizePlanRequest  true  "AI plan optimization request"
// @Success      200  {object}  ai.PlanningResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /v1/ai/plans/optimize [post]
func AIOptimizePlan(w http.ResponseWriter, r *http.Request) {
	var req AIOptimizePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.CurrentPlan) == 0 {
		WriteJSONError(w, "current_plan is required", http.StatusBadRequest)
		return
	}

	if req.ApplicationID == "" {
		WriteJSONError(w, "application_id is required", http.StatusBadRequest)
		return
	}

	// Create AI brain
	brain, err := ai.NewPlatformAIFromConfig(GlobalGraph)
	if err != nil {
		WriteJSONError(w, "AI service unavailable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create deployment plan from the current plan steps
	// Convert []DeploymentStep to []*DeploymentStep
	steps := make([]*ai.DeploymentStep, len(req.CurrentPlan))
	for i := range req.CurrentPlan {
		steps[i] = &req.CurrentPlan[i]
	}

	plan := &ai.DeploymentPlan{
		Steps:    steps,
		Strategy: "ai_optimization",
		Metadata: map[string]interface{}{
			"optimization_request": true,
			"timestamp":            time.Now(),
		},
	}

	// Optimize plan with AI
	optimization, err := brain.OptimizeExistingPlan(ctx, plan, req.ApplicationID)
	if err != nil {
		WriteJSONError(w, "AI plan optimization failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(optimization)
}

// AIProviderStatus godoc
// @Summary      Get AI provider status
// @Description  Returns information about the current AI provider configuration and availability
// @Tags         ai
// @Produce      json
// @Success      200  {object}  AIProviderInfo
// @Failure      500  {object}  map[string]string
// @Router       /v1/ai/provider/status [get]
func AIProviderStatus(w http.ResponseWriter, r *http.Request) {
	// Try to create AI brain to check availability
	brain, err := ai.NewPlatformAIFromConfig(GlobalGraph)

	providerInfo := AIProviderInfo{
		Available:    false,
		Capabilities: []string{"plan_generation", "policy_evaluation", "plan_optimization"},
	}

	if err != nil {
		providerInfo.Name = "OpenAI (Unavailable)"
		providerInfo.Config = map[string]string{
			"error": err.Error(),
		}
	} else {
		// Get provider info from AI brain
		info := brain.GetProviderInfo()
		providerInfo.Name = info.Name
		providerInfo.Version = info.Version
		providerInfo.Capabilities = append(providerInfo.Capabilities, info.Capabilities...)
		providerInfo.Available = true

		// Get model from environment or config
		if model := getEnvOrDefault("OPENAI_MODEL", "gpt-4"); model != "" {
			providerInfo.Model = model
		}

		providerInfo.Config = map[string]string{
			"base_url": getEnvOrDefault("OPENAI_BASE_URL", "https://api.openai.com/v1"),
			"model":    providerInfo.Model,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(providerInfo)
}

// AIMetrics godoc
// @Summary      Get AI performance metrics
// @Description  Returns performance metrics for AI operations
// @Tags         ai
// @Produce      json
// @Param        hours  query  int  false  "Number of hours to look back (default: 24)"
// @Success      200  {object}  map[string]interface{}
// @Router       /v1/ai/metrics [get]
func AIMetrics(w http.ResponseWriter, r *http.Request) {
	hours := 24 // Default to last 24 hours
	if h := r.URL.Query().Get("hours"); h != "" {
		if parsed, err := strconv.Atoi(h); err == nil && parsed > 0 {
			hours = parsed
		}
	}

	// For now, return placeholder metrics
	// TODO: Implement actual metrics collection from AI operations
	metrics := map[string]interface{}{
		"timeframe_hours": hours,
		"plan_generation": map[string]interface{}{
			"total_requests":    0,
			"successful":        0,
			"failed":            0,
			"avg_response_time": "0ms",
			"success_rate":      "0%",
		},
		"policy_evaluation": map[string]interface{}{
			"total_requests":    0,
			"successful":        0,
			"failed":            0,
			"avg_response_time": "0ms",
			"success_rate":      "0%",
		},
		"plan_optimization": map[string]interface{}{
			"total_requests":    0,
			"successful":        0,
			"failed":            0,
			"avg_response_time": "0ms",
			"success_rate":      "0%",
		},
		"note": "Metrics collection is not yet implemented. This endpoint returns placeholder data.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// *** REVOLUTIONARY AI API ENDPOINTS ***
// These endpoints demonstrate groundbreaking AI capabilities impossible with traditional IDPs

// AIChatRequest represents the request for conversational AI
type AIChatRequest struct {
	Query   string   `json:"query"`
	Context string   `json:"context,omitempty"`
	Scope   []string `json:"scope,omitempty"`
	Session string   `json:"session,omitempty"`
	Timeout int      `json:"timeout,omitempty"`
}

// AIImpactRequest represents the request for impact prediction
type AIImpactRequest struct {
	Changes     []map[string]interface{} `json:"changes"`
	Environment string                   `json:"environment"`
	Scope       string                   `json:"scope,omitempty"`
	Timeframe   string                   `json:"timeframe,omitempty"`
	Timeout     int                      `json:"timeout,omitempty"`
}

// AITroubleshootRequest represents the request for intelligent troubleshooting
type AITroubleshootRequest struct {
	IncidentID  string                   `json:"incident_id"`
	Description string                   `json:"description"`
	Symptoms    []string                 `json:"symptoms,omitempty"`
	Timeline    []map[string]interface{} `json:"timeline,omitempty"`
	Logs        []string                 `json:"logs,omitempty"`
	Metrics     map[string]interface{}   `json:"metrics,omitempty"`
	Environment string                   `json:"environment,omitempty"`
	Timeout     int                      `json:"timeout,omitempty"`
}

// AIOptimizeRequest represents the request for proactive optimization
type AIOptimizeRequest struct {
	Target      string                 `json:"target"`
	Areas       []string               `json:"areas,omitempty"`
	Constraints map[string]interface{} `json:"constraints,omitempty"`
	Goals       []string               `json:"goals,omitempty"`
	Timeout     int                    `json:"timeout,omitempty"`
}

// AILearnRequest represents the request for learning from deployments
type AILearnRequest struct {
	DeploymentID string                 `json:"deployment_id"`
	Success      bool                   `json:"success"`
	Duration     string                 `json:"duration,omitempty"`
	Issues       []string               `json:"issues,omitempty"`
	Metrics      map[string]interface{} `json:"metrics,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
	Timeout      int                    `json:"timeout,omitempty"`
}

// AIChatWithPlatform godoc
// @Summary      Chat with Platform using AI
// @Description  Revolutionary conversational AI that allows natural language interaction with platform graph for insights and actions
// @Tags         ai,revolutionary
// @Accept       json
// @Produce      json
// @Param        request  body  AIChatRequest  true  "Chat request"
// @Success      200  {object}  ai.ConversationalResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /v1/ai/chat [post]
func AIChatWithPlatform(w http.ResponseWriter, r *http.Request) {
	var req AIChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		WriteJSONError(w, "query is required", http.StatusBadRequest)
		return
	}

	// Default timeout
	timeout := 60 * time.Second
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	// Create AI brain
	brain, err := ai.NewPlatformAIFromConfig(GlobalGraph)
	if err != nil {
		WriteJSONError(w, "AI service unavailable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Chat with platform using revolutionary AI
	response, err := brain.ChatWithPlatform(ctx, req.Query, req.Context)
	if err != nil {
		WriteJSONError(w, "Conversational AI failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AIPredictImpact godoc
// @Summary      Predict deployment impact using AI
// @Description  Revolutionary AI-driven impact prediction that simulates deployment consequences before they happen
// @Tags         ai,revolutionary
// @Accept       json
// @Produce      json
// @Param        request  body  AIImpactRequest  true  "Impact prediction request"
// @Success      200  {object}  ai.ImpactPrediction
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /v1/ai/predict-impact [post]
func AIPredictImpact(w http.ResponseWriter, r *http.Request) {
	var req AIImpactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.Changes) == 0 {
		WriteJSONError(w, "changes are required", http.StatusBadRequest)
		return
	}

	if req.Environment == "" {
		WriteJSONError(w, "environment is required", http.StatusBadRequest)
		return
	}

	// Default timeout
	timeout := 90 * time.Second
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	// Create AI brain
	brain, err := ai.NewPlatformAIFromConfig(GlobalGraph)
	if err != nil {
		WriteJSONError(w, "AI service unavailable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Convert API request to internal format
	changes := make([]ai.ProposedChange, len(req.Changes))
	for i, change := range req.Changes {
		// Extract fields from map with safe defaults
		changeType, _ := change["type"].(string)
		target, _ := change["target"].(string)

		changes[i] = ai.ProposedChange{
			Type:     changeType,
			Target:   target,
			Details:  change,
			Metadata: change,
		}
	}

	// Predict impact using revolutionary AI
	prediction, err := brain.PredictDeploymentImpact(ctx, changes, req.Environment)
	if err != nil {
		WriteJSONError(w, "Impact prediction failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prediction)
}

// AITroubleshoot godoc
// @Summary      Intelligent troubleshooting using AI
// @Description  Revolutionary AI-driven root cause analysis and intelligent problem diagnosis beyond traditional monitoring
// @Tags         ai,revolutionary
// @Accept       json
// @Produce      json
// @Param        request  body  AITroubleshootRequest  true  "Troubleshooting request"
// @Success      200  {object}  ai.TroubleshootingResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /v1/ai/troubleshoot [post]
func AITroubleshoot(w http.ResponseWriter, r *http.Request) {
	var req AITroubleshootRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.IncidentID == "" {
		WriteJSONError(w, "incident_id is required", http.StatusBadRequest)
		return
	}

	if req.Description == "" {
		WriteJSONError(w, "description is required", http.StatusBadRequest)
		return
	}

	// Default timeout
	timeout := 120 * time.Second
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	// Create AI brain
	brain, err := ai.NewPlatformAIFromConfig(GlobalGraph)
	if err != nil {
		WriteJSONError(w, "AI service unavailable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Perform intelligent troubleshooting using revolutionary AI
	response, err := brain.IntelligentTroubleshooting(ctx, req.IncidentID, req.Description, req.Symptoms)
	if err != nil {
		WriteJSONError(w, "Intelligent troubleshooting failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AIProactiveOptimize godoc
// @Summary      Proactive optimization using AI
// @Description  Revolutionary AI that continuously analyzes patterns and recommends architectural optimizations before problems occur
// @Tags         ai,revolutionary
// @Accept       json
// @Produce      json
// @Param        request  body  AIOptimizeRequest  true  "Proactive optimization request"
// @Success      200  {object}  ai.OptimizationRecommendations
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /v1/ai/proactive-optimize [post]
func AIProactiveOptimize(w http.ResponseWriter, r *http.Request) {
	var req AIOptimizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Target == "" {
		WriteJSONError(w, "target is required", http.StatusBadRequest)
		return
	}

	// Default timeout
	timeout := 120 * time.Second
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	// Create AI brain
	brain, err := ai.NewPlatformAIFromConfig(GlobalGraph)
	if err != nil {
		WriteJSONError(w, "AI service unavailable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Default focus areas if not provided
	focus := req.Areas
	if len(focus) == 0 {
		focus = []string{"performance", "reliability", "security", "cost"}
	}

	// Perform proactive optimization using revolutionary AI
	recommendations, err := brain.ProactiveOptimization(ctx, req.Target, focus)
	if err != nil {
		WriteJSONError(w, "Proactive optimization failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recommendations)
}

// AILearnFromDeployment godoc
// @Summary      Learn from deployment outcomes using AI
// @Description  Revolutionary AI that learns from deployment outcomes to continuously improve future deployments
// @Tags         ai,revolutionary
// @Accept       json
// @Produce      json
// @Param        request  body  AILearnRequest  true  "Learning request"
// @Success      200  {object}  ai.LearningInsights
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /v1/ai/learn [post]
func AILearnFromDeployment(w http.ResponseWriter, r *http.Request) {
	var req AILearnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.DeploymentID == "" {
		WriteJSONError(w, "deployment_id is required", http.StatusBadRequest)
		return
	}

	// Default timeout
	timeout := 90 * time.Second
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	// Create AI brain
	brain, err := ai.NewPlatformAIFromConfig(GlobalGraph)
	if err != nil {
		WriteJSONError(w, "AI service unavailable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Parse duration if provided
	var duration int64 = 0
	if req.Duration != "" {
		if parsedDuration, err := time.ParseDuration(req.Duration); err == nil {
			duration = int64(parsedDuration.Seconds())
		}
	}

	// Convert issues to DeploymentIssue structs
	issues := make([]ai.DeploymentIssue, len(req.Issues))
	for i, issue := range req.Issues {
		issues[i] = ai.DeploymentIssue{
			Type:        "error", // Default type
			Description: issue,
			Severity:    "medium",  // Default severity
			Resolution:  "pending", // Default resolution
			Timestamp:   time.Now().Format(time.RFC3339),
		}
	}

	// Learn from deployment using revolutionary AI
	insights, err := brain.LearnFromDeployment(ctx, req.DeploymentID, req.Success, duration, issues)
	if err != nil {
		WriteJSONError(w, "Learning from deployment failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(insights)
}

// Helper function to get environment variable with fallback
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
