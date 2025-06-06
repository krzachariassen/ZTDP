package resources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

const resourceCatalogNodeID = "resource-catalog"
const resourceCatalogKind = "resource_register"

type Service struct {
	Graph *graph.GlobalGraph
}

func NewService(g *graph.GlobalGraph) *Service {
	return &Service{Graph: g}
}

// ResourceRequest represents a resource creation request
type ResourceRequest struct {
	Kind     string                 `json:"kind"`
	Metadata map[string]interface{} `json:"metadata"`
	Spec     map[string]interface{} `json:"spec"`
}

// ResourceResponse represents a resource creation response
type ResourceResponse struct {
	ID       string                 `json:"id"`
	Kind     string                 `json:"kind"`
	Metadata map[string]interface{} `json:"metadata"`
	Spec     map[string]interface{} `json:"spec"`
}

// ResourceInstanceResponse represents a resource instance operation response
type ResourceInstanceResponse struct {
	Message      string `json:"message"`
	InstanceName string `json:"instance_name"`
	Status       string `json:"status"`
	CatalogRef   string `json:"catalog_ref"`
	Application  string `json:"application"`
}

// ensureResourceCatalogRoot ensures the resource catalog root node exists in the graph
func (s *Service) ensureResourceCatalogRoot() {
	if node, err := s.Graph.GetNode(resourceCatalogNodeID); err != nil || node == nil {
		root := &graph.Node{
			ID:   resourceCatalogNodeID,
			Kind: resourceCatalogKind,
			Metadata: map[string]interface{}{
				"name":  resourceCatalogNodeID,
				"owner": "platform-team",
			},
			Spec: map[string]interface{}{
				"description": "Root node for all resource types in the platform",
			},
		}
		s.Graph.AddNode(root)
	}
}

// repairResourceCatalogRelationships repairs the resource catalog: ensures all resource_type and resource nodes are owned by the catalog root
func (s *Service) repairResourceCatalogRelationships() {
	s.ensureResourceCatalogRoot()

	nodes, err := s.Graph.Nodes()
	if err != nil {
		return // Handle gracefully if we can't get nodes
	}

	edges, err := s.Graph.Edges()
	if err != nil {
		return // Handle gracefully if we can't get edges
	}

	for _, node := range nodes {
		// Ensure both resource_type and catalog resource nodes are owned by resource_register
		if node.Kind == "resource_type" || (node.Kind == "resource" && !s.isResourceInstance(node)) {
			hasEdge := false
			for _, edge := range edges[resourceCatalogNodeID] {
				if edge.To == node.ID && edge.Type == "owns" {
					hasEdge = true
					break
				}
			}
			if !hasEdge {
				s.Graph.AddEdge(resourceCatalogNodeID, node.ID, "owns")
			}
		}
	}
}

// CreateResource creates a new resource (from catalog) and handles the business logic
func (s *Service) CreateResource(req ResourceRequest) (*ResourceResponse, error) {
	if req.Kind != "resource_type" && req.Kind != "resource" {
		return nil, errors.New("invalid kind for resource catalog")
	}

	// Validate required metadata fields
	nameVal, nameOk := req.Metadata["name"].(string)
	ownerVal, ownerOk := req.Metadata["owner"].(string)
	if !nameOk || nameVal == "" {
		return nil, errors.New("resource metadata 'name' is required")
	}
	if req.Kind == "resource_type" && (!ownerOk || ownerVal == "") {
		return nil, errors.New("resource type metadata 'owner' is required")
	}
	if req.Kind == "resource" && !ownerOk {
		// For resource instances, allow owner to be empty (optional), but always set to string
		ownerVal = ""
	}

	// Build the contract
	var node *graph.Node
	if req.Kind == "resource_type" {
		resourceType := contracts.ResourceTypeContract{}
		metadata := contracts.Metadata{
			Name:  nameVal,
			Owner: ownerVal,
		}
		resourceType.Metadata = metadata
		if specBytes, err := json.Marshal(req.Spec); err == nil {
			json.Unmarshal(specBytes, &resourceType.Spec)
		}
		node, _ = graph.ResolveContract(resourceType)
	} else {
		resource := contracts.ResourceContract{}
		metadata := contracts.Metadata{
			Name:  nameVal,
			Owner: ownerVal,
		}
		resource.Metadata = metadata
		if specBytes, err := json.Marshal(req.Spec); err == nil {
			json.Unmarshal(specBytes, &resource.Spec)
		}
		node, _ = graph.ResolveContract(resource)
	}

	s.ensureResourceCatalogRoot()
	s.Graph.AddNode(node)

	// If this is a resource_type, add an 'owns' edge from the catalog root
	if node.Kind == "resource_type" {
		s.Graph.AddEdge(resourceCatalogNodeID, node.ID, "owns")
	}

	s.repairResourceCatalogRelationships()

	if err := s.Graph.Save(); err != nil {
		return nil, fmt.Errorf("failed to save resource: %w", err)
	}

	return &ResourceResponse{
		ID:       node.ID,
		Kind:     node.Kind,
		Metadata: node.Metadata,
		Spec:     node.Spec,
	}, nil
}

