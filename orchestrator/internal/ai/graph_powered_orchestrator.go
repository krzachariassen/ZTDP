package ai

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ztdp/orchestrator/internal/graph"
)

// GraphPoweredAIOrchestrator uses the graph as AI's memory and knowledge base
// This is truly AI-native - the AI explores and learns from the graph dynamically
type GraphPoweredAIOrchestrator struct {
	aiProvider AIProvider
	graph      graph.Graph
	logger     Logger
}

// NewGraphPoweredAIOrchestrator creates a new graph-powered AI orchestrator
func NewGraphPoweredAIOrchestrator(provider AIProvider, graph graph.Graph, logger Logger) *GraphPoweredAIOrchestrator {
	return &GraphPoweredAIOrchestrator{
		aiProvider: provider,
		graph:      graph,
		logger:     logger,
	}
}

// ProcessRequest processes a user request with optimized graph exploration (reduced API calls)
func (ai *GraphPoweredAIOrchestrator) ProcessRequest(ctx context.Context, userInput, userID string) (*ConversationalResponse, error) {
	if ai.logger != nil {
		ai.logger.Info("Processing graph-powered AI request", "input", userInput, "user", userID)
	}

	// Step 1: Combined graph exploration and analysis (single API call)
	analysis, err := ai.exploreAndAnalyze(ctx, userInput, userID)
	if err != nil {
		return nil, fmt.Errorf("graph exploration and analysis failed: %w", err)
	}

	// Step 2: Combined clarification assessment and response generation (single API call)
	response, err := ai.generateOptimizedResponse(ctx, userInput, userID, analysis)
	if err != nil {
		return nil, fmt.Errorf("response generation failed: %w", err)
	}

	// Step 3: Store insights back to graph asynchronously to avoid blocking
	go ai.storeInsightsToGraph(context.Background(), userID, userInput, analysis, response)

	return response, nil
}

// exploreAndAnalyze combines graph exploration and analysis into a single optimized API call
func (ai *GraphPoweredAIOrchestrator) exploreAndAnalyze(ctx context.Context, userInput, userID string) (string, error) {
	systemPrompt := `You are an AI orchestrator with access to a graph database containing:
- Agents and their capabilities
- Past workflows and executions  
- User interaction history
- System relationships and dependencies

Your task is to BOTH explore the graph AND analyze the user request in a single response.

PART 1 - GRAPH EXPLORATION:
Generate specific graph queries to discover relevant information, then execute them.

PART 2 - REQUEST ANALYSIS:
Based on the graph results, analyze:
- Intent: What does the user want to accomplish?
- Category: What domain/area (deployment, security, monitoring, etc.)?
- Context: What relevant history or relationships exist?
- Complexity: Is this a simple or complex multi-step request?
- Confidence: How confident are you in understanding the request?

Respond in this format:
GRAPH_QUERIES:
[specific queries to run]

ANALYSIS:
Intent: [clear intent]
Category: [domain area]
Context: [relevant context from graph]
Complexity: [simple/moderate/complex]
Confidence: [0-100]%
Reasoning: [why this analysis]`

	userPrompt := fmt.Sprintf(`User ID: %s
Request: %s

Explore the graph to understand context, then provide complete analysis.`, userID, userInput)

	response, err := ai.aiProvider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", fmt.Errorf("AI call failed: %w", err)
	}

	// Execute any graph queries mentioned in the response
	if strings.Contains(response, "GRAPH_QUERIES:") {
		parts := strings.Split(response, "GRAPH_QUERIES:")
		if len(parts) > 1 {
			queriesSection := strings.Split(parts[1], "ANALYSIS:")[0]
			ai.executeGraphQueries(ctx, queriesSection)
		}
	}

	return response, nil
}

