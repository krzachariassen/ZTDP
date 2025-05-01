package contracts

import "fmt"

type ServiceContract struct {
    Metadata Metadata `json:"metadata"`
    Spec     struct {
        Application string `json:"application"` // Link to application
        Port        int    `json:"port"`
        Public      bool   `json:"public"`
    } `json:"spec"`
}

func (s ServiceContract) ID() string   { return s.Metadata.Name }
func (s ServiceContract) Kind() string { return "service" }

func (s ServiceContract) Validate() error {
    if s.Metadata.Name == "" {
        return fmt.Errorf("service name is required")
    }
    if s.Spec.Application == "" {
        return fmt.Errorf("linked application is required")
    }
    return nil
}
