package contracts

import (
	"fmt"
)

// EdgeContract represents an edge in the graph with validation rules
type EdgeContract struct {
	FromID   string                 `json:"from_id"`
	ToID     string                 `json:"to_id"`
	Type     string                 `json:"type"`
	FromKind string                 `json:"from_kind"`
	ToKind   string                 `json:"to_kind"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// EdgeValidationRule defines allowed edge types between node kinds
type EdgeValidationRule struct {
	FromKind     string
	ToKind       string
	AllowedTypes []string
	SpecialRules func(from, to interface{}) error // for custom validation logic
}

// Global edge validation rules - this defines the platform's edge policies
var EdgeValidationRules = []EdgeValidationRule{
	{
		FromKind:     "application",
		ToKind:       "service",
		AllowedTypes: []string{"owns"},
	},
	{
		FromKind:     "application",
		ToKind:       "resource",
		AllowedTypes: []string{"owns"},
		SpecialRules: validateApplicationToResource,
	},
	{
		FromKind:     "application",
		ToKind:       "environment",
		AllowedTypes: []string{"allowed_in"}, // Applications can be allowed to deploy to environments
	},
	{
		FromKind:     "service",
		ToKind:       "resource",
		AllowedTypes: []string{"uses"},
		SpecialRules: validateServiceToResource,
	},
	{
		FromKind:     "resource",
		ToKind:       "resource_type",
		AllowedTypes: []string{"instance_of"},
		SpecialRules: validateResourceInstanceToResourceType,
	},
	{
		FromKind:     "resource",
		ToKind:       "environment",
		AllowedTypes: []string{"deploy"}, // Resource instances deploy to environments
	},
	{
		FromKind:     "resource_register",
		ToKind:       "resource_type",
		AllowedTypes: []string{"owns"},
	},
	{
		FromKind:     "service",
		ToKind:       "service_version",
		AllowedTypes: []string{"has_version"},
	},
	{
		FromKind:     "service_version",
		ToKind:       "environment",
		AllowedTypes: []string{"deploy"},
	},
	// Policy-related edge rules
	{
		FromKind:     "check",
		ToKind:       "policy",
		AllowedTypes: []string{"satisfies"},
	},
	{
		FromKind:     "process",
		ToKind:       "policy",
		AllowedTypes: []string{"requires"},
	},
	// BLOCKED RELATIONSHIPS - These should NOT be allowed
	{
		FromKind:     "resource",
		ToKind:       "resource",
		AllowedTypes: []string{}, // NO direct resource-to-resource edges allowed
	},
	{
		FromKind:     "resource_register",
		ToKind:       "application",
		AllowedTypes: []string{}, // Resource register should NOT own applications
	},
	{
		FromKind:     "application",
		ToKind:       "resource_type",
		AllowedTypes: []string{}, // Applications should NOT directly own resource types
	},
	{
		FromKind:     "service",
		ToKind:       "resource_type",
		AllowedTypes: []string{}, // Services should NOT directly use resource types
	},
	// Test node rules (for testing purposes)
	{
		FromKind:     "test",
		ToKind:       "test",
		AllowedTypes: []string{"deploy", "create", "owns", "uses"},
	},
	// Add more rules as needed
}

// Validate validates the edge according to platform policies
func (e EdgeContract) Validate() error {
	// Find applicable rule
	var applicableRule *EdgeValidationRule
	for _, rule := range EdgeValidationRules {
		if rule.FromKind == e.FromKind && rule.ToKind == e.ToKind {
			applicableRule = &rule
			break
		}
	}

	if applicableRule == nil {
		return fmt.Errorf("edge type '%s' not allowed from %s (%s) to %s (%s)",
			e.Type, e.FromID, e.FromKind, e.ToID, e.ToKind)
	}

	// Check if edge type is in allowed list
	typeAllowed := false
	for _, allowedType := range applicableRule.AllowedTypes {
		if allowedType == e.Type {
			typeAllowed = true
			break
		}
	}

	if !typeAllowed {
		return fmt.Errorf("edge type '%s' not allowed from %s to %s (allowed: %v)",
			e.Type, e.FromKind, e.ToKind, applicableRule.AllowedTypes)
	}

	// Apply special validation rules if they exist
	if applicableRule.SpecialRules != nil {
		// Note: we'll need to pass the actual node data here
		// For now, we'll implement this in a way that the graph can call it
		// with full node information
	}

	return nil
}

// Special validation functions

// validateApplicationToResource ensures applications only own resource instances
func validateApplicationToResource(from, to interface{}) error {
	toNode, ok := to.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid node data for resource validation")
	}

	// Check if this is a resource instance (has application and catalog_ref metadata)
	metadata, ok := toNode["metadata"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("resource node missing metadata")
	}

	if app, hasApp := metadata["application"]; !hasApp || app == nil {
		return fmt.Errorf("applications can only own resource instances, not catalog resources")
	}

	if catRef, hasCatRef := metadata["catalog_ref"]; !hasCatRef || catRef == nil {
		return fmt.Errorf("applications can only own resource instances, not catalog resources")
	}

	return nil
}

// validateServiceToResource ensures services only use resource instances
func validateServiceToResource(from, to interface{}) error {
	return validateApplicationToResource(from, to) // Same logic
}

// validateResourceInstanceToResourceType ensures only instance_of edges from instances to resource_types
func validateResourceInstanceToResourceType(from, to interface{}) error {
	fromNode, ok := from.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid from-node data for resource validation")
	}

	// From should be a resource instance, To should be a resource_type
	fromMetadata, ok := fromNode["metadata"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("from-resource node missing metadata")
	}

	// From node should be a resource instance (has application metadata)
	if app, hasApp := fromMetadata["application"]; !hasApp || app == nil {
		return fmt.Errorf("instance_of edges can only originate from resource instances")
	}

	// To node should be a resource_type (validate its kind)
	toNode, ok := to.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid to-node data for resource type validation")
	}

	// The to-node should be a resource_type (this is enforced by the edge rule, but we double-check)
	if kind, ok := toNode["kind"].(string); !ok || kind != "resource_type" {
		return fmt.Errorf("instance_of edges can only target resource_type nodes")
	}

	return nil
}
