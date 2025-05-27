package events

import (
	"time"

	"github.com/google/uuid"
	"github.com/krzachariassen/ZTDP/internal/common"
)

// PolicyEventService handles policy-related event emissions
type PolicyEventService struct {
	eventBus *EventBus
	sourceID string
}

// NewPolicyEventService creates a new policy event service
func NewPolicyEventService(eventBus *EventBus, sourceID string) *PolicyEventService {
	return &PolicyEventService{
		eventBus: eventBus,
		sourceID: sourceID,
	}
}

// EmitPolicyCheck sends an event when a policy check is initiated
func (s *PolicyEventService) EmitPolicyCheck(
	policyID string,
	mutation common.Mutation,
	context map[string]interface{},
) error {
	event := Event{
		Type:      EventTypePolicyCheck,
		Source:    s.sourceID,
		Subject:   policyID,
		Action:    "check",
		Status:    "pending",
		Timestamp: time.Now().UnixNano(),
		ID:        uuid.New().String(),
		Payload: map[string]interface{}{
			"mutation": mutation,
			"context":  context,
		},
	}

	return s.eventBus.PublishDefault(event)
}

// EmitPolicyCheckResult sends an event with the result of a policy check
func (s *PolicyEventService) EmitPolicyCheckResult(
	policyID string,
	result bool,
	reason string,
	mutation common.Mutation,
) error {
	status := "approved"
	if !result {
		status = "rejected"
	}

	event := Event{
		Type:      EventTypePolicyCheckResult,
		Source:    s.sourceID,
		Subject:   policyID,
		Action:    "check.result",
		Status:    status,
		Timestamp: time.Now().UnixNano(),
		ID:        uuid.New().String(),
		Payload: map[string]interface{}{
			"result":   result,
			"reason":   reason,
			"mutation": mutation,
		},
	}

	return s.eventBus.PublishDefault(event)
}

// EmitTransitionAttempt sends an event when a transition is attempted
func (s *PolicyEventService) EmitTransitionAttempt(
	fromID, toID, edgeType string,
	user string,
) error {
	event := Event{
		Type:      EventTypeTransitionAttempt,
		Source:    s.sourceID,
		Subject:   fromID,
		Action:    "transition",
		Status:    "pending",
		Timestamp: time.Now().UnixNano(),
		ID:        uuid.New().String(),
		Payload: map[string]interface{}{
			"from_id":   fromID,
			"to_id":     toID,
			"edge_type": edgeType,
			"user":      user,
		},
	}

	return s.eventBus.PublishDefault(event)
}

// EmitTransitionResult sends an event with the result of a transition attempt
func (s *PolicyEventService) EmitTransitionResult(
	fromID, toID, edgeType string,
	user string,
	approved bool,
	reason string,
) error {
	eventType := EventTypeTransitionApproved
	status := "approved"

	if !approved {
		eventType = EventTypeTransitionRejected
		status = "rejected"
	}

	event := Event{
		Type:      eventType,
		Source:    s.sourceID,
		Subject:   fromID,
		Action:    "transition.result",
		Status:    status,
		Timestamp: time.Now().UnixNano(),
		ID:        uuid.New().String(),
		Payload: map[string]interface{}{
			"from_id":   fromID,
			"to_id":     toID,
			"edge_type": edgeType,
			"user":      user,
			"reason":    reason,
		},
	}

	return s.eventBus.PublishDefault(event)
}
