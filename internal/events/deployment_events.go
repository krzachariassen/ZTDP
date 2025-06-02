package events

import (
	"github.com/krzachariassen/ZTDP/internal/contracts"
)

// Deployment event types
const (
	EventTypeDeploymentStatusChanged  EventType = "deployment.status.changed"
	EventTypeDeploymentProgressUpdate EventType = "deployment.progress.update"
	EventTypeDeploymentStarted        EventType = "deployment.started"
	EventTypeDeploymentCompleted      EventType = "deployment.completed"
	EventTypeDeploymentFailed         EventType = "deployment.failed"
	EventTypeDeploymentCancelled      EventType = "deployment.cancelled"
)

// DeploymentEventBus extends the EventBus with deployment-specific functionality
type DeploymentEventBus struct {
	*EventBus
}

// NewDeploymentEventBus creates a new deployment event bus
func NewDeploymentEventBus(transport EventTransport, defaultAsync bool) *DeploymentEventBus {
	return &DeploymentEventBus{
		EventBus: NewEventBus(transport, defaultAsync),
	}
}

// PublishStatusChange publishes a deployment status change event
func (d *DeploymentEventBus) PublishStatusChange(edgeID string, oldStatus, newStatus contracts.DeploymentStatus, message string) error {
	payload := map[string]interface{}{
		"edge_id":    edgeID,
		"old_status": string(oldStatus),
		"new_status": string(newStatus),
		"message":    message,
	}

	return d.Emit(EventTypeDeploymentStatusChanged, "deployment-engine", edgeID, payload)
}

// PublishProgressUpdate publishes a deployment progress update event
func (d *DeploymentEventBus) PublishProgressUpdate(edgeID string, progress float64, message string) error {
	payload := map[string]interface{}{
		"edge_id":  edgeID,
		"progress": progress,
		"message":  message,
	}

	return d.Emit(EventTypeDeploymentProgressUpdate, "deployment-engine", edgeID, payload)
}

// PublishDeploymentStarted publishes a deployment started event
func (d *DeploymentEventBus) PublishDeploymentStarted(edgeID string, source, target string) error {
	payload := map[string]interface{}{
		"edge_id": edgeID,
		"source":  source,
		"target":  target,
	}

	return d.Emit(EventTypeDeploymentStarted, "deployment-engine", edgeID, payload)
}

// PublishDeploymentCompleted publishes a deployment completed event
func (d *DeploymentEventBus) PublishDeploymentCompleted(edgeID string, success bool, message string) error {
	payload := map[string]interface{}{
		"edge_id": edgeID,
		"success": success,
		"message": message,
	}

	eventType := EventTypeDeploymentCompleted
	if !success {
		eventType = EventTypeDeploymentFailed
	}

	return d.Emit(eventType, "deployment-engine", edgeID, payload)
}

// PublishDeploymentCancelled publishes a deployment cancelled event
func (d *DeploymentEventBus) PublishDeploymentCancelled(edgeID string, reason string) error {
	payload := map[string]interface{}{
		"edge_id": edgeID,
		"reason":  reason,
	}

	return d.Emit(EventTypeDeploymentCancelled, "deployment-engine", edgeID, payload)
}

// SubscribeToDeploymentEvents subscribes to all deployment-related events with a single handler
func (d *DeploymentEventBus) SubscribeToDeploymentEvents(handler EventHandler) {
	d.Subscribe(EventTypeDeploymentStatusChanged, handler)
	d.Subscribe(EventTypeDeploymentProgressUpdate, handler)
	d.Subscribe(EventTypeDeploymentStarted, handler)
	d.Subscribe(EventTypeDeploymentCompleted, handler)
	d.Subscribe(EventTypeDeploymentFailed, handler)
	d.Subscribe(EventTypeDeploymentCancelled, handler)
}

// SubscribeToStatusChanges subscribes specifically to status change events
func (d *DeploymentEventBus) SubscribeToStatusChanges(handler func(edgeID string, oldStatus, newStatus contracts.DeploymentStatus, message string)) {
	d.Subscribe(EventTypeDeploymentStatusChanged, func(event Event) error {
		edgeID, _ := event.Payload["edge_id"].(string)
		oldStatusStr, _ := event.Payload["old_status"].(string)
		newStatusStr, _ := event.Payload["new_status"].(string)
		message, _ := event.Payload["message"].(string)

		oldStatus := contracts.DeploymentStatus(oldStatusStr)
		newStatus := contracts.DeploymentStatus(newStatusStr)

		handler(edgeID, oldStatus, newStatus, message)
		return nil
	})
}

// SubscribeToProgressUpdates subscribes specifically to progress update events
func (d *DeploymentEventBus) SubscribeToProgressUpdates(handler func(edgeID string, progress float64, message string)) {
	d.Subscribe(EventTypeDeploymentProgressUpdate, func(event Event) error {
		edgeID, _ := event.Payload["edge_id"].(string)
		progress, _ := event.Payload["progress"].(float64)
		message, _ := event.Payload["message"].(string)

		handler(edgeID, progress, message)
		return nil
	})
}

// GetEventMetrics returns metrics about deployment events
func (d *DeploymentEventBus) GetEventMetrics() map[string]interface{} {
	// This would typically integrate with a metrics system
	// For now, return a placeholder structure
	return map[string]interface{}{
		"total_events": 0,
		"event_types": map[string]int{
			string(EventTypeDeploymentStatusChanged):  0,
			string(EventTypeDeploymentProgressUpdate): 0,
			string(EventTypeDeploymentStarted):        0,
			string(EventTypeDeploymentCompleted):      0,
			string(EventTypeDeploymentFailed):         0,
			string(EventTypeDeploymentCancelled):      0,
		},
	}
}
