package contracts

import "fmt"

type ApplicationSpec struct {
	Description  string
	Tags         []string
	Environments []string
	Lifecycle    map[string]LifecycleDefinition
}

type ApplicationContract struct {
	Metadata Metadata        `json:"metadata"`
	Spec     ApplicationSpec `json:"spec"`
}

func (a ApplicationContract) ID() string            { return a.Metadata.Name }
func (a ApplicationContract) Kind() string          { return "application" }
func (a ApplicationContract) GetMetadata() Metadata { return a.Metadata }

func (a ApplicationContract) Validate() error {
	if a.Metadata.Name == "" {
		return fmt.Errorf("application name is required")
	}
	if len(a.Spec.Environments) == 0 {
		return fmt.Errorf("at least one environment must be defined")
	}
	return nil
}
