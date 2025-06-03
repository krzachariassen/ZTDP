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
	"github.com/krzachariassen/ZTDP/internal/prompts"
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
type OpenAIProvider struct {
	config            *OpenAIConfig
	client            *http.Client
	logger            *logging.Logger
	planningPrompts   *PlannerPrompts
	policyPrompts     *PolicyPrompts
	deploymentPrompts *prompts.DeploymentPrompts
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
		logger:            logging.GetLogger().ForComponent("ai-openai"),
		planningPrompts:   NewPlannerPrompts(),
		policyPrompts:     NewPolicyPrompts(),
		deploymentPrompts: prompts.NewDeploymentPrompts(),
	}, nil
}

// GeneratePlan creates an intelligent deployment plan using OpenAI
func (p *OpenAIProvider) GeneratePlan(ctx context.Context, request *PlanningRequest) (*PlanningResponse, error) {
	p.logger.Info("🧠 Generating AI deployment plan for application: %s", request.ApplicationID)

	// Create the system prompt for AI planning
	systemPrompt := p.deploymentPrompts.BuildPlanningSystemPrompt()

	// Create the user prompt with context
	userPrompt, err := p.deploymentPrompts.BuildPlanningUserPrompt(request)
	if err != nil {
		return nil, fmt.Errorf("failed to build user prompt: %w", err)
	}

	// Call OpenAI API
	response, err := p.callOpenAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	// Parse the response into a PlanningResponse
	var planResponse PlanningResponse
	if err := json.Unmarshal([]byte(response), &planResponse); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Validate the response structure
	if planResponse.Plan == nil {
		return nil, fmt.Errorf("response missing deployment plan")
	}

	if len(planResponse.Plan.Steps) == 0 {
		return nil, fmt.Errorf("deployment plan has no steps")
	}

	// Set default confidence if not provided
	if planResponse.Confidence == 0 {
		planResponse.Confidence = 0.8
	}

	p.logger.Info("✅ AI deployment plan generated with %d steps (confidence: %.2f)",
		len(planResponse.Plan.Steps), planResponse.Confidence)

	return &planResponse, nil
}

// EvaluatePolicy uses AI to evaluate policy compliance
func (p *OpenAIProvider) EvaluatePolicy(ctx context.Context, policyContext interface{}) (*PolicyEvaluation, error) {
	p.logger.Info("🔍 Evaluating policy compliance using AI")

	// Create policy evaluation prompt
	systemPrompt := p.policyPrompts.BuildPolicySystemPrompt()
	userPrompt, err := p.policyPrompts.BuildPolicyUserPrompt(policyContext)
	if err != nil {
		return nil, fmt.Errorf("failed to build policy prompt: %w", err)
	}

	// Call OpenAI API
	response, err := p.callOpenAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("OpenAI policy evaluation failed: %w", err)
	}

	// Parse the response into PolicyEvaluation
	var evaluation PolicyEvaluation
	if err := json.Unmarshal([]byte(response), &evaluation); err != nil {
		return nil, fmt.Errorf("failed to parse policy evaluation: %w", err)
	}

	// Set default confidence if not provided
	if evaluation.Confidence == 0 {
		evaluation.Confidence = 0.8
	}

	p.logger.Info("✅ Policy evaluation completed (compliant: %t, confidence: %.2f)",
		evaluation.Compliant, evaluation.Confidence)

	return &evaluation, nil
}

