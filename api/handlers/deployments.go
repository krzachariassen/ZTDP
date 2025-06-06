package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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
	service := deployments.NewDeploymentService(GlobalGraph, nil)
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
	agent, err := ai.NewPlatformAgentFromConfig(GlobalGraph, nil, nil)
	if err != nil {
		WriteJSONError(w, "AI service unavailable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer agent.Close()

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
	agent, err := ai.NewPlatformAgentFromConfig(GlobalGraph, nil, nil)
	if err != nil {
		WriteJSONError(w, "AI service unavailable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer agent.Close()

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
