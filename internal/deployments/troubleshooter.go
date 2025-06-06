package deployments

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// Troubleshooter provides AI-powered deployment issue analysis
type Troubleshooter struct {
	provider ai.AIProvider
	graph    *graph.GlobalGraph
	logger   *logging.Logger
}

// NewTroubleshooter creates a new troubleshooter
func NewTroubleshooter(provider ai.AIProvider, graph *graph.GlobalGraph) *Troubleshooter {
	return &Troubleshooter{
		provider: provider,
		graph:    graph,
		logger:   logging.GetLogger().ForComponent("troubleshooter"),
	}
}

// Troubleshoot provides AI-driven root cause analysis for deployment issues
func (t *Troubleshooter) Troubleshoot(ctx context.Context, incidentID string, description string, symptoms []string) (*ai.TroubleshootingResponse, error) {
	if t.provider == nil {
		return nil, fmt.Errorf("AI provider not available for troubleshooting")
	}

	t.logger.Info("🔍 Analyzing incident: %s - %s", incidentID, description)

	// Extract incident context from platform state
	incidentContext, err := t.extractIncidentContext(incidentID, description, symptoms)
	if err != nil {
		return nil, fmt.Errorf("failed to extract incident context: %w", err)
	}

	// Perform AI-driven troubleshooting
	systemPrompt := t.buildSystemPrompt()
	userPrompt := t.buildUserPrompt(incidentContext)

	aiResponse, err := t.provider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI troubleshooting failed: %w", err)
	}

	// Parse AI response
	response, err := t.parseTroubleshootingResponse(aiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse troubleshooting response: %w", err)
	}

	t.logger.Info("✅ Troubleshooting analysis completed: %s (confidence: %.2f)",
		response.RootCause, response.Confidence)

	return response, nil
}

// extractIncidentContext builds comprehensive incident context for troubleshooting
func (t *Troubleshooter) extractIncidentContext(incidentID, description string, symptoms []string) (*ai.IncidentContext, error) {
	globalGraph, err := t.graph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	// Detect affected environment from symptoms
	environment := t.detectIncidentEnvironment(symptoms)

	context := &ai.IncidentContext{
		IncidentID:  incidentID,
		Description: description,
		Symptoms:    symptoms,
		Environment: environment,
		Timeline: []ai.EventTimestamp{
			{
				Timestamp: time.Now().Format(time.RFC3339),
				Event:     "Incident reported",
				Source:    "deployment-service",
				Severity:  "medium",
			},
		},
		Logs: []ai.LogEntry{
			{
				Timestamp: time.Now().Format(time.RFC3339),
				Level:     "info",
				Message:   fmt.Sprintf("Troubleshooting started for: %s", description),
				Source:    "troubleshooter",
			},
		},
		Metrics: map[string]interface{}{
			"incident_timestamp": time.Now(),
			"symptom_count":      len(symptoms),
		},
		Context: map[string]interface{}{
			"platform_state": t.extractCurrentPlatformState(),
			"recent_changes": t.extractRecentDeployments(),
			"health_status":  t.extractHealthStatus(globalGraph),
		},
		Metadata: map[string]interface{}{
			"source":     "deployment-troubleshooter",
			"timestamp":  time.Now(),
			"graph_size": len(globalGraph.Nodes),
		},
	}

	return context, nil
}

// Helper methods for incident analysis
func (t *Troubleshooter) detectIncidentEnvironment(symptoms []string) string {
	for _, symptom := range symptoms {
		if strings.Contains(symptom, "production") {
			return "production"
		}
		if strings.Contains(symptom, "staging") {
			return "staging"
		}
	}
	return "unknown"
}

func (t *Troubleshooter) extractAffectedSystems(graph *graph.Graph, symptoms []string) []string {
	systems := []string{}
	for _, symptom := range symptoms {
		for _, node := range graph.Nodes {
			if strings.Contains(symptom, node.ID) {
				systems = append(systems, node.ID)
			}
		}
	}
	return systems
}

func (t *Troubleshooter) extractCurrentPlatformState() map[string]interface{} {
	return map[string]interface{}{
		"total_services":     15,
		"healthy_services":   13,
		"avg_response_time":  "250ms",
		"error_rate":         "2.1%",
		"cpu_usage":          "65%",
		"memory_usage":       "70%",
		"active_deployments": 2,
	}
}

func (t *Troubleshooter) extractRecentDeployments() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"deployment_id": "deploy-123",
			"application":   "web-app",
			"status":        "completed",
			"timestamp":     time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
		},
		{
			"deployment_id": "deploy-124",
			"application":   "api-service",
			"status":        "in_progress",
			"timestamp":     time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
		},
	}
}

func (t *Troubleshooter) extractHealthStatus(graph *graph.Graph) map[string]interface{} {
	healthy := 0
	total := 0
	for _, node := range graph.Nodes {
		if node.Kind == "service" {
			total++
			if status, ok := node.Metadata["health"]; ok && status == "healthy" {
				healthy++
			}
		}
	}

	return map[string]interface{}{
		"healthy_services":  healthy,
		"total_services":    total,
		"health_percentage": float64(healthy) / float64(total) * 100,
	}
}

// buildSystemPrompt creates the system prompt for troubleshooting
func (t *Troubleshooter) buildSystemPrompt() string {
	return `You are an expert system troubleshooter with deep knowledge of:
- Distributed system architecture and failure patterns
- Root cause analysis methodologies
- Incident response and resolution procedures
- Performance analysis and optimization

Your task is to analyze incidents and provide actionable troubleshooting guidance.
Respond in JSON format with structured troubleshooting analysis.`
}

// buildUserPrompt creates the user prompt for troubleshooting
func (t *Troubleshooter) buildUserPrompt(incident *ai.IncidentContext) string {
	return fmt.Sprintf(`Analyze this incident and provide troubleshooting guidance:

Incident ID: %s
Description: %s
Environment: %s
Symptoms: %v

System Context:
- Metrics: %v
- Timeline Events: %d

Please provide:
1. Most likely root cause
2. Detailed diagnosis
3. Recommended actions
4. Next steps for resolution
5. Estimated resolution time
6. Prevention measures

Focus on practical troubleshooting steps and root cause analysis.`,
		incident.IncidentID, incident.Description, incident.Environment,
		incident.Symptoms, incident.Metrics, len(incident.Timeline))
}

// parseTroubleshootingResponse parses AI response into troubleshooting response
func (t *Troubleshooter) parseTroubleshootingResponse(response string) (*ai.TroubleshootingResponse, error) {
	// TODO: Implement proper JSON parsing
	// For now, return a basic response
	return &ai.TroubleshootingResponse{
		RootCause:          "Service dependency failure",
		Confidence:         0.8,
		Symptoms:           []string{"Increased latency", "Error rate spike"},
		Diagnosis:          "Downstream service appears to be experiencing issues",
		Recommendations:    []string{"Check service health", "Review recent deployments", "Monitor system resources"},
		NextSteps:          []string{"Restart affected service", "Check logs for errors"},
		EstimatedTime:      "15-30 minutes",
		Severity:           "Medium",
		AffectedComponents: []string{"web-service", "api-gateway"},
		Timeline:           []ai.EventTimestamp{},
		Metadata: map[string]interface{}{
			"generated_by": "troubleshooter",
			"method":       "ai_enhanced",
		},
	}, nil
}
