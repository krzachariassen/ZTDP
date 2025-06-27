package domain

import (
	"time"
)

// Analysis represents the AI analysis result for a user request
type Analysis struct {
	Intent         string    `json:"intent"`
	Category       string    `json:"category"`
	Confidence     int       `json:"confidence"` // 0-100
	RequiredAgents []string  `json:"required_agents"`
	Reasoning      string    `json:"reasoning"`
	Timestamp      time.Time `json:"timestamp"`
}

// NewAnalysis creates a new analysis with validation
func NewAnalysis(intent, category string, confidence int, requiredAgents []string, reasoning string) *Analysis {
	// Validate confidence range
	if confidence > 100 {
		confidence = 100
	}
	if confidence < 0 {
		confidence = 0
	}

	return &Analysis{
		Intent:         intent,
		Category:       category,
		Confidence:     confidence,
		RequiredAgents: requiredAgents,
		Reasoning:      reasoning,
		Timestamp:      time.Now(),
	}
}

// IsHighConfidence returns true if confidence is 80% or higher
func (a *Analysis) IsHighConfidence() bool {
	return a.Confidence >= 80
}

// RequiresAgents returns true if specific agents are needed
func (a *Analysis) RequiresAgents() bool {
	return len(a.RequiredAgents) > 0
}
