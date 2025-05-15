// registry.go
package resources

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/krzachariassen/ZTDP/internal/contracts"
)

var (
	// registry is a thread-safe map of resource kinds to factory functions
	registry   = make(map[string]func() contracts.Contract)
	registryMu sync.RWMutex
)

func init() {
	// Register default resource types
	RegisterResourceKind("application", func() contracts.Contract { return &contracts.ApplicationContract{} })
	RegisterResourceKind("service", func() contracts.Contract { return &contracts.ServiceContract{} })
	RegisterResourceKind("environment", func() contracts.Contract { return &contracts.EnvironmentContract{} })
	RegisterResourceKind("resource_type", func() contracts.Contract { return &contracts.ResourceTypeContract{} })
	RegisterResourceKind("resource", func() contracts.Contract { return &contracts.ResourceContract{} })
}

// RegisterResourceKind adds a new resource kind to the registry
// This allows third-party resource providers to register their own custom resource types
func RegisterResourceKind(kind string, factory func() contracts.Contract) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[kind] = factory
}

// GetResourceFactory returns the factory function for a given resource kind
func GetResourceFactory(kind string) (func() contracts.Contract, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	factory, ok := registry[kind]
	return factory, ok
}

// LoadNodeFromSpec hydrates a contract from kind and spec
func LoadNodeFromSpec(kind string, spec map[string]interface{}, metadata contracts.Metadata) (contracts.Contract, error) {
	factory, ok := GetResourceFactory(kind)
	if !ok {
		return nil, fmt.Errorf("unknown resource kind: %s", kind)
	}

	resource := factory()

	// Build the full contract map
	contractMap := map[string]interface{}{
		"metadata": metadata,
		"spec":     spec,
	}

	data, err := json.Marshal(contractMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal contract data: %w", err)
	}

	if err := json.Unmarshal(data, resource); err != nil {
		return nil, fmt.Errorf("failed to unmarshal into resource: %w", err)
	}

	return resource, nil
}
