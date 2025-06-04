package policies

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// Service wraps PolicyEvaluator and provides business logic methods for policy operations
// Enhanced with AI-first policy evaluation capabilities using clean AI provider interface
type Service struct {
	evaluator   *PolicyEvaluator
	aiProvider  ai.AIProvider // Clean AI provider for infrastructure-only calls
	graphStore  GraphStoreInterface
	globalGraph *graph.GlobalGraph
	env         string
	logger      *logging.Logger
}

// NewService creates a new policy service with AI capabilities
func NewService(graphStore GraphStoreInterface, globalGraph *graph.GlobalGraph, env string) *Service {
	evaluator := NewPolicyEvaluator(graphStore, env)

	// Initialize clean AI provider from config (no AI Brain dependency)
	var aiProvider ai.AIProvider
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey != "" {
		config := ai.DefaultOpenAIConfig()

		// Allow model override via environment
		if model := os.Getenv("OPENAI_MODEL"); model != "" {
			config.Model = model
		}

		// Allow base URL override for custom deployments
		if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
			config.BaseURL = baseURL
		}

		if provider, err := ai.NewOpenAIProvider(config, apiKey); err == nil {
			aiProvider = provider
		}
	}

	return &Service{
		evaluator:   evaluator,
		aiProvider:  aiProvider,
		graphStore:  graphStore,
		globalGraph: globalGraph,
		env:         env,
		logger:      logging.GetLogger().ForComponent("policy-service"),
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

// EvaluateDeploymentPolicies provides AI-powered policy evaluation for deployments
// This owns the policy domain business logic and uses AI provider for inference
func (s *Service) EvaluateDeploymentPolicies(ctx context.Context, applicationID string, environmentID string) (*ai.PolicyEvaluation, error) {
	s.logger.Info("🔍 Evaluating deployment policies: %s -> %s", applicationID, environmentID)

	// Try AI evaluation first
	if s.aiProvider != nil {
		evaluation, err := s.evaluatePoliciesWithAI(ctx, applicationID, environmentID)
		if err != nil {
			s.logger.Warn("⚠️ AI policy evaluation failed, falling back to basic evaluation: %v", err)
		} else {
			s.logger.Info("✅ AI policy evaluation completed (compliant: %t)", evaluation.Compliant)
			return evaluation, nil
		}
	}

	// Fallback to basic policy evaluation
	s.logger.Info("🔄 Using basic policy evaluation (AI unavailable)")
	return s.evaluateBasicPolicyCompliance(applicationID, environmentID), nil
}

// ValidateTransitionWithAI uses AI to validate complex policy transitions
// This provides intelligent reasoning about policy compliance beyond simple rules
func (s *Service) ValidateTransitionWithAI(ctx context.Context, fromID, toID, edgeType, user string) (*ai.PolicyEvaluation, error) {
	s.logger.Info("🔍 Validating transition with AI: %s -> %s", fromID, toID)

	// Try AI validation first
	if s.aiProvider != nil {
		evaluation, err := s.validateTransitionWithAI(ctx, fromID, toID, edgeType, user)
		if err != nil {
			s.logger.Warn("⚠️ AI transition validation failed, falling back to basic validation: %v", err)
		} else {
			s.logger.Info("✅ AI transition validation completed (allowed: %t)", evaluation.Compliant)
			return evaluation, nil
		}
	}

	// Fallback to basic transition validation
	s.logger.Info("🔄 Using basic transition validation (AI unavailable)")
	return s.validateBasicTransition(fromID, toID, edgeType, user), nil
}

// HasAICapabilities returns whether AI policy evaluation is available
func (s *Service) HasAICapabilities() bool {
	return s.aiProvider != nil
}

// evaluatePoliciesWithAI performs AI-driven policy evaluation
func (s *Service) evaluatePoliciesWithAI(ctx context.Context, applicationID, environmentID string) (*ai.PolicyEvaluation, error) {
	// Build policy context for AI evaluation
	policyContext, err := s.buildPolicyEvaluationContext(applicationID, environmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to build policy context: %w", err)
	}

	// Build prompts for AI
	systemPrompt := s.buildPolicyEvaluationSystemPrompt()
	userPrompt, err := s.buildPolicyEvaluationUserPrompt(policyContext)
	if err != nil {
		return nil, fmt.Errorf("failed to build user prompt: %w", err)
	}

	// Call AI provider
	response, err := s.aiProvider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI call failed: %w", err)
	}

	// Parse response
	evaluation, err := s.parsePolicyEvaluationResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return evaluation, nil
}

