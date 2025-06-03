package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/deployments"
)

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
// @Summary      Deploy an application to an environment
// @Description  Deploys all services of an application to the specified environment. This is the primary deployment interface for MVP v1.
// @Tags         deployments
// @Accept       json
// @Produce      json
// @Param        app_name     path  string                      true  "Application name"
// @Param        deployment   body  DeployApplicationRequest    true  "Deployment request"
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

	// Use deployment service for orchestration (handles AI gracefully)
	service := deployments.NewService(GlobalGraph)
	result, err := service.DeployApplication(context.Background(), appName, req.Environment)
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
	appName := chi.URLParam(r, "app_name")

	var req PredictImpactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Environment == "" {
		WriteJSONError(w, "Environment is required", http.StatusBadRequest)
		return
	}

	// Analyze impact using AI-powered analysis
	analyzer := ai.NewImpactAnalyzer(GlobalGraph)
	analysisResult, err := analyzer.AnalyzeDeploymentImpact(appName, req.Changes, req.Environment)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return successful analysis result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analysisResult)
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
	appName := chi.URLParam(r, "app_name")

	var req TroubleshootRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.IncidentID == "" {
		WriteJSONError(w, "Incident ID is required", http.StatusBadRequest)
		return
	}

	// Perform troubleshooting using AI-powered analysis
	troubleshooter := ai.NewTroubleshooter(GlobalGraph)
	troubleshootingResult, err := troubleshooter.TroubleshootDeploymentIssue(appName, req.IncidentID, req.Description, req.Symptoms)
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
// @Success      200  {object}  ai.OptimizationResult
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

	// Optimize deployment settings using AI-powered optimization
	optimizer := ai.NewOptimizer(GlobalGraph)
	optimizationResult, err := optimizer.OptimizeDeploymentSettings(appName, req.Target, req.FocusAreas)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return successful optimization result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(optimizationResult)
}
