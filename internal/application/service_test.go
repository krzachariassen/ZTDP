package application

import (
	"context"
	"testing"

	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/stretchr/testify/assert"
)

// TestNewServiceService tests the constructor
func TestNewServiceService(t *testing.T) {
	helpers := CreateTestHelpers(t)
	defer helpers.CleanupTestData(t)

	service := helpers.CreateTestServiceService(t)

	assert.NotNil(t, service)
	assert.Equal(t, helpers.Graph, service.Graph)
}

// TestServiceService_CreateService tests service creation
func TestServiceService_CreateService(t *testing.T) {
	tests := []struct {
		name         string
		appName      string
		serviceData  map[string]interface{}
		expectError  bool
		errorMessage string
	}{
		{
			name:    "valid service",
			appName: "test-app",
			serviceData: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":  "test-service",
					"owner": "test-team",
				},
				"spec": map[string]interface{}{
					"application": "test-app",
					"port":        8080,
					"public":      true,
				},
			},
			expectError: false,
		},
		{
			name:        "empty app name",
			appName:     "",
			serviceData: map[string]interface{}{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers := CreateTestHelpers(t)
			defer helpers.CleanupTestData(t)

			service := helpers.CreateTestServiceService(t)

			_, err := service.CreateService(tt.appName, tt.serviceData)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestServiceService_CreateServiceFromContract tests service creation from contract
func TestServiceService_CreateServiceFromContract(t *testing.T) {
	tests := []struct {
		name            string
		serviceContract *contracts.ServiceContract
		expectError     bool
		errorMessage    string
	}{
		{
			name: "valid service contract",
			serviceContract: &contracts.ServiceContract{
				Metadata: contracts.Metadata{
					Name:  "test-service",
					Owner: "test-team",
				},
				Spec: contracts.ServiceSpec{
					Application: "test-app",
					Port:        8080,
					Public:      true,
				},
			},
			expectError: false,
		},
		{
			name:            "nil service contract",
			serviceContract: nil,
			expectError:     true,
			errorMessage:    "service contract cannot be nil",
		},
		{
			name: "invalid service contract - missing name",
			serviceContract: &contracts.ServiceContract{
				Metadata: contracts.Metadata{
					Owner: "test-team",
				},
				Spec: contracts.ServiceSpec{
					Application: "test-app",
					Port:        8080,
				},
			},
			expectError:  true,
			errorMessage: "service name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers := CreateTestHelpers(t)
			defer helpers.CleanupTestData(t)

			service := helpers.CreateTestServiceService(t)

			_, err := service.CreateServiceFromContract(context.Background(), tt.serviceContract)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestServiceService_GetService tests service retrieval
func TestServiceService_GetService(t *testing.T) {
	helpers := CreateTestHelpers(t)
	defer helpers.CleanupTestData(t)

	service := helpers.CreateTestServiceService(t)

	// Test getting non-existent service
	_, err := service.GetService("test-app", "non-existent-service")
	assert.Error(t, err)
}

// TestServiceService_ListServices tests service listing
func TestServiceService_ListServices(t *testing.T) {
	helpers := CreateTestHelpers(t)
	defer helpers.CleanupTestData(t)

	service := helpers.CreateTestServiceService(t)

	// Test listing services for non-existent app
	services, err := service.ListServices("non-existent-app")
	assert.NoError(t, err) // Should return empty list, not error
	assert.Empty(t, services)
}
