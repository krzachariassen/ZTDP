package environment

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvironmentConfig(t *testing.T) {
	config := DefaultEnvironmentConfig()

	t.Run("ResolveEnvironmentName", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"dev", "development"},
			{"develop", "development"},
			{"development", "development"},
			{"staging", "staging"},
			{"stage", "staging"},
			{"stg", "staging"},
			{"prod", "production"},
			{"production", "production"},
			{"live", "production"},
			{"test", "test"},
			{"testing", "test"},
			{"qa", "test"},
			{"preprod", "preprod"},
			{"pre-prod", "preprod"},
			{"preproduction", "preprod"},
			{"sandbox", "sandbox"},
			{"sbx", "sandbox"},
			{"demo", "sandbox"},
			{"local", "local"},
			{"localhost", "local"},
			// Case insensitive
			{"DEV", "development"},
			{"PROD", "production"},
			{"Staging", "staging"},
			// Unknown environment should return as-is
			{"custom-env", "custom-env"},
		}

		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				result := config.ResolveEnvironmentName(tt.input)
				assert.Equal(t, tt.expected, result, "Failed to resolve %s to %s", tt.input, tt.expected)
			})
		}
	})

	t.Run("GetEnvironmentExamples", func(t *testing.T) {
		examples := config.GetEnvironmentExamples()

		// Should contain examples for different aliases
		assert.Contains(t, examples, `"dev environment" -> environment_name: "development"`)
		assert.Contains(t, examples, `"prod environment" -> environment_name: "production"`)
		assert.Contains(t, examples, `"staging environment" -> environment_name: "staging"`)

		// Should be properly formatted
		lines := strings.Split(examples, "\n")
		assert.Greater(t, len(lines), 5, "Should have multiple example lines")

		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				assert.Contains(t, line, "->", "Each line should contain mapping arrow")
				assert.Contains(t, line, "environment_name:", "Each line should contain environment_name")
			}
		}
	})

	t.Run("GetApprovedEnvironmentsList", func(t *testing.T) {
		envList := config.GetApprovedEnvironmentsList()

		// Should contain canonical names
		assert.Contains(t, envList, "development")
		assert.Contains(t, envList, "staging")
		assert.Contains(t, envList, "production")
		assert.Contains(t, envList, "test")

		// Should contain aliases
		assert.Contains(t, envList, "dev")
		assert.Contains(t, envList, "prod")
		assert.Contains(t, envList, "stage")

		// Should be comma-separated
		envs := strings.Split(envList, ", ")
		assert.Greater(t, len(envs), 10, "Should have many environment names")
	})
}

func TestCustomEnvironmentConfig(t *testing.T) {
	// Test with custom configuration
	customConfig := &EnvironmentConfig{
		ApprovedEnvironments: map[string][]string{
			"alpha":   {"alpha", "a"},
			"beta":    {"beta", "b"},
			"release": {"release", "rc", "release-candidate"},
		},
	}

	t.Run("CustomResolveEnvironmentName", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"a", "alpha"},
			{"alpha", "alpha"},
			{"b", "beta"},
			{"beta", "beta"},
			{"rc", "release"},
			{"release-candidate", "release"},
			{"release", "release"},
			{"unknown", "unknown"},
		}

		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				result := customConfig.ResolveEnvironmentName(tt.input)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("CustomGetApprovedEnvironmentsList", func(t *testing.T) {
		envList := customConfig.GetApprovedEnvironmentsList()

		assert.Contains(t, envList, "alpha")
		assert.Contains(t, envList, "a")
		assert.Contains(t, envList, "beta")
		assert.Contains(t, envList, "b")
		assert.Contains(t, envList, "release")
		assert.Contains(t, envList, "rc")
		assert.Contains(t, envList, "release-candidate")
	})
}
