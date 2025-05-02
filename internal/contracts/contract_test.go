package contracts

import (
	"testing"
)

func TestApplicationContract_Validate(t *testing.T) {
	tests := []struct {
		name     string
		contract ApplicationContract
		wantErr  bool
	}{
		{
			name: "valid application",
			contract: ApplicationContract{
				Metadata: Metadata{Name: "checkout", Environment: "dev", Owner: "team-x"},
				Spec: ApplicationSpec{
					Description:  "Handles checkout flows",
					Tags:         []string{"payments", "frontend"},
					Environments: []string{"dev", "qa"},
					Lifecycle:    map[string]LifecycleDefinition{},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			contract: ApplicationContract{
				Metadata: Metadata{Name: "", Environment: "dev", Owner: "team-x"},
				Spec: ApplicationSpec{
					Description:  "Missing name",
					Tags:         []string{"test"},
					Environments: []string{"dev"},
					Lifecycle:    map[string]LifecycleDefinition{},
				},
			},
			wantErr: true,
		},
		{
			name: "missing environments",
			contract: ApplicationContract{
				Metadata: Metadata{Name: "checkout", Environment: "dev", Owner: "team-x"},
				Spec: ApplicationSpec{
					Description:  "No envs",
					Tags:         []string{"api"},
					Environments: []string{},
					Lifecycle:    map[string]LifecycleDefinition{},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.contract.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServiceContract_Validate(t *testing.T) {
	type spec struct {
		Application string `json:"application"`
		Port        int    `json:"port"`
		Public      bool   `json:"public"`
	}

	tests := []struct {
		name     string
		contract ServiceContract
		wantErr  bool
	}{
		{
			name: "valid service",
			contract: ServiceContract{
				Metadata: Metadata{Name: "checkout-api", Environment: "dev", Owner: "team-x"},
				Spec: spec{
					Application: "checkout",
					Port:        8080,
					Public:      true,
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			contract: ServiceContract{
				Metadata: Metadata{Name: "", Environment: "dev", Owner: "team-x"},
				Spec: spec{
					Application: "checkout",
					Port:        8080,
					Public:      true,
				},
			},
			wantErr: true,
		},
		{
			name: "missing application link",
			contract: ServiceContract{
				Metadata: Metadata{Name: "checkout-api", Environment: "dev", Owner: "team-x"},
				Spec: spec{
					Application: "",
					Port:        8080,
					Public:      true,
				},
			},
			wantErr: true,
		},
		{
			name: "missing environment",
			contract: ServiceContract{
				Metadata: Metadata{Name: "checkout-api", Environment: "", Owner: "team-x"},
				Spec: spec{
					Application: "checkout",
					Port:        8080,
					Public:      true,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.contract.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
