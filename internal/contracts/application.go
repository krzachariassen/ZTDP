package contracts

import "fmt"

type ApplicationContract struct {
    Metadata Metadata `json:"metadata"`
    Spec     struct {
        Description string   `json:"description"`
        Tags        []string `json:"tags,omitempty"`
    } `json:"spec"`
}

func (a ApplicationContract) ID() string   { return a.Metadata.Name }
func (a ApplicationContract) Kind() string { return "application" }

func (a ApplicationContract) Validate() error {
    if a.Metadata.Name == "" {
        return fmt.Errorf("application name is required")
    }
    return nil
}
