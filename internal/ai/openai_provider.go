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
// This is PURE INFRASTRUCTURE - only handles HTTP communication with OpenAI API
// All business logic is in AIService
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

// CallAI makes a raw AI inference call with system and user prompts
// This is pure infrastructure - only handles OpenAI API communication
func (p *OpenAIProvider) CallAI(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	p.logger.Info("🔗 Making OpenAI API call")

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
		return "", fmt.Errorf("OpenAI API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse OpenAI response
	var openAIResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &openAIResponse); err != nil {
		return "", fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	// Check for API errors
	if openAIResponse.Error != nil {
		return "", fmt.Errorf("OpenAI API error: %s", openAIResponse.Error.Message)
	}

	// Extract the response content
	if len(openAIResponse.Choices) == 0 {
		return "", fmt.Errorf("no response choices from OpenAI")
	}

	content := openAIResponse.Choices[0].Message.Content
	p.logger.Info("✅ OpenAI API call completed successfully")

	return content, nil
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
