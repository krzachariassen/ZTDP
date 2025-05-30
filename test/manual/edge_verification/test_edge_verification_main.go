package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/krzachariassen/ZTDP/api/handlers"
	"github.com/krzachariassen/ZTDP/api/server"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

func newTestRouter() http.Handler {
	var backend graph.GraphBackend
	switch os.Getenv("ZTDP_GRAPH_BACKEND") {
	case "redis":
		backend = graph.NewRedisGraph(graph.RedisGraphConfig{})
	default:
		fmt.Println("⚙️  Using backend: Memory")
		backend = graph.NewMemoryGraph()
	}

	// Set up the graph
	handlers.GlobalGraph = graph.NewGlobalGraph(backend)

	// Set up the event system like main.go does
	eventTransport := events.NewMemoryTransport()

	// Create event services
	events.InitializeEventBus(eventTransport)

	return server.NewRouter()
}

func createResource(router http.Handler, resourceName, resourceType, configRef string) error {
	resource := map[string]interface{}{
		"kind": "resource",
		"metadata": map[string]interface{}{
			"name": resourceName,
		},
		"spec": map[string]interface{}{
			"type":       resourceType,
			"config_ref": configRef,
		},
	}
	body, _ := json.Marshal(resource)
	req := httptest.NewRequest("POST", "/v1/resources", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated && resp.Code != http.StatusConflict {
		return fmt.Errorf("failed to create resource %s, status: %d", resourceName, resp.Code)
	}
	return nil
}

func getKeys(edges map[string][]graph.Edge) []string {
	keys := make([]string, 0, len(edges))
	for k := range edges {
		keys = append(keys, k)
	}
	return keys
}

func main() {
	router := newTestRouter()

	// Create a resource in the catalog
	if err := createResource(router, "pg-db", "postgres", "config/postgres/pg-db"); err != nil {
		fmt.Printf("Error creating resource: %v\n", err)
		return
	}

	// Create policy and check nodes to demonstrate policy edges
	policyNode := &graph.Node{
		ID:   "policy-dev-before-prod",
		Kind: "policy",
		Metadata: map[string]interface{}{
			"name":        "Must Deploy To Dev Before Prod",
			"description": "Requires deployment to dev before prod",
			"type":        "system",
			"status":      "active",
		},
		Spec: map[string]interface{}{},
	}
	handlers.GlobalGraph.AddNode(policyNode)

	checkNode := &graph.Node{
		ID:   "check-dev-deployment-checkout",
		Kind: "check",
		Metadata: map[string]interface{}{
			"name":   "Dev Deployment Check for Checkout",
			"type":   "deployment_prerequisite",
			"status": "pending",
		},
		Spec: map[string]interface{}{
			"application":  "checkout",
			"required_env": "dev",
			"target_env":   "prod",
		},
	}
	handlers.GlobalGraph.AddNode(checkNode)

	// Create satisfies edge from check to policy
	if err := handlers.GlobalGraph.AddEdge("check-dev-deployment-checkout", "policy-dev-before-prod", "satisfies"); err != nil {
		// For testing: ignore "edge already exists" errors since Redis persists data
		if err.Error() != "edge already exists" {
			fmt.Printf("Error creating satisfies edge: %v\n", err)
		}
	}

	// Save the graph
	if err := handlers.GlobalGraph.Save(); err != nil {
		fmt.Printf("Error saving graph: %v\n", err)
		return
	}

	// Check the edges
	edges, err := handlers.GlobalGraph.Edges()
	if err != nil {
		fmt.Printf("Error getting edges: %v\n", err)
		return
	}

	nodes, err := handlers.GlobalGraph.Nodes()
	if err != nil {
		fmt.Printf("Error getting nodes: %v\n", err)
		return
	}

	fmt.Println("=== Nodes ===")
	for id, node := range nodes {
		fmt.Printf("Node: %s, Kind: %s\n", id, node.Kind)
	}

	fmt.Println("\n=== Edges ===")
	for fromID, edgeList := range edges {
		for _, edge := range edgeList {
			fmt.Printf("Edge: %s --%s--> %s\n", fromID, edge.Type, edge.To)
		}
	}

	// Check specifically for resource_register edges
	fmt.Println("\n=== Resource Register Edges ===")
	resourceRegisterEdges, exists := edges["resource-catalog"]
	if !exists {
		fmt.Println("No edges found for resource-catalog")
		fmt.Printf("Available edge keys: %v\n", getKeys(edges))
	} else {
		fmt.Printf("resource-catalog has %d edges:\n", len(resourceRegisterEdges))
		for _, edge := range resourceRegisterEdges {
			fmt.Printf("  resource-catalog --%s--> %s\n", edge.Type, edge.To)
		}
	}

	// Check if pg-db has an "owns" edge from resource-catalog
	hasOwnsEdge := false
	if resourceRegisterEdges != nil {
		for _, edge := range resourceRegisterEdges {
			if edge.To == "pg-db" && edge.Type == "owns" {
				hasOwnsEdge = true
				break
			}
		}
	}

	fmt.Printf("\nResource 'pg-db' has 'owns' edge from resource-catalog: %v\n", hasOwnsEdge)
}
