package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/krzachariassen/ZTDP/internal/logging"
)

// AICapabilities provides specialized AI-powered platform capabilities
// These are the revolutionary features that differentiate ZTDP from traditional IDPs
type AICapabilities struct {
	provider        AIProvider
	logger          *logging.Logger
	responseBuilder *ResponseBuilder
}

// NewAICapabilities creates a new AI capabilities engine
func NewAICapabilities(provider AIProvider, logger *logging.Logger) *AICapabilities {
	return &AICapabilities{
		provider:        provider,
		logger:          logger,
		responseBuilder: NewResponseBuilder(logger),
	}
}

// *** REVOLUTIONARY AI CAPABILITIES ***

// performPlatformAnalysis provides AI-driven platform analysis
func (capabilities *AICapabilities) PerformPlatformAnalysis(
	ctx context.Context,
	intent *Intent,
	query string,
	platformContext *PlatformContext,
) (map[string]interface{}, error) {
	capabilities.logger.Info("📊 Performing AI-powered platform analysis")

	// Build analysis system prompt
	systemPrompt := capabilities.buildAnalysisSystemPrompt(platformContext)
	userPrompt := capabilities.buildAnalysisUserPrompt(query, intent, platformContext)

	// Perform analysis using AI
	rawResponse, err := capabilities.provider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI platform analysis failed: %w", err)
	}

	// Parse analysis results
	analysis, err := capabilities.parseAnalysisResponse(rawResponse)
	if err != nil {
		capabilities.logger.Warn("Failed to parse AI analysis, using fallback: %v", err)
		return capabilities.fallbackAnalysis(platformContext), nil
	}

	return analysis, nil
}

// performIntelligentTroubleshooting provides AI-driven troubleshooting
func (capabilities *AICapabilities) PerformIntelligentTroubleshooting(
	ctx context.Context,
	intent *Intent,
	query string,
) (*TroubleshootingResponse, error) {
	capabilities.logger.Info("🔍 Performing AI-powered troubleshooting analysis")

	// Build troubleshooting prompts
	systemPrompt := capabilities.buildTroubleshootingSystemPrompt()
	userPrompt := capabilities.buildTroubleshootingUserPrompt(query, intent)

	// Perform troubleshooting using AI
	rawResponse, err := capabilities.provider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI troubleshooting failed: %w", err)
	}

	// Parse troubleshooting results
	analysisData, err := capabilities.parseTroubleshootingResponse(rawResponse)
	if err != nil {
		capabilities.logger.Warn("Failed to parse AI troubleshooting, using fallback: %v", err)
		analysisData = capabilities.fallbackTroubleshooting(query)
	}

	// Build structured response
	response := capabilities.responseBuilder.BuildTroubleshootingResponse(analysisData)
	return response, nil
}

// PredictDeploymentImpact analyzes potential impact before deployment
func (capabilities *AICapabilities) PredictDeploymentImpact(
	ctx context.Context,
	changes []ProposedChange,
	environment string,
	platformContext *PlatformContext,
) (*ImpactPrediction, error) {
	capabilities.logger.Info("🔮 Predicting deployment impact for %d changes in %s", len(changes), environment)

	// Build impact prediction prompts
	systemPrompt := capabilities.buildImpactPredictionSystemPrompt(platformContext)
	userPrompt := capabilities.buildImpactPredictionUserPrompt(changes, environment, platformContext)

	// Predict impact using AI
	rawResponse, err := capabilities.provider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI impact prediction failed: %w", err)
	}

	// Parse prediction results
	predictionData, err := capabilities.parseImpactPredictionResponse(rawResponse)
	if err != nil {
		capabilities.logger.Warn("Failed to parse AI prediction, using fallback: %v", err)
		predictionData = capabilities.fallbackImpactPrediction(changes, environment)
	}

	// Build structured response
	response := capabilities.responseBuilder.BuildImpactPrediction(predictionData)
	return response, nil
}

