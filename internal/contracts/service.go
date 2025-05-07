package contracts

import "fmt"

type ServiceSpec struct {
	Application string `json:"application"`
	Port        int    `json:"port"`
	Public      bool   `json:"public"`
}

type ServiceContract struct {
	Metadata Metadata    `json:"metadata"`
	Spec     ServiceSpec `json:"spec"`
}

func (s ServiceContract) ID() string            { return s.Metadata.Name }
func (s ServiceContract) Kind() string          { return "service" }
func (s ServiceContract) GetMetadata() Metadata { return s.Metadata }

func (s ServiceContract) Validate() error {
	if s.Metadata.Name == "" {
		return fmt.Errorf("service name is required")
	}
	if s.Spec.Application == "" {
		return fmt.Errorf("linked application is required")
	}
	return nil
}
