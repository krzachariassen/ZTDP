package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/krzachariassen/ZTDP/api/handlers"
	"github.com/krzachariassen/ZTDP/api/server"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/policies"
)

// --- Test router setup with backend selection ---
func newTestRouter(t *testing.T) http.Handler {
	var backend graph.GraphBackend
	switch os.Getenv("ZTDP_GRAPH_BACKEND") {
	case "redis":
		backend = graph.NewRedisGraph(graph.RedisGraphConfig{})
	default:
		t.Logf("⚙️  Using backend: Memory")
		backend = graph.NewMemoryGraph()
	}

	handlers.GlobalGraph = graph.NewGlobalGraph(backend)
	// Set up a fresh policy registry for each test
	reg := policies.NewDefaultPolicyRegistry()
	graph.SetPolicyRegistry(reg)
	return server.NewRouter()
}

// --- Test HTTP helper functions ---
func createApplication(t *testing.T, router http.Handler, appName string) {
	app := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":  appName,
			"owner": "team-x",
		},
		"spec": map[string]interface{}{
			"description": "Test application",
			"tags":        []string{"test"},
			"lifecycle":   map[string]interface{}{},
		},
	}
	body, _ := json.Marshal(app)
	req := httptest.NewRequest("POST", "/v1/applications", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated && resp.Code != http.StatusConflict {
		t.Fatalf("failed to create application %s, status: %d", appName, resp.Code)
	}
}

func createService(t *testing.T, router http.Handler, appName, serviceName string) {
	service := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": serviceName,
		},
		"spec": map[string]interface{}{
			"application": appName,
			"port":        8080,
			"public":      false,
			"description": "Test service",
			"tags":        []string{"test"},
		},
	}
	body, _ := json.Marshal(service)
	url := "/v1/applications/" + appName + "/services"
	req := httptest.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated && resp.Code != http.StatusConflict {
		t.Fatalf("failed to create service %s, status: %d", serviceName, resp.Code)
	}
}

func createEnvironment(t *testing.T, router http.Handler, envName string) {
	env := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": envName,
		},
		"spec": map[string]interface{}{},
	}
	body, _ := json.Marshal(env)
	req := httptest.NewRequest("POST", "/v1/environments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated && resp.Code != http.StatusConflict {
		t.Fatalf("failed to create environment %s, status: %d", envName, resp.Code)
	}
}

func createServiceVersion(t *testing.T, router http.Handler, appName, serviceName, version string) {
	ver := map[string]interface{}{
		"version": version,
		"spec": map[string]interface{}{
			"description": "Test version",
		},
	}
	body, _ := json.Marshal(ver)
	url := "/v1/applications/" + appName + "/services/" + serviceName + "/versions"
	req := httptest.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated && resp.Code != http.StatusConflict {
		t.Fatalf("failed to create service version %s, status: %d", version, resp.Code)
	}
}

func deployServiceVersion(t *testing.T, router http.Handler, appName, serviceName, version, env string) {
	payload := map[string]interface{}{"environment": env}
	body, _ := json.Marshal(payload)
	url := "/v1/applications/" + appName + "/services/" + serviceName + "/versions/" + version + "/deploy"
	req := httptest.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated && resp.Code != http.StatusConflict {
		t.Fatalf("failed to deploy service version %s to env %s, status: %d", version, env, resp.Code)
	}
}

