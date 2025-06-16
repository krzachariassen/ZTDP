package contracts

import (
	"time"
)

// Release represents a versioned deployment package containing specific service versions
type Release struct {
	ID          string    `json:"id" yaml:"id"`
	Application string    `json:"application" yaml:"application"`
	Environment string    `json:"environment" yaml:"environment"`
	Version     string    `json:"version" yaml:"version"`
	Status      string    `json:"status" yaml:"status"` // planned, deployed, failed, rolled_back
	CreatedAt   time.Time `json:"created_at" yaml:"created_at"`
	CreatedBy   string    `json:"created_by" yaml:"created_by"`

	// Service versions included in this release
	Services []ReleaseService `json:"services" yaml:"services"`

	// Deployment strategy and configuration
	Strategy string                 `json:"strategy" yaml:"strategy"` // rolling_update, blue_green, canary
	Config   map[string]interface{} `json:"config,omitempty" yaml:"config,omitempty"`

	// Metadata for tracking and management
	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// ReleaseService represents a specific service version in a release
type ReleaseService struct {
	Name        string `json:"name" yaml:"name"`
	Version     string `json:"version" yaml:"version"`
	Image       string `json:"image" yaml:"image"`
	Replicas    int    `json:"replicas" yaml:"replicas"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// CreateReleaseRequest represents a request to create a new release
type CreateReleaseRequest struct {
	Application string                 `json:"application" yaml:"application"`
	Environment string                 `json:"environment" yaml:"environment"`
	Strategy    string                 `json:"strategy,omitempty" yaml:"strategy,omitempty"`
	UserMessage string                 `json:"user_message,omitempty" yaml:"user_message,omitempty"`
	CreatedBy   string                 `json:"created_by,omitempty" yaml:"created_by,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// CreateReleaseResponse represents the response from creating a release
type CreateReleaseResponse struct {
	Release *Release `json:"release" yaml:"release"`
	Message string   `json:"message" yaml:"message"`
	Status  string   `json:"status" yaml:"status"` // success, error
	Error   string   `json:"error,omitempty" yaml:"error,omitempty"`
}

// GetReleaseRequest represents a request to get release information
type GetReleaseRequest struct {
	ReleaseID string `json:"release_id" yaml:"release_id"`
}

// GetReleaseResponse represents the response from getting a release
type GetReleaseResponse struct {
	Release *Release `json:"release" yaml:"release"`
	Status  string   `json:"status" yaml:"status"`
	Error   string   `json:"error,omitempty" yaml:"error,omitempty"`
}

// ListReleasesRequest represents a request to list releases
type ListReleasesRequest struct {
	Application string `json:"application,omitempty" yaml:"application,omitempty"`
	Environment string `json:"environment,omitempty" yaml:"environment,omitempty"`
	Status      string `json:"status,omitempty" yaml:"status,omitempty"`
	Limit       int    `json:"limit,omitempty" yaml:"limit,omitempty"`
}

// ListReleasesResponse represents the response from listing releases
type ListReleasesResponse struct {
	Releases []Release `json:"releases" yaml:"releases"`
	Total    int       `json:"total" yaml:"total"`
	Status   string    `json:"status" yaml:"status"`
	Error    string    `json:"error,omitempty" yaml:"error,omitempty"`
}

// RollbackReleaseRequest represents a request to rollback to a previous release
type RollbackReleaseRequest struct {
	CurrentReleaseID string `json:"current_release_id" yaml:"current_release_id"`
	TargetReleaseID  string `json:"target_release_id" yaml:"target_release_id"`
	Application      string `json:"application" yaml:"application"`
	Environment      string `json:"environment" yaml:"environment"`
	UserMessage      string `json:"user_message,omitempty" yaml:"user_message,omitempty"`
	CreatedBy        string `json:"created_by,omitempty" yaml:"created_by,omitempty"`
}

// RollbackReleaseResponse represents the response from a rollback operation
type RollbackReleaseResponse struct {
	NewRelease *Release `json:"new_release" yaml:"new_release"`
	Message    string   `json:"message" yaml:"message"`
	Status     string   `json:"status" yaml:"status"`
	Error      string   `json:"error,omitempty" yaml:"error,omitempty"`
}
