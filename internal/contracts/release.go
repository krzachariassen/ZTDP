package contracts

import (
	"fmt"
	"time"
)

type ReleaseChange struct {
	Service     string `json:"service"`
	FromVersion string `json:"from_version"`
	ToVersion   string `json:"to_version"`
	ChangeType  string `json:"change_type"` // "added", "updated", "removed"
}

type ReleaseSpec struct {
	Application     string            `json:"application"`
	Version         string            `json:"version"`
	ServiceVersions []string          `json:"service_versions"`
	Status          string            `json:"status"`
	Strategy        string            `json:"strategy"`
	Configuration   map[string]string `json:"configuration"`
	Notes           string            `json:"notes"`
	Timestamp       time.Time         `json:"timestamp"`
	Changes         []ReleaseChange   `json:"changes,omitempty"`
}

type ReleaseContract struct {
	Metadata Metadata    `json:"metadata"`
	Spec     ReleaseSpec `json:"spec"`
}

func (r ReleaseContract) ID() string            { return r.Metadata.Name }
func (r ReleaseContract) Kind() string          { return "release" }
func (r ReleaseContract) GetMetadata() Metadata { return r.Metadata }

func (r ReleaseContract) Validate() error {
	if r.Metadata.Name == "" {
		return fmt.Errorf("release name is required")
	}
	if r.Spec.Application == "" {
		return fmt.Errorf("application is required")
	}
	if r.Spec.Version == "" {
		return fmt.Errorf("version is required")
	}
	if len(r.Spec.ServiceVersions) == 0 {
		return fmt.Errorf("at least one service_version is required")
	}
	if r.Spec.Status == "" {
		r.Spec.Status = "pending"
	}
	if r.Spec.Strategy == "" {
		r.Spec.Strategy = "rolling"
	}
	if r.Spec.Timestamp.IsZero() {
		r.Spec.Timestamp = time.Now()
	}
	return nil
}

// Release-specific request/response types for agent communication
type CreateReleaseRequest struct {
	Application     string   `json:"application"`
	ServiceVersions []string `json:"service_versions"`
	Strategy        string   `json:"strategy,omitempty"`
	Notes           string   `json:"notes,omitempty"`
	UserMessage     string   `json:"user_message,omitempty"`
}

type CreateReleaseResponse struct {
	Release *ReleaseContract `json:"release"`
	Message string           `json:"message"`
	Success bool             `json:"success"`
}

type ListReleasesRequest struct {
	Application string `json:"application,omitempty"`
	Limit       int    `json:"limit,omitempty"`
}

type ListReleasesResponse struct {
	Releases []*ReleaseContract `json:"releases"`
	Count    int                `json:"count"`
	Success  bool               `json:"success"`
}

type GetReleaseRequest struct {
	ReleaseID string `json:"release_id"`
}

type GetReleaseResponse struct {
	Release *ReleaseContract `json:"release"`
	Success bool             `json:"success"`
}
