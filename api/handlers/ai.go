package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/krzachariassen/ZTDP/internal/ai"
)

// AIGeneratePlanRequest represents the request for AI plan generation
type AIGeneratePlanRequest struct {
	AppName   string   `json:"app_name"`
	EdgeTypes []string `json:"edge_types,omitempty"`
	Timeout   int      `json:"timeout,omitempty"` // Timeout in seconds
}

// AIEvaluatePolicyRequest represents the request for AI policy evaluation
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

	// Default edge types
	if len(req.EdgeTypes) == 0 {
		req.EdgeTypes = []string{"deploy", "create", "owns"}
	}

	// Default timeout
	timeout := 30 * time.Second
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	// Create AI brain
	brain, err := ai.NewAIBrainFromConfig(GlobalGraph)
	if err != nil {
		WriteJSONError(w, "AI service unavailable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Generate AI plan
	planResponse, err := brain.GenerateDeploymentPlan(ctx, req.AppName, req.EdgeTypes)
	if err != nil {
		WriteJSONError(w, "AI plan generation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(planResponse)
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
	brain, err := ai.NewAIBrainFromConfig(GlobalGraph)
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
	brain, err := ai.NewAIBrainFromConfig(GlobalGraph)
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
	brain, err := ai.NewAIBrainFromConfig(GlobalGraph)

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

// Helper function to get environment variable with default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
