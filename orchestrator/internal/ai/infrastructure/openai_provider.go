package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ztdp/orchestrator/internal/ai/domain"
	"github.com/ztdp/orchestrator/internal/logging"
)

// OpenAIConfig contains configuration for OpenAI provider
type OpenAIConfig struct {
	APIKey      string        `json:"api_key"`
	Model       string        `json:"model"`
	BaseURL     string        `json:"base_url"`
	Timeout     time.Duration `json:"timeout"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float32       `json:"temperature"`
}

// DefaultOpenAIConfig returns a default configuration for OpenAI
func DefaultOpenAIConfig() *OpenAIConfig {
	return &OpenAIConfig{
		Model:       "gpt-4",
		BaseURL:     "https://api.openai.com/v1",
		Timeout:     30 * time.Second,
		MaxTokens:   4000,
		Temperature: 0.7,
	}
}

// OpenAIProvider implements domain.AIProvider using OpenAI GPT models
// This is PURE INFRASTRUCTURE - only handles HTTP communication with OpenAI API
type OpenAIProvider struct {
	config *OpenAIConfig
	client *http.Client
	logger logging.Logger
}

// NewOpenAIProvider creates a new OpenAI provider instance
func NewOpenAIProvider(config *OpenAIConfig, logger logging.Logger) *OpenAIProvider {
	if config == nil {
		config = DefaultOpenAIConfig()
	}

	return &OpenAIProvider{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
	}
}

// CallAI makes a raw AI inference call with system and user prompts
// This is pure infrastructure - only handles OpenAI API communication
func (p *OpenAIProvider) CallAI(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if p.logger != nil {
		p.logger.Info("Making OpenAI API call", "model", p.config.Model)
	}

	// Build the request payload
	payload := map[string]interface{}{
		"model": p.config.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"max_tokens":  p.config.MaxTokens,
		"temperature": p.config.Temperature,
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

	if p.logger != nil {
		p.logger.Debug("Sending request to OpenAI", "url", req.URL.String())
	}

	// Make the request
	resp, err := p.client.Do(req)
	if err != nil {
		if p.logger != nil {
			p.logger.Error("OpenAI API request failed", err)
		}
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if p.logger != nil {
		p.logger.Debug("Received response from OpenAI", "status", resp.StatusCode)
	}

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if p.logger != nil {
			p.logger.Error("Failed to read response body", err)
		}
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
	if p.logger != nil {
		p.logger.Info("OpenAI API call completed successfully", "response_length", len(content))
	}

	return content, nil
}

// GetProviderInfo returns information about the OpenAI provider
func (p *OpenAIProvider) GetProviderInfo() *domain.ProviderInfo {
	return &domain.ProviderInfo{
		Name:    "openai",
		Model:   p.config.Model,
		Version: "1.0.0",
	}
}

// Close cleans up OpenAI provider resources
func (p *OpenAIProvider) Close() error {
	if p.logger != nil {
		p.logger.Info("Closing OpenAI provider")
	}
	return nil
}