// ProactiveOptimization continuously analyzes platform for improvements
func (capabilities *AICapabilities) ProactiveOptimization(
	ctx context.Context,
	target string,
	focus []string,
	platformContext *PlatformContext,
) (*OptimizationRecommendations, error) {
	capabilities.logger.Info("⚡ Performing proactive optimization for %s", target)

	// Build optimization prompts
	systemPrompt := capabilities.buildOptimizationSystemPrompt(platformContext)
	userPrompt := capabilities.buildOptimizationUserPrompt(target, focus, platformContext)

	// Generate optimization recommendations using AI
	rawResponse, err := capabilities.provider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI optimization failed: %w", err)
	}

	// Parse optimization results
	optimizationData, err := capabilities.parseOptimizationResponse(rawResponse)
	if err != nil {
		capabilities.logger.Warn("Failed to parse AI optimization, using fallback: %v", err)
		optimizationData = capabilities.fallbackOptimization(target, focus)
	}

	// Build structured response
	response := capabilities.responseBuilder.BuildOptimizationRecommendations(optimizationData)
	return response, nil
}

// LearnFromDeployment captures deployment outcomes for continuous learning
func (capabilities *AICapabilities) LearnFromDeployment(
	ctx context.Context,
	deploymentID string,
	success bool,
	duration int64,
	issues []DeploymentIssue,
) (*LearningInsights, error) {
	capabilities.logger.Info("🧠 Learning from deployment %s (success: %t, duration: %ds)", deploymentID, success, duration)

	// Build learning prompts
	systemPrompt := capabilities.buildLearningSystemPrompt()
	userPrompt := capabilities.buildLearningUserPrompt(deploymentID, success, duration, issues)

	// Generate learning insights using AI
	rawResponse, err := capabilities.provider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI learning failed: %w", err)
	}

	// Parse learning results
	learningData, err := capabilities.parseLearningResponse(rawResponse)
	if err != nil {
		capabilities.logger.Warn("Failed to parse AI learning, using fallback: %v", err)
		learningData = capabilities.fallbackLearning(success, duration, issues)
	}

	// Build structured response
	response := capabilities.responseBuilder.BuildLearningInsights(learningData)
	return response, nil
}

// *** SYSTEM PROMPT BUILDERS ***

func (capabilities *AICapabilities) buildAnalysisSystemPrompt(context *PlatformContext) string {
	return fmt.Sprintf(`You are an AI platform analyst for ZTDP infrastructure platform.

PLATFORM STATE:
- Applications: %d
- Services: %d  
- Policies: %d
- Health: %s

ANALYSIS GOALS:
1. Provide actionable insights about platform state
2. Identify potential issues or optimization opportunities
3. Suggest improvements and next steps
4. Highlight important trends or patterns

RESPONSE FORMAT (JSON):
{
  "insights": ["insight1", "insight2"],
  "health_summary": "overall health status",
  "recommendations": ["rec1", "rec2"],
  "issues": ["issue1", "issue2"],
  "trends": ["trend1", "trend2"],
  "next_steps": ["step1", "step2"]
}`,
		len(context.Applications),
		len(context.Services),
		len(context.Policies),
		context.Health["status"])
}

func (capabilities *AICapabilities) buildTroubleshootingSystemPrompt() string {
	return `You are an AI troubleshooting expert for infrastructure platforms.

TROUBLESHOOTING APPROACH:
1. Analyze symptoms systematically
2. Identify most likely root causes
3. Provide step-by-step diagnosis
4. Suggest remediation actions
5. Estimate resolution time

RESPONSE FORMAT (JSON):
{
  "root_cause": "most likely cause",
  "confidence": 0.0-1.0,
  "symptoms": ["symptom1", "symptom2"],
  "diagnosis": "detailed analysis",
  "recommendations": ["action1", "action2"],
  "next_steps": ["step1", "step2"],
  "estimated_time": "time estimate",
  "severity": "Low|Medium|High|Critical"
}`
}

func (capabilities *AICapabilities) buildImpactPredictionSystemPrompt(context *PlatformContext) string {
	return `You are an AI impact prediction system for deployment changes.

PREDICTION SCOPE:
1. Analyze proposed changes for potential risks
2. Predict system impact and downtime
3. Identify affected components
4. Assess rollback complexity
5. Recommend monitoring points

RESPONSE FORMAT (JSON):
{
  "overall_risk": "Low|Medium|High|Critical",
  "confidence": 0.0-1.0,
  "affected_systems": ["system1", "system2"],
  "risk_factors": ["factor1", "factor2"],
  "recommendations": ["rec1", "rec2"],
  "estimated_downtime": "time estimate",
  "rollback_plan": "rollback strategy",
  "monitoring_points": ["metric1", "metric2"]
}`
}

