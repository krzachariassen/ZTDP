package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/krzachariassen/ZTDP/api/handlers"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/healthz", handlers.HealthCheck)
	r.Post("/contracts", handlers.SubmitContract)
	r.Post("/apply", handlers.ApplyGraph)
	r.Get("/graph", handlers.GetGraph)
	return r
}
