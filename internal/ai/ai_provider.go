package ai

import (
	"context"
)

// AIProvider defines the interface for AI infrastructure providers
// This handles ONLY the communication with AI services (OpenAI, Anthropic, etc.)
// Business logic should be in domain services, not in providers
type AIProvider interface {
	// CallAI makes a raw AI inference call with system and user prompts
	// Returns the raw response from the AI provider
	CallAI(ctx context.Context, systemPrompt, userPrompt string) (string, error)

	// GetProviderInfo returns information about the AI provider
	GetProviderInfo() *ProviderInfo

	// Close cleans up provider resources
	Close() error
}

// ProviderInfo contains metadata about an AI provider
type ProviderInfo struct {
	Name         string                 `json:"name"`         // Provider name (e.g., "openai-gpt4")
	Version      string                 `json:"version"`      // Provider version
	Capabilities []string               `json:"capabilities"` // Supported capabilities
	Metadata     map[string]interface{} `json:"metadata"`     // Provider-specific metadata
}
