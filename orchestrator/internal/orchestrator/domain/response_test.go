package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseParser_ExtractSection(t *testing.T) {
	parser := NewResponseParser()

	t.Run("should extract section from AI response", func(t *testing.T) {
		aiResponse := `DECISION: EXECUTE
CONFIDENCE: 85
REASONING: Clear request
EXECUTION_PLAN:
- Step 1: Deploy application
- Step 2: Run tests
AGENT_COORDINATION:
- Primary Agent: deploy-agent`

		plan := parser.ExtractSection(aiResponse, "EXECUTION_PLAN:")
		assert.Equal(t, "- Step 1: Deploy application\n- Step 2: Run tests", plan)

		coordination := parser.ExtractSection(aiResponse, "AGENT_COORDINATION:")
		assert.Equal(t, "- Primary Agent: deploy-agent", coordination)
	})

	t.Run("should return empty string if section not found", func(t *testing.T) {
		aiResponse := "DECISION: EXECUTE"
		result := parser.ExtractSection(aiResponse, "MISSING_SECTION:")
		assert.Empty(t, result)
	})
}

func TestResponseParser_ParseConfidence(t *testing.T) {
	parser := NewResponseParser()

	t.Run("should parse confidence from string", func(t *testing.T) {
		confidence := parser.ParseConfidence("85 percent confident")
		assert.Equal(t, 85, confidence)

		confidence = parser.ParseConfidence("90")
		assert.Equal(t, 90, confidence)
	})

	t.Run("should return 0 for invalid confidence", func(t *testing.T) {
		confidence := parser.ParseConfidence("not a number")
		assert.Equal(t, 0, confidence)
	})
}

func TestResponseParser_ExtractIntent(t *testing.T) {
	parser := NewResponseParser()

	t.Run("should extract and normalize intent", func(t *testing.T) {
		analysis := "Intent: Deploy Application\nCategory: deployment"
		intent := parser.ExtractIntent(analysis)
		assert.Equal(t, "deploy_application", intent)
	})

	t.Run("should return default intent if not found", func(t *testing.T) {
		analysis := "Category: deployment"
		intent := parser.ExtractIntent(analysis)
		assert.Equal(t, "general_assistance", intent)
	})
}

func TestResponseParser_ExtractCategory(t *testing.T) {
	parser := NewResponseParser()

	t.Run("should extract and normalize category", func(t *testing.T) {
		analysis := "Category: Deployment Operations\nIntent: deploy"
		category := parser.ExtractCategory(analysis)
		assert.Equal(t, "deployment_operations", category)
	})

	t.Run("should return default category if not found", func(t *testing.T) {
		analysis := "Intent: deploy"
		category := parser.ExtractCategory(analysis)
		assert.Equal(t, "general", category)
	})
}

func TestResponseParser_ExtractRequiredAgents(t *testing.T) {
	parser := NewResponseParser()

	t.Run("should parse comma-separated agents", func(t *testing.T) {
		analysis := "Required_Agents: deploy-agent, test-agent, monitor-agent"
		agents := parser.ExtractRequiredAgents(analysis)
		assert.Equal(t, []string{"deploy-agent", "test-agent", "monitor-agent"}, agents)
	})

	t.Run("should handle single agent", func(t *testing.T) {
		analysis := "Required_Agents: deploy-agent"
		agents := parser.ExtractRequiredAgents(analysis)
		assert.Equal(t, []string{"deploy-agent"}, agents)
	})

	t.Run("should return empty slice if no agents", func(t *testing.T) {
		analysis := "Required_Agents: none"
		agents := parser.ExtractRequiredAgents(analysis)
		assert.Empty(t, agents)
	})
}
