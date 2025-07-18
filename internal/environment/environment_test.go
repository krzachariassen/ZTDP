package environment

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

func TestEnvironmentService_ExtractEnvironmentParameters(t *testing.T) {
	tests := []struct {
		name        string
		userMessage string
		expected    EnvironmentDomainParams
		wantErr     bool
	}{
		{
			name:        "create environment with full details",
			userMessage: "Create an environment called production owned by devops team for the main application",
			expected: EnvironmentDomainParams{
				Action:          "create",
				EnvironmentName: "production",
				Owner:           "devops",
				Description:     "main application",
				EnvType:         "production",
				Confidence:      0.95,
			},
			wantErr: false,
		},
		{
			name:        "list environments",
			userMessage: "list all environments",
			expected: EnvironmentDomainParams{
				Action:     "list",
				Confidence: 0.9,
			},
			wantErr: false,
		},
		{
			name:        "show specific environment",
			userMessage: "show the staging environment",
			expected: EnvironmentDomainParams{
				Action:          "show",
				EnvironmentName: "staging",
				Confidence:      0.85,
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

			// Create environment service with AI provider
			service := NewAIEnvironmentService(nil, mockAI, &events.EventBus{})

			// Debug: Print what we're testing
			t.Logf("Testing with message: %s", tt.userMessage)
			t.Logf("Expected response: %s", mockAI.expectedResponse)

			// Test parameter extraction
			params, err := service.ExtractEnvironmentParameters(context.Background(), tt.userMessage)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Action, params.Action)
			assert.Equal(t, tt.expected.EnvironmentName, params.EnvironmentName)
			assert.Equal(t, tt.expected.Owner, params.Owner)

			if tt.expected.Description != "" {
				assert.Equal(t, tt.expected.Description, params.Description)
			}
			if tt.expected.EnvType != "" {
				assert.Equal(t, tt.expected.EnvType, params.EnvType)
			}
		})
	}
}
