package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/krzachariassen/ZTDP/internal/resources"
)

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
	var req resources.ResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	resourceService := resources.NewService(GlobalGraph)
	response, err := resourceService.CreateResource(req)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// AddResourceToApplication godoc
// @Summary      Create a resource instance for an application
// @Description  Creates a named resource instance for an application based on a catalog resource. Uses predictable naming (app-resource) by default, or custom name via query parameter. Operation is idempotent - returns success if resource already exists.
// @Tags         resources
// @Produce      json
// @Param        app_name      path  string  true  "Application name"
// @Param        resource_name path  string  true  "Resource name from catalog"
// @Param        instance_name query string  false "Custom instance name (defaults to app-resource format)"
// @Success      201  {object}  map[string]interface{}  "Resource instance created"
// @Success      200  {object}  map[string]interface{}  "Resource instance already exists"
// @Failure      404  {object}  map[string]string       "Application or catalog resource not found"
// @Failure      409  {object}  map[string]string       "Name conflict with existing non-resource node"
// @Router       /v1/applications/{app_name}/resources/{resource_name} [post]
func AddResourceToApplication(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "app_name")
	resourceName := chi.URLParam(r, "resource_name")
	instanceName := r.URL.Query().Get("instance_name")

	resourceService := resources.NewService(GlobalGraph)
	response, err := resourceService.AddResourceToApplication(appName, resourceName, instanceName)
	if err != nil {
		if err.Error() == "application not found" || err.Error() == "resource not found in catalog" {
			WriteJSONError(w, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "a node with this name already exists but is not a resource" {
			WriteJSONError(w, err.Error(), http.StatusConflict)
			return
		}
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if response.Status == "exists" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	json.NewEncoder(w).Encode(response)
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
	appName := chi.URLParam(r, "app_name")
	serviceName := chi.URLParam(r, "service_name")
	resourceName := chi.URLParam(r, "resource_name")

	resourceService := resources.NewService(GlobalGraph)
	response, err := resourceService.LinkServiceToResource(appName, serviceName, resourceName)
	if err != nil {
		if err.Error() == "application not found" || err.Error() == "service not found" {
			WriteJSONError(w, err.Error(), http.StatusNotFound)
			return
		}
		if fmt.Sprintf("resource instance not found in application: %v", err) == err.Error() {
			WriteJSONError(w, err.Error(), http.StatusNotFound)
			return
		}
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status":               response.Message,
		"resource_instance_id": response.InstanceName,
	})
}

// ListResources godoc
// @Summary      List all resources in the resource catalog
// @Description  Returns all resource nodes in the global graph
// @Tags         resources
// @Produce      json
// @Success      200  {array}  map[string]interface{}
// @Router       /v1/resources [get]
func ListResources(w http.ResponseWriter, r *http.Request) {
	resourceService := resources.NewService(GlobalGraph)
	resourceList, err := resourceService.ListResources()
	if err != nil {
		WriteJSONError(w, "Failed to get resources", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resourceList)
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

	resourceService := resources.NewService(GlobalGraph)
	resourceList, err := resourceService.ListApplicationResources(appName)
	if err != nil {
		if err.Error() == "application not found" {
			WriteJSONError(w, err.Error(), http.StatusNotFound)
			return
		}
		WriteJSONError(w, "Failed to get application resources", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resourceList)
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

	resourceService := resources.NewService(GlobalGraph)
	resourceList, err := resourceService.ListServiceResources(serviceName)
	if err != nil {
		if err.Error() == "service not found" {
			WriteJSONError(w, err.Error(), http.StatusNotFound)
			return
		}
		WriteJSONError(w, "Failed to get service resources", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resourceList)
}