// validateTransitionWithAI performs AI-driven transition validation
func (s *Service) validateTransitionWithAI(ctx context.Context, fromID, toID, edgeType, user string) (*ai.PolicyEvaluation, error) {
	// Build transition context for AI evaluation
	transitionContext := s.buildTransitionValidationContext(fromID, toID, edgeType, user)

	// Build prompts for AI
	systemPrompt := s.buildTransitionValidationSystemPrompt()
	userPrompt, err := s.buildTransitionValidationUserPrompt(transitionContext)
	if err != nil {
		return nil, fmt.Errorf("failed to build user prompt: %w", err)
	}

	// Call AI provider
	response, err := s.aiProvider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI call failed: %w", err)
	}

	// Parse response
	evaluation, err := s.parsePolicyEvaluationResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return evaluation, nil
}

// evaluateBasicPolicyCompliance provides fallback policy evaluation
func (s *Service) evaluateBasicPolicyCompliance(applicationID, environmentID string) *ai.PolicyEvaluation {
	s.logger.Info("🔄 Generating basic policy evaluation for %s -> %s", applicationID, environmentID)

	// Basic heuristic: production environments have stricter policies
	compliant := true
	violations := []ai.PolicyViolation{}

	if environmentID == "production" {
		// In production, require additional checks
		violations = append(violations, ai.PolicyViolation{
			PolicyID:    "prod-approval-required",
			Severity:    "medium",
			Description: "Production deployments require approval",
			Suggestion:  "Obtain approval from operations team",
		})
		compliant = false
	}

	return &ai.PolicyEvaluation{
		Compliant:       compliant,
		Violations:      violations,
		Recommendations: []string{"Follow standard deployment procedures", "Monitor deployment progress"},
		Confidence:      0.7, // Lower confidence for basic evaluation
		Reasoning:       "Basic heuristic-based policy evaluation",
		Metadata: map[string]interface{}{
			"evaluation_type": "deterministic",
			"timestamp":       time.Now(),
			"environment":     environmentID,
		},
	}
}

// validateBasicTransition provides fallback transition validation
func (s *Service) validateBasicTransition(fromID, toID, edgeType, user string) *ai.PolicyEvaluation {
	s.logger.Info("🔄 Generating basic transition validation: %s -> %s", fromID, toID)

	// Basic validation rules
	compliant := true
	violations := []ai.PolicyViolation{}

	// Simple rule: deployments require proper permissions
	if edgeType == "deploy" && user != "admin" {
		violations = append(violations, ai.PolicyViolation{
			PolicyID:    "deploy-permission-required",
			Severity:    "high",
			Description: "Deployment requires administrative permissions",
			Suggestion:  "Request deployment permissions or use administrative account",
		})
		compliant = false
	}

	return &ai.PolicyEvaluation{
		Compliant:       compliant,
		Violations:      violations,
		Recommendations: []string{"Verify user permissions", "Follow deployment protocols"},
		Confidence:      0.6,
		Reasoning:       "Simple rule-based transition validation",
		Metadata: map[string]interface{}{
			"evaluation_type": "deterministic",
			"timestamp":       time.Now(),
			"transition":      fmt.Sprintf("%s->%s", fromID, toID),
		},
	}
}

// buildPolicyEvaluationContext creates comprehensive context for AI policy evaluation
func (s *Service) buildPolicyEvaluationContext(applicationID, environmentID string) (map[string]interface{}, error) {
	graph, err := s.globalGraph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	context := map[string]interface{}{
		"application_id":   applicationID,
		"environment_id":   environmentID,
		"graph_state":      s.extractGraphState(graph, applicationID),
		"policies":         s.extractRelevantPolicies(graph, environmentID),
		"constraints":      s.extractEnvironmentConstraints(environmentID),
		"current_time":     time.Now(),
		"evaluation_scope": "deployment",
	}

	return context, nil
}

