package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/krzachariassen/ZTDP/api/handlers"
	_ "github.com/krzachariassen/ZTDP/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()
	r.Route("/v1", func(v1 chi.Router) {
		v1.Get("/healthz", handlers.HealthCheck)
		v1.Post("/apply", handlers.ApplyGraph)
		v1.Get("/graph", handlers.GetGraph)
		v1.Get("/status", handlers.Status)
		// Applications
		v1.Post("/applications", handlers.CreateApplication)
		v1.Get("/applications", handlers.ListApplications)
		v1.Get("/applications/{app_name}", handlers.GetApplication)
		v1.Put("/applications/{app_name}", handlers.UpdateApplication)
		v1.Get("/applications/schema", handlers.ApplicationSchema)
		// Services should always be created under an application
		v1.Post("/applications/{app_name}/services", handlers.CreateService)
		v1.Get("/applications/{app_name}/services", handlers.ListServices)
		v1.Get("/applications/{app_name}/services/{service_name}", handlers.GetService)
		// Environments
		v1.Post("/environments", handlers.CreateEnvironment)
		v1.Get("/environments", handlers.ListEnvironments)
		// Link Service to Environment (deployed_in) - now nested under application
		v1.Post("/applications/{app_name}/services/{service_name}/environments/{env_name}", handlers.LinkServiceToEnvironment)
		// Application allowed_in policy edge
		v1.Post("/applications/{app_name}/environments/{env_name}/allowed", handlers.LinkAppAllowedInEnvironment)
		// List allowed environments for an application (policy)
		v1.Get("/applications/{app_name}/environments/allowed", handlers.ListAllowedEnvironments)
		// Update and Add allowed environments for an application
		v1.Put("/applications/{app_name}/environments/allowed", handlers.UpdateAllowedEnvironments)
		v1.Post("/applications/{app_name}/environments/allowed", handlers.AddAllowedEnvironments)
		// Service schema
		v1.Get("/applications/{app_name}/services/schema", handlers.ServiceSchema)
		// Swagger UI
		r.Get("/swagger/*", httpSwagger.WrapHandler)
		// Graph Visualization
		r.Handle("/graph.html", http.FileServer(http.Dir("static")))
	})
	return r
}
