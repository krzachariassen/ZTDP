package ai

import (
	"context"

	"github.com/ztdp/orchestrator/internal/types"
)

// ConversationalResponse represents the response structure for AI orchestrator interactions
type ConversationalResponse struct {
	Message    string                 `json:"message"`
	Answer     string                 `json:"answer,omitempty"`
	Intent     string                 `json:"intent,omitempty"`
	Actions    []Action               `json:"actions,omitempty"`
	Insights   []string               `json:"insights,omitempty"`
	Confidence float64                `json:"confidence,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// Action represents an action taken by the orchestrator
type Action struct {
	Type        string                 `json:"type"`
	Target      string                 `json:"target,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Description string                 `json:"description,omitempty"`
}

// AIProvider defines the interface for AI services (OpenAI, Anthropic, etc.)
type AIProvider interface {
	CallAI(ctx context.Context, systemPrompt, userPrompt string) (string, error)
	GetProviderInfo() *ProviderInfo
}

// ProviderInfo contains information about the AI provider
type ProviderInfo struct {
	Name    string `json:"name"`
	Model   string `json:"model"`
	Version string `json:"version"`
}

// OrchestratorService defines the interface for orchestration operations
type OrchestratorService interface {
	PlanWorkflow(ctx context.Context, request *types.WorkflowRequest) (*types.WorkflowPlan, error)
	ExecuteWorkflow(ctx context.Context, request *types.WorkflowRequest) (*types.WorkflowExecution, error)
	GetWorkflowExecution(ctx context.Context, executionID string) (*types.WorkflowExecution, error)
}

// RegistryService defines the interface for agent registry operations
type RegistryService interface {
	RegisterAgent(ctx context.Context, agent *types.Agent) error
	GetAgentsByCapability(ctx context.Context, capability string) ([]*types.Agent, error)
	UpdateAgentStatus(ctx context.Context, agentID, status string) error
	GetActiveAgents(ctx context.Context) ([]*types.Agent, error)
}

// Logger defines the interface for logging
type Logger interface {
	Info(msg string, keyvals ...interface{})
	Debug(msg string, keyvals ...interface{})
	Error(msg string, err error, keyvals ...interface{})
}