// buildTransitionValidationContext creates context for transition validation
func (s *Service) buildTransitionValidationContext(fromID, toID, edgeType, user string) map[string]interface{} {
	return map[string]interface{}{
		"from_id":          fromID,
		"to_id":            toID,
		"edge_type":        edgeType,
		"user":             user,
		"current_time":     time.Now(),
		"validation_scope": "transition",
		"environment":      s.env,
	}
}

// buildPolicyEvaluationSystemPrompt creates the system prompt for policy evaluation
func (s *Service) buildPolicyEvaluationSystemPrompt() string {
	return `You are an expert AI policy evaluator for ZTDP (Zero Touch Developer Platform). Your role is to analyze policy compliance and provide intelligent recommendations.

CONTEXT:
- ZTDP enforces governance through graph-based policies
- Policies can be attached to nodes, edges, and transitions
- Your job is to evaluate compliance and suggest remediation

CAPABILITIES:
- Deep analysis of policy requirements and current state
- Intelligent compliance evaluation beyond simple rule matching
- Context-aware recommendations for policy satisfaction
- Risk assessment and mitigation strategies

RESPONSE FORMAT:
You must respond with valid JSON only:
{
  "compliant": true|false,
  "violations": [
    {
      "policy_id": "policy-name",
      "severity": "low|medium|high|critical", 
      "description": "violation description",
      "suggestion": "remediation suggestion"
    }
  ],
  "recommendations": ["recommendation1", "recommendation2"],
  "reasoning": "detailed reasoning for evaluation",
  "confidence": 0.95,
  "metadata": {}
}

PRINCIPLES:
1. Understand the intent behind policies, not just literal rules
2. Consider context and nuanced scenarios
3. Provide actionable suggestions for compliance
4. Balance governance with developer productivity
5. Explain your reasoning clearly and thoroughly`
}

// buildTransitionValidationSystemPrompt creates the system prompt for transition validation
func (s *Service) buildTransitionValidationSystemPrompt() string {
	return `You are an expert AI transition validator for ZTDP (Zero Touch Developer Platform). Your role is to validate graph transitions according to policies and security requirements.

CONTEXT:
- ZTDP manages infrastructure through a graph model
- Transitions represent state changes (deploy, create, configure, etc.)
- Your job is to validate if transitions are allowed

CAPABILITIES:
- Intelligent validation beyond simple permission checks
- Context-aware security and compliance evaluation
- Risk assessment for infrastructure changes
- Policy interpretation and application

RESPONSE FORMAT:
You must respond with valid JSON only:
{
  "compliant": true|false,
  "violations": [
    {
      "policy_id": "policy-name",
      "severity": "low|medium|high|critical",
      "description": "violation description", 
      "suggestion": "remediation suggestion"
    }
  ],
  "recommendations": ["recommendation1", "recommendation2"],
  "reasoning": "detailed reasoning for validation decision",
  "confidence": 0.95,
  "metadata": {}
}

PRINCIPLES:
1. Security and compliance come first
2. Consider blast radius and potential impact
3. Validate user permissions and context
4. Provide clear explanations for rejections
5. Suggest alternative approaches when possible`
}

// buildPolicyEvaluationUserPrompt creates the user prompt with policy context
func (s *Service) buildPolicyEvaluationUserPrompt(policyContext map[string]interface{}) (string, error) {
	return fmt.Sprintf(`Please evaluate policy compliance for this deployment context:

POLICY CONTEXT:
Application ID: %v
Environment ID: %v
Current Time: %v
Evaluation Scope: %v

GRAPH STATE:
%v

APPLICABLE POLICIES:
%v

ENVIRONMENT CONSTRAINTS:
%v

EVALUATION REQUIREMENTS:
1. Analyze all applicable policies and their requirements
2. Assess current compliance status for the deployment
3. Identify any violations or potential issues
4. Provide actionable recommendations for achieving compliance
5. Consider the business context and practical constraints

Focus on practical, actionable guidance that helps achieve compliance while maintaining developer productivity.`,
		policyContext["application_id"],
		policyContext["environment_id"],
		policyContext["current_time"],
		policyContext["evaluation_scope"],
		policyContext["graph_state"],
		policyContext["policies"],
		policyContext["constraints"]), nil
}

