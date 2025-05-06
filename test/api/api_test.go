package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func TestSubmitContract(t *testing.T) {
	router := server.NewRouter()

	contract := map[string]interface{}{
		"kind": "application",
		"metadata": map[string]interface{}{
			"name":  "checkout",
			"owner": "team-a",
		},
		"spec": map[string]interface{}{
			"description":  "Handles checkout flows",
			"tags":         []string{"payments", "frontend"},
			"environments": []string{"dev", "qa"},
			"lifecycle":    map[string]interface{}{},
		},
	}

	body, _ := json.Marshal(contract)
	req := httptest.NewRequest("POST", "/v1/contracts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.Code)
	}
}

func TestSubmitContract_InvalidJSON(t *testing.T) {
	router := server.NewRouter()

	req := httptest.NewRequest("POST", "/v1/contracts", bytes.NewBuffer([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid JSON, got %d", resp.Code)
	}
}

func TestSubmitContract_MissingKind(t *testing.T) {
	router := server.NewRouter()

	contract := map[string]interface{}{
		"name":  "checkout",
		"owner": "team-a",
	}
	body, _ := json.Marshal(contract)
	req := httptest.NewRequest("POST", "/v1/contracts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing kind, got %d", resp.Code)
	}
}

func TestSubmitContract_UnknownKind(t *testing.T) {
	router := server.NewRouter()

	contract := map[string]interface{}{
		"kind":  "unknown",
		"name":  "checkout",
		"owner": "team-a",
	}
	body, _ := json.Marshal(contract)
	req := httptest.NewRequest("POST", "/v1/contracts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for unknown kind, got %d", resp.Code)
	}
}

func TestSubmitContract_ServiceContract(t *testing.T) {
	router := server.NewRouter()

	contract := map[string]interface{}{
		"kind": "service",
		"metadata": map[string]interface{}{
			"name":  "checkout-api",
			"owner": "team-a",
		},
		"spec": map[string]interface{}{
			"application": "checkout",
			"port":        8080,
			"public":      true,
		},
	}
	body, _ := json.Marshal(contract)
	req := httptest.NewRequest("POST", "/v1/contracts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Errorf("expected status 201 for service contract, got %d", resp.Code)
	}
}

func TestApplyGraph(t *testing.T) {
	router := server.NewRouter()

	req := httptest.NewRequest("POST", "/v1/apply?env=dev", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
}

func TestGetGraph(t *testing.T) {
	router := server.NewRouter()

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

func TestGetContractSchema(t *testing.T) {
	router := server.NewRouter()
	req := httptest.NewRequest("GET", "/v1/contracts/schema", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Error("expected non-empty schema response")
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
	body, _ := ioutil.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Error("expected non-empty status response")
	}
}
