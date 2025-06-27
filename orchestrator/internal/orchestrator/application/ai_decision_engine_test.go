package application

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	aiInfrastructure "github.com/ztdp/orchestrator/internal/ai/infrastructure"
	"github.com/ztdp/orchestrator/internal/logging"
	"github.com/ztdp/orchestrator/internal/orchestrator/domain"
)

// setupRealAIProvider creates a real OpenAI provider for testing
func setupRealAIProvider(t *testing.T) *aiInfrastructure.OpenAIProvider {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY environment variable not set, skipping AI provider tests")
	}

	config := aiInfrastructure.DefaultOpenAIConfig()
	config.APIKey = apiKey
	config.Model = "gpt-4.1-mini" // Use faster model for tests
	config.MaxTokens = 1000       // Limit tokens for faster tests

	logger := logging.NewNoOpLogger() // Use no-op logger for tests
	provider := aiInfrastructure.NewOpenAIProvider(config, logger)

	return provider
}

func TestAIDecisionEngine_ExploreAndAnalyze(t *testing.T) {
	t.Run("should analyze user request with agent context using real AI", func(t *testing.T) {
		aiProvider := setupRealAIProvider(t)
		engine := NewAIDecisionEngine(aiProvider)

		agentContext := "Agent: deploy-agent | Status: available | Capabilities: deploy, test"
		userInput := "Deploy my application to production"
		userID := "user123"

		analysis, err := engine.ExploreAndAnalyze(context.Background(), userInput, userID, agentContext)

		assert.NoError(t, err)
		assert.NotNil(t, analysis)
		assert.NotEmpty(t, analysis.Intent)
		assert.NotEmpty(t, analysis.Category)
		assert.Greater(t, analysis.Confidence, 0)
		assert.Less(t, analysis.Confidence, 101)
		assert.NotEmpty(t, analysis.Reasoning)

		// Since we're using real AI, we can't predict exact responses
		// but we can validate the structure and reasonable expectations
		t.Logf("AI Analysis - Intent: %s, Category: %s, Confidence: %d",
			analysis.Intent, analysis.Category, analysis.Confidence)
	})
}

func TestAIDecisionEngine_MakeDecision(t *testing.T) {
	t.Run("should make decision based on analysis using real AI", func(t *testing.T) {
		aiProvider := setupRealAIProvider(t)
		engine := NewAIDecisionEngine(aiProvider)

		// Create a clear analysis that should result in execute decision
		analysis := domain.NewAnalysis("deploy_application", "deployment", 90,
			[]string{"deploy-agent"}, "Clear deployment request with specific target")

		decision, err := engine.MakeDecision(context.Background(),
			"Deploy my application to production", "user123", analysis)

		assert.NoError(t, err)
		assert.NotNil(t, decision)

		// Validate that we get either CLARIFY or EXECUTE decision
		assert.True(t, decision.Type == domain.DecisionTypeClarify ||
			decision.Type == domain.DecisionTypeExecute)

		if decision.Type == domain.DecisionTypeClarify {
			assert.NotEmpty(t, decision.ClarificationQuestion)
			t.Logf("AI Decision: CLARIFY - %s", decision.ClarificationQuestion)
		} else {
			assert.True(t, decision.IsExecutable())
			t.Logf("AI Decision: EXECUTE - Plan: %s", decision.ExecutionPlan)
		}

		assert.NotEmpty(t, decision.Reasoning)
	})

	t.Run("should handle low confidence request appropriately", func(t *testing.T) {
		aiProvider := setupRealAIProvider(t)
		engine := NewAIDecisionEngine(aiProvider)

		// Create an unclear analysis
		analysis := domain.NewAnalysis("unclear_request", "general", 30,
			[]string{}, "Request is vague and unclear")

		decision, err := engine.MakeDecision(context.Background(),
			"do something", "user123", analysis)

		assert.NoError(t, err)
		assert.NotNil(t, decision)
		assert.NotEmpty(t, decision.Reasoning)

		// With low confidence, AI might choose to clarify or still try to execute
		// We just validate the response is structured correctly
		t.Logf("AI Decision for unclear request: %s - %s",
			decision.Type, decision.Reasoning)
	})
}