// OptimizePlan refines an existing plan using AI
func (p *OpenAIProvider) OptimizePlan(ctx context.Context, plan *DeploymentPlan, context *PlanningContext) (*PlanningResponse, error) {
	p.logger.Info("⚡ Optimizing deployment plan using AI")

	// Create optimization prompt
	systemPrompt := p.deploymentPrompts.BuildOptimizationSystemPrompt()
	userPrompt, err := p.deploymentPrompts.BuildOptimizationPrompt(plan, context)
	if err != nil {
		return nil, fmt.Errorf("failed to build optimization prompt: %w", err)
	}

	// Call OpenAI API
	response, err := p.callOpenAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("OpenAI optimization failed: %w", err)
	}

	// Parse the response
	var planResponse PlanningResponse
	if err := json.Unmarshal([]byte(response), &planResponse); err != nil {
		return nil, fmt.Errorf("failed to parse optimization response: %w", err)
	}

	// Validate the response structure
	if planResponse.Plan == nil {
		return nil, fmt.Errorf("response missing deployment plan")
	}

	// Set default confidence if not provided
	if planResponse.Confidence == 0 {
		planResponse.Confidence = 0.8
	}

	p.logger.Info("✅ Plan optimization completed with %d steps", len(planResponse.Plan.Steps))

	return &planResponse, nil
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

// callOpenAI makes a request to the OpenAI API
func (p *OpenAIProvider) callOpenAI(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
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
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error (%d): %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned from OpenAI")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

// *** REVOLUTIONARY AI IMPLEMENTATIONS FOR OPENAI PROVIDER ***
// These methods implement groundbreaking AI capabilities impossible with traditional IDPs

// ChatWithPlatform implements conversational AI for platform interaction
func (p *OpenAIProvider) ChatWithPlatform(ctx context.Context, query *ConversationalQuery) (*ConversationalResponse, error) {
	p.logger.Info("💬 OpenAI processing conversational query: %s", query.Query)

	// Build conversational system prompt
	systemPrompt := p.buildConversationalSystemPrompt()
	userPrompt := p.buildConversationalUserPrompt(query)

	// Call OpenAI API
	response, err := p.callOpenAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("OpenAI conversational API failed: %w", err)
	}

	// Parse the response
	conversationalResponse, err := p.parseConversationalResponse(response, query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse conversational response: %w", err)
	}

	p.logger.Info("✅ Conversational response generated with %d insights", len(conversationalResponse.Insights))
	return conversationalResponse, nil
}

// PredictImpact implements AI-driven impact prediction
func (p *OpenAIProvider) PredictImpact(ctx context.Context, request *ImpactAnalysisRequest) (*ImpactPrediction, error) {
	p.logger.Info("🔮 OpenAI predicting impact of %d changes", len(request.Changes))

	// Build impact analysis prompts
	systemPrompt := p.buildImpactSystemPrompt()
	userPrompt := p.buildImpactUserPrompt(request)

	// Call OpenAI API
	response, err := p.callOpenAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("OpenAI impact prediction failed: %w", err)
	}

	// Parse impact prediction
	prediction, err := p.parseImpactPrediction(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse impact prediction: %w", err)
	}

	p.logger.Info("✅ Impact prediction completed: %s risk", prediction.OverallRisk)
	return prediction, nil
}

// IntelligentTroubleshooting implements AI-driven troubleshooting
func (p *OpenAIProvider) IntelligentTroubleshooting(ctx context.Context, incident *IncidentContext) (*TroubleshootingResponse, error) {
	p.logger.Info("🔍 OpenAI analyzing incident: %s", incident.IncidentID)

	// Build troubleshooting prompts
	systemPrompt := p.buildTroubleshootingSystemPrompt()
	userPrompt := p.buildTroubleshootingUserPrompt(incident)

	// Call OpenAI API
	response, err := p.callOpenAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("OpenAI troubleshooting failed: %w", err)
	}

	// Parse troubleshooting response
	troubleshootingResponse, err := p.parseTroubleshootingResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse troubleshooting response: %w", err)
	}

	p.logger.Info("✅ Troubleshooting completed: %s", troubleshootingResponse.RootCause)
	return troubleshootingResponse, nil
}

