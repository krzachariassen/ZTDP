package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/krzachariassen/ZTDP/api/handlers"
	"github.com/krzachariassen/ZTDP/api/server"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

func setupTestRouter() http.Handler {
	backend := graph.NewMemoryGraph()
	handlers.GlobalGraph = graph.NewGlobalGraph(backend)

	// Set up the event system
	eventTransport := events.NewMemoryTransport()
	events.InitializeEventBus(eventTransport)

	return server.NewRouter()
}

func createTestResource(router http.Handler, name string) {
	body := `{
		"kind": "resource",
		"metadata": {"name": "` + name + `"},
		"spec": {"type": "postgres", "config_ref": "config/test"}
	}`
	req := httptest.NewRequest("POST", "/v1/resources", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated && resp.Code != http.StatusConflict {
		panic(fmt.Sprintf("Failed to create resource: %d", resp.Code))
	}
}

func createTestApplication(router http.Handler, name string) {
	body := `{
		"metadata": {"name": "` + name + `", "owner": "test-team"},
		"spec": {"description": "Test app"}
	}`
	req := httptest.NewRequest("POST", "/v1/applications", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated && resp.Code != http.StatusConflict {
		panic(fmt.Sprintf("Failed to create application: %d", resp.Code))
	}
}

func addResourceToApp(router http.Handler, appName, resourceName string) {
	url := "/v1/applications/" + appName + "/resources/" + resourceName
	req := httptest.NewRequest("POST", url, nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		panic(fmt.Sprintf("Failed to add resource to app: %d - %s", resp.Code, resp.Body.String()))
	}
}

func getGraph(router http.Handler) map[string]interface{} {
	req := httptest.NewRequest("GET", "/v1/graph", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		panic(fmt.Sprintf("Failed to get graph: %d", resp.Code))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}

func main() {
	fmt.Println("🧪 Testing Resource Instance Edge Structure")

	router := setupTestRouter()

	// 1. Create a resource in the catalog
	fmt.Println("📦 Creating resource 'test-db' in catalog...")
	createTestResource(router, "test-db")

	// 2. Create an application
	fmt.Println("🏗️  Creating application 'test-app'...")
	createTestApplication(router, "test-app")

	// 3. Add resource to application (this should create instance + edges)
	fmt.Println("🔗 Adding resource to application...")
	addResourceToApp(router, "test-app", "test-db")

	// 4. Get graph and analyze structure
	fmt.Println("📊 Analyzing graph structure...")
	graphData := getGraph(router)

	nodes := graphData["nodes"].(map[string]interface{})
	edges := graphData["edges"].(map[string]interface{})

	fmt.Println("\n=== NODES ===")
	var resourceInstances []string
	for nodeID, nodeData := range nodes {
		node := nodeData.(map[string]interface{})
		kind := node["kind"].(string)
		fmt.Printf("Node: %s (kind: %s)\n", nodeID, kind)

		// Look for resource instances (should have format test-app-test-db-<uuid>)
		if kind == "resource" && strings.HasPrefix(nodeID, "test-app-test-db-") {
			resourceInstances = append(resourceInstances, nodeID)
			metadata := node["metadata"].(map[string]interface{})
			fmt.Printf("  ✅ Found resource instance: %s\n", nodeID)
			fmt.Printf("     - catalog_ref: %v\n", metadata["catalog_ref"])
			fmt.Printf("     - application: %v\n", metadata["application"])
		}
	}

	fmt.Println("\n=== EDGES ===")
	var foundInstanceOfEdge, foundOwnsEdge bool

	for fromID, edgeList := range edges {
		edgeArray := edgeList.([]interface{})
		for _, edgeData := range edgeArray {
			edge := edgeData.(map[string]interface{})
			toID := edge["to"].(string)
			edgeType := edge["type"].(string)

			fmt.Printf("Edge: %s --%s--> %s\n", fromID, edgeType, toID)

			// Check for instance_of edge from resource instance to catalog resource
			for _, instanceID := range resourceInstances {
				if fromID == instanceID && toID == "test-db" && edgeType == "instance_of" {
					foundInstanceOfEdge = true
					fmt.Printf("  ✅ Found instance_of edge: %s -> %s\n", fromID, toID)
				}
			}

			// Check for owns edge from application to resource instance
			for _, instanceID := range resourceInstances {
				if fromID == "test-app" && toID == instanceID && edgeType == "owns" {
					foundOwnsEdge = true
					fmt.Printf("  ✅ Found owns edge: %s -> %s\n", fromID, toID)
				}
			}
		}
	}

	fmt.Println("\n=== VERIFICATION ===")

	if len(resourceInstances) == 0 {
		fmt.Println("❌ ERROR: No resource instances found!")
		os.Exit(1)
	} else {
		fmt.Printf("✅ Found %d resource instance(s)\n", len(resourceInstances))
	}

	if !foundInstanceOfEdge {
		fmt.Println("❌ ERROR: instance_of edge not found!")
		os.Exit(1)
	} else {
		fmt.Println("✅ instance_of edge found")
	}

	if !foundOwnsEdge {
		fmt.Println("❌ ERROR: owns edge from app to instance not found!")
		os.Exit(1)
	} else {
		fmt.Println("✅ owns edge from app to instance found")
	}

	// Check that there's NO direct owns edge from app to catalog resource
	if appEdges, exists := edges["test-app"]; exists {
		edgeArray := appEdges.([]interface{})
		for _, edgeData := range edgeArray {
			edge := edgeData.(map[string]interface{})
			if edge["to"].(string) == "test-db" && edge["type"].(string) == "owns" {
				fmt.Println("❌ ERROR: Found direct owns edge from app to catalog resource (should not exist!)")
				os.Exit(1)
			}
		}
	}
	fmt.Println("✅ No direct owns edge from app to catalog resource")

	fmt.Println("\n🎉 ALL CHECKS PASSED! Resource instance structure is correct.")
}
