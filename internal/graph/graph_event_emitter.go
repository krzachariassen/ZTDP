package graph

import (
	"github.com/krzachariassen/ZTDP/internal/events"
)

// GraphEventEmitter extends the GraphStore with event emission capabilities
type GraphEventEmitter struct {
	*GraphStore
	eventService *events.GraphEventService
}

// NewGraphEventEmitter creates a new GraphEventEmitter
func NewGraphEventEmitter(store *GraphStore, eventService *events.GraphEventService) *GraphEventEmitter {
	return &GraphEventEmitter{
		GraphStore:   store,
		eventService: eventService,
	}
}

// AddNode adds a node to the store and emits an event
func (e *GraphEventEmitter) AddNode(env string, node *Node) error {
	// Check if node exists already to determine if this is an update
	existingNode, err := e.GetNode(env, node.ID)
	isUpdate := err == nil && existingNode != nil

	// Perform the actual operation
	err = e.GraphStore.AddNode(env, node)
	if err != nil {
		return err
	}

	// Emit event if event service is available
	if e.eventService != nil {
		if isUpdate {
			e.eventService.EmitNodeUpdated(env, node.ID, node.Kind, node.Metadata)
		} else {
			e.eventService.EmitNodeAdded(env, node.ID, node.Kind, node.Metadata)
		}
	}

	return nil
}

// AddEdge adds an edge to the store and emits an event
func (e *GraphEventEmitter) AddEdge(env, fromID, toID, relType string) error {
	// Perform the actual operation
	err := e.GraphStore.AddEdge(env, fromID, toID, relType)
	if err != nil {
		return err
	}

	// Emit event if event service is available
	if e.eventService != nil {
		e.eventService.EmitEdgeAdded(env, fromID, toID, relType)
	}

	return nil
}
