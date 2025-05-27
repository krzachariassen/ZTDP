package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/krzachariassen/ZTDP/internal/policies"
)

// PolicyRequest represents a request to check or satisfy a policy
type PolicyRequest struct {
	Operation string `json:"operation"` // "check", "create_policy", "create_check", "update_check", "satisfy"
	FromID    string `json:"from_id,omitempty"`
	ToID      string `json:"to_id,omitempty"`
	EdgeType  string `json:"edge_type,omitempty"`
	PolicyID  string `json:"policy_id,omitempty"`
	CheckID   string `json:"check_id,omitempty"`

	// For create operations
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Results     map[string]interface{} `json:"results,omitempty"`
}

// PolicyHandler godoc
// @Summary      Manage policies in the graph
// @Description  Create, check, and satisfy policies
// @Tags         policies
// @Accept       json
// @Produce      json
// @Param        body body PolicyRequest true "Policy request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/policies [post]
func PolicyHandler(w http.ResponseWriter, r *http.Request) {
	// Parse environment from query parameter or use default
	env := r.URL.Query().Get("environment")
	if env == "" {
		env = "default"
	}

	// Parse request body
	var req PolicyRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request format: %v", err), http.StatusBadRequest)
		return
	}

	// Get the evaluator
	evaluator := getPolicyEvaluator(env)

	// Process based on operation
	switch req.Operation {
	case "check":
		// Check if a transition is allowed
		err = evaluator.ValidateTransition(req.FromID, req.ToID, req.EdgeType, getUserFromRequest(r))
		if err != nil {
			respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
				"allowed": false,
				"error":   err.Error(),
			})
			return
		}
		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"allowed": true,
		})

	case "create_policy":
		// Create a new policy node
		policyNode, err := evaluator.CreatePolicyNode(
			req.Name,
			req.Description,
			req.Type,
			req.Parameters,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create policy: %v", err), http.StatusInternalServerError)
			return
		}
		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"policy_id": policyNode.ID,
			"message":   "Policy created",
		})

	case "create_check":
		// Create a check node
		checkNode, err := evaluator.CreateCheckNode(
			req.CheckID,
			req.Name,
			req.Type,
			req.Parameters,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create check: %v", err), http.StatusInternalServerError)
			return
		}
		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"check_id": checkNode.ID,
			"message":  "Check created",
		})

	case "update_check":
		// Update check status
		err = evaluator.UpdateCheckStatus(req.CheckID, req.Status, req.Results)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update check: %v", err), http.StatusInternalServerError)
			return
		}
		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"message": "Check updated",
		})

	case "satisfy":
		// Mark a check as satisfying a policy
		err = evaluator.SatisfyPolicy(req.CheckID, req.PolicyID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to satisfy policy: %v", err), http.StatusInternalServerError)
			return
		}
		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"message": "Policy satisfied",
		})

	default:
		http.Error(w, fmt.Sprintf("Unknown operation: %s", req.Operation), http.StatusBadRequest)
	}
}

// ListPolicies returns all policies
func ListPolicies(w http.ResponseWriter, r *http.Request) {
	// Collect all policy nodes from the global graph
	policies := []interface{}{}
	for _, node := range GlobalGraph.Graph.Nodes {
		if node.Kind == "policy" {
			policies = append(policies, node)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policies)
}

// GetPolicy returns a policy by ID
func GetPolicy(w http.ResponseWriter, r *http.Request) {
	policyID := r.URL.Query().Get("policy_id")
	if policyID == "" {
		// Try chi param if available
		if param := r.Context().Value("policy_id"); param != nil {
			policyID, _ = param.(string)
		}
	}
	if policyID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "policy_id is required"})
		return
	}
	policy, ok := GlobalGraph.Graph.Nodes[policyID]
	if !ok || policy.Kind != "policy" {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "policy not found"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policy)
}

// Helper function to get the evaluator for an environment
func getPolicyEvaluator(env string) *policies.PolicyEvaluator {
	// Get the graph store from the global registry
	graphStore := getGraphStore()

	// Create the evaluator with the simplified constructor (no policy registry)
	evaluator := policies.NewPolicyEvaluator(graphStore, env)

	// Set the event service if available
	if PolicyEventService != nil {
		evaluator.SetEventService(PolicyEventService)
		logger.Printf("Using policy event service for environment %s", env)
	}

	return evaluator
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

// Helper function for JSON responses
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
