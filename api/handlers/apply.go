package handlers

import (
	"net/http"
)

// ApplyGraph godoc
// @Summary      Apply the dependency graph
// @Description  Applies the current dependency graph for the given environment
// @Tags         graph
// @Produce      json
// @Success      200  {string}  string  "Graph applied"
// @Failure      400  {object}  map[string]string
// @Router       /v1/apply [post]
func ApplyGraph(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Graph applied"))
}
