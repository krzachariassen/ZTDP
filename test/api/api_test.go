package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krzachariassen/ZTDP/api/server"
)

func TestSubmitContract(t *testing.T) {
	router := server.NewRouter()

	contract := map[string]interface{}{
		"kind":  "application",
		"name":  "checkout",
		"owner": "team-a",
	}

	body, _ := json.Marshal(contract)
	req := httptest.NewRequest("POST", "/contracts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.Code)
	}
}

func TestApplyGraph(t *testing.T) {
	router := server.NewRouter()

	req := httptest.NewRequest("POST", "/apply?env=dev", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
}

func TestGetGraph(t *testing.T) {
	router := server.NewRouter()

	req := httptest.NewRequest("GET", "/graph?env=dev", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
}
