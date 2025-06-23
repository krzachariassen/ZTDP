package types

import "time"

// Agent represents an agent in the orchestrator
type Agent struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         string            `json:"type,omitempty"`
	Status       string            `json:"status"`
	Capabilities []string          `json:"capabilities"`
	Endpoint     string            `json:"endpoint,omitempty"`
	LastSeen     time.Time         `json:"last_seen,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// AgentStatus constants
const (
	AgentStatusActive   = "active"
	AgentStatusInactive = "inactive"
	AgentStatusError    = "error"
)
