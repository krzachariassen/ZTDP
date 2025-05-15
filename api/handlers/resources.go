package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

const resourceCatalogNodeID = "resource-catalog"
const resourceCatalogKind = "resource_register"

// ensureResourceCatalogRoot ensures the resource catalog root node exists in the graph
func ensureResourceCatalogRoot(g *graph.GlobalGraph) {
	if _, ok := g.Graph.Nodes[resourceCatalogNodeID]; !ok {
		root := &graph.Node{
			ID:   resourceCatalogNodeID,
			Kind: resourceCatalogKind,
			Metadata: map[string]interface{}{
				"name":  resourceCatalogNodeID,
				"owner": "platform-team",
			},
			Spec: map[string]interface{}{
				"description": "Root node for all resource types in the platform",
			},
		}
		g.AddNode(root)
	}
}

// Repairs the resource catalog: ensures all resource_type nodes are owned by the catalog root
func repairResourceCatalogRelationships(g *graph.GlobalGraph) {
	ensureResourceCatalogRoot(g)
	for _, node := range g.Graph.Nodes {
		if node.Kind == "resource_type" {
			hasEdge := false
			for _, edge := range g.Graph.Edges[resourceCatalogNodeID] {
				if edge.To == node.ID && edge.Type == "owns" {
					hasEdge = true
					break
				}
			}
			if !hasEdge {
				g.AddEdge(resourceCatalogNodeID, node.ID, "owns")
			}
		}
	}
}

