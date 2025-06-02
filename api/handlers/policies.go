package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/krzachariassen/ZTDP/internal/policies"
)

// PolicyHandler godoc
// @Summary      Manage policies in the graph
// @Description  Create, check, and satisfy policies
// @Tags         policies
// @Accept       json
// @Produce      json
// @Param        body body policies.PolicyOperationRequest true "Policy request"
// @Success      200  {object}  policies.PolicyOperationResponse
// @Failure      400  {object}  map[string]string
// @Router       /v1/policies [post]
func PolicyHandler(w http.ResponseWriter, r *http.Request) {
	// Parse environment from query parameter or use default
	env := r.URL.Query().Get("environment")
	if env == "" {
		env = "default"
	}

	// Parse and validate request body
	var req policies.PolicyOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid request format: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Operation == "" {
		WriteJSONError(w, "Operation is required", http.StatusBadRequest)
		return
	}

	// Get user from request
	user := getUserFromRequest(r)

	// Create policy service
	policyService := policies.NewService(getGraphStore(), env)

	// Execute operation - let service handle all business logic
	response, err := policyService.ExecuteOperation(req, user)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Service determines the response format and status
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		WriteJSONError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ListPolicies returns all policies
func ListPolicies(w http.ResponseWriter, r *http.Request) {
	// Parse environment from query parameter or use default
	env := r.URL.Query().Get("environment")
	if env == "" {
		env = "default"
	}

	// Create policy service
	policyService := policies.NewService(getGraphStore(), env)

	// Get policies
	policyList, err := policyService.ListPolicies()
	if err != nil {
		WriteJSONError(w, "Failed to get policies: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policyList)
}

// GetPolicy returns a policy by ID
func GetPolicy(w http.ResponseWriter, r *http.Request) {
	policyID := r.URL.Query().Get("policy_id")
	if policyID == "" {
		// Try chi URL param if available
		policyID = chi.URLParam(r, "policy_id")
	}
	if policyID == "" {
		WriteJSONError(w, "policy_id is required", http.StatusBadRequest)
		return
	}

	// Parse environment from query parameter or use default
	env := r.URL.Query().Get("environment")
	if env == "" {
		env = "default"
	}

	// Create policy service
	policyService := policies.NewService(getGraphStore(), env)

	// Get policy
	policy, err := policyService.GetPolicy(policyID)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policy)
}

// Helper function to get user from a request
func getUserFromRequest(r *http.Request) string {
	// In a real system, this would use authentication
	user := r.Header.Get("X-User")
	if user == "" {
		return "system"
	}
	return user
}
