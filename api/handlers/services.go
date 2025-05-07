package handlers

import (
	"encoding/json"
	"net/http"

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