// AddResourceToApplication creates a resource instance for an application
func (s *Service) AddResourceToApplication(appName, resourceName, instanceName string) (*ResourceInstanceResponse, error) {
	// Use default naming if no custom instance name provided
	if instanceName == "" {
		instanceName = appName + "-" + resourceName
	}

	// Check if application exists
	appNode, err := s.Graph.GetNode(appName)
	if err != nil || appNode == nil || appNode.Kind != "application" {
		return nil, errors.New("application not found")
	}

	// Check if catalog resource exists
	catalogNode, err := s.Graph.GetNode(resourceName)
	if err != nil || catalogNode == nil || catalogNode.Kind != "resource" {
		return nil, errors.New("resource not found in catalog")
	}

	// Get the resource type name from the catalog resource's spec
	resourceTypeName, ok := catalogNode.Spec["type"].(string)
	if !ok || resourceTypeName == "" {
		return nil, errors.New("catalog resource missing or invalid 'type' field")
	}

	// Check if the resource type exists
	resourceTypeNode, err := s.Graph.GetNode(resourceTypeName)
	if err != nil || resourceTypeNode == nil || resourceTypeNode.Kind != "resource_type" {
		return nil, fmt.Errorf("resource type '%s' not found", resourceTypeName)
	}

	// Check if instance already exists
	if existingNode, err := s.Graph.GetNode(instanceName); err == nil && existingNode != nil {
		if existingNode.Kind == "resource" {
			// Resource instance already exists - return success (idempotent)
			return &ResourceInstanceResponse{
				Message:      "Resource instance already exists",
				InstanceName: instanceName,
				Status:       "exists",
				CatalogRef:   resourceName,
				Application:  appName,
			}, nil
		} else {
			return nil, errors.New("a node with this name already exists but is not a resource")
		}
	}

	// Create the resource instance
	resourceInstance := &graph.Node{
		ID:   instanceName,
		Kind: "resource",
		Metadata: map[string]interface{}{
			"name":        instanceName,
			"owner":       catalogNode.Metadata["owner"],
			"application": appName,
			"catalog_ref": resourceName,
		},
		Spec: catalogNode.Spec, // Inherit spec from catalog resource
	}

	// Add the resource instance to the graph
	s.Graph.AddNode(resourceInstance)

	// Create relationships
	if err := s.Graph.AddEdge(appName, instanceName, graph.EdgeTypeOwns); err != nil {
		if err.Error() != "edge already exists" {
			return nil, fmt.Errorf("failed to link resource to application: %w", err)
		}
	}

	// Create instance_of edge to resource_type (not catalog resource)
	if err := s.Graph.AddEdge(instanceName, resourceTypeName, graph.EdgeTypeInstanceOf); err != nil {
		if err.Error() != "edge already exists" {
			return nil, fmt.Errorf("failed to link instance to resource type: %w", err)
		}
	}

	if err := s.Graph.Save(); err != nil {
		return nil, errors.New("failed to save resource instance")
	}

	return &ResourceInstanceResponse{
		Message:      "Resource instance created successfully",
		InstanceName: instanceName,
		Status:       "created",
		CatalogRef:   resourceName,
		Application:  appName,
	}, nil
}

