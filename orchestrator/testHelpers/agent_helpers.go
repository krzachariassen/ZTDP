package testHelpers

import (
	"context"
	"time"

	"github.com/ztdp/orchestrator/internal/agent/domain"
	"github.com/ztdp/orchestrator/internal/logging"
)

// TestLogger returns a no-op logger for testing
func TestLogger() logging.Logger {
	return logging.NewNoOpLogger()
}

// TestContext returns a context for testing (without timeout to avoid leaks in tests)
func TestContext() context.Context {
	return context.Background()
}

// TestAgent creates a test agent with default values for clean architecture tests
func TestAgentForDomain(id, name string, capabilities ...string) *domain.Agent {
	caps := make([]domain.AgentCapability, len(capabilities))
	for i, cap := range capabilities {
		caps[i] = domain.AgentCapability{
			Name:        cap,
			Description: "Test capability: " + cap,
		}
	}

	return &domain.Agent{
		ID:           id,
		Name:         name,
		Description:  "Test agent: " + name,
		Status:       domain.AgentStatusOnline,
		Capabilities: caps,
		Metadata:     map[string]string{"test": "true"},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastSeen:     time.Now(),
	}
}

// AssertAgentEquals checks if two agents are functionally equal (ignoring timestamps)
func AssertAgentEquals(expected, actual *domain.Agent) bool {
	if expected.ID != actual.ID {
		return false
	}
	if expected.Name != actual.Name {
		return false
	}
	if expected.Description != actual.Description {
		return false
	}
	if expected.Status != actual.Status {
		return false
	}
	if len(expected.Capabilities) != len(actual.Capabilities) {
		return false
	}

	// Check capabilities
	for i, cap := range expected.Capabilities {
		if actual.Capabilities[i].Name != cap.Name {
			return false
		}
	}

	// Check metadata
	if len(expected.Metadata) != len(actual.Metadata) {
		return false
	}
	for k, v := range expected.Metadata {
		if actual.Metadata[k] != v {
			return false
		}
	}

	return true
}
