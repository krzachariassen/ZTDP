package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test-first: Define what an Analysis should contain based on current functionality
func TestAnalysis_Creation(t *testing.T) {
	t.Run("should create valid analysis", func(t *testing.T) {
		analysis := NewAnalysis("deploy_application", "deployment", 85, []string{"deploy-agent", "test-agent"}, "User wants to deploy their application")

		assert.Equal(t, "deploy_application", analysis.Intent)
		assert.Equal(t, "deployment", analysis.Category)
		assert.Equal(t, 85, analysis.Confidence)
		assert.Equal(t, []string{"deploy-agent", "test-agent"}, analysis.RequiredAgents)
		assert.Equal(t, "User wants to deploy their application", analysis.Reasoning)
		assert.NotZero(t, analysis.Timestamp)
	})

	t.Run("should validate confidence range", func(t *testing.T) {
		analysis := NewAnalysis("test", "test", 150, []string{}, "test")
		assert.Equal(t, 100, analysis.Confidence) // Should cap at 100

		analysis = NewAnalysis("test", "test", -10, []string{}, "test")
		assert.Equal(t, 0, analysis.Confidence) // Should floor at 0
	})
}

func TestAnalysis_IsHighConfidence(t *testing.T) {
	t.Run("should return true for confidence >= 80", func(t *testing.T) {
		analysis := NewAnalysis("test", "test", 85, []string{}, "test")
		assert.True(t, analysis.IsHighConfidence())
	})

	t.Run("should return false for confidence < 80", func(t *testing.T) {
		analysis := NewAnalysis("test", "test", 75, []string{}, "test")
		assert.False(t, analysis.IsHighConfidence())
	})
}

func TestAnalysis_RequiresAgents(t *testing.T) {
	t.Run("should return true when agents are required", func(t *testing.T) {
		analysis := NewAnalysis("test", "test", 85, []string{"agent-1"}, "test")
		assert.True(t, analysis.RequiresAgents())
	})

	t.Run("should return false when no agents required", func(t *testing.T) {
		analysis := NewAnalysis("test", "test", 85, []string{}, "test")
		assert.False(t, analysis.RequiresAgents())
	})
}
