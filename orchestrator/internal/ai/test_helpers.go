package ai

import (
	"context"
	"os"
	"testing"

	"github.com/ztdp/orchestrator/internal/graph"
)

// TestHelpers provides shared test utilities for AI orchestrator tests
// This centralizes common setup code and makes tests more maintainable

// SetupOpenAIProvider creates a real OpenAI provider for testing
// This is the standard way to get an OpenAI provider in tests
func SetupOpenAIProvider(t *testing.T) AIProvider {
	t.Helper()

	// Skip if no OpenAI API key is available
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY environment variable not set, skipping OpenAI integration test")
	}

	// Setup real OpenAI provider with consistent configuration
	logger := NewTestLogger(t.Name())
	config := DefaultOpenAIConfig()
	config.Temperature = 0.1 // Low temperature for consistent, logical responses

	provider, err := NewOpenAIProvider(config, apiKey, logger)
	if err != nil {
		t.Fatalf("Failed to create OpenAI provider: %v", err)
	}

	return provider
}

// NewTestLogger creates a logger for test use with a consistent format
func NewTestLogger(name string) Logger {
	return NewSimpleLogger(name)
}

// SetupEmbeddedGraph creates an embedded graph for testing
func SetupEmbeddedGraph(t *testing.T) graph.Graph {
	t.Helper()
	logger := NewTestLogger(t.Name())
	return graph.NewEmbeddedGraph(logger)
}

// SetupGraphPoweredOrchestrator creates a graph-powered orchestrator for testing
func SetupGraphPoweredOrchestrator(t *testing.T) (*GraphPoweredAIOrchestrator, graph.Graph) {
	t.Helper()

	provider := SetupOpenAIProvider(t)
	logger := NewTestLogger(t.Name())
	testGraph := SetupEmbeddedGraph(t)

	orchestrator := NewGraphPoweredAIOrchestrator(provider, testGraph, logger)
	return orchestrator, testGraph
}

// SetupBasicTestGraph populates the graph with basic test data (agents and workflows)
func SetupBasicTestGraph(t *testing.T, g graph.Graph) {
	t.Helper()
	ctx := context.Background()

	// Add basic test agents
	agents := []map[string]interface{}{
		{
			"id":           "agent-deploy",
			"name":         "Deployment Agent",
			"type":         "deployment",
			"status":       "active",
			"capabilities": []string{"deploy", "rollback", "scale"},
			"endpoint":     "http://deploy-agent:8080",
		},
		{
			"id":           "agent-monitor",
			"name":         "Monitoring Agent",
			"type":         "monitoring",
			"status":       "active",
			"capabilities": []string{"metrics", "alerts", "dashboards"},
			"endpoint":     "http://monitor-agent:8080",
		},
	}

	for _, agent := range agents {
		if err := g.AddNode(ctx, "agent", agent["id"].(string), agent); err != nil {
			t.Fatalf("Failed to add agent: %v", err)
		}
	}

	// Add basic test workflows
	workflows := []map[string]interface{}{
		{
			"id":          "workflow-safe-deploy",
			"name":        "Safe Production Deployment",
			"description": "Deployment with canary rollout and monitoring",
			"steps":       []string{"validate", "canary", "monitor", "full-deploy"},
		},
	}

	for _, workflow := range workflows {
		if err := g.AddNode(ctx, "workflow", workflow["id"].(string), workflow); err != nil {
			t.Fatalf("Failed to add workflow: %v", err)
		}
	}
}

// SetupRichTestGraph creates a more comprehensive graph for complex testing scenarios
func SetupRichTestGraph(t *testing.T, g graph.Graph) {
	t.Helper()
	ctx := context.Background()

	// Start with basic setup
	SetupBasicTestGraph(t, g)

	// Add more sophisticated agents
	sophisticatedAgents := []map[string]interface{}{
		{
			"id":           "agent-kubernetes",
			"name":         "Kubernetes Orchestrator",
			"type":         "orchestration",
			"status":       "active",
			"capabilities": []string{"pod-management", "scaling", "rolling-updates", "health-checks"},
			"endpoint":     "http://k8s-agent:8080",
		},
		{
			"id":           "agent-nodejs",
			"name":         "Node.js Specialist",
			"type":         "runtime",
			"status":       "active",
			"capabilities": []string{"memory-profiling", "performance-tuning", "dependency-management"},
			"endpoint":     "http://nodejs-agent:8080",
		},
		{
			"id":           "agent-database",
			"name":         "Database Optimizer",
			"type":         "database",
			"status":       "active",
			"capabilities": []string{"mongodb-tuning", "redis-optimization", "connection-pooling", "query-analysis"},
			"endpoint":     "http://db-agent:8080",
		},
		{
			"id":           "agent-monitoring",
			"name":         "Advanced Monitoring",
			"type":         "observability",
			"status":       "active",
			"capabilities": []string{"alerting", "dashboards", "log-analysis", "performance-metrics"},
			"endpoint":     "http://monitor-agent:8080",
		},
		{
			"id":           "agent-security",
			"name":         "Security Guardian",
			"type":         "security",
			"status":       "active",
			"capabilities": []string{"vulnerability-scanning", "jwt-validation", "ssl-management"},
			"endpoint":     "http://security-agent:8080",
		},
	}

	for _, agent := range sophisticatedAgents {
		if err := g.AddNode(ctx, "agent", agent["id"].(string), agent); err != nil {
			t.Fatalf("Failed to add sophisticated agent: %v", err)
		}
	}

	// Add complex workflows
	complexWorkflows := []map[string]interface{}{
		{
			"id":          "workflow-crisis-response",
			"name":        "Crisis Response Protocol",
			"description": "Emergency response for production issues during high-traffic events",
			"steps":       []string{"assess-impact", "stabilize-system", "implement-monitoring", "gradual-recovery"},
			"agents":      []string{"agent-kubernetes", "agent-nodejs", "agent-database", "agent-monitoring"},
		},
		{
			"id":          "workflow-performance-optimization",
			"name":        "Performance Optimization Suite",
			"description": "Comprehensive performance improvement for Node.js microservices",
			"steps":       []string{"profile-memory", "optimize-queries", "tune-cache", "validate-improvements"},
			"agents":      []string{"agent-nodejs", "agent-database", "agent-monitoring"},
		},
		{
			"id":          "workflow-safe-deployment",
			"name":        "Zero-Downtime Deployment",
			"description": "Production deployment with canary rollout and comprehensive monitoring",
			"steps":       []string{"security-scan", "canary-deploy", "monitor-metrics", "full-rollout"},
			"agents":      []string{"agent-kubernetes", "agent-monitoring", "agent-security"},
		},
	}

	for _, workflow := range complexWorkflows {
		if err := g.AddNode(ctx, "workflow", workflow["id"].(string), workflow); err != nil {
			t.Fatalf("Failed to add complex workflow: %v", err)
		}
	}
}