// ProactiveOptimization implements AI-driven proactive optimization
func (p *OpenAIProvider) ProactiveOptimization(ctx context.Context, scope *OptimizationScope) (*OptimizationRecommendations, error) {
	p.logger.Info("⚡ OpenAI performing proactive optimization for %s", scope.Target)

	// Build optimization prompts
	systemPrompt := p.buildProactiveOptimizationSystemPrompt()
	userPrompt := p.buildProactiveOptimizationUserPrompt(scope)

	// Call OpenAI API
	response, err := p.callOpenAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("OpenAI optimization failed: %w", err)
	}

	// Parse optimization recommendations
	recommendations, err := p.parseOptimizationRecommendations(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse optimization recommendations: %w", err)
	}

	p.logger.Info("✅ Optimization completed with %d recommendations", len(recommendations.Recommendations))
	return recommendations, nil
}

// LearningFromFailures implements AI learning from deployment outcomes
func (p *OpenAIProvider) LearningFromFailures(ctx context.Context, outcome *DeploymentOutcome) (*LearningInsights, error) {
	p.logger.Info("🧠 OpenAI learning from deployment %s", outcome.DeploymentID)

	// Build learning prompts
	systemPrompt := p.buildLearningSystemPrompt()
	userPrompt := p.buildLearningUserPrompt(outcome)

	// Call OpenAI API
	response, err := p.callOpenAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("OpenAI learning failed: %w", err)
	}

	// Parse learning insights
	insights, err := p.parseLearningInsights(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse learning insights: %w", err)
	}

	p.logger.Info("✅ Learning completed with %d insights", len(insights.Insights))
	return insights, nil
}

// *** REVOLUTIONARY AI PROMPT BUILDERS ***

// buildConversationalSystemPrompt creates system prompt for conversational AI
func (p *OpenAIProvider) buildConversationalSystemPrompt() string {
	return `You are ZTDP AI, an advanced platform intelligence assistant that can converse naturally about infrastructure, deployments, and platform operations.

REVOLUTIONARY CAPABILITIES:
- Answer complex questions about platform state and relationships
- Provide intelligent insights about infrastructure health and dependencies
- Suggest actions based on current platform conditions
- Explain deployment strategies and their implications
- Analyze system patterns and recommend optimizations

RESPONSE FORMAT:
Respond with a JSON object containing:
{
  "answer": "Natural language response to the user's query",
  "insights": ["Key insight 1", "Key insight 2", ...],
  "actions": [
    {
      "id": "action-1",
      "title": "Action Title",
      "description": "What this action does",
      "type": "deploy|scale|fix|optimize|investigate",
      "urgency": "low|medium|high|critical",
      "impact": "Expected impact",
      "command": "Optional executable command"
    }
  ],
  "confidence": 0.85,
  "follow_up": ["Follow-up question 1", "Follow-up question 2"],
  "reasoning": "Why you provided this response",
  "graph_data": { "relevant": "graph visualization data" }
}

Be conversational, insightful, and actionable. Focus on providing value that traditional deployment tools cannot offer.`
}

// buildConversationalUserPrompt creates user prompt for conversational queries
func (p *OpenAIProvider) buildConversationalUserPrompt(query *ConversationalQuery) string {
	prompt := fmt.Sprintf(`User Query: %s

Context: %s
Intent: %s
Scope: %v

Platform Context:
%s

Please provide a comprehensive, conversational response that demonstrates the power of AI-driven platform intelligence.`,
		query.Query, query.Context, query.Intent, query.Scope,
		p.formatContextForPrompt(query.Metadata))

	return prompt
}

// buildImpactSystemPrompt creates system prompt for impact prediction
func (p *OpenAIProvider) buildImpactSystemPrompt() string {
	return `You are ZTDP Impact Predictor, an AI system that can predict the consequences of deployment changes before they happen.

REVOLUTIONARY CAPABILITY:
This is impossible with traditional deployment tools - you can simulate and predict complex systemic impacts across interconnected services, environments, and infrastructure.

ANALYSIS APPROACH:
1. Analyze proposed changes and their direct effects
2. Trace impact propagation through dependencies
3. Consider timing, load, and environmental factors
4. Predict cascade effects and failure modes
5. Assess probability and severity of impacts

RESPONSE FORMAT:
{
  "overall_risk": "low|medium|high|critical",
  "affected_systems": ["system1", "system2", ...],
  "predictions": [
    {
      "system": "affected system",
      "probability": 0.75,
      "severity": "low|medium|high|critical",
      "description": "What will happen",
      "timeline": "when it happens"
    }
  ],
  "recommendations": ["mitigation strategy 1", "mitigation strategy 2"],
  "confidence": 0.85,
  "reasoning": "Detailed analysis reasoning"
}

Focus on actionable predictions that help prevent issues before they occur.`
}

