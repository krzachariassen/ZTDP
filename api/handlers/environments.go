package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/policies"
)

// CreateEnvironment godoc
// @Summary      Create a new environment
// @Description  Creates a new environment node
// @Tags         environments
// @Accept       json
// @Produce      json
// @Param        environment  body      contracts.EnvironmentContract  true  "Environment payload"
// @Success      201  {object}  contracts.EnvironmentContract
// @Failure      400  {object}  map[string]string
// @Router       /v1/environments [post]
func CreateEnvironment(w http.ResponseWriter, r *http.Request) {
	var env contracts.EnvironmentContract
	if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	node, _ := graph.ResolveContract(env)
	GlobalGraph.AddNode(node)
	if err := GlobalGraph.Save(); err != nil {
		WriteJSONError(w, "Failed to save environment", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(env)
}

// ListEnvironments godoc
// @Summary      List all environments
// @Description  Returns all environment nodes
// @Tags         environments
// @Produce      json
// @Success      200  {array}  contracts.EnvironmentContract
// @Router       /v1/environments [get]
func ListEnvironments(w http.ResponseWriter, r *http.Request) {
	envs := []contracts.EnvironmentContract{}
	for _, node := range GlobalGraph.Graph.Nodes {
		if node.Kind == "environment" {
			contract, err := graph.LoadNode(node.Kind, node.Spec, contracts.Metadata{
				Name:  node.Metadata["name"].(string),
				Owner: node.Metadata["owner"].(string),
			})
			if err == nil {
				if env, ok := contract.(*contracts.EnvironmentContract); ok {
					envs = append(envs, *env)
				}
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(envs)
}

// LinkServiceToEnvironment godoc
// @Summary      Link a service to an environment
// @Description  Creates a 'deployed_in' edge from a service to an environment, ensuring the service belongs to the specified application
// @Tags         environments
// @Produce      json
// @Param        app_name      path  string  true  "Application name"
// @Param        service_name  path  string  true  "Service name"
// @Param        env_name      path  string  true  "Environment name"
// @Success      201  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name}/services/{service_name}/environments/{env_name} [post]
func LinkServiceToEnvironment(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "app_name")
	serviceName := chi.URLParam(r, "service_name")
	envName := chi.URLParam(r, "env_name")
	serviceNode, serviceOk := GlobalGraph.Graph.Nodes[serviceName]
	envNode, envOk := GlobalGraph.Graph.Nodes[envName]
	if !serviceOk || serviceNode.Kind != "service" {
		WriteJSONError(w, "Service not found", http.StatusNotFound)
		return
	}
	if !envOk || envNode.Kind != "environment" {
		WriteJSONError(w, "Environment not found", http.StatusNotFound)
		return
	}
	// Ensure the service belongs to the specified application
	appEdgeFound := false
	for _, edge := range GlobalGraph.Graph.Edges[appName] {
		if edge.To == serviceName && edge.Type == "owns" {
			appEdgeFound = true
			break
		}
	}
	if !appEdgeFound {
		WriteJSONError(w, "Service does not belong to the specified application", http.StatusNotFound)
		return
	}
	// Enforce allowed_in policy
	if !policies.IsEnvironmentAllowedForApp(GlobalGraph.Graph, appName, envName) {
		WriteJSONError(w, "Environment is not allowed for this application", http.StatusForbidden)
		return
	}
	if err := GlobalGraph.AddEdge(serviceName, envName, "deployed_in"); err != nil {
		fmt.Printf("AddEdge error: %v\n", err)
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := GlobalGraph.Save(); err != nil {
		WriteJSONError(w, "Failed to save graph", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "linked"})

}

// LinkAppAllowedInEnvironment godoc
// @Summary      Add an allowed_in policy edge from an application to an environment
// @Description  Creates an 'allowed_in' policy edge from an application to an environment
// @Tags         environments
// @Produce      json
// @Param        app_name  path  string  true  "Application name"
// @Param        env_name  path  string  true  "Environment name"
// @Success      201  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name}/environments/{env_name}/allowed [post]
func LinkAppAllowedInEnvironment(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "app_name")
	envName := chi.URLParam(r, "env_name")
	appNode, appOk := GlobalGraph.Graph.Nodes[appName]
	envNode, envOk := GlobalGraph.Graph.Nodes[envName]
	if !appOk || appNode.Kind != "application" {
		WriteJSONError(w, "Application not found", http.StatusNotFound)
		return
	}
	if !envOk || envNode.Kind != "environment" {
		WriteJSONError(w, "Environment not found", http.StatusNotFound)
		return
	}
	// Use policy function to check if already allowed
	if policies.IsEnvironmentAllowedForApp(GlobalGraph.Graph, appName, envName) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "already allowed"})
		return
	}
	if err := GlobalGraph.AddEdge(appName, envName, "allowed_in"); err != nil {
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := GlobalGraph.Save(); err != nil {
		WriteJSONError(w, "Failed to save graph", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "allowed"})
}

// ListAllowedEnvironments godoc
// @Summary      List allowed environments for an application
// @Description  Returns all environments the application is allowed to deploy to (policy)
// @Tags         environments
// @Produce      json
// @Param        app_name  path  string  true  "Application name"
// @Success      200  {array}  contracts.EnvironmentContract
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name}/environments/allowed [get]
func ListAllowedEnvironments(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "app_name")
	appNode, appOk := GlobalGraph.Graph.Nodes[appName]
	if !appOk || appNode.Kind != "application" {
		WriteJSONError(w, "Application not found", http.StatusNotFound)
		return
	}
	allowedIDs := policies.GetAllowedEnvironmentsForApp(GlobalGraph.Graph, appName)
	allowedEnvs := []contracts.EnvironmentContract{}
	for _, envID := range allowedIDs {
		node, ok := GlobalGraph.Graph.Nodes[envID]
		if ok && node.Kind == "environment" {
			contract, err := graph.LoadNode(node.Kind, node.Spec, contracts.Metadata{
				Name:  node.Metadata["name"].(string),
				Owner: node.Metadata["owner"].(string),
			})
			if err == nil {
				if env, ok := contract.(*contracts.EnvironmentContract); ok {
					allowedEnvs = append(allowedEnvs, *env)
				}
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allowedEnvs)
}

// UpdateAllowedEnvironments godoc
// @Summary      Replace allowed environments for an application
// @Description  Replaces the allowed_in policy edges for an application
// @Tags         environments
// @Accept       json
// @Produce      json
// @Param        app_name  path  string  true  "Application name"
// @Param        envs      body  []string true "List of environment names"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name}/environments/allowed [put]
func UpdateAllowedEnvironments(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "app_name")
	appNode, appOk := GlobalGraph.Graph.Nodes[appName]
	if !appOk || appNode.Kind != "application" {
		WriteJSONError(w, "Application not found", http.StatusNotFound)
		return
	}
	var envs []string
	if err := json.NewDecoder(r.Body).Decode(&envs); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	// Remove all existing allowed_in edges
	newEdges := []graph.Edge{}
	for _, edge := range GlobalGraph.Graph.Edges[appName] {
		if edge.Type != "allowed_in" {
			newEdges = append(newEdges, edge)
		}
	}
	GlobalGraph.Graph.Edges[appName] = newEdges
	// Add new allowed_in edges
	for _, envName := range envs {
		envNode, envOk := GlobalGraph.Graph.Nodes[envName]
		if !envOk || envNode.Kind != "environment" {
			WriteJSONError(w, "Environment not found: "+envName, http.StatusNotFound)
			return
		}
		GlobalGraph.AddEdge(appName, envName, "allowed_in")
	}
	if err := GlobalGraph.Save(); err != nil {
		WriteJSONError(w, "Failed to save graph", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// AddAllowedEnvironments godoc
// @Summary      Add allowed environments for an application
// @Description  Adds allowed_in policy edges for an application (does not remove existing)
// @Tags         environments
// @Accept       json
// @Produce      json
// @Param        app_name  path  string  true  "Application name"
// @Param        envs      body  []string true "List of environment names"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name}/environments/allowed [post]
func AddAllowedEnvironments(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "app_name")
	appNode, appOk := GlobalGraph.Graph.Nodes[appName]
	if !appOk || appNode.Kind != "application" {
		WriteJSONError(w, "Application not found", http.StatusNotFound)
		return
	}
	var envs []string
	if err := json.NewDecoder(r.Body).Decode(&envs); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	for _, envName := range envs {
		envNode, envOk := GlobalGraph.Graph.Nodes[envName]
		if !envOk || envNode.Kind != "environment" {
			WriteJSONError(w, "Environment not found: "+envName, http.StatusNotFound)
			return
		}
		if !policies.IsEnvironmentAllowedForApp(GlobalGraph.Graph, appName, envName) {
			GlobalGraph.AddEdge(appName, envName, "allowed_in")
		}
	}
	if err := GlobalGraph.Save(); err != nil {
		WriteJSONError(w, "Failed to save graph", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "added"})
}
