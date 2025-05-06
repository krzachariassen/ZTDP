package handlers

import (
	"net/http"
)

// GetGraph godoc
// @Summary      Get the current graph
// @Description  Returns the current dependency graph
// @Tags         graph
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /v1/graph [get]
func GetGraph(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Graph details"))
}
