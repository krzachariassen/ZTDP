package events

// GlobalGraphEmitter is a global reference to the event-emitting graph store wrapper.
// It is set by the control plane or API main and used by event consumers.
var GlobalGraphEmitter interface{}

// SetGraphEmitter sets the global graph emitter (should be a *graph.GraphEventEmitter).
func SetGraphEmitter(emitter interface{}) {
	GlobalGraphEmitter = emitter
}
