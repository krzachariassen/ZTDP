package contracts

type Contract interface {
	ID() string
	Kind() string
	Validate() error
	GetMetadata() Metadata
}

type Metadata struct {
	Name        string `json:"name"`
	Environment string `json:"environment"`
	Owner       string `json:"owner"`
}

type LifecycleDefinition struct {
	Gates []string `json:"gates"`
}
