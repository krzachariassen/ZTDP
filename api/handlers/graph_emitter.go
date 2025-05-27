package handlers

import (
	"github.com/krzachariassen/ZTDP/internal/graph"
)

var (
	// GlobalGraphEmitter is the event-emitting graph store wrapper
	GlobalGraphEmitter *graph.GraphEventEmitter
)

// SetGraphEmitter sets the global graph emitter
func SetGraphEmitter(emitter *graph.GraphEventEmitter) {
	GlobalGraphEmitter = emitter
}
