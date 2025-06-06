package contracts

import "fmt"

type EnvironmentContract struct {
	Metadata Metadata        `json:"metadata"`
	Spec     EnvironmentSpec `json:"spec"`
}

type EnvironmentSpec struct {
	Description string `json:"description"`
}

func (e EnvironmentContract) ID() string            { return e.Metadata.Name }
func (e EnvironmentContract) Kind() string          { return "environment" }
func (e EnvironmentContract) GetMetadata() Metadata { return e.Metadata }

func (e EnvironmentContract) Validate() error {
	if e.Metadata.Name == "" {
		return fmt.Errorf("environment name is required")
	}
	return nil
}
