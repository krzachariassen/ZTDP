package environment

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// EnvironmentService - ALL domain logic for environments (business logic, AI extraction, persistence)
type EnvironmentService struct {
	Graph      *graph.GlobalGraph
	aiProvider ai.AIProvider
	eventBus   *events.EventBus
	logger     *logging.Logger
	config     *EnvironmentConfig
}

// EnvironmentDomainParams represents extracted parameters from AI parsing
type EnvironmentDomainParams struct {
	Action          string  `json:"action"`
	EnvironmentName string  `json:"environment_name"`
	Owner           string  `json:"owner"`
	Description     string  `json:"description"`
	EnvType         string  `json:"env_type"`
	Confidence      float64 `json:"confidence"`
	Clarification   string  `json:"clarification"`
}

func NewEnvironmentService(g *graph.GlobalGraph) *EnvironmentService {
	return &EnvironmentService{
		Graph:      g,
		aiProvider: nil, // Will be set when AI-native methods are used
		eventBus:   nil, // Will be set when needed
		logger:     logging.GetLogger().ForComponent("environment-domain"),
		config:     DefaultEnvironmentConfig(),
	}
}

// NewAIEnvironmentService creates AI-native environment service with all dependencies
func NewAIEnvironmentService(g *graph.GlobalGraph, aiProvider ai.AIProvider, eventBus *events.EventBus) *EnvironmentService {
	return &EnvironmentService{
		Graph:      g,
		aiProvider: aiProvider,
		eventBus:   eventBus,
		logger:     logging.GetLogger().ForComponent("environment-domain"),
		config:     DefaultEnvironmentConfig(),
	}
}

// HandleEnvironmentEvent - AI-native event handler (ALL domain logic)
func (s *EnvironmentService) HandleEnvironmentEvent(ctx context.Context, event *events.Event, userMessage string) (*events.Event, error) {
	s.logger.Info("🌍 Environment domain processing: %s", userMessage)

	// Extract intent and parameters using AI (domain owns this)
	params, err := s.ExtractEnvironmentParameters(ctx, userMessage)
	if err != nil {
		return s.createErrorResponse(event, fmt.Sprintf("Failed to extract parameters: %v", err)), nil
	}

	s.logger.Info("🤖 AI extracted - action: %s, env: %s, owner: %s, confidence: %.2f",
		params.Action, params.EnvironmentName, params.Owner, params.Confidence)

	// Check confidence level
	if params.Confidence < 0.7 {
		return s.createClarificationResponse(event, params.Clarification), nil
	}

	// Route to appropriate handler based on AI-extracted action
	switch params.Action {
	case "create":
		return s.handleCreateEnvironment(ctx, event, params)
	case "list":
		return s.handleListEnvironments(ctx, event, params)
	case "get", "show":
		return s.handleGetEnvironment(ctx, event, params)
	default:
		return s.createErrorResponse(event, fmt.Sprintf("Unknown action: %s", params.Action)), nil
	}
}