// CreateResource godoc
// @Summary      Create a new resource (from catalog)
// @Description  Creates a new resource node in the global graph
// @Tags         resources
// @Accept       json
// @Produce      json
// @Param        resource  body  map[string]interface{}  true  "Resource payload"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/resources [post]
func CreateResource(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Kind     string                 `json:"kind"`
		Metadata map[string]interface{} `json:"metadata"`
		Spec     map[string]interface{} `json:"spec"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Kind != "resource_type" && req.Kind != "resource" {
		WriteJSONError(w, "Invalid kind for resource catalog", http.StatusBadRequest)
		return
	}

	// Validate required metadata fields
	nameVal, nameOk := req.Metadata["name"].(string)
	ownerVal, ownerOk := req.Metadata["owner"].(string)
	if !nameOk || nameVal == "" {
		WriteJSONError(w, "Resource metadata 'name' is required", http.StatusBadRequest)
		return
	}
	if req.Kind == "resource_type" && (!ownerOk || ownerVal == "") {
		WriteJSONError(w, "Resource type metadata 'owner' is required", http.StatusBadRequest)
		return
	}
	if req.Kind == "resource" && !ownerOk {
		// For resource instances, allow owner to be empty (optional), but always set to string
		ownerVal = ""
	}

	// Build the contract
	var node *graph.Node
	if req.Kind == "resource_type" {
		resourceType := contracts.ResourceTypeContract{}
		metadata := contracts.Metadata{
			Name:  nameVal,
			Owner: ownerVal,
		}
		resourceType.Metadata = metadata
		if specBytes, err := json.Marshal(req.Spec); err == nil {
			json.Unmarshal(specBytes, &resourceType.Spec)
		}
		node, _ = graph.ResolveContract(resourceType)
	} else {
		resource := contracts.ResourceContract{}
		metadata := contracts.Metadata{
			Name:  nameVal,
			Owner: ownerVal,
		}
		resource.Metadata = metadata
		if specBytes, err := json.Marshal(req.Spec); err == nil {
			json.Unmarshal(specBytes, &resource.Spec)
		}
		node, _ = graph.ResolveContract(resource)
	}

	g := GlobalGraph
	ensureResourceCatalogRoot(g)
	g.AddNode(node)

	// If this is a resource_type, add an 'owns' edge from the catalog root
	if node.Kind == "resource_type" {
		g.AddEdge(resourceCatalogNodeID, node.ID, "owns")
	}

	repairResourceCatalogRelationships(g)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(node)
}

// AddResourceToApplication godoc
// @Summary      Add a resource to an application
// @Description  Creates an 'owns' edge from application to resource
// @Tags         resources
// @Produce      json
// @Param        app_name      path  string  true  "Application name"
// @Param        resource_name path  string  true  "Resource name"
// @Success      201  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name}/resources/{resource_name} [post]
func AddResourceToApplication(w http.ResponseWriter, r *http.Request) {
	resourceName := chi.URLParam(r, "resource_name")
	appNode, appOk := GlobalGraph.Graph.Nodes[chi.URLParam(r, "app_name")]
	_, resOk := GlobalGraph.Graph.Nodes[resourceName]
	if !appOk || appNode.Kind != "application" {
		WriteJSONError(w, "Application not found", http.StatusNotFound)
		return
	}
	if !resOk {
		WriteJSONError(w, "Resource not found", http.StatusNotFound)
		return
	}
	GlobalGraph.AddEdge(chi.URLParam(r, "app_name"), resourceName, "owns")
	if err := GlobalGraph.Save(); err != nil {
		WriteJSONError(w, "Failed to save graph", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "resource added to application"})
}

// LinkServiceToResource godoc
// @Summary      Link a service to a resource (creates 'uses' edge)
// @Description  Creates a 'uses' edge from service to resource in the application
// @Tags         resources
// @Produce      json
// @Param        app_name      path  string  true  "Application name"
// @Param        service_name  path  string  true  "Service name"
// @Param        resource_name path  string  true  "Resource name"
// @Success      201  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name}/services/{service_name}/resources/{resource_name} [post]
func LinkServiceToResource(w http.ResponseWriter, r *http.Request) {
	serviceName := chi.URLParam(r, "service_name")
	resourceName := chi.URLParam(r, "resource_name")
	serviceNode, svcOk := GlobalGraph.Graph.Nodes[serviceName]
	_, resOk := GlobalGraph.Graph.Nodes[resourceName]
	if !svcOk || serviceNode.Kind != "service" {
		WriteJSONError(w, "Service not found", http.StatusNotFound)
		return
	}
	if !resOk {
		WriteJSONError(w, "Resource not found", http.StatusNotFound)
		return
	}
	GlobalGraph.AddEdge(serviceName, resourceName, "uses")
	if err := GlobalGraph.Save(); err != nil {
		WriteJSONError(w, "Failed to save graph", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "service linked to resource"})
}

// ListResources godoc
// @Summary      List all resources in the resource catalog
// @Description  Returns all resource nodes in the global graph
// @Tags         resources
// @Produce      json
// @Success      200  {array}  map[string]interface{}
// @Router       /v1/resources [get]
func ListResources(w http.ResponseWriter, r *http.Request) {
	repairResourceCatalogRelationships(GlobalGraph)
	resources := []map[string]interface{}{}
	for _, node := range GlobalGraph.Graph.Nodes {
		if node.Kind == "resource" || node.Kind == "resource_type" {
			resource := map[string]interface{}{
				"id":       node.ID,
				"kind":     node.Kind,
				"metadata": node.Metadata,
				"spec":     node.Spec,
			}
			resources = append(resources, resource)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources)
}

// ListApplicationResources godoc
// @Summary      List all resources for an application
// @Description  Returns all resource nodes owned by the application
// @Tags         resources
// @Produce      json
// @Param        app_name  path  string  true  "Application name"
// @Success      200  {array}  map[string]interface{}
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name}/resources [get]
func ListApplicationResources(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "app_name")
	appNode, appOk := GlobalGraph.Graph.Nodes[appName]
	if !appOk || appNode.Kind != "application" {
		WriteJSONError(w, "Application not found", http.StatusNotFound)
		return
	}
	resources := []map[string]interface{}{}
	for _, edge := range GlobalGraph.Graph.Edges[appName] {
		if edge.Type == "owns" {
			if node, ok := GlobalGraph.Graph.Nodes[edge.To]; ok && node.Kind == "resource" {
				resource := map[string]interface{}{
					"id":       node.ID,
					"kind":     node.Kind,
					"metadata": node.Metadata,
					"spec":     node.Spec,
				}
				resources = append(resources, resource)
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources)
}

// ListServiceResources godoc
// @Summary      List all resources used by a service
// @Description  Returns all resource nodes linked by 'uses' edge from the service
// @Tags         resources
// @Produce      json
// @Param        app_name      path  string  true  "Application name"
// @Param        service_name  path  string  true  "Service name"
// @Success      200  {array}  map[string]interface{}
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name}/services/{service_name}/resources [get]
func ListServiceResources(w http.ResponseWriter, r *http.Request) {
	serviceName := chi.URLParam(r, "service_name")
	serviceNode, svcOk := GlobalGraph.Graph.Nodes[serviceName]
	if !svcOk || serviceNode.Kind != "service" {
		WriteJSONError(w, "Service not found", http.StatusNotFound)
		return
	}
	resources := []map[string]interface{}{}
	for _, edge := range GlobalGraph.Graph.Edges[serviceName] {
		if edge.Type == "uses" {
			if node, ok := GlobalGraph.Graph.Nodes[edge.To]; ok && node.Kind == "resource" {
				resource := map[string]interface{}{
					"id":       node.ID,
					"kind":     node.Kind,
					"metadata": node.Metadata,
					"spec":     node.Spec,
				}
				resources = append(resources, resource)
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources)
}
