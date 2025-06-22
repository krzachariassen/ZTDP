package application

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

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
