package application

import (
	"testing"

	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/stretchr/testify/assert"
)

// TestNewReleaseService tests the constructor
func TestNewReleaseService(t *testing.T) {
	helpers := CreateTestHelpers(t)
	defer helpers.CleanupTestData(t)

	service := helpers.CreateTestReleaseService(t)

	assert.NotNil(t, service)
	assert.Equal(t, helpers.Graph, service.Graph)
}

// TestReleaseService_CreateRelease tests release creation
func TestReleaseService_CreateRelease(t *testing.T) {
	tests := []struct {
		name            string
		releaseContract *contracts.ReleaseContract
		expectError     bool
		errorMessage    string
	}{
		{
			name: "valid release contract",
			releaseContract: &contracts.ReleaseContract{
				Metadata: contracts.Metadata{
					Name:  "test-release-v1.0",
					Owner: "test-team",
				},
				Spec: contracts.ReleaseSpec{
					Version:         "1.0.0",
					Application:     "test-app",
					ServiceVersions: []string{"test-service:1.0.0"},
				},
			},
			expectError: false,
		},
		{
			name:            "nil release contract",
			releaseContract: nil,
			expectError:     true,
			errorMessage:    "release name is required",
		},
		{
			name: "invalid release contract - missing name",
			releaseContract: &contracts.ReleaseContract{
				Metadata: contracts.Metadata{
					Owner: "test-team",
				},
				Spec: contracts.ReleaseSpec{
					Version:     "1.0.0",
					Application: "test-app",
				},
			},
			expectError:  true,
			errorMessage: "release name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers := CreateTestHelpers(t)
			defer helpers.CleanupTestData(t)

			service := helpers.CreateTestReleaseService(t)

			var err error
			if tt.releaseContract == nil {
				// For nil contracts, we need to pass an empty contract instead
				err = service.CreateRelease(contracts.ReleaseContract{})
			} else {
				err = service.CreateRelease(*tt.releaseContract)
			}

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

// TestReleaseService_GetRelease tests release retrieval
func TestReleaseService_GetRelease(t *testing.T) {
	helpers := CreateTestHelpers(t)
	defer helpers.CleanupTestData(t)

	service := helpers.CreateTestReleaseService(t)

	// Test getting non-existent release
	_, err := service.GetRelease("non-existent-release")
	assert.Error(t, err)
}

// TestReleaseService_ListReleases tests release listing
func TestReleaseService_ListReleases(t *testing.T) {
	helpers := CreateTestHelpers(t)
	defer helpers.CleanupTestData(t)

	service := helpers.CreateTestReleaseService(t)

	// Test listing releases for non-existent app
	releases, err := service.ListReleases("non-existent-app")
	assert.NoError(t, err) // Should return empty list, not error
	assert.Empty(t, releases)
}

// TestReleaseService_DeleteRelease tests release deletion
func TestReleaseService_DeleteRelease(t *testing.T) {
	helpers := CreateTestHelpers(t)
	defer helpers.CleanupTestData(t)

	service := helpers.CreateTestReleaseService(t)

	// Test deleting non-existent release
	err := service.DeleteRelease("non-existent-release")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "release not found")
}
