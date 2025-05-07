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
		// Services
		v1.Post("/applications/{app_name}/services", handlers.CreateService)
		v1.Get("/applications/{app_name}/services", handlers.ListServices)
		v1.Get("/applications/{app_name}/services/{service_name}", handlers.GetService)
		v1.Get("/services/schema", handlers.ServiceSchema)
		// Swagger UI
		r.Get("/swagger/*", httpSwagger.WrapHandler)
	})
	return r
}
