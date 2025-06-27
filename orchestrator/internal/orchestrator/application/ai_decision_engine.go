package application

import (
	"context"
	"fmt"
	"strings"

	aiDomain "github.com/ztdp/orchestrator/internal/ai/domain"
	"github.com/ztdp/orchestrator/internal/orchestrator/domain"
)

// AIDecisionEngine handles AI-powered decision making
type AIDecisionEngine struct {
	aiProvider     aiDomain.AIProvider
	responseParser *domain.ResponseParser
}

// NewAIDecisionEngine creates a new AI decision engine
func NewAIDecisionEngine(aiProvider aiDomain.AIProvider) *AIDecisionEngine {
	return &AIDecisionEngine{
		aiProvider:     aiProvider,
		responseParser: domain.NewResponseParser(),
	}
}

// ExploreAndAnalyze analyzes user request with agent context and returns structured analysis
func (e *AIDecisionEngine) ExploreAndAnalyze(ctx context.Context, userInput, userID, agentContext string) (*domain.Analysis, error) {
	systemPrompt := `You are an AI orchestrator. You have access to the following agents and their capabilities:

AVAILABLE_AGENTS:
` + agentContext + `

Analyze the user request and determine:
- Intent: What does the user want to accomplish?
- Category: What domain/area (deployment, security, monitoring, etc.)?
- Confidence: How confident are you in understanding the request?
- Required_Agents: Which agents (if any) would be needed to fulfill this request?

Respond in this format:
ANALYSIS:
Intent: [clear intent]
Category: [domain area]
Confidence: [0-100 percent]
Required_Agents: [list specific agents needed]
Reasoning: [why this analysis]`

	userPrompt := fmt.Sprintf(`User ID: %s
Request: %s

Analyze this request based on available agents.`, userID, userInput)

	response, err := e.aiProvider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI call failed: %w", err)
	}

	// Parse the response into structured analysis
	intent := e.responseParser.ExtractIntent(response)
	category := e.responseParser.ExtractCategory(response)
	confidenceStr := e.responseParser.ExtractSection(response, "Confidence:")
	confidence := e.responseParser.ParseConfidence(confidenceStr)
	requiredAgents := e.responseParser.ExtractRequiredAgents(response)
	reasoning := e.responseParser.ExtractSection(response, "Reasoning:")

	return domain.NewAnalysis(intent, category, confidence, requiredAgents, reasoning), nil
}

// MakeDecision determines whether to clarify or execute based on analysis
func (e *AIDecisionEngine) MakeDecision(ctx context.Context, userInput, userID string, analysis *domain.Analysis) (*domain.Decision, error) {
	systemPrompt := `You are an AI orchestrator that decides whether to ask for clarification or execute a request.

Based on the provided analysis, you must:

1. ASSESS if you need clarification (confidence < 80 percent OR complex multi-step request)
2. IF clarification needed: Generate a helpful clarification question
3. IF ready to execute: Provide comprehensive execution plan with agent coordination

Your analysis includes graph context with available agents and capabilities. When generating execution plans, you MUST:
- Reference specific agents by name that were found in the graph exploration
- Mention the agent capabilities that are relevant
- Create realistic agent coordination plans

Respond in this format:
DECISION: [CLARIFY|EXECUTE]
CONFIDENCE: [0-100 percent]
REASONING: [why this decision]

[If CLARIFY]:
CLARIFICATION: [specific question to ask]

[If EXECUTE]:
EXECUTION_PLAN:
- Step 1: [action using specific agent name from graph]
- Step 2: [action using specific agent name from graph]
- etc.

AGENT_COORDINATION:
- Primary Agent: [specific agent name from analysis and why]
- Supporting Agents: [list specific agent names and roles]
- Workflow Dependencies: [any sequencing needed]`

	analysisText := fmt.Sprintf(`Intent: %s
Category: %s
Confidence: %d
Required_Agents: %s
Reasoning: %s`, analysis.Intent, analysis.Category, analysis.Confidence, strings.Join(analysis.RequiredAgents, ", "), analysis.Reasoning)

	userPrompt := fmt.Sprintf(`User ID: %s
Original Request: %s

ANALYSIS:
%s

Based on this analysis, decide whether to clarify or execute.`, userID, userInput, analysisText)

	response, err := e.aiProvider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI call failed: %w", err)
	}

	// Parse the decision
	if strings.Contains(response, "DECISION: CLARIFY") {
		clarificationQuestion := e.responseParser.ExtractSection(response, "CLARIFICATION:")
		reasoning := e.responseParser.ExtractSection(response, "REASONING:")
		return domain.NewClarifyDecision(clarificationQuestion, reasoning), nil
	}

	// Extract execution details
	executionPlan := e.responseParser.ExtractSection(response, "EXECUTION_PLAN:")
	agentCoordination := e.responseParser.ExtractSection(response, "AGENT_COORDINATION:")
	reasoning := e.responseParser.ExtractSection(response, "REASONING:")

	return domain.NewExecuteDecision(executionPlan, agentCoordination, reasoning), nil
}
