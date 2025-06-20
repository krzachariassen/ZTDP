package application

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
)

// Domain types for application management
type ApplicationEnvironment struct {
	Name        string            `json:"name"`
	Application string            `json:"application"`
	Description string            `json:"description"`
	Status      string            `json:"status"`
	Variables   map[string]string `json:"variables,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type ApplicationService struct {
	Name        string            `json:"name"`
	Application string            `json:"application"`
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Port        int               `json:"port,omitempty"`
	Replicas    int               `json:"replicas"`
	Variables   map[string]string `json:"variables,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type ServiceVersion struct {
	Version      string            `json:"version"`
	Service      string            `json:"service"`
	Application  string            `json:"application"`
	Image        string            `json:"image"`
	Status       string            `json:"status"`
	Variables    map[string]string `json:"variables,omitempty"`
	Dependencies []string          `json:"dependencies,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type ApplicationRelease struct {
	Version       string                 `json:"version"`
	Application   string                 `json:"application"`
	Environment   string                 `json:"environment"`
	Status        string                 `json:"status"`
	Services      map[string]string      `json:"services"` // service -> version mapping
	Configuration map[string]interface{} `json:"configuration,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// Service for application business logic
type Service struct {
	Graph      *graph.GlobalGraph
	aiProvider ai.AIProvider
}

func NewService(g *graph.GlobalGraph, aiProvider ai.AIProvider) *Service {
	return &Service{
		Graph:      g,
		aiProvider: aiProvider,
	}
}

// CreateApplication validates and creates an application node in the graph
func (s *Service) CreateApplication(app contracts.ApplicationContract) error {
	if err := app.Validate(); err != nil {
		return err
	}

	node, _ := graph.ResolveContract(app)
	s.Graph.AddNode(node)

	// Save the graph
	if err := s.Graph.Save(); err != nil {
		return err
	}

	// Emit HIGH-LEVEL BUSINESS EVENT - Simple and clean!
	if events.GlobalEventBus != nil {
		payload := map[string]interface{}{
			"application_name": app.Metadata.Name,
			"description":      app.Spec.Description,
			"owner":            app.Metadata.Owner,
			"tags":             app.Spec.Tags,
		}
		events.GlobalEventBus.Emit(events.EventTypeNotify, "ztdp-platform", "application_created", payload)
	}

	return nil
}

// CreateApplicationFromContract creates application from contract with context support
// This method supports contract-driven AI operations while maintaining business logic
func (s *Service) CreateApplicationFromContract(ctx context.Context, app *contracts.ApplicationContract) (interface{}, error) {
	if err := app.Validate(); err != nil {
		return nil, err
	}

	node, _ := graph.ResolveContract(*app)
	s.Graph.AddNode(node)

	// Save the graph
	if err := s.Graph.Save(); err != nil {
		return nil, err
	}

	// Emit HIGH-LEVEL BUSINESS EVENT
	if events.GlobalEventBus != nil {
		payload := map[string]interface{}{
			"application_name": app.Metadata.Name,
			"description":      app.Spec.Description,
			"owner":            app.Metadata.Owner,
			"tags":             app.Spec.Tags,
		}
		events.GlobalEventBus.Emit(events.EventTypeNotify, "ztdp-platform", "application_created", payload)
	}

	return map[string]interface{}{
		"name":        app.Metadata.Name,
		"status":      "created",
		"description": app.Spec.Description,
		"owner":       app.Metadata.Owner,
		"tags":        app.Spec.Tags,
	}, nil
}

// ListApplications returns all applications in the graph
func (s *Service) ListApplications() ([]contracts.ApplicationContract, error) {
	nodes, err := s.Graph.Nodes()
	if err != nil {
		return nil, err
	}

	apps := []contracts.ApplicationContract{}
	for _, node := range nodes {
		if node.Kind == "application" {
			app := contracts.ApplicationContract{
				Metadata: contracts.Metadata{
					Name:  node.Metadata["name"].(string),
					Owner: node.Metadata["owner"].(string),
				},
				Spec: contracts.ApplicationSpec{},
			}

			// Copy spec data
			if specBytes, err := json.Marshal(node.Spec); err == nil {
				json.Unmarshal(specBytes, &app.Spec)
			}

			apps = append(apps, app)
		}
	}
	return apps, nil
}

// GetApplication returns a specific application by name
func (s *Service) GetApplication(appName string) (*contracts.ApplicationContract, error) {
	node, err := s.Graph.GetNode(appName)
	if err != nil || node == nil || node.Kind != "application" {
		return nil, errors.New("application not found")
	}

	app := contracts.ApplicationContract{
		Metadata: contracts.Metadata{
			Name:  node.Metadata["name"].(string),
			Owner: node.Metadata["owner"].(string),
		},
		Spec: contracts.ApplicationSpec{},
	}

	// Copy spec data
	if specBytes, err := json.Marshal(node.Spec); err == nil {
		json.Unmarshal(specBytes, &app.Spec)
	}

	return &app, nil
}

// UpdateApplication validates and updates an existing application
func (s *Service) UpdateApplication(appName string, app contracts.ApplicationContract) error {
	if app.Metadata.Name != appName {
		return errors.New("application name mismatch")
	}

	if err := app.Validate(); err != nil {
		return err
	}

	node, _ := graph.ResolveContract(app)
	s.Graph.AddNode(node)

	// Save the graph
	if err := s.Graph.Save(); err != nil {
		return err
	}

	// Events are automatically emitted by the graph layer
	return nil
}

// DeleteApplication removes an application from the graph
func (s *Service) DeleteApplication(appName string) error {
	// TODO: Implement proper node deletion when graph supports it
	// For now, we'll update the node to mark it as deleted
	node, err := s.Graph.GetNode(appName)
	if err != nil || node == nil || node.Kind != "application" {
		return errors.New("application not found")
	}

	// Mark node as deleted by updating its metadata
	node.Metadata["deleted"] = true

	// Get current graph and update the node
	currentGraph, err := s.Graph.Graph()
	if err != nil {
		return err
	}

	err = currentGraph.UpdateNode(node)
	if err != nil {
		return err
	}

	// Save the graph
	if err := s.Graph.Save(); err != nil {
		return err
	}

	// Emit deletion event
	if events.GlobalEventBus != nil {
		payload := map[string]interface{}{
			"application_name": appName,
		}
		events.GlobalEventBus.Emit(events.EventTypeNotify, "ztdp-platform", "application_deleted", payload)
	}

	return nil
}

// Environment Management Methods
func (s *Service) CreateApplicationEnvironment(appName, envName, description string) error {
	if appName == "" {
		return fmt.Errorf("application name is required")
	}
	if envName == "" {
		return fmt.Errorf("environment name is required")
	}

	// Check if application exists
	if !s.applicationExists(appName) {
		return fmt.Errorf("application not found: %s", appName)
	}

	// Create environment contract
	envContract := contracts.EnvironmentContract{
		Metadata: contracts.Metadata{
			Name:  fmt.Sprintf("%s-%s", appName, envName),
			Owner: "system",
		},
		Spec: contracts.EnvironmentSpec{
			Description: description,
		},
	}

	// Add to graph
	node, err := graph.ResolveContract(envContract)
	if err != nil {
		return fmt.Errorf("failed to resolve environment contract: %w", err)
	}

	// Add application and environment to node metadata for querying
	node.Metadata["application"] = appName
	node.Metadata["environment"] = envName
	node.Metadata["status"] = "active"

	s.Graph.AddNode(node)
	if err := s.Graph.Save(); err != nil {
		return fmt.Errorf("failed to save environment: %w", err)
	}

	// Emit event
	if events.GlobalEventBus != nil {
		payload := map[string]interface{}{
			"application": appName,
			"environment": envName,
			"description": description,
		}
		events.GlobalEventBus.Emit(events.EventTypeNotify, "ztdp-platform", "application_environment_created", payload)
	}

	return nil
}

func (s *Service) ListApplicationEnvironments(appName string) ([]ApplicationEnvironment, error) {
	if appName == "" {
		return nil, fmt.Errorf("application name is required")
	}

	// Get current graph
	currentGraph, err := s.Graph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	var environments []ApplicationEnvironment

	for _, node := range currentGraph.Nodes {
		if node.Kind == "environment" {
			if appName, ok := node.Metadata["application"].(string); ok && appName == appName {
				env := ApplicationEnvironment{
					Name:        node.Metadata["environment"].(string),
					Application: appName,
					Description: node.Spec["description"].(string),
					Status:      node.Metadata["status"].(string),
					Variables:   make(map[string]string), // TODO: Add variables support
					CreatedAt:   time.Now(),              // TODO: Get from node metadata
					UpdatedAt:   time.Now(),
				}
				environments = append(environments, env)
			}
		}
	}

	return environments, nil
}

// Service Management Methods
func (s *Service) CreateApplicationService(appName, serviceName, serviceType, description string, port, replicas int) error {
	if appName == "" {
		return fmt.Errorf("application name is required")
	}
	if serviceName == "" {
		return fmt.Errorf("service name is required")
	}

	// Check if application exists
	if !s.applicationExists(appName) {
		return fmt.Errorf("application not found: %s", appName)
	}

	// Create service contract - using existing service contract structure
	serviceContract := contracts.ServiceContract{
		Metadata: contracts.Metadata{
			Name:  fmt.Sprintf("%s-%s", appName, serviceName),
			Owner: "system",
		},
		Spec: contracts.ServiceSpec{
			Application: appName,
			Port:        port,
			Public:      false, // Default to private
		},
	}

	// Add to graph
	node, err := graph.ResolveContract(serviceContract)
	if err != nil {
		return fmt.Errorf("failed to resolve service contract: %w", err)
	}

	// Add additional metadata for querying and management
	node.Metadata["service_name"] = serviceName
	node.Metadata["service_type"] = serviceType
	node.Metadata["description"] = description
	node.Metadata["replicas"] = replicas

	s.Graph.AddNode(node)
	if err := s.Graph.Save(); err != nil {
		return fmt.Errorf("failed to save service: %w", err)
	}

	// Emit event
	if events.GlobalEventBus != nil {
		payload := map[string]interface{}{
			"application": appName,
			"service":     serviceName,
			"type":        serviceType,
			"description": description,
		}
		events.GlobalEventBus.Emit(events.EventTypeNotify, "ztdp-platform", "application_service_created", payload)
	}

	return nil
}

func (s *Service) ListApplicationServices(appName string) ([]ApplicationService, error) {
	if appName == "" {
		return nil, fmt.Errorf("application name is required")
	}

	// Get current graph
	currentGraph, err := s.Graph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	var services []ApplicationService

	for _, node := range currentGraph.Nodes {
		if node.Kind == "service" {
			if nodeApp, ok := node.Spec["application"].(string); ok && nodeApp == appName {
				// Extract port with proper type handling
				var port int
				if portVal, ok := node.Spec["port"]; ok {
					switch v := portVal.(type) {
					case int:
						port = v
					case float64:
						port = int(v)
					default:
						port = 0
					}
				}

				// Extract replicas with proper type handling
				var replicas int
				if replicasVal, ok := node.Metadata["replicas"]; ok {
					switch v := replicasVal.(type) {
					case int:
						replicas = v
					case float64:
						replicas = int(v)
					default:
						replicas = 0
					}
				}

				svc := ApplicationService{
					Name:        node.Metadata["service_name"].(string),
					Application: appName,
					Type:        node.Metadata["service_type"].(string),
					Description: node.Metadata["description"].(string),
					Port:        port,
					Replicas:    replicas,
					Variables:   make(map[string]string), // TODO: Add variables support
					CreatedAt:   time.Now(),              // TODO: Get from node metadata
					UpdatedAt:   time.Now(),
				}
				services = append(services, svc)
			}
		}
	}

	return services, nil
}

// Service Version Management Methods
func (s *Service) CreateServiceVersion(appName, serviceName, version, image string) (*ServiceVersion, error) {
	if appName == "" {
		return nil, fmt.Errorf("application name is required")
	}
	if serviceName == "" {
		return nil, fmt.Errorf("service name is required")
	}
	if version == "" {
		return nil, fmt.Errorf("version is required")
	}

	// Check if application and service exist
	if !s.applicationExists(appName) {
		return nil, fmt.Errorf("application not found: %s", appName)
	}
	if !s.serviceExists(appName, serviceName) {
		return nil, fmt.Errorf("service not found: %s in application %s", serviceName, appName)
	}

	serviceVersion := &ServiceVersion{
		Version:     version,
		Service:     serviceName,
		Application: appName,
		Image:       image,
		Status:      "created",
		Variables:   make(map[string]string),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create a simple node for service version (not using contracts for now)
	node := &graph.Node{
		ID:   fmt.Sprintf("%s-%s-%s", appName, serviceName, version),
		Kind: "service-version",
		Metadata: map[string]interface{}{
			"application": appName,
			"service":     serviceName,
			"version":     version,
			"image":       image,
			"status":      "created",
		},
		Spec: map[string]interface{}{
			"application": appName,
			"service":     serviceName,
			"version":     version,
			"image":       image,
		},
	}

	s.Graph.AddNode(node)
	if err := s.Graph.Save(); err != nil {
		return nil, fmt.Errorf("failed to save service version: %w", err)
	}

	// Emit event
	if events.GlobalEventBus != nil {
		payload := map[string]interface{}{
			"application": appName,
			"service":     serviceName,
			"version":     version,
			"image":       image,
		}
		events.GlobalEventBus.Emit(events.EventTypeNotify, "ztdp-platform", "service_version_created", payload)
	}

	return serviceVersion, nil
}

func (s *Service) GetServiceVersion(appName, serviceName, version string) (*ServiceVersion, error) {
	// Get current graph
	currentGraph, err := s.Graph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	nodeID := fmt.Sprintf("%s-%s-%s", appName, serviceName, version)
	if node, ok := currentGraph.Nodes[nodeID]; ok && node.Kind == "service-version" {
		serviceVersion := &ServiceVersion{
			Version:     node.Metadata["version"].(string),
			Service:     node.Metadata["service"].(string),
			Application: node.Metadata["application"].(string),
			Image:       node.Metadata["image"].(string),
			Status:      node.Metadata["status"].(string),
			Variables:   make(map[string]string), // TODO: Add variables support
			CreatedAt:   time.Now(),              // TODO: Get from node metadata
			UpdatedAt:   time.Now(),
		}
		return serviceVersion, nil
	}

	return nil, fmt.Errorf("service version not found: %s-%s-%s", appName, serviceName, version)
}

// Release Management Methods
func (s *Service) CreateApplicationRelease(appName, environment, version string, services map[string]string) (*ApplicationRelease, error) {
	if appName == "" {
		return nil, fmt.Errorf("application name is required")
	}
	if environment == "" {
		return nil, fmt.Errorf("environment is required")
	}
	if version == "" {
		return nil, fmt.Errorf("release version is required")
	}

	// Check if application exists
	if !s.applicationExists(appName) {
		return nil, fmt.Errorf("application not found: %s", appName)
	}

	release := &ApplicationRelease{
		Version:       version,
		Application:   appName,
		Environment:   environment,
		Status:        "created",
		Services:      services,
		Configuration: make(map[string]interface{}),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Create node for release
	node := &graph.Node{
		ID:   fmt.Sprintf("%s-%s-%s", appName, environment, version),
		Kind: "application-release",
		Metadata: map[string]interface{}{
			"application": appName,
			"environment": environment,
			"version":     version,
			"status":      "created",
		},
		Spec: map[string]interface{}{
			"application":   appName,
			"environment":   environment,
			"version":       version,
			"services":      services,
			"configuration": make(map[string]interface{}),
		},
	}

	s.Graph.AddNode(node)
	if err := s.Graph.Save(); err != nil {
		return nil, fmt.Errorf("failed to save release: %w", err)
	}

	// Emit event
	if events.GlobalEventBus != nil {
		payload := map[string]interface{}{
			"application": appName,
			"environment": environment,
			"version":     version,
			"services":    services,
		}
		events.GlobalEventBus.Emit(events.EventTypeNotify, "ztdp-platform", "application_release_created", payload)
	}

	return release, nil
}

func (s *Service) ListApplicationReleases(appName string) ([]ApplicationRelease, error) {
	if appName == "" {
		return nil, fmt.Errorf("application name is required")
	}

	// Get current graph
	currentGraph, err := s.Graph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	var releases []ApplicationRelease

	for _, node := range currentGraph.Nodes {
		if node.Kind == "application-release" {
			if nodeApp, ok := node.Metadata["application"].(string); ok && nodeApp == appName {
				release := ApplicationRelease{
					Version:       node.Metadata["version"].(string),
					Application:   appName,
					Environment:   node.Metadata["environment"].(string),
					Status:        node.Metadata["status"].(string),
					Services:      node.Spec["services"].(map[string]string),
					Configuration: node.Spec["configuration"].(map[string]interface{}),
					CreatedAt:     time.Now(), // TODO: Get from node metadata
					UpdatedAt:     time.Now(),
				}
				releases = append(releases, release)
			}
		}
	}

	return releases, nil
}

// Helper methods
func (s *Service) applicationExists(appName string) bool {
	// Get current graph
	currentGraph, err := s.Graph.Graph()
	if err != nil {
		return false
	}

	for _, node := range currentGraph.Nodes {
		if node.Kind == "application" {
			if nodeName, ok := node.Metadata["name"].(string); ok && nodeName == appName {
				return true
			}
		}
	}
	return false
}

func (s *Service) serviceExists(appName, serviceName string) bool {
	// Get current graph
	currentGraph, err := s.Graph.Graph()
	if err != nil {
		return false
	}

	for _, node := range currentGraph.Nodes {
		if node.Kind == "service" {
			if nodeApp, ok := node.Spec["application"].(string); ok && nodeApp == appName {
				if nodeService, ok := node.Metadata["service_name"].(string); ok && nodeService == serviceName {
					return true
				}
			}
		}
	}
	return false
}
