package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/krzachariassen/ZTDP/api/handlers"
	"github.com/krzachariassen/ZTDP/api/server"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

func init() {
	var backend graph.GraphBackend
	switch os.Getenv("ZTDP_GRAPH_BACKEND") {
	case "redis":
		backend = graph.NewRedisGraph(graph.RedisGraphConfig{})
	default:
		fmt.Println("⚙️  Using backend: Memory")
		backend = graph.NewMemoryGraph()
	}
	handlers.GlobalGraph = graph.NewGlobalGraph(backend)
}

func createApplication(t *testing.T, router http.Handler, name string) {
	app := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":  name,
			"owner": "team-x",
		},
		"spec": map[string]interface{}{
			"description":  "Handles " + name + " flows",
			"tags":         []string{"payments", "frontend"},
			"environments": []string{"dev", "qa"},
			"lifecycle":    map[string]interface{}{},
		},
	}
	body, _ := json.Marshal(app)
	req := httptest.NewRequest("POST", "/v1/applications", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("failed to create application: %s, status: %d", name, resp.Code)
	}
}

func createService(t *testing.T, router http.Handler, appName, svcName string) {
	svc := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":  svcName,
			"owner": "team-x",
		},
		"spec": map[string]interface{}{
			"application": appName,
			"port":        8080,
			"public":      true,
		},
	}
	body, _ := json.Marshal(svc)
	req := httptest.NewRequest("POST", fmt.Sprintf("/v1/applications/%s/services", appName), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("failed to create service: %s, status: %d", svcName, resp.Code)
	}
}

func TestCreateAndGetApplication(t *testing.T) {
	router := server.NewRouter()
	createApplication(t, router, "checkout")

	req := httptest.NewRequest("GET", "/v1/applications/checkout", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
}

func TestListApplications(t *testing.T) {
	router := server.NewRouter()
	createApplication(t, router, "checkout")
	createApplication(t, router, "billing")

	req := httptest.NewRequest("GET", "/v1/applications", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Error("expected non-empty applications list")
	}
}

func TestUpdateApplication(t *testing.T) {
	router := server.NewRouter()
	createApplication(t, router, "checkout")

	app := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":  "checkout",
			"owner": "team-x",
		},
		"spec": map[string]interface{}{
			"description":  "Handles checkout flows - updated",
			"tags":         []string{"payments", "frontend"},
			"environments": []string{"dev", "qa"},
			"lifecycle":    map[string]interface{}{},
		},
	}
	body, _ := json.Marshal(app)
	req := httptest.NewRequest("PUT", "/v1/applications/checkout", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
}

func TestCreateAndGetService(t *testing.T) {
	router := server.NewRouter()
	createApplication(t, router, "checkout")
	createService(t, router, "checkout", "checkout-api")

	// Now test GET
	req := httptest.NewRequest("GET", "/v1/applications/checkout/services/checkout-api", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
}

func TestListServices(t *testing.T) {
	router := server.NewRouter()
	createApplication(t, router, "checkout")
	createService(t, router, "checkout", "checkout-api")
	createService(t, router, "checkout", "checkout-worker")

	req := httptest.NewRequest("GET", "/v1/applications/checkout/services", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Error("expected non-empty services list")
	}
}

func TestApplyGraph(t *testing.T) {
	router := server.NewRouter()
	createApplication(t, router, "checkout")
	req := httptest.NewRequest("POST", "/v1/apply?env=dev", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
}

func TestGetGraph(t *testing.T) {
	router := server.NewRouter()
	createApplication(t, router, "checkout")
	req := httptest.NewRequest("GET", "/v1/graph?env=dev", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
}

func TestHealthz(t *testing.T) {
	router := server.NewRouter()
	req := httptest.NewRequest("GET", "/v1/healthz", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
}

func TestStatusEndpoint(t *testing.T) {
	router := server.NewRouter()
	req := httptest.NewRequest("GET", "/v1/status", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Error("expected non-empty status response")
	}
}

func TestGetApplicationSchema(t *testing.T) {
	router := server.NewRouter()
	req := httptest.NewRequest("GET", "/v1/applications/schema", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Error("expected non-empty application schema response")
	}
}

func TestGetServiceSchema(t *testing.T) {
	router := server.NewRouter()
	req := httptest.NewRequest("GET", "/v1/services/schema", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Error("expected non-empty service schema response")
	}
}
