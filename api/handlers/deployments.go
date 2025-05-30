package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/krzachariassen/ZTDP/internal/deployments"
)

// DeployApplicationRequest represents the payload for application deployment
type DeployApplicationRequest struct {
	Environment string `json:"environment"`
	Version     string `json:"version,omitempty"` // Optional: future feature for specific app versions
}

// DeployApplication godoc
// @Summary      Deploy an application to an environment
// @Description  Deploys all services of an application to the specified environment. This is the primary deployment interface for MVP v1.
// @Tags         applications
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

	// Execute deployment using event-driven deployment engine
	engine := deployments.NewEngine(GlobalGraph)
	result, err := engine.ExecuteApplicationDeployment(appName, req.Environment)
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
