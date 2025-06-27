package domain

import (
	"fmt"
	"strings"
)

// ResponseParser handles parsing of AI responses into structured data
type ResponseParser struct{}

// NewResponseParser creates a new response parser
func NewResponseParser() *ResponseParser {
	return &ResponseParser{}
}

// ExtractSection extracts a specific section from AI response text
func (r *ResponseParser) ExtractSection(text, marker string) string {
	parts := strings.Split(text, marker)
	if len(parts) < 2 {
		return ""
	}

	section := parts[1]
	// Find the end of this section (next marker or end of text)
	nextMarkers := []string{"DECISION:", "CONFIDENCE:", "REASONING:", "CLARIFICATION:", "EXECUTION_PLAN:", "AGENT_COORDINATION:", "Intent:", "Category:", "Required_Agents:"}
	minIndex := len(section)

	for _, nextMarker := range nextMarkers {
		if nextMarker != marker { // Don't end on the same marker we're looking for
			if idx := strings.Index(section, nextMarker); idx > 0 && idx < minIndex {
				minIndex = idx
			}
		}
	}

	if minIndex < len(section) {
		section = section[:minIndex]
	}

	return strings.TrimSpace(section)
}

// ParseConfidence extracts confidence percentage from text
func (r *ResponseParser) ParseConfidence(confidenceStr string) int {
	// Simple extraction - look for numbers
	for i, char := range confidenceStr {
		if char >= '0' && char <= '9' {
			end := i
			for end < len(confidenceStr) && confidenceStr[end] >= '0' && confidenceStr[end] <= '9' {
				end++
			}
			if val := confidenceStr[i:end]; val != "" {
				var num int
				if n, err := fmt.Sscanf(val, "%d", &num); n == 1 && err == nil {
					return num
				}
			}
			break
		}
	}
	return 0
}

// ExtractIntent extracts and normalizes intent from analysis
func (r *ResponseParser) ExtractIntent(analysis string) string {
	intent := r.ExtractSection(analysis, "Intent:")
	if intent == "" {
		return "general_assistance"
	}
	return strings.ToLower(strings.ReplaceAll(intent, " ", "_"))
}

// ExtractCategory extracts and normalizes category from analysis
func (r *ResponseParser) ExtractCategory(analysis string) string {
	category := r.ExtractSection(analysis, "Category:")
	if category == "" {
		return "general"
	}
	return strings.ToLower(strings.ReplaceAll(category, " ", "_"))
}

// ExtractRequiredAgents parses required agents from analysis
func (r *ResponseParser) ExtractRequiredAgents(analysis string) []string {
	agentsStr := r.ExtractSection(analysis, "Required_Agents:")
	if agentsStr == "" {
		return []string{}
	}

	// Parse comma-separated agent names
	agents := strings.Split(agentsStr, ",")
	result := make([]string, 0, len(agents))
	for _, agent := range agents {
		agent = strings.TrimSpace(agent)
		if agent != "" && agent != "none" && agent != "None" {
			result = append(result, agent)
		}
	}
	return result
}
