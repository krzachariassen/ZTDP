package environment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/resources"
)

type Service struct {
	Graph *graph.GlobalGraph
}

func NewService(g *graph.GlobalGraph) *Service {
	return &Service{Graph: g}
}

// CreateEnvironment validates and creates an environment node in the graph
func (s *Service) CreateEnvironment(env contracts.EnvironmentContract) error {
	node, err := graph.ResolveContract(env)
	if err != nil {
		return err
	}
	s.Graph.AddNode(node)
	return s.Graph.Save()
}

// CreateEnvironmentFromContract creates environment from contract with context support
// This method supports contract-driven AI operations while maintaining business logic
func (s *Service) CreateEnvironmentFromContract(ctx context.Context, env *contracts.EnvironmentContract) (interface{}, error) {
	if env == nil {
		return nil, fmt.Errorf("environment contract cannot be nil")
	}

	node, err := graph.ResolveContract(*env)
	if err != nil {
		return nil, err
	}

	s.Graph.AddNode(node)

	if err := s.Graph.Save(); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":        env.Metadata.Name,
		"status":      "created",
		"description": env.Spec.Description,
		"owner":       env.Metadata.Owner,
	}, nil
}

// CreateEnvironmentFromData creates an environment from raw data
func (s *Service) CreateEnvironmentFromData(envData map[string]interface{}) (map[string]interface{}, error) {
	// Convert raw data to contract internally
	var env contracts.EnvironmentContract
	if err := mapToContract(envData, &env); err != nil {
		return nil, err
	}

	// Use existing contract-based logic
	if err := s.CreateEnvironment(env); err != nil {
		return nil, err
	}

	return contractToMap(env), nil
}

// ListEnvironments returns all environment nodes from the graph
func (s *Service) ListEnvironments() ([]contracts.EnvironmentContract, error) {
	nodes, err := s.Graph.Nodes()
	if err != nil {
		return nil, errors.New("failed to get current graph: " + err.Error())
	}

	envs := []contracts.EnvironmentContract{}
	for _, node := range nodes {
		if node.Kind == "environment" {
			contract, err := resources.LoadNode(node.Kind, node.Spec, contracts.Metadata{
				Name:  node.Metadata["name"].(string),
				Owner: node.Metadata["owner"].(string),
			})
			if err == nil {
				if env, ok := contract.(*contracts.EnvironmentContract); ok {
					envs = append(envs, *env)
				}
			}
		}
	}
	return envs, nil
}

// ListEnvironmentsAsData returns environments as basic maps
func (s *Service) ListEnvironmentsAsData() ([]map[string]interface{}, error) {
	// Use existing logic to get contracts
	contracts, err := s.ListEnvironments()
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

// LinkAppAllowedInEnvironment links an application to an environment with an 'allowed_in' edge
func (s *Service) LinkAppAllowedInEnvironment(appName, envName string) error {
	appNode, err := s.Graph.GetNode(appName)
	if err != nil || appNode == nil || appNode.Kind != "application" {
		return errors.New("application not found")
	}
	envNode, err := s.Graph.GetNode(envName)
	if err != nil || envNode == nil || envNode.Kind != "environment" {
		return errors.New("environment not found")
	}
	if err := s.Graph.AddEdge(appName, envName, "allowed_in"); err != nil {
		return err
	}
	return s.Graph.Save()
}

// ListAllowedEnvironments returns all allowed environments for an application
func (s *Service) ListAllowedEnvironments(appName string) ([]contracts.EnvironmentContract, error) {
	appNode, err := s.Graph.GetNode(appName)
	if err != nil || appNode == nil || appNode.Kind != "application" {
		return nil, errors.New("application not found")
	}

	edges, err := s.Graph.Edges()
	if err != nil {
		return nil, errors.New("failed to get allowed environments: " + err.Error())
	}

	nodes, err := s.Graph.Nodes()
	if err != nil {
		return nil, errors.New("failed to get nodes: " + err.Error())
	}

	var envs []contracts.EnvironmentContract
	for _, edge := range edges[appName] {
		if edge.Type == "allowed_in" {
			if envNode, ok := nodes[edge.To]; ok && envNode.Kind == "environment" {
				contract, err := resources.LoadNode(envNode.Kind, envNode.Spec, contracts.Metadata{
					Name:  envNode.Metadata["name"].(string),
					Owner: envNode.Metadata["owner"].(string),
				})
				if err == nil {
					if env, ok := contract.(*contracts.EnvironmentContract); ok {
						envs = append(envs, *env)
					}
				}
			}
		}
	}
	return envs, nil
}

// ListAllowedEnvironmentsAsData returns allowed environments as basic maps
func (s *Service) ListAllowedEnvironmentsAsData(appName string) ([]map[string]interface{}, error) {
	// Use existing logic to get contracts
	contracts, err := s.ListAllowedEnvironments(appName)
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

// UpdateAllowedEnvironments replaces allowed environments for an application
func (s *Service) UpdateAllowedEnvironments(appName string, envs []string) error {
	appNode, err := s.Graph.GetNode(appName)
	if err != nil || appNode == nil || appNode.Kind != "application" {
		return errors.New("application not found")
	}

	// For the MVP, we'll use additive behavior and ignore duplicate edge errors
	for _, envName := range envs {
		envNode, err := s.Graph.GetNode(envName)
		if err != nil || envNode == nil || envNode.Kind != "environment" {
			return errors.New("environment not found: " + envName)
		}
		if err := s.Graph.AddEdge(appName, envName, "allowed_in"); err != nil {
			// If edge already exists, ignore the error
			if err.Error() != "edge already exists" {
				return errors.New("failed to add permission: " + err.Error())
			}
		}
	}
	return s.Graph.Save()
}

// AddAllowedEnvironments adds allowed_in edges for an application (does not remove existing)
func (s *Service) AddAllowedEnvironments(appName string, envs []string) error {
	appNode, err := s.Graph.GetNode(appName)
	if err != nil || appNode == nil || appNode.Kind != "application" {
		return errors.New("application not found")
	}
	for _, envName := range envs {
		envNode, err := s.Graph.GetNode(envName)
		if err != nil || envNode == nil || envNode.Kind != "environment" {
			return errors.New("environment not found: " + envName)
		}
		if err := s.Graph.AddEdge(appName, envName, "allowed_in"); err != nil {
			// If edge already exists, ignore the error
			if err.Error() != "edge already exists" {
				return errors.New("failed to add allowed environment: " + err.Error())
			}
		}
	}
	return s.Graph.Save()
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
