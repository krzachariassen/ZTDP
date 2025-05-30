package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/krzachariassen/ZTDP/internal/environment"
)

// CreateEnvironment godoc
// @Summary      Create a new environment
// @Description  Creates a new environment node
// @Tags         environments
// @Accept       json
// @Produce      json
// @Param        environment  body      map[string]interface{}  true  "Environment payload"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/environments [post]
func CreateEnvironment(w http.ResponseWriter, r *http.Request) {
	var envData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&envData); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	envService := environment.NewService(GlobalGraph)
	createdEnv, err := envService.CreateEnvironmentFromData(envData)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdEnv)
}

// ListEnvironments godoc
// @Summary      List all environments
// @Description  Returns all environment nodes
// @Tags         environments
// @Produce      json
// @Success      200  {array}  map[string]interface{}
// @Router       /v1/environments [get]
func ListEnvironments(w http.ResponseWriter, r *http.Request) {
	envService := environment.NewService(GlobalGraph)
	envs, err := envService.ListEnvironmentsAsData()
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(envs)
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
	envService := environment.NewService(GlobalGraph)
	if err := envService.LinkAppAllowedInEnvironment(appName, envName); err != nil {
		if err.Error() == "application not found" || err.Error() == "environment not found" {
			WriteJSONError(w, err.Error(), http.StatusNotFound)
			return
		}
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
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
// @Success      200  {array}  map[string]interface{}
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name}/environments/allowed [get]
func ListAllowedEnvironments(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "app_name")
	envService := environment.NewService(GlobalGraph)
	allowedEnvs, err := envService.ListAllowedEnvironmentsAsData(appName)
	if err != nil {
		if err.Error() == "application not found" {
			WriteJSONError(w, err.Error(), http.StatusNotFound)
			return
		}
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
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
	var envs []string
	if err := json.NewDecoder(r.Body).Decode(&envs); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	envService := environment.NewService(GlobalGraph)
	if err := envService.UpdateAllowedEnvironments(appName, envs); err != nil {
		if err.Error() == "application not found" || (len(err.Error()) > 19 && err.Error()[:19] == "environment not found") {
			WriteJSONError(w, err.Error(), http.StatusNotFound)
			return
		}
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
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
	var envs []string
	if err := json.NewDecoder(r.Body).Decode(&envs); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	envService := environment.NewService(GlobalGraph)
	if err := envService.AddAllowedEnvironments(appName, envs); err != nil {
		if err.Error() == "application not found" || (len(err.Error()) > 19 && err.Error()[:19] == "environment not found") {
			WriteJSONError(w, err.Error(), http.StatusNotFound)
			return
		}
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "added"})
}
