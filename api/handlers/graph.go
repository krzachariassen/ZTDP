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
	if err := GlobalGraph.Load(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to load graph from backend"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(GlobalGraph.Graph); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to encode graph"})
	}
}