// ExtractEnvironmentParameters - Environment domain owns AI extraction
func (s *EnvironmentService) ExtractEnvironmentParameters(ctx context.Context, userMessage string) (*EnvironmentDomainParams, error) {
	if s.aiProvider == nil {
		return nil, fmt.Errorf("AI provider not available")
	}

	systemPrompt := fmt.Sprintf(`You are an environment management assistant. Parse the user's request and extract the action and parameters.

Available actions: list, create, update, delete, show, get

IMPORTANT: Environment Name Inference Rules:
%s

Approved environment names: %s

ALWAYS try to infer the canonical environment name from context. Look for patterns like:
- "staging environment" -> "staging"
- "production env" -> "production"  
- "dev environment" -> "development"
- Use the approved environment names list above as your reference

Response format must be valid JSON:
{
  "action": "list|create|update|delete|show|get",
  "environment_name": "canonical environment name (infer from context using approved names)",
  "owner": "owner if specified or null",
  "description": "description if specified or null", 
  "env_type": "development|staging|production|test if specified or null",
  "confidence": 0.0-1.0,
  "clarification": "what to ask if confidence < 0.7"
}

Examples:
- "list environments" -> {"action": "list", "confidence": 0.9}
- "create environment dev" -> {"action": "create", "environment_name": "development", "confidence": 0.9}
- "Create a development environment called dev owned by platform-team for development work" -> {"action": "create", "environment_name": "development", "owner": "platform-team", "description": "for development work", "env_type": "development", "confidence": 0.95}
- "Create a staging environment for testing" -> {"action": "create", "environment_name": "staging", "description": "for testing", "env_type": "staging", "confidence": 0.9}
- "Create a production environment with strict policies" -> {"action": "create", "environment_name": "production", "description": "with strict policies", "env_type": "production", "confidence": 0.9}`,
		s.config.GetEnvironmentExamples(), s.config.GetApprovedEnvironmentsList())

	response, err := s.aiProvider.CallAI(ctx, systemPrompt, userMessage)
	if err != nil {
		return nil, fmt.Errorf("AI extraction failed: %w", err)
	}

	var params EnvironmentDomainParams
	if err := json.Unmarshal([]byte(response), &params); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Post-process: resolve environment name using our configuration
	if params.EnvironmentName != "" {
		params.EnvironmentName = s.config.ResolveEnvironmentName(params.EnvironmentName)
	}

	// Emit extraction completed event
	if s.eventBus != nil {
		s.eventBus.Emit("environment", "environment.parameter_extraction.completed", "environment", map[string]interface{}{
			"user_message": userMessage,
			"parameters":   &params,
			"success":      true,
		})
	}

	return &params, nil
}

// AI-native action handlers
func (s *EnvironmentService) handleCreateEnvironment(ctx context.Context, event *events.Event, params *EnvironmentDomainParams) (*events.Event, error) {
	// Validate required fields
	if params.EnvironmentName == "" {
		return s.createErrorResponse(event, "environment name is required"), nil
	}

	// Create environment using domain logic
	envContract := contracts.EnvironmentContract{
		Metadata: contracts.Metadata{
			Name: params.EnvironmentName,
		},
		Spec: contracts.EnvironmentSpec{
			Description: params.Description,
		},
	}

	if err := s.CreateEnvironment(envContract); err != nil {
		return s.createErrorResponse(event, fmt.Sprintf("Failed to create environment: %v", err)), nil
	}

	// Emit domain event
	if s.eventBus != nil {
		s.eventBus.Emit("environment", "environment.created", params.EnvironmentName, map[string]interface{}{
			"environment_name": params.EnvironmentName,
			"owner":            params.Owner,
			"correlation_id":   event.Payload["correlation_id"],
		})
	}

	return &events.Event{
		ID:        fmt.Sprintf("environment-response-%d", time.Now().UnixNano()),
		Type:      events.EventTypeResponse,
		Subject:   "environment.response",
		Source:    "environment-agent",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"status":           "success",
			"message":          fmt.Sprintf("Environment '%s' created successfully", params.EnvironmentName),
			"environment_name": params.EnvironmentName,
			"owner":            params.Owner,
			"correlation_id":   event.Payload["correlation_id"],
			"request_id":       event.Payload["request_id"],
		},
	}, nil
}

func (s *EnvironmentService) handleListEnvironments(ctx context.Context, event *events.Event, params *EnvironmentDomainParams) (*events.Event, error) {
	environments, err := s.ListEnvironmentsAsData()
	if err != nil {
		return s.createErrorResponse(event, fmt.Sprintf("Failed to list environments: %v", err)), nil
	}

	return &events.Event{
		ID:        fmt.Sprintf("environment-response-%d", time.Now().UnixNano()),
		Type:      events.EventTypeResponse,
		Subject:   "environment.response",
		Source:    "environment-agent",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"status":         "success",
			"environments":   environments,
			"count":          len(environments),
			"correlation_id": event.Payload["correlation_id"],
			"request_id":     event.Payload["request_id"],
		},
	}, nil
}

