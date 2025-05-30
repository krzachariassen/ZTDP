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
	SetupRoutes(r)
	return r
}

func SetupRoutes(r *chi.Mux) {
	r.Route("/v1", func(v1 chi.Router) {
		// =============================================================================
		// SYSTEM ENDPOINTS
		// =============================================================================
		v1.Get("/healthz", handlers.HealthCheck)
		v1.Get("/status", handlers.Status)
		v1.Get("/graph", handlers.GetGraph)

		// =============================================================================
		// APPLICATION MANAGEMENT
		// =============================================================================
		v1.Post("/applications", handlers.CreateApplication)
		v1.Get("/applications", handlers.ListApplications)
		v1.Get("/applications/{app_name}", handlers.GetApplication)
		v1.Put("/applications/{app_name}", handlers.UpdateApplication)
		v1.Get("/applications/schema", handlers.ApplicationSchema)

		// Application Deployment (Primary Interface)
		v1.Post("/applications/{app_name}/deploy", handlers.DeployApplication)

		// Application-Environment Policies
		v1.Post("/applications/{app_name}/environments/{env_name}/allowed", handlers.LinkAppAllowedInEnvironment)
		v1.Get("/applications/{app_name}/environments/allowed", handlers.ListAllowedEnvironments)
		v1.Put("/applications/{app_name}/environments/allowed", handlers.UpdateAllowedEnvironments)
		v1.Post("/applications/{app_name}/environments/allowed", handlers.AddAllowedEnvironments)

		// =============================================================================
		// SERVICE MANAGEMENT
		// =============================================================================
		v1.Post("/applications/{app_name}/services", handlers.CreateService)
		v1.Get("/applications/{app_name}/services", handlers.ListServices)
		v1.Get("/applications/{app_name}/services/{service_name}", handlers.GetService)
		v1.Get("/applications/{app_name}/services/schema", handlers.ServiceSchema)

		// Service Versioning
		v1.Post("/applications/{app_name}/services/{service_name}/versions", handlers.CreateServiceVersion)
		v1.Get("/applications/{app_name}/services/{service_name}/versions", handlers.ListServiceVersions)

		// =============================================================================
		// ENVIRONMENT MANAGEMENT
		// =============================================================================
		v1.Post("/environments", handlers.CreateEnvironment)
		v1.Get("/environments", handlers.ListEnvironments)

		// =============================================================================
		// RESOURCE MANAGEMENT
		// =============================================================================
		v1.Post("/resources", handlers.CreateResource)
		v1.Get("/resources", handlers.ListResources)
		v1.Post("/applications/{app_name}/resources/{resource_name}", handlers.AddResourceToApplication)
		v1.Get("/applications/{app_name}/resources", handlers.ListApplicationResources)
		v1.Post("/applications/{app_name}/services/{service_name}/resources/{resource_name}", handlers.LinkServiceToResource)
		v1.Get("/applications/{app_name}/services/{service_name}/resources", handlers.ListServiceResources)

		// =============================================================================
		// POLICY MANAGEMENT
		// =============================================================================
		v1.Post("/policies", handlers.PolicyHandler)
		v1.Get("/policies", handlers.ListPolicies)
		v1.Get("/policies/{policy_id}", handlers.GetPolicy)

		// =============================================================================
		// REAL-TIME LOGS & EVENTS
		// =============================================================================
		v1.Get("/logs/stream", handlers.LogsWebSocket)
	})

	// =============================================================================
	// STATIC CONTENT & DOCUMENTATION
	// =============================================================================
	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Handle("/graph.html", http.FileServer(http.Dir("static")))
	r.Handle("/graph-modern.html", http.FileServer(http.Dir("static")))
	r.Handle("/graph-modern.css", http.FileServer(http.Dir("static")))
}
