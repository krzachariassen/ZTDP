package contracts

import (
	"fmt"
)

// ResourceTypeSpec defines the specification for a resource type in the catalog
type ResourceTypeSpec struct {
	Version         string   `json:"version"`
	DefaultTier     string   `json:"default_tier"`
	TierOptions     []string `json:"tier_options"`
	ConfigTemplate  string   `json:"config_template"`
	AvailablePlans  []string `json:"available_plans"`
	DefaultCapacity string   `json:"default_capacity"`
	// Additional fields can be added by specific resource providers
	ProviderMetadata map[string]interface{} `json:"provider_metadata,omitempty"`
}

// ResourceTypeContract represents a resource type in the catalog (template)
type ResourceTypeContract struct {
	Metadata Metadata         `json:"metadata"`
	Spec     ResourceTypeSpec `json:"spec"`
}

func (rt ResourceTypeContract) ID() string            { return rt.Metadata.Name }
func (rt ResourceTypeContract) Kind() string          { return "resource_type" }
func (rt ResourceTypeContract) GetMetadata() Metadata { return rt.Metadata }

func (rt ResourceTypeContract) Validate() error {
	if rt.Metadata.Name == "" {
		return fmt.Errorf("resource type name is required")
	}
	if rt.Spec.Version == "" {
		return fmt.Errorf("resource type version is required")
	}
	return nil
}

// ResourceSpec defines the specification for a resource instance
type ResourceSpec struct {
	Type     string `json:"type"` // References the resource_type
	Version  string `json:"version"`
	Tier     string `json:"tier"`
	Capacity string `json:"capacity,omitempty"`
	Plan     string `json:"plan,omitempty"`
	// Additional fields can be added by specific resource providers
	ProviderConfig map[string]interface{} `json:"provider_config,omitempty"`
}

// ResourceContract represents a resource instance owned by an application
type ResourceContract struct {
	Metadata Metadata     `json:"metadata"`
	Spec     ResourceSpec `json:"spec"`
}

func (r ResourceContract) ID() string            { return r.Metadata.Name }
func (r ResourceContract) Kind() string          { return "resource" }
func (r ResourceContract) GetMetadata() Metadata { return r.Metadata }

func (r ResourceContract) Validate() error {
	if r.Metadata.Name == "" {
		return fmt.Errorf("resource name is required")
	}
	if r.Spec.Type == "" {
		return fmt.Errorf("resource type reference is required")
	}
	return nil
}