func (capabilities *AICapabilities) buildOptimizationSystemPrompt(context *PlatformContext) string {
	return `You are an AI optimization advisor for infrastructure platforms.

OPTIMIZATION FOCUS:
1. Identify performance improvements
2. Suggest cost optimizations  
3. Recommend architectural enhancements
4. Highlight efficiency gains
5. Assess implementation effort

RESPONSE FORMAT (JSON):
{
  "recommendations": [
    {
      "title": "optimization title",
      "description": "detailed description",
      "impact": "Low|Medium|High",
      "effort": "Low|Medium|High", 
      "priority": "Low|Medium|High",
      "steps": ["step1", "step2"]
    }
  ],
  "patterns": ["pattern1", "pattern2"],
  "confidence": 0.0-1.0,
  "estimated_impact": "impact description"
}`
}

func (capabilities *AICapabilities) buildLearningSystemPrompt() string {
	return `You are an AI learning system that analyzes deployment outcomes.

LEARNING OBJECTIVES:
1. Extract insights from deployment results
2. Identify patterns and trends
3. Generate actionable recommendations
4. Build institutional knowledge
5. Improve future deployments

RESPONSE FORMAT (JSON):
{
  "insights": ["insight1", "insight2"],
  "patterns": ["pattern1", "pattern2"],
  "confidence": 0.0-1.0,
  "actionable": true/false,
  "impact": "Low|Medium|High",
  "categories": ["category1", "category2"],
  "trends": ["trend1", "trend2"],
  "predictions": ["prediction1", "prediction2"]
}`
}

// *** USER PROMPT BUILDERS ***

func (capabilities *AICapabilities) buildAnalysisUserPrompt(query string, intent *Intent, context *PlatformContext) string {
	prompt := fmt.Sprintf("User Query: %s\n", query)
	prompt += fmt.Sprintf("Intent: %s\n", intent.Type)

	if len(context.Applications) > 0 {
		prompt += fmt.Sprintf("Platform has %d applications\n", len(context.Applications))
	}

	if len(context.RecentEvents) > 0 {
		prompt += "Recent activity detected\n"
	}

	prompt += "\nAnalyze the platform state and provide insights."
	return prompt
}

func (capabilities *AICapabilities) buildTroubleshootingUserPrompt(query string, intent *Intent) string {
	prompt := fmt.Sprintf("Problem Description: %s\n", query)

	if symptoms, ok := intent.Parameters["symptoms"].([]string); ok {
		prompt += fmt.Sprintf("Symptoms: %s\n", strings.Join(symptoms, ", "))
	}

	prompt += "\nPerform troubleshooting analysis and provide recommendations."
	return prompt
}

func (capabilities *AICapabilities) buildImpactPredictionUserPrompt(
	changes []ProposedChange,
	environment string,
	context *PlatformContext,
) string {
	prompt := fmt.Sprintf("Environment: %s\n", environment)
	prompt += fmt.Sprintf("Number of changes: %d\n", len(changes))

	for i, change := range changes {
		prompt += fmt.Sprintf("Change %d: %s -> %s\n", i+1, change.Type, change.Description)
	}

	prompt += "\nPredict the impact of these changes."
	return prompt
}

func (capabilities *AICapabilities) buildOptimizationUserPrompt(
	target string,
	focus []string,
	context *PlatformContext,
) string {
	prompt := fmt.Sprintf("Optimization Target: %s\n", target)

	if len(focus) > 0 {
		prompt += fmt.Sprintf("Focus Areas: %s\n", strings.Join(focus, ", "))
	}

	prompt += fmt.Sprintf("Platform Scale: %d applications, %d services\n",
		len(context.Applications), len(context.Services))

	prompt += "\nGenerate optimization recommendations."
	return prompt
}

func (capabilities *AICapabilities) buildLearningUserPrompt(
	deploymentID string,
	success bool,
	duration int64,
	issues []DeploymentIssue,
) string {
	prompt := fmt.Sprintf("Deployment ID: %s\n", deploymentID)
	prompt += fmt.Sprintf("Success: %t\n", success)
	prompt += fmt.Sprintf("Duration: %d seconds\n", duration)

	if len(issues) > 0 {
		prompt += "Issues encountered:\n"
		for _, issue := range issues {
			prompt += fmt.Sprintf("- %s: %s\n", issue.Type, issue.Description)
		}
	}

	prompt += "\nAnalyze this deployment outcome and extract learning insights."
	return prompt
}

