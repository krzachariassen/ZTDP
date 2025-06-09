package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/deployments"
)

// === AI-ENHANCED DEPLOYMENT REQUEST TYPES ===

// AIOptimizePlanRequest represents the request for AI plan optimization
type AIOptimizePlanRequest struct {
	CurrentPlan   []ai.DeploymentStep `json:"current_plan"`
	ApplicationID string              `json:"application_id"`
}

// AIGeneratePlanRequest represents the request for AI plan generation
type AIGeneratePlanRequest struct {
	AppName   string   `json:"app_name"`
	EdgeTypes []string `json:"edge_types,omitempty"`
	Timeout   int      `json:"timeout,omitempty"` // Timeout in seconds
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

// === EXISTING REQUEST TYPES ===

// DeployApplicationRequest represents the payload for application deployment
type DeployApplicationRequest struct {
	Environment string `json:"environment"`
	Version     string `json:"version,omitempty"` // Optional: future feature for specific app versions
}

// PredictImpactRequest represents the request for deployment impact analysis
type PredictImpactRequest struct {
	Changes     []map[string]interface{} `json:"changes"`
	Environment string                   `json:"environment"`
}

// TroubleshootRequest represents the request for deployment troubleshooting
type TroubleshootRequest struct {
	IncidentID  string   `json:"incident_id"`
	Description string   `json:"description"`
	Symptoms    []string `json:"symptoms,omitempty"`
}

// OptimizeRequest represents the request for deployment optimization
type OptimizeRequest struct {
	Target     string   `json:"target"`
	FocusAreas []string `json:"focus_areas,omitempty"`
}

// DeployApplication godoc
// @Summary      Deploy an application to an environment (with AI-powered planning)
// @Description  Deploys all services of an application to the specified environment. Supports preview operations via query parameters: ?plan=true (generate plan), ?dry-run=true (preview), ?optimize=true (optimize plan), ?analyze=true (impact analysis). AI planning is integrated automatically.
// @Tags         deployments
// @Accept       json
// @Produce      json
// @Param        app_name     path  string                      true  "Application name"
// @Param        deployment   body  DeployApplicationRequest    true  "Deployment request"
// @Param        plan         query bool                       false "Return deployment plan without executing (preview mode)"
// @Param        dry-run      query bool                       false "Preview deployment without executing (alias for plan)"
// @Param        optimize     query bool                       false "Generate optimized deployment plan"
// @Param        analyze      query bool                       false "Include impact analysis in plan"
// @Success      200  {object}  deployments.DeploymentResult
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name}/deploy [post]
func DeployApplication(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "app_name")

	var req DeployApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Environment == "" {
		WriteJSONError(w, "Environment is required", http.StatusBadRequest)
		return
	}

	// Check for preview/planning query parameters
	queryParams := r.URL.Query()
	isPlan := queryParams.Get("plan") == "true"
	isDryRun := queryParams.Get("dry-run") == "true"
	isOptimize := queryParams.Get("optimize") == "true"
	isAnalyze := queryParams.Get("analyze") == "true"

	// Create AI platform agent - AI is now required for all deployment operations (AI-native platform)
	agent := GetGlobalV3Agent()
	var aiProvider ai.AIProvider
	if agent != nil {
		aiProvider = agent.Provider()
	}

	// Create deployment service
	service := deployments.NewDeploymentService(GlobalGraph, aiProvider)

	// Handle preview operations (don't actually deploy)
	if isPlan || isDryRun {
		plan, err := service.GenerateDeploymentPlan(r.Context(), appName)
		if err != nil {
			statusCode := http.StatusInternalServerError
			if strings.Contains(err.Error(), "not found") {
				statusCode = http.StatusNotFound
			}
			WriteJSONError(w, "Failed to generate deployment plan: "+err.Error(), statusCode)
			return
		}

		// If optimize flag is also set, optimize the plan
		if isOptimize && len(plan.Steps) > 0 {
			optimizedResult, err := service.OptimizeDeploymentPlan(r.Context(), appName, plan.Steps)
			if err != nil {
				WriteJSONError(w, "Failed to optimize deployment plan: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// Return optimized plan instead
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"operation":    "plan-optimized",
				"plan":         plan,
				"optimization": optimizedResult,
			})
			return
		}

		// If analyze flag is also set, include impact analysis
		if isAnalyze {
			// For now, return basic analysis - could be enhanced with actual impact prediction
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"operation": "plan-analyzed",
				"plan":      plan,
				"analysis": map[string]interface{}{
					"estimated_duration": "5-10 minutes",
					"risk_level":         "low",
					"affected_services":  len(plan.Steps),
				},
			})
			return
		}

		// Return the deployment plan
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"operation": "plan",
			"plan":      plan,
		})
		return
	}

	// Handle actual deployment (AI integrated internally)
	result, err := service.DeployApplication(r.Context(), appName, req.Environment)
	if err != nil {
		// Determine appropriate HTTP status code based on error
		statusCode := http.StatusInternalServerError

		// Check for specific error types
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(errMsg, "not allowed") || strings.Contains(errMsg, "blocked by policy") {
			statusCode = http.StatusForbidden
		}

		WriteJSONError(w, errMsg, statusCode)
		return
	}

	// Return successful deployment result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// PredictImpact godoc
