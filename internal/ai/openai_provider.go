package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/krzachariassen/ZTDP/internal/logging"
)

// OpenAIConfig contains configuration for OpenAI provider
type OpenAIConfig struct {
	APIKey      string        `json:"api_key"`
	Model       string        `json:"model"`       // e.g., "gpt-4o-mini"
	BaseURL     string        `json:"base_url"`    // OpenAI API base URL
	Timeout     time.Duration `json:"timeout"`     // Request timeout
	MaxTokens   int           `json:"max_tokens"`  // Maximum tokens for responses
	Temperature float32       `json:"temperature"` // Response creativity (0-1)
}

// DefaultOpenAIConfig returns a default configuration for OpenAI
func DefaultOpenAIConfig() *OpenAIConfig {
	// Default timeout of 90 seconds, configurable via environment
	timeout := 90 * time.Second
	if timeoutEnv := os.Getenv("ZTDP_OPENAI_TIMEOUT"); timeoutEnv != "" {
		if parsedTimeout, err := time.ParseDuration(timeoutEnv); err == nil {
			timeout = parsedTimeout
		}
	}

	return &OpenAIConfig{
		Model:       "gpt-4o-mini",
		BaseURL:     "https://api.openai.com/v1",
		Timeout:     timeout,
		MaxTokens:   4000,
		Temperature: 0.1, // Low temperature for consistent, logical planning
	}
}

// OpenAIProvider implements AIProvider using OpenAI GPT models
type OpenAIProvider struct {
	config *OpenAIConfig
	client *http.Client
	logger *logging.Logger
}

// NewOpenAIProvider creates a new OpenAI provider instance
func NewOpenAIProvider(config *OpenAIConfig, apiKey string) (*OpenAIProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	if config == nil {
		config = DefaultOpenAIConfig()
	}

	config.APIKey = apiKey

	return &OpenAIProvider{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logging.GetLogger().ForComponent("ai-openai"),
	}, nil
}

// GeneratePlan creates an intelligent deployment plan using OpenAI
func (p *OpenAIProvider) GeneratePlan(ctx context.Context, request *PlanningRequest) (*PlanningResponse, error) {
	p.logger.Info("🧠 Generating AI deployment plan for application: %s", request.ApplicationID)

	// Create the system prompt for AI planning
	systemPrompt := p.buildPlanningSystemPrompt()

	// Create the user prompt with context
	userPrompt, err := p.buildPlanningUserPrompt(request)
	if err != nil {
		return nil, fmt.Errorf("failed to build user prompt: %w", err)
	}

	// Call OpenAI API
	response, err := p.callOpenAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	// Parse the response into a PlanningResponse
	planResponse, err := p.parsePlanningResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	p.logger.Info("✅ AI deployment plan generated with %d steps (confidence: %.2f)",
		len(planResponse.Plan.Steps), planResponse.Confidence)

	return planResponse, nil
}

// EvaluatePolicy uses AI to evaluate policy compliance
func (p *OpenAIProvider) EvaluatePolicy(ctx context.Context, policyContext interface{}) (*PolicyEvaluation, error) {
	p.logger.Info("🔍 Evaluating policy compliance using AI")

	// Create policy evaluation prompt
	systemPrompt := p.buildPolicySystemPrompt()
	userPrompt, err := p.buildPolicyUserPrompt(policyContext)
	if err != nil {
		return nil, fmt.Errorf("failed to build policy prompt: %w", err)
	}

	// Call OpenAI API
	response, err := p.callOpenAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("OpenAI policy evaluation failed: %w", err)
	}

	// Parse the response into PolicyEvaluation
	evaluation, err := p.parsePolicyEvaluation(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse policy evaluation: %w", err)
	}

	p.logger.Info("✅ Policy evaluation completed (compliant: %t, confidence: %.2f)",
		evaluation.Compliant, evaluation.Confidence)

	return evaluation, nil
}

// OptimizePlan refines an existing plan using AI
func (p *OpenAIProvider) OptimizePlan(ctx context.Context, plan *DeploymentPlan, context *PlanningContext) (*PlanningResponse, error) {
	p.logger.Info("⚡ Optimizing deployment plan using AI")

	// Create optimization prompt
	systemPrompt := p.buildOptimizationSystemPrompt()
	userPrompt, err := p.buildOptimizationUserPrompt(plan, context)
	if err != nil {
		return nil, fmt.Errorf("failed to build optimization prompt: %w", err)
	}

	// Call OpenAI API
	response, err := p.callOpenAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("OpenAI optimization failed: %w", err)
	}

	// Parse the response
	planResponse, err := p.parsePlanningResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse optimization response: %w", err)
	}

	p.logger.Info("✅ Plan optimization completed with %d steps", len(planResponse.Plan.Steps))

	return planResponse, nil
}

// GetProviderInfo returns information about the OpenAI provider
func (p *OpenAIProvider) GetProviderInfo() *ProviderInfo {
	return &ProviderInfo{
		Name:    "openai-gpt",
		Version: p.config.Model,
		Capabilities: []string{
			"plan_generation",
			"policy_evaluation",
			"plan_optimization",
			"reasoning_explanation",
		},
		Metadata: map[string]interface{}{
			"max_tokens":  p.config.MaxTokens,
			"temperature": p.config.Temperature,
			"model":       p.config.Model,
		},
	}
}

// Close cleans up OpenAI provider resources
func (p *OpenAIProvider) Close() error {
	p.logger.Info("🔌 Closing OpenAI provider")
	return nil
}

// callOpenAI makes a request to the OpenAI API
func (p *OpenAIProvider) callOpenAI(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	// Build the request payload
	payload := map[string]interface{}{
		"model": p.config.Model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": userPrompt,
			},
		},
		"max_tokens":  p.config.MaxTokens,
		"temperature": p.config.Temperature,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	// Marshal the payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	// Make the request
	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error (%d): %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned from OpenAI")
	}

	return openAIResp.Choices[0].Message.Content, nil
}
