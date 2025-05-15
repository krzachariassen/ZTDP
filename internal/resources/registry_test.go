package resources

import (
	"fmt"
	"testing"

	"github.com/krzachariassen/ZTDP/internal/contracts"
)

// Define a mock resource contract for testing
type MockResourceContract struct {
	Metadata contracts.Metadata `json:"metadata"`
	Spec     struct {
		TestField string `json:"test_field"`
	} `json:"spec"`
}

func (m MockResourceContract) ID() string {
	return m.Metadata.Name
}

func (m MockResourceContract) Kind() string {
	return "mock_resource"
}

func (m MockResourceContract) GetMetadata() contracts.Metadata {
	return m.Metadata
}

func (m MockResourceContract) Validate() error {
	if m.Metadata.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

func TestResourceRegistration(t *testing.T) {
	// Register a mock resource
	RegisterResourceKind("mock_resource", func() contracts.Contract {
		return &MockResourceContract{}
	})

	// Test that we can get the factory
	factory, ok := GetResourceFactory("mock_resource")
	if !ok {
		t.Fatal("Expected mock_resource to be registered")
	}

	// Create an instance
	instance := factory()
	if instance.Kind() != "mock_resource" {
		t.Errorf("Expected kind 'mock_resource', got %s", instance.Kind())
	}
}

func TestLoadNodeFromSpec(t *testing.T) {
	// Register a mock resource
	RegisterResourceKind("mock_resource", func() contracts.Contract {
		return &MockResourceContract{}
	})

	// Test data
	metadata := contracts.Metadata{
		Name:  "test-resource",
		Owner: "test-owner",
	}

	spec := map[string]interface{}{
		"test_field": "test-value",
	}

	// Load the contract
	contract, err := LoadNodeFromSpec("mock_resource", spec, metadata)
	if err != nil {
		t.Fatalf("Failed to load node: %v", err)
	}

	// Check the contract properties
	if contract.ID() != "test-resource" {
		t.Errorf("Expected ID 'test-resource', got %s", contract.ID())
	}

	// Cast to our mock type
	mockContract, ok := contract.(*MockResourceContract)
	if !ok {
		t.Fatal("Failed to cast contract to MockResourceContract")
	}

	if mockContract.Spec.TestField != "test-value" {
		t.Errorf("Expected test_field 'test-value', got %s", mockContract.Spec.TestField)
	}
}

func TestLoadNodeFromSpec_UnknownKind(t *testing.T) {
	_, err := LoadNodeFromSpec("unknown_kind", map[string]interface{}{}, contracts.Metadata{})
	if err == nil {
		t.Error("Expected error for unknown kind, got nil")
	}
}
