package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

// Service for release business logic
type ReleaseService struct {
	Graph *graph.GlobalGraph
}

func NewReleaseService(g *graph.GlobalGraph) *ReleaseService {
	return &ReleaseService{Graph: g}
}

// CreateRelease validates and creates a release node in the graph
func (r *ReleaseService) CreateRelease(release contracts.ReleaseContract) error {
	if err := release.Validate(); err != nil {
		return err
	}

	node, _ := graph.ResolveContract(release)
	r.Graph.AddNode(node)

	// Create edges to link release to application and service versions
	r.linkReleaseToApplication(release.Spec.Application, release.Metadata.Name)
	for _, serviceVersion := range release.Spec.ServiceVersions {
		r.linkReleaseToServiceVersion(serviceVersion, release.Metadata.Name)
	}

	// Save the graph
	if err := r.Graph.Save(); err != nil {
		return err
	}

	// Emit HIGH-LEVEL BUSINESS EVENT - Simple and clean!
	if events.GlobalEventBus != nil {
		payload := map[string]interface{}{
			"release_name":     release.Metadata.Name,
			"application":      release.Spec.Application,
			"version":          release.Spec.Version,
			"service_versions": release.Spec.ServiceVersions,
			"status":           release.Spec.Status,
			"strategy":         release.Spec.Strategy,
			"owner":            release.Metadata.Owner,
			"timestamp":        release.Spec.Timestamp,
		}
		events.GlobalEventBus.Emit(events.EventTypeNotify, "ztdp-platform", "release_created", payload)
	}

	return nil
}

// ListReleases returns all releases for an application
func (r *ReleaseService) ListReleases(application string) ([]contracts.ReleaseContract, error) {
	nodesMap, err := r.Graph.Nodes()
	if err != nil {
		return nil, err
	}

	var releases []contracts.ReleaseContract
	for _, node := range nodesMap {
		if node.Kind == "release" {
			// Convert node back to contract
			release := contracts.ReleaseContract{
				Metadata: contracts.Metadata{
					Name:  node.ID,
					Owner: getStringFromInterface(node.Metadata["owner"]),
				},
				Spec: contracts.ReleaseSpec{
					Application:     getStringFromInterface(node.Spec["application"]),
					Version:         getStringFromInterface(node.Spec["version"]),
					ServiceVersions: getStringSliceFromInterface(node.Spec["service_versions"]),
					Status:          getStringFromInterface(node.Spec["status"]),
					Strategy:        getStringFromInterface(node.Spec["strategy"]),
					Notes:           getStringFromInterface(node.Spec["notes"]),
				},
			}

			if application == "" || release.Spec.Application == application {
				releases = append(releases, release)
			}
		}
	}

	return releases, nil
}

// GetRelease retrieves a specific release by name
func (r *ReleaseService) GetRelease(releaseName string) (*contracts.ReleaseContract, error) {
	nodesMap, err := r.Graph.Nodes()
	if err != nil {
		return nil, err
	}

	for _, node := range nodesMap {
		if node.Kind == "release" && node.ID == releaseName {
			release := contracts.ReleaseContract{
				Metadata: contracts.Metadata{
					Name:  node.ID,
					Owner: getStringFromInterface(node.Metadata["owner"]),
				},
				Spec: contracts.ReleaseSpec{
					Application:     getStringFromInterface(node.Spec["application"]),
					Version:         getStringFromInterface(node.Spec["version"]),
					ServiceVersions: getStringSliceFromInterface(node.Spec["service_versions"]),
					Status:          getStringFromInterface(node.Spec["status"]),
					Strategy:        getStringFromInterface(node.Spec["strategy"]),
					Notes:           getStringFromInterface(node.Spec["notes"]),
				},
			}
			return &release, nil
		}
	}

	return nil, errors.New("release not found")
}

// CreateReleaseFromRequest creates a release from deployment request
func (r *ReleaseService) CreateReleaseFromRequest(ctx context.Context, application, environment string, serviceVersions []string, notes string) (*contracts.ReleaseContract, error) {
	if application == "" {
		return nil, errors.New("application is required")
	}
	if len(serviceVersions) == 0 {
		return nil, errors.New("at least one service version is required")
	}

	timestamp := time.Now()
	releaseName := fmt.Sprintf("%s-rel-%d", application, timestamp.Unix())
	version := fmt.Sprintf("v%d", timestamp.Unix())

	release := contracts.ReleaseContract{
		Metadata: contracts.Metadata{
			Name:  releaseName,
			Owner: "ztdp-system", // Could be extracted from context
		},
		Spec: contracts.ReleaseSpec{
			Application:     application,
			Version:         version,
			ServiceVersions: serviceVersions,
			Status:          "pending",
			Strategy:        "rolling",
			Configuration:   make(map[string]string),
			Notes:           notes,
			Timestamp:       timestamp,
		},
	}

	if err := r.CreateRelease(release); err != nil {
		return nil, err
	}

	return &release, nil
}

// DeleteRelease removes a release from the graph
func (r *ReleaseService) DeleteRelease(releaseName string) error {
	// Check if release exists
	_, err := r.GetRelease(releaseName)
	if err != nil {
		return fmt.Errorf("release not found: %s", releaseName)
	}

	// For now, we don't have a direct way to remove from the graph
	// This would need to be implemented when graph deletion is supported
	// Return an error indicating this is not yet implemented
	return fmt.Errorf("release deletion not yet implemented in graph backend")
}

// Helper functions for linking
func (r *ReleaseService) linkReleaseToApplication(applicationName, releaseName string) {
	r.Graph.AddEdge(releaseName, applicationName, "targets")
}

func (r *ReleaseService) linkReleaseToServiceVersion(serviceVersionName, releaseName string) {
	r.Graph.AddEdge(releaseName, serviceVersionName, "includes")
}

// Helper functions for type conversion
func getStringFromInterface(val interface{}) string {
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

func getStringSliceFromInterface(val interface{}) []string {
	if slice, ok := val.([]interface{}); ok {
		result := make([]string, len(slice))
		for i, v := range slice {
			if str, ok := v.(string); ok {
				result[i] = str
			}
		}
		return result
	}
	if slice, ok := val.([]string); ok {
		return slice
	}
	return []string{}
}
