package contracts

type Contract interface {
    ID() string
    Kind() string
    Validate() error
}

type Metadata struct {
    Name        string `json:"name"`
    Environment string `json:"environment"`
    Owner       string `json:"owner"`
}
