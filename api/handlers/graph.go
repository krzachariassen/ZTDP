package handlers

import (
	"encoding/json"
	"net/http"
)

// GetGraph godoc
// @Summary      Get the current graph
// @Description  Loads the latest graph from the backend and returns it as JSON
// @Tags         graph
// @Produce      json
// @Param        env  query     string  false  "Environment name (optional)"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /v1/graph [get]
func GetGraph(w http.ResponseWriter, r *http.Request) {
	currentGraph, err := GlobalGraph.Graph()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to load graph from backend"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(currentGraph); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to encode graph"})
	}
}

// ReloadGraph godoc
// @Summary      Reload the graph from backend
// @Description  Gets the current graph state from the backend (always fresh in the new architecture)
// @Tags         graph
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /v1/graph/reload [post]
func ReloadGraph(w http.ResponseWriter, r *http.Request) {
	// In the new architecture, graph is always fresh from backend
	// So this just fetches the current state
	currentGraph, err := GlobalGraph.Graph()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to load graph from backend"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"message": "Graph loaded from backend",
		"nodes":   len(currentGraph.Nodes),
		"edges":   len(currentGraph.Edges),
	}
	json.NewEncoder(w).Encode(response)
}
