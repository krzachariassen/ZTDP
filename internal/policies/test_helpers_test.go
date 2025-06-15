package policies

import (
	"fmt"
	"testing"
	"time"

	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

// MockEventBus for testing
type MockEventBus struct {
	events []events.Event
}

func NewMockEventBus() *MockEventBus {
	return &MockEventBus{
		events: make([]events.Event, 0),
	}
}

func (m *MockEventBus) Emit(eventType string, data map[string]interface{}) error {
	m.events = append(m.events, events.Event{
		Type:      events.EventType(eventType),
		Payload:   data,
		Timestamp: time.Now().Unix(),
		ID:        fmt.Sprintf("test-event-%d", len(m.events)),
	})
	return nil
}

func (m *MockEventBus) GetEvents() []events.Event {
	return m.events
}

func (m *MockEventBus) ClearEvents() {
	m.events = make([]events.Event, 0)
}

// MockPolicyStore for testing
type MockPolicyStore struct {
	policies map[string]*Policy
}

func NewMockPolicyStore() *MockPolicyStore {
	return &MockPolicyStore{
		policies: make(map[string]*Policy),
	}
}

func (m *MockPolicyStore) Store(policy *Policy) error {
	m.policies[policy.ID] = policy
	return nil
}

func (m *MockPolicyStore) Get(id string) (*Policy, error) {
	if policy, exists := m.policies[id]; exists {
		return policy, nil
	}
	return nil, ErrPolicyNotFound
}

func (m *MockPolicyStore) GetPoliciesForNodeType(nodeType string) ([]*Policy, error) {
	var policies []*Policy
	for _, policy := range m.policies {
		if policy.Scope == PolicyScopeNode {
			// If no specific node types defined, applies to all
			if len(policy.NodeTypes) == 0 {
				policies = append(policies, policy)
			} else {
				// Check if node type matches
				for _, policyNodeType := range policy.NodeTypes {
					if policyNodeType == nodeType {
						policies = append(policies, policy)
						break
					}
				}
			}
		}
	}
	return policies, nil
}

func (m *MockPolicyStore) GetPoliciesForEdgeType(edgeType string) ([]*Policy, error) {
	var policies []*Policy
	for _, policy := range m.policies {
		if policy.Scope == PolicyScopeEdge {
			// If no specific edge types defined, applies to all
			if len(policy.EdgeTypes) == 0 {
				policies = append(policies, policy)
			} else {
				// Check if edge type matches
				for _, policyEdgeType := range policy.EdgeTypes {
					if policyEdgeType == edgeType {
						policies = append(policies, policy)
						break
					}
				}
			}
		}
	}
	return policies, nil
}

func (m *MockPolicyStore) GetGraphPolicies() ([]*Policy, error) {
	var policies []*Policy
	for _, policy := range m.policies {
		if policy.Scope == PolicyScopeGraph {
			policies = append(policies, policy)
		}
	}
	return policies, nil
}

// Test helper functions
func createTestPolicyService(t *testing.T) (*Service, *MockEventBus, *MockPolicyStore) {
	eventBus := NewMockEventBus()
	store := NewMockPolicyStore()

	// Create a real graph backend and store for testing
	backend := graph.NewMemoryGraph()
	graphStore := graph.NewGraphStore(backend)

	// Create test service with real AI provider (not mocked!)
	// The service will use the real OpenAI provider if OPENAI_API_KEY is set
	service := NewServiceWithPolicyStore(graphStore, nil, store, "test-env", eventBus)

	if service == nil {
		t.Fatal("Failed to create policy service")
	}

	return service, eventBus, store
}

func createTestApplicationNode() *graph.Node {
	return &graph.Node{
		ID:   "test-app",
		Kind: graph.KindApplication,
		Metadata: map[string]interface{}{
			"name":        "Test Application",
			"environment": "test",
		},
		Spec: map[string]interface{}{
			"services": []string{"service1", "service2"},
		},
	}
}

func createTestDatabaseNode() *graph.Node {
	return &graph.Node{
		ID:   "test-db",
		Kind: graph.KindResource,
		Metadata: map[string]interface{}{
			"name":          "Test Database",
			"resource_type": "database",
		},
		Spec: map[string]interface{}{},
	}
}

func createApplicationServiceLimitPolicy() *Policy {
	return &Policy{
		ID:                  "app-service-limit",
		Name:                "Application Service Limit",
		Description:         "Applications must have fewer than 10 services",
		Scope:               PolicyScopeNode,
		NodeTypes:           []string{graph.KindApplication},
		NaturalLanguageRule: "Applications must have fewer than 10 services to maintain manageable complexity",
		Enforcement:         EnforcementBlock,
		RequiredConfidence:  0.8,
		CreatedAt:           time.Now(),
		Enabled:             true,
	}
}

func createDatabaseBackupPolicy() *Policy {
	return &Policy{
		ID:                  "db-backup-required",
		Name:                "Database Backup Required",
		Description:         "Database resources must have backup configuration",
		Scope:               PolicyScopeNode,
		NodeTypes:           []string{graph.KindResource},
		NaturalLanguageRule: "Database resources must have backup configuration enabled with appropriate schedules",
		Enforcement:         EnforcementBlock,
		RequiredConfidence:  0.8,
		CreatedAt:           time.Now(),
		Enabled:             true,
	}
}

func createNoDirectProdDeploymentPolicy() *Policy {
	return &Policy{
		ID:                  "no-direct-prod",
		Name:                "No Direct Production Deployment",
		Description:         "Applications must not be deployed directly to production. A deployment to production is only allowed if the same application version has been successfully deployed to at least one non-production environment (such as staging, QA, or testing) or has passed all required pre-production checks.",
		Scope:               PolicyScopeEdge,
		EdgeTypes:           []string{graph.EdgeTypeDeploy},
		NaturalLanguageRule: "Block any deployment to a production environment unless there is clear evidence in the deployment graph or metadata that the same application version has been successfully deployed to at least one non-production environment or has passed all required pre-production checks (such as code scanning, QA, or security review).",
		Enforcement:         EnforcementBlock,
		RequiredConfidence:  0.8,
		CreatedAt:           time.Now(),
		Enabled:             true,
	}
}

func createMaxAppsPerCustomerPolicy() *Policy {
	return &Policy{
		ID:                  "max-apps-per-customer",
		Name:                "Maximum Applications Per Customer",
		Description:         "Customers cannot have more than 5 applications",
		Scope:               PolicyScopeGraph,
		NaturalLanguageRule: "Each customer should not have more than 5 applications to maintain manageable complexity and resource usage",
		Enforcement:         EnforcementBlock,
		RequiredConfidence:  0.8,
		CreatedAt:           time.Now(),
		Enabled:             true,
	}
}
