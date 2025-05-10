package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/krzachariassen/ZTDP/internal/contracts"
)

// DeployServiceVersion godoc
// @Summary      Deploy a service version to an environment
// @Description  Deploys a specific service version to an environment
// @Tags         deployments
// @Accept       json
// @Produce      json
// @Param        app_name     path  string  true  "Application name"
// @Param        service_name path  string  true  "Service name"
// @Param        version      path  string  true  "Service version"
// @Param        env         body  object  true  "Deployment target"
// @Success      201  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Router       /v1/applications/{app_name}/services/{service_name}/versions/{version}/deploy [post]
func DeployServiceVersion(w http.ResponseWriter, r *http.Request) {
	serviceName := chi.URLParam(r, "service_name")
	version := chi.URLParam(r, "version")
	var req struct {
		Environment string `json:"environment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	verID := serviceName + ":" + version
	// Debug log to help trace service version lookup
	if _, ok := GlobalGraph.Graph.Nodes[verID]; !ok {
		WriteJSONError(w, "Service version not found", http.StatusNotFound)
		return
	}
	if _, ok := GlobalGraph.Graph.Nodes[req.Environment]; !ok {
		WriteJSONError(w, "Environment not found", http.StatusNotFound)
		return
	}
	if err := GlobalGraph.Graph.AddEdge(verID, req.Environment, "deployed_in"); err != nil {
		// Policy errors should return 403 Forbidden
		if err.Error() == "deployment to environment '"+req.Environment+"' is not allowed for application '"+chi.URLParam(r, "app_name")+"'" ||
			err.Error() == "must deploy service version "+verID+" to 'dev' before deploying to 'prod'" ||
			(err != nil && (err.Error() == "service version node not found or not a service_version" || err.Error() == "source node "+verID+" does not exist")) {
			WriteJSONError(w, err.Error(), http.StatusForbidden)
			return
		}
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := GlobalGraph.Save(); err != nil {
		WriteJSONError(w, "Failed to save deployment", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "deployed"})
}

// ListEnvironmentDeployments godoc
// @Summary      List deployments in an environment
// @Description  Returns all service versions deployed in the environment
// @Tags         deployments
// @Produce      json
// @Param        env_name  path  string  true  "Environment name"
// @Success      200  {array}  contracts.ServiceVersionContract
// @Router       /v1/environments/{env_name}/deployments [get]
func ListEnvironmentDeployments(w http.ResponseWriter, r *http.Request) {
	envName := chi.URLParam(r, "env_name")
	deployments := []contracts.ServiceVersionContract{}
	for from, edges := range GlobalGraph.Graph.Edges {
		for _, edge := range edges {
			if edge.Type == "deployed_in" && edge.To == envName {
				if node, ok := GlobalGraph.Graph.Nodes[from]; ok && node.Kind == "service_version" {
					var ver contracts.ServiceVersionContract
					b, _ := json.Marshal(node)
					_ = json.Unmarshal(b, &ver)
					deployments = append(deployments, ver)
				}
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployments)
}
