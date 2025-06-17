package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/krzachariassen/ZTDP/internal/logging"
)

// IntentRecognizer analyzes user queries to understand intent and extract parameters
// This enables the platform to route requests to appropriate domain services
type IntentRecognizer struct {
	provider AIProvider
	logger   *logging.Logger
}

// NewIntentRecognizer creates a new intent recognition engine
func NewIntentRecognizer(provider AIProvider, logger *logging.Logger) *IntentRecognizer {
	return &IntentRecognizer{
		provider: provider,
		logger:   logger,
	}
}

// AnalyzeIntent processes a user query to determine intent and extract parameters
func (recognizer *IntentRecognizer) AnalyzeIntent(
	ctx context.Context,
	query string,
	platformContext *PlatformContext,
) (*Intent, error) {
	recognizer.logger.Info("🧠 Analyzing intent for query: %s", query)

	// Build intent analysis prompts
	systemPrompt := recognizer.buildIntentSystemPrompt(platformContext)
	userPrompt := recognizer.buildIntentUserPrompt(query, platformContext)

	// Analyze intent using AI
	rawResponse, err := recognizer.provider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI intent analysis failed: %w", err)
	}

	// Parse intent from AI response
	intent, err := recognizer.parseIntentResponse(rawResponse, query)
	if err != nil {
		// Fallback to simple pattern matching if AI parsing fails
		recognizer.logger.Warn("AI intent parsing failed, using fallback: %v", err)
		return recognizer.fallbackIntentDetection(query, platformContext), nil
	}

	recognizer.logger.Info("✅ Intent analyzed: %s with confidence %.2f", intent.Type, intent.Confidence)
	return intent, nil
}

// buildIntentSystemPrompt creates the system prompt for intent recognition
func (recognizer *IntentRecognizer) buildIntentSystemPrompt(context *PlatformContext) string {
	return `You are an expert platform orchestration agent for ZTDP that understands complex infrastructure requests and can create execution plans.

Your job is to analyze user queries and determine:
1. Whether this is a simple single-step operation or complex multi-step scenario
2. All the resources and operations involved
3. The correct sequence and dependencies
4. Required parameters for each step

AVAILABLE INTENT TYPES:
- complex_orchestration: Multi-step scenarios requiring multiple operations
- application_creation: Create new applications
- service_creation: Create new services  
- resource_creation: Create infrastructure resources (databases, storage, etc.)
- environment_creation: Create new environments
- policy_creation: Create policies
- deployment: Deploy applications or services
- resource_linking: Link services to resources
- policy_check: Validate policies
- analysis: Analyze platform state
- troubleshooting: Diagnose problems
- question: General questions
- optimization: Performance optimization

RESOURCE TYPES:
- application: Top-level application container
- service: API services, microservices, etc.
- database: SQL/NoSQL databases
- storage: File storage, object storage
- queue: Message queues, event systems
- environment: deployment environments

COMPLEX ORCHESTRATION EXAMPLES:
"Build an application with API that stores data in database" = 
1. Create application
2. Create database resource
3. Create API service
4. Link service to database
5. Link service to application

RESPONSE FORMAT (JSON):
{
  "type": "complex_orchestration" | "simple_operation",
  "confidence": 0.0-1.0,
  "scenario": "description of what user wants to achieve",
  "steps": [
    {
      "operation": "application_creation|service_creation|resource_creation|resource_linking",
      "resource_type": "application|service|database|storage|queue",
      "resource_name": "extracted_name",
      "description": "what this step does",
      "dependencies": ["step_1", "step_2"], // which steps must complete first
      "parameters": {}
    }
  ],
  "parameters": {
    "primary_resource": "main_resource_name",
    "environment": "environment_name",
    "description": "overall_description"
  }
}

For simple operations, provide single step. For complex scenarios, break into logical steps with dependencies.`
}

// buildIntentUserPrompt creates the user prompt with query and context
func (recognizer *IntentRecognizer) buildIntentUserPrompt(query string, context *PlatformContext) string {
	prompt := fmt.Sprintf("Analyze this user query: \"%s\"\n\n", query)

	// Add available applications for context
	if len(context.Applications) > 0 {
		prompt += "Available applications: "
		apps := []string{}
		for app := range context.Applications {
			apps = append(apps, app)
		}
		prompt += strings.Join(apps, ", ") + "\n\n"
	}

	// Add available environments for context
	if len(context.Environments) > 0 {
		prompt += "Available environments: "
		envs := []string{}
		for env := range context.Environments {
			envs = append(envs, env)
		}
		prompt += strings.Join(envs, ", ") + "\n\n"
	}

	prompt += "Return JSON with intent analysis."

	return prompt
}

