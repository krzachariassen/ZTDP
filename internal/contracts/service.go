package contracts

import (
	"fmt"
	"time"
)

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

// ServiceVersionContract represents a versioned service artifact.
type ServiceVersionContract struct {
	IDValue   string    `json:"id"`
	Name      string    `json:"name"`
	Owner     string    `json:"owner,omitempty"`
	Version   string    `json:"version"`
	ConfigRef string    `json:"config_ref"`
	CreatedAt time.Time `json:"created_at"`
}

func (svc ServiceVersionContract) ID() string   { return svc.IDValue }
func (svc ServiceVersionContract) Kind() string { return "service_version" }
func (svc ServiceVersionContract) GetMetadata() Metadata {
	return Metadata{Name: svc.Name, Owner: svc.Owner}
}
func (svc ServiceVersionContract) Validate() error {
	if svc.Name == "" || svc.Version == "" {
		return fmt.Errorf("service_version: name and version are required")
	}
	return nil
}