// generateOptimizedResponse combines clarification assessment and response generation
func (ai *GraphPoweredAIOrchestrator) generateOptimizedResponse(ctx context.Context, userInput, userID string, analysis string) (*ConversationalResponse, error) {
	systemPrompt := `You are an AI orchestrator that decides whether to ask for clarification or execute a request.

Based on the provided analysis, you must:

1. ASSESS if you need clarification (confidence < 80% OR complex multi-step request)
2. IF clarification needed: Generate a helpful clarification question
3. IF ready to execute: Provide comprehensive execution plan with agent coordination

Your analysis includes graph context, so use that knowledge to make informed decisions.

Respond in this format:
DECISION: [CLARIFY|EXECUTE]
CONFIDENCE: [0-100]%
REASONING: [why this decision]

[If CLARIFY]:
CLARIFICATION: [specific question to ask]

[If EXECUTE]:
EXECUTION_PLAN:
- Step 1: [action using specific agent]
- Step 2: [action using specific agent]
- etc.

AGENT_COORDINATION:
- Primary Agent: [agent name and why]
- Supporting Agents: [list and roles]
- Workflow Dependencies: [any sequencing needed]`

	userPrompt := fmt.Sprintf(`User ID: %s
Original Request: %s

ANALYSIS:
%s

Based on this analysis, decide whether to clarify or execute.`, userID, userInput, analysis)

	response, err := ai.aiProvider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI call failed: %w", err)
	}

	// Parse the response to determine if clarification is needed
	if strings.Contains(response, "DECISION: CLARIFY") {
		clarificationText := ai.extractSection(response, "CLARIFICATION:")
		return &ConversationalResponse{
			Message:    clarificationText,
			Intent:     "clarification",
			Confidence: 0.5, // Lower confidence when clarification needed
			Actions: []Action{{
				Type:        "clarification_needed",
				Description: "More information required",
			}},
			Context: map[string]interface{}{
				"analysis":            analysis,
				"user_id":             userID,
				"needs_clarification": true,
			},
		}, nil
	}

	// Extract execution plan
	executionPlan := ai.extractSection(response, "EXECUTION_PLAN:")
	agentCoordination := ai.extractSection(response, "AGENT_COORDINATION:")

	// Parse confidence
	confidence := 0.85 // Default high confidence for execution
	if confidenceStr := ai.extractSection(response, "CONFIDENCE:"); confidenceStr != "" {
		if parsed := ai.parseConfidence(confidenceStr); parsed > 0 {
			confidence = float64(parsed) / 100.0
		}
	}

	return &ConversationalResponse{
		Message:    fmt.Sprintf("I'll help you with that. Here's my plan:\n\n%s\n\nAgent Coordination:\n%s", executionPlan, agentCoordination),
		Intent:     ai.extractIntent(analysis),
		Confidence: confidence,
		Actions:    ai.parseExecutionActions(executionPlan),
		Context: map[string]interface{}{
			"analysis":           analysis,
			"execution_plan":     executionPlan,
			"agent_coordination": agentCoordination,
			"user_id":            userID,
		},
	}, nil
}

// executeGraphQueries executes AI-generated graph queries (used by exploreAndAnalyze)
func (ai *GraphPoweredAIOrchestrator) executeGraphQueries(ctx context.Context, queries string) []string {
	var results []string

	queryLines := strings.Split(queries, "\n")
	for _, query := range queryLines {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}

		result := ai.executeGraphQuery(ctx, query)
		results = append(results, fmt.Sprintf("Query: %s\nResult: %s", query, result))
	}

	return results
}

// executeGraphQuery translates natural language query to graph operation
func (ai *GraphPoweredAIOrchestrator) executeGraphQuery(ctx context.Context, query string) string {
	query = strings.ToLower(query)

	// Agent queries
	if strings.Contains(query, "find all agents") || strings.Contains(query, "agents with capability") {
		agents, err := ai.graph.QueryNodes(ctx, "agent", map[string]interface{}{})
		if err != nil {
			return fmt.Sprintf("Error querying agents: %v", err)
		}
		return ai.formatAgentResults(agents)
	}

	// Workflow queries
	if strings.Contains(query, "find workflows") || strings.Contains(query, "workflows related") {
		workflows, err := ai.graph.QueryNodes(ctx, "workflow", map[string]interface{}{})
		if err != nil {
			return fmt.Sprintf("Error querying workflows: %v", err)
		}
		return ai.formatWorkflowResults(workflows)
	}

	// User history queries
	if strings.Contains(query, "user history") || strings.Contains(query, "past interactions") {
		history, err := ai.graph.QueryNodes(ctx, "conversation", map[string]interface{}{})
		if err != nil {
			return fmt.Sprintf("Error querying user history: %v", err)
		}
		return ai.formatHistoryResults(history)
	}

	// Generic exploration
	stats := ai.graph.GetStats()
	return fmt.Sprintf("Graph stats: %v", stats)
}

