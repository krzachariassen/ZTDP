package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
	"github.com/krzachariassen/ZTDP/internal/resources"
)

// ServiceService - ALL domain logic for services (business logic, AI extraction, persistence)
type ServiceService struct {
	Graph      *graph.GlobalGraph
	aiProvider ai.AIProvider
	eventBus   *events.EventBus
	logger     *logging.Logger
}

// ServiceParams represents extracted parameters from AI parsing
type ServiceDomainParams struct {
	Action          string  `json:"action"`
	ServiceName     string  `json:"service_name"`
	ApplicationName string  `json:"application_name"`
	Port            int     `json:"port"`
	Public          bool    `json:"public"`
	Version         string  `json:"version"`
	Details         string  `json:"details"`
	Confidence      float64 `json:"confidence"`
	Clarification   string  `json:"clarification"`
}

func NewServiceService(g *graph.GlobalGraph) *ServiceService {
	return &ServiceService{
		Graph:      g,
		aiProvider: nil, // Will be set when AI-native methods are used
		eventBus:   nil, // Will be set when needed
		logger:     logging.GetLogger().ForComponent("service-domain"),
	}
}

// NewAIServiceService creates AI-native service with all dependencies
func NewAIServiceService(g *graph.GlobalGraph, aiProvider ai.AIProvider, eventBus *events.EventBus) *ServiceService {
	return &ServiceService{
		Graph:      g,
		aiProvider: aiProvider,
		eventBus:   eventBus,
		logger:     logging.GetLogger().ForComponent("service-domain"),
	}
}

// HandleServiceEvent - AI-native event handler (ALL domain logic)
func (s *ServiceService) HandleServiceEvent(ctx context.Context, event *events.Event, userMessage string) (*events.Event, error) {
	s.logger.Info("🔧 Service domain processing: %s", userMessage)

	// Extract intent and parameters using AI (domain owns this)
	params, err := s.ExtractServiceParameters(ctx, userMessage)
	if err != nil {
		return s.createErrorResponse(event, fmt.Sprintf("Failed to extract parameters: %v", err)), nil
	}

	s.logger.Info("🤖 AI extracted - action: %s, service: %s, app: %s, confidence: %.2f",
		params.Action, params.ServiceName, params.ApplicationName, params.Confidence)

	// Check confidence level
	if params.Confidence < 0.7 {
		return s.createClarificationResponse(event, params.Clarification), nil
	}

	// Route to appropriate handler based on AI-extracted action
	s.logger.Info("🎯 Routing to action handler: %s", params.Action)
	switch params.Action {
	case "create":
		s.logger.Info("🏗️ Calling handleCreateService for service: %s", params.ServiceName)
		return s.handleCreateService(ctx, event, params)
	case "list":
		return s.handleListServices(ctx, event, params)
	case "get", "show":
		return s.handleGetService(ctx, event, params)
	case "version":
		return s.handleVersionService(ctx, event, params)
	default:
		s.logger.Error("❌ Unknown action: %s", params.Action)
		return s.createErrorResponse(event, fmt.Sprintf("Unknown action: %s", params.Action)), nil
	}
}

// ExtractServiceParameters - Service domain owns AI extraction
func (s *ServiceService) ExtractServiceParameters(ctx context.Context, userMessage string) (*ServiceDomainParams, error) {
	if s.aiProvider == nil {
		return nil, fmt.Errorf("AI provider not available")
	}

	systemPrompt := `You are a service management assistant. Parse the user's request and extract the action and parameters.

Available actions: list, create, update, delete, show, get, version

Response format must be valid JSON:
{
  "action": "list|create|update|delete|show|get|version",
  "service_name": "service name if specified or null",
  "application_name": "application name if specified or null",
  "port": 8080,
  "public": true,
  "version": "version if specified or null",
  "details": "any additional context",
  "confidence": 0.0-1.0,
  "clarification": "what to ask if confidence < 0.7"
}

IMPORTANT: port must be a number (integer), not a string. If no port specified, use 0.
IMPORTANT: public must be a boolean (true/false), not a string.

Examples:
- "list services for myapp" -> {"action": "list", "application_name": "myapp", "port": 0, "public": false, "confidence": 0.9}
- "create service api in myapp" -> {"action": "create", "application_name": "myapp", "service_name": "api", "port": 0, "public": false, "confidence": 0.9}
- "create service checkout-api for checkout application on port 8080 that is public facing" -> {"action": "create", "service_name": "checkout-api", "application_name": "checkout", "port": 8080, "public": true, "confidence": 0.95}
- "show me the payment service details" -> {"action": "show", "service_name": "payment", "port": 0, "public": false, "confidence": 0.9}`

	response, err := s.aiProvider.CallAI(ctx, systemPrompt, userMessage)
	if err != nil {
		return nil, fmt.Errorf("AI extraction failed: %w", err)
	}

	var params ServiceDomainParams
	if err := json.Unmarshal([]byte(response), &params); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &params, nil
}

