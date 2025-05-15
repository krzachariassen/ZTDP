package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/resources"
)

// CreateApplication godoc
// @Summary      Create a new application
// @Description  Creates a new application resource
// @Tags         applications
// @Accept       json
// @Produce      json
// @Param        application  body      contracts.ApplicationContract  true  "Application payload"
// @Success      201  {object}  contracts.ApplicationContract
// @Failure      400  {object}  map[string]string
// @Router       /v1/applications [post]
func CreateApplication(w http.ResponseWriter, r *http.Request) {
	var app contracts.ApplicationContract
	if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if err := app.Validate(); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	node, _ := graph.ResolveContract(app)
	GlobalGraph.AddNode(node)
	if err := GlobalGraph.Save(); err != nil {
		WriteJSONError(w, "Failed to save application", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(app)
}

// ListApplications godoc
// @Summary      List all applications
// @Description  Returns all application resources
// @Tags         applications
// @Produce      json
// @Success      200  {array}  contracts.ApplicationContract
// @Router       /v1/applications [get]
func ListApplications(w http.ResponseWriter, r *http.Request) {
	apps := []contracts.ApplicationContract{}
	for _, node := range GlobalGraph.Graph.Nodes {
		if node.Kind == "application" {
			contract, err := resources.LoadNode(node.Kind, node.Spec, contracts.Metadata{
				Name:  node.Metadata["name"].(string),
				Owner: node.Metadata["owner"].(string),
			})
			if err == nil {
				if app, ok := contract.(*contracts.ApplicationContract); ok {
					apps = append(apps, *app)
				}
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apps)
}

// GetApplication godoc
// @Summary      Get an application
// @Description  Returns a specific application by name
// @Tags         applications
// @Produce      json
// @Param        app_name  path      string  true  "Application name"
// @Success      200  {object}  contracts.ApplicationContract
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name} [get]
func GetApplication(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "app_name")
	node, ok := GlobalGraph.Graph.Nodes[appName]
	if !ok || node.Kind != "application" {
		WriteJSONError(w, "Application not found", http.StatusNotFound)
		return
	}
	contract, err := resources.LoadNode(node.Kind, node.Spec, contracts.Metadata{
		Name:  node.Metadata["name"].(string),
		Owner: node.Metadata["owner"].(string),
	})
	if err != nil {
		WriteJSONError(w, "Invalid application contract", http.StatusInternalServerError)
		return
	}
	app, ok := contract.(*contracts.ApplicationContract)
	if !ok {
		WriteJSONError(w, "Invalid application contract", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(app)
}

// UpdateApplication godoc
// @Summary      Update an application
// @Description  Updates an existing application resource
// @Tags         applications
// @Accept       json
// @Produce      json
// @Param        app_name     path      string                        true  "Application name"
// @Param        application  body      contracts.ApplicationContract true  "Application payload"
// @Success      200  {object}  contracts.ApplicationContract
// @Failure      400  {object}  map[string]string
// @Router       /v1/applications/{app_name} [put]
func UpdateApplication(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "app_name")
	var app contracts.ApplicationContract
	if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if app.Metadata.Name != appName {
		WriteJSONError(w, "Application name mismatch", http.StatusBadRequest)
		return
	}
	if err := app.Validate(); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	node, _ := graph.ResolveContract(app)
	GlobalGraph.AddNode(node)
	if err := GlobalGraph.Save(); err != nil {
		WriteJSONError(w, "Failed to save application", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(app)
}
