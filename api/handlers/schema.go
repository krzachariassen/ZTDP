package handlers

import (
	"encoding/json"
	"net/http"
)

// ContractSchema godoc
// @Summary      Get contract schemas
// @Description  Returns example schemas for supported contract kinds
// @Tags         contracts
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /v1/contracts/schema [get]
func ContractSchema(w http.ResponseWriter, r *http.Request) {
	schemas := map[string]interface{}{
		"application": map[string]interface{}{
			"kind": "application",
			"metadata": map[string]interface{}{
				"name":  "string",
				"owner": "string",
			},
			"spec": map[string]interface{}{
				"description":  "string",
				"tags":         []string{"string"},
				"environments": []string{"string"},
				"lifecycle":    map[string]interface{}{},
			},
		},
		"service": map[string]interface{}{
			"kind": "service",
			"metadata": map[string]interface{}{
				"name":  "string",
				"owner": "string",
			},
			"spec": map[string]interface{}{
				"application": "string",
				"port":        8080,
				"public":      true,
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schemas)
}

// ApplicationSchema godoc
// @Summary      Get application contract schema
// @Description  Returns example schema for application contract
// @Tags         applications
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /v1/applications/schema [get]
func ApplicationSchema(w http.ResponseWriter, r *http.Request) {
	schema := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":  "string",
			"owner": "string",
		},
		"spec": map[string]interface{}{
			"description":  "string",
			"tags":         []string{"string"},
			"environments": []string{"string"},
			"lifecycle":    map[string]interface{}{},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schema)
}

// ServiceSchema godoc
// @Summary      Get service contract schema
// @Description  Returns example schema for service contract
// @Tags         services
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /v1/services/schema [get]
func ServiceSchema(w http.ResponseWriter, r *http.Request) {
	schema := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":  "string",
			"owner": "string",
		},
		"spec": map[string]interface{}{
			"application": "string",
			"port":        8080,
			"public":      true,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schema)
}
