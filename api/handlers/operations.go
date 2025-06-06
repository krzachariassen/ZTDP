package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/analytics"
	"github.com/krzachariassen/ZTDP/internal/operations"
)

// === AI-ENHANCED OPERATIONS REQUEST TYPES ===

// AIOptimizeRequest represents the request for proactive optimization
type AIOptimizeRequest struct {
	Target      string                 `json:"target"`
	Areas       []string               `json:"areas,omitempty"`
	Constraints map[string]interface{} `json:"constraints,omitempty"`
	Goals       []string               `json:"goals,omitempty"`
	Timeout     int                    `json:"timeout,omitempty"`
}

// AILearnRequest represents the request for learning from deployments
type AILearnRequest struct {
	DeploymentID string                 `json:"deployment_id"`
	Success      bool                   `json:"success"`
	Duration     string                 `json:"duration,omitempty"`
	Issues       []string               `json:"issues,omitempty"`
	Metrics      map[string]interface{} `json:"metrics,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
	Timeout      int                    `json:"timeout,omitempty"`
}

// === AI-ENHANCED OPERATIONS HANDLERS ===

// AIProactiveOptimize godoc
// @Summary      Proactive optimization using AI
// @Description  Revolutionary AI that continuously analyzes patterns and recommends architectural optimizations before problems occur
// @Tags         operations,ai,revolutionary
// @Accept       json
// @Produce      json
// @Param        request  body  AIOptimizeRequest  true  "Proactive optimization request"
// @Success      200  {object}  ai.OptimizationRecommendations
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /v1/ai/proactive-optimize [post]
func AIProactiveOptimize(w http.ResponseWriter, r *http.Request) {
	var req AIOptimizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Target == "" {
		WriteJSONError(w, "target is required", http.StatusBadRequest)
		return
	}

	// Default focus areas if not provided
	focus := req.Areas
	if len(focus) == 0 {
		focus = []string{"performance", "reliability", "security", "cost"}
	}

	// Default timeout
	timeout := 60 * time.Second
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create AI platform agent for operations service (infrastructure layer)
	agent := GetGlobalV3Agent()
	if agent == nil {
		WriteJSONError(w, "AI service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Use operations service for proactive optimization (clean architecture - business logic in domain service)
	operationsService := operations.NewOperationsService(GlobalGraph, agent.Provider())
	recommendations, err := operationsService.OptimizeOperations(ctx, req.Target, focus)
	if err != nil {
		WriteJSONError(w, "Proactive optimization failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recommendations)
}

// AILearnFromDeployment godoc
// @Summary      Learn from deployment outcomes using AI
// @Description  Revolutionary AI that learns from deployment outcomes to continuously improve future deployments
// @Tags         operations,ai,revolutionary
// @Accept       json
// @Produce      json
// @Param        request  body  AILearnRequest  true  "Learning request"
// @Success      200  {object}  ai.LearningInsights
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /v1/ai/learn [post]
func AILearnFromDeployment(w http.ResponseWriter, r *http.Request) {
	var req AILearnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.DeploymentID == "" {
		WriteJSONError(w, "deployment_id is required", http.StatusBadRequest)
		return
	}

	// Default timeout
	timeout := 60 * time.Second
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Parse duration if provided
	var duration int64 = 0
	if req.Duration != "" {
		if parsedDuration, err := time.ParseDuration(req.Duration); err == nil {
			duration = int64(parsedDuration.Seconds())
		}
	}

	// Convert issues to DeploymentIssue structs
	issues := make([]ai.DeploymentIssue, len(req.Issues))
	for i, issue := range req.Issues {
		issues[i] = ai.DeploymentIssue{
			Type:        "error", // Default type
			Description: issue,
			Severity:    "medium",  // Default severity
			Resolution:  "pending", // Default resolution
			Metadata: map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
			},
		}
	}

	// Create AI platform agent for analytics service (infrastructure layer)
	agent := GetGlobalV3Agent()
	if agent == nil {
		WriteJSONError(w, "AI service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Create deployment outcome for learning
	outcome := &ai.DeploymentOutcome{
		DeploymentID: req.DeploymentID,
		Success:      req.Success,
		Duration:     duration,
		Issues:       issues,
		Metadata: map[string]interface{}{
			"endpoint": "ai_learning",
		},
	}

	// Use analytics service for learning (clean architecture - business logic in domain service)
	analyticsService := analytics.NewAnalyticsService(GlobalGraph, agent.Provider())
	insights, err := analyticsService.LearnFromDeployment(ctx, outcome)
	if err != nil {
		WriteJSONError(w, "Learning from deployment failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(insights)
}