// buildTransitionValidationUserPrompt creates the user prompt for transition validation
func (s *Service) buildTransitionValidationUserPrompt(transitionContext map[string]interface{}) (string, error) {
	return fmt.Sprintf(`Please validate this infrastructure transition:

TRANSITION CONTEXT:
From ID: %v
To ID: %v
Edge Type: %v
User: %v
Environment: %v
Current Time: %v

VALIDATION REQUIREMENTS:
1. Verify user has appropriate permissions for this transition
2. Check if the transition violates any security policies
3. Assess potential impact and risk
4. Validate the transition makes sense in the current context
5. Consider environment-specific restrictions

Provide a clear decision with detailed reasoning for your validation result.`,
		transitionContext["from_id"],
		transitionContext["to_id"],
		transitionContext["edge_type"],
		transitionContext["user"],
		transitionContext["environment"],
		transitionContext["current_time"]), nil
}

// parsePolicyEvaluationResponse parses AI response into PolicyEvaluation
func (s *Service) parsePolicyEvaluationResponse(response string) (*ai.PolicyEvaluation, error) {
	var evaluation ai.PolicyEvaluation

	if err := json.Unmarshal([]byte(response), &evaluation); err != nil {
		return nil, fmt.Errorf("failed to parse policy evaluation: %w", err)
	}

	// Set default confidence if not provided
	if evaluation.Confidence == 0 {
		evaluation.Confidence = 0.8
	}

	return &evaluation, nil
}

// Helper methods for context extraction
func (s *Service) extractGraphState(graph *graph.Graph, applicationID string) map[string]interface{} {
	appState := map[string]interface{}{
		"nodes": []map[string]interface{}{},
		"edges": []map[string]interface{}{},
	}

	// Find application node and related nodes
	for nodeID, node := range graph.Nodes {
		if nodeID == applicationID || s.isRelatedToApplication(graph, nodeID, applicationID) {
			appState["nodes"] = append(appState["nodes"].([]map[string]interface{}), map[string]interface{}{
				"id":       node.ID,
				"kind":     node.Kind,
				"metadata": node.Metadata,
			})
		}
	}

	// Find related edges
	for fromID, edges := range graph.Edges {
		if fromID == applicationID || s.isRelatedToApplication(graph, fromID, applicationID) {
			for _, edge := range edges {
				appState["edges"] = append(appState["edges"].([]map[string]interface{}), map[string]interface{}{
					"from": fromID,
					"to":   edge.To,
					"type": edge.Type,
				})
			}
		}
	}

	return appState
}

func (s *Service) extractRelevantPolicies(graph *graph.Graph, environmentID string) []map[string]interface{} {
	policies := []map[string]interface{}{}

	for _, node := range graph.Nodes {
		if node.Kind == "policy" {
			if env, exists := node.Metadata["environment"]; exists && env == environmentID {
				policies = append(policies, map[string]interface{}{
					"policy_id":   node.ID,
					"policy_type": node.Metadata["type"],
					"severity":    node.Metadata["severity"],
					"description": node.Metadata["description"],
				})
			}
		}
	}

	return policies
}

func (s *Service) extractEnvironmentConstraints(environmentID string) map[string]interface{} {
	constraints := map[string]interface{}{
		"environment": environmentID,
	}

	// Add environment-specific constraints
	if environmentID == "production" {
		constraints["approval_required"] = true
		constraints["change_window"] = "maintenance"
		constraints["rollback_required"] = true
	}

	return constraints
}

func (s *Service) isRelatedToApplication(graph *graph.Graph, nodeID, applicationID string) bool {
	// Check if node is directly connected to application
	if edges, exists := graph.Edges[applicationID]; exists {
		for _, edge := range edges {
			if edge.To == nodeID {
				return true
			}
		}
	}

	// Check reverse direction
	if edges, exists := graph.Edges[nodeID]; exists {
		for _, edge := range edges {
			if edge.To == applicationID {
				return true
			}
		}
	}

	return false
}
