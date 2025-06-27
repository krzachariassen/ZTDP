// Package ai provides backward compatibility types for the web interface
// This is a minimal package to support existing web integrations
// while the core orchestrator logic has moved to clean architecture
package ai

// ConversationalResponse represents a response from the AI orchestrator for web interfaces
type ConversationalResponse struct {
	Message    string   `json:"message"`
	Intent     string   `json:"intent"`
	Confidence float64  `json:"confidence"`
	Actions    []Action `json:"actions"`
}

// Action represents an action that can be taken by the orchestrator
type Action struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
}