// buildImpactUserPrompt creates user prompt for impact analysis
func (p *OpenAIProvider) buildImpactUserPrompt(request *ImpactAnalysisRequest) string {
	changesJson, _ := json.MarshalIndent(request.Changes, "", "  ")

	return fmt.Sprintf(`Predict the impact of these proposed changes:

Changes:
%s

Environment: %s
Scope: %s
Timeframe: %s

Context:
%s

Please predict all potential impacts, their probabilities, and provide actionable mitigation strategies.`,
		string(changesJson), request.Environment, request.Scope, request.Timeframe,
		p.formatContextForPrompt(request.Metadata))
}

// buildTroubleshootingSystemPrompt creates system prompt for intelligent troubleshooting
func (p *OpenAIProvider) buildTroubleshootingSystemPrompt() string {
	return `You are ZTDP Troubleshooter, an AI system that performs intelligent root cause analysis and provides sophisticated problem diagnosis.

REVOLUTIONARY CAPABILITY:
Unlike traditional monitoring tools that only show symptoms, you can:
- Correlate complex patterns across multiple data sources
- Identify non-obvious root causes through AI reasoning
- Suggest intelligent investigation paths
- Learn from similar historical incidents
- Provide step-by-step diagnostic procedures

ANALYSIS APPROACH:
1. Analyze symptoms and timeline to understand the problem scope
2. Correlate logs, metrics, and events to identify patterns
3. Apply AI reasoning to determine root cause
4. Generate investigation steps for confirmation
5. Provide multiple solution options with trade-offs
6. Reference similar historical incidents for context

RESPONSE FORMAT:
{
  "root_cause": "Identified root cause",
  "diagnosis": "Detailed technical diagnosis",
  "solutions": [
    {
      "id": "solution-1",
      "title": "Solution Title",
      "description": "What this solution does",
      "steps": ["step 1", "step 2"],
      "risk": "low|medium|high",
      "effort": "low|medium|high",
      "success": 0.85
    }
  ],
  "investigation": [
    {
      "step": "Check system logs",
      "command": "kubectl logs service-name",
      "expected": "Should show error messages",
      "purpose": "Confirm suspected issue"
    }
  ],
  "prevention": ["prevention measure 1", "prevention measure 2"],
  "confidence": 0.90,
  "reasoning": "Detailed reasoning process"
}

Provide intelligent, actionable troubleshooting that goes beyond basic monitoring.`
}

// buildTroubleshootingUserPrompt creates user prompt for troubleshooting
func (p *OpenAIProvider) buildTroubleshootingUserPrompt(incident *IncidentContext) string {
	incidentJson, _ := json.MarshalIndent(incident, "", "  ")

	return fmt.Sprintf(`Analyze this incident and provide intelligent troubleshooting:

Incident Context:
%s

Please provide a comprehensive root cause analysis, investigation steps, and solution recommendations based on the available data.`,
		string(incidentJson))
}