// LinkServiceToResource creates a 'uses' edge from service to resource
func (s *Service) LinkServiceToResource(appName, serviceName, resourceName string) (*ResourceInstanceResponse, error) {
	// Validate application exists
	appNode, err := s.Graph.GetNode(appName)
	if err != nil || appNode == nil || appNode.Kind != "application" {
		return nil, errors.New("application not found")
	}

	// Validate service exists
	serviceNode, err := s.Graph.GetNode(serviceName)
	if err != nil || serviceNode == nil || serviceNode.Kind != "service" {
		return nil, errors.New("service not found")
	}

	// Find the resource instance in this application's context
	resourceInstanceID, err := s.findResourceInstanceInApplication(appName, resourceName)
	if err != nil {
		return nil, fmt.Errorf("resource instance not found in application: %w", err)
	}

	// Add edge - validation is now handled at the graph level
	if err := s.Graph.AddEdge(serviceName, resourceInstanceID, graph.EdgeTypeUses); err != nil {
		// For MVP: ignore "edge already exists" errors - this is additive behavior
		if err.Error() == "edge already exists" {
			return &ResourceInstanceResponse{
				Message:      "service already linked to resource instance",
				InstanceName: resourceInstanceID,
				Status:       "exists",
			}, nil
		}
		return nil, fmt.Errorf("failed to link service to resource: %w", err)
	}

	if err := s.Graph.Save(); err != nil {
		return nil, errors.New("failed to save graph")
	}

	return &ResourceInstanceResponse{
		Message:      "service linked to resource instance",
		InstanceName: resourceInstanceID,
		Status:       "created",
	}, nil
}

