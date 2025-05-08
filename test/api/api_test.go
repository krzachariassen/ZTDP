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
			"description": "Handles " + name + " flows",
			"tags":        []string{"payments", "frontend"},
			"lifecycle":   map[string]interface{}{},
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

func createEnvironment(t *testing.T, router http.Handler, name string) {
	env := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":  name,
			"owner": "platform-team",
		},
		"spec": map[string]interface{}{
			"description": "Test environment: " + name,
		},
	}
	body, _ := json.Marshal(env)
	req := httptest.NewRequest("POST", "/v1/environments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("failed to create environment: %s, status: %d", name, resp.Code)
	}
}

func linkServiceToEnvironment(t *testing.T, router http.Handler, appName, svcName, envName string) {
	url := fmt.Sprintf("/v1/applications/%s/services/%s/environments/%s", appName, svcName, envName)
	req := httptest.NewRequest("POST", url, nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("failed to link service %s to environment %s, status: %d", svcName, envName, resp.Code)
	}
}

func linkAppAllowedInEnvironment(t *testing.T, router http.Handler, appName, envName string) {
	url := fmt.Sprintf("/v1/applications/%s/environments/%s/allowed", appName, envName)
	req := httptest.NewRequest("POST", url, nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("failed to link allowed_in: %s -> %s, status: %d", appName, envName, resp.Code)
	}
}

func createServiceVersion(t *testing.T, router http.Handler, appName, svcName, version string) {
	ver := map[string]interface{}{
		"version":    version,
		"config_ref": "default-config",
		"owner":      "team-x",
	}
	url := fmt.Sprintf("/v1/applications/%s/services/%s/versions", appName, svcName)
	body, _ := json.Marshal(ver)
	req := httptest.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated && resp.Code != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("failed to create service version: %s, status: %d, body: %s", version, resp.Code, string(respBody))
	}
}

func deployServiceVersion(t *testing.T, router http.Handler, appName, svcName, version, envName string) {
	payload := map[string]interface{}{"environment": envName}
	url := fmt.Sprintf("/v1/applications/%s/services/%s/versions/%s/deploy", appName, svcName, version)
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("failed to deploy service version %s to env %s, status: %d", version, envName, resp.Code)
	}
}

func createTestData(t *testing.T, router http.Handler) {
	createApplication(t, router, "checkout")
	createService(t, router, "checkout", "checkout-api")
	createService(t, router, "checkout", "checkout-worker")
	createEnvironment(t, router, "dev")
	createEnvironment(t, router, "prod")
	linkAppAllowedInEnvironment(t, router, "checkout", "dev")
	linkAppAllowedInEnvironment(t, router, "checkout", "prod")
	// Create service versions and deploy them
	createServiceVersion(t, router, "checkout", "checkout-api", "1.0.0")
	createServiceVersion(t, router, "checkout", "checkout-worker", "1.0.0")
	deployServiceVersion(t, router, "checkout", "checkout-api", "1.0.0", "dev")
	deployServiceVersion(t, router, "checkout", "checkout-api", "1.0.0", "prod")
	deployServiceVersion(t, router, "checkout", "checkout-worker", "1.0.0", "dev")
}

func setupTestData(t *testing.T, router http.Handler) {
	createTestData(t, router)
}

func TestCreateAndGetApplication(t *testing.T) {
	router := server.NewRouter()
	setupTestData(t, router)

	req := httptest.NewRequest("GET", "/v1/applications/checkout", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
}

func TestListApplications(t *testing.T) {
	router := server.NewRouter()
	setupTestData(t, router)

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
	setupTestData(t, router)

	app := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":  "checkout",
			"owner": "team-x",
		},
		"spec": map[string]interface{}{
			"description": "Handles checkout flows - updated",
			"tags":        []string{"payments", "frontend"},
			"lifecycle":   map[string]interface{}{},
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
	setupTestData(t, router)

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
	setupTestData(t, router)

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
	setupTestData(t, router)

	req := httptest.NewRequest("POST", "/v1/apply?env=dev", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
}

func TestGetGraph(t *testing.T) {
	router := server.NewRouter()
	setupTestData(t, router)

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
	// Use the new endpoint under the application scope
	req := httptest.NewRequest("GET", "/v1/applications/checkout/services/schema", nil)
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

func TestCreateAndListEnvironments(t *testing.T) {
	router := server.NewRouter()
	setupTestData(t, router)

	req := httptest.NewRequest("GET", "/v1/environments", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Error("expected non-empty environments list")
	}
}

func TestCreateAndLinkServiceToEnvironment(t *testing.T) {
	router := server.NewRouter()
	setupTestData(t, router)
	// Already linked in setupTestData, so just verify no error on relink
	linkServiceToEnvironment(t, router, "checkout", "checkout-api", "dev")
}

func TestAllowedInPolicyEdge(t *testing.T) {
	router := server.NewRouter()
	setupTestData(t, router)
	linkAppAllowedInEnvironment(t, router, "checkout", "prod")
}

func TestAllowedEnvironmentsPolicyAPI(t *testing.T) {
	router := server.NewRouter()
	setupTestData(t, router)
	// Set allowed environments to only dev for checkout
	putBody, _ := json.Marshal([]string{"dev"})
	req := httptest.NewRequest("PUT", "/v1/applications/checkout/environments/allowed", bytes.NewBuffer(putBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("failed to set allowed environments, status: %d", resp.Code)
	}
	// Should succeed: link service to allowed env (dev)
	linkServiceToEnvironment(t, router, "checkout", "checkout-api", "dev")
	// Should fail: link service to not-allowed env (prod)
	url := "/v1/applications/checkout/services/checkout-api/environments/prod"
	req = httptest.NewRequest("POST", url, nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code == http.StatusCreated {
		t.Error("expected failure when linking service to not-allowed environment, but got success")
	}
}

func TestListServiceVersions(t *testing.T) {
	router := server.NewRouter()
	setupTestData(t, router)

	// List versions for checkout-api
	url := "/v1/applications/checkout/services/checkout-api/versions"
	req := httptest.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
	var versions []map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&versions)
	if len(versions) == 0 {
		t.Error("expected at least one service version for checkout-api")
	}
}

func TestListEnvironmentDeployments(t *testing.T) {
	router := server.NewRouter()
	setupTestData(t, router)

	// List deployments in dev environment
	url := "/v1/environments/dev/deployments"
	req := httptest.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
	var deployments []map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&deployments)
	if len(deployments) == 0 {
		t.Error("expected at least one deployment in dev environment")
	}
}