// @Summary      Predict deployment impact
// @Description  Analyzes proposed changes and predicts their impact on the deployment.
// @Tags         deployments
// @Accept       json
// @Produce      json
// @Param        app_name     path  string                      true  "Application name"
// @Param        changes      body  PredictImpactRequest        true  "Changes request"
// @Success      200  {object}  ai.ImpactAnalysisResult
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name}/predict-impact [post]
func PredictImpact(w http.ResponseWriter, r *http.Request) {
	var req PredictImpactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Environment == "" {
		WriteJSONError(w, "Environment is required", http.StatusBadRequest)
		return
	}

	// Create AI platform agent for deployment service (infrastructure layer)
	agent := GetGlobalV3Agent()
	if agent == nil {
		WriteJSONError(w, "AI service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Use deployment service for impact prediction (clean architecture - business logic in domain service)
	deploymentService := deployments.NewDeploymentService(GlobalGraph, agent.Provider())

	// Convert request changes to proper format
	changes := make([]ai.ProposedChange, len(req.Changes))
	for i, change := range req.Changes {
		changes[i] = ai.ProposedChange{
			Description: fmt.Sprintf("Change %d", i+1),
			Metadata:    change,
		}
	}

	// Predict impact using domain service
	prediction, err := deploymentService.PredictDeploymentImpact(r.Context(), changes, req.Environment)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return successful prediction result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prediction)
}

// TroubleshootDeployment godoc
// @Summary      Troubleshoot a deployment issue
// @Description  Provides troubleshooting assistance for deployment incidents.
// @Tags         deployments
// @Accept       json
// @Produce      json
// @Param        app_name     path  string                      true  "Application name"
// @Param        incident     body  TroubleshootRequest         true  "Incident request"
// @Success      200  {object}  ai.TroubleshootingResult
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name}/troubleshoot [post]
func TroubleshootDeployment(w http.ResponseWriter, r *http.Request) {
	var req TroubleshootRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.IncidentID == "" {
		WriteJSONError(w, "Incident ID is required", http.StatusBadRequest)
		return
	}

	// Create AI platform agent for deployment service (infrastructure layer)
	agent := GetGlobalV3Agent()
	if agent == nil {
		WriteJSONError(w, "AI service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Use deployment service for troubleshooting (clean architecture - business logic in domain service)
	deploymentService := deployments.NewDeploymentService(GlobalGraph, agent.Provider())

	// Troubleshoot using domain service method
	troubleshootingResult, err := deploymentService.TroubleshootDeployment(r.Context(), req.IncidentID, req.Description, req.Symptoms)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return successful troubleshooting result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(troubleshootingResult)
}

// OptimizeDeployment godoc
// @Summary      Optimize deployment settings
// @Description  Recommends optimal deployment settings based on desired targets and focus areas.
// @Tags         deployments
// @Accept       json
// @Produce      json
// @Param        app_name     path  string                      true  "Application name"
// @Param        optimization  body  OptimizeRequest            true  "Optimization request"
// @Success      200  {object}  ai.OptimizationRecommendations
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name}/optimize [post]
func OptimizeDeployment(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "app_name")

	var req OptimizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Target == "" {
		WriteJSONError(w, "Target is required", http.StatusBadRequest)
		return
	}

	// TODO: Add OptimizeDeployment method to deployment service
	// For now, return a basic optimization response
	optimizationResult := map[string]interface{}{
		"recommendations": []string{"Deployment optimization will be available after adding method to deployment service"},
		"app_name":        appName,
		"target":          req.Target,
		"focus_areas":     req.FocusAreas,
	}

	// Return optimization result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(optimizationResult)
}