// AI-native action handlers
func (s *ServiceService) handleCreateService(ctx context.Context, event *events.Event, params *ServiceDomainParams) (*events.Event, error) {
	s.logger.Info("🏗️ Starting service creation: service=%s, app=%s", params.ServiceName, params.ApplicationName)

	// Validate required fields
	if params.ServiceName == "" {
		s.logger.Error("❌ Service name is required")
		return s.createErrorResponse(event, "service name is required"), nil
	}
	if params.ApplicationName == "" {
		s.logger.Error("❌ Application name is required")
		return s.createErrorResponse(event, "application name is required"), nil
	}

	// Create service data from AI-extracted parameters
	serviceData := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": params.ServiceName,
		},
		"spec": map[string]interface{}{
			"application": params.ApplicationName,
		},
	}

	if params.Port > 0 {
		serviceData["spec"].(map[string]interface{})["port"] = params.Port
	}
	if params.Public {
		serviceData["spec"].(map[string]interface{})["public"] = true
	}

	// Use existing domain logic
	s.logger.Info("🔧 Creating service in graph with data: %+v", serviceData)
	result, err := s.CreateService(params.ApplicationName, serviceData)
	if err != nil {
		s.logger.Error("❌ Failed to create service in graph: %v", err)
		return s.createErrorResponse(event, fmt.Sprintf("Failed to create service: %v", err)), nil
	}

	s.logger.Info("✅ Service created successfully in graph: %+v", result)

	// Emit domain event
	if s.eventBus != nil {
		s.eventBus.Emit("service", "service.created", params.ServiceName, map[string]interface{}{
			"service_name":     params.ServiceName,
			"application_name": params.ApplicationName,
			"correlation_id":   event.Payload["correlation_id"],
		})
	}

	responseEvent := &events.Event{
		Source:  "service-agent",
		Subject: "service.response",
		Payload: map[string]interface{}{
			"status":         "success",
			"message":        fmt.Sprintf("Service '%s' created successfully", params.ServiceName),
			"service":        result,
			"correlation_id": event.Payload["correlation_id"],
			"request_id":     event.Payload["request_id"],
		},
	}

	s.logger.Info("📤 Returning response event: %+v", responseEvent)
	return responseEvent, nil
}

func (s *ServiceService) handleListServices(ctx context.Context, event *events.Event, params *ServiceDomainParams) (*events.Event, error) {
	appName := params.ApplicationName
	if appName == "" {
		// List all services if no app specified
		appName = ""
	}

	services, err := s.ListServices(appName)
	if err != nil {
		return s.createErrorResponse(event, fmt.Sprintf("Failed to list services: %v", err)), nil
	}

	return &events.Event{
		Source:  "service-agent",
		Subject: "service.response",
		Payload: map[string]interface{}{
			"status":         "success",
			"services":       services,
			"count":          len(services),
			"correlation_id": event.Payload["correlation_id"],
			"request_id":     event.Payload["request_id"],
		},
	}, nil
}

func (s *ServiceService) handleGetService(ctx context.Context, event *events.Event, params *ServiceDomainParams) (*events.Event, error) {
	if params.ServiceName == "" {
		return s.createErrorResponse(event, "service name is required"), nil
	}

	service, err := s.GetService(params.ApplicationName, params.ServiceName)
	if err != nil {
		return s.createErrorResponse(event, fmt.Sprintf("Failed to get service: %v", err)), nil
	}

	return &events.Event{
		Source:  "service-agent",
		Subject: "service.response",
		Payload: map[string]interface{}{
			"status":         "success",
			"service":        service,
			"correlation_id": event.Payload["correlation_id"],
			"request_id":     event.Payload["request_id"],
		},
	}, nil
}

func (s *ServiceService) handleVersionService(ctx context.Context, event *events.Event, params *ServiceDomainParams) (*events.Event, error) {
	if params.ServiceName == "" {
		return s.createErrorResponse(event, "service name is required"), nil
	}

	versions, err := s.ListServiceVersions(params.ServiceName)
	if err != nil {
		return s.createErrorResponse(event, fmt.Sprintf("Failed to get service versions: %v", err)), nil
	}

	return &events.Event{
		Source:  "service-agent",
		Subject: "service.response",
		Payload: map[string]interface{}{
			"status":         "success",
			"service":        params.ServiceName,
			"versions":       versions,
			"count":          len(versions),
			"correlation_id": event.Payload["correlation_id"],
			"request_id":     event.Payload["request_id"],
		},
	}, nil
}

