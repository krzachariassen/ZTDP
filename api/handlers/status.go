package handlers

import (
	"encoding/json"
	"net/http"
)

// Status godoc
// @Summary      Get platform status
// @Description  Returns high-level platform status and graph node count
// @Tags         status
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /v1/status [get]
func Status(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"graph_nodes": len(GlobalGraph.Graph.Nodes),
		// Add more fields as needed
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
