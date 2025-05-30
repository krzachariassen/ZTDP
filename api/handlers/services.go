package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	servicecore "github.com/krzachariassen/ZTDP/internal/service"
)

// CreateService godoc
// @Summary      Create a new service for an application
// @Description  Creates a new service resource linked to an application
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        app_name  path      string                  true  "Application name"
// @Param        service   body      map[string]interface{} true  "Service payload"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/applications/{app_name}/services [post]
func CreateService(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "app_name")
	var svcData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&svcData); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	serviceService := servicecore.NewServiceService(GlobalGraph)
	createdSvc, err := serviceService.CreateService(appName, svcData)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdSvc)
}

// ListServices godoc
// @Summary      List all services for an application
// @Description  Returns all services linked to an application
// @Tags         services
// @Produce      json
// @Param        app_name  path      string  true  "Application name"
// @Success      200  {array}   map[string]interface{}
// @Router       /v1/applications/{app_name}/services [get]
func ListServices(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "app_name")
	serviceService := servicecore.NewServiceService(GlobalGraph)
	services, err := serviceService.ListServices(appName)
	if err != nil {
		WriteJSONError(w, "Failed to get services", http.StatusInternalServerError)
		return
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
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name}/services/{service_name} [get]
func GetService(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "app_name")
	serviceName := chi.URLParam(r, "service_name")
	serviceService := servicecore.NewServiceService(GlobalGraph)
	service, err := serviceService.GetService(appName, serviceName)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(service)
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
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/applications/{app_name}/services/{service_name}/versions [post]
func CreateServiceVersion(w http.ResponseWriter, r *http.Request) {
	serviceName := chi.URLParam(r, "service_name")

	var versionData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&versionData); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	serviceService := servicecore.NewServiceService(GlobalGraph)
	createdVersion, err := serviceService.CreateServiceVersion(serviceName, versionData)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdVersion)
}

// ListServiceVersions godoc
// @Summary      List all versions for a service
// @Description  Returns all versions for a service
// @Tags         services
// @Produce      json
// @Param        app_name     path  string  true  "Application name"
// @Param        service_name path  string  true  "Service name"
// @Success      200  {array}  map[string]interface{}
// @Router       /v1/applications/{app_name}/services/{service_name}/versions [get]
func ListServiceVersions(w http.ResponseWriter, r *http.Request) {
	serviceName := chi.URLParam(r, "service_name")
	serviceService := servicecore.NewServiceService(GlobalGraph)
	versions, err := serviceService.ListServiceVersions(serviceName)
	if err != nil {
		WriteJSONError(w, "Failed to get service versions", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versions)
}