// parseIntentResponse parses the AI response into an Intent struct
func (recognizer *IntentRecognizer) parseIntentResponse(rawResponse, originalQuery string) (*Intent, error) {
	// Try to extract JSON from the response
	jsonStart := strings.Index(rawResponse, "{")
	jsonEnd := strings.LastIndex(rawResponse, "}") + 1

	if jsonStart == -1 || jsonEnd <= jsonStart {
		return nil, fmt.Errorf("no JSON found in response")
	}

	jsonStr := rawResponse[jsonStart:jsonEnd]

	var intentData struct {
		Type       string                 `json:"type"`
		Confidence float64                `json:"confidence"`
		Parameters map[string]interface{} `json:"parameters"`
		Reasoning  string                 `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &intentData); err != nil {
		return nil, fmt.Errorf("failed to parse intent JSON: %w", err)
	}

	intent := &Intent{
		Type:       intentData.Type,
		Confidence: intentData.Confidence,
		Parameters: intentData.Parameters,
		Query:      originalQuery,
		Reasoning:  intentData.Reasoning,
	}

	return intent, nil
}

// fallbackIntentDetection provides simple pattern-based intent detection as fallback
func (recognizer *IntentRecognizer) fallbackIntentDetection(query string, context *PlatformContext) *Intent {
	queryLower := strings.ToLower(query)

	// Deployment patterns
	if containsAny(queryLower, []string{"deploy", "deployment", "release", "rollout"}) {
		return &Intent{
			Type:       "deployment",
			Confidence: 0.7,
			Parameters: recognizer.extractApplicationFromQuery(queryLower, context),
			Query:      query,
			Reasoning:  "Pattern match: deployment keywords detected",
		}
	}

	// Policy patterns
	if containsAny(queryLower, []string{"policy", "compliance", "validate", "check"}) {
		return &Intent{
			Type:       "policy_check",
			Confidence: 0.7,
			Parameters: recognizer.extractApplicationFromQuery(queryLower, context),
			Query:      query,
			Reasoning:  "Pattern match: policy keywords detected",
		}
	}

	// Analysis patterns
	if containsAny(queryLower, []string{"status", "health", "show", "list", "analyze"}) {
		return &Intent{
			Type:       "analysis",
			Confidence: 0.7,
			Parameters: recognizer.extractApplicationFromQuery(queryLower, context),
			Query:      query,
			Reasoning:  "Pattern match: analysis keywords detected",
		}
	}

	// Troubleshooting patterns
	if containsAny(queryLower, []string{"error", "issue", "problem", "debug", "troubleshoot", "fail"}) {
		return &Intent{
			Type:       "troubleshooting",
			Confidence: 0.7,
			Parameters: recognizer.extractApplicationFromQuery(queryLower, context),
			Query:      query,
			Reasoning:  "Pattern match: troubleshooting keywords detected",
		}
	}

	// Question patterns
	if containsAny(queryLower, []string{"what", "how", "why", "when", "where", "?"}) {
		return &Intent{
			Type:       "question",
			Confidence: 0.6,
			Parameters: map[string]interface{}{},
			Query:      query,
			Reasoning:  "Pattern match: question keywords detected",
		}
	}

	// Default to question intent
	return &Intent{
		Type:       "question",
		Confidence: 0.5,
		Parameters: map[string]interface{}{},
		Query:      query,
		Reasoning:  "Fallback: no specific pattern matched",
	}
}

// extractApplicationFromQuery tries to extract application names from the query
func (recognizer *IntentRecognizer) extractApplicationFromQuery(query string, context *PlatformContext) map[string]interface{} {
	parameters := make(map[string]interface{})

	// Check if any known applications are mentioned
	for appName := range context.Applications {
		if strings.Contains(query, strings.ToLower(appName)) {
			parameters["app"] = appName
			break
		}
	}

	// Check if any known environments are mentioned
	for envName := range context.Environments {
		if strings.Contains(query, strings.ToLower(envName)) {
			parameters["environment"] = envName
			break
		}
	}

	// Extract common environment keywords
	if strings.Contains(query, "prod") || strings.Contains(query, "production") {
		parameters["environment"] = "production"
	} else if strings.Contains(query, "stag") || strings.Contains(query, "staging") {
		parameters["environment"] = "staging"
	} else if strings.Contains(query, "dev") || strings.Contains(query, "development") {
		parameters["environment"] = "development"
	}

	return parameters
}

// containsAny checks if a string contains any of the specified substrings
func containsAny(s string, substrings []string) bool {
	for _, substring := range substrings {
		if strings.Contains(s, substring) {
			return true
		}
	}
	return false
}