func (s *EnvironmentService) handleGetEnvironment(ctx context.Context, event *events.Event, params *EnvironmentDomainParams) (*events.Event, error) {
	if params.EnvironmentName == "" {
		return s.createErrorResponse(event, "environment name is required"), nil
	}

	// Get environment by name
	node, err := s.Graph.GetNode(params.EnvironmentName)
	if err != nil || node == nil {
		return s.createErrorResponse(event, fmt.Sprintf("Environment '%s' not found", params.EnvironmentName)), nil
	}

	return &events.Event{
		ID:        fmt.Sprintf("environment-response-%d", time.Now().UnixNano()),
		Type:      events.EventTypeResponse,
		Subject:   "environment.response",
		Source:    "environment-agent",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"status":         "success",
			"environment":    node.Metadata,
			"correlation_id": event.Payload["correlation_id"],
			"request_id":     event.Payload["request_id"],
		},
	}, nil
}

// Helper methods for responses
func (s *EnvironmentService) createErrorResponse(originalEvent *events.Event, errorMsg string) *events.Event {
	return &events.Event{
		ID:        fmt.Sprintf("environment-error-%d", time.Now().UnixNano()),
		Type:      events.EventTypeResponse,
		Subject:   "environment.error",
		Source:    "environment-agent",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"status":         "error",
			"message":        errorMsg,
			"correlation_id": originalEvent.Payload["correlation_id"],
			"request_id":     originalEvent.Payload["request_id"],
		},
	}
}

func (s *EnvironmentService) createClarificationResponse(originalEvent *events.Event, clarification string) *events.Event {
	return &events.Event{
		ID:        fmt.Sprintf("environment-clarification-%d", time.Now().UnixNano()),
		Type:      events.EventTypeResponse,
		Subject:   "environment.clarification",
		Source:    "environment-agent",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"status":         "clarification",
			"message":        clarification,
			"correlation_id": originalEvent.Payload["correlation_id"],
			"request_id":     originalEvent.Payload["request_id"],
		},
	}
}

// Traditional domain methods (CRUD operations)

// CreateEnvironment validates and creates an environment node in the graph
func (s *EnvironmentService) CreateEnvironment(env contracts.EnvironmentContract) error {
	node, err := graph.ResolveContract(env)
	if err != nil {
		return err
	}
	s.Graph.AddNode(node)
	return s.Graph.Save()
}

// CreateEnvironmentFromData creates an environment from raw data
func (s *EnvironmentService) CreateEnvironmentFromData(envData map[string]interface{}) (map[string]interface{}, error) {
	var env contracts.EnvironmentContract
	if err := mapToContract(envData, &env); err != nil {
		return nil, fmt.Errorf("failed to convert data to contract: %w", err)
	}

	if err := s.CreateEnvironment(env); err != nil {
		return nil, err
	}

	return contractToMap(env), nil
}

// ListEnvironments returns all environments as contracts
func (s *EnvironmentService) ListEnvironments() ([]contracts.EnvironmentContract, error) {
	nodes, err := s.Graph.Nodes()
	if err != nil {
		return nil, err
	}

	var environments []contracts.EnvironmentContract
	for _, node := range nodes {
		if node.Kind == "environment" {
			var env contracts.EnvironmentContract
			if node.Metadata != nil {
				if err := mapToContract(node.Metadata, &env); err == nil {
					environments = append(environments, env)
				}
			}
		}
	}
	return environments, nil
}

// ListEnvironmentsAsData returns all environments as data maps
func (s *EnvironmentService) ListEnvironmentsAsData() ([]map[string]interface{}, error) {
	environments, err := s.ListEnvironments()
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for _, env := range environments {
		result = append(result, contractToMap(env))
	}
	return result, nil
}

// GetEnvironments returns environments as maps (for compatibility)
func (s *EnvironmentService) GetEnvironments() ([]map[string]interface{}, error) {
	return s.ListEnvironmentsAsData()
}