// *** RESPONSE PARSERS ***

func (capabilities *AICapabilities) parseAnalysisResponse(rawResponse string) (map[string]interface{}, error) {
	return capabilities.responseBuilder.ParseJSONResponse(capabilities.extractJSON(rawResponse))
}

func (capabilities *AICapabilities) parseTroubleshootingResponse(rawResponse string) (map[string]interface{}, error) {
	return capabilities.responseBuilder.ParseJSONResponse(capabilities.extractJSON(rawResponse))
}

func (capabilities *AICapabilities) parseImpactPredictionResponse(rawResponse string) (map[string]interface{}, error) {
	return capabilities.responseBuilder.ParseJSONResponse(capabilities.extractJSON(rawResponse))
}

func (capabilities *AICapabilities) parseOptimizationResponse(rawResponse string) (map[string]interface{}, error) {
	return capabilities.responseBuilder.ParseJSONResponse(capabilities.extractJSON(rawResponse))
}

func (capabilities *AICapabilities) parseLearningResponse(rawResponse string) (map[string]interface{}, error) {
	return capabilities.responseBuilder.ParseJSONResponse(capabilities.extractJSON(rawResponse))
}

// extractJSON extracts JSON from AI response
func (capabilities *AICapabilities) extractJSON(response string) string {
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}") + 1

	if jsonStart == -1 || jsonEnd <= jsonStart {
		// No JSON found, return empty object
		return "{}"
	}

	return response[jsonStart:jsonEnd]
}

// *** FALLBACK METHODS ***

func (capabilities *AICapabilities) fallbackAnalysis(context *PlatformContext) map[string]interface{} {
	return map[string]interface{}{
		"insights":        []string{"Platform analysis completed", "System appears operational"},
		"health_summary":  context.Health["status"],
		"recommendations": []string{"Continue monitoring", "Review recent changes"},
		"issues":          []string{},
		"trends":          []string{},
		"next_steps":      []string{"Monitor system health", "Review metrics"},
	}
}

func (capabilities *AICapabilities) fallbackTroubleshooting(query string) map[string]interface{} {
	return map[string]interface{}{
		"root_cause":      "Investigation required",
		"confidence":      0.3,
		"symptoms":        []string{query},
		"diagnosis":       "Further analysis needed",
		"recommendations": []string{"Check logs", "Monitor metrics", "Review recent changes"},
		"next_steps":      []string{"Gather more information", "Check system status"},
		"estimated_time":  "30-60 minutes",
		"severity":        "Medium",
	}
}

func (capabilities *AICapabilities) fallbackImpactPrediction(changes []ProposedChange, environment string) map[string]interface{} {
	return map[string]interface{}{
		"overall_risk":       "Medium",
		"confidence":         0.5,
		"affected_systems":   []string{environment},
		"risk_factors":       []string{"Multiple changes", "Production environment"},
		"recommendations":    []string{"Deploy during maintenance window", "Monitor closely"},
		"estimated_downtime": "5-10 minutes",
		"rollback_plan":      "Standard rollback procedures",
		"monitoring_points":  []string{"Response time", "Error rate", "System health"},
	}
}

func (capabilities *AICapabilities) fallbackOptimization(target string, focus []string) map[string]interface{} {
	return map[string]interface{}{
		"recommendations": []map[string]interface{}{
			{
				"title":       "Performance Review",
				"description": "Review system performance metrics",
				"impact":      "Medium",
				"effort":      "Low",
				"priority":    "Medium",
				"steps":       []string{"Review metrics", "Identify bottlenecks"},
			},
		},
		"patterns":         []string{"Standard optimization patterns"},
		"confidence":       0.4,
		"estimated_impact": "Moderate improvements expected",
	}
}

func (capabilities *AICapabilities) fallbackLearning(success bool, duration int64, issues []DeploymentIssue) map[string]interface{} {
	insights := []string{"Deployment completed"}
	if !success {
		insights = append(insights, "Deployment encountered issues")
	}
	if duration > 300 { // 5 minutes
		insights = append(insights, "Deployment took longer than expected")
	}

	return map[string]interface{}{
		"insights":    insights,
		"patterns":    []string{"Standard deployment patterns"},
		"confidence":  0.3,
		"actionable":  false,
		"impact":      "Low",
		"categories":  []string{"deployment"},
		"trends":      []string{},
		"predictions": []string{},
	}
}
