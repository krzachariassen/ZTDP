package application

import (
	"encoding/json"
	"errors"

	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

// Service for application business logic
type Service struct {
	Graph *graph.GlobalGraph
}

func NewService(g *graph.GlobalGraph) *Service {
	return &Service{
		Graph: g,
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
		events.GlobalEventBus.Emit(events.EventTypeApplicationCreated, "ztdp-platform", app.Metadata.Name, payload)
	}

	return nil
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
