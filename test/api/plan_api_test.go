package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/krzachariassen/ZTDP/api/handlers"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

func TestGetPlan_Endpoint(t *testing.T) {
	handlers.GlobalGraph = &graph.GlobalGraph{
		Graph: graph.NewGraph(),
	}

	req := httptest.NewRequest("GET", "/v1/applications/test-app/plan", nil)
	rw := httptest.NewRecorder()

	// Set chi route context with app_name param
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("app_name", "test-app")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Setup a minimal graph for the test
	handlers.GlobalGraph.Graph.Nodes = map[string]*graph.Node{
		"test-app": {ID: "test-app", Kind: "application", Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}},
		"svc":      {ID: "svc", Kind: "service", Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}},
	}
	handlers.GlobalGraph.Graph.Edges = map[string][]graph.Edge{
		"test-app": {{To: "svc", Type: "deploy"}},
	}

	handlers.GetPlan(rw, req)
	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}
	var order []string
	if err := json.Unmarshal(rw.Body.Bytes(), &order); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(order) != 2 {
		t.Errorf("expected 2 nodes in plan, got %d", len(order))
	}

	// Assert the plan order is as expected (test-app before svc)
	expectedOrder := []string{"test-app", "svc"}
	for i, id := range expectedOrder {
		if order[i] != id {
			t.Errorf("expected node %d to be %q, got %q", i, id, order[i])
		}
	}
}