// LinkAppAllowedInEnvironment creates a relationship between an app and environment
func (s *EnvironmentService) LinkAppAllowedInEnvironment(appName, envName string) error {
	appNode, err := s.Graph.GetNode(appName)
	if err != nil || appNode == nil {
		return fmt.Errorf("application '%s' not found", appName)
	}

	envNode, err := s.Graph.GetNode(envName)
	if err != nil || envNode == nil {
		return fmt.Errorf("environment '%s' not found", envName)
	}

	return s.Graph.AddEdge(appNode.ID, envNode.ID, "allowed_in")
}

// ListAllowedEnvironments returns environments where the app is allowed
func (s *EnvironmentService) ListAllowedEnvironments(appName string) ([]contracts.EnvironmentContract, error) {
	appNode, err := s.Graph.GetNode(appName)
	if err != nil || appNode == nil {
		return nil, fmt.Errorf("application '%s' not found", appName)
	}

	// Get all edges from this app
	edges, err := s.Graph.Edges()
	if err != nil {
		return nil, fmt.Errorf("failed to get edges: %w", err)
	}

	var environments []contracts.EnvironmentContract

	// Find edges from this app with "allowed_in" relationship
	if appEdges, exists := edges[appNode.ID]; exists {
		for _, edge := range appEdges {
			if edge.Type == "allowed_in" {
				// Get the target environment node
				envNode, err := s.Graph.GetNode(edge.To)
				if err == nil && envNode != nil && envNode.Kind == "environment" {
					var env contracts.EnvironmentContract
					if envNode.Metadata != nil {
						if err := mapToContract(envNode.Metadata, &env); err == nil {
							environments = append(environments, env)
						}
					}
				}
			}
		}
	}

	return environments, nil
}

// EnvironmentConfig holds configuration for environment management including approved names and aliases
type EnvironmentConfig struct {
	ApprovedEnvironments map[string][]string // Map of canonical name to aliases
}

// DefaultEnvironmentConfig returns the default configuration with common environment names and aliases
func DefaultEnvironmentConfig() *EnvironmentConfig {
	return &EnvironmentConfig{
		ApprovedEnvironments: map[string][]string{
			"development": {"dev", "develop", "development"},
			"staging":     {"stage", "staging", "stg"},
			"production":  {"prod", "production", "live"},
			"test":        {"test", "testing", "qa"},
			"preprod":     {"preprod", "pre-prod", "preproduction"},
			"sandbox":     {"sandbox", "sbx", "demo"},
			"local":       {"local", "localhost"},
		},
	}
}

// GetEnvironmentExamples generates dynamic examples for the AI prompt based on approved environments
func (c *EnvironmentConfig) GetEnvironmentExamples() string {
	var examples []string

	for canonical, aliases := range c.ApprovedEnvironments {
		// Create example for each alias pointing to canonical name
		for _, alias := range aliases {
			examples = append(examples, fmt.Sprintf(`- "%s environment" -> environment_name: "%s"`, alias, canonical))
		}
	}

	return strings.Join(examples, "\n")
}

// GetApprovedEnvironmentsList generates a list of all approved environment names for the AI prompt
func (c *EnvironmentConfig) GetApprovedEnvironmentsList() string {
	var allNames []string

	for canonical, aliases := range c.ApprovedEnvironments {
		allNames = append(allNames, canonical)
		allNames = append(allNames, aliases...)
	}

	return strings.Join(allNames, ", ")
}

// ResolveEnvironmentName resolves an alias to its canonical name
func (c *EnvironmentConfig) ResolveEnvironmentName(input string) string {
	inputLower := strings.ToLower(strings.TrimSpace(input))

	for canonical, aliases := range c.ApprovedEnvironments {
		for _, alias := range aliases {
			if strings.ToLower(alias) == inputLower {
				return canonical
			}
		}
	}

	// If no match found, return the input as-is (might be a custom environment)
	return input
}

// Helper functions for converting between contracts and maps
func mapToContract(data map[string]interface{}, contract interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, contract)
}

func contractToMap(contract interface{}) map[string]interface{} {
	jsonData, _ := json.Marshal(contract)
	var result map[string]interface{}
	json.Unmarshal(jsonData, &result)
	return result
}
