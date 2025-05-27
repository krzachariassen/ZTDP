package events

import (
	"time"

	"github.com/google/uuid"
	"github.com/krzachariassen/ZTDP/internal/common"
)

// GraphEventService handles graph-related event emissions
type GraphEventService struct {
	eventBus *EventBus
	sourceID string
}

// NewGraphEventService creates a new graph event service
func NewGraphEventService(eventBus *EventBus, sourceID string) *GraphEventService {
	return &GraphEventService{
		eventBus: eventBus,
		sourceID: sourceID,
	}
}

// EmitNodeAdded sends an event when a node is added to the graph
func (s *GraphEventService) EmitNodeAdded(env string, nodeID string, kind string, metadata map[string]interface{}) error {
	event := Event{
		Type:      EventTypeGraphNodeAdded,
		Source:    s.sourceID,
		Subject:   nodeID,
		Action:    "add_node",
		Status:    "success",
		Timestamp: time.Now().UnixNano(),
		ID:        uuid.New().String(),
		Payload: map[string]interface{}{
			"environment": env,
			"node": common.EventNodeView{
				ID:       nodeID,
				Kind:     kind,
				Metadata: metadata,
			},
		},
	}

	return s.eventBus.PublishDefault(event)
}

// EmitNodeUpdated sends an event when a node is updated in the graph
func (s *GraphEventService) EmitNodeUpdated(env string, nodeID string, kind string, metadata map[string]interface{}) error {
	event := Event{
		Type:      EventTypeGraphNodeUpdated,
		Source:    s.sourceID,
		Subject:   nodeID,
		Action:    "update_node",
		Status:    "success",
		Timestamp: time.Now().UnixNano(),
		ID:        uuid.New().String(),
		Payload: map[string]interface{}{
			"environment": env,
			"node": common.EventNodeView{
				ID:       nodeID,
				Kind:     kind,
				Metadata: metadata,
			},
		},
	}

	return s.eventBus.PublishDefault(event)
}

// EmitNodeRemoved sends an event when a node is removed from the graph
func (s *GraphEventService) EmitNodeRemoved(env string, nodeID string) error {
	event := Event{
		Type:      EventTypeGraphNodeRemoved,
		Source:    s.sourceID,
		Subject:   nodeID,
		Action:    "remove_node",
		Status:    "success",
		Timestamp: time.Now().UnixNano(),
		ID:        uuid.New().String(),
		Payload: map[string]interface{}{
			"environment": env,
			"node_id":     nodeID,
		},
	}

	return s.eventBus.PublishDefault(event)
}

// EmitEdgeAdded sends an event when an edge is added to the graph
func (s *GraphEventService) EmitEdgeAdded(env string, fromID, toID, edgeType string) error {
	event := Event{
		Type:      EventTypeGraphEdgeAdded,
		Source:    s.sourceID,
		Subject:   fromID + "->" + toID,
		Action:    "add_edge",
		Status:    "success",
		Timestamp: time.Now().UnixNano(),
		ID:        uuid.New().String(),
		Payload: map[string]interface{}{
			"environment": env,
			"edge": common.EventEdgeView{
				From: fromID,
				To:   toID,
				Type: edgeType,
			},
		},
	}

	return s.eventBus.PublishDefault(event)
}

// EmitEdgeRemoved sends an event when an edge is removed from the graph
func (s *GraphEventService) EmitEdgeRemoved(env string, fromID, toID, edgeType string) error {
	event := Event{
		Type:      EventTypeGraphEdgeRemoved,
		Source:    s.sourceID,
		Subject:   fromID + "->" + toID,
		Action:    "remove_edge",
		Status:    "success",
		Timestamp: time.Now().UnixNano(),
		ID:        uuid.New().String(),
		Payload: map[string]interface{}{
			"environment": env,
			"edge": common.EventEdgeView{
				From: fromID,
				To:   toID,
				Type: edgeType,
			},
		},
	}

	return s.eventBus.PublishDefault(event)
}
