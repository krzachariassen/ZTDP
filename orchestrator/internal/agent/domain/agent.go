package domain

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

// AgentStatus represents the current status of an agent
type AgentStatus string

const (
	AgentStatusOnline      AgentStatus = "online"
	AgentStatusOffline     AgentStatus = "offline"
	AgentStatusBusy        AgentStatus = "busy"
	AgentStatusMaintenance AgentStatus = "maintenance"
)

// AgentCapability represents a specific capability an agent provides
type AgentCapability struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Parameters  map[string]string `json:"parameters,omitempty"`
}

// Agent represents an agent in the system with full type safety and validation
type Agent struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Status       AgentStatus       `json:"status"`
	Capabilities []AgentCapability `json:"capabilities"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	LastSeen     time.Time         `json:"last_seen"`
}

// Agent business rules and validation
var (
	ErrInvalidAgentID          = errors.New("agent ID must be non-empty and contain only alphanumeric characters, hyphens, and underscores")
	ErrInvalidAgentName        = errors.New("agent name must be non-empty and less than 100 characters")
	ErrInvalidAgentDescription = errors.New("agent description must be less than 500 characters")
	ErrInvalidStatus           = errors.New("invalid agent status")
	ErrNoCapabilities          = errors.New("agent must have at least one capability")
	ErrInvalidCapability       = errors.New("capability name must be non-empty")
)

// agentIDPattern defines valid agent ID format
var agentIDPattern = regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)

// NewAgent creates a new agent with validation
func NewAgent(id, name, description string, capabilities []AgentCapability) (*Agent, error) {
	agent := &Agent{
		ID:           id,
		Name:         name,
		Description:  description,
		Status:       AgentStatusOffline, // Default to offline
		Capabilities: capabilities,
		Metadata:     make(map[string]string),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastSeen:     time.Now(),
	}

	if err := agent.Validate(); err != nil {
		return nil, err
	}

	return agent, nil
}

// Validate enforces business rules for agent data
func (a *Agent) Validate() error {
	// Validate ID
	if a.ID == "" {
		return ErrInvalidAgentID
	}
	if !agentIDPattern.MatchString(a.ID) {
		return ErrInvalidAgentID
	}

	// Validate name
	if a.Name == "" || len(a.Name) > 100 {
		return ErrInvalidAgentName
	}

	// Validate description
	if len(a.Description) > 500 {
		return ErrInvalidAgentDescription
	}

	// Validate status
	if !a.Status.IsValid() {
		return ErrInvalidStatus
	}

	// Validate capabilities
	if len(a.Capabilities) == 0 {
		return ErrNoCapabilities
	}

	for _, capability := range a.Capabilities {
		if err := capability.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// IsValid checks if the agent status is valid
func (s AgentStatus) IsValid() bool {
	return s == AgentStatusOnline || s == AgentStatusOffline ||
		s == AgentStatusBusy || s == AgentStatusMaintenance
}

// Validate enforces business rules for capabilities
func (c *AgentCapability) Validate() error {
	if c.Name == "" {
		return ErrInvalidCapability
	}
	return nil
}

// UpdateStatus updates the agent status with validation
func (a *Agent) UpdateStatus(status AgentStatus) error {
	if !status.IsValid() {
		return ErrInvalidStatus
	}

	a.Status = status
	a.UpdatedAt = time.Now()

	// Update last seen time when coming online
	if status == AgentStatusOnline {
		a.LastSeen = time.Now()
	}

	return nil
}

// UpdateLastSeen updates the last seen timestamp
func (a *Agent) UpdateLastSeen() {
	a.LastSeen = time.Now()
	a.UpdatedAt = time.Now()
}

// HasCapability checks if the agent has a specific capability
func (a *Agent) HasCapability(capabilityName string) bool {
	for _, capability := range a.Capabilities {
		if capability.Name == capabilityName {
			return true
		}
	}
	return false
}

// IsAvailable checks if the agent is available for work
func (a *Agent) IsAvailable() bool {
	return a.Status == AgentStatusOnline
}

// IsHealthy checks if the agent is healthy (online/busy and recently seen)
func (a *Agent) IsHealthy() bool {
	// Agent must be online or busy to be considered healthy
	if a.Status != AgentStatusOnline && a.Status != AgentStatusBusy {
		return false
	}

	// Agent must have been seen within the last 10 minutes
	healthyThreshold := 10 * time.Minute
	return time.Since(a.LastSeen) <= healthyThreshold
}

// ToMap converts the agent to a map for storage (legacy compatibility)
func (a *Agent) ToMap() map[string]interface{} {
	capabilities := make([]map[string]interface{}, len(a.Capabilities))
	for i, cap := range a.Capabilities {
		capabilities[i] = map[string]interface{}{
			"name":        cap.Name,
			"description": cap.Description,
			"parameters":  cap.Parameters,
		}
	}

	return map[string]interface{}{
		"id":           a.ID,
		"name":         a.Name,
		"description":  a.Description,
		"status":       string(a.Status),
		"capabilities": capabilities,
		"metadata":     a.Metadata,
		"created_at":   a.CreatedAt.Format(time.RFC3339),
		"updated_at":   a.UpdatedAt.Format(time.RFC3339),
		"last_seen":    a.LastSeen.Format(time.RFC3339),
	}
}

// FromMap creates an agent from a map (legacy compatibility)
func AgentFromMap(data map[string]interface{}) (*Agent, error) {
	agent := &Agent{
		Metadata: make(map[string]string),
	}

	// Extract basic fields with type safety
	if id, ok := data["id"].(string); ok {
		agent.ID = id
	} else {
		return nil, errors.New("missing or invalid agent ID")
	}

	if name, ok := data["name"].(string); ok {
		agent.Name = name
	} else {
		return nil, errors.New("missing or invalid agent name")
	}

	if description, ok := data["description"].(string); ok {
		agent.Description = description
	}

	if statusStr, ok := data["status"].(string); ok {
		agent.Status = AgentStatus(statusStr)
	} else {
		agent.Status = AgentStatusOffline
	}

	// Parse capabilities
	if capsData, ok := data["capabilities"].([]interface{}); ok {
		for _, capData := range capsData {
			if capMap, ok := capData.(map[string]interface{}); ok {
				capability := AgentCapability{}
				if name, ok := capMap["name"].(string); ok {
					capability.Name = name
				}
				if desc, ok := capMap["description"].(string); ok {
					capability.Description = desc
				}
				if params, ok := capMap["parameters"].(map[string]string); ok {
					capability.Parameters = params
				}
				agent.Capabilities = append(agent.Capabilities, capability)
			}
		}
	}

	// Parse timestamps
	if createdStr, ok := data["created_at"].(string); ok {
		if createdAt, err := time.Parse(time.RFC3339, createdStr); err == nil {
			agent.CreatedAt = createdAt
		}
	}
	if updatedStr, ok := data["updated_at"].(string); ok {
		if updatedAt, err := time.Parse(time.RFC3339, updatedStr); err == nil {
			agent.UpdatedAt = updatedAt
		}
	}
	if lastSeenStr, ok := data["last_seen"].(string); ok {
		if lastSeen, err := time.Parse(time.RFC3339, lastSeenStr); err == nil {
			agent.LastSeen = lastSeen
		}
	}

	// Set defaults for missing timestamps
	now := time.Now()
	if agent.CreatedAt.IsZero() {
		agent.CreatedAt = now
	}
	if agent.UpdatedAt.IsZero() {
		agent.UpdatedAt = now
	}
	if agent.LastSeen.IsZero() {
		agent.LastSeen = now
	}

	// Validate the constructed agent
	if err := agent.Validate(); err != nil {
		return nil, fmt.Errorf("invalid agent data: %w", err)
	}

	return agent, nil
}