// buildProactiveOptimizationSystemPrompt creates system prompt for proactive optimization
func (p *OpenAIProvider) buildProactiveOptimizationSystemPrompt() string {
	return `You are ZTDP Optimizer, an AI system that continuously analyzes platform patterns and recommends architectural optimizations.

REVOLUTIONARY CAPABILITY:
Traditional tools only react to problems. You proactively identify optimization opportunities by:
- Analyzing usage patterns and inefficiencies
- Detecting architectural anti-patterns
- Predicting future scaling needs
- Recommending cost and performance optimizations
- Suggesting architectural improvements

ANALYSIS APPROACH:
1. Analyze current platform state and usage patterns
2. Identify inefficiencies, bottlenecks, and waste
3. Detect architectural improvements opportunities
4. Consider cost, performance, reliability, and maintainability
5. Prioritize recommendations by impact and effort
6. Provide implementation guidance

RESPONSE FORMAT:
{
  "summary": "Overall optimization assessment",
  "recommendations": [
    {
      "id": "opt-1",
      "category": "performance|cost|reliability|architecture",
      "title": "Optimization Title",
      "description": "What this optimization achieves",
      "benefits": ["benefit 1", "benefit 2"],
      "effort": "low|medium|high",
      "priority": "low|medium|high|critical",
      "actions": [{"id": "action-1", "title": "Action", "description": "What to do"}]
    }
  ],
  "patterns": [
    {
      "pattern": "Detected pattern description",
      "frequency": 5,
      "significance": "high",
      "examples": ["example 1", "example 2"]
    }
  ],
  "opportunities": [
    {
      "area": "Performance optimization",
      "potential": "20% improvement",
      "complexity": "medium",
      "roi": "high"
    }
  ],
  "reasoning": "Detailed optimization reasoning"
}

Focus on actionable, high-impact optimizations that improve platform efficiency.`
}

// buildProactiveOptimizationUserPrompt creates user prompt for optimization analysis
func (p *OpenAIProvider) buildProactiveOptimizationUserPrompt(scope *OptimizationScope) string {
	scopeJson, _ := json.MarshalIndent(scope, "", "  ")

	return fmt.Sprintf(`Analyze this scope for optimization opportunities:

Optimization Scope:
%s

Please provide comprehensive optimization recommendations focusing on the specified areas and constraints.`,
		string(scopeJson))
}

// buildLearningSystemPrompt creates system prompt for learning from deployments
func (p *OpenAIProvider) buildLearningSystemPrompt() string {
	return `You are ZTDP Learner, an AI system that learns from deployment outcomes to continuously improve future deployments.

REVOLUTIONARY CAPABILITY:
Traditional deployment tools don't learn from experience. You build institutional knowledge by:
- Analyzing successful and failed deployment patterns
- Identifying factors that contribute to success or failure
- Learning from deployment timing, strategies, and environments
- Building predictive models for future deployments
- Developing organizational best practices

LEARNING APPROACH:
1. Analyze deployment outcome and context
2. Identify success/failure factors and patterns
3. Extract actionable insights for future deployments
4. Compare with historical deployment data
5. Generate predictive insights for similar future scenarios
6. Suggest process improvements

RESPONSE FORMAT:
{
  "insights": [
    {
      "id": "insight-1",
      "type": "pattern|antipattern|success_factor|risk_factor",
      "description": "What was learned",
      "evidence": ["supporting evidence 1", "supporting evidence 2"],
      "confidence": 0.85,
      "impact": "low|medium|high"
    }
  ],
  "patterns": [
    {
      "pattern": "Learned pattern",
      "conditions": ["when pattern applies"],
      "outcomes": ["typical results"],
      "reliability": 0.90
    }
  ],
  "improvements": [
    {
      "area": "deployment process",
      "current": "current approach",
      "proposed": "improved approach",
      "benefits": ["expected benefits"]
    }
  ],
  "predictions": [
    {
      "prediction": "future prediction",
      "likelihood": 0.75,
      "timeframe": "when applies",
      "indicators": ["leading indicators"],
      "actions": ["recommended actions"]
    }
  ],
  "confidence": 0.85,
  "reasoning": "Learning analysis reasoning"
}

Focus on building actionable knowledge that improves future deployment success.`
}

// buildLearningUserPrompt creates user prompt for learning analysis
func (p *OpenAIProvider) buildLearningUserPrompt(outcome *DeploymentOutcome) string {
	outcomeJson, _ := json.MarshalIndent(outcome, "", "  ")

	return fmt.Sprintf(`Learn from this deployment outcome:

Deployment Outcome:
%s

Please analyze this deployment to extract insights, patterns, and learnings that can improve future deployments.`,
		string(outcomeJson))
}

