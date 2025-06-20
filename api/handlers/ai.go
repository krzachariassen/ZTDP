package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/krzachariassen/ZTDP/internal/ai"
)

// AIProviderInfo represents AI provider information
type AIProviderInfo struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Capabilities []string          `json:"capabilities"`
	Available    bool              `json:"available"`
	Model        string            `json:"model,omitempty"`
	Config       map[string]string `json:"config,omitempty"`
}

// AIProviderStatus godoc
// @Summary      Get AI provider status
// @Description  Returns information about the current AI provider configuration and availability
// @Tags         ai
// @Produce      json
// @Success      200  {object}  AIProviderInfo
// @Failure      500  {object}  map[string]string
// @Router       /v1/ai/provider/status [get]
func AIProviderStatus(w http.ResponseWriter, r *http.Request) {
	// Try to get global AI platform agent to check availability
	agent := GetGlobalOrchestrator()

	providerInfo := AIProviderInfo{
		Available:    false,
		Capabilities: []string{"plan_generation", "policy_evaluation", "plan_optimization"},
	}

	if agent == nil {
		providerInfo.Name = "OpenAI (Unavailable)"
		providerInfo.Config = map[string]string{
			"error": "AI agent not initialized",
		}
	} else {
		// Get provider info from AI platform agent (simplified)
		providerInfo.Name = "ZTDP AI Orchestrator"
		providerInfo.Version = "1.0.0"
		providerInfo.Available = true

		// Get model from environment or config
		if model := getEnvOrDefault("OPENAI_MODEL", "gpt-4"); model != "" {
			providerInfo.Model = model
		}

		providerInfo.Config = map[string]string{
			"base_url": getEnvOrDefault("OPENAI_BASE_URL", "https://api.openai.com/v1"),
			"model":    providerInfo.Model,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(providerInfo)
}

// AIMetrics godoc
// @Summary      Get AI performance metrics
// @Description  Returns performance metrics for AI operations
// @Tags         ai
// @Produce      json
// @Param        hours  query  int  false  "Number of hours to look back (default: 24)"
// @Success      200  {object}  map[string]interface{}
// @Router       /v1/ai/metrics [get]
func AIMetrics(w http.ResponseWriter, r *http.Request) {
	hours := 24 // Default to last 24 hours
	if h := r.URL.Query().Get("hours"); h != "" {
		if parsed, err := strconv.Atoi(h); err == nil && parsed > 0 {
			hours = parsed
		}
	}

	// For now, return placeholder metrics
	// TODO: Implement actual metrics collection from AI operations
	metrics := map[string]interface{}{
		"timeframe_hours": hours,
		"plan_generation": map[string]interface{}{
			"total_requests":    0,
			"successful":        0,
			"failed":            0,
			"avg_response_time": "0ms",
			"success_rate":      "0%",
		},
		"policy_evaluation": map[string]interface{}{
			"total_requests":    0,
			"successful":        0,
			"failed":            0,
			"avg_response_time": "0ms",
			"success_rate":      "0%",
		},
		"plan_optimization": map[string]interface{}{
			"total_requests":    0,
			"successful":        0,
			"failed":            0,
			"avg_response_time": "0ms",
			"success_rate":      "0%",
		},
		"note": "Metrics collection is not yet implemented. This endpoint returns placeholder data.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// *** REVOLUTIONARY AI API ENDPOINTS ***
// These endpoints demonstrate groundbreaking AI capabilities impossible with traditional IDPs

// AIChatRequest represents the request for conversational AI
type AIChatRequest struct {
	Query   string   `json:"query"`
	Context string   `json:"context,omitempty"`
	Scope   []string `json:"scope,omitempty"`
	Session string   `json:"session,omitempty"`
	Timeout int      `json:"timeout,omitempty"`
}

// AIChatWithPlatform godoc
// @Summary      Chat with Platform using AI
// @Description  Revolutionary conversational AI that allows natural language interaction with platform graph for insights and actions
// @Tags         ai,revolutionary
// @Accept       json
// @Produce      json
// @Param        request  body  AIChatRequest  true  "Chat request"
// @Success      200  {object}  ai.ConversationalResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /v1/ai/chat [post]
func AIChatWithPlatform(w http.ResponseWriter, r *http.Request) {
	var req AIChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		WriteJSONError(w, "query is required", http.StatusBadRequest)
		return
	}

	// Default timeout
	timeout := 60 * time.Second
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	// Use global AI platform agent with all domain services already injected
	agent := GetGlobalOrchestrator()
	if agent == nil {
		WriteJSONError(w, "AI service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Chat with platform using revolutionary AI
	response, err := agent.Chat(ctx, req.Query)
	if err != nil {
		WriteJSONError(w, "Conversational AI failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// V3ChatRequest represents a request to the V3 AI chat endpoint
type V3ChatRequest struct {
	Message string `json:"message" binding:"required"`
}

// V3AIChat godoc
// @Summary      Chat with V3 AI Platform Agent (Ultra Simple)
// @Description  Ultra-simple ChatGPT-style AI interface. AI drives everything naturally.
// @Tags         ai
// @Accept       json
// @Produce      json
// @Param        request  body      V3ChatRequest  true  "Chat request"
// @Success      200      {object}  ai.ConversationalResponse
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /v3/ai/chat [post]
func V3AIChat(w http.ResponseWriter, r *http.Request) {
	var req V3ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		WriteJSONError(w, "Message is required", http.StatusBadRequest)
		return
	}

	// Use global orchestrator
	orchestrator := GetGlobalOrchestrator()
	if orchestrator == nil {
		WriteJSONError(w, "Orchestrator not available", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	// Use the ultra simple Chat method!
	response, err := orchestrator.Chat(ctx, req.Message)
	if err != nil {
		WriteJSONError(w, "Orchestrator chat failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// createAIProvider creates an AI provider based on environment configuration
func createAIProvider() (ai.AIProvider, error) {
	providerName := os.Getenv("AI_PROVIDER")
	if providerName == "" {
		providerName = "openai"
	}

	switch providerName {
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
		}

		config := ai.DefaultOpenAIConfig()
		if model := os.Getenv("OPENAI_MODEL"); model != "" {
			config.Model = model
		}
		if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
			config.BaseURL = baseURL
		}

		return ai.NewOpenAIProvider(config, apiKey)
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", providerName)
	}
}

// Helper function to get environment variable with fallback
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