// findResourceInstanceInApplication finds the resource instance owned by an application that references the catalog resource
func (s *Service) findResourceInstanceInApplication(appName, catalogResourceName string) (string, error) {
	// First try the new predictable naming scheme
	predictableInstanceName := appName + "-" + catalogResourceName
	if node, err := s.Graph.GetNode(predictableInstanceName); err == nil && node != nil {
		if node.Kind == "resource" {
			// Verify it's actually owned by the application and references the catalog resource
			if appRef, ok := node.Metadata["application"].(string); ok && appRef == appName {
				if catRef, ok := node.Metadata["catalog_ref"].(string); ok && catRef == catalogResourceName {
					return predictableInstanceName, nil
				}
			}
		}
	}

	// Fallback to searching through all nodes (for backward compatibility with old UUID-based instances)
	nodes, err := s.Graph.Nodes()
	if err != nil {
		return "", fmt.Errorf("failed to get nodes: %w", err)
	}

	// Get edges to find ownership relationships
	edges, err := s.Graph.Edges()
	if err != nil {
		return "", fmt.Errorf("failed to get edges: %w", err)
	}

	// Find resource instances owned by this application
	for _, edge := range edges[appName] {
		if edge.Type == graph.EdgeTypeOwns {
			// Check if the target is a resource instance
			if node, ok := nodes[edge.To]; ok && s.isResourceInstance(node) {
				// Check if this instance references the catalog resource
				if catRef, ok := node.Metadata["catalog_ref"].(string); ok && catRef == catalogResourceName {
					return edge.To, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no resource instance found for catalog resource '%s' in application '%s'", catalogResourceName, appName)
}

// ListResources returns all resources as basic maps
func (s *Service) ListResources() ([]map[string]interface{}, error) {
	nodes, err := s.Graph.Nodes()
	if err != nil {
		return nil, err
	}

	var resources []map[string]interface{}
	for _, node := range nodes {
		if node.Kind == "resource" || node.Kind == "resource_type" {
			resource := map[string]interface{}{
				"id":       node.ID,
				"kind":     node.Kind,
				"metadata": node.Metadata,
				"spec":     node.Spec,
			}
			resources = append(resources, resource)
		}
	}
	return resources, nil
}

// ListApplicationResources returns all resource nodes owned by the application
func (s *Service) ListApplicationResources(appName string) ([]map[string]interface{}, error) {
	appNode, err := s.Graph.GetNode(appName)
	if err != nil || appNode == nil || appNode.Kind != "application" {
		return nil, errors.New("application not found")
	}

	edges, err := s.Graph.Edges()
	if err != nil {
		return nil, errors.New("failed to get application resources")
	}

	nodes, err := s.Graph.Nodes()
	if err != nil {
		return nil, errors.New("failed to get application resources")
	}

	resources := []map[string]interface{}{}
	for _, edge := range edges[appName] {
		if edge.Type == "owns" {
			if node, ok := nodes[edge.To]; ok && node.Kind == "resource" {
				resource := map[string]interface{}{
					"id":       node.ID,
					"kind":     node.Kind,
					"metadata": node.Metadata,
					"spec":     node.Spec,
				}
				resources = append(resources, resource)
			}
		}
	}
	return resources, nil
}

// ListServiceResources returns all resource nodes linked by 'uses' edge from the service
func (s *Service) ListServiceResources(serviceName string) ([]map[string]interface{}, error) {
	serviceNode, err := s.Graph.GetNode(serviceName)
	if err != nil || serviceNode == nil || serviceNode.Kind != "service" {
		return nil, errors.New("service not found")
	}

	edges, err := s.Graph.Edges()
	if err != nil {
		return nil, errors.New("failed to get service resources")
	}

	nodes, err := s.Graph.Nodes()
	if err != nil {
		return nil, errors.New("failed to get service resources")
	}

	resources := []map[string]interface{}{}
	for _, edge := range edges[serviceName] {
		if edge.Type == "uses" {
			if node, ok := nodes[edge.To]; ok && node.Kind == "resource" {
				resource := map[string]interface{}{
					"id":       node.ID,
					"kind":     node.Kind,
					"metadata": node.Metadata,
					"spec":     node.Spec,
				}
				resources = append(resources, resource)
			}
		}
	}
	return resources, nil
}

// isResourceInstance checks if a resource node is an instance (owned by an application) vs a catalog resource
func (s *Service) isResourceInstance(node *graph.Node) bool {
	if node.Kind != "resource" {
		return false
	}
	// Resource instances have "application" and "catalog_ref" in metadata
	if app, hasApp := node.Metadata["application"]; hasApp && app != nil {
		if catRef, hasCatRef := node.Metadata["catalog_ref"]; hasCatRef && catRef != nil {
			return true
		}
	}
	return false
}

// CreateResourceFromData creates a resource from raw data (legacy method for backward compatibility)
func (s *Service) CreateResourceFromData(resourceData map[string]interface{}) (map[string]interface{}, error) {
	// Convert raw data to ResourceRequest
	kindVal, ok := resourceData["kind"].(string)
	if !ok {
		return nil, errors.New("kind field is required")
	}

	metadataVal, ok := resourceData["metadata"].(map[string]interface{})
	if !ok {
		return nil, errors.New("metadata field is required")
	}

	specVal, ok := resourceData["spec"].(map[string]interface{})
	if !ok {
		specVal = make(map[string]interface{})
	}

	req := ResourceRequest{
		Kind:     kindVal,
		Metadata: metadataVal,
		Spec:     specVal,
	}

	// Use the new method
	resp, err := s.CreateResource(req)
	if err != nil {
		return nil, err
	}

	// Convert back to raw map
	return map[string]interface{}{
		"id":       resp.ID,
		"kind":     resp.Kind,
		"metadata": resp.Metadata,
		"spec":     resp.Spec,
	}, nil
}

// CreateResourceFromContract creates resource from contract with context support
// This method supports contract-driven AI operations while maintaining business logic
func (s *Service) CreateResourceFromContract(ctx context.Context, resource *contracts.ResourceContract) (interface{}, error) {
	if err := resource.Validate(); err != nil {
		return nil, err
	}

	// Check if resource type exists
	resourceTypeNode, err := s.Graph.GetNode(resource.Spec.Type)
	if err != nil || resourceTypeNode == nil || resourceTypeNode.Kind != "resource_type" {
		return nil, fmt.Errorf("resource type '%s' not found", resource.Spec.Type)
	}

	// Create resource instance
	resourceInstance := &graph.Node{
		ID:   resource.Metadata.Name,
		Kind: "resource",
		Metadata: map[string]interface{}{
			"name":  resource.Metadata.Name,
			"owner": resource.Metadata.Owner,
		},
		Spec: map[string]interface{}{
			"type":     resource.Spec.Type,
			"version":  resource.Spec.Version,
			"tier":     resource.Spec.Tier,
			"capacity": resource.Spec.Capacity,
			"plan":     resource.Spec.Plan,
		},
	}

	s.Graph.AddNode(resourceInstance)
	s.Graph.AddEdge(resource.Metadata.Name, resource.Spec.Type, graph.EdgeTypeInstanceOf)

	if err := s.Graph.Save(); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":   resource.Metadata.Name,
		"status": "created",
		"type":   resource.Spec.Type,
		"tier":   resource.Spec.Tier,
	}, nil
}
