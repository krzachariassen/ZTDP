package ai

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/krzachariassen/ZTDP/internal/logging"
)

// ResponseBuilder constructs structured responses for different AI capabilities
type ResponseBuilder struct {
	logger *logging.Logger
}

// NewResponseBuilder creates a new response builder
func NewResponseBuilder(logger *logging.Logger) *ResponseBuilder {
	return &ResponseBuilder{
		logger: logger,
	}
}

// BuildTroubleshootingResponse creates a structured troubleshooting response
func (builder *ResponseBuilder) BuildTroubleshootingResponse(
	analysis map[string]interface{},
) *TroubleshootingResponse {
	response := &TroubleshootingResponse{
		RootCause:          extractStringField(analysis, "root_cause", "Analysis in progress"),
		Confidence:         extractFloatField(analysis, "confidence", 0.5),
		Symptoms:           extractStringArrayField(analysis, "symptoms"),
		Diagnosis:          extractStringField(analysis, "diagnosis", ""),
		Recommendations:    extractStringArrayField(analysis, "recommendations"),
		NextSteps:          extractStringArrayField(analysis, "next_steps"),
		EstimatedTime:      extractStringField(analysis, "estimated_time", "Unknown"),
		Severity:           extractStringField(analysis, "severity", "Medium"),
		AffectedComponents: extractStringArrayField(analysis, "affected_components"),
		Timeline:           []EventTimestamp{}, // Would be populated from actual events
		Metadata: map[string]interface{}{
			"analysis_time": time.Now(),
			"ai_generated":  true,
		},
	}

	return response
}

// BuildImpactPrediction creates a structured impact prediction response
func (builder *ResponseBuilder) BuildImpactPrediction(
	prediction map[string]interface{},
) *ImpactPrediction {
	response := &ImpactPrediction{
		OverallRisk:       extractStringField(prediction, "overall_risk", "Medium"),
		Confidence:        extractFloatField(prediction, "confidence", 0.5),
		AffectedSystems:   extractStringArrayField(prediction, "affected_systems"),
		RiskFactors:       extractStringArrayField(prediction, "risk_factors"),
		Recommendations:   extractStringArrayField(prediction, "recommendations"),
		EstimatedDowntime: extractStringField(prediction, "estimated_downtime", "Unknown"),
		RollbackPlan:      extractStringField(prediction, "rollback_plan", ""),
		MonitoringPoints:  extractStringArrayField(prediction, "monitoring_points"),
		Timeline: []TimelineEvent{
			{
				Time:        time.Now(),
				Event:       "Impact prediction generated",
				Probability: 1.0,
			},
		},
		Metadata: map[string]interface{}{
			"prediction_time": time.Now(),
			"ai_generated":    true,
		},
	}

	return response
}

// BuildOptimizationRecommendations creates structured optimization recommendations
func (builder *ResponseBuilder) BuildOptimizationRecommendations(
	optimization map[string]interface{},
) *OptimizationRecommendations {
	response := &OptimizationRecommendations{
		Recommendations: builder.extractRecommendations(optimization),
		Patterns:        extractStringArrayField(optimization, "patterns"),
		Confidence:      extractFloatField(optimization, "confidence", 0.5),
		EstimatedImpact: extractStringField(optimization, "estimated_impact", "Unknown"),
		Priority:        extractStringField(optimization, "priority", "Medium"),
		Timeline:        extractStringField(optimization, "timeline", "Unknown"),
		Resources:       extractStringArrayField(optimization, "required_resources"),
		RiskLevel:       extractStringField(optimization, "risk_level", "Low"),
		Validation:      extractStringArrayField(optimization, "validation_steps"),
		Metadata: map[string]interface{}{
			"optimization_time": time.Now(),
			"ai_generated":      true,
		},
	}

	return response
}

// BuildLearningInsights creates structured learning insights
func (builder *ResponseBuilder) BuildLearningInsights(
	learning map[string]interface{},
) *LearningInsights {
	response := &LearningInsights{
		Insights:    extractStringArrayField(learning, "insights"),
		Patterns:    extractStringArrayField(learning, "patterns"),
		Confidence:  extractFloatField(learning, "confidence", 0.5),
		Actionable:  extractBoolField(learning, "actionable", false),
		Impact:      extractStringField(learning, "impact", "Medium"),
		Categories:  extractStringArrayField(learning, "categories"),
		Trends:      extractStringArrayField(learning, "trends"),
		Predictions: extractStringArrayField(learning, "predictions"),
		Metadata: map[string]interface{}{
			"learning_time": time.Now(),
			"ai_generated":  true,
		},
	}

	return response
}

// extractRecommendations extracts recommendation objects from the analysis
func (builder *ResponseBuilder) extractRecommendations(data map[string]interface{}) []Recommendation {
	recommendations := []Recommendation{}

	if recData, ok := data["recommendations"].([]interface{}); ok {
		for _, rec := range recData {
			if recMap, ok := rec.(map[string]interface{}); ok {
				recommendation := Recommendation{
					Title:       extractStringField(recMap, "title", "Optimization Recommendation"),
					Description: extractStringField(recMap, "description", ""),
					Impact:      extractStringField(recMap, "impact", "Medium"),
					Effort:      extractStringField(recMap, "effort", "Medium"),
					Priority:    extractStringField(recMap, "priority", "Medium"),
					Category:    extractStringField(recMap, "category", "General"),
					Steps:       extractStringArrayField(recMap, "steps"),
					Risks:       extractStringArrayField(recMap, "risks"),
					Benefits:    extractStringArrayField(recMap, "benefits"),
					Timeline:    extractStringField(recMap, "timeline", "Unknown"),
				}
				recommendations = append(recommendations, recommendation)
			}
		}
	}

	// If no structured recommendations found, create a generic one
	if len(recommendations) == 0 && len(extractStringArrayField(data, "recommendations")) > 0 {
		recommendations = append(recommendations, Recommendation{
			Title:       "Platform Optimization",
			Description: "AI-generated optimization recommendations",
			Impact:      "Medium",
			Effort:      "Medium",
			Priority:    "Medium",
			Category:    "General",
			Steps:       extractStringArrayField(data, "recommendations"),
		})
	}

	return recommendations
}

// Helper functions to safely extract fields from map[string]interface{}

func extractStringField(data map[string]interface{}, key, defaultValue string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return defaultValue
}

func extractFloatField(data map[string]interface{}, key string, defaultValue float64) float64 {
	if value, ok := data[key].(float64); ok {
		return value
	}
	if value, ok := data[key].(int); ok {
		return float64(value)
	}
	return defaultValue
}

func extractBoolField(data map[string]interface{}, key string, defaultValue bool) bool {
	if value, ok := data[key].(bool); ok {
		return value
	}
	return defaultValue
}

func extractStringArrayField(data map[string]interface{}, key string) []string {
	var result []string

	if value, ok := data[key].([]interface{}); ok {
		for _, item := range value {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
	} else if value, ok := data[key].([]string); ok {
		result = value
	}

	return result
}

// ParseJSONResponse safely parses a JSON response into a map
func (builder *ResponseBuilder) ParseJSONResponse(jsonStr string) (map[string]interface{}, error) {
	var result map[string]interface{}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return result, nil
}
