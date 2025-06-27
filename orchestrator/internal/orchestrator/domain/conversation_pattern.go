package domain

import (
	"time"
)

// ConversationPattern represents patterns learned from conversation history
type ConversationPattern struct {
	SessionID         string         `json:"session_id"`
	CommonIntents     []string       `json:"common_intents"`
	FrequentEntities  map[string]int `json:"frequent_entities"`
	TotalInteractions int            `json:"total_interactions"`
	SuccessRate       float64        `json:"success_rate"`
	TopActions        []string       `json:"top_actions"`
	LastAnalyzed      time.Time      `json:"last_analyzed"`
}

// NewConversationPattern creates a new conversation pattern
func NewConversationPattern(sessionID string) *ConversationPattern {
	return &ConversationPattern{
		SessionID:         sessionID,
		CommonIntents:     make([]string, 0),
		FrequentEntities:  make(map[string]int),
		TotalInteractions: 0,
		SuccessRate:       0.0,
		TopActions:        make([]string, 0),
		LastAnalyzed:      time.Now(),
	}
}

// AddIntent adds an intent to the pattern analysis
func (cp *ConversationPattern) AddIntent(intent string) {
	// Check if intent is already in the list
	for _, existingIntent := range cp.CommonIntents {
		if existingIntent == intent {
			return
		}
	}
	cp.CommonIntents = append(cp.CommonIntents, intent)
}

// AddEntity adds or increments an entity frequency
func (cp *ConversationPattern) AddEntity(entity string) {
	cp.FrequentEntities[entity]++
}

// AddAction adds an action to the top actions if not already present
func (cp *ConversationPattern) AddAction(action string) {
	for _, existingAction := range cp.TopActions {
		if existingAction == action {
			return
		}
	}
	cp.TopActions = append(cp.TopActions, action)
}

// IncrementInteractions increments the total interaction count
func (cp *ConversationPattern) IncrementInteractions() {
	cp.TotalInteractions++
}

// UpdateSuccessRate updates the success rate based on successful vs total interactions
func (cp *ConversationPattern) UpdateSuccessRate(successfulInteractions int) {
	if cp.TotalInteractions > 0 {
		cp.SuccessRate = float64(successfulInteractions) / float64(cp.TotalInteractions)
	}
}