func addAllowedEnvironments(t *testing.T, router http.Handler, appName string, envs []string) {
	body, _ := json.Marshal(envs)
	url := "/v1/applications/" + appName + "/environments/allowed"
	req := httptest.NewRequest("PUT", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK && resp.Code != http.StatusConflict {
		t.Fatalf("failed to set allowed environments for %s, status: %d", appName, resp.Code)
	}
}

// --- Focused setup helpers ---
func setupApplications(t *testing.T, router http.Handler) {
	createApplication(t, router, "checkout")
}

func setupServices(t *testing.T, router http.Handler) {
	createService(t, router, "checkout", "checkout-api")
	createService(t, router, "checkout", "checkout-worker")
}

func setupEnvironments(t *testing.T, router http.Handler) {
	createEnvironment(t, router, "dev")
	createEnvironment(t, router, "prod")
}

func setupAllowedEnvironments(t *testing.T, router http.Handler) {
	addAllowedEnvironments(t, router, "checkout", []string{"dev", "prod"})
}

func setupServiceVersions(t *testing.T, router http.Handler) {
	createServiceVersion(t, router, "checkout", "checkout-api", "1.0.0")
	createServiceVersion(t, router, "checkout", "checkout-worker", "1.0.0")
}

func setupDeployments(t *testing.T, router http.Handler) {
	deployServiceVersion(t, router, "checkout", "checkout-api", "1.0.0", "dev")
	deployServiceVersion(t, router, "checkout", "checkout-api", "1.0.0", "prod")
	deployServiceVersion(t, router, "checkout", "checkout-worker", "1.0.0", "dev")
}

// --- Per-test setup ---

func TestCreateAndGetApplication(t *testing.T) {
	router := newTestRouter(t)
	setupApplications(t, router)

	req := httptest.NewRequest("GET", "/v1/applications/checkout", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
}

func TestListApplications(t *testing.T) {
	router := newTestRouter(t)
	setupApplications(t, router)

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
	router := newTestRouter(t)
	setupApplications(t, router)

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
	router := newTestRouter(t)
	setupApplications(t, router)
	setupServices(t, router)

	// Now test GET
	req := httptest.NewRequest("GET", "/v1/applications/checkout/services/checkout-api", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
}

func TestListServices(t *testing.T) {
	router := newTestRouter(t)
	setupApplications(t, router)
	setupServices(t, router)

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
	router := newTestRouter(t)
	setupApplications(t, router)
	setupServices(t, router)
	setupEnvironments(t, router)
	setupAllowedEnvironments(t, router)
	setupServiceVersions(t, router)
	setupDeployments(t, router)

	req := httptest.NewRequest("POST", "/v1/apply?env=dev", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
}

func TestGetGrap(t *testing.T) {
	router := newTestRouter(t)
	setupApplications(t, router)
	setupServices(t, router)
	setupEnvironments(t, router)
	setupAllowedEnvironments(t, router)
	setupServiceVersions(t, router)
	setupDeployments(t, router)

	req := httptest.NewRequest("GET", "/v1/graph", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Error("expected non-empty graph response")
	}
	// Print the graph JSON for debugging
	t.Logf("Graph JSON: %s", string(body))
}

func TestHealthz(t *testing.T) {
	router := newTestRouter(t)
	req := httptest.NewRequest("GET", "/v1/healthz", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
}

func TestStatusEndpoint(t *testing.T) {
	router := newTestRouter(t)
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
	router := newTestRouter(t)
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
	router := newTestRouter(t)
	setupApplications(t, router)
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
	router := newTestRouter(t)
	setupEnvironments(t, router)

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

// --- Policy enforcement tests ---
func TestDisallowDirectProductionDeployment(t *testing.T) {
	router := newTestRouter(t)
	setupApplications(t, router)
	setupServices(t, router)
	setupEnvironments(t, router)
	setupAllowedEnvironments(t, router)

	// Only deploy to prod, skip dev
	createServiceVersion(t, router, "checkout", "checkout-api", "2.0.0")
	resp := httptest.NewRecorder()
	payload := map[string]interface{}{"environment": "prod"}
	body, _ := json.Marshal(payload)
	url := "/v1/applications/checkout/services/checkout-api/versions/2.0.0/deploy"
	req := httptest.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden && resp.Code != http.StatusBadRequest {
		t.Errorf("expected forbidden or bad request when deploying directly to production, got %d", resp.Code)
	}
}

func TestDisallowDeploymentToNotAllowedEnv(t *testing.T) {
	router := newTestRouter(t)
	setupApplications(t, router)
	setupServices(t, router)
	setupEnvironments(t, router)
	addAllowedEnvironments(t, router, "checkout", []string{"dev"})
	createServiceVersion(t, router, "checkout", "checkout-api", "3.0.0")
	// Should succeed: deploy to allowed env (dev)
	deployServiceVersion(t, router, "checkout", "checkout-api", "3.0.0", "dev")
	// Should fail: deploy to not-allowed env (prod)
	resp := httptest.NewRecorder()
	payload := map[string]interface{}{"environment": "prod"}
	body, _ := json.Marshal(payload)
	url := "/v1/applications/checkout/services/checkout-api/versions/3.0.0/deploy"
	req := httptest.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden && resp.Code != http.StatusBadRequest {
		t.Errorf("expected forbidden or bad request when deploying to not-allowed environment, got %d", resp.Code)
	}
}