// Helper methods for responses
func (s *ServiceService) createErrorResponse(originalEvent *events.Event, errorMsg string) *events.Event {
	return &events.Event{
		Source:  "service-agent",
		Subject: "service.error",
		Payload: map[string]interface{}{
			"status":         "error",
			"message":        errorMsg,
			"correlation_id": originalEvent.Payload["correlation_id"],
			"request_id":     originalEvent.Payload["request_id"],
		},
	}
}

func (s *ServiceService) createClarificationResponse(originalEvent *events.Event, clarification string) *events.Event {
	return &events.Event{
		Source:  "service-agent",
		Subject: "service.clarification",
		Payload: map[string]interface{}{
			"status":         "clarification",
			"message":        clarification,
			"correlation_id": originalEvent.Payload["correlation_id"],
			"request_id":     originalEvent.Payload["request_id"],
		},
	}
}

// CreateService creates a new service from raw data
func (s *ServiceService) CreateService(appName string, serviceData map[string]interface{}) (map[string]interface{}, error) {
	// Convert raw data to contract internally
	var svc contracts.ServiceContract
	if err := mapToContract(serviceData, &svc); err != nil {
		return nil, err
	}

	// Use existing contract-based logic
	if err := s.createServiceInternal(appName, svc); err != nil {
		return nil, err
	}

	// Return as basic map
	return contractToMap(svc), nil
}

// ListServices returns services as basic maps
func (s *ServiceService) ListServices(appName string) ([]map[string]interface{}, error) {
	// Use existing logic to get contracts
	contracts, err := s.listServicesInternal(appName)
	if err != nil {
		return nil, err
	}

	// Convert to basic maps
	var result []map[string]interface{}
	for _, contract := range contracts {
		result = append(result, contractToMap(contract))
	}
	return result, nil
}

// GetService returns a service as basic map
func (s *ServiceService) GetService(appName, serviceName string) (map[string]interface{}, error) {
	// Use existing logic to get contract
	contract, err := s.getServiceInternal(appName, serviceName)
	if err != nil {
		return nil, err
	}

	return contractToMap(contract), nil
}

// CreateServiceVersion creates a service version from raw data
func (s *ServiceService) CreateServiceVersion(serviceName string, versionData map[string]interface{}) (map[string]interface{}, error) {
	// Convert raw data to contract internally
	var ver contracts.ServiceVersionContract
	if err := mapToContract(versionData, &ver); err != nil {
		return nil, err
	}

	// Use existing contract-based logic
	if err := s.createServiceVersionInternal(serviceName, ver); err != nil {
		return nil, err
	}

	return contractToMap(ver), nil
}

// ListServiceVersions returns service versions as basic maps
func (s *ServiceService) ListServiceVersions(serviceName string) ([]map[string]interface{}, error) {
	// Use existing logic to get contracts
	contracts, err := s.listServiceVersionsInternal(serviceName)
	if err != nil {
		return nil, err
	}

	// Convert to basic maps
	var result []map[string]interface{}
	for _, contract := range contracts {
		result = append(result, contractToMap(contract))
	}
	return result, nil
}

// CreateServiceFromContract creates service from contract with context support
// This method supports contract-driven AI operations while maintaining business logic
func (s *ServiceService) CreateServiceFromContract(ctx context.Context, svc *contracts.ServiceContract) (interface{}, error) {
	if err := svc.Validate(); err != nil {
		return nil, err
	}

	node, err := graph.ResolveContract(*svc)
	if err != nil {
		return nil, err
	}

	s.Graph.AddNode(node)
	s.Graph.AddEdge(svc.Spec.Application, svc.Metadata.Name, "owns")

	if err := s.Graph.Save(); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":        svc.Metadata.Name,
		"status":      "created",
		"application": svc.Spec.Application,
		"port":        svc.Spec.Port,
		"public":      svc.Spec.Public,
	}, nil
}

// Helper functions to convert between contracts and maps
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

