package agents

import (
	"context"
	"time"

	"github.com/krzachariassen/ZTDP/internal/events"
)

// BaseAgent provides common functionality for all AI agents
type BaseAgent struct {
	ID           string
	Type         string
	Version      string
	Capabilities []AgentCapability
	Status       string
	EventBus     *events.EventBus
	Registry     AgentRegistry

	// Agent state
	StartTime    time.Time
	LastActivity time.Time
}

// NewBaseAgent creates a new base agent
func NewBaseAgent(id, agentType, version string, capabilities []AgentCapability) *BaseAgent {
	return &BaseAgent{
		ID:           id,
		Type:         agentType,
		Version:      version,
		Capabilities: capabilities,
		Status:       "stopped",
		StartTime:    time.Now(),
		LastActivity: time.Now(),
	}
}

// GetStatus implements AgentInterface
func (b *BaseAgent) GetStatus() AgentStatus {
	return AgentStatus{
		ID:           b.ID,
		Type:         b.Type,
		Status:       b.Status,
		LastActivity: b.LastActivity,
		LoadFactor:   0.0, // Base implementation
		Version:      b.Version,
		Metadata: map[string]interface{}{
			"start_time": b.StartTime,
		},
	}
}

// GetCapabilities implements AgentInterface
func (b *BaseAgent) GetCapabilities() []AgentCapability {
	return b.Capabilities
}

// Start implements AgentInterface
func (b *BaseAgent) Start(ctx context.Context) error {
	b.Status = "running"
	b.StartTime = time.Now()
	b.LastActivity = time.Now()

	// Register with registry if available
	if b.Registry != nil {
		return b.Registry.RegisterAgent(ctx, b)
	}

	return nil
}

// Stop implements AgentInterface
func (b *BaseAgent) Stop(ctx context.Context) error {
	b.Status = "stopped"

	// Unregister from registry if available
	if b.Registry != nil {
		return b.Registry.UnregisterAgent(ctx, b.ID)
	}

	return nil
}

// Health implements AgentInterface
func (b *BaseAgent) Health() HealthStatus {
	healthy := b.Status == "running"
	status := "healthy"
	if !healthy {
		status = "unhealthy"
	}

	return HealthStatus{
		Healthy:   healthy,
		Status:    status,
		Message:   "Base agent health check",
		CheckedAt: time.Now(),
		Checks: map[string]interface{}{
			"status":        b.Status,
			"uptime":        time.Since(b.StartTime).String(),
			"last_activity": b.LastActivity,
		},
	}
}

// ProcessEvent must be implemented by concrete agents
func (b *BaseAgent) ProcessEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	// Update activity timestamp
	b.LastActivity = time.Now()

	// Base implementation returns a generic response
	return &events.Event{
		Type:    events.EventTypeResponse,
		Source:  b.ID,
		Subject: "base_response",
		Payload: map[string]interface{}{
			"message":    "Base agent received event but no specific handler implemented",
			"event_type": event.Type,
		},
	}, nil
}

// UpdateCapabilities updates agent capabilities
func (b *BaseAgent) UpdateCapabilities(capabilities []AgentCapability) {
	b.Capabilities = capabilities
	b.LastActivity = time.Now()
}

// SetEventBus sets the event bus for the agent
func (b *BaseAgent) SetEventBus(eventBus *events.EventBus) {
	b.EventBus = eventBus
}

// SetRegistry sets the agent registry
func (b *BaseAgent) SetRegistry(registry AgentRegistry) {
	b.Registry = registry
}
