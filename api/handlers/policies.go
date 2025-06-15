package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// === AI-ENHANCED POLICY REQUEST TYPES ===

// AIEvaluatePolicyRequest represents the request for AI policy evaluation
type AIEvaluatePolicyRequest struct {
	ApplicationID string                 `json:"application_id"`
	Environment   string                 `json:"environment"`
	Action        string                 `json:"action"`
	Context       map[string]interface{} `json:"context,omitempty"`
	Timeout       int                    `json:"timeout,omitempty"`
}

// === AI-ENHANCED POLICY HANDLERS ===

// AIEvaluatePolicy godoc
// @Summary      Evaluate policies using AI
// @Description  Evaluates policies for deployments using AI-enhanced analysis
// @Tags         policies,ai
// @Accept       json
// @Produce      json
// @Param        request  body  AIEvaluatePolicyRequest  true  "AI policy evaluation request"
// @Success      200  {object}  ai.PolicyEvaluation
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /v1/ai/policies/evaluate [post]
func AIEvaluatePolicy(w http.ResponseWriter, r *http.Request) {
	var req AIEvaluatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.ApplicationID == "" {
		WriteJSONError(w, "application_id is required", http.StatusBadRequest)
		return
	}

	if req.Action == "" {
		WriteJSONError(w, "action is required", http.StatusBadRequest)
		return
	}

	// Default environment if not provided
	if req.Environment == "" {
		req.Environment = "default"
	}

	// Default timeout
	timeout := 60 * time.Second
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create AI platform agent for policy service (infrastructure layer)
	agent := GetGlobalV3Agent()
	if agent == nil {
		WriteJSONError(w, "AI service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Use event-driven PolicyAgent for evaluation (NEW: agent-to-agent architecture)
	response, err := agent.ChatWithPlatform(ctx,
		fmt.Sprintf("Evaluate policy for application %s in environment %s for action %s",
			req.ApplicationID, req.Environment, req.Action),
		fmt.Sprintf("Context: %v", req.Context))

	if err != nil {
		WriteJSONError(w, "Policy evaluation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract policy decision from agent response
	evaluation := map[string]interface{}{
		"application_id": req.ApplicationID,
		"environment":    req.Environment,
		"action":         req.Action,
		"decision":       "allowed", // Default, will be enhanced by AI response parsing
		"reasoning":      response.Message,
		"confidence":     0.8,
	}

	// Return successful evaluation result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(evaluation)
}

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
	ctx := r.Context()

	// Parse environment from query parameter or use default
	env := r.URL.Query().Get("environment")
	if env == "" {
		env = "default"
	}

	// Parse and validate request body
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid request format: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get V3Agent for policy operations
	agent := GetGlobalV3Agent()
	if agent == nil {
		WriteJSONError(w, "AI service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Execute policy operation through AI agent
	query := fmt.Sprintf("Execute policy operation in environment %s: %v", env, req)
	response, err := agent.ChatWithPlatform(ctx, query, "policy operation")
	if err != nil {
		WriteJSONError(w, "Policy operation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return AI agent response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"environment": env,
		"operation":   req,
		"response":    response,
		"timestamp":   time.Now(),
	})
}

// ListPolicies returns all policies - simplified implementation using V3Agent
func ListPolicies(w http.ResponseWriter, r *http.Request) {
	// Since we're moving to event-driven architecture, use V3Agent for policy queries
	agent := GetGlobalV3Agent()
	if agent == nil {
		WriteJSONError(w, "AI agent not initialized", http.StatusInternalServerError)
		return
	}

	// Use V3Agent to handle policy listing through chat interface
	response, err := agent.Chat(r.Context(), "List all policies available in the system")
	if err != nil {
		WriteJSONError(w, "Failed to query policies: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// For now, return the AI response as a simple structure
	result := map[string]interface{}{
		"message": response.Message,
		"note":    "Policy listing is now handled through the AI agent interface",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetPolicy returns a policy by ID - simplified implementation using V3Agent
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

	// Since we're moving to event-driven architecture, use V3Agent for policy queries
	agent := GetGlobalV3Agent()
	if agent == nil {
		WriteJSONError(w, "AI agent not initialized", http.StatusInternalServerError)
		return
	}

	// Use V3Agent to handle policy retrieval through chat interface
	query := fmt.Sprintf("Get details for policy with ID: %s", policyID)
	response, err := agent.Chat(r.Context(), query)
	if err != nil {
		WriteJSONError(w, "Failed to query policy: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// For now, return the AI response as a simple structure
	result := map[string]interface{}{
		"policy_id": policyID,
		"message":   response.Message,
		"note":      "Policy retrieval is now handled through the AI agent interface",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
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
