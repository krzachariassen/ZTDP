package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

// CreateService godoc
// @Summary      Create a new service for an application
// @Description  Creates a new service resource linked to an application
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        app_name  path      string                  true  "Application name"
// @Param        service   body      contracts.ServiceContract true  "Service payload"
// @Success      201  {object}  contracts.ServiceContract
// @Failure      400  {object}  map[string]string
// @Router       /v1/applications/{app_name}/services [post]
func CreateService(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "app_name")
	var svc contracts.ServiceContract
	if err := json.NewDecoder(r.Body).Decode(&svc); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if svc.Spec.Application != appName {
		WriteJSONError(w, "Service must be linked to the specified application", http.StatusBadRequest)
		return
	}
	if err := svc.Validate(); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	node, _ := graph.ResolveContract(svc)
	GlobalGraph.AddNode(node)
	// Add edge with relationship type 'owns'
	GlobalGraph.AddEdge(appName, svc.Metadata.Name, "owns")
	if err := GlobalGraph.Save(); err != nil {
		WriteJSONError(w, "Failed to save service", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(svc)
}

// ListServices godoc
// @Summary      List all services for an application
// @Description  Returns all services linked to an application
// @Tags         services
// @Produce      json
// @Param        app_name  path      string  true  "Application name"
// @Success      200  {array}   contracts.ServiceContract
// @Router       /v1/applications/{app_name}/services [get]
func ListServices(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "app_name")
	services := []contracts.ServiceContract{}
	for _, node := range GlobalGraph.Graph.Nodes {
		if node.Kind == "service" {
			contract, err := graph.LoadNode(node.Kind, node.Spec, contracts.Metadata{
				Name:  node.Metadata["name"].(string),
				Owner: node.Metadata["owner"].(string),
			})
			if err == nil {
				if svc, ok := contract.(*contracts.ServiceContract); ok && svc.Spec.Application == appName {
					services = append(services, *svc)
				}
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

// GetService godoc
// @Summary      Get a service for an application
// @Description  Returns a specific service by name for an application
// @Tags         services
// @Produce      json
// @Param        app_name     path      string  true  "Application name"
// @Param        service_name path      string  true  "Service name"
// @Success      200  {object}  contracts.ServiceContract
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name}/services/{service_name} [get]
func GetService(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "app_name")
	serviceName := chi.URLParam(r, "service_name")
	node, ok := GlobalGraph.Graph.Nodes[serviceName]
	if !ok || node.Kind != "service" {
		WriteJSONError(w, "Service not found", http.StatusNotFound)
		return
	}
	contract, err := graph.LoadNode(node.Kind, node.Spec, contracts.Metadata{
		Name:  node.Metadata["name"].(string),
		Owner: node.Metadata["owner"].(string),
	})
	if err != nil {
		WriteJSONError(w, "Invalid service contract", http.StatusInternalServerError)
		return
	}
	svc, ok := contract.(*contracts.ServiceContract)
	if !ok || svc.Spec.Application != appName {
		WriteJSONError(w, "Service not found for this application", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(svc)
}

// CreateServiceVersion godoc
// @Summary      Create a new service version
// @Description  Creates a new version for a service
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        app_name     path  string  true  "Application name"
// @Param        service_name path  string  true  "Service name"
// @Param        version      body  object  true  "Service version payload"
// @Success      201  {object}  contracts.ServiceVersionContract
// @Failure      400  {object}  map[string]string
// @Router       /v1/applications/{app_name}/services/{service_name}/versions [post]
func CreateServiceVersion(w http.ResponseWriter, r *http.Request) {
	serviceName := chi.URLParam(r, "service_name")
	if _, ok := GlobalGraph.Graph.Nodes[serviceName]; !ok {
		WriteJSONError(w, "Service does not exist", http.StatusBadRequest)
		return
	}
	var req struct {
		Version   string `json:"version"`
		ConfigRef string `json:"config_ref"`
		Owner     string `json:"owner"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		body, _ := io.ReadAll(r.Body)
		WriteJSONError(w, "Invalid JSON: "+string(body), http.StatusBadRequest)
		return
	}
	if req.Version == "" {
		WriteJSONError(w, "Version is required", http.StatusBadRequest)
		return
	}
	id := serviceName + ":" + req.Version
	if existingNode, exists := GlobalGraph.Graph.Nodes[id]; exists {
		w.WriteHeader(http.StatusOK)
		var existingVer contracts.ServiceVersionContract
		b, _ := json.Marshal(existingNode)
		_ = json.Unmarshal(b, &existingVer)
		json.NewEncoder(w).Encode(existingVer)
		return
	}
	ver := contracts.ServiceVersionContract{
		IDValue:   id,
		Name:      serviceName,
		Owner:     req.Owner,
		Version:   req.Version,
		ConfigRef: req.ConfigRef,
		CreatedAt: time.Now(),
	}
	if err := ver.Validate(); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	node, _ := graph.ResolveContract(ver)
	GlobalGraph.AddNode(node)
	GlobalGraph.AddEdge(serviceName, id, "has_version")
	if err := GlobalGraph.Save(); err != nil {
		WriteJSONError(w, "Failed to save service version", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ver)
}

// ListServiceVersions godoc
// @Summary      List all versions for a service
// @Description  Returns all versions for a service
// @Tags         services
// @Produce      json
// @Param        app_name     path  string  true  "Application name"
// @Param        service_name path  string  true  "Service name"
// @Success      200  {array}  contracts.ServiceVersionContract
// @Router       /v1/applications/{app_name}/services/{service_name}/versions [get]
func ListServiceVersions(w http.ResponseWriter, r *http.Request) {
	serviceName := chi.URLParam(r, "service_name")
	versions := []contracts.ServiceVersionContract{}
	for _, edge := range GlobalGraph.Graph.Edges[serviceName] {
		if edge.Type == "has_version" {
			if node, ok := GlobalGraph.Graph.Nodes[edge.To]; ok && node.Kind == "service_version" {
				var ver contracts.ServiceVersionContract
				b, _ := json.Marshal(node)
				_ = json.Unmarshal(b, &ver)
				versions = append(versions, ver)
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versions)
}

// DeployServiceVersion godoc
// @Summary      Deploy a service version to an environment
// @Description  Deploys a specific service version to an environment
// @Tags         services
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
	if _, ok := GlobalGraph.Graph.Nodes[verID]; !ok {
		WriteJSONError(w, "Service version not found", http.StatusNotFound)
		return
	}
	if _, ok := GlobalGraph.Graph.Nodes[req.Environment]; !ok {
		WriteJSONError(w, "Environment not found", http.StatusNotFound)
		return
	}
	GlobalGraph.AddEdge(verID, req.Environment, "deployed_in")
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
// @Tags         environments
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

// getNodeKeys returns a slice of all keys in the given map (for logging)
func getNodeKeys(m map[string]*graph.Node) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
