package policies

import (
	"fmt"

	"github.com/krzachariassen/ZTDP/internal/graph"
)

// Service wraps PolicyEvaluator and provides business logic methods for policy operations
type Service struct {
	evaluator  *PolicyEvaluator
	graphStore GraphStoreInterface
	env        string
}

// NewService creates a new policy service
func NewService(graphStore GraphStoreInterface, env string) *Service {
	evaluator := NewPolicyEvaluator(graphStore, env)

	return &Service{
		evaluator:  evaluator,
		graphStore: graphStore,
		env:        env,
	}
}

// PolicyOperationRequest represents a policy operation request
type PolicyOperationRequest struct {
	Operation   string                 `json:"operation"`
	FromID      string                 `json:"from_id,omitempty"`
	ToID        string                 `json:"to_id,omitempty"`
	EdgeType    string                 `json:"edge_type,omitempty"`
	PolicyID    string                 `json:"policy_id,omitempty"`
	CheckID     string                 `json:"check_id,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Results     map[string]interface{} `json:"results,omitempty"`
}

// PolicyOperationResponse represents a policy operation response
type PolicyOperationResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// ExecuteOperation executes a policy operation
func (s *Service) ExecuteOperation(req PolicyOperationRequest, user string) (*PolicyOperationResponse, error) {
	switch req.Operation {
	case "check":
		return s.checkTransition(req, user)
	case "create_policy":
		return s.createPolicy(req)
	case "create_check":
		return s.createCheck(req)
	case "update_check":
		return s.updateCheck(req)
	case "satisfy":
		return s.satisfyPolicy(req)
	default:
		return &PolicyOperationResponse{
			Success: false,
			Error:   fmt.Sprintf("Unknown operation: %s", req.Operation),
		}, fmt.Errorf("unknown operation: %s", req.Operation)
	}
}

// checkTransition validates if a transition is allowed
func (s *Service) checkTransition(req PolicyOperationRequest, user string) (*PolicyOperationResponse, error) {
	err := s.evaluator.ValidateTransition(req.FromID, req.ToID, req.EdgeType, user)
	if err != nil {
		return &PolicyOperationResponse{
			Success: false,
			Error:   err.Error(),
			Data: map[string]interface{}{
				"allowed": false,
			},
		}, nil // Don't return error as this is a valid business response
	}

	return &PolicyOperationResponse{
		Success: true,
		Data: map[string]interface{}{
			"allowed": true,
		},
	}, nil
}

// createPolicy creates a new policy node
func (s *Service) createPolicy(req PolicyOperationRequest) (*PolicyOperationResponse, error) {
	policyNode, err := s.evaluator.CreatePolicyNode(
		req.Name,
		req.Description,
		req.Type,
		req.Parameters,
	)
	if err != nil {
		return &PolicyOperationResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create policy: %v", err),
		}, err
	}

	return &PolicyOperationResponse{
		Success: true,
		Message: "Policy created",
		Data: map[string]interface{}{
			"policy_id": policyNode.ID,
		},
	}, nil
}

// createCheck creates a new check node
func (s *Service) createCheck(req PolicyOperationRequest) (*PolicyOperationResponse, error) {
	checkNode, err := s.evaluator.CreateCheckNode(
		req.CheckID,
		req.Name,
		req.Type,
		req.Parameters,
	)
	if err != nil {
		return &PolicyOperationResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create check: %v", err),
		}, err
	}

	return &PolicyOperationResponse{
		Success: true,
		Message: "Check created",
		Data: map[string]interface{}{
			"check_id": checkNode.ID,
		},
	}, nil
}

// updateCheck updates a check's status
func (s *Service) updateCheck(req PolicyOperationRequest) (*PolicyOperationResponse, error) {
	err := s.evaluator.UpdateCheckStatus(req.CheckID, req.Status, req.Results)
	if err != nil {
		return &PolicyOperationResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to update check: %v", err),
		}, err
	}

	return &PolicyOperationResponse{
		Success: true,
		Message: "Check updated",
	}, nil
}

// satisfyPolicy marks a check as satisfying a policy
func (s *Service) satisfyPolicy(req PolicyOperationRequest) (*PolicyOperationResponse, error) {
	err := s.evaluator.SatisfyPolicy(req.CheckID, req.PolicyID)
	if err != nil {
		return &PolicyOperationResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to satisfy policy: %v", err),
		}, err
	}

	return &PolicyOperationResponse{
		Success: true,
		Message: "Policy satisfied",
	}, nil
}

// ListPolicies returns all policies in the environment
func (s *Service) ListPolicies() ([]interface{}, error) {
	policies := []interface{}{}

	graph, err := s.graphStore.GetGraph(s.env)
	if err != nil {
		// If environment graph doesn't exist, return empty array
		return policies, nil
	}

	for _, node := range graph.Nodes {
		if node.Kind == "policy" {
			policies = append(policies, node)
		}
	}

	return policies, nil
}

// GetPolicy returns a policy by ID
func (s *Service) GetPolicy(policyID string) (*graph.Node, error) {
	graph, err := s.graphStore.GetGraph(s.env)
	if err != nil {
		return nil, fmt.Errorf("environment not found")
	}

	policy, ok := graph.Nodes[policyID]
	if !ok || policy.Kind != "policy" {
		return nil, fmt.Errorf("policy not found")
	}

	return policy, nil
}