// TestScenarios provides common test scenarios for consistency across tests
type TestScenario struct {
	Name        string
	Input       string
	UserID      string
	ExpectType  string
	Description string
}

// GetBasicTestScenarios returns standard test scenarios for basic orchestrator testing
func GetBasicTestScenarios() []TestScenario {
	return []TestScenario{
		{
			Name:        "simple deployment request",
			Input:       "Deploy myapp to production",
			UserID:      "test_user",
			ExpectType:  "deployment",
			Description: "Basic deployment request should be understood and handled",
		},
		{
			Name:        "monitoring setup request",
			Input:       "Set up monitoring for my application",
			UserID:      "test_user",
			ExpectType:  "monitoring",
			Description: "Monitoring request should leverage monitoring agent capabilities",
		},
		{
			Name:        "capability discovery",
			Input:       "What can this platform help me with?",
			UserID:      "new_user",
			ExpectType:  "discovery",
			Description: "Platform capability discovery should show agent knowledge",
		},
	}
}

// GetComplexTestScenarios returns advanced test scenarios for comprehensive testing
func GetComplexTestScenarios() []TestScenario {
	return []TestScenario{
		{
			Name:   "crisis management scenario",
			UserID: "crisis_user",
			Input: `We have a Node.js microservice for user authentication that's been running in production for 3 months. 
        Here's our current situation:

        TECHNICAL CONTEXT:
        - Service handles 50k requests/day
        - MongoDB backend with Redis cache
        - Running on Kubernetes with 3 replicas
        - Current version: v2.1.4

        BUSINESS CONTEXT:
        - Black Friday sale starts tomorrow
        - Expecting 10x traffic spike
        - Marketing team already sent newsletters
        - CEO is breathing down our necks

        CURRENT ISSUES:
        - Memory leaks causing OOM kills every 6 hours
        - JWT validation taking 200ms (should be 5ms)
        - Redis connection pool exhausted during peaks
        - New hire deployed a hotfix yesterday that may have introduced bugs

        REQUIREMENTS:
        - Must be stable for tomorrow's traffic
        - Can't afford any downtime during business hours
        - Need monitoring and alerting in place
        - Want to implement gradual rollout for safety

        What's your complete strategy to handle this before tomorrow?`,
			ExpectType:  "crisis_response",
			Description: "Complex crisis scenario should trigger sophisticated multi-agent response",
		},
		{
			Name:        "performance optimization request",
			Input:       "My application is slow during peak hours, help me optimize it",
			UserID:      "performance_user",
			ExpectType:  "optimization",
			Description: "Performance optimization should leverage specialized agents",
		},
		{
			Name:        "security and deployment",
			Input:       "Deploy my application securely with full monitoring",
			UserID:      "security_user",
			ExpectType:  "secure_deployment",
			Description: "Security-focused deployment should use security and monitoring agents",
		},
	}
}

// AssertGraphKnowledge checks if a response demonstrates graph knowledge
func AssertGraphKnowledge(t *testing.T, response *ConversationalResponse, description string) {
	t.Helper()

	if response == nil {
		t.Fatalf("%s: response is nil", description)
	}

	// Check for evidence of graph exploration
	hasGraphContext := response.Context != nil && response.Context["graph_context"] != nil
	if !hasGraphContext {
		t.Errorf("%s: response lacks graph context", description)
	}

	// Check for sophisticated response (graph-powered responses should be more detailed)
	if len(response.Message) < 100 {
		t.Errorf("%s: response too short, expected detailed graph-powered response", description)
	}

	// Check for agent-specific knowledge (responses should mention specific agents)
	responseText := response.Message
	hasAgentKnowledge := false
	agentIndicators := []string{"Agent", "agent", "Specialist", "Orchestrator", "Guardian", "Optimizer"}
	for _, indicator := range agentIndicators {
		if contains(responseText, indicator) {
			hasAgentKnowledge = true
			break
		}
	}

	if !hasAgentKnowledge {
		t.Errorf("%s: response doesn't show agent knowledge", description)
	}
}

// Helper function to check if text contains substring
func contains(text, substr string) bool {
	return len(text) > 0 && len(substr) > 0 &&
		(text == substr || len(text) >= len(substr) &&
			(text[:len(substr)] == substr || text[len(text)-len(substr):] == substr ||
				findInString(text, substr)))
}

func findInString(text, substr string) bool {
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TruncateString truncates a string for display purposes
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}
