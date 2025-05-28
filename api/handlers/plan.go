package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/krzachariassen/ZTDP/internal/planner"
)

// GetPlan godoc
// @Summary      Get execution plan for an application
// @Description  Returns the topological execution order for a given application
// @Tags         planner
// @Produce      json
// @Param        app_name  path      string  true  "Application name"
// @Success      200  {array}  string
// @Failure      404  {object}  map[string]string
// @Router       /v1/applications/{app_name}/plan [get]
func GetPlan(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "app_name")
	if _, ok := GlobalGraph.Graph.Nodes[appName]; !ok {
		WriteJSONError(w, "Application not found", http.StatusNotFound)
		return
	}
	// For now, use all allowed edge types for planning
	edgeTypes := make([]string, 0, len(GlobalGraph.Graph.Edges))
	for t := range GlobalGraph.Graph.Edges {
		edgeTypes = append(edgeTypes, t)
	}
	// But for strictness, use the static allowed edge types
	// edgeTypes := []string{"deploy", "create", "satisfies", "enforces", "requires"}
	p := planner.NewPlanner(GlobalGraph.Graph)
	order, err := p.PlanWithEdgeTypes([]string{"deploy", "create", "satisfies", "enforces", "requires"})
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