// storeInsightsToGraph stores learnings and insights back to graph
func (ai *GraphPoweredAIOrchestrator) storeInsightsToGraph(ctx context.Context, userID, userInput, analysis string, response *ConversationalResponse) {
	insightData := map[string]interface{}{
		"user_id":         userID,
		"input":           userInput,
		"analysis":        analysis,
		"response":        response.Message,
		"confidence":      response.Confidence,
		"timestamp":       "now",
		"learned_pattern": ai.extractLearningPattern(userInput, analysis, response),
	}

	if err := ai.graph.AddNode(ctx, "insight", fmt.Sprintf("insight_%s_%d", userID, len(userInput)), insightData); err != nil {
		if ai.logger != nil {
			ai.logger.Error("Failed to store insights to graph", err)
		}
	}
}

// extractLearningPattern identifies patterns for future learning
func (ai *GraphPoweredAIOrchestrator) extractLearningPattern(userInput, analysis string, response *ConversationalResponse) string {
	// This could also be AI-powered pattern extraction
	return fmt.Sprintf("User requested: %s, Analysis showed: %s, Response type: %s",
		userInput, analysis, response.Answer)
}

// Helper methods for response parsing
func (ai *GraphPoweredAIOrchestrator) extractSection(text, marker string) string {
	parts := strings.Split(text, marker)
	if len(parts) < 2 {
		return ""
	}

	section := parts[1]
	// Find the end of this section (next marker or end of text)
	nextMarkers := []string{"DECISION:", "CONFIDENCE:", "REASONING:", "CLARIFICATION:", "EXECUTION_PLAN:", "AGENT_COORDINATION:"}
	for _, nextMarker := range nextMarkers {
		if idx := strings.Index(section, nextMarker); idx > 0 {
			section = section[:idx]
		}
	}

	return strings.TrimSpace(section)
}

func (ai *GraphPoweredAIOrchestrator) parseConfidence(confidenceStr string) int {
	// Extract number from strings like "85%" or "85"
	re := regexp.MustCompile(`(\d+)`)
	matches := re.FindStringSubmatch(confidenceStr)
	if len(matches) > 1 {
		if val, err := strconv.Atoi(matches[1]); err == nil {
			return val
		}
	}
	return 0
}

func (ai *GraphPoweredAIOrchestrator) extractIntent(analysis string) string {
	intent := ai.extractSection(analysis, "Intent:")
	if intent == "" {
		return "general_assistance"
	}
	return strings.ToLower(strings.ReplaceAll(intent, " ", "_"))
}

func (ai *GraphPoweredAIOrchestrator) parseExecutionActions(executionPlan string) []Action {
	var actions []Action
	lines := strings.Split(executionPlan, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- Step") || strings.HasPrefix(line, "Step") {
			actions = append(actions, Action{
				Type:        "execution_step",
				Description: line,
			})
		}
	}
	if len(actions) == 0 {
		return []Action{{
			Type:        "execute_plan",
			Description: "Execute the generated plan",
		}}
	}
	return actions
}

// Helper methods for formatting graph results
func (ai *GraphPoweredAIOrchestrator) formatAgentResults(agents []map[string]interface{}) string {
	if len(agents) == 0 {
		return "No agents found"
	}

	var result strings.Builder
	for i, agent := range agents {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString(fmt.Sprintf("Agent %v", agent))
	}
	return result.String()
}

func (ai *GraphPoweredAIOrchestrator) formatWorkflowResults(workflows []map[string]interface{}) string {
	if len(workflows) == 0 {
		return "No workflows found"
	}

	var result strings.Builder
	for i, workflow := range workflows {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString(fmt.Sprintf("Workflow %v", workflow))
	}
	return result.String()
}

func (ai *GraphPoweredAIOrchestrator) formatHistoryResults(history []map[string]interface{}) string {
	if len(history) == 0 {
		return "No conversation history found"
	}

	var result strings.Builder
	for i, conv := range history {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString(fmt.Sprintf("Conversation %v", conv))
	}
	return result.String()
}
