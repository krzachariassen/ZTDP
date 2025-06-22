package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/stretchr/testify/assert"
)

// MockAIProvider for testing
type MockAIProvider struct {
	expectedResponse string
	shouldError      bool
}

func (m *MockAIProvider) CallAI(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if m.shouldError {
		return "", assert.AnError
	}
	return m.expectedResponse, nil
}

func (m *MockAIProvider) GetProviderInfo() *ai.ProviderInfo {
	return &ai.ProviderInfo{Name: "mock", Version: "1.0"}
}

func (m *MockAIProvider) Close() error {
	return nil
}

func TestServiceService_ExtractServiceParameters(t *testing.T) {
	tests := []struct {
		name        string
		userMessage string
		expected    ServiceDomainParams
		wantErr     bool
	}{
		{
			name:        "create service with full details",
			userMessage: "Create a service called checkout-api for the checkout application on port 8080 that is public facing",
			expected: ServiceDomainParams{
				Action:          "create",
				ServiceName:     "checkout-api",
				ApplicationName: "checkout",
				Port:            8080,
				Public:          true,
				Confidence:      0.95,
			},
			wantErr: false,
		},
		{
			name:        "list services for application",
			userMessage: "list services for myapp",
			expected: ServiceDomainParams{
				Action:          "list",
				ApplicationName: "myapp",
				Confidence:      0.9,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock AI provider
			mockAI := &MockAIProvider{}

			// Mock the AI response based on expected result
			expectedResponse, _ := json.Marshal(tt.expected)
			mockAI.expectedResponse = string(expectedResponse)

			// Create service with AI provider
			service := NewAIServiceService(nil, mockAI, &events.EventBus{})

			// Debug: Print what we're testing
			t.Logf("Testing with message: %s", tt.userMessage)
			t.Logf("Expected response: %s", mockAI.expectedResponse)

			// Test parameter extraction
			params, err := service.ExtractServiceParameters(context.Background(), tt.userMessage)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Action, params.Action)
			assert.Equal(t, tt.expected.ServiceName, params.ServiceName)
			assert.Equal(t, tt.expected.ApplicationName, params.ApplicationName)

			if tt.expected.Port > 0 {
				assert.Equal(t, tt.expected.Port, params.Port)
			}
			if tt.expected.Public {
				assert.Equal(t, tt.expected.Public, params.Public)
			}
		})
	}
}
