package common

// EventNodeView defines a lightweight representation of a graph node for events
type EventNodeView struct {
	ID       string                 `json:"id"`
	Kind     string                 `json:"kind"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Spec     map[string]interface{} `json:"spec,omitempty"`
}

// EventEdgeView defines a lightweight representation of a graph edge for events
type EventEdgeView struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

// GraphEvent defines standard event information for graph operations
type GraphEvent struct {
	EventType   string                 `json:"event_type"`
	Environment string                 `json:"environment"`
	NodeID      string                 `json:"node_id,omitempty"`
	Node        *EventNodeView         `json:"node,omitempty"`
	FromID      string                 `json:"from_id,omitempty"`
	ToID        string                 `json:"to_id,omitempty"`
	EdgeType    string                 `json:"edge_type,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}