// Internal methods that work with contracts (actual existing logic)
func (s *ServiceService) createServiceInternal(appName string, svc contracts.ServiceContract) error {
	// Auto-populate application from URL parameter to eliminate redundant validation
	svc.Spec.Application = appName

	// CRITICAL: Validate that the application exists before creating service
	appNode, err := s.Graph.GetNode(appName)
	if err != nil || appNode == nil {
		s.logger.Error("❌ Cannot create service '%s': application '%s' does not exist", svc.Metadata.Name, appName)
		return fmt.Errorf("application '%s' does not exist - create the application first", appName)
	}
	if appNode.Kind != "application" {
		s.logger.Error("❌ Cannot create service '%s': '%s' exists but is not an application (kind: %s)", svc.Metadata.Name, appName, appNode.Kind)
		return fmt.Errorf("'%s' is not an application (kind: %s) - services can only belong to applications", appName, appNode.Kind)
	}

	s.logger.Info("✅ Application '%s' exists - proceeding with service creation", appName)

	if err := svc.Validate(); err != nil {
		return err
	}
	node, err := graph.ResolveContract(svc)
	if err != nil {
		return err
	}

	s.logger.Info("🔗 Adding service node and creating edge: %s -> %s (owns)", appName, svc.Metadata.Name)
	s.Graph.AddNode(node)
	s.Graph.AddEdge(appName, svc.Metadata.Name, "owns")

	if err := s.Graph.Save(); err != nil {
		s.logger.Error("❌ Failed to save graph after service creation: %v", err)
		return fmt.Errorf("failed to persist service creation: %w", err)
	}

	s.logger.Info("✅ Service '%s' created and linked to application '%s'", svc.Metadata.Name, appName)
	return nil
}

func (s *ServiceService) listServicesInternal(appName string) ([]contracts.ServiceContract, error) {
	nodes, err := s.Graph.Nodes()
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	services := []contracts.ServiceContract{}
	for _, node := range nodes {
		if node.Kind == "service" {
			contract, err := resources.LoadNode(node.Kind, node.Spec, contracts.Metadata{
				Name:  node.Metadata["name"].(string),
				Owner: node.Metadata["owner"].(string),
			})
			if err == nil {
				if svc, ok := contract.(*contracts.ServiceContract); ok && svc.Spec.Application == appName {
					services = append(services, *svc)
				}
			}
		}
	}
	return services, nil
}

func (s *ServiceService) getServiceInternal(appName, serviceName string) (contracts.ServiceContract, error) {
	node, err := s.Graph.GetNode(serviceName)
	if err != nil || node == nil || node.Kind != "service" {
		return contracts.ServiceContract{}, errors.New("service not found")
	}

	contract, err := resources.LoadNode(node.Kind, node.Spec, contracts.Metadata{
		Name:  node.Metadata["name"].(string),
		Owner: node.Metadata["owner"].(string),
	})
	if err != nil {
		return contracts.ServiceContract{}, errors.New("invalid service contract")
	}
	svc, ok := contract.(*contracts.ServiceContract)
	if !ok || svc.Spec.Application != appName {
		return contracts.ServiceContract{}, errors.New("service not found for this application")
	}
	return *svc, nil
}

func (s *ServiceService) createServiceVersionInternal(serviceName string, ver contracts.ServiceVersionContract) error {
	if node, err := s.Graph.GetNode(serviceName); err != nil || node == nil {
		return errors.New("service does not exist")
	}

	if ver.Version == "" {
		return errors.New("version is required")
	}

	id := serviceName + ":" + ver.Version
	if existingNode, err := s.Graph.GetNode(id); err == nil && existingNode != nil {
		// Version already exists, return existing
		return nil
	}

	ver.IDValue = id
	ver.Name = serviceName
	ver.CreatedAt = time.Now()

	if err := ver.Validate(); err != nil {
		return err
	}

	node, err := graph.ResolveContract(ver)
	if err != nil {
		return err
	}

	s.Graph.AddNode(node)
	s.Graph.AddEdge(serviceName, id, "has_version")
	return s.Graph.Save()
}

func (s *ServiceService) listServiceVersionsInternal(serviceName string) ([]contracts.ServiceVersionContract, error) {
	edges, err := s.Graph.Edges()
	if err != nil {
		return nil, fmt.Errorf("failed to get service versions: %w", err)
	}

	nodes, err := s.Graph.Nodes()
	if err != nil {
		return nil, fmt.Errorf("failed to get service versions: %w", err)
	}

	versions := []contracts.ServiceVersionContract{}
	for _, edge := range edges[serviceName] {
		if edge.Type == "has_version" {
			if node, ok := nodes[edge.To]; ok && node.Kind == "service_version" {
				var ver contracts.ServiceVersionContract
				b, _ := json.Marshal(node)
				_ = json.Unmarshal(b, &ver)
				versions = append(versions, ver)
			}
		}
	}
	return versions, nil
}
