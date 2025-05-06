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
		v1.Post("/contracts", handlers.SubmitContract)
		v1.Post("/apply", handlers.ApplyGraph)
		v1.Get("/graph", handlers.GetGraph)
		v1.Get("/contracts/schema", handlers.ContractSchema)
		v1.Get("/status", handlers.Status)
		r.Get("/swagger/*", httpSwagger.WrapHandler)
	})
	return r
}