// *** REVOLUTIONARY AI RESPONSE PARSERS ***

// parseConversationalResponse parses OpenAI response for conversational queries
func (p *OpenAIProvider) parseConversationalResponse(response string, query *ConversationalQuery) (*ConversationalResponse, error) {
	var result ConversationalResponse
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// Fallback for non-JSON responses
		result = ConversationalResponse{
			Answer:     response,
			Insights:   []string{"AI provided a natural language response"},
			Actions:    []SuggestedAction{},
			Confidence: 0.7,
			FollowUp:   []string{"Would you like me to analyze anything specific?"},
			Reasoning:  "Conversational AI response",
			Metadata:   map[string]interface{}{"query": query.Query},
		}
	}
	return &result, nil
}

// parseImpactPrediction parses OpenAI response for impact predictions
func (p *OpenAIProvider) parseImpactPrediction(response string) (*ImpactPrediction, error) {
	var result ImpactPrediction
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// Fallback for parsing errors
		result = ImpactPrediction{
			OverallRisk:     "medium",
			AffectedSystems: []string{"unknown"},
			Predictions:     []ImpactForecast{},
			Recommendations: []string{"Monitor deployment closely", "Have rollback plan ready"},
			Confidence:      0.6,
			Reasoning:       "Impact analysis completed with limited data",
			Metadata:        map[string]interface{}{"fallback": true},
		}
	}
	return &result, nil
}

// parseTroubleshootingResponse parses OpenAI response for troubleshooting
func (p *OpenAIProvider) parseTroubleshootingResponse(response string) (*TroubleshootingResponse, error) {
	var result TroubleshootingResponse
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// Fallback for parsing errors
		result = TroubleshootingResponse{
			RootCause:     "Analysis in progress",
			Diagnosis:     response,
			Solutions:     []Solution{},
			Investigation: []InvestigationStep{},
			Prevention:    []string{"Implement monitoring", "Add health checks"},
			Confidence:    0.5,
			SimilarIssues: []HistoricalIncident{},
			Reasoning:     "Troubleshooting analysis provided",
			Metadata:      map[string]interface{}{"fallback": true},
		}
	}
	return &result, nil
}

// parseOptimizationRecommendations parses OpenAI response for optimization
func (p *OpenAIProvider) parseOptimizationRecommendations(response string) (*OptimizationRecommendations, error) {
	var result OptimizationRecommendations
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// Fallback for parsing errors
		result = OptimizationRecommendations{
			Summary:         "Optimization analysis completed",
			Recommendations: []Recommendation{},
			Patterns:        []DetectedPattern{},
			Opportunities:   []OptimizationOpportunity{},
			Impact:          ImpactAssessment{},
			Priority:        []string{},
			Reasoning:       response,
			Metadata:        map[string]interface{}{"fallback": true},
		}
	}
	return &result, nil
}

// parseLearningInsights parses OpenAI response for learning insights
func (p *OpenAIProvider) parseLearningInsights(response string) (*LearningInsights, error) {
	var result LearningInsights
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// Fallback for parsing errors
		result = LearningInsights{
			Insights:      []Insight{},
			Patterns:      []LearnedPattern{},
			Improvements:  []ProcessImprovement{},
			Predictions:   []FuturePrediction{},
			Confidence:    0.6,
			Applicability: []string{"future deployments"},
			Reasoning:     response,
			Metadata:      map[string]interface{}{"fallback": true},
		}
	}
	return &result, nil
}

// formatContextForPrompt formats context for inclusion in prompts
func (p *OpenAIProvider) formatContextForPrompt(context map[string]interface{}) string {
	if context == nil {
		return "No additional context available"
	}

	contextJson, err := json.MarshalIndent(context, "", "  ")
	if err != nil {
		return "Context formatting error"
	}

	return string(contextJson)
}
