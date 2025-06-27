package application

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	aiInfrastructure "github.com/ztdp/orchestrator/internal/ai/infrastructure"
	"github.com/ztdp/orchestrator/internal/logging"
)

func TestServiceFactory_CreateAIProvider(t *testing.T) {
	// TDD: RED - Test creating AI provider with factory

	// Setup
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY environment variable not set, skipping AI provider factory tests")
	}

	config := aiInfrastructure.DefaultOpenAIConfig()
	config.APIKey = apiKey
	config.Model = "gpt-3.5-turbo"

	logger, err := logging.NewLogger(false)
	require.NoError(t, err)

	// Execute: Create AI provider using factory method
	aiProvider := CreateAIProvider(config, logger)

	// Verify: Provider should be properly created
	assert.NotNil(t, aiProvider)

	// Get provider info
	info := aiProvider.GetProviderInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "openai", info.Name)
	assert.Equal(t, "gpt-3.5-turbo", info.Model)

	t.Log("Service factory successfully created AI provider")
}

func TestNewServiceFactory(t *testing.T) {
	// TDD: RED - Test factory constructor

	logger, err := logging.NewLogger(false)
	require.NoError(t, err)

	// We don't need real graph/messageBus for constructor test
	factory := NewServiceFactory(logger, nil, nil, nil)

	// Verify: Factory should be created
	assert.NotNil(t, factory)
	assert.Equal(t, logger, factory.logger)

	t.Log("Service factory constructor works correctly")
}
