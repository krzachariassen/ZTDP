package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// === SIMPLIFIED DEPLOYMENT REQUEST TYPES ===

// DeploymentRequest represents a basic deployment request that will be handled by orchestrator
type DeploymentRequest struct {
	Application string                 `json:"application"`
	Environment string                 `json:"environment"`
	Version     string                 `json:"version,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

// PlanOptimizationRequest represents a request to optimize a deployment plan
type PlanOptimizationRequest struct {
	Application string                 `json:"application"`
	CurrentPlan map[string]interface{} `json:"current_plan"`
}

// ImpactAnalysisRequest represents a request to analyze deployment impact
type ImpactAnalysisRequest struct {
	Application string                 `json:"application"`
	Environment string                 `json:"environment"`
	Changes     map[string]interface{} `json:"changes"`
}

// === DEPLOYMENT HANDLERS USING ORCHESTRATOR ===

// GenerateDeploymentPlan generates deployment plans using the orchestrator chat interface
// @Summary      Generate AI deployment plan
// @Description  Uses the orchestrator to generate intelligent deployment plans
// @Tags         deployments
// @Accept       json
// @Produce      json
// @Param        request body DeploymentRequest true "Deployment request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /v1/deployments/plan/generate [post]
func GenerateDeploymentPlan(w http.ResponseWriter, r *http.Request) {
	var req DeploymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Application == "" {
		WriteJSONError(w, "Application name is required", http.StatusBadRequest)
		return
	}

	if req.Environment == "" {
		WriteJSONError(w, "Environment is required", http.StatusBadRequest)
		return
	}

	// Use orchestrator for deployment planning
	orchestrator := GetGlobalOrchestrator()
	if orchestrator == nil {
		WriteJSONError(w, "Orchestrator not available", http.StatusServiceUnavailable)
		return
	}

	// Create natural language request for orchestrator
	message := fmt.Sprintf("Generate a deployment plan for application '%s' to environment '%s'", req.Application, req.Environment)
	if req.Version != "" {
		message += fmt.Sprintf(" version '%s'", req.Version)
	}

	response, err := orchestrator.Chat(r.Context(), message)
	if err != nil {
		WriteJSONError(w, "Failed to generate deployment plan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"message":  response.Message,
		"intent":   response.Intent,
		"actions":  response.Actions,
		"request":  req,
	})
}

// OptimizeDeploymentPlan optimizes deployment plans using the orchestrator
// @Summary      Optimize deployment plan with AI
// @Description  Uses the orchestrator to optimize existing deployment plans
// @Tags         deployments
// @Accept       json
// @Produce      json
// @Param        request body PlanOptimizationRequest true "Plan optimization request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /v1/deployments/plan/optimize [post]
func OptimizeDeploymentPlan(w http.ResponseWriter, r *http.Request) {
	var req PlanOptimizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Application == "" {
		WriteJSONError(w, "Application name is required", http.StatusBadRequest)
		return
	}

	// Use orchestrator for plan optimization
	orchestrator := GetGlobalOrchestrator()
	if orchestrator == nil {
		WriteJSONError(w, "Orchestrator not available", http.StatusServiceUnavailable)
		return
	}

	// Create natural language request for orchestrator
	message := fmt.Sprintf("Optimize the deployment plan for application '%s'. Current plan: %+v", req.Application, req.CurrentPlan)

	response, err := orchestrator.Chat(r.Context(), message)
	if err != nil {
		WriteJSONError(w, "Failed to optimize deployment plan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"message":  response.Message,
		"intent":   response.Intent,
		"actions":  response.Actions,
		"request":  req,
	})
}

// AnalyzeDeploymentImpact analyzes deployment impact using the orchestrator
// @Summary      Analyze deployment impact
// @Description  Uses the orchestrator to analyze potential deployment impact
// @Tags         deployments
// @Accept       json
// @Produce      json
// @Param        request body ImpactAnalysisRequest true "Impact analysis request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /v1/deployments/impact/analyze [post]
func AnalyzeDeploymentImpact(w http.ResponseWriter, r *http.Request) {
	var req ImpactAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Application == "" {
		WriteJSONError(w, "Application name is required", http.StatusBadRequest)
		return
	}

	if req.Environment == "" {
		WriteJSONError(w, "Environment is required", http.StatusBadRequest)
		return
	}

	// Use orchestrator for impact analysis
	orchestrator := GetGlobalOrchestrator()
	if orchestrator == nil {
		WriteJSONError(w, "Orchestrator not available", http.StatusServiceUnavailable)
		return
	}

	// Create natural language request for orchestrator
	message := fmt.Sprintf("Analyze the impact of deploying application '%s' to environment '%s' with changes: %+v", 
		req.Application, req.Environment, req.Changes)

	response, err := orchestrator.Chat(r.Context(), message)
	if err != nil {
		WriteJSONError(w, "Failed to analyze deployment impact: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"message":  response.Message,
		"intent":   response.Intent,
		"actions":  response.Actions,
		"request":  req,
	})
}

// ExecuteDeployment executes deployments using the orchestrator
// @Summary      Execute deployment
// @Description  Uses the orchestrator to execute deployments
// @Tags         deployments
// @Accept       json
// @Produce      json
// @Param        app          path   string  true  "Application name"
// @Param        environment  path   string  true  "Environment name"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /v1/deployments/{app}/{environment}/execute [post]
func ExecuteDeployment(w http.ResponseWriter, r *http.Request) {
	app := chi.URLParam(r, "app")
	environment := chi.URLParam(r, "environment")

	if app == "" {
		WriteJSONError(w, "Application name is required", http.StatusBadRequest)
		return
	}

	if environment == "" {
		WriteJSONError(w, "Environment is required", http.StatusBadRequest)
		return
	}

	// Use orchestrator for deployment execution
	orchestrator := GetGlobalOrchestrator()
	if orchestrator == nil {
		WriteJSONError(w, "Orchestrator not available", http.StatusServiceUnavailable)
		return
	}

	// Create natural language request for orchestrator
	message := fmt.Sprintf("Deploy application '%s' to environment '%s'", app, environment)

	response, err := orchestrator.Chat(r.Context(), message)
	if err != nil {
		WriteJSONError(w, "Failed to execute deployment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "success",
		"message":     response.Message,
		"intent":      response.Intent,
		"actions":     response.Actions,
		"application": app,
		"environment": environment,
	})
}

// GetDeploymentStatus gets deployment status using the orchestrator
// @Summary      Get deployment status
// @Description  Uses the orchestrator to get deployment status
// @Tags         deployments
// @Produce      json
// @Param        app          path   string  true  "Application name"
// @Param        environment  path   string  true  "Environment name"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /v1/deployments/{app}/{environment}/status [get]
func GetDeploymentStatus(w http.ResponseWriter, r *http.Request) {
	app := chi.URLParam(r, "app")
	environment := chi.URLParam(r, "environment")

	if app == "" {
		WriteJSONError(w, "Application name is required", http.StatusBadRequest)
		return
	}

	if environment == "" {
		WriteJSONError(w, "Environment is required", http.StatusBadRequest)
		return
	}

	// Use orchestrator for status checking
	orchestrator := GetGlobalOrchestrator()
	if orchestrator == nil {
		WriteJSONError(w, "Orchestrator not available", http.StatusServiceUnavailable)
		return
	}

	// Create natural language request for orchestrator
	message := fmt.Sprintf("What is the deployment status of application '%s' in environment '%s'?", app, environment)

	response, err := orchestrator.Chat(r.Context(), message)
	if err != nil {
		WriteJSONError(w, "Failed to get deployment status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "success",
		"message":     response.Message,
		"intent":      response.Intent,
		"actions":     response.Actions,
		"application": app,
		"environment": environment,
	})
}

// ListDeployments lists deployments using the orchestrator
// @Summary      List deployments
// @Description  Uses the orchestrator to list deployments
// @Tags         deployments
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /v1/deployments [get]
func ListDeployments(w http.ResponseWriter, r *http.Request) {
	// Use orchestrator for listing deployments
	orchestrator := GetGlobalOrchestrator()
	if orchestrator == nil {
		WriteJSONError(w, "Orchestrator not available", http.StatusServiceUnavailable)
		return
	}

	// Create natural language request for orchestrator
	message := "List all deployments and their current status"

	response, err := orchestrator.Chat(r.Context(), message)
	if err != nil {
		WriteJSONError(w, "Failed to list deployments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": response.Message,
		"intent":  response.Intent,
		"actions": response.Actions,
	})
}
