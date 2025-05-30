package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/krzachariassen/ZTDP/internal/application"
	"github.com/krzachariassen/ZTDP/internal/contracts"
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

	// Create application service - simple and clean!
	appService := application.NewService(GlobalGraph)

	if err := appService.CreateApplication(app); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
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
	appService := application.NewService(GlobalGraph)
	apps, err := appService.ListApplications()
	if err != nil {
		WriteJSONError(w, "Failed to get applications: "+err.Error(), http.StatusInternalServerError)
		return
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

	appService := application.NewService(GlobalGraph)
	app, err := appService.GetApplication(appName)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusNotFound)
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

	appService := application.NewService(GlobalGraph)

	if err := appService.UpdateApplication(appName, app); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(app)
}

// DeleteApplication godoc
// @Summary      Delete an application
// @Description  Deletes an existing application resource
// @Tags         applications
// @Param        app_name  path      string  true  "Application name"
// @Success      204
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name} [delete]
func DeleteApplication(w http.ResponseWriter, r *http.Request) {
	// NOT IMPLEMENTED
	//appName := chi.URLParam(r, "app_name")
	//
	//appService := application.NewService(GlobalGraph)
	//
	//if err := appService.DeleteApplication(appName); err != nil {
	//	WriteJSONError(w, err.Error(), http.StatusNotFound)
	//	return
	//}
	//
	//w.WriteHeader(http.StatusNoContent)
}
