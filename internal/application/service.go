package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/resources"
)

type ServiceService struct {
	Graph *graph.GlobalGraph
}

func NewServiceService(g *graph.GlobalGraph) *ServiceService {
	return &ServiceService{Graph: g}
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
	if svc == nil {
		return nil, fmt.Errorf("service contract cannot be nil")
	}

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

// Internal methods that work with contracts (actual existing logic)
func (s *ServiceService) createServiceInternal(appName string, svc contracts.ServiceContract) error {
	// Auto-populate application from URL parameter to eliminate redundant validation
	svc.Spec.Application = appName

	if err := svc.Validate(); err != nil {
		return err
	}
	node, err := graph.ResolveContract(svc)
	if err != nil {
		return err
	}
	s.Graph.AddNode(node)
	s.Graph.AddEdge(appName, svc.Metadata.Name, "owns")
	return s.Graph.Save()
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