// AIGeneratePlan godoc
// @Summary      Generate deployment plan using AI
// @Description  Creates an optimal deployment plan based on application requirements and edge characteristics.
// @Tags         deployments
// @Accept       json
// @Produce      json
// @Param        app_name     path  string                      true  "Application name"
// @Param        plan_request body  AIGeneratePlanRequest       true  "AI Plan request"
// @Success      200  {object}  ai.DeploymentPlan
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name}/ai/generate-plan [post]
func AIGeneratePlan(w http.ResponseWriter, r *http.Request) {
	var req AIGeneratePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.AppName == "" {
		WriteJSONError(w, "Application name is required", http.StatusBadRequest)
		return
	}

	// Set default timeout if not provided
	if req.Timeout == 0 {
		req.Timeout = 60 // 60 seconds default timeout
	}

	// Create AI platform agent for deployment service (infrastructure layer)
	agent := GetGlobalV3Agent()
	if agent == nil {
		WriteJSONError(w, "AI service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Use deployment service for AI plan generation (clean architecture - business logic in domain service)
	deploymentService := deployments.NewDeploymentService(GlobalGraph, agent.Provider())

	// Generate deployment plan using domain service method
	plan, err := deploymentService.GenerateDeploymentPlan(r.Context(), req.AppName)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return successful plan generation result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

// AIImpactAnalysis godoc
// @Summary      Analyze deployment impact using AI
// @Description  Evaluates the potential impact of changes using AI-driven analysis.
// @Tags         deployments
// @Accept       json
// @Produce      json
// @Param        app_name     path  string                      true  "Application name"
// @Param        impact_request body  AIImpactRequest            true  "AI Impact request"
// @Success      200  {object}  ai.ImpactAnalysisResult
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name}/ai/impact-analysis [post]
func AIImpactAnalysis(w http.ResponseWriter, r *http.Request) {
	var req AIImpactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Environment == "" {
		WriteJSONError(w, "Environment is required", http.StatusBadRequest)
		return
	}

	// Create AI platform agent for deployment service (infrastructure layer)
	agent := GetGlobalV3Agent()
	if agent == nil {
		WriteJSONError(w, "AI service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Use deployment service for impact analysis (clean architecture - business logic in domain service)
	deploymentService := deployments.NewDeploymentService(GlobalGraph, agent.Provider())

	// Convert request changes to proper format
	changes := make([]ai.ProposedChange, len(req.Changes))
	for i, change := range req.Changes {
		// Extract fields from map with safe defaults
		changeType, _ := change["type"].(string)
		target, _ := change["target"].(string)

		changes[i] = ai.ProposedChange{
			Type:        changeType,
			Target:      target,
			Description: fmt.Sprintf("Change %d: %s", i+1, changeType),
			Metadata:    change,
		}
	}

	// Analyze impact using domain service method
	analysis, err := deploymentService.PredictDeploymentImpact(r.Context(), changes, req.Environment)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return successful impact analysis result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analysis)
}

// AITroubleshootDeployment godoc
// @Summary      Intelligent deployment troubleshooting
// @Description  Performs AI-driven troubleshooting for deployment incidents.
// @Tags         deployments
// @Accept       json
// @Produce      json
// @Param        app_name     path  string                      true  "Application name"
// @Param        troubleshoot_request body  AITroubleshootRequest   true  "AI Troubleshoot request"
// @Success      200  {object}  ai.TroubleshootingResult
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name}/ai/troubleshoot [post]
func AITroubleshootDeployment(w http.ResponseWriter, r *http.Request) {
	var req AITroubleshootRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.IncidentID == "" {
		WriteJSONError(w, "Incident ID is required", http.StatusBadRequest)
		return
	}

	// Set default timeout if not provided
	if req.Timeout == 0 {
		req.Timeout = 60 // 60 seconds default timeout
	}

	// Create AI platform agent for deployment service (infrastructure layer)
	agent := GetGlobalV3Agent()
	if agent == nil {
		WriteJSONError(w, "AI service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Use deployment service for AI troubleshooting (clean architecture - business logic in domain service)
	deploymentService := deployments.NewDeploymentService(GlobalGraph, agent.Provider())

	// Perform AI-driven troubleshooting using domain service method
	troubleshootingResult, err := deploymentService.TroubleshootDeployment(r.Context(), req.IncidentID, req.Description, req.Symptoms)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return successful AI troubleshooting result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(troubleshootingResult)
}

// AIOptimizePlan godoc
// @Summary      Optimize deployment plan using AI
// @Description  Refines an existing deployment plan based on AI analysis and recommendations.
// @Tags         deployments
// @Accept       json
// @Produce      json
// @Param        app_name     path  string                      true  "Application name"
// @Param        optimize_request body  AIOptimizePlanRequest     true  "AI Optimize request"
// @Success      200  {object}  ai.OptimizationRecommendations
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name}/ai/optimize-plan [post]
func AIOptimizePlan(w http.ResponseWriter, r *http.Request) {
	var req AIOptimizePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.ApplicationID == "" {
		WriteJSONError(w, "Application ID is required", http.StatusBadRequest)
		return
	}

	// Create AI platform agent for deployment service (infrastructure layer)
	agent := GetGlobalV3Agent()
	if agent == nil {
		WriteJSONError(w, "AI service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Use deployment service for AI plan optimization (clean architecture - business logic in domain service)
	deploymentService := deployments.NewDeploymentService(GlobalGraph, agent.Provider())

	// Optimize deployment plan using domain service method
	optimizationResult, err := deploymentService.OptimizeDeploymentPlan(r.Context(), req.ApplicationID, req.CurrentPlan)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return successful optimization result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(optimizationResult)
}
